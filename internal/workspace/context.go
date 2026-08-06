package workspace

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
)

type ContextWorkflowState string

const (
	ContextWorkflowNotDeclared     ContextWorkflowState = "not-declared"
	ContextWorkflowPending         ContextWorkflowState = "pending"
	ContextWorkflowReady           ContextWorkflowState = "ready"
	ContextWorkflowInvalid         ContextWorkflowState = "invalid"
	ContextWorkflowUnknownProvider ContextWorkflowState = "unknown-provider"
)

type ContextReport struct {
	SchemaVersion int               `json:"schema_version"`
	Status        Status            `json:"status"`
	Workspace     *WorkspaceContext `json:"workspace,omitempty"`
	Knowledge     *KnowledgeContext `json:"knowledge,omitempty"`
	Source        *SourceContext    `json:"source,omitempty"`
	Workflow      *WorkflowContext  `json:"workflow,omitempty"`
	Problems      []ContextProblem  `json:"problems"`
}

type WorkspaceContext struct {
	Name string `json:"name,omitempty"`
	Root string `json:"root"`
}

type KnowledgeContext struct {
	Path          string `json:"path"`
	ProductPath   string `json:"product_path,omitempty"`
	SpecsPath     string `json:"specs_path,omitempty"`
	DecisionsPath string `json:"decisions_path,omitempty"`
	PoliciesPath  string `json:"policies_path,omitempty"`
}

type SourceContext struct {
	Path            string `json:"path"`
	InsideWorkspace bool   `json:"inside_workspace"`
}

type WorkflowContext struct {
	Declared bool                 `json:"declared"`
	Provider string               `json:"provider,omitempty"`
	State    ContextWorkflowState `json:"state"`
}

type ContextProblem struct {
	Code      string `json:"code"`
	Severity  string `json:"severity"`
	Component string `json:"component"`
}

func Context(start string, resolve WorkflowResolver) ContextReport {
	report := ContextReport{SchemaVersion: 1, Status: Healthy, Problems: []ContextProblem{}}
	root, ok := locateContextWorkspace(start)
	if !ok {
		report.Problems = append(report.Problems, contextProblem("workspace-not-found", "error", "workspace"))
		return finishContext(report)
	}
	report.Workspace = &WorkspaceContext{Root: root}

	knowledge := filepath.Join(root, "knowledge")
	if regularDir(knowledge) != nil {
		report.Problems = append(report.Problems, contextProblem("knowledge-invalid", "error", "knowledge"))
		return finishContext(report)
	}
	knowledge = canonical(knowledge)
	report.Knowledge = &KnowledgeContext{Path: knowledge}
	validateContextDirectory(report.Knowledge, &report.Problems, knowledge, "product", "knowledge.product")
	validateContextDirectory(report.Knowledge, &report.Problems, knowledge, "decisions", "knowledge.decisions")
	validateContextDirectory(report.Knowledge, &report.Problems, knowledge, "policies", "knowledge.policies")
	validateContextDirectory(report.Knowledge, &report.Problems, knowledge, "runs", "knowledge.runs")

	manifestPath := filepath.Join(knowledge, "cerne.json")
	info, err := os.Lstat(manifestPath)
	if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		report.Problems = append(report.Problems, contextProblem("manifest-invalid", "error", "manifest"))
		return finishContext(report)
	}
	data, err := readManifest(manifestPath)
	if err != nil {
		report.Problems = append(report.Problems, contextProblem("manifest-invalid", "error", "manifest"))
		return finishContext(report)
	}
	if data.VersionErr != nil {
		report.Problems = append(report.Problems, contextProblem("manifest-version-unsupported", "error", "manifest"))
		return finishContext(report)
	}
	if data.WorkflowErr != nil || data.Name != filepath.Base(root) {
		report.Problems = append(report.Problems, contextProblem("manifest-invalid", "error", "manifest"))
		return finishContext(report)
	}
	report.Workspace.Name = data.Name
	addContextSource(&report, root, knowledge, data.Source)
	addContextWorkflow(&report, knowledge, data, resolve)
	return finishContext(report)
}

func locateContextWorkspace(start string) (string, bool) {
	current, err := filepath.Abs(start)
	if err != nil {
		return "", false
	}
	if info, err := os.Stat(current); err == nil && !info.IsDir() {
		current = filepath.Dir(current)
	}
	current = canonical(current)
	for {
		knowledge := filepath.Join(current, "knowledge")
		manifest := filepath.Join(knowledge, "cerne.json")
		if contextPathExists(knowledge) || contextPathExists(manifest) {
			return canonical(current), true
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", false
		}
		current = parent
	}
}

func contextPathExists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil || !errors.Is(err, os.ErrNotExist)
}

func validateContextDirectory(knowledge *KnowledgeContext, problems *[]ContextProblem, root, name, component string) {
	path := filepath.Join(root, name)
	if regularDir(path) != nil {
		*problems = append(*problems, contextProblem("required-directory-invalid", "error", component))
		return
	}
	path = canonical(path)
	switch name {
	case "product":
		knowledge.ProductPath = path
	case "decisions":
		knowledge.DecisionsPath = path
	case "policies":
		knowledge.PoliciesPath = path
	}
}

func addContextSource(report *ContextReport, root, knowledge, sourceValue string) {
	source, err := validateSourcePath(knowledge, sourceValue)
	if err != nil || regularDir(source) != nil || samePath(root, source) || samePath(knowledge, source) || containsPath(knowledge, source) || containsPath(source, knowledge) || containsPath(source, root) {
		report.Problems = append(report.Problems, contextProblem("source-invalid", "error", "source"))
		return
	}
	report.Source = &SourceContext{Path: canonical(source), InsideWorkspace: containsPath(root, source)}
}

func addContextWorkflow(report *ContextReport, knowledge string, data manifest, resolve WorkflowResolver) {
	if !data.WorkflowDeclared {
		report.Workflow = &WorkflowContext{State: ContextWorkflowNotDeclared}
		validateContextSpecs(report, filepath.Join(knowledge, "specs"), true)
		return
	}
	report.Workflow = &WorkflowContext{Declared: true, Provider: data.WorkflowProvider}
	if resolve == nil {
		report.Workflow.State = ContextWorkflowUnknownProvider
		report.Problems = append(report.Problems, contextProblem("workflow-unknown-provider", "error", "workflow"))
		return
	}
	definition, err := resolve(data.WorkflowProvider)
	if err != nil {
		report.Workflow.State = ContextWorkflowUnknownProvider
		report.Problems = append(report.Problems, contextProblem("workflow-unknown-provider", "error", "workflow"))
		return
	}
	root, marker, err := workflowPaths(knowledge, definition)
	if err != nil {
		report.Workflow.State = ContextWorkflowInvalid
		report.Problems = append(report.Problems, contextProblem("workflow-invalid", "error", "workflow"))
		return
	}
	state, layoutErr := contextWorkflowLayout(root, marker)
	specs, specsErr := safeWorkflowPath(knowledge, definition.CanonicalSpecs)
	if state == WorkflowPending && layoutErr == nil {
		report.Workflow.State = ContextWorkflowPending
		report.Problems = append(report.Problems, contextProblem("workflow-pending", "warning", "workflow"))
		if specsErr == nil && regularDir(specs) == nil {
			report.Knowledge.SpecsPath = canonical(specs)
		}
		return
	}
	specsInvalid := specsErr != nil || regularDir(specs) != nil
	if specsInvalid {
		report.Problems = append(report.Problems, contextProblem("required-directory-invalid", "error", "knowledge.specs"))
	} else {
		report.Knowledge.SpecsPath = canonical(specs)
	}
	if layoutErr != nil || specsInvalid {
		report.Workflow.State = ContextWorkflowInvalid
		report.Problems = append(report.Problems, contextProblem("workflow-invalid", "error", "workflow"))
		return
	}
	report.Workflow.State = ContextWorkflowReady
}

func contextWorkflowLayout(root, marker string) (WorkflowState, error) {
	state, err := workflowLayout(root, marker)
	if err != nil || state == WorkflowPending {
		return state, err
	}
	err = filepath.WalkDir(root, func(_ string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return errors.New("symlink")
		}
		return nil
	})
	return state, err
}

func validateContextSpecs(report *ContextReport, path string, required bool) {
	if regularDir(path) == nil {
		report.Knowledge.SpecsPath = canonical(path)
		return
	}
	if required {
		report.Problems = append(report.Problems, contextProblem("required-directory-invalid", "error", "knowledge.specs"))
	}
}

func contextProblem(code, severity, component string) ContextProblem {
	return ContextProblem{Code: code, Severity: severity, Component: component}
}

func finishContext(report ContextReport) ContextReport {
	sort.SliceStable(report.Problems, func(i, j int) bool {
		left, right := contextProblemOrder(report.Problems[i]), contextProblemOrder(report.Problems[j])
		if left != right {
			return left < right
		}
		return contextComponentOrder(report.Problems[i].Component) < contextComponentOrder(report.Problems[j].Component)
	})
	report.Status = Healthy
	for _, problem := range report.Problems {
		if problem.Severity == "error" {
			report.Status = Invalid
			return report
		}
		report.Status = Warnings
	}
	return report
}

func contextProblemOrder(problem ContextProblem) int {
	order := map[string]int{"workspace-not-found": 0, "knowledge-invalid": 1, "manifest-invalid": 2, "manifest-version-unsupported": 3, "source-invalid": 4, "required-directory-invalid": 5, "workflow-pending": 6, "workflow-invalid": 7, "workflow-unknown-provider": 8}
	return order[problem.Code]
}

func contextComponentOrder(component string) int {
	order := map[string]int{"knowledge.product": 0, "knowledge.specs": 1, "knowledge.decisions": 2, "knowledge.policies": 3, "knowledge.runs": 4}
	if value, ok := order[component]; ok {
		return value
	}
	return 9
}
