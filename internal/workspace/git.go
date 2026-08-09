package workspace

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/WilliamSampaio/cerne-cli/internal/gitexec"
)

type GitInspectRequest struct {
	Agent  string
	TaskID string
	Home   string
}

type GitBranchRequest struct {
	Agent   string
	TaskID  string
	Home    string
	StateID string
	Name    string
	Bases   map[string]string
	Confirm bool
}

type GitCommitRequest struct {
	Agent      string
	TaskID     string
	Home       string
	StateID    string
	Repository string
	Message    string
	Paths      []string
	Confirm    bool
}

type GitPushRequest struct {
	Agent      string
	TaskID     string
	Home       string
	StateID    string
	Repository string
	Remote     string
	Branch     string
	Confirm    bool
}

type GitPullRequestRequest struct {
	Agent      string
	TaskID     string
	Home       string
	StateID    string
	Repository string
	Remote     string
	Base       string
	Head       string
	Title      string
	Body       string
	Confirm    bool
}

type PullRequestOpener func(context.Context, GitPullRequestRequest, string) (GitPullRequestResult, error)

type GitPullRequestResult struct {
	Number  int    `json:"number"`
	URL     string `json:"url"`
	Outcome string `json:"outcome"`
}

type WorkspaceGitMutationReport struct {
	SchemaVersion int                   `json:"schema_version"`
	Operation     string                `json:"operation"`
	Status        string                `json:"status"`
	StateBefore   string                `json:"state_before,omitempty"`
	StateAfter    string                `json:"state_after,omitempty"`
	Aligned       bool                  `json:"aligned"`
	AuditID       string                `json:"audit_id,omitempty"`
	Repositories  []RepositoryGitEffect `json:"repositories"`
	PullRequest   *GitPullRequestResult `json:"pull_request,omitempty"`
	Problems      []GitProblem          `json:"problems"`
	ErrorCode     string                `json:"error_code,omitempty"`
}

type RepositoryGitEffect struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	Status    string `json:"status"`
	Branch    string `json:"branch,omitempty"`
	Head      string `json:"head,omitempty"`
	ErrorCode string `json:"error_code,omitempty"`
}

type WorkspaceGitSnapshot struct {
	SchemaVersion int                  `json:"schema_version"`
	Status        string               `json:"status"`
	StateID       string               `json:"state_id"`
	AuditID       string               `json:"audit_id,omitempty"`
	Workspace     *GitWorkspace        `json:"workspace,omitempty"`
	Repositories  []RepositoryGitState `json:"repositories"`
	Problems      []GitProblem         `json:"problems"`
}

type GitWorkspace struct {
	Name string `json:"name"`
	Root string `json:"root"`
}

type RepositoryGitState struct {
	Name           string                    `json:"name"`
	Path           string                    `json:"path"`
	Branch         string                    `json:"branch"`
	Head           string                    `json:"head"`
	DefaultBranch  string                    `json:"default_branch"`
	Clean          bool                      `json:"clean"`
	Changes        []GitChange               `json:"changes"`
	LocalBranches  []string                  `json:"local_branches"`
	RemoteBranches []string                  `json:"remote_branches"`
	Upstream       *gitexec.WorkflowUpstream `json:"upstream,omitempty"`
	Remotes        []gitexec.WorkflowRemote  `json:"remotes"`
}

type GitChange struct {
	Path     string `json:"path"`
	Index    string `json:"index"`
	Worktree string `json:"worktree"`
	Digest   string `json:"digest"`
}

type GitProblem struct {
	Code      string `json:"code"`
	Component string `json:"component"`
}

type GitFailure struct {
	Code       string
	Cause      string
	Correction string
}

func (failure GitFailure) Error() string { return failure.Cause }

type gitAuditRecord struct {
	SchemaVersion int         `json:"schema_version"`
	Executor      string      `json:"executor"`
	Agent         string      `json:"agent"`
	TaskID        string      `json:"task_id"`
	Operation     string      `json:"operation"`
	Authorization string      `json:"authorization"`
	StateID       string      `json:"state_id,omitempty"`
	Targets       []gitTarget `json:"targets"`
	Status        string      `json:"status"`
	StartedAt     string      `json:"started_at"`
	FinishedAt    string      `json:"finished_at,omitempty"`
	Phases        []gitPhase  `json:"phases,omitempty"`
}

type gitTarget struct {
	Repository string `json:"repository"`
}

type gitPhase struct {
	Repository string `json:"repository"`
	Operation  string `json:"operation"`
	Status     string `json:"status"`
	ErrorCode  string `json:"error_code,omitempty"`
}

var replaceGitAudit = atomicReplaceFile
var workflowOperationInProgress = gitexec.WorkflowOperationInProgress

func InspectGit(start string, request GitInspectRequest, inspect gitexec.WorkflowInspector) (WorkspaceGitSnapshot, error) {
	if !supportedGitAgent(request.Agent) || !validTaskID(request.TaskID) || request.Home == "" || inspect == nil {
		return WorkspaceGitSnapshot{}, gitFailure("validation_failed", "argumento inválido", "informe agente, tarefa e formato válidos")
	}
	root, manifestPath, err := locateWorkspace(start)
	if err != nil {
		return invalidGitSnapshot(err), nil
	}
	data, err := readManifest(manifestPath)
	if err != nil || data.VersionErr != nil || data.WorkflowErr != nil {
		return invalidGitSnapshot(statusFailure("manifest-invalid", "manifesto ausente ou inválido", manifestPath, "corrija knowledge/cerne.json")), nil
	}
	participants, err := workspaceRepositories(root, data)
	if err != nil {
		return invalidGitSnapshot(err), nil
	}
	audit, auditID, err := startGitAudit(request.Home, request.Agent, request.TaskID, "inspect", "not-required", participants)
	if err != nil {
		return WorkspaceGitSnapshot{}, gitFailure("audit_unavailable", "audit_unavailable", "verifique ~/.cerne/audit")
	}
	snapshot := WorkspaceGitSnapshot{
		SchemaVersion: 1,
		Status:        "healthy",
		AuditID:       auditID,
		Workspace:     &GitWorkspace{Name: data.Name, Root: root},
		Problems:      []GitProblem{},
	}
	for _, participant := range participants {
		repository, err := inspect(participant.Path)
		if err != nil {
			snapshot.Status = "invalid"
			snapshot.Problems = append(snapshot.Problems, GitProblem{Code: "repository_invalid", Component: participant.Name})
			continue
		}
		snapshot.Repositories = append(snapshot.Repositories, repositoryGitState(participant.Name, repository))
	}
	if snapshot.Status == "healthy" {
		snapshot.StateID = snapshotStateID(snapshot)
	}
	status := "succeeded"
	if snapshot.Status != "healthy" {
		status = "failed"
	}
	if err := audit.finish(status, snapshot.StateID, snapshot.Problems); err != nil {
		return snapshot, gitFailure("audit_incomplete", "audit_incomplete", "verifique ~/.cerne/audit")
	}
	return snapshot, nil
}

func CreateGitBranch(start string, request GitBranchRequest, inspect gitexec.WorkflowInspector, branch gitexec.WorkflowBrancher) (WorkspaceGitMutationReport, error) {
	if !supportedGitAgent(request.Agent) || !validTaskID(request.TaskID) || request.Home == "" || request.StateID == "" ||
		request.Name == "" || !request.Confirm || inspect == nil || branch == nil {
		return WorkspaceGitMutationReport{}, gitFailure("validation_failed", "argumento inválido", "informe agente, tarefa, estado, nome, bases e confirmação")
	}
	if err := gitexec.ValidateWorkflowBranchName(request.Name); err != nil {
		return WorkspaceGitMutationReport{}, gitFailure("validation_failed", "branch inválida", "informe um nome aceito por git check-ref-format --branch")
	}
	if !validWorkflowBranchSlug(request.Name) {
		return WorkspaceGitMutationReport{}, gitFailure("validation_failed", "branch inválida", "use feat|fix|refactor|chore|spec/<slug>")
	}
	root, manifestPath, err := locateWorkspace(start)
	if err != nil {
		return WorkspaceGitMutationReport{SchemaVersion: 1, Operation: "branch_create", Status: "failed", Problems: invalidGitSnapshot(err).Problems}, nil
	}
	data, err := readManifest(manifestPath)
	if err != nil || data.VersionErr != nil || data.WorkflowErr != nil {
		return WorkspaceGitMutationReport{SchemaVersion: 1, Operation: "branch_create", Status: "failed", Problems: invalidGitSnapshot(statusFailure("manifest-invalid", "manifesto ausente ou inválido", manifestPath, "corrija knowledge/cerne.json")).Problems}, nil
	}
	participants, err := workspaceRepositories(root, data)
	if err != nil {
		return WorkspaceGitMutationReport{SchemaVersion: 1, Operation: "branch_create", Status: "failed", Problems: invalidGitSnapshot(err).Problems}, nil
	}
	if !hasExactBases(participants, request.Bases) {
		return WorkspaceGitMutationReport{}, gitFailure("validation_failed", "bases inválidas", "informe exatamente uma base por repositório")
	}
	audit, auditID, err := startGitAudit(request.Home, request.Agent, request.TaskID, "branch_create", "confirmed", participants)
	if err != nil {
		return WorkspaceGitMutationReport{}, gitFailure("audit_unavailable", "audit_unavailable", "verifique ~/.cerne/audit")
	}
	report := WorkspaceGitMutationReport{SchemaVersion: 1, Operation: "branch_create", Status: "succeeded", AuditID: auditID}
	before := inspectParticipants(participants, inspect)
	report.StateBefore = before.StateID
	report.Repositories = mutationEffects(before.Repositories, "validated")
	report.Problems = before.Problems
	if before.Status != "healthy" {
		report.Status = "failed"
		return finishGitMutation(audit, &report, nil)
	}
	if before.StateID != request.StateID {
		report.Status = "blocked"
		report.Problems = append(report.Problems, GitProblem{Code: "stale_state", Component: "workspace"})
		return finishGitMutation(audit, &report, nil)
	}
	if problems := validateBranchPreconditions(before.Repositories, request); len(problems) > 0 {
		report.Status = "blocked"
		report.Problems = append(report.Problems, problems...)
		return finishGitMutation(audit, &report, nil)
	}
	succeeded := 0
	for i, repo := range before.Repositories {
		if err := branch(repo.Path, request.Name, request.Bases[repo.Name]); err != nil {
			report.Repositories[i].Status = "failed"
			report.Repositories[i].ErrorCode = "branch_create_failed"
			report.Status = "failed"
			if succeeded > 0 {
				report.Status = "partial"
			}
			break
		}
		succeeded++
		report.Repositories[i].Status = "succeeded"
		report.Repositories[i].Branch = request.Name
	}
	for i := range report.Repositories {
		if report.Repositories[i].Status == "validated" {
			report.Repositories[i].Status = "not-run"
		}
	}
	after := inspectParticipants(participants, inspect)
	report.StateAfter = after.StateID
	if report.Status == "succeeded" && after.Status != "healthy" {
		report.Status = "failed"
		report.Problems = append(report.Problems, after.Problems...)
	}
	report.Aligned = report.Status == "succeeded"
	return finishGitMutation(audit, &report, nil)
}

func CommitGit(start string, request GitCommitRequest, inspect gitexec.WorkflowInspector, commit gitexec.WorkflowCommitter) (WorkspaceGitMutationReport, error) {
	if !validMutationRequest(request.Agent, request.TaskID, request.Home, request.StateID, request.Confirm) ||
		request.Repository == "" || strings.TrimSpace(request.Message) == "" || strings.ContainsAny(request.Message, "\r\n") || len(request.Paths) == 0 ||
		inspect == nil || commit == nil {
		return WorkspaceGitMutationReport{}, gitFailure("validation_failed", "argumento inválido", "informe repositório, mensagem, paths, estado e confirmação")
	}
	for _, path := range request.Paths {
		if !gitexec.ValidateWorkflowLiteralPath(path) {
			return WorkspaceGitMutationReport{}, gitFailure("validation_failed", "path inválido", "use paths relativos literais")
		}
	}
	report, participant, before, audit, err := prepareSingleGitMutation(start, request.Home, request.Agent, request.TaskID, request.StateID, request.Repository, "commit", inspect)
	if err != nil {
		return report, err
	}
	if report.Status != "succeeded" {
		return finishGitMutation(audit, &report, nil)
	}
	repo := findRepository(before.Repositories, request.Repository)
	if repo.Branch == gitexec.DetachedHEAD || workflowOperationInProgress(repo.Path) {
		report.Status = "blocked"
		report.ErrorCode = "git_operation_in_progress"
		report.Problems = append(report.Problems, GitProblem{Code: "git_operation_in_progress", Component: repo.Name})
		return finishGitMutation(audit, &report, nil)
	}
	for _, path := range request.Paths {
		if !changedPath(repo, path) {
			report.Status = "blocked"
			report.ErrorCode = "path_not_changed"
			report.Problems = append(report.Problems, GitProblem{Code: "path_not_changed", Component: repo.Name})
			return finishGitMutation(audit, &report, nil)
		}
	}
	if err := commit(participant.Path, request.Message, request.Paths); err != nil {
		report.Status = "failed"
		report.ErrorCode = "commit_failed"
		report.Repositories[0].Status = "failed"
		report.Repositories[0].ErrorCode = "commit_failed"
		return finishAfterMutation(participantsFrom(participant), inspect, audit, &report)
	}
	report.Repositories[0].Status = "succeeded"
	return finishAfterMutation(participantsFrom(participant), inspect, audit, &report)
}

func PushGit(start string, request GitPushRequest, inspect gitexec.WorkflowInspector, push gitexec.WorkflowPusher) (WorkspaceGitMutationReport, error) {
	if !validMutationRequest(request.Agent, request.TaskID, request.Home, request.StateID, request.Confirm) ||
		request.Repository == "" || request.Remote == "" || request.Branch == "" || inspect == nil || push == nil {
		return WorkspaceGitMutationReport{}, gitFailure("validation_failed", "argumento inválido", "informe repositório, remote, branch, estado e confirmação")
	}
	if !validRemoteName(request.Remote) || gitexec.ValidateWorkflowBranchName(request.Branch) != nil {
		return WorkspaceGitMutationReport{}, gitFailure("validation_failed", "argumento inválido", "use remote local e branch local explícitos")
	}
	report, participant, before, audit, err := prepareSingleGitMutation(start, request.Home, request.Agent, request.TaskID, request.StateID, request.Repository, "push", inspect)
	if err != nil {
		return report, err
	}
	if report.Status != "succeeded" {
		return finishGitMutation(audit, &report, nil)
	}
	repo := findRepository(before.Repositories, request.Repository)
	if repo.Branch != request.Branch || !containsGitName(repo.LocalBranches, request.Branch) || !containsRemote(repo.Remotes, request.Remote) {
		report.Status = "blocked"
		report.ErrorCode = "validation_failed"
		report.Problems = append(report.Problems, GitProblem{Code: "validation_failed", Component: repo.Name})
		return finishGitMutation(audit, &report, nil)
	}
	if err := push(participant.Path, request.Remote, request.Branch); err != nil {
		report.Status = "failed"
		report.ErrorCode = "push_rejected"
		report.Repositories[0].Status = "failed"
		report.Repositories[0].ErrorCode = "push_rejected"
		return finishAfterMutation(participantsFrom(participant), inspect, audit, &report)
	}
	report.Repositories[0].Status = "succeeded"
	return finishAfterMutation(participantsFrom(participant), inspect, audit, &report)
}

func CreateGitPullRequest(start string, request GitPullRequestRequest, inspect gitexec.WorkflowInspector, remoteURL gitexec.WorkflowRemoteURLReader, open PullRequestOpener) (WorkspaceGitMutationReport, error) {
	if !validMutationRequest(request.Agent, request.TaskID, request.Home, request.StateID, request.Confirm) ||
		request.Repository == "" || request.Remote == "" || request.Base == "" || request.Head == "" || strings.TrimSpace(request.Title) == "" ||
		inspect == nil || remoteURL == nil || open == nil {
		return WorkspaceGitMutationReport{}, gitFailure("validation_failed", "argumento inválido", "informe repositório, remote, base, head, título, estado e confirmação")
	}
	if !validRemoteName(request.Remote) || gitexec.ValidateWorkflowBranchName(request.Base) != nil || gitexec.ValidateWorkflowBranchName(request.Head) != nil {
		return WorkspaceGitMutationReport{}, gitFailure("validation_failed", "argumento inválido", "use remote local e branches explícitas")
	}
	report, participant, before, audit, err := prepareSingleGitMutation(start, request.Home, request.Agent, request.TaskID, request.StateID, request.Repository, "pull_request_create", inspect)
	if err != nil {
		return report, err
	}
	if report.Status != "succeeded" {
		return finishGitMutation(audit, &report, nil)
	}
	repo := findRepository(before.Repositories, request.Repository)
	if !containsRemote(repo.Remotes, request.Remote) || !containsGitName(repo.RemoteBranches, request.Remote+"/"+request.Head) {
		report.Status = "blocked"
		report.ErrorCode = "github_remote_required"
		report.Problems = append(report.Problems, GitProblem{Code: "github_remote_required", Component: repo.Name})
		return finishGitMutation(audit, &report, nil)
	}
	rawURL, err := remoteURL(participant.Path, request.Remote)
	if err != nil {
		report.Status = "blocked"
		report.ErrorCode = "remote_missing"
		report.Problems = append(report.Problems, GitProblem{Code: "remote_missing", Component: repo.Name})
		return finishGitMutation(audit, &report, nil)
	}
	result, err := open(context.Background(), request, rawURL)
	if err != nil {
		report.Status = "failed"
		report.ErrorCode = safeErrorCode(err, "remote_result_unknown")
		report.Repositories[0].Status = "failed"
		report.Repositories[0].ErrorCode = report.ErrorCode
		return finishGitMutation(audit, &report, nil)
	}
	report.PullRequest = &result
	report.Repositories[0].Status = "succeeded"
	return finishGitMutation(audit, &report, nil)
}

func repositoryGitState(name string, repository gitexec.WorkflowRepository) RepositoryGitState {
	changes := make([]GitChange, 0, len(repository.Changes))
	for _, change := range repository.Changes {
		changes = append(changes, GitChange{Path: change.Path, Index: change.Index, Worktree: change.Worktree, Digest: change.Digest})
	}
	return RepositoryGitState{
		Name:           name,
		Path:           repository.Path,
		Branch:         repository.Branch,
		Head:           repository.Head,
		DefaultBranch:  repository.DefaultBranch,
		Clean:          repository.Clean,
		Changes:        changes,
		LocalBranches:  repository.LocalBranches,
		RemoteBranches: repository.RemoteBranches,
		Upstream:       repository.Upstream,
		Remotes:        repository.Remotes,
	}
}

func invalidGitSnapshot(err error) WorkspaceGitSnapshot {
	code := "workspace_not_found"
	var failure StatusFailure
	if errors.As(err, &failure) {
		code = strings.ReplaceAll(failure.Code, "-", "_")
	}
	return WorkspaceGitSnapshot{SchemaVersion: 1, Status: "invalid", Repositories: []RepositoryGitState{}, Problems: []GitProblem{{Code: code, Component: "workspace"}}}
}

func inspectParticipants(participants []RepositoryParticipant, inspect gitexec.WorkflowInspector) WorkspaceGitSnapshot {
	snapshot := WorkspaceGitSnapshot{SchemaVersion: 1, Status: "healthy", Problems: []GitProblem{}}
	for _, participant := range participants {
		repository, err := inspect(participant.Path)
		if err != nil {
			snapshot.Status = "invalid"
			snapshot.Problems = append(snapshot.Problems, GitProblem{Code: "repository_invalid", Component: participant.Name})
			continue
		}
		snapshot.Repositories = append(snapshot.Repositories, repositoryGitState(participant.Name, repository))
	}
	if snapshot.Status == "healthy" {
		snapshot.StateID = snapshotStateID(snapshot)
	}
	return snapshot
}

func mutationEffects(repositories []RepositoryGitState, status string) []RepositoryGitEffect {
	effects := make([]RepositoryGitEffect, 0, len(repositories))
	for _, repo := range repositories {
		effects = append(effects, RepositoryGitEffect{Name: repo.Name, Path: repo.Path, Status: status, Branch: repo.Branch, Head: repo.Head})
	}
	return effects
}

func validateBranchPreconditions(repositories []RepositoryGitState, request GitBranchRequest) []GitProblem {
	var problems []GitProblem
	for _, repo := range repositories {
		switch {
		case !repo.Clean:
			problems = append(problems, GitProblem{Code: "repository_dirty", Component: repo.Name})
		case repo.Branch == gitexec.DetachedHEAD:
			problems = append(problems, GitProblem{Code: "detached_head", Component: repo.Name})
		case workflowOperationInProgress(repo.Path):
			problems = append(problems, GitProblem{Code: "operation_in_progress", Component: repo.Name})
		case !containsGitName(repo.LocalBranches, request.Bases[repo.Name]):
			problems = append(problems, GitProblem{Code: "base_missing", Component: repo.Name})
		case containsGitName(repo.LocalBranches, request.Name) || containsGitSuffix(repo.RemoteBranches, "/"+request.Name):
			problems = append(problems, GitProblem{Code: "target_exists", Component: repo.Name})
		}
	}
	return problems
}

func hasExactBases(participants []RepositoryParticipant, bases map[string]string) bool {
	if len(bases) != len(participants) {
		return false
	}
	for _, participant := range participants {
		if bases[participant.Name] == "" {
			return false
		}
	}
	return true
}

func containsGitName(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func containsGitSuffix(values []string, suffix string) bool {
	for _, value := range values {
		if strings.HasSuffix(value, suffix) {
			return true
		}
	}
	return false
}

func validMutationRequest(agent, taskID, home, stateID string, confirm bool) bool {
	return supportedGitAgent(agent) && validTaskID(taskID) && home != "" && stateID != "" && confirm
}

func validWorkflowBranchSlug(name string) bool {
	prefix, slug, ok := strings.Cut(name, "/")
	if !ok || slug == "" {
		return false
	}
	if prefix != "feat" && prefix != "fix" && prefix != "refactor" && prefix != "chore" && prefix != "spec" {
		return false
	}
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '/' {
			continue
		}
		return false
	}
	return !strings.Contains(slug, "//") && !strings.HasPrefix(slug, "/") && !strings.HasSuffix(slug, "/")
}

func validRemoteName(name string) bool {
	if name == "" || strings.HasPrefix(name, "-") || strings.ContainsAny(name, "\\:\r\n\x00") {
		return false
	}
	for _, part := range strings.Split(name, "/") {
		if part == "" || part == "." || part == ".." {
			return false
		}
	}
	return true
}

func prepareSingleGitMutation(start, home, agent, taskID, stateID, repository, operation string, inspect gitexec.WorkflowInspector) (WorkspaceGitMutationReport, RepositoryParticipant, WorkspaceGitSnapshot, gitAudit, error) {
	participants, err := gitParticipants(start)
	if err != nil {
		return WorkspaceGitMutationReport{SchemaVersion: 1, Operation: operation, Status: "failed", Problems: invalidGitSnapshot(err).Problems}, RepositoryParticipant{}, WorkspaceGitSnapshot{}, gitAudit{}, nil
	}
	participant, ok := findParticipant(participants, repository)
	if !ok {
		return WorkspaceGitMutationReport{}, RepositoryParticipant{}, WorkspaceGitSnapshot{}, gitAudit{}, gitFailure("validation_failed", "repositório inválido", "use um repositório do snapshot")
	}
	audit, auditID, err := startGitAudit(home, agent, taskID, operation, "explicit-confirmation", []RepositoryParticipant{participant})
	if err != nil {
		return WorkspaceGitMutationReport{}, RepositoryParticipant{}, WorkspaceGitSnapshot{}, gitAudit{}, gitFailure("audit_unavailable", "audit_unavailable", "verifique ~/.cerne/audit")
	}
	before := inspectParticipants(participants, inspect)
	report := WorkspaceGitMutationReport{SchemaVersion: 1, Operation: operation, Status: "succeeded", StateBefore: before.StateID, AuditID: auditID, Repositories: mutationEffects([]RepositoryGitState{findRepository(before.Repositories, repository)}, "validated"), Problems: before.Problems}
	if before.Status != "healthy" {
		report.Status = "failed"
		return report, participant, before, audit, nil
	}
	if before.StateID != stateID {
		report.Status = "blocked"
		report.ErrorCode = "state_changed"
		report.Problems = append(report.Problems, GitProblem{Code: "state_changed", Component: "workspace"})
		return report, participant, before, audit, nil
	}
	return report, participant, before, audit, nil
}

func gitParticipants(start string) ([]RepositoryParticipant, error) {
	root, manifestPath, err := locateWorkspace(start)
	if err != nil {
		return nil, err
	}
	data, err := readManifest(manifestPath)
	if err != nil || data.VersionErr != nil || data.WorkflowErr != nil {
		return nil, statusFailure("manifest-invalid", "manifesto ausente ou inválido", manifestPath, "corrija knowledge/cerne.json")
	}
	return workspaceRepositories(root, data)
}

func findParticipant(participants []RepositoryParticipant, name string) (RepositoryParticipant, bool) {
	for _, participant := range participants {
		if participant.Name == name {
			return participant, true
		}
	}
	return RepositoryParticipant{}, false
}

func findRepository(repositories []RepositoryGitState, name string) RepositoryGitState {
	for _, repository := range repositories {
		if repository.Name == name {
			return repository
		}
	}
	return RepositoryGitState{Name: name}
}

func participantsFrom(participant RepositoryParticipant) []RepositoryParticipant {
	return []RepositoryParticipant{participant}
}

func finishAfterMutation(participants []RepositoryParticipant, inspect gitexec.WorkflowInspector, audit gitAudit, report *WorkspaceGitMutationReport) (WorkspaceGitMutationReport, error) {
	after := inspectParticipants(participants, inspect)
	report.StateAfter = after.StateID
	if report.Status == "succeeded" && after.Status != "healthy" {
		report.Status = "failed"
		report.ErrorCode = "repository_invalid"
		report.Problems = append(report.Problems, after.Problems...)
	}
	return finishGitMutation(audit, report, nil)
}

func changedPath(repo RepositoryGitState, path string) bool {
	for _, change := range repo.Changes {
		if change.Path == path {
			return true
		}
	}
	return false
}

func containsRemote(remotes []gitexec.WorkflowRemote, name string) bool {
	for _, remote := range remotes {
		if remote.Name == name {
			return true
		}
	}
	return false
}

func safeErrorCode(err error, fallback string) string {
	var failure GitFailure
	if errors.As(err, &failure) && failure.Code != "" {
		return failure.Code
	}
	type coded interface{ Error() string }
	var withCode coded
	if errors.As(err, &withCode) && withCode.Error() != "" && validTaskID(withCode.Error()) {
		return withCode.Error()
	}
	return fallback
}

func snapshotStateID(snapshot WorkspaceGitSnapshot) string {
	var lines []string
	for _, repo := range snapshot.Repositories {
		lines = append(lines, repo.Name, repo.Branch, repo.Head, repo.DefaultBranch)
		for _, branch := range repo.LocalBranches {
			lines = append(lines, "local:"+branch)
		}
		for _, branch := range repo.RemoteBranches {
			lines = append(lines, "remote:"+branch)
		}
		for _, change := range repo.Changes {
			lines = append(lines, change.Path, change.Index, change.Worktree, change.Digest)
		}
	}
	return digest(lines...)
}

func supportedGitAgent(agent string) bool {
	return agent == "codex" || agent == "claude" || agent == "gemini"
}

func validTaskID(value string) bool {
	if value == "" || len(value) > 80 {
		return false
	}
	for _, r := range value {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' || r == '.' {
			continue
		}
		return false
	}
	return true
}

type gitAudit struct {
	path string
}

func startGitAudit(home, agent, taskID, operation, authorization string, participants []RepositoryParticipant) (gitAudit, string, error) {
	auditDir := filepath.Join(canonical(home), ".cerne", "audit")
	for _, dir := range []string{filepath.Dir(auditDir), auditDir} {
		if err := ensureGitAuditDir(dir); err != nil {
			return gitAudit{}, "", err
		}
	}
	id := randomGitID()
	record := gitAuditRecord{
		SchemaVersion: 1,
		Executor:      "cerne-cli",
		Agent:         agent,
		TaskID:        taskID,
		Operation:     operation,
		Authorization: authorization,
		Status:        "started",
		StartedAt:     time.Now().UTC().Format(time.RFC3339Nano),
	}
	for _, participant := range participants {
		record.Targets = append(record.Targets, gitTarget{Repository: participant.Name})
	}
	audit := gitAudit{path: filepath.Join(auditDir, "git-"+id+".json")}
	if err := audit.write(record); err != nil {
		return gitAudit{}, "", err
	}
	return audit, id, nil
}

func ensureGitAuditDir(path string) error {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		if err := os.Mkdir(path, 0o700); err != nil {
			return err
		}
		return secureAuditPath(path, true)
	}
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return fmt.Errorf("unsafe audit directory")
	}
	return secureAuditPath(path, true)
}

func (audit gitAudit) finish(status, stateID string, problems []GitProblem) error {
	data, err := os.ReadFile(audit.path)
	if err != nil {
		return err
	}
	var record gitAuditRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return err
	}
	record.Status = status
	record.StateID = stateID
	record.FinishedAt = time.Now().UTC().Format(time.RFC3339Nano)
	for _, problem := range problems {
		record.Phases = append(record.Phases, gitPhase{Repository: problem.Component, Operation: "inspect", Status: "failed", ErrorCode: problem.Code})
	}
	sort.Slice(record.Phases, func(i, j int) bool { return record.Phases[i].Repository < record.Phases[j].Repository })
	return audit.write(record)
}

func finishGitMutation(audit gitAudit, report *WorkspaceGitMutationReport, phases []gitPhase) (WorkspaceGitMutationReport, error) {
	if phases == nil {
		for _, effect := range report.Repositories {
			phase := gitPhase{Repository: effect.Name, Operation: report.Operation, Status: effect.Status, ErrorCode: effect.ErrorCode}
			if phase.Status == "validated" {
				phase.Status = report.Status
			}
			phases = append(phases, phase)
		}
		for _, problem := range report.Problems {
			phases = append(phases, gitPhase{Repository: problem.Component, Operation: report.Operation, Status: "failed", ErrorCode: problem.Code})
		}
	}
	data, err := os.ReadFile(audit.path)
	if err != nil {
		return *report, gitFailure("audit_incomplete", "audit_incomplete", "verifique ~/.cerne/audit")
	}
	var record gitAuditRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return *report, gitFailure("audit_incomplete", "audit_incomplete", "verifique ~/.cerne/audit")
	}
	record.Status = report.Status
	record.StateID = report.StateAfter
	if record.StateID == "" {
		record.StateID = report.StateBefore
	}
	record.FinishedAt = time.Now().UTC().Format(time.RFC3339Nano)
	record.Phases = phases
	if err := audit.write(record); err != nil {
		return *report, gitFailure("audit_incomplete", "audit_incomplete", "verifique ~/.cerne/audit")
	}
	return *report, nil
}

func (audit gitAudit) write(record gitAuditRecord) error {
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(audit.path), ".git-audit-*.tmp")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)
	if _, err := temp.Write(append(data, '\n')); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	if err := secureAuditPath(tempPath, false); err != nil {
		return err
	}
	if err := replaceGitAudit(tempPath, audit.path); err != nil {
		return err
	}
	return secureAuditPath(audit.path, false)
}

func randomGitID() string {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return fmt.Sprint(time.Now().UnixNano())
	}
	return hex.EncodeToString(value[:])
}

func digest(values ...string) string {
	h := sha256.New()
	for _, value := range values {
		h.Write([]byte(value))
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))
}

func gitFailure(code, cause, correction string) GitFailure {
	return GitFailure{Code: code, Cause: cause, Correction: correction}
}
