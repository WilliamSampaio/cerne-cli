package gitexec

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestClassifyCloneOriginAllowsOnlySafeLocations(t *testing.T) {
	local := t.TempDir()
	allowed := map[string]string{
		local:                              "local",
		"file:///tmp/example.git":          "file",
		"https://example.com/org/repo.git": "https",
		"ssh://git@example.com/org/repo":   "ssh",
		"git@example.com:org/repo.git":     "ssh",
	}
	for input, transport := range allowed {
		t.Run(transport+input, func(t *testing.T) {
			got, err := ClassifyCloneOrigin(t.TempDir(), input)
			wantFingerprint := fmt.Sprintf("%x", sha256.Sum256([]byte(input)))
			if err != nil || got.Transport != transport || got.Fingerprint != wantFingerprint || got.Location == "" {
				t.Fatalf("origem=%#v erro=%v", got, err)
			}
		})
	}
	for _, input := range []string{
		"", "--upload-pack=evil", "http://example.com/repo", "git://example.com/repo",
		"ext::evil", "ftp://example.com/repo", "https://user@example.com/repo",
		"https://example.com/repo?token=secret", "https://example.com/repo#fragment",
		"ssh://user:password@example.com/repo", "missing-local-path",
		"user:password@example.com:repo", "directory/name:repo",
	} {
		t.Run("reject "+input, func(t *testing.T) {
			if _, err := ClassifyCloneOrigin(t.TempDir(), input); !errors.Is(err, ErrInvalidCloneOrigin) {
				t.Fatalf("origem %q aceita: %v", input, err)
			}
		})
	}
}

func TestCloneUsesFixedArgumentsAndSanitizedEnvironment(t *testing.T) {
	original := runCloneCommand
	t.Cleanup(func() { runCloneCommand = original })
	t.Setenv("GIT_DIR", "redirected-secret")
	t.Setenv("TOKEN_FOR_TEST", "secret-value")
	var executable string
	var arguments, environment []string
	runCloneCommand = func(git string, args, env []string) error {
		executable, arguments, environment = git, args, env
		return nil
	}
	destination := filepath.Join(t.TempDir(), "private staging")
	if err := cloneWithGit("absolute-git")("https://example.com/repo.git", destination); err != nil {
		t.Fatal(err)
	}
	if executable != "absolute-git" {
		t.Fatalf("executável=%q", executable)
	}
	wantPrefix := []string{
		"-c", "credential.interactive=false", "-c", "protocol.allow=never",
		"-c", "protocol.file.allow=always", "-c", "protocol.https.allow=always",
		"-c", "protocol.ssh.allow=always", "-c",
	}
	if len(arguments) != 20 || !reflect.DeepEqual(arguments[:11], wantPrefix) ||
		!strings.HasPrefix(arguments[11], "core.hooksPath=") ||
		!reflect.DeepEqual(arguments[12:18], []string{"clone", "--quiet", "--origin=origin", "--no-local", arguments[16], "--"}) ||
		!strings.HasPrefix(arguments[16], "--template=") || arguments[18] != "https://example.com/repo.git" || arguments[19] != destination {
		t.Fatalf("argumentos=%q", arguments)
	}
	joined := strings.Join(environment, "\n")
	for _, required := range []string{"GIT_TERMINAL_PROMPT=0", "GCM_INTERACTIVE=Never", "SSH_ASKPASS_REQUIRE=never"} {
		if !strings.Contains(joined, required) {
			t.Fatalf("ambiente sem %s: %q", required, environment)
		}
	}
	if strings.Contains(joined, "redirected-secret") {
		t.Fatalf("redirecionamento Git herdado: %q", environment)
	}
}

func TestCloneFailureDoesNotExposeGitOutput(t *testing.T) {
	original := runCloneCommand
	t.Cleanup(func() { runCloneCommand = original })
	runCloneCommand = func(string, []string, []string) error { return errors.New("token-super-secreto") }
	err := cloneWithGit("git")("https://example.com/repo.git", t.TempDir())
	if err == nil || strings.Contains(err.Error(), "token-super-secreto") || strings.Contains(err.Error(), "example.com") {
		t.Fatalf("erro inseguro=%v", err)
	}
}

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
