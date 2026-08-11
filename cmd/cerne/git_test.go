package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCLIGitInspectJSONAndAudit(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("Git não está disponível")
	}
	binary := buildCLI(t)
	root := newCLIGitWorkspace(t)
	home := t.TempDir()

	status, stdout, stderr := executeCLI(t, binary, root, skillHomeEnvironment(home), "git", "inspect", "--agent", "codex", "--task", "task-1", "--json")
	if status != 0 || stderr != "" || !strings.HasSuffix(stdout, "\n") {
		t.Fatalf("status=%d stdout=%q stderr=%q", status, stdout, stderr)
	}
	var snapshot struct {
		SchemaVersion int    `json:"schema_version"`
		Status        string `json:"status"`
		StateID       string `json:"state_id"`
		AuditID       string `json:"audit_id"`
		Repositories  []struct {
			Name    string `json:"name"`
			Head    string `json:"head"`
			Remotes []struct {
				Name     string `json:"name"`
				Provider string `json:"provider"`
			} `json:"remotes"`
		} `json:"repositories"`
	}
	if err := json.Unmarshal([]byte(stdout), &snapshot); err != nil {
		t.Fatal(err)
	}
	if snapshot.SchemaVersion != 1 || snapshot.Status != "healthy" || snapshot.StateID == "" ||
		snapshot.AuditID == "" || len(snapshot.Repositories) != 2 {
		t.Fatalf("snapshot=%#v", snapshot)
	}
	if strings.Contains(stdout, "github.com/example/private") || strings.Contains(stdout, "token") {
		t.Fatalf("remote leaked in stdout: %s", stdout)
	}
	audit := readTestFile(t, filepath.Join(home, ".cerne", "audit", "git-"+snapshot.AuditID+".json"))
	if !strings.Contains(audit, `"operation": "inspect"`) || !strings.Contains(audit, `"authorization": "not-required"`) ||
		strings.Contains(audit, root) || strings.Contains(audit, "github.com") {
		t.Fatalf("unsafe audit: %s", audit)
	}
}

func TestCLIGitInspectUsageHelpAndFailures(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("Git não está disponível")
	}
	binary := buildCLI(t)
	root := newCLIGitWorkspace(t)

	status, stdout, stderr := executeCLI(t, binary, root, nil, "git", "--help")
	if status != 0 || stderr != "" || !strings.Contains(stdout, "cerne git inspect --agent") || strings.Contains(stdout, "cerne git branch") {
		t.Fatalf("help: status=%d stdout=%q stderr=%q", status, stdout, stderr)
	}
	for _, args := range [][]string{
		{"git"},
		{"git", "inspect", "--agent", "codex", "--json"},
		{"git", "inspect", "--agent", "Codex", "--task", "task", "--json"},
		{"git", "inspect", "--agent", "codex", "--task", "bad task", "--json"},
		{"git", "merge"},
	} {
		status, stdout, stderr := executeCLI(t, binary, root, nil, args...)
		if status != 2 || stdout != "" || !strings.Contains(stderr, "cerne git inspect") {
			t.Fatalf("%v: status=%d stdout=%q stderr=%q", args, status, stdout, stderr)
		}
	}

	status, stdout, stderr = executeCLI(t, binary, t.TempDir(), skillHomeEnvironment(t.TempDir()), "git", "inspect", "--agent", "codex", "--task", "task", "--json")
	if status != 1 || stderr != "" || !strings.Contains(stdout, `"status": "invalid"`) {
		t.Fatalf("invalid workspace: status=%d stdout=%q stderr=%q", status, stdout, stderr)
	}

	if runtime.GOOS == "windows" {
		t.Skip("symlink audit check is covered on Unix")
	}
	home := t.TempDir()
	if err := os.Mkdir(filepath.Join(home, ".cerne"), 0o700); err != nil {
		t.Fatal(err)
	}
	auditTarget := t.TempDir()
	if err := os.Symlink(auditTarget, filepath.Join(home, ".cerne", "audit")); err != nil {
		t.Skip(err)
	}
	status, stdout, stderr = executeCLI(t, binary, root, skillHomeEnvironment(home), "git", "inspect", "--agent", "codex", "--task", "task", "--json")
	if status != 1 || stdout != "" || !strings.Contains(stderr, "audit_unavailable") {
		t.Fatalf("audit failure: status=%d stdout=%q stderr=%q", status, stdout, stderr)
	}
	entries, err := os.ReadDir(auditTarget)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("audit created unexpectedly: %v", entries)
	}
}

func TestCLIGitDelegatesEffectsWithoutAudit(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("Git não está disponível")
	}
	binary := buildCLI(t)
	root := newCLIGitWorkspace(t)
	state := cliGitState(t, binary, root, t.TempDir())

	for _, args := range [][]string{
		{"git", "branch", "create", "--name", "feat/x", "--base", "knowledge=main", "--base", "source=main", "--state", state, "--confirm", "--agent", "codex", "--task", "task-1", "--json"},
		{"git", "commit", "source", "--message", "feat: checkpoint", "--include", "file.txt", "--state", state, "--confirm", "--agent", "codex", "--task", "task-2", "--json"},
		{"git", "push", "source", "--remote", "origin", "--branch", "main", "--state", state, "--confirm", "--agent", "codex", "--task", "task-3", "--json"},
		{"git", "pr", "prepare", "source", "--remote", "origin", "--base", "main", "--head", "feat/x", "--title", "Title", "--body-file", "body.md", "--state", state, "--confirm", "--agent", "codex", "--task", "task-4", "--json"},
	} {
		home := t.TempDir()
		status, stdout, stderr := executeCLI(t, binary, root, skillHomeEnvironment(home), args...)
		if status != 2 || stdout != "" || !strings.Contains(stderr, "cerne git inspect") {
			t.Fatalf("%v: status=%d stdout=%q stderr=%q", args, status, stdout, stderr)
		}
		if entries, _ := os.ReadDir(filepath.Join(home, ".cerne", "audit")); len(entries) != 0 {
			t.Fatalf("%v created audit entries: %v", args, entries)
		}
	}
}

func TestCLIGitRefusesForbiddenOperationsWithoutAudit(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("Git não está disponível")
	}
	binary := buildCLI(t)
	root := newCLIGitWorkspace(t)
	home := t.TempDir()
	state := cliGitState(t, binary, root, home)
	for _, args := range [][]string{
		{"git", "merge", "main"},
		{"git", "rebase", "main"},
		{"git", "reset", "--hard"},
		{"git", "stash"},
		{"git", "clean", "-fd"},
		{"git", "commit", "source", "--amend", "--state", state, "--confirm", "--agent", "codex", "--task", "task", "--json"},
		{"git", "push", "source", "--remote", "origin", "--branch", "main", "--force", "--state", state, "--confirm", "--agent", "codex", "--task", "task", "--json"},
		{"git", "branch", "delete", "feat/x"},
		{"git", "pr", "merge", "source"},
	} {
		badHome := t.TempDir()
		status, stdout, stderr := executeCLI(t, binary, root, skillHomeEnvironment(badHome), args...)
		if status != 2 || stdout != "" || stderr == "" {
			t.Fatalf("%v: status=%d stdout=%q stderr=%q", args, status, stdout, stderr)
		}
		if entries, _ := os.ReadDir(filepath.Join(badHome, ".cerne", "audit")); len(entries) != 0 {
			t.Fatalf("%v created audit entries: %v", args, entries)
		}
	}
}

func newCLIGitWorkspace(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	knowledge := filepath.Join(root, "knowledge")
	source := filepath.Join(root, "source")
	if err := os.MkdirAll(knowledge, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(knowledge, "cerne.json"), []byte(`{"name":"example","source":"../source"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	initCLIGitRepo(t, knowledge)
	initCLIGitRepo(t, source)
	mustCLIGit(t, source, "remote", "add", "origin", "https://token@github.com/example/private.git")
	return root
}

func cliGitState(t *testing.T, binary, root, home string) string {
	t.Helper()
	status, stdout, stderr := executeCLI(t, binary, root, skillHomeEnvironment(home), "git", "inspect", "--agent", "codex", "--task", "state", "--json")
	if status != 0 || stderr != "" {
		t.Fatalf("inspect: status=%d stdout=%q stderr=%q", status, stdout, stderr)
	}
	var snapshot struct {
		StateID string `json:"state_id"`
	}
	if err := json.Unmarshal([]byte(stdout), &snapshot); err != nil {
		t.Fatal(err)
	}
	return snapshot.StateID
}

func initCLIGitRepo(t *testing.T, dir string) {
	t.Helper()
	mustCLIGit(t, dir, "init", "--quiet", "--initial-branch", "main")
	mustCLIGit(t, dir, "config", "user.email", "test@example.com")
	mustCLIGit(t, dir, "config", "user.name", "Test")
	if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("init\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mustCLIGit(t, dir, "add", "-A")
	mustCLIGit(t, dir, "commit", "-m", "init")
}

func mustCLIGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	_ = mustCLIGitOutput(t, dir, args...)
}

func mustCLIGitOutput(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v: %s", strings.Join(args, " "), err, output)
	}
	return string(output)
}
