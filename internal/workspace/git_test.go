package workspace

import (
	"context"
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

func TestCreateGitBranchValidationAndPreflight(t *testing.T) {
	originalOperation := workflowOperationInProgress
	workflowOperationInProgress = func(string) bool { return false }
	t.Cleanup(func() { workflowOperationInProgress = originalOperation })
	root := newGitWorkflowWorkspace(t)
	home := t.TempDir()
	state := branchState(t, root, map[string]gitexec.WorkflowRepository{
		filepath.Join(root, "knowledge"): fakeWorkflowRepo(filepath.Join(root, "knowledge"), "main", "a"),
		filepath.Join(root, "source"):    fakeWorkflowRepo(filepath.Join(root, "source"), "main", "b"),
	})
	t.Run("requires authorization fields", func(t *testing.T) {
		_, err := CreateGitBranch(root, GitBranchRequest{Agent: "codex", TaskID: "task", Home: home, Name: "feat/x", Bases: map[string]string{"knowledge": "main", "source": "main"}}, fakeWorkflowInspector(nil, nil), func(string, string, string) error { return nil })
		var failure GitFailure
		if !errors.As(err, &failure) || failure.Code != "validation_failed" {
			t.Fatalf("err = %#v", err)
		}
	})
	t.Run("blocks stale state before effects", func(t *testing.T) {
		calls := 0
		report, err := CreateGitBranch(root, GitBranchRequest{Agent: "codex", TaskID: "task", Home: home, StateID: "old", Name: "feat/x", Confirm: true, Bases: map[string]string{"knowledge": "main", "source": "main"}}, fakeWorkflowInspector(map[string]gitexec.WorkflowRepository{
			filepath.Join(root, "knowledge"): fakeWorkflowRepo(filepath.Join(root, "knowledge"), "main", "a"),
			filepath.Join(root, "source"):    fakeWorkflowRepo(filepath.Join(root, "source"), "main", "b"),
		}, nil), func(string, string, string) error {
			calls++
			return nil
		})
		if err != nil || report.Status != "blocked" || calls != 0 || report.Problems[0].Code != "stale_state" {
			t.Fatalf("report=%#v err=%#v calls=%d", report, err, calls)
		}
	})
	t.Run("blocks dirty detached missing base existing target before effects", func(t *testing.T) {
		repos := map[string]gitexec.WorkflowRepository{
			filepath.Join(root, "knowledge"): fakeWorkflowRepo(filepath.Join(root, "knowledge"), gitexec.DetachedHEAD, "a"),
			filepath.Join(root, "source"):    fakeWorkflowRepo(filepath.Join(root, "source"), "main", "b"),
		}
		source := repos[filepath.Join(root, "source")]
		source.Clean = false
		source.LocalBranches = []string{"main", "feat/x"}
		repos[filepath.Join(root, "source")] = source
		current := branchState(t, root, repos)
		calls := 0
		report, err := CreateGitBranch(root, GitBranchRequest{Agent: "codex", TaskID: "task", Home: home, StateID: current.StateID, Name: "feat/x", Confirm: true, Bases: map[string]string{"knowledge": "missing", "source": "main"}}, fakeWorkflowInspector(repos, nil), func(string, string, string) error {
			calls++
			return nil
		})
		if err != nil || report.Status != "blocked" || calls != 0 || len(report.Problems) != 2 {
			t.Fatalf("report=%#v err=%#v calls=%d", report, err, calls)
		}
	})
	t.Run("audit failure happens before inspection or branch effects", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("symlink audit check is covered on Unix")
		}
		badHome := t.TempDir()
		if err := os.Symlink(t.TempDir(), filepath.Join(badHome, ".cerne")); err != nil {
			t.Skip(err)
		}
		queries := 0
		_, err := CreateGitBranch(root, GitBranchRequest{Agent: "codex", TaskID: "task", Home: badHome, StateID: state.StateID, Name: "feat/x", Confirm: true, Bases: map[string]string{"knowledge": "main", "source": "main"}}, func(string) (gitexec.WorkflowRepository, error) {
			queries++
			return gitexec.WorkflowRepository{}, nil
		}, func(string, string, string) error {
			queries++
			return nil
		})
		var failure GitFailure
		if !errors.As(err, &failure) || failure.Code != "audit_unavailable" || queries != 0 {
			t.Fatalf("err=%#v queries=%d", err, queries)
		}
	})
}

func TestCreateGitBranchSuccessPartialAndAuditIncomplete(t *testing.T) {
	originalOperation := workflowOperationInProgress
	workflowOperationInProgress = func(string) bool { return false }
	t.Cleanup(func() { workflowOperationInProgress = originalOperation })
	root := newGitWorkflowWorkspace(t)
	repos := map[string]gitexec.WorkflowRepository{
		filepath.Join(root, "knowledge"): fakeWorkflowRepo(filepath.Join(root, "knowledge"), "main", "a"),
		filepath.Join(root, "source"):    fakeWorkflowRepo(filepath.Join(root, "source"), "develop", "b"),
	}
	source := repos[filepath.Join(root, "source")]
	source.LocalBranches = []string{"develop", "main"}
	repos[filepath.Join(root, "source")] = source
	state := branchState(t, root, repos)
	t.Run("creates from different bases in order", func(t *testing.T) {
		var calls []string
		report, err := CreateGitBranch(root, GitBranchRequest{Agent: "gemini", TaskID: "task", Home: t.TempDir(), StateID: state.StateID, Name: "feat/x", Confirm: true, Bases: map[string]string{"knowledge": "main", "source": "develop"}}, fakeWorkflowInspector(repos, nil), func(path, name, base string) error {
			calls = append(calls, filepath.Base(path)+":"+name+":"+base)
			return nil
		})
		if err != nil || report.Status != "succeeded" || strings.Join(calls, ",") != "knowledge:feat/x:main,source:feat/x:develop" || report.StateAfter == "" {
			t.Fatalf("report=%#v err=%#v calls=%v", report, err, calls)
		}
	})
	t.Run("partial failure does not rollback", func(t *testing.T) {
		var calls []string
		report, err := CreateGitBranch(root, GitBranchRequest{Agent: "codex", TaskID: "task", Home: t.TempDir(), StateID: state.StateID, Name: "feat/y", Confirm: true, Bases: map[string]string{"knowledge": "main", "source": "develop"}}, fakeWorkflowInspector(repos, nil), func(path, name, base string) error {
			calls = append(calls, filepath.Base(path))
			if filepath.Base(path) == "source" {
				return errors.New("boom")
			}
			return nil
		})
		if err != nil || report.Status != "partial" || strings.Join(calls, ",") != "knowledge,source" ||
			report.Repositories[0].Status != "succeeded" || report.Repositories[1].Status != "failed" {
			t.Fatalf("report=%#v err=%#v calls=%v", report, err, calls)
		}
	})
	t.Run("audit finalization failure reports incomplete after effects", func(t *testing.T) {
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
		report, err := CreateGitBranch(root, GitBranchRequest{Agent: "codex", TaskID: "task", Home: home, StateID: state.StateID, Name: "feat/z", Confirm: true, Bases: map[string]string{"knowledge": "main", "source": "develop"}}, fakeWorkflowInspector(repos, nil), func(string, string, string) error { return nil })
		var failure GitFailure
		if !errors.As(err, &failure) || failure.Code != "audit_incomplete" || report.Status != "succeeded" {
			t.Fatalf("report=%#v err=%#v", report, err)
		}
	})
}

func TestGitCommitPushAndPullRequestCoordinator(t *testing.T) {
	originalOperation := workflowOperationInProgress
	workflowOperationInProgress = func(string) bool { return false }
	t.Cleanup(func() { workflowOperationInProgress = originalOperation })
	root := newGitWorkflowWorkspace(t)
	knowledge := filepath.Join(root, "knowledge")
	source := filepath.Join(root, "source")
	repos := map[string]gitexec.WorkflowRepository{
		knowledge: fakeWorkflowRepo(knowledge, "main", "a"),
		source:    fakeWorkflowRepo(source, "feat/x", "b"),
	}
	sourceRepo := repos[source]
	sourceRepo.Changes = []gitexec.WorkflowChange{{Path: "changed.txt", Index: " ", Worktree: "M", Digest: strings.Repeat("d", 64)}}
	sourceRepo.Clean = false
	sourceRepo.LocalBranches = []string{"feat/x", "main"}
	sourceRepo.RemoteBranches = []string{"origin/feat/x"}
	sourceRepo.Remotes = []gitexec.WorkflowRemote{{Name: "origin", Provider: "github"}}
	repos[source] = sourceRepo
	state := branchState(t, root, repos)

	t.Run("commit uses selected repository and paths", func(t *testing.T) {
		var gotPath string
		report, err := CommitGit(root, GitCommitRequest{Agent: "codex", TaskID: "task", Home: t.TempDir(), StateID: state.StateID, Repository: "source", Message: "feat: change", Paths: []string{"changed.txt"}, Confirm: true}, fakeWorkflowInspector(repos, nil), func(path, message string, paths []string) error {
			gotPath = path
			if message != "feat: change" || strings.Join(paths, ",") != "changed.txt" {
				t.Fatalf("message=%q paths=%v", message, paths)
			}
			return nil
		})
		if err != nil || report.Status != "succeeded" || gotPath != source {
			t.Fatalf("report=%#v err=%#v path=%q", report, err, gotPath)
		}
	})
	t.Run("commit blocks unchanged paths before effects", func(t *testing.T) {
		calls := 0
		report, err := CommitGit(root, GitCommitRequest{Agent: "codex", TaskID: "task", Home: t.TempDir(), StateID: state.StateID, Repository: "source", Message: "feat: change", Paths: []string{"missing.txt"}, Confirm: true}, fakeWorkflowInspector(repos, nil), func(string, string, []string) error {
			calls++
			return nil
		})
		if err != nil || report.Status != "blocked" || report.ErrorCode != "path_not_changed" || calls != 0 {
			t.Fatalf("report=%#v err=%#v calls=%d", report, err, calls)
		}
	})
	t.Run("push validates explicit remote and branch", func(t *testing.T) {
		var got string
		report, err := PushGit(root, GitPushRequest{Agent: "claude", TaskID: "task", Home: t.TempDir(), StateID: state.StateID, Repository: "source", Remote: "origin", Branch: "feat/x", Confirm: true}, fakeWorkflowInspector(repos, nil), func(path, remote, branch string) error {
			got = filepath.Base(path) + ":" + remote + ":" + branch
			return nil
		})
		if err != nil || report.Status != "succeeded" || got != "source:origin:feat/x" {
			t.Fatalf("report=%#v err=%#v got=%q", report, err, got)
		}
	})
	t.Run("pull request opens only when head is published", func(t *testing.T) {
		report, err := CreateGitPullRequest(root, GitPullRequestRequest{Agent: "gemini", TaskID: "task", Home: t.TempDir(), StateID: state.StateID, Repository: "source", Remote: "origin", Base: "main", Head: "feat/x", Title: "Title", Body: "Body", Confirm: true}, fakeWorkflowInspector(repos, nil), func(path, remote string) (string, error) {
			return "https://github.com/acme/app.git", nil
		}, func(_ context.Context, request GitPullRequestRequest, remote string) (GitPullRequestResult, error) {
			if remote != "https://github.com/acme/app.git" || request.Title != "Title" {
				t.Fatalf("request=%#v remote=%q", request, remote)
			}
			return GitPullRequestResult{Number: 1, URL: "https://github.com/acme/app/pull/1", Outcome: "created"}, nil
		})
		if err != nil || report.Status != "succeeded" || report.PullRequest == nil || report.PullRequest.Number != 1 {
			t.Fatalf("report=%#v err=%#v", report, err)
		}
	})
}

func TestGitWorkflowDomainRefusesUnsafeInputsBeforeEffects(t *testing.T) {
	root := newGitWorkflowWorkspace(t)
	repos := map[string]gitexec.WorkflowRepository{
		filepath.Join(root, "knowledge"): fakeWorkflowRepo(filepath.Join(root, "knowledge"), "main", "a"),
		filepath.Join(root, "source"):    fakeWorkflowRepo(filepath.Join(root, "source"), "main", "b"),
	}
	state := branchState(t, root, repos)
	calls := 0
	_, err := CreateGitBranch(root, GitBranchRequest{Agent: "codex", TaskID: "task", Home: t.TempDir(), StateID: state.StateID, Name: "topic/free", Confirm: true, Bases: map[string]string{"knowledge": "main", "source": "main"}}, fakeWorkflowInspector(repos, nil), func(string, string, string) error {
		calls++
		return nil
	})
	var failure GitFailure
	if !errors.As(err, &failure) || failure.Code != "validation_failed" || calls != 0 {
		t.Fatalf("branch err=%#v calls=%d", err, calls)
	}
	_, err = CommitGit(root, GitCommitRequest{Agent: "codex", TaskID: "task", Home: t.TempDir(), StateID: state.StateID, Repository: "source", Message: "bad\nmessage", Paths: []string{"file.txt"}, Confirm: true}, fakeWorkflowInspector(repos, nil), func(string, string, []string) error {
		calls++
		return nil
	})
	if !errors.As(err, &failure) || failure.Code != "validation_failed" || calls != 0 {
		t.Fatalf("commit err=%#v calls=%d", err, calls)
	}
	_, err = PushGit(root, GitPushRequest{Agent: "codex", TaskID: "task", Home: t.TempDir(), StateID: state.StateID, Repository: "source", Remote: "https://github.com/acme/app.git", Branch: "main:main", Confirm: true}, fakeWorkflowInspector(repos, nil), func(string, string, string) error {
		calls++
		return nil
	})
	if !errors.As(err, &failure) || failure.Code != "validation_failed" || calls != 0 {
		t.Fatalf("push err=%#v calls=%d", err, calls)
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

func branchState(t *testing.T, root string, repos map[string]gitexec.WorkflowRepository) WorkspaceGitSnapshot {
	t.Helper()
	got, err := InspectGit(root, GitInspectRequest{Agent: "codex", TaskID: "task", Home: t.TempDir()}, fakeWorkflowInspector(repos, nil))
	if err != nil {
		t.Fatal(err)
	}
	return got
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
	return func(path string) (gitexec.WorkflowRepository, error) {
		path = canonical(path)
		if err, ok := failures[path]; ok {
			return gitexec.WorkflowRepository{}, err
		}
		if repo, ok := repos[path]; ok {
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
