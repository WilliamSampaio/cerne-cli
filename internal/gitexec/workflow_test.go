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

func TestWorkflowBranchCreateValidatesAndUsesExactArgv(t *testing.T) {
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
	mustGitInspect(t, repo, "branch", "develop")

	if err := ValidateWorkflowBranchName("feat/work"); err != nil {
		t.Fatalf("valid branch rejected: %v", err)
	}
	if err := ValidateWorkflowBranchName("bad name"); err == nil {
		t.Fatal("invalid branch accepted")
	}
	if WorkflowOperationInProgress(repo) {
		t.Fatal("operation detected in clean repo")
	}
	brancher, err := FindWorkflowBrancher()
	if err != nil {
		t.Fatal(err)
	}
	if err := brancher(repo, "feat/work", "develop"); err != nil {
		t.Fatal(err)
	}
	got := strings.TrimSpace(mustGitInspectOutput(t, repo, "branch", "--show-current"))
	if got != "feat/work" {
		t.Fatalf("branch=%q", got)
	}
}

func TestWorkflowOperationInProgressDetectsMerge(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("Git não está disponível")
	}
	repo := t.TempDir()
	mustGitInspect(t, repo, "init", "--quiet", "--initial-branch", "main")
	if err := os.WriteFile(filepath.Join(repo, ".git", "MERGE_HEAD"), []byte(strings.Repeat("a", 40)+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !WorkflowOperationInProgress(repo) {
		t.Fatal("merge not detected")
	}
}

func TestWorkflowCommitLiteralPathsPreservesUnrelatedStageAndHooks(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("Git não está disponível")
	}
	repo := t.TempDir()
	mustGitInspect(t, repo, "init", "--quiet", "--initial-branch", "main")
	mustGitInspect(t, repo, "config", "user.email", "test@example.com")
	mustGitInspect(t, repo, "config", "user.name", "Test")
	writeWorkflowFile(t, filepath.Join(repo, "selected.txt"), "one\n")
	writeWorkflowFile(t, filepath.Join(repo, "other.txt"), "one\n")
	mustGitInspect(t, repo, "add", "-A")
	mustGitInspect(t, repo, "commit", "-m", "init")
	writeWorkflowFile(t, filepath.Join(repo, "selected.txt"), "two\n")
	writeWorkflowFile(t, filepath.Join(repo, "other.txt"), "two\n")
	mustGitInspect(t, repo, "add", "other.txt")

	committer, err := FindWorkflowCommitter()
	if err != nil {
		t.Fatal(err)
	}
	if err := committer(repo, "commit selected", []string{"selected.txt"}); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(mustGitInspectOutput(t, repo, "show", "--name-only", "--format=", "HEAD")); got != "selected.txt" {
		t.Fatalf("committed files=%q", got)
	}
	if got := strings.TrimSpace(mustGitInspectOutput(t, repo, "diff", "--cached", "--name-only")); got != "other.txt" {
		t.Fatalf("unrelated stage=%q", got)
	}

	writeWorkflowFile(t, filepath.Join(repo, "selected.txt"), "three\n")
	hook := filepath.Join(repo, ".git", "hooks", "pre-commit")
	writeWorkflowFile(t, hook, "#!/bin/sh\nexit 1\n")
	if err := os.Chmod(hook, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := committer(repo, "hook should fail", []string{"selected.txt"}); err == nil {
		t.Fatal("hook failure was ignored")
	}
}

func TestWorkflowCommitRejectsUnsafeInputsAndMissingIdentity(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("Git não está disponível")
	}
	for _, path := range []string{"", "../x", "/tmp/x", "-x", ":(glob)*"} {
		if ValidateWorkflowLiteralPath(path) {
			t.Fatalf("unsafe path accepted: %q", path)
		}
	}
	repo := t.TempDir()
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(t.TempDir(), "missing"))
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("EMAIL", "")
	mustGitInspect(t, repo, "init", "--quiet", "--initial-branch", "main")
	mustGitInspect(t, repo, "config", "user.useConfigOnly", "true")
	writeWorkflowFile(t, filepath.Join(repo, "file.txt"), "one\n")
	committer, err := FindWorkflowCommitter()
	if err != nil {
		t.Fatal(err)
	}
	if err := committer(repo, "message", []string{"file.txt"}); err == nil {
		t.Fatal("missing identity accepted")
	}
}

func TestWorkflowPushUsesExplicitNonForceRefspecAndNoPrompts(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("Git não está disponível")
	}
	remote := filepath.Join(t.TempDir(), "remote.git")
	mustGitInspect(t, t.TempDir(), "init", "--bare", "--quiet", remote)
	repo := t.TempDir()
	mustGitInspect(t, repo, "init", "--quiet", "--initial-branch", "main")
	mustGitInspect(t, repo, "config", "user.email", "test@example.com")
	mustGitInspect(t, repo, "config", "user.name", "Test")
	writeWorkflowFile(t, filepath.Join(repo, "file.txt"), "one\n")
	mustGitInspect(t, repo, "add", "-A")
	mustGitInspect(t, repo, "commit", "-m", "init")
	mustGitInspect(t, repo, "switch", "--create", "feat/push")
	mustGitInspect(t, repo, "remote", "add", "origin", remote)

	pusher, err := FindWorkflowPusher()
	if err != nil {
		t.Fatal(err)
	}
	if err := pusher(repo, "origin", "feat/push"); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(mustGitInspectOutput(t, repo, "rev-parse", "--verify", "origin/feat/push")); got == "" {
		t.Fatal("remote branch missing")
	}
	env := strings.Join(gitEnvironment(nil), "\n")
	for _, want := range []string{"GIT_TERMINAL_PROMPT=0", "GCM_INTERACTIVE=Never", "GIT_ASKPASS=", "SSH_ASKPASS="} {
		if !strings.Contains(env, want) {
			t.Fatalf("missing env %s in %q", want, env)
		}
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

func mustGitInspectOutput(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v: %s", strings.Join(args, " "), err, output)
	}
	return string(output)
}
