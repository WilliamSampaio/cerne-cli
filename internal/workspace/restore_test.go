package workspace

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestRestoreClonesKnowledgeAndSourceAndAuditsBeforeGit(t *testing.T) {
	parent, home := t.TempDir(), t.TempDir()
	knowledgeOrigin := restoreRepository(t, "knowledge", map[string]string{
		"cerne.json":       `{"name":"example","source":"../source"}`,
		"product/.gitkeep": "", "specs/.gitkeep": "", "decisions/.gitkeep": "",
		"policies/.gitkeep": "", "runs/.gitkeep": "",
	})
	sourceOrigin := restoreRepository(t, "source", map[string]string{"README.md": "source"})
	cloneCalls := 0
	clone := func(origin, destination string) error {
		cloneCalls++
		entries, err := os.ReadDir(filepath.Join(home, ".cerne", "audit"))
		if err != nil || len(entries) != 1 {
			t.Fatalf("audit deve existir antes do clone: entries=%d err=%v", len(entries), err)
		}
		return copyRestoreTree(origin, destination)
	}
	result, err := Restore(parent, home, RestoreRequest{
		KnowledgeOrigin: knowledgeOrigin, SourceMode: SourceClone, SourceInput: sourceOrigin,
	}, restoreInspector, clone, nil)
	if err != nil {
		t.Fatal(err)
	}
	if cloneCalls != 2 || result.Name != "example" || result.SourceMode != SourceClone {
		t.Fatalf("resultado inesperado: calls=%d result=%#v", cloneCalls, result)
	}
	if _, err := os.Stat(filepath.Join(parent, "example", "source", "README.md")); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(result.AuditPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) == "" || containsAny(string(data), knowledgeOrigin, sourceOrigin) {
		t.Fatalf("audit ausente ou vazou origem: %s", data)
	}
}

func TestRestoreLinksLocalSourceAndPreservesManifestFields(t *testing.T) {
	parent, home := t.TempDir(), t.TempDir()
	knowledgeOrigin := restoreRepository(t, "knowledge", map[string]string{
		"cerne.json":       `{"name":"example","source":"../old","custom":{"keep":true}}`,
		"product/.gitkeep": "", "specs/.gitkeep": "", "decisions/.gitkeep": "",
		"policies/.gitkeep": "", "runs/.gitkeep": "",
	})
	source := restoreRepository(t, "source", map[string]string{"untouched": "yes"})
	result, err := Restore(parent, home, RestoreRequest{
		KnowledgeOrigin: knowledgeOrigin, SourceMode: SourceLocal, SourceInput: source,
	}, restoreInspector, copyRestoreTree, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !result.ManifestChanged || result.SourcePath != canonical(source) {
		t.Fatalf("resultado inesperado: %#v", result)
	}
	data, err := os.ReadFile(filepath.Join(result.KnowledgePath, "cerne.json"))
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	var custom map[string]bool
	if json.Unmarshal(data, &raw) != nil || json.Unmarshal(raw["custom"], &custom) != nil || !custom["keep"] {
		t.Fatalf("campo desconhecido perdido: %s", data)
	}
	if got, err := os.ReadFile(filepath.Join(source, "untouched")); err != nil || string(got) != "yes" {
		t.Fatalf("source local alterado: %q, %v", got, err)
	}
}

func TestRestoreCloneFixesExternalManifestSource(t *testing.T) {
	parent, home := t.TempDir(), t.TempDir()
	external, err := json.Marshal(filepath.ToSlash(filepath.Join(t.TempDir(), "old-source")))
	if err != nil {
		t.Fatal(err)
	}
	knowledgeOrigin := restoreRepository(t, "knowledge", map[string]string{
		"cerne.json":       `{"name":"example","source":` + string(external) + `,"custom":{"keep":true}}`,
		"product/.gitkeep": "", "specs/.gitkeep": "", "decisions/.gitkeep": "",
		"policies/.gitkeep": "", "runs/.gitkeep": "",
	})
	sourceOrigin := restoreRepository(t, "source", map[string]string{"README.md": "source"})
	result, err := Restore(parent, home, RestoreRequest{
		KnowledgeOrigin: knowledgeOrigin, SourceMode: SourceClone, SourceInput: sourceOrigin,
	}, restoreInspector, copyRestoreTree, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !result.ManifestChanged || result.SourcePath != canonical(filepath.Join(parent, "example", "source")) {
		t.Fatalf("resultado inesperado: %#v", result)
	}
	data, err := os.ReadFile(filepath.Join(result.KnowledgePath, "cerne.json"))
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	var source string
	var custom map[string]bool
	if json.Unmarshal(data, &raw) != nil || json.Unmarshal(raw["source"], &source) != nil || json.Unmarshal(raw["custom"], &custom) != nil {
		t.Fatalf("manifesto inválido: %s", data)
	}
	if source != "../source" || !custom["keep"] {
		t.Fatalf("manifesto não atualizado corretamente: %s", data)
	}
}

func TestRestorePreservesPendingWorkflowWithoutExecutingProvider(t *testing.T) {
	parent, home := t.TempDir(), t.TempDir()
	knowledge := restoreRepository(t, "knowledge", map[string]string{
		"cerne.json":       `{"name":"example","source":"../source","workflow":{"provider":"speckit"}}`,
		"product/.gitkeep": "", "decisions/.gitkeep": "", "policies/.gitkeep": "", "runs/.gitkeep": "",
	})
	source := restoreRepository(t, "source", map[string]string{})
	called := false
	resolve := func(provider string) (WorkflowDefinition, error) {
		return WorkflowDefinition{Provider: provider, Executor: "specify", CanonicalSpecs: "specs", OwnedRoot: ".specify", Marker: ".specify/init-options.json", Available: true, Setup: func(string) error { called = true; return nil }}, nil
	}
	if _, err := Restore(parent, home, RestoreRequest{KnowledgeOrigin: knowledge, SourceMode: SourceClone, SourceInput: source}, restoreInspector, copyRestoreTree, resolve); err != nil {
		t.Fatal(err)
	}
	if called {
		t.Fatal("restore executou provider de workflow")
	}
}

func restoreRepository(t *testing.T, name string, files map[string]string) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), name)
	if err := os.MkdirAll(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	for path, content := range files {
		full := filepath.Join(root, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func copyRestoreTree(origin, destination string) error {
	return filepath.WalkDir(origin, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(origin, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, relative)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}

func restoreInspector(path string) (LinkRepositoryFacts, error) {
	root := canonical(path)
	return LinkRepositoryFacts{RequestedPath: root, WorktreeRoot: root, CommonDir: filepath.Join(root, ".git"), HasWorktree: true}, nil
}

func containsAny(value string, candidates ...string) bool {
	for _, candidate := range candidates {
		if candidate != "" && len(value) >= len(candidate) {
			for index := 0; index+len(candidate) <= len(value); index++ {
				if value[index:index+len(candidate)] == candidate {
					return true
				}
			}
		}
	}
	return false
}
