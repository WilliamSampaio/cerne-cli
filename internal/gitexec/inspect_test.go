package gitexec

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestInspectRepositoryOwnRootAndCommonDir(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("Git não está disponível")
	}
	repo := t.TempDir()
	mustGitInspect(t, repo, "init", "--quiet")

	inspect, err := FindInspector()
	if err != nil {
		t.Fatal(err)
	}
	got, err := inspect(repo)
	if err != nil {
		t.Fatal(err)
	}
	if !sameInspectPath(got.RequestedRoot, repo) || !sameInspectPath(got.WorktreeRoot, repo) {
		t.Fatalf("raízes = %#v, esperado %s", got, repo)
	}
	if got.CommonDir == "" {
		t.Fatal("common dir vazio")
	}
}

func TestInspectDetectsAncestorRepository(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("Git não está disponível")
	}
	repo := t.TempDir()
	child := filepath.Join(repo, "child")
	if err := os.Mkdir(child, 0o755); err != nil {
		t.Fatal(err)
	}
	mustGitInspect(t, repo, "init", "--quiet")

	inspect, err := FindInspector()
	if err != nil {
		t.Fatal(err)
	}
	got, err := inspect(child)
	if err != nil {
		t.Fatal(err)
	}
	if sameInspectPath(got.WorktreeRoot, child) || !sameInspectPath(got.WorktreeRoot, repo) {
		t.Fatalf("worktree root = %#v", got)
	}
}

func TestInspectSharedCommonDir(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("Git não está disponível")
	}
	parent := t.TempDir()
	repo := filepath.Join(parent, "repo")
	worktree := filepath.Join(parent, "worktree")
	if err := os.Mkdir(repo, 0o755); err != nil {
		t.Fatal(err)
	}
	mustGitInspect(t, repo, "init", "--quiet")
	mustGitInspect(t, repo, "config", "user.email", "test@example.com")
	mustGitInspect(t, repo, "config", "user.name", "Test")
	if err := os.WriteFile(filepath.Join(repo, "file"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	mustGitInspect(t, repo, "add", "file")
	mustGitInspect(t, repo, "commit", "-m", "init")
	mustGitInspect(t, repo, "worktree", "add", "--quiet", worktree)

	inspect, err := FindInspector()
	if err != nil {
		t.Fatal(err)
	}
	left, err := inspect(repo)
	if err != nil {
		t.Fatal(err)
	}
	right, err := inspect(worktree)
	if err != nil {
		t.Fatal(err)
	}
	if !sameInspectPath(left.CommonDir, right.CommonDir) {
		t.Fatalf("common dirs deveriam ser compartilhados: %#v %#v", left, right)
	}
}

func TestInspectSanitizesHostileGitEnvironmentAndUsesOnlyRevParse(t *testing.T) {
	dir := t.TempDir()
	fake := filepath.Join(dir, "git")
	if runtime.GOOS == "windows" {
		fake += ".bat"
	}
	writeFakeGit(t, fake)
	t.Setenv("PATH", dir)
	t.Setenv("GIT_DIR", filepath.Join(dir, "bad.git"))
	t.Setenv("GIT_WORK_TREE", filepath.Join(dir, "bad-worktree"))

	inspect, err := FindInspector()
	if err != nil {
		t.Fatal(err)
	}
	got, err := inspect(filepath.Join(dir, "repo"))
	if err != nil {
		t.Fatal(err)
	}
	if got.WorktreeRoot == "" || got.CommonDir == "" {
		t.Fatalf("resultado incompleto: %#v", got)
	}
}

func writeFakeGit(t *testing.T, path string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		script := "@echo off\r\nif not \"%1\"==\"-C\" exit /b 3\r\nif not \"%3\"==\"rev-parse\" exit /b 4\r\nif \"%4\"==\"--show-toplevel\" echo %2\r\nif \"%4\"==\"--git-common-dir\" echo %2\\.git\r\nif not \"%4\"==\"--show-toplevel\" if not \"%4\"==\"--git-common-dir\" exit /b 5\r\n"
		if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
			t.Fatal(err)
		}
		return
	}
	script := "#!/bin/sh\n[ \"$1\" = \"-C\" ] || exit 3\n[ \"$3\" = \"rev-parse\" ] || exit 4\ncase \"$4\" in\n  --show-toplevel) printf '%s\\n' \"$2\" ;;\n  --git-common-dir) printf '%s/.git\\n' \"$2\" ;;\n  *) exit 5 ;;\nesac\n"
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
}

func mustGitInspect(t *testing.T, repository string, args ...string) {
	t.Helper()
	command := exec.Command("git", append([]string{"-C", repository}, args...)...)
	command.Env = gitEnvironment(os.Environ())
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v: %s", strings.Join(args, " "), err, output)
	}
}

func sameInspectPath(left, right string) bool {
	if resolved, err := filepath.EvalSymlinks(left); err == nil {
		left = resolved
	}
	if resolved, err := filepath.EvalSymlinks(right); err == nil {
		right = resolved
	}
	return strings.EqualFold(filepath.Clean(left), filepath.Clean(right))
}
