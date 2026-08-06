package workspace

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRestoreFailureRollsBackOnlyOwnedStagingAndKeepsAudit(t *testing.T) {
	parent, home := t.TempDir(), t.TempDir()
	knowledge := restoreRepository(t, "knowledge", map[string]string{
		"cerne.json":       `{"name":"example","source":"../source"}`,
		"product/.gitkeep": "", "specs/.gitkeep": "", "decisions/.gitkeep": "",
		"policies/.gitkeep": "", "runs/.gitkeep": "",
	})
	cloneCalls := 0
	clone := func(origin, destination string) error {
		cloneCalls++
		if cloneCalls == 1 {
			return copyRestoreTree(origin, destination)
		}
		return errors.New("SECRET origin failure")
	}
	_, err := Restore(parent, home, RestoreRequest{KnowledgeOrigin: knowledge, SourceMode: SourceClone, SourceInput: "secret"}, restoreInspector, clone, nil)
	if err == nil || pathExists(filepath.Join(parent, "example")) {
		t.Fatalf("falha deveria reverter workspace: %v", err)
	}
	entries, readErr := os.ReadDir(filepath.Join(home, ".cerne", "audit"))
	if readErr != nil || len(entries) != 1 {
		t.Fatalf("audit ausente: %v, %d", readErr, len(entries))
	}
	data, readErr := os.ReadFile(filepath.Join(home, ".cerne", "audit", entries[0].Name()))
	if readErr != nil || containsAny(string(data), "secret", "SECRET") {
		t.Fatalf("audit vazou falha: %s, %v", data, readErr)
	}
}

func TestRestorePreservesExistingDestination(t *testing.T) {
	parent, home := t.TempDir(), t.TempDir()
	target := filepath.Join(parent, "example")
	if err := os.Mkdir(target, 0o755); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(target, "theirs")
	if err := os.WriteFile(marker, []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}
	knowledge := restoreRepository(t, "knowledge", map[string]string{
		"cerne.json":       `{"name":"example","source":"../source"}`,
		"product/.gitkeep": "", "specs/.gitkeep": "", "decisions/.gitkeep": "",
		"policies/.gitkeep": "", "runs/.gitkeep": "",
	})
	_, err := Restore(parent, home, RestoreRequest{KnowledgeOrigin: knowledge, SourceMode: SourceClone, SourceInput: knowledge}, restoreInspector, copyRestoreTree, nil)
	if err == nil {
		t.Fatal("destino existente deveria ser recusado")
	}
	if data, err := os.ReadFile(marker); err != nil || string(data) != "keep" {
		t.Fatalf("destino existente foi alterado: %q, %v", data, err)
	}
}

func TestCleanupOwnedRestoreRejectsReplacedIdentity(t *testing.T) {
	parent := t.TempDir()
	target := filepath.Join(parent, ".cerne-restore-owned")
	if err := os.Mkdir(target, 0o700); err != nil {
		t.Fatal(err)
	}
	owned, err := restorePathIdentity(target)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(target, filepath.Join(parent, "retained-owned-object")); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(target, 0o700); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(target, "concurrent")
	if err := os.WriteFile(marker, []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := cleanupOwnedRestore(target, parent, ".cerne-restore-", owned); err == nil {
		t.Fatal("cleanup deveria recusar identidade substituída")
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatalf("conteúdo concorrente removido: %v", err)
	}
}

func TestRestoreRollsBackPromotionWhenFinalAuditCannotPersist(t *testing.T) {
	parent, home := t.TempDir(), t.TempDir()
	knowledge := restoreRepository(t, "knowledge", map[string]string{
		"cerne.json":       `{"name":"example","source":"../source"}`,
		"product/.gitkeep": "", "specs/.gitkeep": "", "decisions/.gitkeep": "",
		"policies/.gitkeep": "", "runs/.gitkeep": "",
	})
	source := restoreRepository(t, "source", map[string]string{})
	original := replaceRestoreAudit
	replaceRestoreAudit = func(temp, target string) error {
		data, err := os.ReadFile(temp)
		if err == nil && strings.Contains(string(data), `"status": "succeeded"`) {
			return errors.New("injected final audit failure")
		}
		return atomicReplaceFile(temp, target)
	}
	t.Cleanup(func() { replaceRestoreAudit = original })

	_, err := Restore(parent, home, RestoreRequest{KnowledgeOrigin: knowledge, SourceMode: SourceClone, SourceInput: source}, restoreInspector, copyRestoreTree, nil)
	if err == nil || pathExists(filepath.Join(parent, "example")) {
		t.Fatalf("falha final deveria reverter root: %v", err)
	}
	entries, err := os.ReadDir(filepath.Join(home, ".cerne", "audit"))
	if err != nil || len(entries) != 1 {
		t.Fatalf("audit inconclusivo ausente: %v, %d", err, len(entries))
	}
	data, err := os.ReadFile(filepath.Join(home, ".cerne", "audit", entries[0].Name()))
	if err != nil || !strings.Contains(string(data), `"status": "started"`) {
		t.Fatalf("audit deveria preservar último estado durável: %s, %v", data, err)
	}
}
