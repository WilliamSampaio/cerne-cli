package workspace

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type LinkRequest struct {
	SourceInput string
	Replace     bool
}

type LinkResult struct {
	ProjectName    string
	PreviousSource string
	NewSource      string
	Changed        bool
}

type LinkRepositoryFacts struct {
	RequestedPath string
	WorktreeRoot  string
	CommonDir     string
	IsBare        bool
	HasWorktree   bool
}

type LinkGitInspect func(string) (LinkRepositoryFacts, error)

type LinkFailure struct {
	Code       string
	Cause      string
	Path       string
	Correction string
}

func (failure LinkFailure) Error() string {
	if failure.Path == "" {
		return failure.Cause
	}
	return failure.Cause + ": " + failure.Path
}

type linkManifest struct {
	manifest
	raw map[string]json.RawMessage
}

var replaceManifestFile = atomicReplaceFile

func Link(start string, request LinkRequest, inspect LinkGitInspect) (LinkResult, error) {
	root, manifestPath, err := locateLinkWorkspace(start)
	if err != nil {
		return LinkResult{}, err
	}
	data, err := readLinkManifest(manifestPath)
	if err != nil {
		return LinkResult{}, linkFailure("manifest-invalid", "manifesto ausente ou inválido", manifestPath, "corrija ou restaure knowledge/cerne.json")
	}
	if data.VersionErr != nil {
		return LinkResult{}, linkFailure("manifest-version-unsupported", "versão do manifesto não suportada", manifestPath, "use version como inteiro JSON 1 ou remova o campo")
	}
	if data.WorkflowErr != nil {
		return LinkResult{}, linkFailure("workflow-invalid", "workflow inválido no manifesto", manifestPath, "corrija o objeto workflow.provider")
	}
	if inspect == nil {
		return LinkResult{}, linkFailure("git-unavailable", "Git indisponível", "", "instale o Git e disponibilize-o no PATH")
	}

	knowledge := filepath.Join(root, "knowledge")
	if err := regularDir(knowledge); err != nil {
		return LinkResult{}, linkFailure("knowledge-missing", "repositório de conhecimento não encontrado", knowledge, "restaure o diretório knowledge")
	}
	candidate, err := resolveLinkPath(start, request.SourceInput)
	if err != nil {
		return LinkResult{}, err
	}
	sourceFacts, err := validLinkRepository(inspect, candidate)
	if err != nil {
		return LinkResult{}, err
	}
	knowledgeFacts, err := validLinkRepository(inspect, knowledge)
	if err != nil {
		return LinkResult{}, linkFailure("knowledge-invalid", "repositório knowledge inválido", knowledge, "inicialize ou restaure knowledge como repositório Git local")
	}
	existing := workspaceNamedFacts(inspect, knowledge, knowledgeFacts, data, "")
	if err := validateRepositorySeparation(candidate, sourceFacts, filterOutSource(existing)); err != nil {
		return LinkResult{}, err
	}

	previousPath := canonical(manifestSourcePath(knowledge, data.Source))
	if linkSameSource(inspect, previousPath, candidate) {
		return LinkResult{ProjectName: data.Name, PreviousSource: data.Source, NewSource: data.Source, Changed: false}, nil
	}
	if !request.Replace {
		return LinkResult{}, linkFailure("source-already-configured", "outro source já está configurado", previousPath, "execute novamente com --replace para substituir apenas a referência do manifesto")
	}

	newSource := manifestLinkSource(knowledge, candidate)
	data.raw["source"], _ = json.Marshal(newSource)
	content, err := json.MarshalIndent(data.raw, "", "  ")
	if err != nil {
		return LinkResult{}, linkFailure("manifest-update-unsafe", "manifesto não pode ser atualizado com segurança", manifestPath, "verifique o manifesto e tente novamente")
	}
	content = append(content, '\n')
	if err := writeManifestAtomically(manifestPath, content); err != nil {
		return LinkResult{}, linkFailure("manifest-update-failed", "manifesto não pode ser atualizado com segurança", manifestPath, "verifique permissões e tente novamente")
	}
	return LinkResult{ProjectName: data.Name, PreviousSource: data.Source, NewSource: newSource, Changed: true}, nil
}

func locateLinkWorkspace(start string) (string, string, error) {
	current, err := filepath.Abs(start)
	if err != nil {
		return "", "", linkFailure("workspace-not-found", "workspace Cerne não localizado", start, "execute o comando dentro de um workspace Cerne")
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
		return "", "", linkFailure("manifest-missing", "manifesto Cerne ausente", path, "restaure knowledge/cerne.json ou execute em outro workspace")
	}
	return "", "", linkFailure("workspace-not-found", "workspace Cerne não localizado", start, "execute o comando dentro de um workspace Cerne")
}

func readLinkManifest(path string) (linkManifest, error) {
	base, err := readManifest(path)
	if err != nil {
		return linkManifest{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return linkManifest{}, err
	}
	raw := map[string]json.RawMessage{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return linkManifest{}, err
	}
	return linkManifest{manifest: base, raw: raw}, nil
}

func resolveLinkPath(start, input string) (string, error) {
	if input == "" {
		return "", linkFailure("source-path-missing", "caminho source não informado", "", "informe um diretório de repositório Git local")
	}
	path := input
	if !filepath.IsAbs(path) {
		path = filepath.Join(start, path)
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", linkFailure("source-path-invalid", "caminho source inválido", input, "informe um caminho relativo ou absoluto válido")
	}
	info, err := os.Stat(abs)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", linkFailure("source-path-not-found", "caminho source inexistente", abs, "informe um diretório existente")
		}
		return "", linkFailure("source-path-inaccessible", "caminho source inacessível", abs, "verifique permissões do caminho informado")
	}
	if !info.IsDir() {
		return "", linkFailure("source-not-directory", "caminho source não é diretório", abs, "informe um diretório de repositório Git local")
	}
	return canonical(abs), nil
}

func validLinkRepository(inspect LinkGitInspect, path string) (LinkRepositoryFacts, error) {
	facts, err := inspect(path)
	if err != nil {
		return LinkRepositoryFacts{}, linkFailure("source-not-git", "caminho source não é um repositório Git válido", path, "informe a raiz de um repositório Git local")
	}
	if facts.IsBare {
		return LinkRepositoryFacts{}, linkFailure("source-bare", "repositório source bare não é aceito", path, "informe um repositório Git com árvore de trabalho")
	}
	if !facts.HasWorktree || facts.WorktreeRoot == "" {
		return LinkRepositoryFacts{}, linkFailure("source-no-worktree", "repositório source sem árvore de trabalho", path, "informe um repositório Git não-bare")
	}
	if !samePath(facts.WorktreeRoot, path) {
		return LinkRepositoryFacts{}, linkFailure("source-not-git-root", "caminho source não é raiz Git própria", path, "informe a raiz do repositório Git")
	}
	return facts, nil
}

// namedFacts é um repositório já associado ao workspace, com os fatos Git necessários para
// comparar identidade e sobreposição contra um candidato.
type namedFacts struct {
	Name  string
	Path  string
	Facts LinkRepositoryFacts
}

// validateRepositorySeparation verifica o candidato contra todos os repositórios já associados ao
// workspace — knowledge, source e cada adicional registrado. Como o conjunto é válido antes de
// cada operação, comparar apenas o candidato contra cada membro preserva o invariante.
//
// Dois repositórios que compartilham CommonDir compartilham histórico e NÃO são independentes,
// ainda que suas árvores de trabalho sejam distintas: é o caso de um Git worktree de um
// repositório já registrado, recusado por decisão registrada na especificação (FR-013, FR-014).
func validateRepositorySeparation(candidate string, candidateFacts LinkRepositoryFacts, existing []namedFacts) error {
	for _, other := range existing {
		if samePath(other.Facts.WorktreeRoot, candidateFacts.WorktreeRoot) ||
			samePath(other.Facts.CommonDir, candidateFacts.CommonDir) {
			return linkFailure("repositories-not-independent", "é o mesmo repositório que "+other.Name, candidate, "informe um repositório com histórico Git próprio")
		}
		if containsPath(other.Path, candidate) || containsPath(candidate, other.Path) ||
			containsPath(other.Facts.WorktreeRoot, candidateFacts.WorktreeRoot) ||
			containsPath(candidateFacts.WorktreeRoot, other.Facts.WorktreeRoot) {
			return linkFailure("repositories-overlap", "sobreposição perigosa com "+other.Name, candidate, "mantenha os repositórios em diretórios separados")
		}
	}
	return nil
}

// workspaceNamedFacts coleta os fatos Git de knowledge, source e dos repositórios adicionais já
// registrados, pulando entradas cujo caminho não pode mais ser inspecionado — uma entrada quebrada
// não impede validar o candidato contra as demais.
func workspaceNamedFacts(inspect LinkGitInspect, knowledge string, knowledgeFacts LinkRepositoryFacts, data linkManifest, skipName string) []namedFacts {
	existing := []namedFacts{{Name: "knowledge", Path: knowledge, Facts: knowledgeFacts}}
	if source, err := validateSourcePath(knowledge, data.Source); err == nil {
		if facts, err := inspect(source); err == nil && facts.HasWorktree && !facts.IsBare {
			existing = append(existing, namedFacts{Name: "source", Path: canonical(source), Facts: facts})
		}
	}
	for _, entry := range data.Repositories {
		if entry.Name == skipName {
			continue
		}
		path, err := validateSourcePath(knowledge, entry.Path)
		if err != nil {
			continue
		}
		facts, err := inspect(path)
		if err != nil || !facts.HasWorktree || facts.IsBare {
			continue
		}
		existing = append(existing, namedFacts{Name: entry.Name, Path: canonical(path), Facts: facts})
	}
	return existing
}

func linkSameSource(inspect LinkGitInspect, previous, candidate string) bool {
	if samePath(previous, candidate) {
		return true
	}
	previousFacts, err := inspect(previous)
	if err != nil || previousFacts.IsBare || !previousFacts.HasWorktree {
		return false
	}
	candidateFacts, err := inspect(candidate)
	if err != nil || candidateFacts.IsBare || !candidateFacts.HasWorktree {
		return false
	}
	return samePath(previousFacts.WorktreeRoot, candidateFacts.WorktreeRoot) &&
		samePath(previousFacts.CommonDir, candidateFacts.CommonDir)
}

func manifestLinkSource(knowledge, source string) string {
	if rel, err := filepath.Rel(knowledge, source); err == nil {
		return filepath.ToSlash(filepath.Clean(rel))
	}
	return filepath.Clean(source)
}

func writeManifestAtomically(path string, content []byte) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	temp, err := os.CreateTemp(dir, ".cerne-*.tmp")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)
	if err := temp.Chmod(info.Mode().Perm()); err != nil {
		temp.Close()
		return err
	}
	if _, err := temp.Write(content); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Sync(); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return replaceManifestFile(tempPath, path)
}

func linkFailure(code, cause, path, correction string) LinkFailure {
	if path != "" {
		path = filepath.Clean(path)
	}
	return LinkFailure{Code: code, Cause: cause, Path: path, Correction: correction}
}

// filterOutSource remove o source atual do conjunto de comparação: em `cerne link` ele está sendo
// substituído, então comparar o candidato contra ele recusaria uma troca legítima. O caso do
// candidato ser o próprio source já configurado é tratado antes, por linkSameSource.
func filterOutSource(existing []namedFacts) []namedFacts {
	filtered := make([]namedFacts, 0, len(existing))
	for _, entry := range existing {
		if entry.Name == "source" {
			continue
		}
		filtered = append(filtered, entry)
	}
	return filtered
}

// validateLinkSeparation é o caso de dois repositórios do invariante generalizado, usado por init e
// restore, onde o workspace ainda não tem repositórios adicionais registrados.
func validateLinkSeparation(knowledge, source string, knowledgeFacts, sourceFacts LinkRepositoryFacts) error {
	return validateRepositorySeparation(source, sourceFacts, []namedFacts{
		{Name: "knowledge", Path: knowledge, Facts: knowledgeFacts},
	})
}
