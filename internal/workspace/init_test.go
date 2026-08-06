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

func TestInitWithLocalSourceLinksWithoutCreatingOrChangingSource(t *testing.T) {
	parent := t.TempDir()
	external := filepath.Join(parent, "código existente Ω")
	if err := os.Mkdir(external, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(external, "sentinel"), []byte("keep"), 0o640); err != nil {
		t.Fatal(err)
	}
	before := readText(t, filepath.Join(external, "sentinel"))
	input, err := filepath.Rel(parent, external)
	if err != nil {
		t.Fatal(err)
	}
	result, err := InitWithSource(parent, "example", SourceInitRequest{Mode: SourceLocal, Input: input}, fakeInitRepository, fakeLinkInspect(nil, nil), nil)
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(parent, "example")
	if result.SourceMode != SourceLocal || !samePath(result.SourcePath, external) {
		t.Fatalf("resultado=%#v", result)
	}
	if _, err := os.Stat(filepath.Join(root, "source")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("source interno criado: %v", err)
	}
	if readText(t, filepath.Join(external, "sentinel")) != before {
		t.Fatal("source externo alterado")
	}
	raw := readManifestJSON(t, root)
	var source string
	if err := json.Unmarshal(raw["source"], &source); err != nil {
		t.Fatal(err)
	}
	if source != filepath.ToSlash(mustRel(t, filepath.Join(root, "knowledge"), external)) {
		t.Fatalf("source no manifesto=%q", source)
	}
}

func TestInitWithLocalSourceRevalidatesAndRollsBack(t *testing.T) {
	parent := t.TempDir()
	external := filepath.Join(parent, "external")
	if err := os.Mkdir(external, 0o755); err != nil {
		t.Fatal(err)
	}
	calls := 0
	inspect := func(path string) (LinkRepositoryFacts, error) {
		calls++
		common := filepath.Join(path, ".git")
		if calls == 2 {
			common = filepath.Join(path, ".git-changed")
		}
		return LinkRepositoryFacts{RequestedPath: path, WorktreeRoot: path, CommonDir: common, HasWorktree: true}, nil
	}
	_, err := InitWithSource(parent, "example", SourceInitRequest{Mode: SourceLocal, Input: external}, fakeInitRepository, inspect, nil)
	if err == nil {
		t.Fatal("mudança concorrente aceita")
	}
	if _, err := os.Stat(filepath.Join(parent, "example")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("workspace não revertido: %v", err)
	}
}

func TestInitWithClonePromotesValidSourceAndFinalizesRedactedAudit(t *testing.T) {
	parent := t.TempDir()
	origin := "https://token@example.invalid/private/repo.git"
	request := SourceInitRequest{Mode: SourceClone, Input: origin, OriginTransport: "https", OriginFingerprint: strings.Repeat("a", 64)}
	clone := func(_ string, staging string) error {
		audit := filepath.Join(parent, "example", "knowledge", "runs", "source-clone.json")
		if !strings.Contains(readText(t, audit), `"status": "started"`) {
			t.Fatal("clone começou antes da auditoria")
		}
		return os.Mkdir(filepath.Join(staging, ".git"), 0o755)
	}
	result, err := InitWithSource(parent, "example", request, fakeInitRepository, fakeLinkInspect(nil, nil), clone)
	if err != nil {
		t.Fatal(err)
	}
	if result.SourceMode != SourceClone || result.AuditPath == "" {
		t.Fatalf("resultado=%#v", result)
	}
	if _, err := os.Stat(filepath.Join(result.SourcePath, ".git")); err != nil {
		t.Fatal(err)
	}
	audit := readText(t, result.AuditPath)
	if !strings.Contains(audit, `"status": "succeeded"`) || strings.Contains(audit, origin) || strings.Contains(audit, "token") {
		t.Fatalf("auditoria insegura=%s", audit)
	}
	entries, err := os.ReadDir(filepath.Join(parent, "example"))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".source-clone-") {
			t.Fatalf("staging remanescente=%s", entry.Name())
		}
	}
}

func TestInitWithCloneFailurePreservesKnowledgeAndCleansOnlyStaging(t *testing.T) {
	parent := t.TempDir()
	request := SourceInitRequest{Mode: SourceClone, Input: "ssh://example.invalid/repo", OriginTransport: "ssh", OriginFingerprint: strings.Repeat("b", 64)}
	result, err := InitWithSource(parent, "example", request, fakeInitRepository, fakeLinkInspect(nil, nil), func(_ string, staging string) error {
		if err := os.WriteFile(filepath.Join(staging, "partial"), []byte("secret-provider-output"), 0o644); err != nil {
			return err
		}
		return errors.New("token-super-secreto")
	})
	var failure SourceInitFailure
	if !errors.As(err, &failure) || !failure.Incomplete || strings.Contains(err.Error(), "token-super-secreto") {
		t.Fatalf("erro=%#v", err)
	}
	if _, err := os.Stat(result.KnowledgePath); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(result.SourcePath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("source parcial existe: %v", err)
	}
	if !strings.Contains(readText(t, result.AuditPath), `"failure": "clone-failed"`) {
		t.Fatal("falha não auditada")
	}
}

func TestInitWithCloneNeverReplacesConcurrentSourceAndKeepsPromotedSourceOnAuditFailure(t *testing.T) {
	t.Run("concurrent source", func(t *testing.T) {
		parent := t.TempDir()
		request := SourceInitRequest{Mode: SourceClone, Input: "file:///origin", OriginTransport: "file", OriginFingerprint: strings.Repeat("c", 64)}
		result, err := InitWithSource(parent, "example", request, fakeInitRepository, fakeLinkInspect(nil, nil), func(_ string, staging string) error {
			if err := os.Mkdir(filepath.Join(staging, ".git"), 0o755); err != nil {
				return err
			}
			if err := os.Mkdir(resultSource(parent), 0o755); err != nil {
				return err
			}
			return os.WriteFile(filepath.Join(resultSource(parent), "sentinel"), []byte("keep"), 0o644)
		})
		if err == nil || readText(t, filepath.Join(result.SourcePath, "sentinel")) != "keep" {
			t.Fatalf("source concorrente alterado: erro=%v", err)
		}
	})

	t.Run("audit finalization", func(t *testing.T) {
		parent := t.TempDir()
		original := replaceManifestFile
		replaceManifestFile = func(string, string) error { return errors.New("disk full") }
		t.Cleanup(func() { replaceManifestFile = original })
		request := SourceInitRequest{Mode: SourceClone, Input: "file:///origin", OriginTransport: "file", OriginFingerprint: strings.Repeat("d", 64)}
		result, err := InitWithSource(parent, "example", request, fakeInitRepository, fakeLinkInspect(nil, nil), func(_ string, staging string) error {
			return os.Mkdir(filepath.Join(staging, ".git"), 0o755)
		})
		if err == nil {
			t.Fatal("falha de auditoria ignorada")
		}
		if _, statErr := os.Stat(filepath.Join(result.SourcePath, ".git")); statErr != nil {
			t.Fatalf("source promovido removido: %v", statErr)
		}
		if !strings.Contains(readText(t, result.AuditPath), `"status": "started"`) {
			t.Fatal("auditoria não permaneceu started")
		}
	})
}

func TestInitWithCloneAuditAndCleanupFailuresStaySafe(t *testing.T) {
	t.Run("started audit blocks clone and rolls back", func(t *testing.T) {
		parent := t.TempDir()
		root := filepath.Join(parent, "example")
		if err := os.Mkdir(root, 0o755); err != nil {
			t.Fatal(err)
		}
		original := openCloneAudit
		openCloneAudit = func(string, int, os.FileMode) (*os.File, error) { return nil, errors.New("disk full") }
		t.Cleanup(func() { openCloneAudit = original })
		called := false
		request := SourceInitRequest{Mode: SourceClone, Input: "file:///origin", OriginTransport: "file", OriginFingerprint: strings.Repeat("e", 64)}
		_, err := InitWithSource(parent, "example", request, fakeInitRepository, fakeLinkInspect(nil, nil), func(string, string) error { called = true; return nil })
		if err == nil || called {
			t.Fatalf("erro=%v clone chamado=%v", err, called)
		}
		assertEntries(t, root, nil)
	})

	t.Run("cleanup failure is audited", func(t *testing.T) {
		parent := t.TempDir()
		original := removeCloneStaging
		removeCloneStaging = func(string) error { return errors.New("busy") }
		t.Cleanup(func() { removeCloneStaging = original })
		request := SourceInitRequest{Mode: SourceClone, Input: "file:///origin", OriginTransport: "file", OriginFingerprint: strings.Repeat("f", 64)}
		result, err := InitWithSource(parent, "example", request, fakeInitRepository, fakeLinkInspect(nil, nil), func(_ string, staging string) error {
			if err := os.WriteFile(filepath.Join(staging, "partial"), []byte("x"), 0o644); err != nil {
				return err
			}
			return errors.New("failed")
		})
		if err == nil || !strings.Contains(readText(t, result.AuditPath), `"failure": "cleanup-failed"`) {
			t.Fatalf("erro=%v auditoria=%s", err, readText(t, result.AuditPath))
		}
	})

	t.Run("invalid clone result is removed and audited", func(t *testing.T) {
		parent := t.TempDir()
		request := SourceInitRequest{Mode: SourceClone, Input: "file:///origin", OriginTransport: "file", OriginFingerprint: strings.Repeat("0", 64)}
		inspect := func(path string) (LinkRepositoryFacts, error) {
			if strings.HasPrefix(filepath.Base(path), ".source-clone-") {
				return LinkRepositoryFacts{}, errors.New("not a repository")
			}
			return fakeLinkInspect(nil, nil)(path)
		}
		result, err := InitWithSource(parent, "example", request, fakeInitRepository, inspect, func(_ string, staging string) error {
			return os.WriteFile(filepath.Join(staging, "partial"), []byte("x"), 0o644)
		})
		if err == nil || !strings.Contains(readText(t, result.AuditPath), `"failure": "invalid-result"`) {
			t.Fatalf("erro=%v auditoria=%s", err, readText(t, result.AuditPath))
		}
		if _, statErr := os.Stat(result.SourcePath); !errors.Is(statErr, os.ErrNotExist) {
			t.Fatalf("resultado inválido promovido: %v", statErr)
		}
	})
}

func resultSource(parent string) string { return filepath.Join(parent, "example", "source") }

func assertKnowledge(t *testing.T, knowledge string) {
	t.Helper()
	for _, directory := range []string{"product", "specs", "decisions", "policies", "runs"} {
		info, err := os.Stat(filepath.Join(knowledge, directory))
		if err != nil || !info.IsDir() {
			t.Fatalf("diretório %s ausente: %v", directory, err)
		}
		info, err = os.Stat(filepath.Join(knowledge, directory, ".gitkeep"))
		if err != nil || !info.Mode().IsRegular() {
			t.Fatalf("placeholder de %s ausente: %v", directory, err)
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

func TestPromoteDirectoryNoReplacePreservesConcurrentTarget(t *testing.T) {
	parent := t.TempDir()
	staging := filepath.Join(parent, "staging")
	target := filepath.Join(parent, "target")
	if err := os.Mkdir(staging, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(staging, "ours"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(target, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, "theirs"), nil, 0o600); err != nil {
		t.Fatal(err)
	}

	if err := promoteDirectoryNoReplace(staging, target); err == nil {
		t.Fatal("promoção deveria recusar target existente")
	}
	if _, err := os.Stat(filepath.Join(target, "theirs")); err != nil {
		t.Fatalf("target concorrente foi alterado: %v", err)
	}
	if _, err := os.Stat(filepath.Join(staging, "ours")); err != nil {
		t.Fatalf("staging foi perdido: %v", err)
	}
}
