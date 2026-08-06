package workspace

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type RestoreRequest struct {
	KnowledgeOrigin string
	SourceMode      SourceMode
	SourceInput     string
}

type RestoreResult struct {
	Name            string
	KnowledgePath   string
	SourcePath      string
	SourceMode      SourceMode
	ManifestChanged bool
	AuditPath       string
}

type RestoreFailure struct {
	Cause      string
	Correction string
}

func (failure RestoreFailure) Error() string { return failure.Cause }

type restorePhase struct {
	Operation  string `json:"operation"`
	Status     string `json:"status"`
	StartedAt  string `json:"started_at,omitempty"`
	FinishedAt string `json:"finished_at,omitempty"`
	Failure    string `json:"failure,omitempty"`
}

type restoreAttempt struct {
	Kind          string        `json:"kind"`
	Executor      string        `json:"executor"`
	Operation     string        `json:"operation"`
	Authorization string        `json:"authorization"`
	SourceMode    SourceMode    `json:"source_mode"`
	WorkspaceName string        `json:"workspace_name,omitempty"`
	Status        string        `json:"status"`
	StartedAt     string        `json:"started_at"`
	FinishedAt    string        `json:"finished_at,omitempty"`
	Phases        restorePhases `json:"phases"`
}

type restorePhases struct {
	Knowledge restorePhase `json:"knowledge"`
	Source    restorePhase `json:"source"`
}

var replaceRestoreAudit = atomicReplaceFile

func Restore(parent, home string, request RestoreRequest, inspect LinkGitInspect, clone CloneSource, resolve WorkflowResolver) (result RestoreResult, err error) {
	parent = canonical(parent)
	if request.KnowledgeOrigin == "" || request.SourceInput == "" || clone == nil || inspect == nil ||
		request.SourceMode != SourceClone && request.SourceMode != SourceLocal {
		return result, restoreFailure("restauração indisponível", "instale o Git e informe uma origem e um source válidos")
	}
	if err := regularDir(parent); err != nil {
		return result, restoreFailure("destino da restauração é inválido", "execute o comando em um diretório regular e acessível")
	}

	localSource := ""
	if request.SourceMode == SourceLocal {
		localSource, err = resolveRestoreLocalSource(parent, home, request.SourceInput)
		if err != nil {
			return result, err
		}
	}

	audit, attempt, err := startRestoreAudit(home, request.SourceMode)
	if err != nil {
		return result, restoreFailure("não foi possível registrar a tentativa de restauração", "verifique a segurança e as permissões de ~/.cerne/audit")
	}
	result.AuditPath = audit.path
	defer audit.root.Close()
	staging := ""
	var stagingInfo os.FileInfo
	finalRoot := ""
	fail := func(category, cause, correction string) (RestoreResult, error) {
		auditWritable := category != "audit-failed" && category != "audit-finalization-failed"
		cleanupFailed := false
		if finalRoot != "" && stagingInfo != nil {
			if cleanupOwnedRestore(finalRoot, filepath.Dir(finalRoot), filepath.Base(finalRoot), stagingInfo) != nil {
				category = "cleanup-failed"
				cleanupFailed = true
			}
		} else if staging != "" && stagingInfo != nil {
			if cleanupOwnedRestore(staging, parent, ".cerne-restore-", stagingInfo) != nil {
				category = "cleanup-failed"
				cleanupFailed = true
			}
		}
		attempt.Status = "failed"
		attempt.FinishedAt = restoreTimestamp()
		finishRestorePhases(&attempt, category)
		if auditWritable {
			_ = audit.write(attempt)
		}
		if cleanupFailed {
			cause = "restauração falhou e a limpeza não pôde ser confirmada"
			correction = "preserve o alvo concorrente e inspecione ~/.cerne/audit antes de agir manualmente"
		}
		return result, restoreFailure(cause, correction)
	}

	staging, err = os.MkdirTemp(parent, ".cerne-restore-")
	if err != nil {
		return fail("staging-failed", "não foi possível preparar a restauração", "verifique as permissões do diretório atual")
	}
	stagingInfo, err = os.Lstat(staging)
	if err != nil {
		return fail("staging-failed", "não foi possível confirmar a restauração", "verifique o sistema de arquivos")
	}
	if err := os.Chmod(staging, 0o700); err != nil {
		return fail("staging-failed", "não foi possível proteger a restauração", "verifique as permissões do diretório atual")
	}

	knowledge := filepath.Join(staging, "knowledge")
	startRestorePhase(&attempt.Phases.Knowledge, "clone")
	if err := audit.write(attempt); err != nil {
		return fail("audit-failed", "não foi possível atualizar a auditoria", "verifique ~/.cerne/audit")
	}
	if err := clone(request.KnowledgeOrigin, knowledge); err != nil {
		attempt.Phases.Knowledge.Failure = "clone-failed"
		return fail("knowledge-clone-failed", "não foi possível clonar knowledge", "verifique acesso à origem e tente novamente")
	}
	manifestPath := filepath.Join(knowledge, "cerne.json")
	data, err := validateRestoreKnowledge(knowledge, manifestPath, inspect, resolve)
	if err != nil {
		attempt.Phases.Knowledge.Failure = "invalid-result"
		return fail("knowledge-invalid", "knowledge restaurado é inválido", "corrija o repositório de knowledge e tente novamente")
	}
	attempt.WorkspaceName = data.Name
	finishRestorePhase(&attempt.Phases.Knowledge)

	finalRoot = filepath.Join(parent, data.Name)
	if _, err := os.Lstat(finalRoot); !errors.Is(err, os.ErrNotExist) {
		return fail("destination-exists", "destino do workspace já existe", "escolha outro diretório ou remova manualmente o destino existente")
	}
	finalKnowledge := filepath.Join(finalRoot, "knowledge")
	var source string
	startRestorePhase(&attempt.Phases.Source, map[SourceMode]string{SourceClone: "clone", SourceLocal: "link"}[request.SourceMode])
	if err := audit.write(attempt); err != nil {
		return fail("audit-failed", "não foi possível atualizar a auditoria", "verifique ~/.cerne/audit")
	}

	if request.SourceMode == SourceClone {
		relative, relativeErr := portableRestoreSource(data.Source)
		if relativeErr != nil {
			attempt.Phases.Source.Failure = "invalid-path"
			return fail("source-path-invalid", "caminho source do manifesto é inseguro", "use um caminho portátil ../<diretório> dentro do workspace")
		}
		source = filepath.Clean(filepath.Join(knowledge, filepath.FromSlash(relative)))
		if !containsPath(staging, source) || containsPath(knowledge, source) || samePath(staging, source) {
			return fail("source-path-invalid", "caminho source do manifesto é inseguro", "use um diretório irmão de knowledge dentro do workspace")
		}
		if _, err := os.Lstat(source); !errors.Is(err, os.ErrNotExist) {
			return fail("source-destination-exists", "destino do source já existe", "remova o artefato do repositório knowledge")
		}
		if err := os.MkdirAll(filepath.Dir(source), 0o700); err != nil {
			return fail("source-staging-failed", "não foi possível preparar o source", "verifique o manifesto e as permissões")
		}
		if err := clone(request.SourceInput, source); err != nil {
			attempt.Phases.Source.Failure = "clone-failed"
			return fail("source-clone-failed", "não foi possível clonar source", "verifique acesso à origem e tente novamente")
		}
		sourceFacts, sourceErr := validLinkRepository(inspect, source)
		knowledgeFacts, knowledgeErr := validLinkRepository(inspect, knowledge)
		if sourceErr != nil || knowledgeErr != nil || validateLinkSeparation(knowledge, source, knowledgeFacts, sourceFacts) != nil {
			return fail("source-invalid", "source restaurado é inválido", "use repositórios Git independentes")
		}
		result.SourcePath = filepath.Join(finalRoot, strings.TrimPrefix(filepath.ToSlash(relative), "../"))
	} else {
		before, sourceErr := validLinkRepository(inspect, localSource)
		knowledgeFacts, knowledgeErr := validLinkRepository(inspect, knowledge)
		if sourceErr != nil || knowledgeErr != nil || validateLinkSeparation(finalKnowledge, localSource, knowledgeFacts, before) != nil {
			return fail("source-invalid", "source local é inválido ou inseguro", "informe a raiz de um repositório Git local independente")
		}
		newSource := manifestLinkSource(finalKnowledge, localSource)
		if newSource != data.Source {
			manifest, readErr := readLinkManifest(manifestPath)
			if readErr != nil {
				return fail("manifest-update-failed", "manifesto não pode ser atualizado com segurança", "corrija knowledge/cerne.json")
			}
			manifest.raw["source"], _ = json.Marshal(newSource)
			content, marshalErr := json.MarshalIndent(manifest.raw, "", "  ")
			if marshalErr != nil || writeManifestAtomically(manifestPath, append(content, '\n')) != nil {
				return fail("manifest-update-failed", "manifesto não pode ser atualizado com segurança", "verifique permissões e tente novamente")
			}
			result.ManifestChanged = true
		}
		after, sourceErr := validLinkRepository(inspect, localSource)
		if sourceErr != nil || !sameRepositoryFacts(before, after) {
			return fail("source-changed", "source local mudou durante a restauração", "estabilize o repositório e tente novamente")
		}
		result.SourcePath = localSource
	}
	finishRestorePhase(&attempt.Phases.Source)

	if err := promoteDirectoryNoReplace(staging, finalRoot); err != nil {
		finalRoot = ""
		return fail("promotion-failed", "não foi possível promover o workspace", "verifique se o destino foi criado concorrentemente")
	}
	staging = ""
	result.Name, result.KnowledgePath, result.SourceMode = data.Name, finalKnowledge, request.SourceMode
	if _, err := os.Lstat(filepath.Join(finalRoot, ".git")); !errors.Is(err, os.ErrNotExist) {
		return fail("post-validation-failed", "workspace promovido falhou na validação", "inspecione o sistema de arquivos e tente novamente")
	}
	finalManifest, err := validateRestoreKnowledge(finalKnowledge, filepath.Join(finalKnowledge, "cerne.json"), inspect, resolve)
	if err != nil {
		return fail("post-validation-failed", "workspace promovido falhou na validação", "inspecione o sistema de arquivos e tente novamente")
	}
	resolvedSource, err := validateSourcePath(finalKnowledge, finalManifest.Source)
	if err != nil || !samePath(resolvedSource, result.SourcePath) {
		return fail("post-validation-failed", "workspace promovido falhou na validação", "inspecione o sistema de arquivos e tente novamente")
	}
	result.SourcePath = resolvedSource
	knowledgeFacts, knowledgeErr := validLinkRepository(inspect, finalKnowledge)
	sourceFacts, sourceErr := validLinkRepository(inspect, result.SourcePath)
	if knowledgeErr != nil || sourceErr != nil || validateLinkSeparation(finalKnowledge, result.SourcePath, knowledgeFacts, sourceFacts) != nil {
		return fail("post-validation-failed", "workspace promovido falhou na validação", "inspecione o sistema de arquivos e tente novamente")
	}
	attempt.Status, attempt.FinishedAt = "succeeded", restoreTimestamp()
	if err := audit.write(attempt); err != nil {
		return fail("audit-finalization-failed", "não foi possível finalizar a auditoria", "verifique ~/.cerne/audit antes de tentar novamente")
	}
	finalRoot = ""
	return result, nil
}

func validateRestoreKnowledge(knowledge, manifestPath string, inspect LinkGitInspect, resolve WorkflowResolver) (manifest, error) {
	info, err := os.Lstat(manifestPath)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return manifest{}, errors.New("manifesto irregular")
	}
	data, err := readManifest(manifestPath)
	if err != nil || data.VersionErr != nil || data.WorkflowErr != nil {
		return manifest{}, errors.New("manifesto inválido")
	}
	if _, err := validLinkRepository(inspect, knowledge); err != nil {
		return manifest{}, err
	}
	directories := commonKnowledgeDirectories(knowledge)
	if !data.WorkflowDeclared {
		directories = append(directories, filepath.Join(knowledge, "specs"))
	}
	for _, directory := range directories {
		if regularDir(directory) != nil || noSymlink(directory) != nil {
			return manifest{}, errors.New("diretório obrigatório ausente")
		}
	}
	if data.WorkflowDeclared {
		if resolve == nil {
			return manifest{}, errors.New("provider desconhecido")
		}
		definition, err := resolve(data.WorkflowProvider)
		if err != nil {
			return manifest{}, err
		}
		root, marker, err := workflowPaths(knowledge, definition)
		if err != nil {
			return manifest{}, err
		}
		state, err := workflowLayout(root, marker)
		if err != nil || state == WorkflowUnchanged && !workflowSpecsValid(knowledge, root, definition) {
			return manifest{}, errors.New("workflow parcial")
		}
	}
	return data, nil
}

func portableRestoreSource(source string) (string, error) {
	if source == "" || strings.Contains(source, `\`) || strings.Contains(source, ":") ||
		strings.HasPrefix(source, "/") || !strings.HasPrefix(source, "../") || path.Clean(source) != source {
		return "", errors.New("source não portátil")
	}
	segments := strings.Split(strings.TrimPrefix(source, "../"), "/")
	if len(segments) == 0 {
		return "", errors.New("source ausente")
	}
	for _, segment := range segments {
		if segment == "" || segment == "." || segment == ".." || ValidateName(segment) != nil {
			return "", errors.New("segmento inválido")
		}
	}
	return source, nil
}

func resolveRestoreLocalSource(parent, home, input string) (string, error) {
	source, err := resolveLinkPath(parent, input)
	if err != nil {
		return "", restoreFailure("source local inválido", "informe a raiz de um repositório Git local existente")
	}
	auditRoot := filepath.Join(canonical(home), ".cerne")
	if samePath(source, parent) || containsPath(source, parent) || containsPath(parent, source) ||
		samePath(source, auditRoot) || containsPath(source, auditRoot) || containsPath(auditRoot, source) {
		return "", restoreFailure("source local sobrepõe área protegida", "use um repositório fora do destino e de ~/.cerne")
	}
	return source, nil
}

type restoreAudit struct {
	root *os.Root
	home string
	path string
}

func startRestoreAudit(home string, mode SourceMode) (restoreAudit, restoreAttempt, error) {
	home = canonical(home)
	root, err := os.OpenRoot(home)
	if err != nil {
		return restoreAudit{}, restoreAttempt{}, err
	}
	closeOnError := true
	defer func() {
		if closeOnError {
			root.Close()
		}
	}()
	for _, directory := range []string{".cerne", ".cerne/audit"} {
		info, err := root.Lstat(directory)
		if errors.Is(err, os.ErrNotExist) {
			if err := root.Mkdir(directory, 0o700); err != nil {
				return restoreAudit{}, restoreAttempt{}, err
			}
			info, err = root.Lstat(directory)
		}
		if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return restoreAudit{}, restoreAttempt{}, errors.New("diretório de audit inseguro")
		}
		if err := secureAuditPath(filepath.Join(home, filepath.FromSlash(directory)), true); err != nil {
			return restoreAudit{}, restoreAttempt{}, err
		}
	}
	id, err := restoreID()
	if err != nil {
		return restoreAudit{}, restoreAttempt{}, err
	}
	rel := ".cerne/audit/restore-" + id + ".json"
	file, err := root.OpenFile(rel, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return restoreAudit{}, restoreAttempt{}, err
	}
	if err := file.Close(); err != nil {
		return restoreAudit{}, restoreAttempt{}, err
	}
	if err := secureAuditPath(filepath.Join(home, filepath.FromSlash(rel)), false); err != nil {
		return restoreAudit{}, restoreAttempt{}, err
	}
	audit := restoreAudit{root: root, home: home, path: filepath.Join(home, filepath.FromSlash(rel))}
	attempt := restoreAttempt{
		Kind: "workspace-restore", Executor: "cerne", Operation: "restore",
		Authorization: "restore --" + string(mode), SourceMode: mode,
		Status: "started", StartedAt: restoreTimestamp(),
		Phases: restorePhases{
			Knowledge: restorePhase{Operation: "clone", Status: "pending"},
			Source:    restorePhase{Operation: map[SourceMode]string{SourceClone: "clone", SourceLocal: "link"}[mode], Status: "pending"},
		},
	}
	if err := audit.write(attempt); err != nil {
		return restoreAudit{}, restoreAttempt{}, err
	}
	closeOnError = false
	return audit, attempt, nil
}

func (audit restoreAudit) write(attempt restoreAttempt) error {
	content, err := json.MarshalIndent(attempt, "", "  ")
	if err != nil {
		return err
	}
	tempID, err := restoreID()
	if err != nil {
		return err
	}
	tempRel := ".cerne/audit/.restore-" + tempID + ".tmp"
	file, err := audit.root.OpenFile(tempRel, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	tempPath := filepath.Join(audit.home, filepath.FromSlash(tempRel))
	defer audit.root.Remove(tempRel)
	if _, err := file.Write(append(content, '\n')); err != nil {
		file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	if err := secureAuditPath(tempPath, false); err != nil {
		return err
	}
	return replaceRestoreAudit(tempPath, audit.path)
}

func restoreID() (string, error) {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return hex.EncodeToString(value), nil
}

func restoreTimestamp() string { return time.Now().UTC().Format(time.RFC3339Nano) }

func startRestorePhase(phase *restorePhase, operation string) {
	phase.Operation, phase.Status, phase.StartedAt = operation, "started", restoreTimestamp()
}

func finishRestorePhase(phase *restorePhase) {
	phase.Status, phase.FinishedAt = "succeeded", restoreTimestamp()
}

func finishRestorePhases(attempt *restoreAttempt, category string) {
	for _, phase := range []*restorePhase{&attempt.Phases.Knowledge, &attempt.Phases.Source} {
		if phase.Status == "started" {
			phase.Status, phase.FinishedAt = "failed", restoreTimestamp()
			if phase.Failure == "" {
				phase.Failure = category
			}
		}
	}
}

func cleanupOwnedRestore(target, parent, prefix string, expected os.FileInfo) error {
	clean := filepath.Clean(target)
	if filepath.Dir(clean) != filepath.Clean(parent) || !strings.HasPrefix(filepath.Base(clean), prefix) {
		return errors.New("ownership do restore não confirmada")
	}
	info, err := os.Lstat(clean)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.IsDir() || !os.SameFile(info, expected) {
		return errors.New("ownership do restore não confirmada")
	}
	return os.RemoveAll(clean)
}

func restoreFailure(cause, correction string) RestoreFailure {
	return RestoreFailure{Cause: cause, Correction: correction}
}
