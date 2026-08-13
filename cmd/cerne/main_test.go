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

	"github.com/WilliamSampaio/cerne-cli/internal/localization"
)

const expectedGlobalHelp = `Cerne administra workspaces com repositórios Git independentes de conhecimento e código-fonte.

Uso:
  cerne [--lang <en|pt-BR>] <comando> [argumentos]
  cerne --help
  cerne --version

Comandos:
  init      Cria um workspace Cerne
  restore   Restaura um workspace Cerne existente
  doctor    Valida a estrutura e a segurança do workspace
  status    Exibe o estado local dos repositórios
  link      Vincula um repositório Git local como source
  workflow  Inicializa o workflow declarado no workspace
  context   Exibe o contexto estrutural do workspace
  skill     Instala skills Cerne no perfil do agente
  git       Coordena inspeção Git segura
  config    Administra preferências do usuário

Opções:
  --lang       Usa en ou pt-BR somente nesta execução
  --help       Exibe esta ajuda
  --version    Exibe a versão do Cerne

Idioma:
  CERNE_LANG seleciona temporariamente o idioma. A ordem é --lang, CERNE_LANG,
  preferência salva e pt-BR. Use "cerne config --help" para persistir a escolha.

Use "cerne <comando> --help" para detalhes de um comando.
`

const expectedInitHelp = `Inicializa um workspace Cerne com repositórios Git independentes.

Uso:
  cerne init <project-name>
  cerne init <project-name> --source <caminho>
  cerne init <project-name> --clone <origem>
  cerne init <project-name> --workflow <speckit|openspec>
  cerne init <project-name> --workflow speckit --agent <codex|claude>
  cerne init <project-name> --source <caminho> --workflow <speckit|openspec>
  cerne init <project-name> --clone <origem> --workflow <speckit|openspec>

Nome:
  1 a 255 caracteres ASCII; começa por letra ou número e continua com
  letras, números, ponto, hífen ou sublinhado. Nomes reservados e ponto final
  não são aceitos.

Source:
  Sem flag, cria source como repositório Git vazio. --source vincula a raiz de
  um working tree Git local non-bare, resolvido a partir do diretório atual,
  sem criar source interno nem alterar o repositório informado. --clone aceita
  path local, file, HTTPS ou SSH, inclusive SCP-like, e cria remoto origin.
  --source e --clone são exclusivos; --workflow pode acompanhar qualquer modo.
  As opções ficam após o nome e combinações aceitam qualquer ordem.

Clone:
  HTTP, git://, ext::, helpers desconhecidos, opções, credenciais embutidas,
  query e fragmento são recusados. O clone é completo, sem submódulos, usa o
  checkout padrão e pode seguir redirects, autenticação externa e filtros já
  configurados no Git. Prompts controláveis pelo Git ficam desabilitados;
  helpers externos ainda podem falhar ou ter comportamento próprio.

Workflow:
  Sem a opção, mantém o layout padrão em knowledge/specs. Spec Kit também usa
  specs e cria .specify. OpenSpec usa openspec/specs e cria openspec.
  --agent codex|claude pode acompanhar somente --workflow speckit. A escolha
  é local, cria descoberta na raiz do workspace e não entra no manifesto.
  O Cerne usa somente instalações locais existentes, sem instalar agentes,
  atualizar providers ou fornecer credenciais. Se ausente, o setup fica pendente.
  Para instalar a skill global, use cerne skill install <agent>.

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
  cerne init exemplo --workflow speckit --agent codex
  cerne init exemplo --clone https://host/organizacao/aplicacao.git --workflow speckit
`

func TestCLIGlobalHelpAndVersion(t *testing.T) {
	binary := buildCLI(t)
	for _, test := range []struct {
		argument string
		expected string
	}{
		{"--help", expectedGlobalHelp},
		{"--version", "cerne 0.11.0\n"},
	} {
		t.Run(test.argument, func(t *testing.T) {
			status, stdout, stderr := executeCLI(t, binary, t.TempDir(), nil, test.argument)
			if status != 0 || stdout != test.expected || stderr != "" {
				t.Fatalf("status = %d\nstdout = %q\nstderr = %q", status, stdout, stderr)
			}
		})
	}
}

func TestCLIVersionCanBeOverriddenAtBuildTime(t *testing.T) {
	name := "cerne"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	binary := filepath.Join(t.TempDir(), name)
	command := exec.Command("go", "build", "-ldflags", "-X main.version=v9.8.7", "-o", binary, ".")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("go build: %v: %s", err, output)
	}
	status, stdout, stderr := executeCLI(t, binary, t.TempDir(), nil, "--version")
	if status != 0 || stdout != "cerne v9.8.7\n" || stderr != "" {
		t.Fatalf("status = %d\nstdout = %q\nstderr = %q", status, stdout, stderr)
	}
}

func TestCLIConfigLanguage(t *testing.T) {
	binary := buildCLI(t)
	home := t.TempDir()
	environment := skillHomeEnvironment(home)

	status, stdout, stderr := executeCLI(t, binary, t.TempDir(), environment, "config", "set", "language", "en")
	if status != 0 || stdout != "Idioma salvo: en\n" || stderr != "" {
		t.Fatalf("set: status=%d stdout=%q stderr=%q", status, stdout, stderr)
	}
	if got := readTestFile(t, filepath.Join(home, ".cerne", "config.json")); got != "{\n  \"language\": \"en\"\n}\n" {
		t.Fatalf("config = %q", got)
	}

	status, stdout, stderr = executeCLI(t, binary, t.TempDir(), environment, "config", "get", "language")
	if status != 0 || stdout != "Saved language: en\n" || stderr != "" {
		t.Fatalf("get: status=%d stdout=%q stderr=%q", status, stdout, stderr)
	}
	status, stdout, stderr = executeCLI(t, binary, t.TempDir(), environment, "--help")
	if status != 0 || !strings.HasPrefix(stdout, "Cerne manages workspaces") || stderr != "" {
		t.Fatalf("english help: status=%d stdout=%q stderr=%q", status, stdout, stderr)
	}

	status, stdout, stderr = executeCLI(t, binary, t.TempDir(), environment, "config", "unset", "language")
	if status != 0 || stdout != "Language preference removed.\n" || stderr != "" {
		t.Fatalf("unset: status=%d stdout=%q stderr=%q", status, stdout, stderr)
	}
	status, stdout, stderr = executeCLI(t, binary, t.TempDir(), environment, "config", "get", "language")
	if status != 0 || stdout != "Idioma não definido. Padrão atual: pt-BR\n" || stderr != "" {
		t.Fatalf("get unset: status=%d stdout=%q stderr=%q", status, stdout, stderr)
	}
}

func TestCLILanguagePrecedenceDoesNotChangePreference(t *testing.T) {
	binary := buildCLI(t)
	home := t.TempDir()
	environment := skillHomeEnvironment(home)
	status, _, stderr := executeCLI(t, binary, t.TempDir(), environment, "config", "set", "language", "pt-BR")
	if status != 0 || stderr != "" {
		t.Fatalf("set: status=%d stderr=%q", status, stderr)
	}

	englishEnvironment := replaceEnvironment(environment, "CERNE_LANG", "en")
	status, stdout, stderr := executeCLI(t, binary, t.TempDir(), englishEnvironment, "--help")
	if status != 0 || !strings.HasPrefix(stdout, "Cerne manages workspaces") || stderr != "" {
		t.Fatalf("environment: status=%d stdout=%q stderr=%q", status, stdout, stderr)
	}
	status, stdout, stderr = executeCLI(t, binary, t.TempDir(), englishEnvironment, "--lang", "pt-BR", "--help")
	if status != 0 || !strings.HasPrefix(stdout, "Cerne administra workspaces") || stderr != "" {
		t.Fatalf("flag: status=%d stdout=%q stderr=%q", status, stdout, stderr)
	}
	status, stdout, stderr = executeCLI(t, binary, t.TempDir(), environment, "config", "get", "language")
	if status != 0 || stdout != "Idioma salvo: pt-BR\n" || stderr != "" {
		t.Fatalf("saved: status=%d stdout=%q stderr=%q", status, stdout, stderr)
	}
}

func TestCLIRejectsInvalidLanguageWithoutChangingConfig(t *testing.T) {
	binary := buildCLI(t)
	home := t.TempDir()
	environment := skillHomeEnvironment(home)
	status, stdout, stderr := executeCLI(t, binary, t.TempDir(), environment, "config", "set", "language", "es")
	if status != 2 || stdout != "" || stderr != "erro: idioma inválido: \"es\"\ncorreção: use en ou pt-BR\n" {
		t.Fatalf("set: status=%d stdout=%q stderr=%q", status, stdout, stderr)
	}
	if _, err := os.Stat(filepath.Join(home, ".cerne")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("uso inválido criou configuração: %v", err)
	}
	status, stdout, stderr = executeCLI(t, binary, t.TempDir(), environment, "--lang", "es", "--help")
	if status != 2 || stdout != "" || stderr != "erro: idioma inválido: \"es\"\ncorreção: use en ou pt-BR\n" {
		t.Fatalf("flag: status=%d stdout=%q stderr=%q", status, stdout, stderr)
	}
}

func TestCLIEnglishHelpForEveryCommand(t *testing.T) {
	binary := buildCLI(t)
	environment := skillHomeEnvironment(t.TempDir())
	commands := [][]string{
		{"--help"},
		{"init", "--help"},
		{"restore", "--help"},
		{"doctor", "--help"},
		{"status", "--help"},
		{"link", "--help"},
		{"workflow", "--help"},
		{"context", "--help"},
		{"skill", "--help"},
		{"config", "--help"},
	}
	for _, args := range commands {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			arguments := append([]string{"--lang", "en"}, args...)
			status, stdout, stderr := executeCLI(t, binary, t.TempDir(), environment, arguments...)
			if status != 0 || stderr != "" || !strings.Contains(stdout, "Usage:") {
				t.Fatalf("status=%d stdout=%q stderr=%q", status, stdout, stderr)
			}
			for _, portuguese := range []string{"Uso:", "Saídas:", "Efeitos:", "Exemplo:", "correção:"} {
				if strings.Contains(stdout, portuguese) {
					t.Fatalf("help contém português %q: %q", portuguese, stdout)
				}
			}
		})
	}
}

func TestCLIEnglishReportsAndJSONInvariance(t *testing.T) {
	binary := buildCLI(t)
	parent := t.TempDir()
	home := t.TempDir()
	environment := skillHomeEnvironment(home)

	status, stdout, stderr := executeCLI(t, binary, parent, environment, "--lang", "en", "init", "example")
	if status != 0 || stderr != "" || !strings.Contains(stdout, "Created workspace \"example\".") {
		t.Fatalf("init: status=%d stdout=%q stderr=%q", status, stdout, stderr)
	}
	root := filepath.Join(parent, "example")

	for _, test := range []struct {
		args []string
		want string
	}{
		{args: []string{"doctor"}, want: "Healthy workspace\n"},
		{args: []string{"status"}, want: "Project: example\n"},
		{args: []string{"context"}, want: "Status: healthy\n"},
		{args: []string{"link", "source"}, want: "No changes required.\n"},
	} {
		arguments := append([]string{"--lang", "en"}, test.args...)
		status, stdout, stderr = executeCLI(t, binary, root, environment, arguments...)
		if status != 0 || stderr != "" || !strings.Contains(stdout, test.want) {
			t.Fatalf("%v: status=%d stdout=%q stderr=%q", test.args, status, stdout, stderr)
		}
		for _, portuguese := range []string{"Projeto:", "Caminho:", "Estado:", "saudável", "Nenhuma alteração"} {
			if strings.Contains(stdout, portuguese) {
				t.Fatalf("%v contém português %q: %q", test.args, portuguese, stdout)
			}
		}
	}

	statusEN, jsonEN, stderrEN := executeCLI(t, binary, root, environment, "--lang", "en", "context", "--json")
	statusPT, jsonPT, stderrPT := executeCLI(t, binary, root, environment, "--lang", "pt-BR", "context", "--json")
	if statusEN != statusPT || jsonEN != jsonPT || stderrEN != stderrPT {
		t.Fatalf("JSON varia por idioma:\nen=%d %q %q\npt=%d %q %q", statusEN, jsonEN, stderrEN, statusPT, jsonPT, stderrPT)
	}
}

func TestCLIEnglishFailuresAndNeutralVersion(t *testing.T) {
	binary := buildCLI(t)
	environment := skillHomeEnvironment(t.TempDir())
	for _, test := range []struct {
		args   []string
		status int
		want   string
	}{
		{args: []string{"unknown"}, status: 2, want: "error: unknown command\n"},
		{args: []string{"status", "extra"}, status: 2, want: "error: invalid argument\n"},
		{args: []string{"workflow", "setup"}, status: 1, want: "error: Cerne workspace not found"},
	} {
		arguments := append([]string{"--lang", "en"}, test.args...)
		status, stdout, stderr := executeCLI(t, binary, t.TempDir(), environment, arguments...)
		if status != test.status || stdout != "" || !strings.Contains(stderr, test.want) {
			t.Fatalf("%v: status=%d stdout=%q stderr=%q", test.args, status, stdout, stderr)
		}
		if strings.Contains(stderr, "erro:") || strings.Contains(stderr, "correção:") {
			t.Fatalf("%v contém português: %q", test.args, stderr)
		}
	}
	status, stdout, stderr := executeCLI(t, binary, t.TempDir(), environment, "--lang", "en", "--version")
	if status != 0 || stdout != "cerne 0.11.0\n" || stderr != "" {
		t.Fatalf("version: status=%d stdout=%q stderr=%q", status, stdout, stderr)
	}
}

func TestSkillInstallCommand(t *testing.T) {
	binary := buildCLI(t)

	t.Run("help", func(t *testing.T) {
		status, stdout, stderr := executeCLI(t, binary, t.TempDir(), nil, "skill", "--help")
		if status != 0 || stderr != "" || !strings.Contains(stdout, "cerne skill install <codex|claude|gemini>") ||
			!strings.Contains(stdout, "cerne skill install <codex|claude|gemini> <cerne-context|cerne-git-workflow>") {
			t.Fatalf("help: status=%d stdout=%q stderr=%q", status, stdout, stderr)
		}
		status, stdout, stderr = executeCLI(t, binary, t.TempDir(), nil, "skill", "install", "--help")
		if status != 0 || stderr != "" || !strings.Contains(stdout, "cerne skill install <codex|claude|gemini>") ||
			!strings.Contains(stdout, "cerne skill install <codex|claude|gemini> <cerne-context|cerne-git-workflow>") {
			t.Fatalf("install help: status=%d stdout=%q stderr=%q", status, stdout, stderr)
		}
	})

	t.Run("default installs all supported skills", func(t *testing.T) {
		for _, agent := range []string{"codex", "claude", "gemini"} {
			home := t.TempDir()
			status, stdout, stderr := executeCLI(t, binary, t.TempDir(), skillHomeEnvironment(home), "skill", "install", agent)
			if status != 0 || stderr != "" || !strings.Contains(stdout, "Agente: "+agent+"\n") || !strings.Contains(stdout, "Versão: 0.4.2\n") {
				t.Fatalf("%s: status=%d stdout=%q stderr=%q", agent, status, stdout, stderr)
			}
			root := map[string]string{"codex": ".codex", "claude": ".claude", "gemini": ".gemini"}[agent]
			want := []string{"cerne-git-workflow"}
			if agent != "gemini" {
				want = append([]string{"cerne-context"}, want...)
			}
			for _, skill := range want {
				target := filepath.Join(home, root, "skills", skill)
				if !strings.Contains(stdout, "Skill instalada: "+skill+"\n") || !strings.Contains(stdout, "Destino: "+target+"\n") {
					t.Fatalf("%s/%s não apareceu no output: %q", agent, skill, stdout)
				}
				if !strings.Contains(readTestFile(t, filepath.Join(target, "SKILL.md")), "name: "+skill) {
					t.Fatalf("%s/%s skill não instalada", agent, skill)
				}
			}
		}
	})

	t.Run("named git workflow skill", func(t *testing.T) {
		for _, agent := range []string{"codex", "claude", "gemini"} {
			home := t.TempDir()
			status, stdout, stderr := executeCLI(t, binary, t.TempDir(), skillHomeEnvironment(home), "skill", "install", agent, "cerne-git-workflow")
			target := filepath.Join(home, map[string]string{
				"codex":  ".codex",
				"claude": ".claude",
				"gemini": ".gemini",
			}[agent], "skills", "cerne-git-workflow")
			if status != 0 || stderr != "" || !strings.Contains(stdout, "Skill instalada: cerne-git-workflow\n") ||
				!strings.Contains(stdout, "Agente: "+agent+"\n") || !strings.Contains(stdout, "Destino: "+target+"\n") {
				t.Fatalf("%s: status=%d stdout=%q stderr=%q", agent, status, stdout, stderr)
			}
			if !strings.Contains(readTestFile(t, filepath.Join(target, "SKILL.md")), "name: cerne-git-workflow") {
				t.Fatalf("%s git workflow skill não instalada", agent)
			}
		}
	})

	t.Run("invalid usage", func(t *testing.T) {
		for _, args := range [][]string{
			{"skill", "install"},
			{"skill", "install", "generic"},
			{"skill", "install", "Codex"},
			{"skill", "install", "codex", "extra"},
			{"skill", "install", "gemini", "cerne-context"},
			{"skill", "install", "codex", "Cerne-Git-Workflow"},
		} {
			home := t.TempDir()
			status, stdout, stderr := executeCLI(t, binary, t.TempDir(), skillHomeEnvironment(home), args...)
			if status != 2 || stdout != "" || stderr != "erro: argumento inválido\nuso: cerne skill install <codex|claude|gemini> [cerne-context|cerne-git-workflow]\n" {
				t.Fatalf("%v: status=%d stdout=%q stderr=%q", args, status, stdout, stderr)
			}
			if _, err := os.Stat(filepath.Join(home, ".cerne", "audit")); !os.IsNotExist(err) {
				t.Fatalf("%v criou audit em uso inválido: %v", args, err)
			}
		}
	})
}

func TestWorkspaceCommandsDoNotInstallGlobalSkills(t *testing.T) {
	binary := buildCLI(t)

	t.Run("init and workflow setup", func(t *testing.T) {
		parent, home := t.TempDir(), t.TempDir()
		tools := buildWorkflowTools(t, "speckit", false)
		env := skillHomeEnvironment(home)
		env = replaceEnvironment(env, "PATH", tools)
		status, stdout, stderr := executeCLI(t, binary, parent, env, "init", "app", "--workflow", "speckit", "--agent", "codex")
		if status != 0 || stderr != "" || !strings.Contains(stdout, "Descoberta: pronta") {
			t.Fatalf("init: status=%d stdout=%q stderr=%q", status, stdout, stderr)
		}
		assertNoGlobalSkills(t, home)

		status, stdout, stderr = executeCLI(t, binary, filepath.Join(parent, "app"), env, "workflow", "setup", "--agent", "claude")
		if status != 0 || stderr != "" || !strings.Contains(stdout, "Descoberta: pronta") {
			t.Fatalf("workflow setup: status=%d stdout=%q stderr=%q", status, stdout, stderr)
		}
		assertNoGlobalSkills(t, home)
	})

	t.Run("restore", func(t *testing.T) {
		parent, home := t.TempDir(), t.TempDir()
		knowledge := createRestoreGitRepository(t, map[string]string{
			"cerne.json":       `{"name":"example","source":"../source"}`,
			"product/.gitkeep": "", "specs/.gitkeep": "", "decisions/.gitkeep": "",
			"policies/.gitkeep": "", "runs/.gitkeep": "",
		})
		source := createRestoreGitRepository(t, map[string]string{"README.md": "source\n"})
		status, stdout, stderr := executeCLI(t, binary, parent, skillHomeEnvironment(home), "restore", knowledge, "--source", source)
		if status != 0 || stderr != "" || !strings.Contains(stdout, "Source vinculado:") {
			t.Fatalf("restore: status=%d stdout=%q stderr=%q", status, stdout, stderr)
		}
		assertNoGlobalSkills(t, home)
	})
}

func TestCLIContextJSONContract(t *testing.T) {
	binary := buildCLI(t)
	root := filepath.Join(t.TempDir(), "example")
	for _, path := range []string{"knowledge/product", "knowledge/specs", "knowledge/decisions", "knowledge/policies", "knowledge/runs", "source", "source/sub"} {
		if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(path)), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	manifest := `{"name":"example","source":"../source","version":1}`
	if err := os.WriteFile(filepath.Join(root, "knowledge", "cerne.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	root = canonicalCLIPath(t, root)
	before := snapshotTree(t, root)
	status, stdout, stderr := executeCLI(t, binary, filepath.Join(root, "source", "sub"), replaceEnvironment(os.Environ(), "PATH", t.TempDir()), "context", "--json")
	if status != 0 || stderr != "" || !strings.HasSuffix(stdout, "\n") {
		t.Fatalf("status=%d stdout=%q stderr=%q", status, stdout, stderr)
	}
	expected := "{\n  \"schema_version\": 1,\n  \"status\": \"healthy\",\n  \"workspace\": {\n    \"name\": \"example\",\n    \"root\": " + strconv.Quote(root) + "\n  },\n  \"knowledge\": {\n    \"path\": " + strconv.Quote(filepath.Join(root, "knowledge")) + ",\n    \"product_path\": " + strconv.Quote(filepath.Join(root, "knowledge", "product")) + ",\n    \"specs_path\": " + strconv.Quote(filepath.Join(root, "knowledge", "specs")) + ",\n    \"decisions_path\": " + strconv.Quote(filepath.Join(root, "knowledge", "decisions")) + ",\n    \"policies_path\": " + strconv.Quote(filepath.Join(root, "knowledge", "policies")) + "\n  },\n  \"source\": {\n    \"path\": " + strconv.Quote(filepath.Join(root, "source")) + ",\n    \"inside_workspace\": true\n  },\n  \"workflow\": {\n    \"declared\": false,\n    \"state\": \"not-declared\"\n  },\n  \"problems\": []\n}\n"
	if stdout != expected {
		t.Fatalf("stdout = %q\nesperado = %q", stdout, expected)
	}
	_, repeated, _ := executeCLI(t, binary, root, nil, "context", "--json")
	if repeated != stdout {
		t.Fatalf("saída não determinística")
	}
	if after := snapshotTree(t, root); !reflect.DeepEqual(before, after) {
		t.Fatalf("context alterou o workspace\nantes=%#v\ndepois=%#v", before, after)
	}
}

func TestCLIContextHumanHelpAndFailures(t *testing.T) {
	binary := buildCLI(t)
	root := filepath.Join(t.TempDir(), "example")
	for _, path := range []string{"knowledge/product", "knowledge/specs", "knowledge/decisions", "knowledge/policies", "knowledge/runs", "source"} {
		if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(path)), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "knowledge", "cerne.json"), []byte(`{"name":"example","source":"../source","version":1}`), 0o644); err != nil {
		t.Fatal(err)
	}
	root = canonicalCLIPath(t, root)
	status, stdout, stderr := executeCLI(t, binary, root, nil, "context")
	expected := "Workspace: example\nStatus: saudável\nRoot: " + root + "\n\nKnowledge: " + filepath.Join(root, "knowledge") + "\nProduct: " + filepath.Join(root, "knowledge", "product") + "\nSpecs: " + filepath.Join(root, "knowledge", "specs") + "\nDecisions: " + filepath.Join(root, "knowledge", "decisions") + "\nPolicies: " + filepath.Join(root, "knowledge", "policies") + "\n\nSource: " + filepath.Join(root, "source") + "\nLocalização: interno ao workspace\n\nWorkflow: não declarado\n"
	if status != 0 || stdout != expected || stderr != "" {
		t.Fatalf("status=%d\nstdout=%q\nstderr=%q", status, stdout, stderr)
	}

	status, stdout, stderr = executeCLI(t, binary, t.TempDir(), nil, "context", "--help")
	if status != 0 || stderr != "" || !strings.Contains(stdout, "cerne context --json") || !strings.Contains(stdout, "Leitura estrutural exclusiva") {
		t.Fatalf("help: status=%d stdout=%q stderr=%q", status, stdout, stderr)
	}

	outside := t.TempDir()
	status, stdout, stderr = executeCLI(t, binary, outside, nil, "context", "--json")
	expectedJSON := "{\n  \"schema_version\": 1,\n  \"status\": \"invalid\",\n  \"problems\": [\n    {\n      \"code\": \"workspace-not-found\",\n      \"severity\": \"error\",\n      \"component\": \"workspace\"\n    }\n  ]\n}\n"
	if status != 1 || stdout != expectedJSON || stderr != "" {
		t.Fatalf("outside: status=%d stdout=%q stderr=%q", status, stdout, stderr)
	}

	status, stdout, stderr = executeCLI(t, binary, outside, nil, "context", "--json", "extra")
	if status != 2 || stdout != "" || stderr != "erro: argumento inválido\nuso: cerne context [--json]\n" {
		t.Fatalf("usage: status=%d stdout=%q stderr=%q", status, stdout, stderr)
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
		expected := "erro: argumento inválido\nuso: cerne init <project-name> [--source <caminho> | --clone <origem>] [--workflow <speckit|openspec> [--agent <codex|claude>]]\n"
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
	before := snapshotTreeWithoutGit(t, source)
	worktreeBefore := snapshotTreeWithoutGit(t, worktree)
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
	if !reflect.DeepEqual(before, snapshotTreeWithoutGit(t, source)) {
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
	if !reflect.DeepEqual(worktreeBefore, snapshotTreeWithoutGit(t, worktree)) || !reflect.DeepEqual(before, snapshotTreeWithoutGit(t, source)) {
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
		{"init", "example", "--source", "path", "--clone", "origin", "--workflow", "speckit"},
		{"init", "example", "--workflow", "speckit", "--workflow", "openspec"},
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

func TestCLIInitCombinesWorkflowWithEitherSourceMode(t *testing.T) {
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Skip("Git não está disponível")
	}
	binary := buildCLI(t)
	for _, test := range []struct {
		name, provider, sourceLabel, marker string
		clone                               bool
	}{
		{"local", "speckit", "Source vinculado:", ".specify/init-options.json", false},
		{"clone", "openspec", "Source clonado:", "openspec/config.yaml", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			parent := t.TempDir()
			source := filepath.Join(parent, "origin")
			if err := os.Mkdir(source, 0o755); err != nil {
				t.Fatal(err)
			}
			gitOutput(t, source, "init", "--quiet")
			before := snapshotTree(t, source)
			tools := buildWorkflowTools(t, test.provider, false)
			environment := replaceEnvironment(os.Environ(), "PATH", tools)
			environment = replaceEnvironment(environment, "CERNE_TEST_REAL_GIT", realGit)
			args := []string{"init", "example", "--workflow", test.provider, "--source", source}
			if test.clone {
				args = []string{"init", "example", "--clone", source, "--workflow", test.provider}
			}
			status, stdout, stderr := executeCLI(t, binary, parent, environment, args...)
			knowledge := filepath.Join(parent, "example", "knowledge")
			sourcePath := source
			if test.clone {
				sourcePath = filepath.Join(parent, "example", "source")
			}
			lines := strings.Split(stdout, "\n")
			if status != 0 || stderr != "" || len(lines) != 6 || lines[0] != `Workspace "example" criado.` ||
				!samePath(strings.TrimPrefix(lines[1], "Knowledge: "), knowledge) ||
				!samePath(strings.TrimPrefix(lines[2], test.sourceLabel+" "), sourcePath) ||
				lines[3] != "Workflow: "+test.provider || lines[4] != "Setup: concluído" || lines[5] != "" {
				t.Fatalf("status=%d stdout=%q stderr=%q", status, stdout, stderr)
			}
			if !reflect.DeepEqual(before, snapshotTree(t, source)) {
				t.Fatal("source informado foi alterado")
			}
			if _, err := os.Stat(filepath.Join(knowledge, filepath.FromSlash(test.marker))); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(readFile(t, filepath.Join(knowledge, "cerne.json")), `"provider": "`+test.provider+`"`) {
				t.Fatal("workflow ausente do manifesto")
			}
		})
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

func TestCLISpecKitAgentDiscovery(t *testing.T) {
	binary := buildCLI(t)

	t.Run("init codex creates root bridge and keeps manifest neutral", func(t *testing.T) {
		tools := buildWorkflowTools(t, "speckit", false)
		parent := t.TempDir()
		environment := replaceEnvironment(os.Environ(), "PATH", tools)
		status, stdout, stderr := executeCLI(t, binary, parent, environment, "init", "codex-project", "--workflow", "speckit", "--agent", "codex")
		root := filepath.Join(parent, "codex-project")
		knowledge := filepath.Join(root, "knowledge")
		lines := strings.Split(stdout, "\n")
		if status != 0 || stderr != "" || len(lines) != 8 || lines[0] != `Workspace "codex-project" criado.` ||
			!samePath(strings.TrimPrefix(lines[1], "Knowledge: "), knowledge) ||
			!samePath(strings.TrimPrefix(lines[2], "Source: "), filepath.Join(root, "source")) ||
			lines[3] != "Workflow: speckit" || lines[4] != "Setup: concluído" ||
			lines[5] != "Agent: codex" || lines[6] != "Descoberta: pronta" || lines[7] != "" {
			t.Fatalf("status=%d stdout=%q stderr=%q", status, stdout, stderr)
		}
		for _, path := range []string{
			filepath.Join(knowledge, ".agents", "skills", "speckit-analyze", "SKILL.md"),
			filepath.Join(root, ".agents", "skills", "speckit-analyze", "SKILL.md"),
			filepath.Join(knowledge, ".specify", "init-options.json"),
		} {
			if _, err := os.Stat(path); err != nil {
				t.Fatal(err)
			}
		}
		manifest := readFile(t, filepath.Join(knowledge, "cerne.json"))
		if strings.Contains(manifest, "agent") || !strings.Contains(manifest, `"provider": "speckit"`) {
			t.Fatalf("manifesto inválido: %s", manifest)
		}
		bridge := readFile(t, filepath.Join(root, ".agents", "skills", "speckit-analyze", "SKILL.md"))
		if !strings.Contains(bridge, "knowledge/.agents/skills/speckit-analyze/SKILL.md") || strings.Contains(bridge, root) {
			t.Fatalf("ponte inválida: %s", bridge)
		}
	})

	t.Run("workflow setup claude refreshes local discovery", func(t *testing.T) {
		tools := buildWorkflowTools(t, "speckit", false)
		parent := t.TempDir()
		environment := replaceEnvironment(os.Environ(), "PATH", tools)
		status, _, stderr := executeCLI(t, binary, parent, environment, "init", "restored", "--workflow", "speckit", "--agent", "codex")
		if status != 0 || stderr != "" {
			t.Fatalf("init status=%d stderr=%q", status, stderr)
		}
		root := filepath.Join(parent, "restored")
		sourceBefore := snapshotTree(t, filepath.Join(root, "source"))
		status, stdout, stderr := executeCLI(t, binary, filepath.Join(root, "knowledge", "product"), environment, "workflow", "setup", "--agent", "claude")
		expected := "Workflow: speckit\nKnowledge: " + displayPath(filepath.Join(root, "knowledge")) + "\nNenhuma alteração necessária.\nAgent: claude\nDescoberta: pronta\n"
		if status != 0 || stdout != expected || stderr != "" {
			t.Fatalf("status=%d stdout=%q stderr=%q", status, stdout, stderr)
		}
		if _, err := os.Stat(filepath.Join(root, ".claude", "skills", "speckit-analyze", "SKILL.md")); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(sourceBefore, snapshotTree(t, filepath.Join(root, "source"))) {
			t.Fatal("source alterado")
		}
		if strings.Contains(readFile(t, filepath.Join(root, "knowledge", "cerne.json")), "agent") {
			t.Fatal("agente persistido no manifesto")
		}
	})

	t.Run("missing provider creates no fake bridge", func(t *testing.T) {
		tools := buildWorkflowTools(t, "", false)
		parent := t.TempDir()
		environment := replaceEnvironment(os.Environ(), "PATH", tools)
		status, stdout, stderr := executeCLI(t, binary, parent, environment, "init", "pending", "--workflow", "speckit", "--agent", "codex")
		root := filepath.Join(parent, "pending")
		if status != 0 || !strings.Contains(stdout, "Setup: pendente") || strings.Contains(stdout, "Agent:") ||
			!strings.Contains(stderr, `cerne workflow setup --agent codex`) {
			t.Fatalf("status=%d stdout=%q stderr=%q", status, stdout, stderr)
		}
		if _, err := os.Stat(filepath.Join(root, ".agents")); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("ponte falsa criada: %v", err)
		}
		status, stdout, stderr = executeCLI(t, binary, root, environment, "workflow", "setup", "--agent", "codex")
		if status != 1 || stdout != "" || !strings.Contains(stderr, `executável "specify" não encontrado`) {
			t.Fatalf("status=%d stdout=%q stderr=%q", status, stdout, stderr)
		}
		if _, err := os.Stat(filepath.Join(root, ".agents")); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("ponte falsa criada no setup: %v", err)
		}
	})

	t.Run("provider failure with agent creates no fake bridge", func(t *testing.T) {
		tools := buildWorkflowTools(t, "speckit", true)
		parent := t.TempDir()
		environment := replaceEnvironment(os.Environ(), "PATH", tools)
		status, stdout, stderr := executeCLI(t, binary, parent, environment, "init", "failed", "--workflow", "speckit", "--agent", "codex")
		root := filepath.Join(parent, "failed")
		if status != 1 || stdout != "" || strings.Contains(stderr, "SECRET") || !strings.Contains(stderr, "provider não concluiu") {
			t.Fatalf("status=%d stdout=%q stderr=%q", status, stdout, stderr)
		}
		if _, err := os.Stat(filepath.Join(root, ".agents")); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("ponte falsa criada: %v", err)
		}
		if _, err := os.Stat(filepath.Join(root, "source", ".git")); err != nil {
			t.Fatal(err)
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
		{"init", "example", "--agent", "codex"},
		{"init", "example", "--workflow", "openspec", "--agent", "codex"},
		{"init", "example", "--workflow", "speckit", "--agent", "generic"},
		{"init", "example", "--workflow", "speckit", "--agent", "codex", "--agent", "claude"},
		{"workflow", "setup", "extra"},
		{"workflow", "setup", "--agent"},
		{"workflow", "setup", "--agent", "generic"},
		{"workflow", "setup", "--agent", "codex", "extra"},
	} {
		status, stdout, stderr := executeCLI(t, binary, t.TempDir(), nil, args...)
		if status != 2 || stdout != "" || !strings.Contains(stderr, "uso:") {
			t.Fatalf("args=%v status=%d stdout=%q stderr=%q", args, status, stdout, stderr)
		}
	}
	status, stdout, stderr := executeCLI(t, binary, t.TempDir(), nil, "workflow", "--help")
	if status != 0 || stderr != "" || !strings.Contains(stdout, "cerne workflow setup --agent <codex|claude>") {
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
	status := runDoctor(nil, &stdout, &stderr, localizer{language: localization.Default})
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
	status := runStatus(nil, &stdout, &stderr, localizer{language: localization.Default})
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
	status := runLink([]string{"."}, &stdout, &stderr, localizer{language: localization.Default})
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

func canonicalCLIPath(t *testing.T, path string) string {
	t.Helper()
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Clean(resolved)
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

func skillHomeEnvironment(home string) []string {
	environment := replaceEnvironment(os.Environ(), "HOME", home)
	return replaceEnvironment(environment, "USERPROFILE", home)
}

func auditEntries(t *testing.T, home string) []os.DirEntry {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(home, ".cerne", "audit"))
	if err != nil {
		t.Fatal(err)
	}
	return entries
}

func assertNoGlobalSkills(t *testing.T, home string) {
	t.Helper()
	for _, path := range []string{
		filepath.Join(home, ".codex", "skills", "cerne-context"),
		filepath.Join(home, ".claude", "skills", "cerne-context"),
	} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("global skill destination exists: %s (%v)", path, err)
		}
	}
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readTestFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
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
	Mode fs.FileMode
	Size int64
	Hash [32]byte
}

func snapshotTree(t *testing.T, root string) map[string]snapshotEntry {
	return snapshotTreeFiltered(t, root, false)
}

func snapshotTreeWithoutGit(t *testing.T, root string) map[string]snapshotEntry {
	return snapshotTreeFiltered(t, root, true)
}

func snapshotTreeFiltered(t *testing.T, root string, skipGit bool) map[string]snapshotEntry {
	t.Helper()
	out := map[string]snapshotEntry{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if skipGit && (relative == ".git" || strings.HasPrefix(relative, ".git"+string(filepath.Separator))) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		item := snapshotEntry{Mode: info.Mode()}
		if !info.IsDir() {
			item.Size = info.Size()
		}
		if info.Mode().IsRegular() {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			item.Hash = sha256.Sum256(data)
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
  if name == "specify" && len(os.Args) >= 4 && os.Args[1] == "integration" && os.Args[2] == "install" {
    agentRoot := map[string]string{"codex": ".agents/skills", "claude": ".claude/skills"}[os.Args[3]]
    if agentRoot == "" { os.Exit(2) }
    commands := []string{"speckit-analyze","speckit-checklist","speckit-clarify","speckit-constitution","speckit-converge","speckit-implement","speckit-plan","speckit-specify","speckit-tasks","speckit-taskstoissues"}
    for _, command := range commands {
      skill := filepath.Join(agentRoot, command, "SKILL.md")
      if err := os.MkdirAll(filepath.Dir(skill), 0755); err != nil { os.Exit(1) }
      content := "---\nname: \""+command+"\"\ndescription: \"Run the "+command+" workflow.\"\n---\n\n# "+command+"\n"
      if err := os.WriteFile(skill, []byte(content), 0644); err != nil { os.Exit(1) }
    }
    record := strings.Join(os.Args[1:], "\n") + "\n" + strings.Join(os.Environ(), "\n")
    if err := os.WriteFile(filepath.Join(agentRoot, "record"), []byte(record), 0644); err != nil { os.Exit(1) }
    return
  }
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

func TestParseRestoreArgsRequiresOneExclusiveSourceMode(t *testing.T) {
	tests := []struct {
		args []string
		ok   bool
		mode string
	}{
		{[]string{"knowledge", "--clone", "source"}, true, "clone"},
		{[]string{"knowledge", "--source", "../source"}, true, "local"},
		{[]string{"--clone", "source", "knowledge"}, false, ""},
		{[]string{"knowledge"}, false, ""},
		{[]string{"knowledge", "--clone", "source", "extra"}, false, ""},
		{[]string{"knowledge", "--clone", "--source"}, false, ""},
	}
	for _, test := range tests {
		parsed, ok := parseRestoreArgs(test.args)
		if ok != test.ok || ok && string(parsed.SourceMode) != test.mode {
			t.Fatalf("parseRestoreArgs(%v) = %#v, %v", test.args, parsed, ok)
		}
	}
}

func TestCLIRestoreHelpAndInvalidUsage(t *testing.T) {
	binary := buildCLI(t)
	status, stdout, stderr := executeCLI(t, binary, t.TempDir(), nil, "restore", "--help")
	if status != 0 || stderr != "" || !strings.Contains(stdout, "cerne restore <origem-knowledge> --clone <origem-source>") ||
		!strings.Contains(stdout, "~/.cerne/audit") {
		t.Fatalf("help: status=%d stdout=%q stderr=%q", status, stdout, stderr)
	}
	status, stdout, stderr = executeCLI(t, binary, t.TempDir(), nil, "restore", "knowledge")
	expected := "erro: argumento inválido\nuso: cerne restore <origem-knowledge> (--source <caminho> | --clone <origem-source>)\n"
	if status != 2 || stdout != "" || stderr != expected {
		t.Fatalf("uso: status=%d stdout=%q stderr=%q", status, stdout, stderr)
	}
}

func TestCLIRestoreCloneEndToEnd(t *testing.T) {
	binary := buildCLI(t)
	parent, home := t.TempDir(), t.TempDir()
	knowledge := createRestoreGitRepository(t, map[string]string{
		"cerne.json":       `{"name":"example","source":"../source"}`,
		"product/.gitkeep": "", "specs/.gitkeep": "", "decisions/.gitkeep": "",
		"policies/.gitkeep": "", "runs/.gitkeep": "",
	})
	source := createRestoreGitRepository(t, map[string]string{"README.md": "source\n"})
	environment := replaceEnvironment(os.Environ(), "HOME", home)
	environment = replaceEnvironment(environment, "USERPROFILE", home)
	status, stdout, stderr := executeCLI(t, binary, parent, environment, "restore", knowledge, "--clone", source)
	expected := "Workspace \"example\" restaurado.\nKnowledge: " + displayPath(filepath.Join(parent, "example", "knowledge")) +
		"\nSource clonado: " + displayPath(filepath.Join(parent, "example", "source")) + "\n"
	if status != 0 || stdout != expected || stderr != "" {
		t.Fatalf("status = %d\nstdout = %q\nstderr = %q", status, stdout, stderr)
	}
	entries, err := os.ReadDir(filepath.Join(home, ".cerne", "audit"))
	if err != nil || len(entries) != 1 {
		t.Fatalf("audit = %d, %v", len(entries), err)
	}
	audit := readFile(t, filepath.Join(home, ".cerne", "audit", entries[0].Name()))
	if strings.Contains(audit, knowledge) || strings.Contains(audit, source) || !strings.Contains(audit, `"status": "succeeded"`) {
		t.Fatalf("audit inválido: %s", audit)
	}
}

func TestCLIRestoreLocalSourceUpdatesOnlyManifestReference(t *testing.T) {
	binary := buildCLI(t)
	parent, home := t.TempDir(), t.TempDir()
	knowledge := createRestoreGitRepository(t, map[string]string{
		"cerne.json":       `{"name":"example","source":"../old","custom":true}`,
		"product/.gitkeep": "", "specs/.gitkeep": "", "decisions/.gitkeep": "",
		"policies/.gitkeep": "", "runs/.gitkeep": "",
	})
	source := createRestoreGitRepository(t, map[string]string{"README.md": "untouched\n"})
	before := snapshotTreeWithoutGit(t, source)
	environment := replaceEnvironment(os.Environ(), "HOME", home)
	environment = replaceEnvironment(environment, "USERPROFILE", home)
	status, stdout, stderr := executeCLI(t, binary, parent, environment, "restore", knowledge, "--source", source)
	if status != 0 || stderr != "" || !strings.Contains(stdout, "Source vinculado: "+displayPath(source)+"\n") ||
		!strings.HasSuffix(stdout, "Manifesto: referência de source atualizada.\n") {
		t.Fatalf("status = %d\nstdout = %q\nstderr = %q", status, stdout, stderr)
	}
	if after := snapshotTreeWithoutGit(t, source); !reflect.DeepEqual(before, after) {
		t.Fatal("source local foi alterado")
	}
	manifest := readFile(t, filepath.Join(parent, "example", "knowledge", "cerne.json"))
	if !strings.Contains(manifest, `"custom": true`) {
		t.Fatalf("manifesto perdeu campo: %s", manifest)
	}
}

func TestCLIRestoreRejectsInvalidOriginBeforeAudit(t *testing.T) {
	binary := buildCLI(t)
	parent, home := t.TempDir(), t.TempDir()
	environment := replaceEnvironment(os.Environ(), "HOME", home)
	environment = replaceEnvironment(environment, "USERPROFILE", home)
	secret := "https://user:password@example.invalid/knowledge.git"
	status, stdout, stderr := executeCLI(t, binary, parent, environment, "restore", secret, "--clone", "source")
	if status != 2 || stdout != "" || strings.Contains(stderr, secret) || strings.Contains(stderr, "password") {
		t.Fatalf("status = %d\nstdout = %q\nstderr = %q", status, stdout, stderr)
	}
	if _, err := os.Stat(filepath.Join(home, ".cerne")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("uso inválido criou audit: %v", err)
	}
}

func createRestoreGitRepository(t *testing.T, files map[string]string) string {
	t.Helper()
	repository := t.TempDir()
	gitOutput(t, repository, "init", "--quiet")
	gitOutput(t, repository, "config", "user.email", "restore@example.com")
	gitOutput(t, repository, "config", "user.name", "Restore Test")
	for name, content := range files {
		path := filepath.Join(repository, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	gitOutput(t, repository, "add", ".")
	gitOutput(t, repository, "commit", "--quiet", "-m", "fixture")
	return repository
}
