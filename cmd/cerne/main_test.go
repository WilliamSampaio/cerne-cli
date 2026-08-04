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

const expectedInitHelp = `Inicializa um workspace Cerne com repositórios Git independentes.

Uso:
  cerne init <project-name>

Nome:
  1 a 255 caracteres ASCII; começa por letra ou número e continua com
  letras, números, ponto, hífen ou sublinhado. Nomes reservados e ponto final
  não são aceitos.

Estrutura:
  <project-name>/
  ├── knowledge/
  │   ├── cerne.json
  │   ├── product/
  │   ├── specs/
  │   ├── decisions/
  │   ├── policies/
  │   └── runs/
  └── source/

Efeitos:
  Cria dois repositórios Git locais vazios, sem commit ou remoto.
  Não acessa a rede e não altera conteúdo existente.

Saídas:
  Sucesso e ajuda usam stdout. Erros usam stderr.
  Status 0: sucesso ou ajuda; 1: falha operacional; 2: uso ou nome inválido.

Erros:
  O destino deve estar ausente ou ser um diretório regular vazio.
  Instale Git, corrija o nome ou escolha outro destino conforme o diagnóstico.

Exemplo:
  cerne init exemplo
`

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
		expected := "erro: informe exatamente um nome de projeto\nuso: cerne init <project-name>\n"
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
