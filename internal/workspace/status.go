package workspace

import (
	"errors"
	"os"
	"path/filepath"
)

const (
	RepositoryClean   = "clean"
	RepositoryPending = "pending"
)

type GitRepositoryStatus struct {
	Path           string
	Branch         string
	Commit         string
	ModifiedCount  int
	StagedCount    int
	UntrackedCount int
}

type RepositoryReport struct {
	Name           string
	Path           string
	Branch         string
	Commit         string
	State          string
	ModifiedCount  int
	StagedCount    int
	UntrackedCount int
}

type WorkspaceReport struct {
	ProjectName  string
	Root         string
	Repositories []RepositoryReport
}

type GitStatus func(string) (GitRepositoryStatus, error)

type RepositoryParticipant struct {
	Name string
	Path string
}

type StatusFailure struct {
	Code       string
	Cause      string
	Path       string
	Correction string
}

func (failure StatusFailure) Error() string {
	if failure.Path == "" {
		return failure.Cause
	}
	return failure.Cause + ": " + failure.Path
}

func CurrentStatus(start string, collect GitStatus) (WorkspaceReport, error) {
	root, manifestPath, err := locateWorkspace(start)
	if err != nil {
		return WorkspaceReport{}, err
	}
	data, err := readManifest(manifestPath)
	if err != nil {
		return WorkspaceReport{}, statusFailure("manifest-invalid", "manifesto ausente ou inválido", manifestPath, "corrija ou restaure knowledge/cerne.json")
	}
	if data.VersionErr != nil {
		return WorkspaceReport{}, statusFailure("manifest-version-unsupported", "versão do manifesto não suportada", manifestPath, "use version como inteiro JSON 1 ou remova o campo")
	}
	if data.WorkflowErr != nil {
		return WorkspaceReport{}, statusFailure("workflow-invalid", "workflow inválido no manifesto", manifestPath, "corrija o objeto workflow.provider")
	}

	participants, err := workspaceRepositories(root, data)
	if err != nil {
		return WorkspaceReport{}, err
	}
	if collect == nil {
		return WorkspaceReport{}, statusFailure("git-unavailable", "Git indisponível", "", "instale o Git e disponibilize-o no PATH")
	}

	reports := make([]RepositoryReport, 0, len(participants))
	for _, repository := range participants {
		status, err := collect(repository.Path)
		if err != nil {
			return WorkspaceReport{}, statusFailure("git-status-failed", "não foi possível consultar o repositório Git", repository.Path, "verifique se o diretório é um repositório Git local válido")
		}
		reports = append(reports, repositoryReport(repository.Name, repository.Path, status))
	}
	return WorkspaceReport{ProjectName: data.Name, Root: root, Repositories: reports}, nil
}

func workspaceRepositories(root string, data manifest) ([]RepositoryParticipant, error) {
	knowledge := filepath.Join(root, "knowledge")
	source, err := validateSourcePath(knowledge, data.Source)
	if err != nil {
		return nil, statusFailure("source-invalid", "caminho source inválido no manifesto", manifestSourcePath(knowledge, data.Source), "configure um caminho source existente e seguro")
	}
	if err := regularDir(knowledge); err != nil {
		return nil, statusFailure("knowledge-missing", "repositório de conhecimento não encontrado", knowledge, "restaure o diretório knowledge")
	}
	if err := regularDir(source); err != nil {
		return nil, statusFailure("source-missing", "repositório de código-fonte não encontrado", source, "restaure o diretório source")
	}
	return []RepositoryParticipant{{Name: "knowledge", Path: knowledge}, {Name: "source", Path: source}}, nil
}

func locateWorkspace(start string) (string, string, error) {
	current, err := filepath.Abs(start)
	if err != nil {
		return "", "", statusFailure("workspace-not-found", "workspace Cerne não localizado", start, "execute o comando dentro de um workspace Cerne")
	}
	if info, err := os.Stat(current); err == nil && !info.IsDir() {
		current = filepath.Dir(current)
	}

	candidate := ""
	for {
		manifestPath := filepath.Join(current, "knowledge", "cerne.json")
		if _, err := os.Stat(manifestPath); err == nil || !errors.Is(err, os.ErrNotExist) {
			return canonical(current), manifestPath, nil
		}
		if candidate == "" && looksLikeWorkspace(current) {
			candidate = current
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	if candidate != "" {
		path := filepath.Join(candidate, "knowledge", "cerne.json")
		return "", "", statusFailure("manifest-missing", "manifesto Cerne ausente", path, "restaure knowledge/cerne.json ou execute em outro workspace")
	}
	return "", "", statusFailure("workspace-not-found", "workspace Cerne não localizado", start, "execute o comando dentro de um workspace Cerne")
}

func looksLikeWorkspace(root string) bool {
	return regularDir(filepath.Join(root, "knowledge")) == nil
}

func repositoryReport(name, path string, status GitRepositoryStatus) RepositoryReport {
	state := RepositoryClean
	if status.ModifiedCount+status.StagedCount+status.UntrackedCount > 0 {
		state = RepositoryPending
	}
	return RepositoryReport{
		Name:           name,
		Path:           canonical(path),
		Branch:         status.Branch,
		Commit:         status.Commit,
		State:          state,
		ModifiedCount:  status.ModifiedCount,
		StagedCount:    status.StagedCount,
		UntrackedCount: status.UntrackedCount,
	}
}

func manifestSourcePath(knowledge, source string) string {
	if filepath.IsAbs(source) {
		return source
	}
	return filepath.Join(knowledge, source)
}

func statusFailure(code, cause, path, correction string) StatusFailure {
	if path != "" {
		path = filepath.Clean(path)
	}
	return StatusFailure{Code: code, Cause: cause, Path: path, Correction: correction}
}
