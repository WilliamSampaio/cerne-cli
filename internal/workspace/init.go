package workspace

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	ErrInvalidName       = errors.New("nome de projeto inválido")
	ErrUnsafeDestination = errors.New("destino inseguro")
)

type Result struct {
	Name          string
	KnowledgePath string
	SourcePath    string
	SourceMode    SourceMode
	AuditPath     string
}

type SourceMode string

const (
	SourceEmpty SourceMode = "empty"
	SourceLocal SourceMode = "local"
	SourceClone SourceMode = "clone"
)

type SourceInitRequest struct {
	Mode              SourceMode
	Input             string
	OriginTransport   string
	OriginFingerprint string
}

type CloneSource func(string, string) error

type SourceInitFailure struct {
	Cause      string
	Correction string
	Incomplete bool
}

func (failure SourceInitFailure) Error() string { return failure.Cause }

func Init(parent, name string, initRepository func(string) error) (result Result, err error) {
	return initWorkspace(parent, name, WorkflowDefinition{}, initRepository)
}

func InitWithSource(parent, name string, request SourceInitRequest, initRepository func(string) error, inspect LinkGitInspect, clone CloneSource) (Result, error) {
	result, _, err := InitWithSourceAndWorkflow(parent, name, request, WorkflowDefinition{}, initRepository, inspect, clone)
	return result, err
}

func InitWithSourceAndWorkflow(parent, name string, request SourceInitRequest, definition WorkflowDefinition, initRepository func(string) error, inspect LinkGitInspect, clone CloneSource) (Result, WorkflowResult, error) {
	return InitWithSourceAndWorkflowAndAgent(parent, name, request, definition, "", initRepository, inspect, clone)
}

func InitWithSourceAndWorkflowAndAgent(parent, name string, request SourceInitRequest, definition WorkflowDefinition, agent string, initRepository func(string) error, inspect LinkGitInspect, clone CloneSource) (Result, WorkflowResult, error) {
	if definition.Provider != "" {
		if _, _, err := workflowPaths(filepath.Join(parent, name, "knowledge"), definition); err != nil {
			return Result{}, WorkflowResult{}, errors.New("workflow inválido")
		}
	}
	result, err := initWithSource(parent, name, request, definition, initRepository, inspect, clone)
	if err != nil || definition.Provider == "" {
		return result, WorkflowResult{}, err
	}
	authorization := "--workflow"
	if agent != "" {
		authorization += " --agent " + agent
	}
	workflow, err := applyWorkflow(result.KnowledgePath, definition, "init", authorization, agent)
	workflow.ProjectName = name
	return result, workflow, err
}

func initWithSource(parent, name string, request SourceInitRequest, definition WorkflowDefinition, initRepository func(string) error, inspect LinkGitInspect, clone CloneSource) (Result, error) {
	switch request.Mode {
	case SourceLocal:
		return initWithLocalSource(parent, name, request.Input, definition, initRepository, inspect)
	case SourceClone:
		return initWithClonedSource(parent, name, request, definition, initRepository, inspect, clone)
	default:
		return Result{}, errors.New("modo de source inválido")
	}
}

func InitWithWorkflow(parent, name string, definition WorkflowDefinition, initRepository func(string) error) (Result, WorkflowResult, error) {
	return InitWithWorkflowAndAgent(parent, name, definition, "", initRepository)
}

func InitWithWorkflowAndAgent(parent, name string, definition WorkflowDefinition, agent string, initRepository func(string) error) (Result, WorkflowResult, error) {
	if definition.Provider == "" {
		return Result{}, WorkflowResult{}, errors.New("workflow inválido")
	}
	if _, _, err := workflowPaths(filepath.Join(parent, name, "knowledge"), definition); err != nil {
		return Result{}, WorkflowResult{}, errors.New("workflow inválido")
	}
	result, err := initWorkspace(parent, name, definition, initRepository)
	if err != nil {
		return Result{}, WorkflowResult{}, err
	}
	authorization := "--workflow"
	if agent != "" {
		authorization += " --agent " + agent
	}
	workflow, err := applyWorkflow(result.KnowledgePath, definition, "init", authorization, agent)
	workflow.ProjectName = name
	return result, workflow, err
}

func initWorkspace(parent, name string, definition WorkflowDefinition, initRepository func(string) error) (result Result, err error) {
	return initWorkspaceMode(parent, name, definition, "../source", true, initRepository)
}

func initWorkspaceMode(parent, name string, definition WorkflowDefinition, manifestSource string, createSource bool, initRepository func(string) error) (result Result, err error) {
	if err := ValidateName(name); err != nil {
		return Result{}, err
	}

	root, err := filepath.Abs(filepath.Join(parent, name))
	if err != nil {
		return Result{}, fmt.Errorf("não foi possível resolver o destino: %w", err)
	}

	var created []string
	defer func() {
		if err == nil {
			return
		}
		if rollbackErr := removeCreated(created); rollbackErr != nil {
			err = fmt.Errorf("%w; falha no rollback: %v", err, rollbackErr)
		}
	}()

	if info, err := os.Lstat(root); err == nil {
		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return Result{}, fmt.Errorf("%w: %q não é um diretório regular", ErrUnsafeDestination, root)
		}
		empty, err := isEmpty(root)
		if err != nil {
			return Result{}, fmt.Errorf("não foi possível inspecionar o destino %q: %w", root, err)
		}
		if !empty {
			return Result{}, fmt.Errorf("%w: %q não está vazio", ErrUnsafeDestination, root)
		}
	} else if os.IsNotExist(err) {
		if err := os.Mkdir(root, 0o755); err != nil {
			return Result{}, fmt.Errorf("não foi possível criar o destino %q: %w", root, err)
		}
		created = append(created, root)
	} else {
		return Result{}, fmt.Errorf("não foi possível inspecionar o destino %q: %w", root, err)
	}

	knowledge := filepath.Join(root, "knowledge")
	source := filepath.Join(root, "source")
	knowledgeDirs := commonKnowledgeDirectories(knowledge)
	if definition.Provider == "" || filepath.Clean(filepath.FromSlash(definition.CanonicalSpecs)) == "specs" {
		knowledgeDirs = append(knowledgeDirs, filepath.Join(knowledge, "specs"))
	}
	directories := append([]string{knowledge}, knowledgeDirs...)
	if createSource {
		directories = append([]string{knowledge, source}, knowledgeDirs...)
	}
	for _, directory := range directories {
		if err := os.Mkdir(directory, 0o755); err != nil {
			return Result{}, fmt.Errorf("não foi possível criar %q: %w", directory, err)
		}
		created = append(created, directory)
	}
	for _, directory := range knowledgeDirs {
		if err := os.WriteFile(filepath.Join(directory, ".gitkeep"), nil, 0o644); err != nil {
			return Result{}, fmt.Errorf("não foi possível preservar %q no Git: %w", directory, err)
		}
	}

	manifestPath := filepath.Join(knowledge, "cerne.json")
	manifest, err := os.OpenFile(manifestPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return Result{}, fmt.Errorf("não foi possível criar o manifesto: %w", err)
	}
	created = append(created, manifestPath)
	encoder := json.NewEncoder(manifest)
	encoder.SetIndent("", "  ")
	manifestData := struct {
		Name     string `json:"name"`
		Source   string `json:"source"`
		Workflow *struct {
			Provider string `json:"provider"`
		} `json:"workflow,omitempty"`
	}{Name: name, Source: manifestSource}
	if definition.Provider != "" {
		manifestData.Workflow = &struct {
			Provider string `json:"provider"`
		}{Provider: definition.Provider}
	}
	writeErr := encoder.Encode(manifestData)
	closeErr := manifest.Close()
	if writeErr != nil {
		return Result{}, fmt.Errorf("não foi possível gravar o manifesto: %w", writeErr)
	}
	if closeErr != nil {
		return Result{}, fmt.Errorf("não foi possível concluir o manifesto: %w", closeErr)
	}

	repositories := []string{knowledge}
	if createSource {
		repositories = append(repositories, source)
	}
	for _, repository := range repositories {
		if err := initRepository(repository); err != nil {
			return Result{}, err
		}
	}

	return Result{Name: name, KnowledgePath: knowledge, SourcePath: source, SourceMode: SourceEmpty}, nil
}

func initWithLocalSource(parent, name, input string, definition WorkflowDefinition, initRepository func(string) error, inspect LinkGitInspect) (Result, error) {
	if inspect == nil {
		return Result{}, sourceInitFailure("Git indisponível", "instale o Git e disponibilize-o no PATH", false)
	}
	candidate, err := resolveLinkPath(parent, input)
	if err != nil {
		return Result{}, err
	}
	before, err := validLinkRepository(inspect, candidate)
	if err != nil {
		return Result{}, err
	}
	root := filepath.Join(canonical(parent), name)
	knowledge := filepath.Join(root, "knowledge")
	if samePath(candidate, knowledge) || containsPath(candidate, knowledge) || containsPath(knowledge, candidate) {
		return Result{}, linkFailure("sobreposição perigosa entre knowledge e source", candidate, "mantenha os repositórios em diretórios separados")
	}
	rootExisted := pathExists(root)
	manifestSource := manifestLinkSource(knowledge, candidate)
	result, err := initWorkspaceMode(parent, name, definition, manifestSource, false, initRepository)
	if err != nil {
		return Result{}, err
	}
	rollback := func(failure error) (Result, error) {
		if rollbackErr := rollbackInitializedWorkspace(result, rootExisted); rollbackErr != nil {
			failure = fmt.Errorf("%w; falha no rollback: %v", failure, rollbackErr)
		}
		return Result{}, failure
	}
	after, err := validLinkRepository(inspect, candidate)
	if err != nil || !sameRepositoryFacts(before, after) {
		return rollback(sourceInitFailure("source local mudou durante a inicialização", "verifique o repositório e tente novamente", false))
	}
	knowledgeFacts, err := validLinkRepository(inspect, result.KnowledgePath)
	if err != nil {
		return rollback(sourceInitFailure("repositório knowledge inválido", "verifique o Git e tente novamente", false))
	}
	if err := validateLinkSeparation(result.KnowledgePath, candidate, knowledgeFacts, after); err != nil {
		return rollback(err)
	}
	result.SourcePath = candidate
	result.SourceMode = SourceLocal
	return result, nil
}

func sameRepositoryFacts(left, right LinkRepositoryFacts) bool {
	return left.HasWorktree == right.HasWorktree && left.IsBare == right.IsBare &&
		samePath(left.WorktreeRoot, right.WorktreeRoot) && samePath(left.CommonDir, right.CommonDir)
}

func initWithClonedSource(parent, name string, request SourceInitRequest, definition WorkflowDefinition, initRepository func(string) error, inspect LinkGitInspect, clone CloneSource) (Result, error) {
	if inspect == nil || clone == nil || request.Input == "" || request.OriginTransport == "" || request.OriginFingerprint == "" {
		return Result{}, sourceInitFailure("clone do source indisponível", "instale o Git e tente novamente", false)
	}
	root := filepath.Join(canonical(parent), name)
	rootExisted := pathExists(root)
	result, err := initWorkspaceMode(parent, name, definition, "../source", false, initRepository)
	if err != nil {
		return Result{}, err
	}
	result.SourceMode = SourceClone
	auditPath, attempt, err := startCloneAudit(result.KnowledgePath, name, request)
	if err != nil {
		if rollbackErr := rollbackInitializedWorkspace(result, rootExisted); rollbackErr != nil {
			err = fmt.Errorf("%w; falha no rollback: %v", err, rollbackErr)
		}
		return Result{}, sourceInitFailure("não foi possível registrar a tentativa de clone", "verifique as permissões de knowledge/runs", false)
	}
	result.AuditPath = auditPath
	staging, err := os.MkdirTemp(root, ".source-clone-")
	if err != nil {
		_ = finishCloneAudit(auditPath, attempt, "failed", "staging-failed")
		return result, incompleteCloneFailure()
	}
	if err := os.Chmod(staging, 0o700); err != nil {
		_ = cleanupCloneStaging(root, staging)
		_ = finishCloneAudit(auditPath, attempt, "failed", "staging-failed")
		return result, incompleteCloneFailure()
	}
	fail := func(category string) (Result, error) {
		if err := cleanupCloneStaging(root, staging); err != nil {
			category = "cleanup-failed"
		}
		if err := finishCloneAudit(auditPath, attempt, "failed", category); err != nil {
			return result, incompleteCloneFailure()
		}
		return result, incompleteCloneFailure()
	}
	if err := clone(request.Input, staging); err != nil {
		return fail("clone-failed")
	}
	cloneFacts, err := validLinkRepository(inspect, staging)
	if err != nil {
		return fail("invalid-result")
	}
	knowledgeFacts, err := validLinkRepository(inspect, result.KnowledgePath)
	if err != nil || validateLinkSeparation(result.KnowledgePath, staging, knowledgeFacts, cloneFacts) != nil {
		return fail("invalid-result")
	}
	if err := promoteDirectoryNoReplace(staging, result.SourcePath); err != nil {
		return fail("promotion-failed")
	}
	staging = ""
	if err := finishCloneAudit(auditPath, attempt, "succeeded", ""); err != nil {
		return result, incompleteCloneFailure()
	}
	return result, nil
}

type cloneAttempt struct {
	Kind              string `json:"kind"`
	Executor          string `json:"executor"`
	Operation         string `json:"operation"`
	Project           string `json:"project"`
	Destination       string `json:"destination"`
	OriginTransport   string `json:"origin_transport"`
	OriginFingerprint string `json:"origin_fingerprint"`
	Authorization     string `json:"authorization"`
	Status            string `json:"status"`
	StartedAt         string `json:"started_at"`
	FinishedAt        string `json:"finished_at,omitempty"`
	Failure           string `json:"failure,omitempty"`
}

var openCloneAudit = os.OpenFile
var removeCloneStaging = os.RemoveAll

func startCloneAudit(knowledge, project string, request SourceInitRequest) (string, cloneAttempt, error) {
	attempt := cloneAttempt{
		Kind: "source-clone", Executor: "git", Operation: "clone", Project: project,
		Destination: "../source", OriginTransport: request.OriginTransport,
		OriginFingerprint: request.OriginFingerprint, Authorization: "--clone",
		Status: "started", StartedAt: time.Now().UTC().Format(time.RFC3339Nano),
	}
	path := filepath.Join(knowledge, "runs", "source-clone.json")
	file, err := openCloneAudit(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return "", attempt, err
	}
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(attempt); err != nil {
		file.Close()
		os.Remove(path)
		return "", attempt, err
	}
	if err := file.Sync(); err != nil {
		file.Close()
		os.Remove(path)
		return "", attempt, err
	}
	if err := file.Close(); err != nil {
		os.Remove(path)
		return "", attempt, err
	}
	return path, attempt, nil
}

func finishCloneAudit(path string, attempt cloneAttempt, status, failure string) error {
	attempt.Status = status
	attempt.FinishedAt = time.Now().UTC().Format(time.RFC3339Nano)
	attempt.Failure = failure
	content, err := json.MarshalIndent(attempt, "", "  ")
	if err != nil {
		return err
	}
	return writeManifestAtomically(path, append(content, '\n'))
}

func cleanupCloneStaging(root, staging string) error {
	if staging == "" || filepath.Dir(filepath.Clean(staging)) != filepath.Clean(root) ||
		!strings.HasPrefix(filepath.Base(staging), ".source-clone-") {
		return errors.New("ownership do staging não confirmada")
	}
	info, err := os.Lstat(staging)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return errors.New("ownership do staging não confirmada")
	}
	return removeCloneStaging(staging)
}

func rollbackInitializedWorkspace(result Result, rootExisted bool) error {
	root := filepath.Dir(result.KnowledgePath)
	if !rootExisted {
		return os.RemoveAll(root)
	}
	return os.RemoveAll(result.KnowledgePath)
}

func pathExists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}

func sourceInitFailure(cause, correction string, incomplete bool) SourceInitFailure {
	return SourceInitFailure{Cause: cause, Correction: correction, Incomplete: incomplete}
}

func incompleteCloneFailure() SourceInitFailure {
	return sourceInitFailure("não foi possível concluir o clone do source", "inspecione knowledge/runs/source-clone.json e associe um source válido ou remova o workspace incompleto antes de repetir o init", true)
}

func ValidateName(name string) error {
	if len(name) == 0 || len(name) > 255 || !asciiAlphaNumeric(name[0]) {
		return fmt.Errorf("%w: use de 1 a 255 caracteres ASCII e comece por letra ou número",
			ErrInvalidName)
	}
	for index := 1; index < len(name); index++ {
		character := name[index]
		if !asciiAlphaNumeric(character) && character != '.' && character != '_' && character != '-' {
			return fmt.Errorf("%w: use somente letras ASCII, números, ponto, hífen ou sublinhado",
				ErrInvalidName)
		}
	}
	if name[len(name)-1] == '.' {
		return fmt.Errorf("%w: o nome não pode terminar em ponto", ErrInvalidName)
	}
	stem := strings.ToUpper(strings.SplitN(name, ".", 2)[0])
	if stem == "CON" || stem == "PRN" || stem == "AUX" || stem == "NUL" ||
		len(stem) == 4 && (strings.HasPrefix(stem, "COM") || strings.HasPrefix(stem, "LPT")) &&
			stem[3] >= '1' && stem[3] <= '9' {
		return fmt.Errorf("%w: %q é reservado nos sistemas suportados", ErrInvalidName, name)
	}
	return nil
}

func asciiAlphaNumeric(character byte) bool {
	return character >= 'a' && character <= 'z' ||
		character >= 'A' && character <= 'Z' ||
		character >= '0' && character <= '9'
}

func removeCreated(paths []string) error {
	var rollbackErr error
	for index := len(paths) - 1; index >= 0; index-- {
		if err := os.RemoveAll(paths[index]); err != nil {
			rollbackErr = errors.Join(rollbackErr, fmt.Errorf("remover %q: %w", paths[index], err))
		}
	}
	return rollbackErr
}

func knowledgeDirectories(root string) []string {
	names := []string{"product", "specs", "decisions", "policies", "runs"}
	return namedDirectories(root, names)
}

func commonKnowledgeDirectories(root string) []string {
	return namedDirectories(root, []string{"product", "decisions", "policies", "runs"})
}

func namedDirectories(root string, names []string) []string {
	directories := make([]string, len(names))
	for index, name := range names {
		directories[index] = filepath.Join(root, name)
	}
	return directories
}

func isEmpty(directory string) (bool, error) {
	handle, err := os.Open(directory)
	if err != nil {
		return false, err
	}
	defer handle.Close()
	_, err = handle.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}
