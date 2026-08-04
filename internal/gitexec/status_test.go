package gitexec

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestStatusCollectorCountsLocalGitState(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("Git não está disponível")
	}
	repo := t.TempDir()
	gitOutput(t, repo, "init", "--quiet")
	gitOutput(t, repo, "config", "core.autocrlf", "false")
	gitOutput(t, repo, "config", "user.email", "test@example.com")
	gitOutput(t, repo, "config", "user.name", "Test")
	if err := os.WriteFile(filepath.Join(repo, "tracked.txt"), []byte("base\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitOutput(t, repo, "add", "tracked.txt")
	gitOutput(t, repo, "commit", "-m", "init")
	if err := os.WriteFile(filepath.Join(repo, "tracked.txt"), []byte("changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "staged.txt"), []byte("staged\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitOutput(t, repo, "add", "staged.txt")
	if err := os.WriteFile(filepath.Join(repo, "untracked.txt"), []byte("untracked\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	collect, err := FindStatus()
	if err != nil {
		t.Fatal(err)
	}
	got, err := collect(repo)
	if err != nil {
		t.Fatal(err)
	}
	if got.Commit == NoCommits || len(got.Commit) != 7 || got.Branch == DetachedHEAD {
		t.Fatalf("branch/commit = %#v", got)
	}
	if got.ModifiedCount != 1 || got.StagedCount != 1 || got.UntrackedCount != 1 {
		t.Fatalf("contagens = %#v", got)
	}
}

func TestStatusCollectorDetachedHeadNoCommitsAndOwnRoot(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("Git não está disponível")
	}
	empty := t.TempDir()
	gitOutput(t, empty, "init", "--quiet")
	collect, err := FindStatus()
	if err != nil {
		t.Fatal(err)
	}
	got, err := collect(empty)
	if err != nil {
		t.Fatal(err)
	}
	if got.Commit != NoCommits {
		t.Fatalf("commit = %q", got.Commit)
	}

	repo := t.TempDir()
	child := filepath.Join(repo, "child")
	if err := os.Mkdir(child, 0o755); err != nil {
		t.Fatal(err)
	}
	gitOutput(t, repo, "init", "--quiet")
	gitOutput(t, repo, "config", "core.autocrlf", "false")
	gitOutput(t, repo, "config", "user.email", "test@example.com")
	gitOutput(t, repo, "config", "user.name", "Test")
	if err := os.WriteFile(filepath.Join(repo, "file.txt"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitOutput(t, repo, "add", "file.txt")
	gitOutput(t, repo, "commit", "-m", "init")
	gitOutput(t, repo, "checkout", "--detach")
	got, err = collect(repo)
	if err != nil {
		t.Fatal(err)
	}
	if got.Branch != DetachedHEAD {
		t.Fatalf("branch = %q", got.Branch)
	}
	if _, err := collect(child); err == nil {
		t.Fatal("subdiretório versionado por ancestral aceito como raiz própria")
	}
}

func TestStatusCollectorSanitizesEnvironmentAndUsesReadOnlyCommands(t *testing.T) {
	dir := t.TempDir()
	fake := filepath.Join(dir, "git")
	if runtime.GOOS == "windows" {
		fake += ".bat"
	}
	writeFakeStatusGit(t, fake)
	t.Setenv("PATH", dir)
	t.Setenv("GIT_DIR", filepath.Join(dir, "bad.git"))
	t.Setenv("GIT_WORK_TREE", filepath.Join(dir, "bad-worktree"))

	collect, err := FindStatus()
	if err != nil {
		t.Fatal(err)
	}
	got, err := collect(filepath.Join(dir, "repo"))
	if err != nil {
		t.Fatal(err)
	}
	if got.Branch != "main" || got.Commit != "abc1234" ||
		got.ModifiedCount != 1 || got.StagedCount != 1 || got.UntrackedCount != 1 {
		t.Fatalf("status fake = %#v", got)
	}
}

func writeFakeStatusGit(t *testing.T, path string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		script := strings.Join([]string{
			"@echo off",
			"if not \"%GIT_DIR%\"==\"\" exit /b 20",
			"if not \"%GIT_OPTIONAL_LOCKS%\"==\"0\" exit /b 21",
			"if not \"%GIT_TERMINAL_PROMPT%\"==\"0\" exit /b 22",
			"if not \"%1\"==\"-C\" exit /b 3",
			"if \"%3\"==\"rev-parse\" if \"%4\"==\"--show-toplevel\" echo %2&& exit /b 0",
			"if \"%3\"==\"rev-parse\" if \"%4\"==\"--git-common-dir\" echo %2\\.git&& exit /b 0",
			"if \"%3\"==\"rev-parse\" if \"%4\"==\"--verify\" echo abc1234&& exit /b 0",
			"if \"%3\"==\"symbolic-ref\" if \"%4\"==\"--quiet\" if \"%5\"==\"--short\" if \"%6\"==\"HEAD\" echo main&& exit /b 0",
			"if \"%3\"==\"diff\" if \"%4\"==\"--name-only\" echo modified.txt&& exit /b 0",
			"if \"%3\"==\"diff\" if \"%4\"==\"--cached\" if \"%5\"==\"--name-only\" echo staged.txt&& exit /b 0",
			"if \"%3\"==\"ls-files\" if \"%4\"==\"--others\" if \"%5\"==\"--exclude-standard\" echo untracked.txt&& exit /b 0",
			"exit /b 9",
			"",
		}, "\r\n")
		if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
			t.Fatal(err)
		}
		return
	}
	script := `#!/bin/sh
[ -z "$GIT_DIR" ] || exit 20
[ "$GIT_OPTIONAL_LOCKS" = "0" ] || exit 21
[ "$GIT_TERMINAL_PROMPT" = "0" ] || exit 22
[ "$1" = "-C" ] || exit 3
case "$3 $4 $5 $6" in
  "rev-parse --show-toplevel  ") printf '%s\n' "$2" ;;
  "rev-parse --git-common-dir  ") printf '%s/.git\n' "$2" ;;
  "rev-parse --verify --short=7 HEAD") printf 'abc1234\n' ;;
  "symbolic-ref --quiet --short HEAD") printf 'main\n' ;;
  "diff --name-only  ") printf 'modified.txt\n' ;;
  "diff --cached --name-only ") printf 'staged.txt\n' ;;
  "ls-files --others --exclude-standard ") printf 'untracked.txt\n' ;;
  *) exit 9 ;;
esac
`
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
}
