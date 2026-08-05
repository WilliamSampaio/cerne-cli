package main

import (
	"crypto/sha256"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

const expectedGlobalHelp = `Cerne administra workspaces com repositórios Git independentes de conhecimento e código-fonte.

Uso:
  cerne <comando> [argumentos]
  cerne --help
  cerne --version

Comandos:
  init      Cria um workspace Cerne
  doctor    Valida a estrutura e a segurança do workspace
  status    Exibe o estado local dos repositórios
  link      Vincula um repositório Git local como source
  workflow  Inicializa o workflow declarado no workspace

Opções:
  --help       Exibe esta ajuda
  --version    Exibe a versão do Cerne

Use "cerne <comando> --help" para detalhes de um comando.
`

const expectedInitHelp = `Inicializa um workspace Cerne com repositórios Git independentes.

Uso:
  cerne init <project-name>
  cerne init <project-name> --source <caminho>
  cerne init <project-name> --clone <origem>
  cerne init <project-name> --workflow <speckit|openspec>

Nome:
  1 a 255 caracteres ASCII; começa por letra ou número e continua com
  letras, números, ponto, hífen ou sublinhado. Nomes reservados e ponto final
  não são aceitos.

Source:
  Sem flag, cria source como repositório Git vazio. --source vincula a raiz de
  um working tree Git local non-bare, resolvido a partir do diretório atual,
  sem criar source interno nem alterar o repositório informado. --clone aceita
  path local, file, HTTPS ou SSH, inclusive SCP-like, e cria remoto origin.
  --source, --clone e --workflow são opções exclusivas e ficam após o nome.

Clone:
  HTTP, git://, ext::, helpers desconhecidos, opções, credenciais embutidas,
  query e fragmento são recusados. O clone é completo, sem submódulos, usa o
  checkout padrão e pode seguir redirects, autenticação externa e filtros já
  configurados no Git. Prompts controláveis pelo Git ficam desabilitados;
  helpers externos ainda podem falhar ou ter comportamento próprio.

Workflow:
  Sem a opção, mantém o layout padrão em knowledge/specs. Spec Kit também usa
  specs e cria .specify. OpenSpec usa openspec/specs e cria openspec.
  O Cerne usa somente uma instalação local existente, sem instalar, atualizar,
  selecionar agente ou fornecer credenciais. Se ausente, o setup fica pendente.

Efeitos:
  Sempre cria knowledge como Git independente. O destino deve estar ausente ou
  vazio e nunca é substituído. Antes do clone, falhas revertem os artefatos da
  tentativa. Cada clone iniciado cria runs/source-clone.json antes do Git; uma
  falha posterior preserva knowledge e a auditoria, remove somente o staging
  privado e pode deixar o workspace incompleto. A origem e a saída Git não são
  gravadas na auditoria. Nenhum modo autoriza push, merge ou publicação.

Saídas:
  Sucesso e ajuda usam stdout. Erros usam stderr.
  Status 0: sucesso ou ajuda; 1: falha operacional; 2: uso ou nome inválido.

Erros:
  Status 1 após o clone pode preservar um source já promovido se a auditoria
  final falhar; um source concorrente nunca é substituído. Consulte a correção
  exibida e knowledge/runs/source-clone.json antes de remover ou vincular source.

Exemplos:
  cerne init exemplo
  cerne init exemplo --source ../aplicacao
  cerne init exemplo --clone https://host/organizacao/aplicacao.git
  cerne init exemplo --workflow speckit
`

func TestCLIGlobalHelpAndVersion(t *testing.T) {
	binary := buildCLI(t)
	for _, test := range []struct {
		argument string
		expected string
	}{
		{"--help", expectedGlobalHelp},
		{"--version", "cerne 0.3.0\n"},
	} {
		t.Run(test.argument, func(t *testing.T) {
			status, stdout, stderr := executeCLI(t, binary, t.TempDir(), nil, test.argument)
			if status != 0 || stdout != test.expected || stderr != "" {
				t.Fatalf("status = %d\nstdout = %q\nstderr = %q", status, stdout, stderr)
			}
		})
	}
}

func TestCLIInitSuccess(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("Git não está disponível")
	}

	binary := buildCLI(t)
	parent := t.TempDir()
	command := exec.Command(binary, "init", "example")
	command.Dir = parent
	var stdout, stderr strings.Builder
	command.Stdout = &stdout
	command.Stderr = &stderr

	if err := command.Run(); err != nil {
		t.Fatalf("cerne init: %v\nstderr: %s", err, stderr.String())
	}

	knowledge := filepath.Join(parent, "example", "knowledge")
	source := filepath.Join(parent, "example", "source")
	expected := "Workspace \"example\" criado.\nKnowledge: " + knowledge + "\nSource: " + source + "\n"
	if stdout.String() != expected {
		t.Fatalf("stdout = %q, esperado %q", stdout.String(), expected)
	}
	if stderr.String() != "" {
		t.Fatalf("stderr = %q", stderr.String())
	}
	if !samePath(gitOutput(t, knowledge, "rev-parse", "--show-toplevel"), knowledge) {
		t.Fatal("raiz Git de knowledge incorreta")
	}
	if !samePath(gitOutput(t, source, "rev-parse", "--show-toplevel"), source) {
		t.Fatal("raiz Git de source incorreta")
	}
	for _, repository := range []string{knowledge, source} {
		if gitOutput(t, repository, "remote") != "" {
			t.Fatalf("%s possui remoto", repository)
		}
		if gitOutput(t, repository, "rev-list", "--all", "--count") != "0" {
			t.Fatalf("%s possui commits", repository)
		}
	}
}

func TestCLIStableContractAndPortablePath(t *testing.T) {
	binary := buildCLI(t)

	t.Run("help", func(t *testing.T) {
		status, stdout, stderr := executeCLI(t, binary, t.TempDir(), nil, "init", "--help")
		if status != 0 || stdout != expectedInitHelp || stderr != "" {
			t.Fatalf("status = %d\nstdout = %q\nstderr = %q", status, stdout, stderr)
		}
	})

	t.Run("missing argument", func(t *testing.T) {
		status, stdout, stderr := executeCLI(t, binary, t.TempDir(), nil, "init")
		expected := "erro: argumento inválido\nuso: cerne init <project-name> [--workflow <speckit|openspec> | --source <caminho> | --clone <origem>]\n"
		if status != 2 || stdout != "" || stderr != expected {
			t.Fatalf("status = %d\nstdout = %q\nstderr = %q", status, stdout, stderr)
		}
	})

	t.Run("spaces and Unicode in current path", func(t *testing.T) {
		parent := filepath.Join(t.TempDir(), "área com espaços")
		if err := os.Mkdir(parent, 0o755); err != nil {
			t.Fatal(err)
		}
		status, stdout, stderr := executeCLI(t, binary, parent, nil, "init", "portable")
		knowledge := filepath.Join(parent, "portable", "knowledge")
		source := filepath.Join(parent, "portable", "source")
		expected := "Workspace \"portable\" criado.\nKnowledge: " + knowledge + "\nSource: " + source + "\n"
		if status != 0 || stdout != expected || stderr != "" {
			t.Fatalf("status = %d\nstdout = %q\nstderr = %q", status, stdout, stderr)
		}
		if !samePath(gitOutput(t, knowledge, "rev-parse", "--show-toplevel"), knowledge) ||
			!samePath(gitOutput(t, source, "rev-parse", "--show-toplevel"), source) {
			t.Fatal("raízes Git incorretas em caminho portátil")
		}
	})
}

func TestCLIFailures(t *testing.T) {
	binary := buildCLI(t)

	t.Run("invalid name", func(t *testing.T) {
		parent := t.TempDir()
		status, stdout, stderr := executeCLI(t, binary, parent, nil, "init", "../invalid")
		if status != 2 || stdout != "" {
			t.Fatalf("status = %d, stdout = %q", status, stdout)
		}
		if !strings.Contains(stderr, "erro:") || !strings.Contains(stderr, "uso: cerne init <project-name>") {
			t.Fatalf("stderr = %q", stderr)
		}
	})

	t.Run("non-empty destination", func(t *testing.T) {
		parent := t.TempDir()
		target := filepath.Join(parent, "target")
		if err := os.Mkdir(target, 0o755); err != nil {
			t.Fatal(err)
		}
		sentinel := filepath.Join(target, "sentinel")
		if err := os.WriteFile(sentinel, []byte("keep"), 0o644); err != nil {
			t.Fatal(err)
		}
		status, stdout, stderr := executeCLI(t, binary, parent, nil, "init", "target")
		if status != 1 || stdout != "" || readFile(t, sentinel) != "keep" {
			t.Fatalf("status = %d, stdout = %q, stderr = %q", status, stdout, stderr)
		}
		if !strings.Contains(stderr, "erro:") || !strings.Contains(stderr, "correção:") ||
			!strings.Contains(stderr, "inexistente ou vazio") {
			t.Fatalf("stderr = %q", stderr)
		}
	})

	t.Run("Git missing", func(t *testing.T) {
		parent := t.TempDir()
		environment := replaceEnvironment(os.Environ(), "PATH", t.TempDir())
		status, stdout, stderr := executeCLI(t, binary, parent, environment, "init", "target")
		if status != 1 || stdout != "" {
			t.Fatalf("status = %d, stdout = %q", status, stdout)
		}
		if !strings.Contains(stderr, "Git") || !strings.Contains(stderr, "PATH") ||
			!strings.Contains(stderr, "correção:") {
			t.Fatalf("stderr = %q", stderr)
		}
		if _, err := os.Lstat(filepath.Join(parent, "target")); !os.IsNotExist(err) {
			t.Fatalf("workspace parcial encontrado: %v", err)
		}
	})
}

func TestCLIInitWithExistingLocalSource(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("Git não está disponível")
	}
	binary := buildCLI(t)
	parent := t.TempDir()
	source := filepath.Join(parent, "código existente Ω")
	if err := os.Mkdir(source, 0o755); err != nil {
		t.Fatal(err)
	}
	gitOutput(t, source, "init", "--quiet")
	gitOutput(t, source, "config", "user.email", "test@example.com")
	gitOutput(t, source, "config", "user.name", "Test")
	if err := os.WriteFile(filepath.Join(source, "tracked.txt"), []byte("keep\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitOutput(t, source, "add", "tracked.txt")
	gitOutput(t, source, "commit", "-m", "initial")
	worktree := filepath.Join(parent, "linked-worktree")
	gitOutput(t, source, "worktree", "add", "--quiet", worktree)
	before := snapshotTree(t, source)
	worktreeBefore := snapshotTree(t, worktree)
	input, err := filepath.Rel(parent, source)
	if err != nil {
		t.Fatal(err)
	}
	status, stdout, stderr := executeCLI(t, binary, parent, nil, "init", "linked", "--source", input)
	root := filepath.Join(parent, "linked")
	lines := strings.Split(stdout, "\n")
	if status != 0 || stderr != "" || len(lines) != 4 || lines[0] != `Workspace "linked" criado.` ||
		!samePath(strings.TrimPrefix(lines[1], "Knowledge: "), filepath.Join(root, "knowledge")) ||
		!samePath(strings.TrimPrefix(lines[2], "Source vinculado: "), source) || lines[3] != "" {
		t.Fatalf("status=%d stdout=%q stderr=%q", status, stdout, stderr)
	}
	if _, err := os.Stat(filepath.Join(root, "source")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("source interno criado: %v", err)
	}
	if !reflect.DeepEqual(before, snapshotTree(t, source)) {
		t.Fatal("source externo foi alterado")
	}
	status, _, stderr = executeCLI(t, binary, root, nil, "doctor")
	if status != 0 || stderr != "" {
		t.Fatalf("doctor status=%d stderr=%q", status, stderr)
	}

	status, stdout, stderr = executeCLI(t, binary, parent, nil, "init", "from-worktree", "--source", worktree)
	lines = strings.Split(stdout, "\n")
	if status != 0 || stderr != "" || len(lines) != 4 ||
		!samePath(strings.TrimPrefix(lines[2], "Source vinculado: "), worktree) {
		t.Fatalf("worktree status=%d stdout=%q stderr=%q", status, stdout, stderr)
	}
	if !reflect.DeepEqual(worktreeBefore, snapshotTree(t, worktree)) || !reflect.DeepEqual(before, snapshotTree(t, source)) {
		t.Fatal("init com worktree alterou o repositório")
	}
}

func TestCLIInitRejectsInvalidLocalSourcesWithoutWorkspace(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("Git não está disponível")
	}
	binary := buildCLI(t)
	parent := t.TempDir()
	file := filepath.Join(parent, "file")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	bare := filepath.Join(parent, "bare.git")
	if output, err := exec.Command("git", "init", "--bare", "--quiet", bare).CombinedOutput(); err != nil {
		t.Fatalf("git init --bare: %v: %s", err, output)
	}
	repository := filepath.Join(parent, "repository")
	if err := os.Mkdir(repository, 0o755); err != nil {
		t.Fatal(err)
	}
	gitOutput(t, repository, "init", "--quiet")
	if err := os.Mkdir(filepath.Join(repository, "subdirectory"), 0o755); err != nil {
		t.Fatal(err)
	}
	for index, source := range []string{filepath.Join(parent, "missing"), file, bare, filepath.Join(repository, "subdirectory"), repository} {
		name := "invalid-" + strconv.Itoa(index)
		directory := parent
		if source == repository {
			directory = repository
			source = "."
		}
		status, stdout, stderr := executeCLI(t, binary, directory, nil, "init", name, "--source", source)
		if status != 1 || stdout != "" || !strings.Contains(stderr, "correção:") {
			t.Fatalf("source=%q status=%d stdout=%q stderr=%q", source, status, stdout, stderr)
		}
		if _, err := os.Stat(filepath.Join(directory, name)); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("source=%q criou workspace: %v", source, err)
		}
	}
}

func TestCLIInitClonePopulatedAndEmptyLocalOrigins(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("Git não está disponível")
	}
	binary := buildCLI(t)
	for _, populated := range []bool{true, false} {
		t.Run(map[bool]string{true: "populated", false: "empty"}[populated], func(t *testing.T) {
			parent := t.TempDir()
			origin := filepath.Join(parent, "sensitive-origin")
			if err := os.Mkdir(origin, 0o755); err != nil {
				t.Fatal(err)
			}
			gitOutput(t, origin, "init", "--quiet")
			if populated {
				gitOutput(t, origin, "config", "user.email", "test@example.com")
				gitOutput(t, origin, "config", "user.name", "Test")
				if err := os.WriteFile(filepath.Join(origin, "tracked.txt"), []byte("history\n"), 0o644); err != nil {
					t.Fatal(err)
				}
				gitOutput(t, origin, "add", "tracked.txt")
				gitOutput(t, origin, "commit", "-m", "initial")
				gitOutput(t, origin, "tag", "v-test")
			}
			before := snapshotTree(t, origin)
			status, stdout, stderr := executeCLI(t, binary, parent, nil, "init", "cloned", "--clone", origin)
			root := filepath.Join(parent, "cloned")
			source := filepath.Join(root, "source")
			if status != 0 || stderr != "" || !strings.Contains(stdout, "Source clonado:") || strings.Contains(stdout, origin) {
				t.Fatalf("status=%d stdout=%q stderr=%q", status, stdout, stderr)
			}
			if !samePath(gitOutput(t, source, "rev-parse", "--show-toplevel"), source) || gitOutput(t, source, "remote") != "origin" {
				t.Fatal("clone ou remoto inválido")
			}
			if populated {
				if gitOutput(t, source, "rev-list", "--all", "--count") != "1" || gitOutput(t, source, "tag", "--list") != "v-test" {
					t.Fatal("histórico incompleto")
				}
			} else if gitOutput(t, source, "rev-list", "--all", "--count") != "0" {
				t.Fatal("clone vazio ganhou commit")
			}
			if !reflect.DeepEqual(before, snapshotTree(t, origin)) {
				t.Fatal("origem alterada")
			}
			audit := readFile(t, filepath.Join(root, "knowledge", "runs", "source-clone.json"))
			if !strings.Contains(audit, `"status": "succeeded"`) || strings.Contains(audit, origin) {
				t.Fatalf("auditoria=%s", audit)
			}
			for _, command := range []string{"doctor", "status"} {
				status, _, stderr := executeCLI(t, binary, root, nil, command)
				if status != 0 || stderr != "" {
					t.Fatalf("%s status=%d stderr=%q", command, status, stderr)
				}
			}
		})
	}
}

func TestCLIInitSourceSelectionRejectsInvalidShapesAndOriginsWithoutEffects(t *testing.T) {
	binary := buildCLI(t)
	environment := replaceEnvironment(os.Environ(), "PATH", t.TempDir())
	for _, args := range [][]string{
		{"init", "example", "--source"}, {"init", "example", "--clone"},
		{"init", "example", "--source", "path", "--clone", "origin"},
		{"init", "--source", "path", "example"}, {"init", "example", "--source", "--clone"},
		{"init", "example", "--clone", "https://user:secret@example.com/repo"},
		{"init", "example", "--clone", "http://example.com/repo"},
		{"init", "example", "--clone", "--upload-pack=evil"},
	} {
		parent := t.TempDir()
		status, stdout, stderr := executeCLI(t, binary, parent, environment, args...)
		if status != 2 || stdout != "" || !strings.Contains(stderr, "uso: cerne init") {
			t.Fatalf("args=%q status=%d stdout=%q stderr=%q", args, status, stdout, stderr)
		}
		if _, err := os.Stat(filepath.Join(parent, "example")); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("args=%q criou workspace: %v", args, err)
		}
	}
}

func TestCLIInitSourceSelectionOperationalFailures(t *testing.T) {
	binary := buildCLI(t)

	t.Run("Git missing for local source", func(t *testing.T) {
		parent := t.TempDir()
		source := filepath.Join(parent, "source")
		if err := os.Mkdir(source, 0o755); err != nil {
			t.Fatal(err)
		}
		environment := replaceEnvironment(os.Environ(), "PATH", t.TempDir())
		status, stdout, stderr := executeCLI(t, binary, parent, environment, "init", "example", "--source", source)
		if status != 1 || stdout != "" || !strings.Contains(stderr, "Git") || !strings.Contains(stderr, "correção:") {
			t.Fatalf("status=%d stdout=%q stderr=%q", status, stdout, stderr)
		}
		if _, err := os.Stat(filepath.Join(parent, "example")); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("workspace criado: %v", err)
		}
	})

	t.Run("clone process failure is redacted and incomplete", func(t *testing.T) {
		realGit, err := exec.LookPath("git")
		if err != nil {
			t.Skip("Git não está disponível")
		}
		parent := t.TempDir()
		origin := filepath.Join(parent, "token-super-secreto-origin")
		if err := os.Mkdir(origin, 0o755); err != nil {
			t.Fatal(err)
		}
		tools := buildWorkflowTools(t, "", true)
		environment := replaceEnvironment(os.Environ(), "PATH", tools)
		environment = replaceEnvironment(environment, "CERNE_TEST_REAL_GIT", realGit)
		status, stdout, stderr := executeCLI(t, binary, parent, environment, "init", "example", "--clone", origin)
		root := filepath.Join(parent, "example")
		if status != 1 || stdout != "" || !strings.Contains(stderr, "workspace incompleto") ||
			strings.Contains(stderr, origin) || strings.Contains(stderr, "SECRET") {
			t.Fatalf("status=%d stdout=%q stderr=%q", status, stdout, stderr)
		}
		audit := readFile(t, filepath.Join(root, "knowledge", "runs", "source-clone.json"))
		if !strings.Contains(audit, `"failure": "clone-failed"`) || strings.Contains(audit, origin) {
			t.Fatalf("auditoria=%s", audit)
		}
		if _, err := os.Stat(filepath.Join(root, "source")); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("source parcial existe: %v", err)
		}
		for _, command := range []string{"doctor", "status"} {
			status, output, diagnostic := executeCLI(t, binary, root, nil, command)
			if status != 1 || !strings.Contains(strings.ToLower(output+diagnostic), "source") {
				t.Fatalf("%s não diagnosticou source ausente: status=%d stdout=%q stderr=%q", command, status, output, diagnostic)
			}
		}
	})
}

func TestCLIWorkflowInitPendingResumeIdempotentAndFailure(t *testing.T) {
	binary := buildCLI(t)

	t.Run("configured speckit", func(t *testing.T) {
		tools := buildWorkflowTools(t, "speckit", false)
		parent := t.TempDir()
		status, stdout, stderr := executeCLI(t, binary, parent, replaceEnvironment(os.Environ(), "PATH", tools), "init", "spec", "--workflow", "speckit")
		knowledge := filepath.Join(parent, "spec", "knowledge")
		if status != 0 || !strings.HasSuffix(stdout, "Workflow: speckit\nSetup: concluído\n") || stderr != "" {
			t.Fatalf("status=%d stdout=%q stderr=%q", status, stdout, stderr)
		}
		for _, path := range []string{filepath.Join(knowledge, ".specify", "init-options.json"), filepath.Join(knowledge, "specs", ".gitkeep")} {
			if _, err := os.Stat(path); err != nil {
				t.Fatal(err)
			}
		}
	})

	t.Run("configured openspec", func(t *testing.T) {
		tools := buildWorkflowTools(t, "openspec", false)
		parent := t.TempDir()
		environment := replaceEnvironment(os.Environ(), "PATH", tools)
		status, stdout, stderr := executeCLI(t, binary, parent, environment, "init", "example", "--workflow", "openspec")
		knowledge := filepath.Join(parent, "example", "knowledge")
		lines := strings.Split(stdout, "\n")
		if status != 0 || stderr != "" || len(lines) != 6 || lines[0] != `Workspace "example" criado.` ||
			!samePath(strings.TrimPrefix(lines[1], "Knowledge: "), knowledge) ||
			!samePath(strings.TrimPrefix(lines[2], "Source: "), filepath.Join(parent, "example", "source")) ||
			lines[3] != "Workflow: openspec" || lines[4] != "Setup: concluído" || lines[5] != "" {
			t.Fatalf("status=%d stdout=%q stderr=%q", status, stdout, stderr)
		}
		if _, err := os.Stat(filepath.Join(knowledge, "openspec", "config.yaml")); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(filepath.Join(knowledge, "specs")); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("specs top-level existe: %v", err)
		}
		if strings.Contains(readFile(t, filepath.Join(knowledge, "openspec", "record")), "SECRET") {
			t.Fatal("segredo chegou ao provider")
		}
	})

	t.Run("pending resume and no-op", func(t *testing.T) {
		tools := buildWorkflowTools(t, "", false)
		parent := t.TempDir()
		environment := replaceEnvironment(os.Environ(), "PATH", tools)
		status, stdout, stderr := executeCLI(t, binary, parent, environment, "init", "pending", "--workflow", "speckit")
		root := filepath.Join(parent, "pending")
		if status != 0 || !strings.Contains(stdout, "Setup: pendente") || !strings.Contains(stderr, `aviso: executável "specify" não encontrado`) {
			t.Fatalf("status=%d stdout=%q stderr=%q", status, stdout, stderr)
		}
		installWorkflowTool(t, tools, "speckit")
		nested := filepath.Join(root, "knowledge", "product")
		status, stdout, stderr = executeCLI(t, binary, nested, environment, "workflow", "setup")
		knowledge := filepath.Join(root, "knowledge")
		expected := "Workflow: speckit\nKnowledge: " + displayPath(knowledge) + "\nSetup concluído.\n"
		if status != 0 || stdout != expected || stderr != "" {
			t.Fatalf("status=%d stdout=%q stderr=%q", status, stdout, stderr)
		}
		before := snapshotTree(t, root)
		status, stdout, stderr = executeCLI(t, binary, nested, environment, "workflow", "setup")
		if status != 0 || stdout != "Workflow: speckit\nKnowledge: "+displayPath(knowledge)+"\nNenhuma alteração necessária.\n" || stderr != "" || !reflect.DeepEqual(before, snapshotTree(t, root)) {
			t.Fatalf("no-op status=%d stdout=%q stderr=%q", status, stdout, stderr)
		}
	})

	t.Run("safe provider failure preserves base", func(t *testing.T) {
		tools := buildWorkflowTools(t, "speckit", true)
		parent := t.TempDir()
		environment := replaceEnvironment(os.Environ(), "PATH", tools)
		environment = append(environment, "CERNE_SECRET_TOKEN=SECRET-value")
		status, stdout, stderr := executeCLI(t, binary, parent, environment, "init", "failed", "--workflow", "speckit")
		root := filepath.Join(parent, "failed")
		if status != 1 || stdout != "" || strings.Contains(stderr, "SECRET") || !strings.Contains(stderr, "provider não concluiu") {
			t.Fatalf("status=%d stdout=%q stderr=%q", status, stdout, stderr)
		}
		if _, err := os.Stat(filepath.Join(root, "knowledge", "cerne.json")); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(filepath.Join(root, "source", ".git")); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(filepath.Join(root, "knowledge", ".specify")); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("raiz parcial existe: %v", err)
		}
		runs, err := os.ReadDir(filepath.Join(root, "knowledge", "runs"))
		if err != nil {
			t.Fatal(err)
		}
		if len(runs) != 2 {
			t.Fatalf("runs=%v", runs)
		}
	})
}

func TestCLIWorkflowUsageAndHelp(t *testing.T) {
	binary := buildCLI(t)
	for _, args := range [][]string{
		{"init", "example", "--workflow"},
		{"init", "example", "--workflow", "other"},
		{"init", "example", "--workflow", "speckit", "extra"},
		{"init", "--workflow", "speckit", "example"},
		{"workflow", "setup", "extra"},
	} {
		status, stdout, stderr := executeCLI(t, binary, t.TempDir(), nil, args...)
		if status != 2 || stdout != "" || !strings.Contains(stderr, "uso:") {
			t.Fatalf("args=%v status=%d stdout=%q stderr=%q", args, status, stdout, stderr)
		}
	}
	status, stdout, stderr := executeCLI(t, binary, t.TempDir(), nil, "workflow", "--help")
	if status != 0 || stderr != "" || !strings.Contains(stdout, "cerne workflow setup") {
		t.Fatalf("status=%d stdout=%q stderr=%q", status, stdout, stderr)
	}
	parent := t.TempDir()
	root := initWorkspaceWithCLI(t, binary, parent, "legacy")
	status, stdout, stderr = executeCLI(t, binary, root, nil, "workflow", "setup")
	if status != 1 || stdout != "" || !strings.Contains(stderr, "workflow não configurado") {
		t.Fatalf("status=%d stdout=%q stderr=%q", status, stdout, stderr)
	}
}

func TestCLIDoctorHealthyWarningAndInvalidReports(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("Git não está disponível")
	}
	binary := buildCLI(t)

	t.Run("healthy exact stdout", func(t *testing.T) {
		root := initWorkspaceWithCLI(t, binary, t.TempDir(), "example")
		start := time.Now()
		status, stdout, stderr := executeCLI(t, binary, root, nil, "doctor")
		if time.Since(start) > 5*time.Second {
			t.Fatal("doctor excedeu 5 segundos")
		}
		if status != 0 || stdout != expectedDoctorHealthy || stderr != "" {
			t.Fatalf("status = %d\nstdout = %q\nstderr = %q", status, stdout, stderr)
		}
	})

	t.Run("warning name differs", func(t *testing.T) {
		root := initWorkspaceWithCLI(t, binary, t.TempDir(), "example")
		if err := os.WriteFile(filepath.Join(root, "knowledge", "cerne.json"),
			[]byte(`{"name":"other","source":"../source"}`+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		status, stdout, stderr := executeCLI(t, binary, root, nil, "doctor")
		if status != 0 || stderr != "" || !strings.Contains(stdout, "Workspace com avisos") ||
			!strings.Contains(stdout, "! Manifesto: name válido difere do nome da raiz; correção:") {
			t.Fatalf("status = %d\nstdout = %q\nstderr = %q", status, stdout, stderr)
		}
	})

	t.Run("invalid report includes ten lines and no private content", func(t *testing.T) {
		root := initWorkspaceWithCLI(t, binary, t.TempDir(), "example")
		if err := os.RemoveAll(filepath.Join(root, "knowledge", "product")); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, "knowledge", "cerne.json"),
			[]byte(`{"name":"example","source":"../source","secret":"nao-vazar"}`+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		status, stdout, stderr := executeCLI(t, binary, root, nil, "doctor")
		if status != 1 || stderr != "" || countReportLines(stdout) != 10 ||
			!strings.Contains(stdout, "Workspace inválido") || !strings.Contains(stdout, "correção:") ||
			strings.Contains(stdout, "nao-vazar") {
			t.Fatalf("status = %d\nstdout = %q\nstderr = %q", status, stdout, stderr)
		}
	})

	t.Run("Git unavailable is diagnostic", func(t *testing.T) {
		root := initWorkspaceWithCLI(t, binary, t.TempDir(), "example")
		env := replaceEnvironment(os.Environ(), "PATH", t.TempDir())
		status, stdout, stderr := executeCLI(t, binary, root, env, "doctor")
		if status != 1 || stderr != "" || countReportLines(stdout) != 10 ||
			!strings.Contains(stdout, "✗ Git: indisponível; correção: instale o Git") {
			t.Fatalf("status = %d\nstdout = %q\nstderr = %q", status, stdout, stderr)
		}
	})
}

func TestCLIDoctorHelpUsagePortablePathAndReadOnly(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("Git não está disponível")
	}
	binary := buildCLI(t)

	t.Run("help", func(t *testing.T) {
		status, stdout, stderr := executeCLI(t, binary, t.TempDir(), nil, "doctor", "--help")
		if status != 0 || stderr != "" || !strings.Contains(stdout, "cerne doctor") ||
			!strings.Contains(stdout, "Status 0") || !strings.Contains(stdout, "Leitura exclusiva") {
			t.Fatalf("status = %d\nstdout = %q\nstderr = %q", status, stdout, stderr)
		}
	})

	t.Run("invalid usage", func(t *testing.T) {
		status, stdout, stderr := executeCLI(t, binary, t.TempDir(), nil, "doctor", "extra")
		expected := "erro: argumento inválido\nuso: cerne doctor\n"
		if status != 2 || stdout != "" || stderr != expected {
			t.Fatalf("status = %d\nstdout = %q\nstderr = %q", status, stdout, stderr)
		}
	})

	t.Run("spaces unicode and read only", func(t *testing.T) {
		parent := filepath.Join(t.TempDir(), "área com espaços")
		if err := os.Mkdir(parent, 0o755); err != nil {
			t.Fatal(err)
		}
		root := initWorkspaceWithCLI(t, binary, parent, "portable")
		before := snapshotTree(t, root)
		status, stdout, stderr := executeCLI(t, binary, root, nil, "doctor")
		after := snapshotTree(t, root)
		if status != 0 || stdout != expectedDoctorHealthy || stderr != "" {
			t.Fatalf("status = %d\nstdout = %q\nstderr = %q", status, stdout, stderr)
		}
		if !reflect.DeepEqual(before, after) {
			t.Fatalf("doctor alterou o workspace\nantes=%#v\ndepois=%#v", before, after)
		}
	})
}

func TestCLIDoctorPreReportFailure(t *testing.T) {
	original := currentDirectory
	currentDirectory = func() (string, error) { return "", errors.New("cwd indisponível") }
	t.Cleanup(func() { currentDirectory = original })

	var stdout, stderr strings.Builder
	status := runDoctor(nil, &stdout, &stderr)
	if status != 1 || stdout.String() != "" ||
		!strings.Contains(stderr.String(), "correção: execute o comando em um diretório acessível") ||
		strings.Contains(stderr.String(), "Workspace ") {
		t.Fatalf("status = %d\nstdout = %q\nstderr = %q", status, stdout.String(), stderr.String())
	}
}

func TestCLIStatusCleanWorkspace(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("Git não está disponível")
	}
	binary := buildCLI(t)
	root := initWorkspaceWithCLI(t, binary, t.TempDir(), "example")
	knowledge := filepath.Join(root, "knowledge")
	source := filepath.Join(root, "source")
	gitOutput(t, knowledge, "config", "core.autocrlf", "false")
	gitOutput(t, knowledge, "config", "user.email", "test@example.com")
	gitOutput(t, knowledge, "config", "user.name", "Test")
	gitOutput(t, knowledge, "add", ".")
	gitOutput(t, knowledge, "commit", "-m", "init")
	gitOutput(t, source, "config", "core.autocrlf", "false")
	gitOutput(t, source, "config", "user.email", "test@example.com")
	gitOutput(t, source, "config", "user.name", "Test")
	gitOutput(t, source, "commit", "--allow-empty", "-m", "init")
	knowledgeBranch := gitOutput(t, knowledge, "symbolic-ref", "--short", "HEAD")
	sourceBranch := gitOutput(t, source, "symbolic-ref", "--short", "HEAD")
	knowledgeCommit := gitOutput(t, knowledge, "rev-parse", "--short=7", "HEAD")
	sourceCommit := gitOutput(t, source, "rev-parse", "--short=7", "HEAD")
	expected := expectedStatus(root, "example",
		repositoryExpectation{"Knowledge", knowledge, knowledgeBranch, knowledgeCommit, "limpo", 0, 0, 0},
		repositoryExpectation{"Source", source, sourceBranch, sourceCommit, "limpo", 0, 0, 0},
	)

	status, stdout, stderr := executeCLI(t, binary, root, nil, "status")
	if status != 0 || stdout != expected || stderr != "" {
		t.Fatalf("status = %d\nstdout = %q\nstderr = %q\nesperado = %q", status, stdout, stderr, expected)
	}

	subdir := filepath.Join(source, "pkg")
	if err := os.Mkdir(subdir, 0o755); err != nil {
		t.Fatal(err)
	}
	status, stdout, stderr = executeCLI(t, binary, subdir, nil, "status")
	if status != 0 || stdout != expected || stderr != "" {
		t.Fatalf("subdir status = %d\nstdout = %q\nstderr = %q", status, stdout, stderr)
	}
}

func TestCLIStatusPendingDetachedAndNoCommits(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("Git não está disponível")
	}
	binary := buildCLI(t)
	root := initWorkspaceWithCLI(t, binary, t.TempDir(), "example")
	source := filepath.Join(root, "source")
	gitOutput(t, source, "config", "core.autocrlf", "false")
	gitOutput(t, source, "config", "user.email", "test@example.com")
	gitOutput(t, source, "config", "user.name", "Test")
	if err := os.WriteFile(filepath.Join(source, "tracked.txt"), []byte("base\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitOutput(t, source, "add", "tracked.txt")
	gitOutput(t, source, "commit", "-m", "init")
	gitOutput(t, source, "checkout", "--detach")
	commit := gitOutput(t, source, "rev-parse", "--short=7", "HEAD")
	if err := os.WriteFile(filepath.Join(source, "tracked.txt"), []byte("changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "staged.txt"), []byte("staged\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitOutput(t, source, "add", "staged.txt")
	if err := os.WriteFile(filepath.Join(source, "untracked.txt"), []byte("untracked\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	status, stdout, stderr := executeCLI(t, binary, root, nil, "status")
	if status != 0 || stderr != "" {
		t.Fatalf("status = %d\nstdout = %q\nstderr = %q", status, stdout, stderr)
	}
	for _, want := range []string{
		"Knowledge\n",
		"  Commit: sem commits\n",
		"Source\n",
		"  Branch: detached HEAD\n",
		"  Commit: " + commit + "\n",
		"  Estado: alterações pendentes\n",
		"  Modificados: 1\n",
		"  Em stage: 1\n",
		"  Não rastreados: 1\n",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("stdout não contém %q:\n%s", want, stdout)
		}
	}
}

func TestCLIStatusFailuresHelpUsageAndReadOnly(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("Git não está disponível")
	}
	binary := buildCLI(t)

	t.Run("help", func(t *testing.T) {
		status, stdout, stderr := executeCLI(t, binary, t.TempDir(), nil, "status", "--help")
		if status != 0 || stderr != "" || !strings.Contains(stdout, "cerne status") ||
			!strings.Contains(stdout, "Leitura exclusiva") || !strings.Contains(stdout, "detached HEAD") {
			t.Fatalf("status = %d\nstdout = %q\nstderr = %q", status, stdout, stderr)
		}
	})

	t.Run("invalid usage", func(t *testing.T) {
		status, stdout, stderr := executeCLI(t, binary, t.TempDir(), nil, "status", "extra")
		expected := "erro: argumento inválido\nuso: cerne status\n"
		if status != 2 || stdout != "" || stderr != expected {
			t.Fatalf("status = %d\nstdout = %q\nstderr = %q", status, stdout, stderr)
		}
	})

	t.Run("workspace not found", func(t *testing.T) {
		dir := t.TempDir()
		status, stdout, stderr := executeCLI(t, binary, dir, nil, "status")
		if status != 1 || stdout != "" || !strings.Contains(stderr, "workspace Cerne não localizado") ||
			!containsPathAlias(stderr, dir) || !strings.Contains(stderr, "correção:") {
			t.Fatalf("status = %d\nstdout = %q\nstderr = %q", status, stdout, stderr)
		}
	})

	t.Run("missing manifest", func(t *testing.T) {
		root := initWorkspaceWithCLI(t, binary, t.TempDir(), "example")
		manifest := filepath.Join(root, "knowledge", "cerne.json")
		if err := os.Remove(manifest); err != nil {
			t.Fatal(err)
		}
		status, stdout, stderr := executeCLI(t, binary, root, nil, "status")
		if status != 1 || stdout != "" || !strings.Contains(stderr, "manifesto Cerne ausente") ||
			!containsPathAlias(stderr, manifest) {
			t.Fatalf("status = %d\nstdout = %q\nstderr = %q", status, stdout, stderr)
		}
	})

	t.Run("not git repository", func(t *testing.T) {
		root := initWorkspaceWithCLI(t, binary, t.TempDir(), "example")
		source := filepath.Join(root, "source")
		if err := os.RemoveAll(filepath.Join(source, ".git")); err != nil {
			t.Fatal(err)
		}
		status, stdout, stderr := executeCLI(t, binary, root, nil, "status")
		if status != 1 || stdout != "" || !strings.Contains(stderr, "repositório Git") ||
			!containsPathAlias(stderr, source) {
			t.Fatalf("status = %d\nstdout = %q\nstderr = %q", status, stdout, stderr)
		}
	})

	t.Run("read only", func(t *testing.T) {
		root := initWorkspaceWithCLI(t, binary, t.TempDir(), "example")
		before := snapshotTree(t, root)
		status, stdout, stderr := executeCLI(t, binary, root, nil, "status")
		after := snapshotTree(t, root)
		if status != 0 || stdout == "" || stderr != "" {
			t.Fatalf("status = %d\nstdout = %q\nstderr = %q", status, stdout, stderr)
		}
		if !reflect.DeepEqual(before, after) {
			t.Fatalf("status alterou o workspace\nantes=%#v\ndepois=%#v", before, after)
		}
	})
}

func TestCLIStatusPreReportFailure(t *testing.T) {
	original := currentDirectory
	currentDirectory = func() (string, error) { return "", errors.New("cwd indisponível") }
	t.Cleanup(func() { currentDirectory = original })

	var stdout, stderr strings.Builder
	status := runStatus(nil, &stdout, &stderr)
	if status != 1 || stdout.String() != "" ||
		!strings.Contains(stderr.String(), "correção: execute o comando em um diretório acessível") {
		t.Fatalf("status = %d\nstdout = %q\nstderr = %q", status, stdout.String(), stderr.String())
	}
}

func TestCLILinkSuccessReplaceNoopAndReadOnly(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("Git não está disponível")
	}
	binary := buildCLI(t)
	parent := t.TempDir()
	root := initWorkspaceWithCLI(t, binary, parent, "example")
	newSource := filepath.Join(parent, "geo app Ω")
	if err := os.Mkdir(newSource, 0o755); err != nil {
		t.Fatal(err)
	}
	gitOutput(t, newSource, "init", "--quiet")
	before := snapshotTree(t, newSource)

	start := time.Now()
	status, stdout, stderr := executeCLI(t, binary, root, nil, "link", "../geo app Ω", "--replace")
	if time.Since(start) > 5*time.Second {
		t.Fatal("link excedeu 5 segundos")
	}
	expectedNewSource := filepath.ToSlash(filepath.Join("..", "..", "geo app Ω"))
	expected := "Projeto: example\nSource anterior: ../source\nNovo source: " + expectedNewSource + "\nManifesto atualizado.\n"
	if status != 0 || stdout != expected || stderr != "" {
		t.Fatalf("status = %d\nstdout = %q\nstderr = %q\nesperado = %q", status, stdout, stderr, expected)
	}
	if !strings.Contains(readFile(t, filepath.Join(root, "knowledge", "cerne.json")), `"source": "`+expectedNewSource+`"`) {
		t.Fatalf("manifesto não atualizado:\n%s", readFile(t, filepath.Join(root, "knowledge", "cerne.json")))
	}
	if after := snapshotTree(t, newSource); !reflect.DeepEqual(before, after) {
		t.Fatalf("link alterou source\nantes=%#v\ndepois=%#v", before, after)
	}

	status, stdout, stderr = executeCLI(t, binary, root, nil, "link", "../geo app Ω")
	expected = "Projeto: example\nSource atual: " + expectedNewSource + "\nNenhuma alteração necessária.\n"
	if status != 0 || stdout != expected || stderr != "" {
		t.Fatalf("no-op status = %d\nstdout = %q\nstderr = %q", status, stdout, stderr)
	}

	status, stdout, stderr = executeCLI(t, binary, root, nil, "status")
	if status != 0 || stderr != "" || !containsPathAlias(stdout, newSource) {
		t.Fatalf("status após link externo = %d\nstdout = %q\nstderr = %q", status, stdout, stderr)
	}
	status, stdout, stderr = executeCLI(t, binary, root, nil, "doctor")
	if status != 0 || stderr != "" || !strings.HasSuffix(stdout, "Workspace saudável\n") {
		t.Fatalf("doctor após link externo = %d\nstdout = %q\nstderr = %q", status, stdout, stderr)
	}
}

func TestCLILinkRefusesReplacementWithoutFlag(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("Git não está disponível")
	}
	binary := buildCLI(t)
	parent := t.TempDir()
	root := initWorkspaceWithCLI(t, binary, parent, "example")
	newSource := filepath.Join(parent, "new-source")
	if err := os.Mkdir(newSource, 0o755); err != nil {
		t.Fatal(err)
	}
	gitOutput(t, newSource, "init", "--quiet")
	beforeManifest := readFile(t, filepath.Join(root, "knowledge", "cerne.json"))
	beforeOld := snapshotTree(t, filepath.Join(root, "source"))
	beforeNew := snapshotTree(t, newSource)

	status, stdout, stderr := executeCLI(t, binary, root, nil, "link", "../new-source")
	if status != 1 || stdout != "" || !strings.Contains(stderr, "outro source já está configurado") ||
		!strings.Contains(stderr, "--replace") || !strings.Contains(stderr, "correção:") {
		t.Fatalf("status = %d\nstdout = %q\nstderr = %q", status, stdout, stderr)
	}
	if readFile(t, filepath.Join(root, "knowledge", "cerne.json")) != beforeManifest ||
		!reflect.DeepEqual(beforeOld, snapshotTree(t, filepath.Join(root, "source"))) ||
		!reflect.DeepEqual(beforeNew, snapshotTree(t, newSource)) {
		t.Fatal("recusa sem --replace alterou arquivos")
	}
}

func TestCLILinkFailuresHelpUsageAndPreReportFailure(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("Git não está disponível")
	}
	binary := buildCLI(t)

	t.Run("help", func(t *testing.T) {
		status, stdout, stderr := executeCLI(t, binary, t.TempDir(), nil, "link", "--help")
		if status != 0 || stderr != "" || !strings.Contains(stdout, "cerne link <caminho>") ||
			!strings.Contains(stdout, "--replace") || !strings.Contains(stdout, "Status 0") {
			t.Fatalf("status = %d\nstdout = %q\nstderr = %q", status, stdout, stderr)
		}
	})

	t.Run("invalid usage", func(t *testing.T) {
		status, stdout, stderr := executeCLI(t, binary, t.TempDir(), nil, "link", "--replace")
		expected := "erro: argumento inválido\nuso: cerne link <caminho> [--replace]\n"
		if status != 2 || stdout != "" || stderr != expected {
			t.Fatalf("status = %d\nstdout = %q\nstderr = %q", status, stdout, stderr)
		}
	})

	t.Run("workspace not found", func(t *testing.T) {
		dir := t.TempDir()
		status, stdout, stderr := executeCLI(t, binary, dir, nil, "link", dir)
		if status != 1 || stdout != "" || !strings.Contains(stderr, "workspace Cerne não localizado") ||
			!containsPathAlias(stderr, dir) || !strings.Contains(stderr, "correção:") {
			t.Fatalf("status = %d\nstdout = %q\nstderr = %q", status, stdout, stderr)
		}
	})

	t.Run("invalid git path", func(t *testing.T) {
		parent := t.TempDir()
		root := initWorkspaceWithCLI(t, binary, parent, "example")
		dir := filepath.Join(parent, "plain-dir")
		if err := os.Mkdir(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		status, stdout, stderr := executeCLI(t, binary, root, nil, "link", "../plain-dir", "--replace")
		if status != 1 || stdout != "" || !strings.Contains(stderr, "repositório Git válido") ||
			!containsPathAlias(stderr, dir) {
			t.Fatalf("status = %d\nstdout = %q\nstderr = %q", status, stdout, stderr)
		}
	})

	t.Run("bare repository", func(t *testing.T) {
		parent := t.TempDir()
		root := initWorkspaceWithCLI(t, binary, parent, "example")
		bare := filepath.Join(parent, "bare.git")
		if output, err := exec.Command("git", "init", "--bare", "--quiet", bare).CombinedOutput(); err != nil {
			t.Fatalf("git init --bare: %v: %s", err, output)
		}
		status, stdout, stderr := executeCLI(t, binary, root, nil, "link", "../bare.git", "--replace")
		if status != 1 || stdout != "" || !strings.Contains(stderr, "bare") || !containsPathAlias(stderr, bare) {
			t.Fatalf("status = %d\nstdout = %q\nstderr = %q", status, stdout, stderr)
		}
	})
}

func TestCLILinkPreReportFailure(t *testing.T) {
	original := currentDirectory
	currentDirectory = func() (string, error) { return "", errors.New("cwd indisponível") }
	t.Cleanup(func() { currentDirectory = original })

	var stdout, stderr strings.Builder
	status := runLink([]string{"."}, &stdout, &stderr)
	if status != 1 || stdout.String() != "" ||
		!strings.Contains(stderr.String(), "correção: execute o comando em um diretório acessível") {
		t.Fatalf("status = %d\nstdout = %q\nstderr = %q", status, stdout.String(), stderr.String())
	}
}

const expectedDoctorHealthy = `✓ Manifesto: legível
✓ Repositório de conhecimento: encontrado
✓ Repositório de código-fonte: encontrado
✓ Independência Git: raízes e históricos distintos
✓ Isolamento de versionamento: nenhum repositório contém o outro
✓ Caminhos do manifesto: válidos
✓ Diretórios obrigatórios: product, specs, decisions, policies e runs encontrados
✓ Git: disponível
✓ Permissões: leitura e escrita confirmadas
✓ Versão do manifesto: versão 1 implícita e suportada
Workspace saudável
`

type repositoryExpectation struct {
	Title     string
	Path      string
	Branch    string
	Commit    string
	State     string
	Modified  int
	Staged    int
	Untracked int
}

func expectedStatus(root, project string, repositories ...repositoryExpectation) string {
	var output strings.Builder
	output.WriteString("Projeto: " + project + "\n")
	output.WriteString("Workspace: " + displayPath(root) + "\n\n")
	for index, repository := range repositories {
		if index > 0 {
			output.WriteString("\n")
		}
		output.WriteString(repository.Title + "\n")
		output.WriteString("  Caminho: " + displayPath(repository.Path) + "\n")
		output.WriteString("  Branch: " + repository.Branch + "\n")
		output.WriteString("  Commit: " + repository.Commit + "\n")
		output.WriteString("  Estado: " + repository.State + "\n")
		output.WriteString("  Modificados: " + strconv.Itoa(repository.Modified) + "\n")
		output.WriteString("  Em stage: " + strconv.Itoa(repository.Staged) + "\n")
		output.WriteString("  Não rastreados: " + strconv.Itoa(repository.Untracked) + "\n")
	}
	return output.String()
}

func buildCLI(t *testing.T) string {
	t.Helper()
	name := "cerne"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	binary := filepath.Join(t.TempDir(), name)
	command := exec.Command("go", "build", "-o", binary, ".")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("go build: %v: %s", err, output)
	}
	return binary
}

func initWorkspaceWithCLI(t *testing.T, binary, parent, name string) string {
	t.Helper()
	status, stdout, stderr := executeCLI(t, binary, parent, nil, "init", name)
	if status != 0 {
		t.Fatalf("init status = %d\nstdout = %q\nstderr = %q", status, stdout, stderr)
	}
	return filepath.Join(parent, name)
}

func countReportLines(stdout string) int {
	count := 0
	for _, line := range strings.Split(strings.TrimSuffix(stdout, "\n"), "\n") {
		if strings.HasPrefix(line, "✓ ") || strings.HasPrefix(line, "✗ ") || strings.HasPrefix(line, "! ") {
			count++
		}
	}
	return count
}

type snapshotEntry struct {
	Mode    fs.FileMode
	Size    int64
	ModTime int64
	Hash    [32]byte
}

func snapshotTree(t *testing.T, root string) map[string]snapshotEntry {
	t.Helper()
	out := map[string]snapshotEntry{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		item := snapshotEntry{Mode: info.Mode(), Size: info.Size()}
		if info.Mode().IsRegular() {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			item.Hash = sha256.Sum256(data)
			item.ModTime = info.ModTime().UnixNano()
		}
		out[relative] = item
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func buildWorkflowTools(t *testing.T, provider string, fail bool) string {
	t.Helper()
	directory := t.TempDir()
	source := filepath.Join(directory, "main.go")
	program := `package main
import ("fmt"; "os"; "os/exec"; "path/filepath"; "strings")
func main() {
  executable, _ := os.Executable()
  name := strings.TrimSuffix(filepath.Base(executable), filepath.Ext(executable))
  if name == "git" {
    for _, argument := range os.Args[1:] {
      if argument == "clone" {
        if _, err := os.Stat(filepath.Join(filepath.Dir(executable), "fail")); err == nil {
          destination := os.Args[len(os.Args)-1]
          _ = os.WriteFile(filepath.Join(destination, "partial"), []byte("partial"), 0644)
          fmt.Fprintln(os.Stderr, "SECRET raw Git output")
          os.Exit(9)
        }
      }
    }
    if realGit := os.Getenv("CERNE_TEST_REAL_GIT"); realGit != "" {
      command := exec.Command(realGit, os.Args[1:]...)
      command.Stdin, command.Stdout, command.Stderr = os.Stdin, os.Stdout, os.Stderr
      if err := command.Run(); err != nil { os.Exit(1) }
      return
    }
    if err := os.Mkdir(".git", 0755); err != nil { os.Exit(1) }
    return
  }
  root, marker := ".specify", ".specify/init-options.json"
  if name == "openspec" { root, marker = "openspec", "openspec/config.yaml" }
  if _, err := os.Stat(filepath.Join(filepath.Dir(executable), "fail")); err == nil {
    _ = os.MkdirAll(root, 0755); fmt.Fprintln(os.Stderr, "SECRET raw provider output"); os.Exit(9)
  }
  if err := os.MkdirAll(filepath.Dir(marker), 0755); err != nil { os.Exit(1) }
  if name == "openspec" { _ = os.MkdirAll(filepath.Join(root, "specs"), 0755) }
  if err := os.WriteFile(marker, []byte("configured"), 0644); err != nil { os.Exit(1) }
  record := strings.Join(os.Args[1:], "\n") + "\n" + strings.Join(os.Environ(), "\n")
  if err := os.WriteFile(filepath.Join(root, "record"), []byte(record), 0644); err != nil { os.Exit(1) }
}`
	if err := os.WriteFile(source, []byte(program), 0o644); err != nil {
		t.Fatal(err)
	}
	tool := filepath.Join(directory, "workflow-test-tool")
	if runtime.GOOS == "windows" {
		tool += ".exe"
	}
	command := exec.Command("go", "build", "-o", tool, source)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build workflow tool: %v: %s", err, output)
	}
	installToolBinary(t, tool, filepath.Join(directory, executableFilename("git")))
	if provider != "" {
		installWorkflowTool(t, directory, provider)
	}
	if fail {
		if err := os.WriteFile(filepath.Join(directory, "fail"), []byte("1"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return directory
}

func installWorkflowTool(t *testing.T, directory, provider string) {
	t.Helper()
	name := map[string]string{"speckit": "specify", "openspec": "openspec"}[provider]
	installToolBinary(t, filepath.Join(directory, executableFilename("workflow-test-tool")), filepath.Join(directory, executableFilename(name)))
}

func installToolBinary(t *testing.T, source, destination string) {
	t.Helper()
	data, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(destination, data, 0o755); err != nil {
		t.Fatal(err)
	}
}

func executableFilename(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

func executeCLI(t *testing.T, binary, directory string, environment []string, args ...string) (int, string, string) {
	t.Helper()
	command := exec.Command(binary, args...)
	command.Dir = directory
	if environment != nil {
		command.Env = environment
	}
	var stdout, stderr strings.Builder
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	if err == nil {
		return 0, stdout.String(), stderr.String()
	}
	exit, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("executar CLI: %v", err)
	}
	return exit.ExitCode(), stdout.String(), stderr.String()
}

func replaceEnvironment(environment []string, name, value string) []string {
	prefix := name + "="
	replaced := make([]string, 0, len(environment)+1)
	for _, entry := range environment {
		if !strings.EqualFold(strings.SplitN(entry, "=", 2)[0]+"=", prefix) {
			replaced = append(replaced, entry)
		}
	}
	return append(replaced, prefix+value)
}

func gitOutput(t *testing.T, repository string, args ...string) string {
	t.Helper()
	output, err := exec.Command("git", append([]string{"-C", repository}, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v: %s", strings.Join(args, " "), err, output)
	}
	return strings.TrimSpace(string(output))
}

func samePath(left, right string) bool {
	if resolved, err := filepath.EvalSymlinks(left); err == nil {
		left = resolved
	}
	if resolved, err := filepath.EvalSymlinks(right); err == nil {
		right = resolved
	}
	return strings.EqualFold(filepath.Clean(left), filepath.Clean(right))
}

func displayPath(path string) string {
	if abs, err := filepath.Abs(path); err == nil {
		path = abs
	}
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		path = resolved
	}
	return filepath.Clean(path)
}

func containsPathAlias(text, path string) bool {
	return strings.Contains(text, filepath.Clean(path)) || strings.Contains(text, displayPath(path))
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
