package workspace

import (
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
