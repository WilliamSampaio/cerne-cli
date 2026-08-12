package workspace

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/WilliamSampaio/cerne-cli/internal/gitexec"
)

func TestInspectGitSnapshotStateAndAudit(t *testing.T) {
	root := newGitWorkflowWorkspace(t)
	home := t.TempDir()
	inspector := fakeWorkflowInspector(map[string]gitexec.WorkflowRepository{
		filepath.Join(root, "knowledge"): fakeWorkflowRepo(filepath.Join(root, "knowledge"), "main", "a"),
		filepath.Join(root, "source"):    fakeWorkflowRepo(filepath.Join(root, "source"), "main", "b"),
	}, nil)

	first, err := InspectGit(root, GitInspectRequest{Agent: "codex", TaskID: "task-1", Home: home}, inspector)
	if err != nil {
		t.Fatal(err)
	}
	second, err := InspectGit(root, GitInspectRequest{Agent: "codex", TaskID: "task-1", Home: home}, inspector)
	if err != nil {
		t.Fatal(err)
	}
	if first.Status != "healthy" || first.SchemaVersion != 1 || first.StateID == "" || first.StateID != second.StateID {
		t.Fatalf("snapshots = %#v %#v", first, second)
	}
	if len(first.Repositories) != 2 || first.Repositories[0].Name != "knowledge" || first.Repositories[1].Name != "source" {
		t.Fatalf("repositories = %#v", first.Repositories)
	}
	audit := readGitAudit(t, home, first.AuditID)
	if audit.Operation != "inspect" || audit.Authorization != "not-required" || audit.Status != "succeeded" ||
		audit.StateID != first.StateID || len(audit.Targets) != 2 {
		t.Fatalf("audit = %#v", audit)
	}

	changedInspector := fakeWorkflowInspector(map[string]gitexec.WorkflowRepository{
		filepath.Join(root, "knowledge"): fakeWorkflowRepo(filepath.Join(root, "knowledge"), "main", "changed"),
		filepath.Join(root, "source"):    fakeWorkflowRepo(filepath.Join(root, "source"), "main", "b"),
	}, nil)
	changed, err := InspectGit(root, GitInspectRequest{Agent: "codex", TaskID: "task-1", Home: home}, changedInspector)
	if err != nil {
		t.Fatal(err)
	}
	if changed.StateID == first.StateID {
		t.Fatal("state_id did not change")
	}
}

func TestInspectGitInvalidAndAuditFailures(t *testing.T) {
	root := newGitWorkflowWorkspace(t)
	t.Run("invalid usage", func(t *testing.T) {
		_, err := InspectGit(root, GitInspectRequest{Agent: "Codex", TaskID: "bad task", Home: t.TempDir()}, fakeWorkflowInspector(nil, nil))
		var failure GitFailure
		if !errors.As(err, &failure) || failure.Code != "validation_failed" {
			t.Fatalf("err = %#v", err)
		}
	})
	t.Run("repository problem is deterministic", func(t *testing.T) {
		failures := map[string]error{
			filepath.Join(root, "source"): errors.New("not git"),
		}
		snapshot, err := InspectGit(root, GitInspectRequest{Agent: "codex", TaskID: "task", Home: t.TempDir()}, fakeWorkflowInspector(nil, failures))
		if err != nil {
			t.Fatal(err)
		}
		if snapshot.Status != "invalid" || len(snapshot.Problems) != 1 || snapshot.Problems[0].Component != "source" {
			t.Fatalf("snapshot = %#v", snapshot)
		}
	})
	t.Run("audit creation failure before git query", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("symlink audit check is covered on Unix")
		}
		home := t.TempDir()
		if err := os.Symlink(t.TempDir(), filepath.Join(home, ".cerne")); err != nil {
			t.Skip(err)
		}
		queries := 0
		_, err := InspectGit(root, GitInspectRequest{Agent: "codex", TaskID: "task", Home: home}, func(string) (gitexec.WorkflowRepository, error) {
			queries++
			return gitexec.WorkflowRepository{}, nil
		})
		var failure GitFailure
		if !errors.As(err, &failure) || failure.Code != "audit_unavailable" || queries != 0 {
			t.Fatalf("err=%#v queries=%d", err, queries)
		}
	})
	t.Run("audit finalization failure preserves started record", func(t *testing.T) {
		home := t.TempDir()
		calls := 0
		original := replaceGitAudit
		replaceGitAudit = func(temp, target string) error {
			calls++
			if calls == 2 {
				return errors.New("disk full")
			}
			return os.Rename(temp, target)
		}
		t.Cleanup(func() { replaceGitAudit = original })
		snapshot, err := InspectGit(root, GitInspectRequest{Agent: "codex", TaskID: "task", Home: home}, fakeWorkflowInspector(map[string]gitexec.WorkflowRepository{
			filepath.Join(root, "knowledge"): fakeWorkflowRepo(filepath.Join(root, "knowledge"), "main", "a"),
			filepath.Join(root, "source"):    fakeWorkflowRepo(filepath.Join(root, "source"), "main", "b"),
		}, nil))
		var failure GitFailure
		if !errors.As(err, &failure) || failure.Code != "audit_incomplete" || snapshot.AuditID == "" {
			t.Fatalf("snapshot=%#v err=%#v", snapshot, err)
		}
		audit := readGitAudit(t, home, snapshot.AuditID)
		if audit.Status != "started" || audit.FinishedAt != "" {
			t.Fatalf("audit should preserve durable start: %#v", audit)
		}
	})
}

func TestInspectGitAuditPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX mode check")
	}
	root := newGitWorkflowWorkspace(t)
	home := t.TempDir()
	snapshot, err := InspectGit(root, GitInspectRequest{Agent: "codex", TaskID: "task", Home: home}, fakeWorkflowInspector(map[string]gitexec.WorkflowRepository{
		filepath.Join(root, "knowledge"): fakeWorkflowRepo(filepath.Join(root, "knowledge"), "main", "a"),
		filepath.Join(root, "source"):    fakeWorkflowRepo(filepath.Join(root, "source"), "main", "b"),
	}, nil))
	if err != nil {
		t.Fatal(err)
	}
	for path, want := range map[string]os.FileMode{
		filepath.Join(home, ".cerne"):                                           0o700,
		filepath.Join(home, ".cerne", "audit"):                                  0o700,
		filepath.Join(home, ".cerne", "audit", "git-"+snapshot.AuditID+".json"): 0o600,
	} {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if got := info.Mode().Perm(); got != want {
			t.Fatalf("%s mode=%o want=%o", path, got, want)
		}
	}
}

func newGitWorkflowWorkspace(t *testing.T) string {
	t.Helper()
	root := newDoctorWorkspace(t, "example")
	for _, dir := range []string{"knowledge", "source"} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func fakeWorkflowRepo(path, branch, headSeed string) gitexec.WorkflowRepository {
	return gitexec.WorkflowRepository{
		Path:          canonical(path),
		Branch:        branch,
		Head:          strings.Repeat(headSeed, 40)[:40],
		DefaultBranch: "main",
		Clean:         true,
		LocalBranches: []string{"main"},
		Remotes:       []gitexec.WorkflowRemote{{Name: "origin", Provider: "github"}},
	}
}

func fakeWorkflowInspector(repos map[string]gitexec.WorkflowRepository, failures map[string]error) gitexec.WorkflowInspector {
	normalizedRepos := make(map[string]gitexec.WorkflowRepository, len(repos))
	for path, repo := range repos {
		normalizedRepos[canonical(path)] = repo
	}
	normalizedFailures := make(map[string]error, len(failures))
	for path, err := range failures {
		normalizedFailures[canonical(path)] = err
	}
	return func(path string) (gitexec.WorkflowRepository, error) {
		path = canonical(path)
		if err, ok := normalizedFailures[path]; ok {
			return gitexec.WorkflowRepository{}, err
		}
		if repo, ok := normalizedRepos[path]; ok {
			return repo, nil
		}
		return fakeWorkflowRepo(path, "main", "c"), nil
	}
}

func readGitAudit(t *testing.T, home, id string) gitAuditRecord {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(home, ".cerne", "audit", "git-"+id+".json"))
	if err != nil {
		t.Fatal(err)
	}
	var record gitAuditRecord
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatal(err)
	}
	return record
}
