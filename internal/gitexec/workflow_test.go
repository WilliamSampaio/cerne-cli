package gitexec

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestWorkflowInspectReportsLocalFactsAndSanitizesRemotes(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("Git não está disponível")
	}
	repo := t.TempDir()
	mustGitInspect(t, repo, "init", "--quiet", "--initial-branch", "main")
	mustGitInspect(t, repo, "config", "user.email", "test@example.com")
	mustGitInspect(t, repo, "config", "user.name", "Test")
	writeWorkflowFile(t, filepath.Join(repo, "tracked.txt"), "one\n")
	mustGitInspect(t, repo, "add", "tracked.txt")
	mustGitInspect(t, repo, "commit", "-m", "init")
	mustGitInspect(t, repo, "branch", "feature")
	mustGitInspect(t, repo, "remote", "add", "origin", "https://token@github.com/example/project.git")
	mustGitInspect(t, repo, "update-ref", "refs/remotes/origin/main", "HEAD")
	mustGitInspect(t, repo, "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/main")
	writeWorkflowFile(t, filepath.Join(repo, "tracked.txt"), "two\n")
	writeWorkflowFile(t, filepath.Join(repo, "new file.txt"), "new\n")

	inspect, err := FindWorkflowInspector()
	if err != nil {
		t.Fatal(err)
	}
	got, err := inspect(repo)
	if err != nil {
		t.Fatal(err)
	}
	if got.Branch != "main" || len(got.Head) != 40 || got.DefaultBranch != "main" || got.Clean {
		t.Fatalf("snapshot = %#v", got)
	}
	if !containsString(got.LocalBranches, "feature") || !containsString(got.RemoteBranches, "origin/main") {
		t.Fatalf("branches = %#v %#v", got.LocalBranches, got.RemoteBranches)
	}
	if len(got.Remotes) != 1 || got.Remotes[0].Name != "origin" || got.Remotes[0].Provider != "github" {
		t.Fatalf("remotes not sanitized: %#v", got.Remotes)
	}
	if len(got.Changes) != 2 || got.Changes[0].Digest == "" || got.Changes[1].Digest == "" {
		t.Fatalf("changes = %#v", got.Changes)
	}
	if got.Changes[0].Path != "new file.txt" || got.Changes[1].Path != "tracked.txt" {
		t.Fatalf("changes not deterministic/literal: %#v", got.Changes)
	}
}

func TestWorkflowInspectNoCommitAndGitEnvironment(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("Git não está disponível")
	}
	repo := t.TempDir()
	mustGitInspect(t, repo, "init", "--quiet")
	t.Setenv("GIT_DIR", filepath.Join(t.TempDir(), "bad.git"))
	t.Setenv("GIT_WORK_TREE", filepath.Join(t.TempDir(), "bad-worktree"))

	inspect, err := FindWorkflowInspector()
	if err != nil {
		t.Fatal(err)
	}
	got, err := inspect(repo)
	if err != nil {
		t.Fatal(err)
	}
	if got.Head != NoCommits || got.DefaultBranch != "main" {
		t.Fatalf("empty repo = %#v", got)
	}
}

func TestWorkflowInspectDisablesConfiguredFSMonitor(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("Git não está disponível")
	}
	repo := t.TempDir()
	mustGitInspect(t, repo, "init", "--quiet", "--initial-branch", "main")
	mustGitInspect(t, repo, "config", "user.email", "test@example.com")
	mustGitInspect(t, repo, "config", "user.name", "Test")
	writeWorkflowFile(t, filepath.Join(repo, "tracked.txt"), "one\n")
	mustGitInspect(t, repo, "add", "tracked.txt")
	mustGitInspect(t, repo, "commit", "-m", "init")
	hook := filepath.Join(t.TempDir(), "fsmonitor")
	sentinel := hook + ".ran"
	if runtime.GOOS == "windows" {
		hook += ".bat"
		sentinel = hook + ".ran"
		writeWorkflowFile(t, hook, "@echo off\r\necho ran > \"%~f0.ran\"\r\nexit /b 0\r\n")
	} else {
		writeWorkflowFile(t, hook, "#!/bin/sh\nprintf ran > \"$0.ran\"\n")
	}
	if err := os.Chmod(hook, 0o755); err != nil {
		t.Fatal(err)
	}
	mustGitInspect(t, repo, "config", "core.fsmonitor", hook)

	inspect, err := FindWorkflowInspector()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := inspect(repo); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(sentinel); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("fsmonitor configurado foi executado: %v", err)
	}
}

func TestWorkflowInspectDefaultBranchPrefersUpstreamRemoteHead(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("Git não está disponível")
	}
	repo := t.TempDir()
	mustGitInspect(t, repo, "init", "--quiet", "--initial-branch", "main")
	mustGitInspect(t, repo, "config", "user.email", "test@example.com")
	mustGitInspect(t, repo, "config", "user.name", "Test")
	writeWorkflowFile(t, filepath.Join(repo, "tracked.txt"), "one\n")
	mustGitInspect(t, repo, "add", "tracked.txt")
	mustGitInspect(t, repo, "commit", "-m", "init")
	mustGitInspect(t, repo, "remote", "add", "upstream", "git@github.com:example/project.git")
	mustGitInspect(t, repo, "update-ref", "refs/remotes/upstream/trunk", "HEAD")
	mustGitInspect(t, repo, "symbolic-ref", "refs/remotes/upstream/HEAD", "refs/remotes/upstream/trunk")
	mustGitInspect(t, repo, "branch", "--set-upstream-to=upstream/trunk", "main")

	inspect, err := FindWorkflowInspector()
	if err != nil {
		t.Fatal(err)
	}
	got, err := inspect(repo)
	if err != nil {
		t.Fatal(err)
	}
	if got.DefaultBranch != "trunk" || got.Upstream == nil || got.Upstream.Remote != "upstream" {
		t.Fatalf("default/upstream = %#v", got)
	}
}

func TestWorkflowGitEnvironmentCleansHostileVariables(t *testing.T) {
	clean := strings.Join(gitEnvironment([]string{"GIT_DIR=bad", "GIT_WORK_TREE=bad", "PATH=/bin"}), "\n")
	if strings.Contains(clean, "GIT_DIR=") || strings.Contains(clean, "GIT_WORK_TREE=") {
		t.Fatalf("hostile Git environment kept:\n%s", clean)
	}
}

func writeWorkflowFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
