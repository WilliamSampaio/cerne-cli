package gitexec

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestLinkInspectorDetectsNonBareBareAndWorktree(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("Git não está disponível")
	}
	parent := t.TempDir()
	repo := filepath.Join(parent, "repo")
	worktree := filepath.Join(parent, "worktree")
	bare := filepath.Join(parent, "bare.git")
	if err := os.Mkdir(repo, 0o755); err != nil {
		t.Fatal(err)
	}
	mustGitInspect(t, repo, "init", "--quiet")
	mustGitInspect(t, repo, "config", "core.autocrlf", "false")
	mustGitInspect(t, repo, "config", "user.email", "test@example.com")
	mustGitInspect(t, repo, "config", "user.name", "Test")
	if err := os.WriteFile(filepath.Join(repo, "file.txt"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mustGitInspect(t, repo, "add", "file.txt")
	mustGitInspect(t, repo, "commit", "-m", "init")
	mustGitInspect(t, repo, "worktree", "add", "--quiet", worktree)
	if output, err := exec.Command("git", "init", "--bare", "--quiet", bare).CombinedOutput(); err != nil {
		t.Fatalf("git init --bare: %v: %s", err, output)
	}

	inspect, err := FindLinkInspector()
	if err != nil {
		t.Fatal(err)
	}
	got, err := inspect(repo)
	if err != nil {
		t.Fatal(err)
	}
	if got.IsBare || !got.HasWorktree || !sameInspectPath(got.WorktreeRoot, repo) || got.CommonDir == "" {
		t.Fatalf("repo comum = %#v", got)
	}
	got, err = inspect(worktree)
	if err != nil {
		t.Fatal(err)
	}
	if got.IsBare || !got.HasWorktree || !sameInspectPath(got.WorktreeRoot, worktree) || got.CommonDir == "" {
		t.Fatalf("worktree = %#v", got)
	}
	got, err = inspect(bare)
	if err != nil {
		t.Fatal(err)
	}
	if !got.IsBare || got.HasWorktree {
		t.Fatalf("bare = %#v", got)
	}
}

func TestLinkInspectorSanitizesEnvironmentAndUsesOnlyReadOnlyRevParse(t *testing.T) {
	dir := t.TempDir()
	fake := filepath.Join(dir, "git")
	if runtime.GOOS == "windows" {
		fake += ".bat"
	}
	writeFakeLinkGit(t, fake)
	t.Setenv("PATH", dir)
	t.Setenv("GIT_DIR", filepath.Join(dir, "bad.git"))
	t.Setenv("GIT_WORK_TREE", filepath.Join(dir, "bad-worktree"))

	inspect, err := FindLinkInspector()
	if err != nil {
		t.Fatal(err)
	}
	got, err := inspect(filepath.Join(dir, "repo"))
	if err != nil {
		t.Fatal(err)
	}
	if got.IsBare || !got.HasWorktree || got.WorktreeRoot == "" || got.CommonDir == "" {
		t.Fatalf("resultado fake = %#v", got)
	}
}

func writeFakeLinkGit(t *testing.T, path string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		script := strings.Join([]string{
			"@echo off",
			"if not \"%GIT_DIR%\"==\"\" exit /b 20",
			"if not \"%GIT_OPTIONAL_LOCKS%\"==\"0\" exit /b 21",
			"if not \"%GIT_TERMINAL_PROMPT%\"==\"0\" exit /b 22",
			"if not \"%1\"==\"-C\" exit /b 3",
			"if \"%3\"==\"rev-parse\" if \"%4\"==\"--is-bare-repository\" echo false&& exit /b 0",
			"if \"%3\"==\"rev-parse\" if \"%4\"==\"--show-toplevel\" echo %2&& exit /b 0",
			"if \"%3\"==\"rev-parse\" if \"%4\"==\"--git-common-dir\" echo %2\\.git&& exit /b 0",
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
case "$3 $4" in
  "rev-parse --is-bare-repository") printf 'false\n' ;;
  "rev-parse --show-toplevel") printf '%s\n' "$2" ;;
  "rev-parse --git-common-dir") printf '%s/.git\n' "$2" ;;
  *) exit 9 ;;
esac
`
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
}
