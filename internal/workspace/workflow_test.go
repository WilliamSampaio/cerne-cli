package workspace

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestInitWithWorkflowConfiguresGenericLayoutAndAudit(t *testing.T) {
	for _, test := range []struct {
		name, specs, root, marker string
		wantTopSpecs              bool
	}{
		{"alpha", "specs", ".alpha", ".alpha/options.json", true},
		{"beta", "beta/specs", "beta", "beta/config.yaml", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			parent := t.TempDir()
			definition := readyDefinition(test.name, test.specs, test.root, test.marker)
			base, workflow, err := InitWithWorkflow(parent, "example", definition, fakeInitRepository)
			if err != nil || workflow.State != WorkflowConfigured {
				t.Fatalf("base=%#v workflow=%#v erro=%v", base, workflow, err)
			}
			knowledge := base.KnowledgePath
			if _, err := os.Stat(filepath.Join(knowledge, filepath.FromSlash(test.marker))); err != nil {
				t.Fatal(err)
			}
			_, specsErr := os.Stat(filepath.Join(knowledge, "specs", ".gitkeep"))
			if test.wantTopSpecs != (specsErr == nil) {
				t.Fatalf("specs top-level inesperado: %v", specsErr)
			}
			manifest := readManifestJSON(t, filepath.Dir(knowledge))
			var workflowManifest struct {
				Provider string `json:"provider"`
			}
			if err := json.Unmarshal(manifest["workflow"], &workflowManifest); err != nil || workflowManifest.Provider != test.name {
				t.Fatalf("workflow no manifesto = %s, erro=%v", manifest["workflow"], err)
			}
			attempts := auditFiles(t, knowledge)
			if len(attempts) != 1 {
				t.Fatalf("auditorias = %v", attempts)
			}
			var attempt workflowAttempt
			if err := json.Unmarshal([]byte(readText(t, attempts[0])), &attempt); err != nil {
				t.Fatal(err)
			}
			if attempt.Status != "succeeded" || attempt.Operation != "init" || attempt.Context != "knowledge" {
				t.Fatalf("auditoria = %#v", attempt)
			}
			assertEntries(t, base.SourcePath, []string{".git"})
		})
	}
}

func TestWorkflowPendingResumeAndIdempotency(t *testing.T) {
	definition := readyDefinition("alpha", "specs", ".alpha", ".alpha/options.json")
	definition.Available = false
	definition.Setup = nil
	base, pending, err := InitWithWorkflow(t.TempDir(), "example", definition, fakeInitRepository)
	if err != nil || pending.State != WorkflowPending || len(auditFiles(t, base.KnowledgePath)) != 0 {
		t.Fatalf("pending=%#v erro=%v", pending, err)
	}

	calls := 0
	definition = readyDefinition("alpha", "specs", ".alpha", ".alpha/options.json")
	originalSetup := definition.Setup
	definition.Setup = func(knowledge string) error { calls++; return originalSetup(knowledge) }
	resolve := func(string) (WorkflowDefinition, error) { return definition, nil }
	start := filepath.Join(base.KnowledgePath, "product")
	configured, err := SetupWorkflow(start, resolve)
	if err != nil || configured.State != WorkflowConfigured || calls != 1 {
		t.Fatalf("configured=%#v calls=%d erro=%v", configured, calls, err)
	}
	unchanged, err := SetupWorkflow(start, resolve)
	if err != nil || unchanged.State != WorkflowUnchanged || calls != 1 || len(auditFiles(t, base.KnowledgePath)) != 1 {
		t.Fatalf("unchanged=%#v calls=%d erro=%v", unchanged, calls, err)
	}
}

func TestWorkflowFailureCleanupPartialAndAuditFinalization(t *testing.T) {
	t.Run("provider failure", func(t *testing.T) {
		definition := readyDefinition("alpha", "specs", ".alpha", ".alpha/options.json")
		definition.Setup = func(knowledge string) error {
			if err := os.Mkdir(filepath.Join(knowledge, ".alpha"), 0o755); err != nil {
				return err
			}
			return errors.New("SECRET raw provider output")
		}
		base, _, err := InitWithWorkflow(t.TempDir(), "example", definition, fakeInitRepository)
		if err == nil || containsText(err.Error(), "SECRET") {
			t.Fatalf("erro inseguro = %v", err)
		}
		if _, statErr := os.Stat(filepath.Join(base.KnowledgePath, ".alpha")); !errors.Is(statErr, os.ErrNotExist) {
			t.Fatalf("raiz parcial permaneceu: %v", statErr)
		}
		attempts := auditFiles(t, base.KnowledgePath)
		if len(attempts) != 1 || !containsText(readText(t, attempts[0]), `"status": "failed"`) || containsText(readText(t, attempts[0]), "SECRET") {
			t.Fatalf("auditoria insegura: %v", attempts)
		}
	})

	for _, test := range []struct {
		name  string
		setup func(string) error
	}{
		{"missing marker", func(knowledge string) error { return os.Mkdir(filepath.Join(knowledge, ".alpha"), 0o755) }},
		{"nested git", func(knowledge string) error {
			definition := readyDefinition("alpha", "specs", ".alpha", ".alpha/options.json")
			if err := definition.Setup(knowledge); err != nil {
				return err
			}
			return os.MkdirAll(filepath.Join(knowledge, ".alpha", "nested", ".git"), 0o755)
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			definition := readyDefinition("alpha", "specs", ".alpha", ".alpha/options.json")
			definition.Setup = test.setup
			base, _, err := InitWithWorkflow(t.TempDir(), "example", definition, fakeInitRepository)
			if err == nil {
				t.Fatal("estrutura inválida aceita")
			}
			if _, statErr := os.Stat(filepath.Join(base.KnowledgePath, ".alpha")); !errors.Is(statErr, os.ErrNotExist) {
				t.Fatalf("raiz permaneceu: %v", statErr)
			}
		})
	}

	t.Run("partial is preserved and refused", func(t *testing.T) {
		definition := readyDefinition("alpha", "specs", ".alpha", ".alpha/options.json")
		definition.Available = false
		definition.Setup = nil
		base, _, err := InitWithWorkflow(t.TempDir(), "example", definition, fakeInitRepository)
		if err != nil {
			t.Fatal(err)
		}
		sentinel := filepath.Join(base.KnowledgePath, ".alpha", "sentinel")
		if err := os.Mkdir(filepath.Dir(sentinel), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(sentinel, []byte("keep"), 0o644); err != nil {
			t.Fatal(err)
		}
		definition = readyDefinition("alpha", "specs", ".alpha", ".alpha/options.json")
		_, err = SetupWorkflow(filepath.Dir(base.KnowledgePath), func(string) (WorkflowDefinition, error) { return definition, nil })
		if err == nil || readText(t, sentinel) != "keep" || len(auditFiles(t, base.KnowledgePath)) != 0 {
			t.Fatalf("parcial alterado: erro=%v", err)
		}
	})

	t.Run("audit finalization leaves started", func(t *testing.T) {
		original := replaceWorkflowAudit
		replaceWorkflowAudit = func(string, string) error { return errors.New("disk full") }
		t.Cleanup(func() { replaceWorkflowAudit = original })
		definition := readyDefinition("alpha", "specs", ".alpha", ".alpha/options.json")
		base, _, err := InitWithWorkflow(t.TempDir(), "example", definition, fakeInitRepository)
		if err == nil {
			t.Fatal("esperava falha")
		}
		attempts := auditFiles(t, base.KnowledgePath)
		if len(attempts) != 1 || !containsText(readText(t, attempts[0]), `"status": "started"`) {
			t.Fatalf("auditoria = %v", attempts)
		}
		if _, statErr := os.Stat(filepath.Join(base.KnowledgePath, ".alpha")); !errors.Is(statErr, os.ErrNotExist) {
			t.Fatalf("raiz permaneceu: %v", statErr)
		}
	})
}

func readyDefinition(provider, specs, root, marker string) WorkflowDefinition {
	return WorkflowDefinition{
		Provider: provider, Executor: "provider", CanonicalSpecs: specs,
		OwnedRoot: root, Marker: marker, Available: true,
		Setup: func(knowledge string) error {
			path := filepath.Join(knowledge, filepath.FromSlash(marker))
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return err
			}
			return os.WriteFile(path, []byte("configured"), 0o644)
		},
	}
}

func fakeInitRepository(path string) error { return os.Mkdir(filepath.Join(path, ".git"), 0o755) }

func auditFiles(t *testing.T, knowledge string) []string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(knowledge, "runs"))
	if err != nil {
		t.Fatal(err)
	}
	var paths []string
	for _, entry := range entries {
		if entry.Name() != ".gitkeep" {
			paths = append(paths, filepath.Join(knowledge, "runs", entry.Name()))
		}
	}
	return paths
}

func containsText(text, fragment string) bool {
	for index := 0; index+len(fragment) <= len(text); index++ {
		if text[index:index+len(fragment)] == fragment {
			return true
		}
	}
	return false
}
