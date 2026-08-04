package workspace

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Severity int

const (
	Pass Severity = iota
	Warning
	Error
)

type Status string

const (
	Healthy  Status = "healthy"
	Warnings Status = "warning"
	Invalid  Status = "invalid"
)

type CheckResult struct {
	ID         string
	Label      string
	Severity   Severity
	Detail     string
	Correction string
}

type Diagnosis struct {
	Root   string
	Checks []CheckResult
	Status Status
}

type RepositoryFacts struct {
	RequestedRoot string
	WorktreeRoot  string
	CommonDir     string
}

type AccessOutcome string

const (
	AccessAllowed AccessOutcome = "allowed"
	AccessDenied  AccessOutcome = "denied"
	AccessUnknown AccessOutcome = "unknown"
)

type AccessResult struct {
	Path   string
	Read   AccessOutcome
	Write  AccessOutcome
	Reason string
}

type GitInspect func(string) (RepositoryFacts, error)
type AccessCheck func(string) AccessResult

func Doctor(root string, inspectGit GitInspect, checkAccess AccessCheck) Diagnosis {
	root = canonical(root)
	knowledge := filepath.Join(root, "knowledge")
	manifestPath := filepath.Join(knowledge, "cerne.json")

	manifest, manifestErr := readManifest(manifestPath)
	source := filepath.Join(root, "source")
	if manifest.Source != "" {
		source = manifestSourcePath(knowledge, manifest.Source)
	}
	sourceValid, sourceErr := validateSourcePath(knowledge, manifest.Source)
	if sourceValid != "" {
		source = sourceValid
	}

	knowledgeOK := regularDir(knowledge) == nil
	sourceOK := sourceErr == nil && regularDir(source) == nil
	knowledgeGit, knowledgeGitErr := inspectRepository(inspectGit, knowledge)
	sourceGit, sourceGitErr := inspectRepository(inspectGit, source)

	checks := []CheckResult{
		manifestCheck(root, manifest, manifestErr),
		dirCheck("knowledge", "Repositório de conhecimento", knowledgeOK),
		dirCheck("source", "Repositório de código-fonte", sourceOK),
		gitIndependenceCheck(inspectGit, knowledge, source, knowledgeGit, sourceGit, knowledgeGitErr, sourceGitErr),
		versioningIsolationCheck(inspectGit, knowledge, source, knowledgeGit, sourceGit, knowledgeGitErr, sourceGitErr),
		manifestPathsCheck(manifestErr, sourceErr, sourceOK),
		knowledgeDirectoriesCheck(knowledge, knowledgeOK),
		gitAvailableCheck(inspectGit),
		permissionsCheck(checkAccess, manifestPath, knowledge, source, knowledgeOK, sourceOK),
		manifestVersionCheck(manifest.VersionState, manifest.VersionErr),
	}
	status := Healthy
	for _, check := range checks {
		if check.Severity == Error {
			status = Invalid
			break
		}
		if check.Severity == Warning {
			status = Warnings
		}
	}
	return Diagnosis{Root: root, Checks: checks, Status: status}
}

type manifest struct {
	Name         string
	Source       string
	VersionState string
	VersionErr   error
}

func readManifest(path string) (manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return manifest{}, err
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	var raw map[string]json.RawMessage
	if err := decoder.Decode(&raw); err != nil {
		return manifest{}, err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return manifest{}, errors.New("manifesto possui conteúdo extra")
	}

	var out manifest
	if err := decodeString(raw, "name", &out.Name); err != nil {
		return out, err
	}
	if err := ValidateName(out.Name); err != nil {
		return out, err
	}
	if err := decodeString(raw, "source", &out.Source); err != nil {
		return out, err
	}
	if version, ok := raw["version"]; !ok {
		out.VersionState = "implicit"
	} else if strings.TrimSpace(string(version)) == "1" {
		out.VersionState = "explicit"
	} else {
		out.VersionState = "invalid"
		out.VersionErr = errors.New("versão não suportada")
	}
	return out, nil
}

func decodeString(raw map[string]json.RawMessage, key string, target *string) error {
	value, ok := raw[key]
	if !ok {
		return fmt.Errorf("campo %s ausente", key)
	}
	if err := json.Unmarshal(value, target); err != nil || *target == "" {
		return fmt.Errorf("campo %s inválido", key)
	}
	return nil
}

func manifestCheck(root string, data manifest, err error) CheckResult {
	if err != nil {
		return check("manifest", "Manifesto", Error, "inválido ou ilegível", "corrija knowledge/cerne.json")
	}
	if data.Name != filepath.Base(root) {
		return check("manifest", "Manifesto", Warning, "name válido difere do nome da raiz", "alinhe o manifesto ou renomeie o workspace")
	}
	return check("manifest", "Manifesto", Pass, "legível", "")
}

func dirCheck(id, label string, ok bool) CheckResult {
	if !ok {
		return check(id, label, Error, "não encontrado como diretório regular", "crie ou restaure o diretório esperado")
	}
	return check(id, label, Pass, "encontrado", "")
}

func gitIndependenceCheck(inspect GitInspect, knowledge, source string, k RepositoryFacts, s RepositoryFacts, kerr, serr error) CheckResult {
	if inspect == nil {
		return check("git-independence", "Independência Git", Error, "Git indisponível", "instale o Git e disponibilize-o no PATH")
	}
	if kerr != nil || serr != nil || !samePath(k.WorktreeRoot, knowledge) || !samePath(s.WorktreeRoot, source) {
		return check("git-independence", "Independência Git", Error, "raízes Git próprias não confirmadas", "inicialize knowledge e source como repositórios Git independentes")
	}
	if samePath(k.CommonDir, s.CommonDir) {
		return check("git-independence", "Independência Git", Error, "histórico Git compartilhado", "use repositórios Git independentes")
	}
	return check("git-independence", "Independência Git", Pass, "raízes e históricos distintos", "")
}

func versioningIsolationCheck(inspect GitInspect, knowledge, source string, k RepositoryFacts, s RepositoryFacts, kerr, serr error) CheckResult {
	if inspect == nil {
		return check("versioning-isolation", "Isolamento de versionamento", Error, "Git indisponível", "instale o Git e disponibilize-o no PATH")
	}
	if containsPath(knowledge, source) || containsPath(source, knowledge) {
		return check("versioning-isolation", "Isolamento de versionamento", Error, "um repositório contém o outro", "mantenha knowledge e source como diretórios irmãos")
	}
	if kerr != nil || serr != nil {
		return check("versioning-isolation", "Isolamento de versionamento", Error, "limites Git não confirmados", "corrija os repositórios Git locais")
	}
	if containsPath(k.WorktreeRoot, s.WorktreeRoot) || containsPath(s.WorktreeRoot, k.WorktreeRoot) {
		return check("versioning-isolation", "Isolamento de versionamento", Error, "um worktree contém o outro", "separe os repositórios Git")
	}
	return check("versioning-isolation", "Isolamento de versionamento", Pass, "nenhum repositório contém o outro", "")
}

func manifestPathsCheck(manifestErr, sourceErr error, sourceOK bool) CheckResult {
	if manifestErr != nil {
		return check("manifest-paths", "Caminhos do manifesto", Error, "manifesto inválido impede resolver caminhos", "corrija knowledge/cerne.json")
	}
	if sourceErr != nil {
		return check("manifest-paths", "Caminhos do manifesto", Error, "source inválido", "configure um caminho source existente e seguro")
	}
	if !sourceOK {
		return check("manifest-paths", "Caminhos do manifesto", Error, "source não existe como diretório regular", "restaure o diretório de código-fonte")
	}
	return check("manifest-paths", "Caminhos do manifesto", Pass, "válidos", "")
}

func knowledgeDirectoriesCheck(knowledge string, knowledgeOK bool) CheckResult {
	if !knowledgeOK {
		return check("knowledge-directories", "Diretórios obrigatórios", Error, "knowledge indisponível", "restaure o repositório de conhecimento")
	}
	for _, dir := range knowledgeDirectories(knowledge) {
		if err := regularDir(dir); err != nil {
			return check("knowledge-directories", "Diretórios obrigatórios", Error, "diretório obrigatório ausente ou inválido", "restaure product, specs, decisions, policies e runs")
		}
	}
	return check("knowledge-directories", "Diretórios obrigatórios", Pass, "product, specs, decisions, policies e runs encontrados", "")
}

func gitAvailableCheck(inspect GitInspect) CheckResult {
	if inspect == nil {
		return check("git-available", "Git", Error, "indisponível", "instale o Git e disponibilize-o no PATH")
	}
	return check("git-available", "Git", Pass, "disponível", "")
}

func permissionsCheck(access AccessCheck, manifestPath, knowledge, source string, knowledgeOK, sourceOK bool) CheckResult {
	if access == nil {
		return check("permissions", "Permissões", Warning, "não foi possível confirmar permissões", "verifique leitura e escrita manualmente")
	}
	paths := []string{manifestPath}
	if knowledgeOK {
		paths = append(paths, knowledge)
		paths = append(paths, knowledgeDirectories(knowledge)...)
	}
	if sourceOK {
		paths = append(paths, source)
	}
	unknown := false
	for _, path := range paths {
		result := access(path)
		if result.Read == AccessDenied || result.Write == AccessDenied {
			return check("permissions", "Permissões", Error, "leitura ou escrita negada", "ajuste permissões do workspace")
		}
		unknown = unknown || result.Read == AccessUnknown || result.Write == AccessUnknown
	}
	if unknown {
		return check("permissions", "Permissões", Warning, "não foi possível confirmar leitura e escrita", "verifique permissões do workspace")
	}
	return check("permissions", "Permissões", Pass, "leitura e escrita confirmadas", "")
}

func manifestVersionCheck(state string, err error) CheckResult {
	if err != nil || state == "invalid" {
		return check("manifest-version", "Versão do manifesto", Error, "versão não suportada", "use version como inteiro JSON 1 ou remova o campo")
	}
	if state == "explicit" {
		return check("manifest-version", "Versão do manifesto", Pass, "versão 1 suportada", "")
	}
	return check("manifest-version", "Versão do manifesto", Pass, "versão 1 implícita e suportada", "")
}

func validateSourcePath(knowledge, source string) (string, error) {
	if source == "" {
		return "", errors.New("source ausente")
	}
	candidate := source
	if !filepath.IsAbs(candidate) {
		candidate = filepath.Join(knowledge, candidate)
	}
	candidate = filepath.Clean(candidate)
	if err := noSymlink(candidate); err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(candidate)
	if err != nil {
		return "", err
	}
	return filepath.Clean(resolved), nil
}

func regularDir(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return errors.New("não é diretório regular")
	}
	return nil
}

func inspectRepository(inspect GitInspect, path string) (RepositoryFacts, error) {
	if inspect == nil {
		return RepositoryFacts{}, errors.New("Git indisponível")
	}
	return inspect(path)
}

func check(id, label string, severity Severity, detail, correction string) CheckResult {
	return CheckResult{ID: id, Label: label, Severity: severity, Detail: detail, Correction: correction}
}

func canonical(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return filepath.Clean(path)
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		abs = resolved
	}
	return filepath.Clean(abs)
}

func samePath(left, right string) bool {
	left = canonical(left)
	right = canonical(right)
	leftInfo, leftErr := os.Stat(left)
	rightInfo, rightErr := os.Stat(right)
	if leftErr == nil && rightErr == nil {
		return os.SameFile(leftInfo, rightInfo)
	}
	return left == right
}

func containsPath(parent, child string) bool {
	parent = canonical(parent)
	child = canonical(child)
	relative, err := filepath.Rel(parent, child)
	return err == nil && relative != "." && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

func noSymlink(path string) error {
	for {
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return errors.New("link não permitido")
		}
		parent := filepath.Dir(path)
		if parent == path {
			return nil
		}
		path = parent
	}
}
