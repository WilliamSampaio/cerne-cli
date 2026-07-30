package gitexec

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindRequiresGit(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	if _, err := Find(); err == nil {
		t.Fatal("Find() deveria falhar sem Git no PATH")
	}
}

func TestInitCreatesIsolatedEmptyRepository(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("Git não está disponível")
	}

	initRepository, err := Find()
	if err != nil {
		t.Fatal(err)
	}

	repository := t.TempDir()
	redirected := filepath.Join(t.TempDir(), "redirected.git")
	t.Setenv("GIT_DIR", redirected)
	t.Setenv("GIT_WORK_TREE", t.TempDir())

	if err := initRepository(repository); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(repository, ".git")); err != nil {
		t.Fatalf("metadados Git locais não encontrados: %v", err)
	}
	if _, err := os.Stat(redirected); !os.IsNotExist(err) {
		t.Fatalf("GIT_DIR externo foi usado: %v", err)
	}
	if got := gitOutput(t, repository, "remote"); got != "" {
		t.Fatalf("remotos inesperados: %q", got)
	}
	if got := gitOutput(t, repository, "rev-list", "--all", "--count"); got != "0" {
		t.Fatalf("commits = %q, esperado 0", got)
	}
}

func gitOutput(t *testing.T, repository string, args ...string) string {
	t.Helper()
	command := exec.Command("git", append([]string{"-C", repository}, args...)...)
	command.Env = cleanEnvironment(os.Environ())
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v: %s", strings.Join(args, " "), err, output)
	}
	return strings.TrimSpace(string(output))
}
