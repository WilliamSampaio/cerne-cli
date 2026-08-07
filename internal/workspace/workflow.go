package workspace

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type WorkflowDefinition struct {
	Provider       string
	Executor       string
	CanonicalSpecs string
	OwnedRoot      string
	Marker         string
	Available      bool
	Setup          func(string) error
	Agents         map[string]WorkflowAgentTarget
}

type WorkflowAgentTarget struct {
	Name          string
	DiscoveryRoot string
	Setup         func(string) error
}

type WorkflowResolver func(string) (WorkflowDefinition, error)

type WorkflowState string

const (
	WorkflowPending    WorkflowState = "pending"
	WorkflowConfigured WorkflowState = "configured"
	WorkflowUnchanged  WorkflowState = "unchanged"
)

type WorkflowResult struct {
	ProjectName   string
	KnowledgePath string
	Provider      string
	Executor      string
	State         WorkflowState
	AuditPath     string
	Agent         string
	Discovery     WorkflowDiscoveryState
}

type WorkflowFailure struct {
	Cause      string
	Correction string
}

func (failure WorkflowFailure) Error() string { return failure.Cause }

type workflowAttempt struct {
	Kind          string `json:"kind"`
	Provider      string `json:"provider"`
	Executor      string `json:"executor"`
	Operation     string `json:"operation"`
	Context       string `json:"context"`
	Authorization string `json:"authorization"`
	Status        string `json:"status"`
	StartedAt     string `json:"started_at"`
	FinishedAt    string `json:"finished_at,omitempty"`
	Failure       string `json:"failure,omitempty"`
}

var replaceWorkflowAudit = atomicReplaceFile

type WorkflowDiscoveryState string

const (
	WorkflowDiscoveryReady      WorkflowDiscoveryState = "ready"
	WorkflowDiscoveryUnchanged  WorkflowDiscoveryState = "unchanged"
	WorkflowDiscoveryNotCreated WorkflowDiscoveryState = "not-created"
)

var specKitBridgeCommands = []string{
	"speckit-analyze",
	"speckit-checklist",
	"speckit-clarify",
	"speckit-constitution",
	"speckit-converge",
	"speckit-implement",
	"speckit-plan",
	"speckit-specify",
	"speckit-tasks",
	"speckit-taskstoissues",
}

func SetupWorkflow(start string, resolve WorkflowResolver) (WorkflowResult, error) {
	return SetupWorkflowWithAgent(start, resolve, "")
}

func SetupWorkflowWithAgent(start string, resolve WorkflowResolver, agent string) (WorkflowResult, error) {
	root, manifestPath, err := locateWorkspace(start)
	if err != nil {
		return WorkflowResult{}, workflowFailure("workspace Cerne não localizado", "execute o comando dentro de um workspace Cerne")
	}
	data, err := readManifest(manifestPath)
	if err != nil || data.VersionErr != nil || data.WorkflowErr != nil {
		return WorkflowResult{}, workflowFailure("manifesto ausente ou inválido", "corrija ou restaure knowledge/cerne.json")
	}
	if !data.WorkflowDeclared {
		return WorkflowResult{}, workflowFailure("workflow não configurado no manifesto", "crie um novo workspace com \"cerne init <project-name> --workflow <speckit|openspec>\"")
	}
	knowledge := filepath.Join(root, "knowledge")
	source, sourceErr := validateSourcePath(knowledge, data.Source)
	if regularDir(knowledge) != nil || sourceErr != nil || regularDir(source) != nil || samePath(knowledge, source) || containsPath(knowledge, source) || containsPath(source, knowledge) {
		return WorkflowResult{}, workflowFailure("workspace Cerne inválido", "execute cerne doctor e corrija o workspace antes do setup")
	}
	definition, err := resolve(data.WorkflowProvider)
	if err != nil {
		return WorkflowResult{}, workflowFailure("workflow declarado não é suportado", "use speckit ou openspec em um novo workspace")
	}
	result, err := applyWorkflow(knowledge, definition, "setup", "workflow setup", agent)
	result.ProjectName = data.Name
	return result, err
}

func applyWorkflow(knowledge string, definition WorkflowDefinition, operation, authorization, agent string) (WorkflowResult, error) {
	result := WorkflowResult{KnowledgePath: canonical(knowledge), Provider: definition.Provider, Executor: definition.Executor}
	root, marker, err := workflowPaths(knowledge, definition)
	if err != nil {
		return result, workflowFailure("definição de workflow inválida", "atualize o Cerne e tente novamente")
	}
	state, err := workflowLayout(root, marker)
	if err != nil {
		return result, workflowFailure("estrutura do workflow inválida ou parcial", "resolva manualmente a estrutura parcial antes de tentar novamente")
	}
	if state == WorkflowUnchanged {
		if !workflowSpecsValid(knowledge, root, definition) {
			return result, workflowFailure("estrutura do workflow inválida ou parcial", "restaure o diretório canônico de especificações")
		}
		result.State = WorkflowUnchanged
		return applyAgentDiscovery(result, knowledge, definition, agent)
	}
	if !definition.Available || definition.Setup == nil {
		result.State = WorkflowPending
		return result, nil
	}

	auditPath, attempt, err := startWorkflowAudit(knowledge, definition, operation, authorization)
	if err != nil {
		return result, workflowFailure("não foi possível registrar a tentativa de workflow", "verifique as permissões de knowledge/runs")
	}
	result.AuditPath = auditPath
	fail := func(category, cause, correction string) (WorkflowResult, error) {
		cleanupErr := os.RemoveAll(root)
		if cleanupErr != nil {
			category = "cleanup-failed"
			cause = "provider falhou e a estrutura parcial não pôde ser removida"
			correction = "remova manualmente a raiz parcial do provider antes de tentar novamente"
		}
		if err := finishWorkflowAudit(auditPath, attempt, "failed", category); err != nil {
			if cleanupErr != nil {
				return result, workflowFailure("auditoria e limpeza do workflow falharam", "remova manualmente a raiz parcial do provider e verifique knowledge/runs")
			}
			return result, workflowFailure("não foi possível finalizar a auditoria do workflow", "verifique as permissões de knowledge/runs e tente novamente")
		}
		return result, workflowFailure(cause, correction)
	}

	if err := definition.Setup(canonical(knowledge)); err != nil {
		return fail("provider-failed", "provider não concluiu a inicialização", "corrija ou atualize o provider e tente novamente")
	}
	state, err = workflowLayout(root, marker)
	if err != nil || state != WorkflowUnchanged || !workflowSpecsValid(knowledge, root, definition) {
		return fail("layout-invalid", "provider concluiu sem criar uma estrutura válida", "instale uma versão compatível do provider e tente novamente")
	}
	if err := finishWorkflowAudit(auditPath, attempt, "succeeded", ""); err != nil {
		if cleanupErr := os.RemoveAll(root); cleanupErr != nil {
			return result, workflowFailure("auditoria e limpeza do workflow falharam", "remova manualmente a raiz parcial do provider e verifique knowledge/runs")
		}
		return result, workflowFailure("não foi possível finalizar a auditoria do workflow", "verifique as permissões de knowledge/runs e tente novamente")
	}
	result.State = WorkflowConfigured
	return applyAgentDiscovery(result, knowledge, definition, agent)
}

func applyAgentDiscovery(result WorkflowResult, knowledge string, definition WorkflowDefinition, agent string) (WorkflowResult, error) {
	if agent == "" {
		return result, nil
	}
	result.Agent = agent
	target, ok := definition.Agents[agent]
	if !ok || target.Name != agent || target.DiscoveryRoot == "" {
		return result, workflowFailure("agente não suportado para o workflow", "use --agent codex ou --agent claude com workflow speckit")
	}
	if !definition.Available || target.Setup == nil {
		result.State = WorkflowPending
		result.Discovery = WorkflowDiscoveryNotCreated
		return result, nil
	}
	auditPath, attempt, err := startWorkflowAudit(knowledge, definition, "agent-integration", "--agent "+agent)
	if err != nil {
		return result, workflowFailure("não foi possível registrar a tentativa de workflow", "verifique as permissões de knowledge/runs")
	}
	result.AuditPath = auditPath
	if err := target.Setup(canonical(knowledge)); err != nil {
		if auditErr := finishWorkflowAudit(auditPath, attempt, "failed", "agent-integration-failed"); auditErr != nil {
			return result, workflowFailure("não foi possível finalizar a auditoria do workflow", "verifique as permissões de knowledge/runs e tente novamente")
		}
		return result, workflowFailure("integração do agente não concluiu", "corrija ou atualize o provider e tente novamente")
	}
	if err := validateAgentIntegration(knowledge, target); err != nil {
		if auditErr := finishWorkflowAudit(auditPath, attempt, "failed", "agent-layout-invalid"); auditErr != nil {
			return result, workflowFailure("não foi possível finalizar a auditoria do workflow", "verifique as permissões de knowledge/runs e tente novamente")
		}
		return result, workflowFailure("integração do agente incompleta", "instale uma versão compatível do provider e tente novamente")
	}
	if err := finishWorkflowAudit(auditPath, attempt, "succeeded", ""); err != nil {
		return result, workflowFailure("não foi possível finalizar a auditoria do workflow", "verifique as permissões de knowledge/runs e tente novamente")
	}
	if err := createAgentBridge(filepath.Dir(knowledge), target); err != nil {
		result.Discovery = WorkflowDiscoveryNotCreated
		return result, workflowFailure("não foi possível preparar descoberta local do agente", "verifique permissões e artefatos de agente na raiz do workspace")
	}
	result.Discovery = WorkflowDiscoveryReady
	return result, nil
}

func validateAgentIntegration(knowledge string, target WorkflowAgentTarget) error {
	root, err := safeWorkflowPath(knowledge, target.DiscoveryRoot)
	if err != nil {
		return err
	}
	for _, command := range specKitBridgeCommands {
		path := filepath.Join(root, command, "SKILL.md")
		info, err := os.Lstat(path)
		if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() || info.Size() == 0 {
			return errors.New("skill ausente ou inválida")
		}
	}
	return nil
}

func createAgentBridge(workspaceRoot string, target WorkflowAgentTarget) error {
	root, err := safeBridgePath(workspaceRoot, target.DiscoveryRoot)
	if err != nil {
		return err
	}
	if err := ensureRegularDirectory(root); err != nil {
		return err
	}
	for _, command := range specKitBridgeCommands {
		directory := filepath.Join(root, command)
		if err := ensureRegularDirectory(directory); err != nil {
			return err
		}
		skill := filepath.Join(directory, "SKILL.md")
		if info, err := os.Lstat(skill); err == nil {
			if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
				return errors.New("skill gerenciada inválida")
			}
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		if err := os.WriteFile(skill, []byte(agentBridgeContent(command, target)), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func safeBridgePath(workspaceRoot, relative string) (string, error) {
	if relative == "" || filepath.IsAbs(relative) {
		return "", errors.New("caminho inválido")
	}
	base, err := filepath.Abs(workspaceRoot)
	if err != nil {
		return "", err
	}
	path := filepath.Clean(filepath.Join(base, filepath.FromSlash(relative)))
	if !lexicallyContains(base, path) {
		return "", errors.New("caminho fora do workspace")
	}
	return path, nil
}

func ensureRegularDirectory(path string) error {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return os.MkdirAll(path, 0o755)
	}
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return errors.New("diretório inválido")
	}
	return nil
}

func agentBridgeContent(command string, target WorkflowAgentTarget) string {
	knowledgeSkill := filepath.ToSlash(filepath.Join("knowledge", filepath.FromSlash(target.DiscoveryRoot), command, "SKILL.md"))
	return "# " + command + "\n\n" +
		"Este arquivo é uma ponte local gerenciada pelo Cerne.\n\n" +
		"Use `knowledge` como raiz do projeto Spec Kit deste workspace e siga `" + knowledgeSkill + "`.\n"
}

func workflowPaths(knowledge string, definition WorkflowDefinition) (string, string, error) {
	if definition.Provider == "" || definition.Executor == "" || definition.CanonicalSpecs == "" {
		return "", "", errors.New("campos obrigatórios ausentes")
	}
	root, err := safeWorkflowPath(knowledge, definition.OwnedRoot)
	if err != nil {
		return "", "", err
	}
	marker, err := safeWorkflowPath(knowledge, definition.Marker)
	if err != nil || !lexicallyContains(root, marker) {
		return "", "", errors.New("marker fora da raiz do provider")
	}
	if _, err := safeWorkflowPath(knowledge, definition.CanonicalSpecs); err != nil {
		return "", "", err
	}
	return root, marker, nil
}

func workflowSpecsValid(knowledge, root string, definition WorkflowDefinition) bool {
	specs, err := safeWorkflowPath(knowledge, definition.CanonicalSpecs)
	if err != nil {
		return false
	}
	return lexicallyContains(root, specs) || regularDir(specs) == nil
}

func safeWorkflowPath(knowledge, relative string) (string, error) {
	if relative == "" || filepath.IsAbs(relative) {
		return "", errors.New("caminho inválido")
	}
	base, err := filepath.Abs(knowledge)
	if err != nil {
		return "", err
	}
	path := filepath.Clean(filepath.Join(base, filepath.FromSlash(relative)))
	if !lexicallyContains(base, path) {
		return "", errors.New("caminho fora de knowledge")
	}
	return path, nil
}

func lexicallyContains(parent, child string) bool {
	relative, err := filepath.Rel(filepath.Clean(parent), filepath.Clean(child))
	return err == nil && relative != "." && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

func workflowLayout(root, marker string) (WorkflowState, error) {
	info, err := os.Lstat(root)
	if errors.Is(err, os.ErrNotExist) {
		return WorkflowPending, nil
	}
	if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return "", errors.New("raiz inválida")
	}
	if err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.Name() == ".git" {
			return errors.New("repositório Git aninhado")
		}
		return nil
	}); err != nil {
		return "", err
	}
	markerInfo, err := os.Lstat(marker)
	if err != nil || markerInfo.Mode()&os.ModeSymlink != 0 || !markerInfo.Mode().IsRegular() || markerInfo.Size() == 0 {
		return "", errors.New("marker ausente ou inválido")
	}
	if !containsPath(root, marker) {
		return "", errors.New("marker fora da raiz do provider")
	}
	return WorkflowUnchanged, nil
}

func startWorkflowAudit(knowledge string, definition WorkflowDefinition, operation, authorization string) (string, workflowAttempt, error) {
	attempt := workflowAttempt{
		Kind: "workflow-setup", Provider: definition.Provider, Executor: definition.Executor,
		Operation: operation, Context: "knowledge", Authorization: authorization,
		Status: "started", StartedAt: time.Now().UTC().Format(time.RFC3339Nano),
	}
	file, err := os.CreateTemp(filepath.Join(knowledge, "runs"), "workflow-setup-*.json")
	if err != nil {
		return "", attempt, err
	}
	path := file.Name()
	if err := writeAttempt(file, attempt); err != nil {
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

func finishWorkflowAudit(path string, attempt workflowAttempt, status, failure string) error {
	attempt.Status = status
	attempt.FinishedAt = time.Now().UTC().Format(time.RFC3339Nano)
	attempt.Failure = failure
	content, err := json.MarshalIndent(attempt, "", "  ")
	if err != nil {
		return err
	}
	content = append(content, '\n')
	temp, err := os.CreateTemp(filepath.Dir(path), ".workflow-audit-*.tmp")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)
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
	return replaceWorkflowAudit(tempPath, path)
}

func writeAttempt(file *os.File, attempt workflowAttempt) error {
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(attempt); err != nil {
		return fmt.Errorf("gravar auditoria: %w", err)
	}
	return file.Sync()
}

func workflowFailure(cause, correction string) WorkflowFailure {
	return WorkflowFailure{Cause: cause, Correction: correction}
}
