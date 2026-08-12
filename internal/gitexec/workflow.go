package gitexec

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"sort"
	"strings"
)

type WorkflowInspector func(string) (WorkflowRepository, error)

type WorkflowRepository struct {
	Path           string
	Branch         string
	Head           string
	DefaultBranch  string
	Clean          bool
	Changes        []WorkflowChange
	LocalBranches  []string
	RemoteBranches []string
	Upstream       *WorkflowUpstream
	Remotes        []WorkflowRemote
}

type WorkflowChange struct {
	Path     string
	Index    string
	Worktree string
	Digest   string
}

type WorkflowUpstream struct {
	Remote string `json:"remote"`
	Branch string `json:"branch"`
	Ahead  int    `json:"ahead"`
	Behind int    `json:"behind"`
}

type WorkflowRemote struct {
	Name     string `json:"name"`
	Provider string `json:"provider"`
}

func FindWorkflowInspector() (WorkflowInspector, error) {
	git, err := exec.LookPath("git")
	if err != nil {
		return nil, fmt.Errorf("Git não encontrado no PATH: %w", err)
	}
	return func(directory string) (WorkflowRepository, error) {
		return inspectWorkflowRepository(git, directory)
	}, nil
}

func inspectWorkflowRepository(git, directory string) (WorkflowRepository, error) {
	root, err := statusOutput(git, directory, "rev-parse", "--show-toplevel")
	if err != nil {
		return WorkflowRepository{}, err
	}
	if cleanPath(root) != cleanPath(directory) {
		return WorkflowRepository{}, fmt.Errorf("%q não é raiz Git própria", directory)
	}
	branch, err := statusOutput(git, directory, "symbolic-ref", "--quiet", "--short", "HEAD")
	if err != nil {
		branch = DetachedHEAD
	}
	head, err := statusOutput(git, directory, "rev-parse", "--verify", "HEAD")
	if err != nil {
		head = NoCommits
	}
	changes, err := workflowChanges(git, directory)
	if err != nil {
		return WorkflowRepository{}, err
	}
	local, err := refNames(git, directory, "refs/heads")
	if err != nil {
		return WorkflowRepository{}, err
	}
	remoteBranches, err := refNames(git, directory, "refs/remotes")
	if err != nil {
		return WorkflowRepository{}, err
	}
	upstream := workflowUpstream(git, directory)
	remotes, err := workflowRemotes(git, directory)
	if err != nil {
		return WorkflowRepository{}, err
	}
	return WorkflowRepository{
		Path:           cleanPath(root),
		Branch:         branch,
		Head:           head,
		DefaultBranch:  defaultBranch(git, directory, upstream),
		Clean:          len(changes) == 0,
		Changes:        changes,
		LocalBranches:  local,
		RemoteBranches: remoteBranches,
		Upstream:       upstream,
		Remotes:        remotes,
	}, nil
}

func workflowChanges(git, directory string) ([]WorkflowChange, error) {
	output, err := workflowReadOnlyOutput(git, directory, "status", "--porcelain=v1", "-z", "--untracked-files=all")
	if err != nil {
		return nil, err
	}
	if output == "" {
		return nil, nil
	}
	parts := strings.Split(output, "\x00")
	var changes []WorkflowChange
	for i := 0; i < len(parts); i++ {
		item := parts[i]
		if item == "" {
			continue
		}
		if len(item) < 4 {
			continue
		}
		path := item[3:]
		if item[0] == 'R' || item[0] == 'C' {
			i++
		}
		changes = append(changes, WorkflowChange{
			Path:     path,
			Index:    string(item[0]),
			Worktree: string(item[1]),
			Digest:   digestStrings(item[:2], path),
		})
	}
	sort.Slice(changes, func(i, j int) bool { return changes[i].Path < changes[j].Path })
	return changes, nil
}

func workflowReadOnlyOutput(git, directory string, args ...string) (string, error) {
	command := exec.Command(git, append(readOnlyGitArgs(directory), args...)...)
	command.Env = gitEnvironment(os.Environ())
	output, err := command.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("consulta Git falhou em %q: %w", directory, err)
	}
	return string(output), nil
}

func refNames(git, directory, prefix string) ([]string, error) {
	output, err := statusOutput(git, directory, "for-each-ref", "--format=%(refname:short)", prefix)
	if err != nil {
		return nil, err
	}
	return sortedLines(output), nil
}

func workflowUpstream(git, directory string) *WorkflowUpstream {
	name, err := statusOutput(git, directory, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{upstream}")
	if err != nil || !strings.Contains(name, "/") {
		return nil
	}
	remote, branch, _ := strings.Cut(name, "/")
	left, _ := statusOutput(git, directory, "rev-list", "--count", "HEAD..@{upstream}")
	right, _ := statusOutput(git, directory, "rev-list", "--count", "@{upstream}..HEAD")
	return &WorkflowUpstream{Remote: remote, Branch: branch, Behind: atoi(left), Ahead: atoi(right)}
}

func workflowRemotes(git, directory string) ([]WorkflowRemote, error) {
	names, err := statusOutput(git, directory, "remote")
	if err != nil {
		return nil, err
	}
	var remotes []WorkflowRemote
	for _, name := range sortedLines(names) {
		raw, _ := statusOutput(git, directory, "remote", "get-url", "--push", name)
		remotes = append(remotes, WorkflowRemote{Name: name, Provider: remoteProvider(raw)})
	}
	return remotes, nil
}

func defaultBranch(git, directory string, upstream *WorkflowUpstream) string {
	if upstream != nil {
		if branch := remoteHead(git, directory, upstream.Remote); branch != "" {
			return branch
		}
	}
	if branch := remoteHead(git, directory, "origin"); branch != "" {
		return branch
	}
	return "main"
}

func remoteHead(git, directory, remote string) string {
	ref, err := statusOutput(git, directory, "symbolic-ref", "--quiet", "--short", "refs/remotes/"+remote+"/HEAD")
	if err != nil {
		return ""
	}
	_, branch, ok := strings.Cut(ref, "/")
	if !ok || branch == "" {
		return ""
	}
	return branch
}

func remoteProvider(raw string) string {
	host := ""
	if strings.HasPrefix(raw, "git@") {
		rest := strings.TrimPrefix(raw, "git@")
		host, _, _ = strings.Cut(rest, ":")
	} else if parsed, err := url.Parse(raw); err == nil {
		host = parsed.Hostname()
	}
	if strings.EqualFold(host, "github.com") {
		return "github"
	}
	return "other"
}

func sortedLines(output string) []string {
	if strings.TrimSpace(output) == "" {
		return nil
	}
	lines := strings.Split(strings.TrimRight(output, "\r\n"), "\n")
	sort.Strings(lines)
	return lines
}

func digestStrings(values ...string) string {
	h := sha256.New()
	for _, value := range values {
		h.Write([]byte(value))
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))
}

func atoi(value string) int {
	var n int
	for _, r := range strings.TrimSpace(value) {
		if r < '0' || r > '9' {
			return 0
		}
		n = n*10 + int(r-'0')
	}
	return n
}
