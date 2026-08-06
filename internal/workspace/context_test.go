package workspace

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

func TestContextHealthyFromDescendantAndExternalSource(t *testing.T) {
	root := newContextWorkspace(t, "example", "../source", "")
	descendant := filepath.Join(root, "knowledge", "product")
	report := Context(descendant, contextResolver)
	if report.Status != Healthy || report.SchemaVersion != 1 || len(report.Problems) != 0 {
		t.Fatalf("report = %#v", report)
	}
	if report.Workspace == nil || report.Workspace.Name != "example" || report.Workspace.Root != root {
		t.Fatalf("workspace = %#v", report.Workspace)
	}
	if report.Source == nil || !report.Source.InsideWorkspace || report.Source.Path != filepath.Join(root, "source") {
		t.Fatalf("source = %#v", report.Source)
	}
	if report.Workflow == nil || report.Workflow.State != ContextWorkflowNotDeclared {
		t.Fatalf("workflow = %#v", report.Workflow)
	}

	external := filepath.Join(t.TempDir(), "external")
	if err := os.Mkdir(external, 0o755); err != nil {
		t.Fatal(err)
	}
	external = canonical(external)
	writeContextManifest(t, root, "example", external, "")
	report = Context(root, contextResolver)
	if report.Source == nil || report.Source.InsideWorkspace || report.Source.Path != external {
		t.Fatalf("external source = %#v", report.Source)
	}
	outside := Context(external, contextResolver)
	if outside.Status != Invalid || !hasContextProblem(outside, "workspace-not-found", "workspace") {
		t.Fatalf("outside = %#v", outside)
	}
}

func TestContextWorkflowStatesAndCanonicalSpecs(t *testing.T) {
	root := newContextWorkspace(t, "example", "../source", "speckit")
	pending := Context(root, contextResolver)
	if pending.Status != Warnings || pending.Workflow.State != ContextWorkflowPending || !hasContextProblem(pending, "workflow-pending", "workflow") {
		t.Fatalf("pending = %#v", pending)
	}
	if err := os.MkdirAll(filepath.Join(root, "knowledge", ".specify"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "knowledge", ".specify", "init-options.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "knowledge", "specs"), 0o755); err != nil {
		t.Fatal(err)
	}
	ready := Context(root, contextResolver)
	if ready.Status != Healthy || ready.Workflow.State != ContextWorkflowReady || ready.Knowledge.SpecsPath != filepath.Join(root, "knowledge", "specs") {
		t.Fatalf("ready = %#v", ready)
	}
}

func TestContextPartialProblemsAreStableAndDoNotReadAgentFiles(t *testing.T) {
	root := newContextWorkspace(t, "example", "../source", "unknown")
	if err := os.RemoveAll(filepath.Join(root, "knowledge", "product")); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"AGENTS.md", "CLAUDE.md"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("SECRET SENTINEL"), 0); err != nil {
			t.Fatal(err)
		}
	}
	first := Context(root, contextResolver)
	second := Context(root, contextResolver)
	if !reflect.DeepEqual(first, second) || first.Status != Invalid {
		t.Fatalf("reports differ: %#v %#v", first, second)
	}
	if !hasContextProblem(first, "required-directory-invalid", "knowledge.product") || !hasContextProblem(first, "workflow-unknown-provider", "workflow") {
		t.Fatalf("problems = %#v", first.Problems)
	}
}

func TestContextRejectsPartialManifestSourceAndWorkflow(t *testing.T) {
	t.Run("nearest partial", func(t *testing.T) {
		outer := newContextWorkspace(t, "outer", "../source", "")
		inner := filepath.Join(outer, "source", "inner")
		if err := os.MkdirAll(filepath.Join(inner, "knowledge"), 0o755); err != nil {
			t.Fatal(err)
		}
		report := Context(inner, contextResolver)
		if report.Workspace == nil || report.Workspace.Root != inner || !hasContextProblem(report, "manifest-invalid", "manifest") {
			t.Fatalf("report = %#v", report)
		}
	})

	t.Run("future manifest", func(t *testing.T) {
		root := newContextWorkspace(t, "example", "../source", "")
		if err := os.WriteFile(filepath.Join(root, "knowledge", "cerne.json"), []byte(`{"name":"other","source":"../source","version":2,"workflow":{}}`), 0o644); err != nil {
			t.Fatal(err)
		}
		report := Context(root, contextResolver)
		if report.Workspace.Name != "" || report.Source != nil || report.Workflow != nil || !hasContextProblem(report, "manifest-version-unsupported", "manifest") {
			t.Fatalf("report = %#v", report)
		}
	})

	t.Run("unsafe source", func(t *testing.T) {
		root := newContextWorkspace(t, "example", "../source", "")
		writeContextManifest(t, root, "example", "..", "")
		report := Context(root, contextResolver)
		if report.Source != nil || !hasContextProblem(report, "source-invalid", "source") {
			t.Fatalf("report = %#v", report)
		}
	})

	t.Run("materialized workflow without specs", func(t *testing.T) {
		root := newContextWorkspace(t, "example", "../source", "speckit")
		providerRoot := filepath.Join(root, "knowledge", ".specify")
		if err := os.Mkdir(providerRoot, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(providerRoot, "init-options.json"), []byte("{}"), 0o644); err != nil {
			t.Fatal(err)
		}
		report := Context(root, contextResolver)
		if report.Workflow.State != ContextWorkflowInvalid || len(report.Problems) < 2 || report.Problems[len(report.Problems)-2].Code != "required-directory-invalid" || report.Problems[len(report.Problems)-1].Code != "workflow-invalid" {
			t.Fatalf("report = %#v", report)
		}
	})
}

func TestContextInvalidDirectoryAndWorkflowMatrix(t *testing.T) {
	for _, name := range []string{"product", "specs", "decisions", "policies", "runs"} {
		t.Run("missing "+name, func(t *testing.T) {
			root := newContextWorkspace(t, "example", "../source", "")
			if err := os.RemoveAll(filepath.Join(root, "knowledge", name)); err != nil {
				t.Fatal(err)
			}
			report := Context(root, contextResolver)
			if !hasContextProblem(report, "required-directory-invalid", "knowledge."+name) {
				t.Fatalf("problems = %#v", report.Problems)
			}
		})
	}

	for _, manifest := range []string{
		`{"name":`,
		`{"name":"other","source":"../source","version":1}`,
	} {
		t.Run("invalid manifest", func(t *testing.T) {
			root := newContextWorkspace(t, "example", "../source", "")
			if err := os.WriteFile(filepath.Join(root, "knowledge", "cerne.json"), []byte(manifest), 0o644); err != nil {
				t.Fatal(err)
			}
			if report := Context(root, contextResolver); !hasContextProblem(report, "manifest-invalid", "manifest") || report.Source != nil {
				t.Fatalf("report = %#v", report)
			}
		})
	}

	t.Run("openspec ready", func(t *testing.T) {
		root := newContextWorkspace(t, "example", "../source", "openspec")
		if err := os.MkdirAll(filepath.Join(root, "knowledge", "openspec", "specs"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, "knowledge", "openspec", "config.yaml"), []byte("schema: spec-driven\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		report := Context(root, contextResolver)
		if report.Status != Healthy || report.Workflow.State != ContextWorkflowReady || report.Knowledge.SpecsPath != filepath.Join(root, "knowledge", "openspec", "specs") {
			t.Fatalf("report = %#v", report)
		}
	})

	t.Run("nested git", func(t *testing.T) {
		root := newContextWorkspace(t, "example", "../source", "speckit")
		if err := os.MkdirAll(filepath.Join(root, "knowledge", ".specify", ".git"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, "knowledge", ".specify", "init-options.json"), []byte("{}"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Mkdir(filepath.Join(root, "knowledge", "specs"), 0o755); err != nil {
			t.Fatal(err)
		}
		if report := Context(root, contextResolver); report.Workflow.State != ContextWorkflowInvalid || !hasContextProblem(report, "workflow-invalid", "workflow") {
			t.Fatalf("report = %#v", report)
		}
	})

	if runtime.GOOS != "windows" {
		t.Run("workflow symlink", func(t *testing.T) {
			root := newContextWorkspace(t, "example", "../source", "speckit")
			providerRoot := filepath.Join(root, "knowledge", ".specify")
			if err := os.Mkdir(providerRoot, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(providerRoot, "init-options.json"), []byte("{}"), 0o644); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(filepath.Join(root, "source"), filepath.Join(providerRoot, "linked")); err != nil {
				t.Fatal(err)
			}
			if err := os.Mkdir(filepath.Join(root, "knowledge", "specs"), 0o755); err != nil {
				t.Fatal(err)
			}
			if report := Context(root, contextResolver); report.Workflow.State != ContextWorkflowInvalid {
				t.Fatalf("report = %#v", report)
			}
		})
	}
}

func contextResolver(provider string) (WorkflowDefinition, error) {
	switch provider {
	case "speckit":
		return WorkflowDefinition{Provider: provider, Executor: "specify", CanonicalSpecs: "specs", OwnedRoot: ".specify", Marker: ".specify/init-options.json"}, nil
	case "openspec":
		return WorkflowDefinition{Provider: provider, Executor: "openspec", CanonicalSpecs: "openspec/specs", OwnedRoot: "openspec", Marker: "openspec/config.yaml"}, nil
	default:
		return WorkflowDefinition{}, errors.New("unknown")
	}
}

func newContextWorkspace(t *testing.T, name, source, provider string) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), name)
	for _, path := range []string{"knowledge/product", "knowledge/decisions", "knowledge/policies", "knowledge/runs", "source"} {
		if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(path)), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if provider == "" {
		if err := os.Mkdir(filepath.Join(root, "knowledge", "specs"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	writeContextManifest(t, root, name, source, provider)
	return canonical(root)
}

func writeContextManifest(t *testing.T, root, name, source, provider string) {
	t.Helper()
	workflow := ""
	if provider != "" {
		workflow = `,"workflow":{"provider":"` + provider + `"}`
	}
	content := `{"name":"` + name + `","source":"` + filepath.ToSlash(source) + `","version":1` + workflow + `}`
	if err := os.WriteFile(filepath.Join(root, "knowledge", "cerne.json"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func hasContextProblem(report ContextReport, code, component string) bool {
	for _, problem := range report.Problems {
		if problem.Code == code && problem.Component == component {
			return true
		}
	}
	return false
}
