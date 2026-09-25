package workspace

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateRepositoryName(t *testing.T) {
	registered := []RepositoryEntry{{Name: "frontend", Path: "../frontend"}}
	cases := []struct {
		name  string
		input string
		valid bool
	}{
		{name: "simples", input: "infra", valid: true},
		{name: "com hifen ponto e sublinhado", input: "app-web_2.0", valid: true},
		{name: "inicia por numero", input: "2fa", valid: true},
		{name: "vazio", input: ""},
		{name: "com espaco", input: "app web"},
		{name: "inicia por hifen", input: "-app"},
		{name: "termina em ponto", input: "app."},
		{name: "separador de caminho", input: "a/b"},
		{name: "ponto", input: "."},
		{name: "ponto ponto", input: ".."},
		{name: "acentuado", input: "aplicação"},
		{name: "reservado windows CON", input: "CON"},
		{name: "reservado windows COM1", input: "com1"},
		{name: "reservado source", input: "source"},
		{name: "reservado knowledge", input: "knowledge"},
		{name: "duplicado", input: "frontend"},
		{name: "longo demais", input: strings.Repeat("a", 256)},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			err := validateRepositoryName(testCase.input, registered)
			if testCase.valid && err != nil {
				t.Fatalf("nome válido recusado: %v", err)
			}
			if !testCase.valid && err == nil {
				t.Fatalf("nome inválido aceito: %q", testCase.input)
			}
		})
	}
}

func newRepoWorkspace(t *testing.T) (root string, parent string) {
	t.Helper()
	root = newDoctorWorkspace(t, "example")
	return root, filepath.Dir(root)
}

func mkRepo(t *testing.T, parent, name string) string {
	t.Helper()
	path := filepath.Join(parent, name)
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func manifestEntries(t *testing.T, root string) []RepositoryEntry {
	t.Helper()
	data, err := readManifest(filepath.Join(root, "knowledge", "cerne.json"))
	if err != nil {
		t.Fatalf("manifesto ilegível: %v", err)
	}
	return data.Repositories
}

func TestAddRepositoryRegistersEntries(t *testing.T) {
	root, parent := newRepoWorkspace(t)
	frontend := mkRepo(t, parent, "frontend")
	infra := mkRepo(t, parent, "infra")

	first, err := AddRepository(root, AddRepositoryRequest{Name: "frontend", PathInput: frontend}, fakeLinkInspect(nil, nil))
	if err != nil {
		t.Fatalf("primeiro registro falhou: %v", err)
	}
	if !first.Changed || first.Name != "frontend" || first.ProjectName != "example" {
		t.Fatalf("resultado = %#v", first)
	}
	if _, err := AddRepository(root, AddRepositoryRequest{Name: "infra", PathInput: infra}, fakeLinkInspect(nil, nil)); err != nil {
		t.Fatalf("segundo registro falhou: %v", err)
	}

	entries := manifestEntries(t, root)
	if len(entries) != 2 || entries[0].Name != "frontend" || entries[1].Name != "infra" {
		t.Fatalf("ordem de inserção não preservada: %#v", entries)
	}
	if entries[0].Path != "../../frontend" {
		t.Fatalf("caminho não é relativo a knowledge/: %q", entries[0].Path)
	}
}

func TestAddRepositoryReplaceSemantics(t *testing.T) {
	root, parent := newRepoWorkspace(t)
	frontend := mkRepo(t, parent, "frontend")
	infra := mkRepo(t, parent, "infra")
	outro := mkRepo(t, parent, "outro")
	for _, repository := range []RepositoryEntry{{Name: "frontend", Path: frontend}, {Name: "infra", Path: infra}} {
		if _, err := AddRepository(root, AddRepositoryRequest{Name: repository.Name, PathInput: repository.Path}, fakeLinkInspect(nil, nil)); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("mesmo caminho conclui sem alteracao", func(t *testing.T) {
		got, err := AddRepository(root, AddRepositoryRequest{Name: "frontend", PathInput: frontend}, fakeLinkInspect(nil, nil))
		if err != nil || got.Changed {
			t.Fatalf("resultado = %#v, err = %v", got, err)
		}
	})

	t.Run("caminho diferente recusa sem replace", func(t *testing.T) {
		_, err := AddRepository(root, AddRepositoryRequest{Name: "frontend", PathInput: outro}, fakeLinkInspect(nil, nil))
		var failure LinkFailure
		if !errors.As(err, &failure) || failure.Code != "repository-already-registered" {
			t.Fatalf("erro = %#v", err)
		}
		if manifestEntries(t, root)[0].Path != "../../frontend" {
			t.Fatal("manifesto alterado apesar da recusa")
		}
	})

	t.Run("replace preserva a posicao", func(t *testing.T) {
		if _, err := AddRepository(root, AddRepositoryRequest{Name: "frontend", PathInput: outro, Replace: true}, fakeLinkInspect(nil, nil)); err != nil {
			t.Fatal(err)
		}
		entries := manifestEntries(t, root)
		if len(entries) != 2 || entries[0].Name != "frontend" || entries[1].Name != "infra" {
			t.Fatalf("posição não preservada: %#v", entries)
		}
		if entries[0].Path != "../../outro" {
			t.Fatalf("caminho não substituído: %q", entries[0].Path)
		}
	})
}

// TestAddRepositoryDoesNotTouchLinkedRepository cobre FR-017, FR-018, FR-019, SC-001 e SC-003:
// nem o registro bem-sucedido nem as falhas bloqueantes alteram o repositório vinculado, e nenhum
// remoto é acessado em momento algum.
func TestAddRepositoryDoesNotTouchLinkedRepository(t *testing.T) {
	root, parent := newRepoWorkspace(t)
	frontend := mkRepo(t, parent, "frontend")
	marker := filepath.Join(frontend, "arquivo.txt")
	if err := os.WriteFile(marker, []byte("conteúdo"), 0o644); err != nil {
		t.Fatal(err)
	}
	snapshot := func() []string {
		var seen []string
		if err := filepath.WalkDir(frontend, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			info, err := entry.Info()
			if err != nil {
				return err
			}
			seen = append(seen, fmt.Sprintf("%s|%d|%v", path, info.Size(), entry.IsDir()))
			return nil
		}); err != nil {
			t.Fatal(err)
		}
		return seen
	}
	before := snapshot()

	var inspected []string
	tracking := func(path string) (LinkRepositoryFacts, error) {
		inspected = append(inspected, path)
		return fakeLinkInspect(nil, nil)(path)
	}
	if _, err := AddRepository(root, AddRepositoryRequest{Name: "frontend", PathInput: frontend}, tracking); err != nil {
		t.Fatal(err)
	}
	for _, request := range []AddRepositoryRequest{
		{Name: "frontend", PathInput: frontend},
		{Name: "source", PathInput: frontend},
		{Name: "nome inválido", PathInput: frontend},
		{Name: "ausente", PathInput: filepath.Join(parent, "nao-existe")},
	} {
		_, _ = AddRepository(root, request, tracking)
	}

	after := snapshot()
	if len(before) != len(after) {
		t.Fatalf("repositório vinculado alterado: %v -> %v", before, after)
	}
	for i := range before {
		if before[i] != after[i] {
			t.Fatalf("entrada alterada: %q -> %q", before[i], after[i])
		}
	}
	// SC-001: nenhuma inspeção fora do workspace e do candidato; nenhum acesso a remoto é possível
	// porque LinkGitInspect é a única porta para o Git e recebe apenas caminhos locais.
	for _, path := range inspected {
		if !filepath.IsAbs(path) {
			t.Fatalf("inspeção com caminho não local: %q", path)
		}
	}
}

func TestAddRepositoryPathResolution(t *testing.T) {
	root, parent := newRepoWorkspace(t)

	t.Run("relativo a partir de subdiretorio do workspace", func(t *testing.T) {
		app := mkRepo(t, parent, "app-rel")
		start := filepath.Join(root, "knowledge", "product")
		input, err := filepath.Rel(start, app)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := AddRepository(start, AddRepositoryRequest{Name: "rel", PathInput: input}, fakeLinkInspect(nil, nil)); err != nil {
			t.Fatalf("caminho relativo de subdiretório recusado: %v", err)
		}
		if manifestEntries(t, root)[0].Path != "../../app-rel" {
			t.Fatalf("caminho = %q", manifestEntries(t, root)[0].Path)
		}
	})

	t.Run("absoluto", func(t *testing.T) {
		app := mkRepo(t, parent, "app-abs")
		if _, err := AddRepository(root, AddRepositoryRequest{Name: "abs", PathInput: app}, fakeLinkInspect(nil, nil)); err != nil {
			t.Fatalf("caminho absoluto recusado: %v", err)
		}
	})

	// FR-013, caso positivo: um worktree com histórico próprio é aceito. O caso negativo — worktree
	// de um repositório já registrado — está em TestValidateRepositorySeparationOverSet.
	t.Run("worktree independente aceito", func(t *testing.T) {
		worktree := mkRepo(t, parent, "app-worktree")
		origem := mkRepo(t, parent, "app-origem")
		facts := map[string]LinkRepositoryFacts{worktree: {
			WorktreeRoot: canonical(worktree),
			CommonDir:    filepath.Join(canonical(origem), ".git", "worktrees", "wt"),
			HasWorktree:  true,
		}}
		if _, err := AddRepository(root, AddRepositoryRequest{Name: "wt", PathInput: worktree}, fakeLinkInspect(facts, nil)); err != nil {
			t.Fatalf("worktree independente recusado: %v", err)
		}
	})
}

func TestRemoveRepository(t *testing.T) {
	setup := func(t *testing.T) (string, string) {
		root, parent := newRepoWorkspace(t)
		for _, name := range []string{"frontend", "infra", "docs"} {
			path := mkRepo(t, parent, name)
			if _, err := AddRepository(root, AddRepositoryRequest{Name: name, PathInput: path}, fakeLinkInspect(nil, nil)); err != nil {
				t.Fatal(err)
			}
		}
		return root, parent
	}

	t.Run("remove preservando a ordem das demais", func(t *testing.T) {
		root, parent := setup(t)
		before := snapshotDir(t, filepath.Join(parent, "infra"))
		got, err := RemoveRepository(root, "infra")
		if err != nil {
			t.Fatalf("remoção falhou: %v", err)
		}
		if got.Name != "infra" || got.ProjectName != "example" || got.RemovedPath == "" {
			t.Fatalf("resultado = %#v", got)
		}
		entries := manifestEntries(t, root)
		if len(entries) != 2 || entries[0].Name != "frontend" || entries[1].Name != "docs" {
			t.Fatalf("ordem alterada: %#v", entries)
		}
		// FR-023: o repositório vinculado permanece intacto no disco.
		if after := snapshotDir(t, filepath.Join(parent, "infra")); after != before {
			t.Fatalf("repositório vinculado alterado: %q -> %q", before, after)
		}
	})

	t.Run("nome inexistente falha sem alterar o manifesto", func(t *testing.T) {
		root, _ := setup(t)
		before := manifestEntries(t, root)
		_, err := RemoveRepository(root, "fantasma")
		var failure LinkFailure
		if !errors.As(err, &failure) || failure.Code != "repository-not-registered" {
			t.Fatalf("erro = %#v", err)
		}
		if len(manifestEntries(t, root)) != len(before) {
			t.Fatal("manifesto alterado")
		}
	})

	// FR-025: source e knowledge não são removíveis por este comando.
	t.Run("nomes reservados sao recusados", func(t *testing.T) {
		root, _ := setup(t)
		for _, name := range []string{"source", "knowledge"} {
			_, err := RemoveRepository(root, name)
			var failure LinkFailure
			if !errors.As(err, &failure) || failure.Code != "repository-not-removable" {
				t.Fatalf("%s: erro = %#v", name, err)
			}
		}
		if len(manifestEntries(t, root)) != 3 {
			t.Fatal("manifesto alterado")
		}
	})

	t.Run("entrada quebrada e removivel", func(t *testing.T) {
		root, parent := setup(t)
		if err := os.Rename(filepath.Join(parent, "docs"), filepath.Join(parent, "docs-movido")); err != nil {
			t.Fatal(err)
		}
		if _, err := RemoveRepository(root, "docs"); err != nil {
			t.Fatalf("entrada quebrada não pôde ser removida: %v", err)
		}
		if len(manifestEntries(t, root)) != 2 {
			t.Fatalf("entradas = %#v", manifestEntries(t, root))
		}
	})

	t.Run("remover a ultima entrada limpa o campo", func(t *testing.T) {
		root, parent := newRepoWorkspace(t)
		path := mkRepo(t, parent, "unico")
		if _, err := AddRepository(root, AddRepositoryRequest{Name: "unico", PathInput: path}, fakeLinkInspect(nil, nil)); err != nil {
			t.Fatal(err)
		}
		if _, err := RemoveRepository(root, "unico"); err != nil {
			t.Fatal(err)
		}
		content := readFileForWorkspaceTest(t, filepath.Join(root, "knowledge", "cerne.json"))
		if strings.Contains(content, "repositories") {
			t.Fatalf("campo vazio mantido no manifesto:\n%s", content)
		}
	})
}

func snapshotDir(t *testing.T, path string) string {
	t.Helper()
	var builder strings.Builder
	err := filepath.WalkDir(path, func(entry string, info os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		builder.WriteString(entry + "|")
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return builder.String()
}

// TestRepositoryFailuresPreserveManifestByteForByte cobre SC-002 e FR-037: para cada falha
// bloqueante, o manifesto anterior permanece idêntico byte a byte.
func TestRepositoryFailuresPreserveManifestByteForByte(t *testing.T) {
	root, parent := newRepoWorkspace(t)
	frontend := mkRepo(t, parent, "frontend")
	if _, err := AddRepository(root, AddRepositoryRequest{Name: "frontend", PathInput: frontend}, fakeLinkInspect(nil, nil)); err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(root, "knowledge", "cerne.json")
	before := readFileForWorkspaceTest(t, manifestPath)

	file := filepath.Join(parent, "arquivo.txt")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	bare := mkRepo(t, parent, "bare")
	nested := mkRepo(t, parent, "frontend/interno")

	adds := map[string]AddRepositoryRequest{
		"nome invalido":    {Name: "nome inválido", PathInput: frontend},
		"nome reservado":   {Name: "source", PathInput: frontend},
		"nome duplicado":   {Name: "frontend", PathInput: mkRepo(t, parent, "outro")},
		"caminho ausente":  {Name: "x", PathInput: filepath.Join(parent, "nao-existe")},
		"nao diretorio":    {Name: "x", PathInput: file},
		"mesmo registrado": {Name: "x", PathInput: frontend},
		"aninhado":         {Name: "x", PathInput: nested},
		"knowledge":        {Name: "x", PathInput: filepath.Join(root, "knowledge")},
		"caminho vazio":    {Name: "x", PathInput: ""},
	}
	for name, request := range adds {
		t.Run("add/"+name, func(t *testing.T) {
			if _, err := AddRepository(root, request, fakeLinkInspect(nil, nil)); err == nil {
				t.Fatal("falha bloqueante não ocorreu")
			}
			if got := readFileForWorkspaceTest(t, manifestPath); got != before {
				t.Fatalf("manifesto alterado:\n%s\n---\n%s", before, got)
			}
		})
	}
	t.Run("add/bare", func(t *testing.T) {
		facts := map[string]LinkRepositoryFacts{bare: {IsBare: true}}
		if _, err := AddRepository(root, AddRepositoryRequest{Name: "x", PathInput: bare}, fakeLinkInspect(facts, nil)); err == nil {
			t.Fatal("repositório bare aceito")
		}
		if got := readFileForWorkspaceTest(t, manifestPath); got != before {
			t.Fatal("manifesto alterado")
		}
	})
	for name, target := range map[string]string{"nao registrado": "fantasma", "source": "source", "knowledge": "knowledge"} {
		t.Run("remove/"+name, func(t *testing.T) {
			if _, err := RemoveRepository(root, target); err == nil {
				t.Fatal("falha bloqueante não ocorreu")
			}
			if got := readFileForWorkspaceTest(t, manifestPath); got != before {
				t.Fatal("manifesto alterado")
			}
		})
	}
}

// TestRepositoryWriteFailurePreservesManifest cobre FR-021: a gravação é atômica e o manifesto
// anterior permanece válido e utilizável quando a substituição final falha.
func TestRepositoryWriteFailurePreservesManifest(t *testing.T) {
	root, parent := newRepoWorkspace(t)
	frontend := mkRepo(t, parent, "frontend")
	manifestPath := filepath.Join(root, "knowledge", "cerne.json")
	before := readFileForWorkspaceTest(t, manifestPath)

	original := replaceManifestFile
	replaceManifestFile = func(string, string) error { return errors.New("falha simulada") }
	t.Cleanup(func() { replaceManifestFile = original })

	_, err := AddRepository(root, AddRepositoryRequest{Name: "frontend", PathInput: frontend}, fakeLinkInspect(nil, nil))
	var failure LinkFailure
	if !errors.As(err, &failure) || failure.Code != "manifest-update-failed" {
		t.Fatalf("erro = %#v", err)
	}
	if got := readFileForWorkspaceTest(t, manifestPath); got != before {
		t.Fatalf("manifesto alterado:\n%s", got)
	}
	if _, err := readManifest(manifestPath); err != nil {
		t.Fatalf("manifesto ficou inutilizável: %v", err)
	}
}
