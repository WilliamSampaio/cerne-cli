package githubapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOpenPullRequestExistingAndCreated(t *testing.T) {
	t.Run("existing uses one GET and no POST", func(t *testing.T) {
		requests := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requests++
			if r.Method != http.MethodGet || r.URL.Path != "/repos/acme/app/pulls" ||
				r.URL.Query().Get("head") != "acme:feature/x" || r.URL.Query().Get("base") != "main" ||
				r.Header.Get("Authorization") != "Bearer gh" || r.Header.Get("Accept") != "application/vnd.github+json" ||
				r.Header.Get("X-GitHub-Api-Version") != "2026-03-10" {
				t.Fatalf("bad request: %s %s headers=%v", r.Method, r.URL.String(), r.Header)
			}
			w.Header().Set("X-GitHub-Request-Id", "rid")
			json.NewEncoder(w).Encode([]PullRequestResult{{Number: 7, URL: "https://github.com/acme/app/pull/7"}})
		}))
		defer server.Close()
		got, err := OpenPullRequest(context.Background(), PullRequestRequest{RemoteURL: "https://github.com/acme/app.git", Base: "main", Head: "feature/x", Title: "Title", Env: []string{"GH_TOKEN=gh"}, APIBaseURL: server.URL, HTTPClient: server.Client()})
		if err != nil || got.Outcome != "existing" || got.Number != 7 || got.RequestID != "rid" || requests != 1 {
			t.Fatalf("got=%#v err=%v requests=%d", got, err, requests)
		}
	})
	t.Run("created uses GET then one POST", func(t *testing.T) {
		var methods []string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			methods = append(methods, r.Method)
			if r.Method == http.MethodGet {
				json.NewEncoder(w).Encode([]PullRequestResult{})
				return
			}
			var body map[string]string
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body["title"] != "Title" || body["body"] != "Body" || body["head"] != "feature/x" || body["base"] != "main" {
				t.Fatalf("body=%v", body)
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(PullRequestResult{Number: 8, URL: "https://github.com/acme/app/pull/8"})
		}))
		defer server.Close()
		got, err := OpenPullRequest(context.Background(), PullRequestRequest{RemoteURL: "git@github.com:acme/app.git", Base: "main", Head: "feature/x", Title: "Title", Body: "Body", Env: []string{"GITHUB_TOKEN=gh2"}, APIBaseURL: server.URL, HTTPClient: server.Client()})
		if err != nil || got.Outcome != "created" || got.Number != 8 || len(methods) != 2 || methods[1] != http.MethodPost {
			t.Fatalf("got=%#v err=%v methods=%v", got, err, methods)
		}
	})
}

func TestOpenPullRequestReconciles422AndMapsSafeErrors(t *testing.T) {
	posts := 0
	gets := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			gets++
			if gets == 1 {
				json.NewEncoder(w).Encode([]PullRequestResult{})
				return
			}
			json.NewEncoder(w).Encode([]PullRequestResult{{Number: 9, URL: "https://github.com/acme/app/pull/9"}})
			return
		}
		posts++
		w.WriteHeader(http.StatusUnprocessableEntity)
	}))
	defer server.Close()
	got, err := OpenPullRequest(context.Background(), PullRequestRequest{RemoteURL: "ssh://git@github.com/acme/app.git", Base: "main", Head: "feature/x", Title: "Title", Env: []string{"GH_TOKEN=gh"}, APIBaseURL: server.URL, HTTPClient: server.Client()})
	if err != nil || got.Outcome != "reconciled" || posts != 1 || gets != 2 {
		t.Fatalf("got=%#v err=%v posts=%d gets=%d", got, err, posts, gets)
	}

	_, err = OpenPullRequest(context.Background(), PullRequestRequest{RemoteURL: "https://github.com/acme/app", Base: "main", Head: "feature/x", Title: "Title", Env: []string{}, APIBaseURL: server.URL, HTTPClient: server.Client()})
	var failure Failure
	if !errors.As(err, &failure) || failure.Code != "authentication_unavailable" {
		t.Fatalf("err=%#v", err)
	}
}

func TestNormalizeRemoteRejectsUnsafeHostsAndCredentials(t *testing.T) {
	for _, raw := range []string{
		"https://token@github.com/acme/app.git",
		"https://github.enterprise/acme/app.git",
		"https://github.com/acme/app/issues",
		"https://github.com/acme/app.git?x=1",
		"git@gitlab.com:acme/app.git",
	} {
		if _, err := NormalizeRemote(raw); err == nil {
			t.Fatalf("unsafe remote accepted: %s", raw)
		}
	}
}
