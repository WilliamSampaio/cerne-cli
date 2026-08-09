package githubapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type PullRequestRequest struct {
	RemoteURL  string
	Base       string
	Head       string
	Title      string
	Body       string
	Env        []string
	APIBaseURL string
	HTTPClient *http.Client
	UserAgent  string
}

type PullRequestResult struct {
	Number    int    `json:"number"`
	URL       string `json:"url"`
	Outcome   string `json:"outcome"`
	RequestID string `json:"request_id,omitempty"`
}

type Failure struct {
	Code string
}

func (failure Failure) Error() string { return failure.Code }

type repository struct {
	Owner string
	Name  string
}

func OpenPullRequest(ctx context.Context, request PullRequestRequest) (PullRequestResult, error) {
	repo, err := NormalizeRemote(request.RemoteURL)
	if err != nil {
		return PullRequestResult{}, Failure{Code: "github_remote_required"}
	}
	token := tokenFromEnv(request.Env)
	if token == "" {
		return PullRequestResult{}, Failure{Code: "authentication_unavailable"}
	}
	if request.Base == "" || request.Head == "" || request.Title == "" {
		return PullRequestResult{}, Failure{Code: "validation_failed"}
	}
	client := request.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second, CheckRedirect: sameGitHubRedirect}
	}
	apiBase := strings.TrimRight(request.APIBaseURL, "/")
	if apiBase == "" {
		apiBase = "https://api.github.com"
	}
	agent := request.UserAgent
	if agent == "" {
		agent = "cerne/dev"
	}
	if existing, ok, err := findOpenPR(ctx, client, apiBase, agent, token, repo, request.Base, request.Head); err != nil || ok {
		return existing, err
	}
	created, err := createPR(ctx, client, apiBase, agent, token, repo, request)
	if err == nil {
		created.Outcome = "created"
		return created, nil
	}
	var failure Failure
	if !errors.As(err, &failure) || (failure.Code != "validation_failed" && failure.Code != "remote_result_unknown") {
		return PullRequestResult{}, err
	}
	reconciled, ok, findErr := findOpenPR(ctx, client, apiBase, agent, token, repo, request.Base, request.Head)
	if findErr != nil {
		return PullRequestResult{}, err
	}
	if ok {
		reconciled.Outcome = "reconciled"
		return reconciled, nil
	}
	return PullRequestResult{}, err
}

func NormalizeRemote(raw string) (repository, error) {
	var host, path string
	if strings.HasPrefix(raw, "git@") {
		rest := strings.TrimPrefix(raw, "git@")
		host, path, _ = strings.Cut(rest, ":")
	} else {
		parsed, err := url.Parse(raw)
		if err != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
			return repository{}, fmt.Errorf("remote inválido")
		}
		if parsed.User != nil && !(parsed.Scheme == "ssh" && parsed.User.Username() == "git") {
			return repository{}, fmt.Errorf("remote inválido")
		}
		host = parsed.Hostname()
		path = strings.TrimPrefix(parsed.Path, "/")
	}
	if !strings.EqualFold(host, "github.com") {
		return repository{}, fmt.Errorf("remote não GitHub")
	}
	path = strings.TrimSuffix(path, ".git")
	parts := strings.Split(path, "/")
	if len(parts) != 2 || unsafeSegment(parts[0]) || unsafeSegment(parts[1]) {
		return repository{}, fmt.Errorf("remote inválido")
	}
	return repository{Owner: parts[0], Name: parts[1]}, nil
}

func findOpenPR(ctx context.Context, client *http.Client, apiBase, agent, token string, repo repository, base, head string) (PullRequestResult, bool, error) {
	values := url.Values{}
	values.Set("state", "open")
	values.Set("head", repo.Owner+":"+head)
	values.Set("base", base)
	values.Set("per_page", "1")
	endpoint := fmt.Sprintf("%s/repos/%s/%s/pulls?%s", apiBase, url.PathEscape(repo.Owner), url.PathEscape(repo.Name), values.Encode())
	var results []PullRequestResult
	status, requestID, err := doJSON(ctx, client, http.MethodGet, endpoint, agent, token, nil, &results)
	if err != nil {
		return PullRequestResult{}, false, err
	}
	if status != http.StatusOK {
		return PullRequestResult{}, false, mapStatus(status)
	}
	if len(results) == 0 {
		return PullRequestResult{}, false, nil
	}
	results[0].Outcome = "existing"
	results[0].RequestID = requestID
	if !validGitHubURL(results[0].URL) {
		return PullRequestResult{}, false, Failure{Code: "invalid_remote_response"}
	}
	return results[0], true, nil
}

func createPR(ctx context.Context, client *http.Client, apiBase, agent, token string, repo repository, request PullRequestRequest) (PullRequestResult, error) {
	endpoint := fmt.Sprintf("%s/repos/%s/%s/pulls", apiBase, url.PathEscape(repo.Owner), url.PathEscape(repo.Name))
	body, _ := json.Marshal(map[string]string{"title": request.Title, "body": request.Body, "head": request.Head, "base": request.Base})
	var result PullRequestResult
	status, requestID, err := doJSON(ctx, client, http.MethodPost, endpoint, agent, token, body, &result)
	if err != nil {
		return PullRequestResult{}, Failure{Code: "remote_result_unknown"}
	}
	if status != http.StatusCreated {
		return PullRequestResult{}, mapStatus(status)
	}
	result.RequestID = requestID
	if result.Number <= 0 || !validGitHubURL(result.URL) {
		return PullRequestResult{}, Failure{Code: "invalid_remote_response"}
	}
	return result, nil
}

func doJSON(ctx context.Context, client *http.Client, method, endpoint, agent, token string, body []byte, target any) (int, string, error) {
	req, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(body))
	if err != nil {
		return 0, "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-GitHub-Api-Version", "2026-03-10")
	req.Header.Set("User-Agent", agent)
	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	requestID := resp.Header.Get("X-GitHub-Request-Id")
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(target); err != nil {
			return resp.StatusCode, requestID, Failure{Code: "invalid_remote_response"}
		}
	} else {
		io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
	}
	return resp.StatusCode, requestID, nil
}

func mapStatus(status int) error {
	switch status {
	case http.StatusUnauthorized:
		return Failure{Code: "authentication_unavailable"}
	case http.StatusForbidden:
		return Failure{Code: "permission_denied"}
	case http.StatusTooManyRequests:
		return Failure{Code: "rate_limited"}
	case http.StatusNotFound:
		return Failure{Code: "remote_not_found"}
	case http.StatusGone:
		return Failure{Code: "api_version_unsupported"}
	case http.StatusUnprocessableEntity:
		return Failure{Code: "validation_failed"}
	}
	if status >= 500 {
		return Failure{Code: "remote_result_unknown"}
	}
	return Failure{Code: "invalid_remote_response"}
}

func tokenFromEnv(env []string) string {
	if env == nil {
		env = os.Environ()
	}
	values := map[string]string{}
	for _, item := range env {
		key, value, ok := strings.Cut(item, "=")
		if ok {
			values[key] = value
		}
	}
	if values["GH_TOKEN"] != "" {
		return values["GH_TOKEN"]
	}
	return values["GITHUB_TOKEN"]
}

func sameGitHubRedirect(req *http.Request, via []*http.Request) error {
	if req.URL.Scheme != "https" || !strings.EqualFold(req.URL.Hostname(), "api.github.com") {
		return http.ErrUseLastResponse
	}
	return nil
}

func validGitHubURL(raw string) bool {
	parsed, err := url.Parse(raw)
	return err == nil && parsed.Scheme == "https" && strings.EqualFold(parsed.Hostname(), "github.com")
}

func unsafeSegment(value string) bool {
	return value == "" || value == "." || value == ".." || strings.ContainsAny(value, "?#:")
}
