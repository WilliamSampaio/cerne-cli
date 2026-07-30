package workspace

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestInitCreatesWorkspaceInAbsentOrEmptyDestination(t *testing.T) {
	for _, existing := range []bool{false, true} {
		t.Run(map[bool]string{false: "absent", true: "empty"}[existing], func(t *testing.T) {
			parent := t.TempDir()
			root := filepath.Join(parent, "example")
			if existing {
				if err := os.Mkdir(root, 0o755); err != nil {
					t.Fatal(err)
				}
			}

			var initialized []string
			result, err := Init(parent, "example", func(path string) error {
				initialized = append(initialized, path)
				return os.Mkdir(filepath.Join(path, ".git"), 0o755)
			})
			if err != nil {
				t.Fatal(err)
			}

			knowledge := filepath.Join(root, "knowledge")
			source := filepath.Join(root, "source")
			if result.Name != "example" || result.KnowledgePath != knowledge || result.SourcePath != source {
				t.Fatalf("resultado inesperado: %#v", result)
			}
			if !reflect.DeepEqual(initialized, []string{knowledge, source}) {
				t.Fatalf("inicializações Git = %v", initialized)
			}
			assertKnowledge(t, knowledge)
			assertEntries(t, source, []string{".git"})
			if _, err := os.Stat(filepath.Join(root, ".git")); !os.IsNotExist(err) {
				t.Fatalf("a raiz não deve ser repositório Git: %v", err)
			}
		})
	}
}

func TestInitRejectsInvalidPortableNamesWithoutMutation(t *testing.T) {
	names := []string{
		"", ".", "..", "-project", "project.", "project/name", `project\name`,
		"project name", "project:name", "café", strings.Repeat("a", 256),
		"CON", "con.txt", "PRN", "AUX", "NUL", "COM1", "com9.log", "LPT1", "lpt9.txt",
	}
	for index, name := range names {
		t.Run(fmt.Sprintf("case-%d", index), func(t *testing.T) {
			parent := t.TempDir()
			called := false
			_, err := Init(parent, name, func(string) error {
				called = true
				return nil
			})
			if !errors.Is(err, ErrInvalidName) {
				t.Fatalf("erro = %v, esperado ErrInvalidName", err)
			}
			if called {
				t.Fatal("Git foi chamado para nome inválido")
			}
			assertEntries(t, parent, nil)
		})
	}
}

func TestInitRejectsUnsafeDestinationsWithoutMutation(t *testing.T) {
	t.Run("file", func(t *testing.T) {
		parent := t.TempDir()
		target := filepath.Join(parent, "target")
		if err := os.WriteFile(target, []byte("sentinel"), 0o644); err != nil {
			t.Fatal(err)
		}
		_, err := Init(parent, "target", func(string) error { return nil })
		if !errors.Is(err, ErrUnsafeDestination) || readText(t, target) != "sentinel" {
			t.Fatalf("erro = %v; destino foi alterado", err)
		}
	})

	t.Run("non-empty directory", func(t *testing.T) {
		parent := t.TempDir()
		target := filepath.Join(parent, "target")
		if err := os.Mkdir(target, 0o755); err != nil {
			t.Fatal(err)
		}
		sentinel := filepath.Join(target, "sentinel")
		if err := os.WriteFile(sentinel, []byte("keep"), 0o644); err != nil {
			t.Fatal(err)
		}
		_, err := Init(parent, "target", func(string) error { return nil })
		if !errors.Is(err, ErrUnsafeDestination) || readText(t, sentinel) != "keep" {
			t.Fatalf("erro = %v; conteúdo foi alterado", err)
		}
		assertEntries(t, target, []string{"sentinel"})
	})

	t.Run("symlink", func(t *testing.T) {
		parent := t.TempDir()
		outside := t.TempDir()
		target := filepath.Join(parent, "target")
		if err := os.Symlink(outside, target); err != nil {
			t.Skipf("symlink indisponível: %v", err)
		}
		_, err := Init(parent, "target", func(string) error { return nil })
		if !errors.Is(err, ErrUnsafeDestination) {
			t.Fatalf("erro = %v, esperado ErrUnsafeDestination", err)
		}
		assertEntries(t, outside, nil)
	})
}

func TestInitRollsBackOnlyCreatedArtifacts(t *testing.T) {
	for _, existing := range []bool{false, true} {
		t.Run(map[bool]string{false: "new root", true: "existing empty root"}[existing], func(t *testing.T) {
			parent := t.TempDir()
			sentinel := filepath.Join(parent, "sentinel")
			if err := os.WriteFile(sentinel, []byte("keep"), 0o644); err != nil {
				t.Fatal(err)
			}
			root := filepath.Join(parent, "target")
			if existing {
				if err := os.Mkdir(root, 0o755); err != nil {
					t.Fatal(err)
				}
			}

			calls := 0
			_, err := Init(parent, "target", func(path string) error {
				calls++
				if calls == 2 {
					return errors.New("injected failure")
				}
				return os.Mkdir(filepath.Join(path, ".git"), 0o755)
			})
			if err == nil {
				t.Fatal("Init() deveria falhar")
			}
			if readText(t, sentinel) != "keep" {
				t.Fatal("conteúdo preexistente foi alterado")
			}
			if existing {
				assertEntries(t, root, nil)
			} else if _, err := os.Lstat(root); !os.IsNotExist(err) {
				t.Fatalf("raiz criada permaneceu: %v", err)
			}
		})
	}
}

func assertKnowledge(t *testing.T, knowledge string) {
	t.Helper()
	for _, directory := range []string{"product", "specs", "decisions", "policies", "runs"} {
		info, err := os.Stat(filepath.Join(knowledge, directory))
		if err != nil || !info.IsDir() {
			t.Fatalf("diretório %s ausente: %v", directory, err)
		}
	}

	data, err := os.ReadFile(filepath.Join(knowledge, "cerne.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest map[string]string
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	expected := map[string]string{"name": "example", "source": "../source"}
	if !reflect.DeepEqual(manifest, expected) {
		t.Fatalf("manifesto = %#v", manifest)
	}
}

func assertEntries(t *testing.T, directory string, expected []string) {
	t.Helper()
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	if !reflect.DeepEqual(names, expected) {
		t.Fatalf("entradas em %s = %v, esperado %v", directory, names, expected)
	}
}

func readText(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
