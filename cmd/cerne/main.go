package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/WilliamSampaio/cerne-cli/internal/filecheck"
	"github.com/WilliamSampaio/cerne-cli/internal/gitexec"
	"github.com/WilliamSampaio/cerne-cli/internal/workspace"
)

const initHelp = `Inicializa um workspace Cerne com repositórios Git independentes.

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

const doctorHelp = `Analisa o workspace Cerne atual sem modificar arquivos ou repositórios.

Uso:
  cerne doctor
  cerne doctor --help

Raiz:
  O diretório atual é tratado como raiz do workspace.

Verificações:
  Manifesto; repositório de conhecimento; repositório de código-fonte;
  independência Git; isolamento de versionamento; caminhos do manifesto;
  diretórios obrigatórios; Git; permissões; versão do manifesto.

Símbolos e resumos:
  ✓ aprovado
  ✗ erro bloqueante
  ! aviso não bloqueante

  Workspace saudável
  Workspace com avisos
  Workspace inválido

Saídas:
  Relatórios, resumos e ajuda usam stdout. Uso inválido ou falha antes do
  relatório usa stderr. Status 0: saudável, avisos ou ajuda; 1: erro
  bloqueante ou falha operacional; 2: uso inválido.

Efeitos:
  Leitura exclusiva. Não cria, corrige, altera, remove, usa rede, Git remoto,
  credenciais ou agentes de IA. Permissões inconclusivas aparecem como aviso.

Manifesto:
  A ausência de version significa versão 1 implícita. Quando version existe,
  somente o inteiro JSON 1 é aceito. name válido diferente da raiz gera aviso.

Exemplo:
  cerne doctor
`

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		return commandUsageError(stderr, "informe um comando")
	}
	switch args[0] {
	case "init":
		return runInit(args[1:], stdout, stderr)
	case "doctor":
		return runDoctor(args[1:], stdout, stderr)
	default:
		return commandUsageError(stderr, "comando desconhecido")
	}
}

func runInit(args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		return initUsageError(stderr, "informe exatamente um nome de projeto")
	}
	if args[0] == "--help" {
		fmt.Fprint(stdout, initHelp)
		return 0
	}
	if err := workspace.ValidateName(args[0]); err != nil {
		return initUsageError(stderr, err.Error())
	}

	initRepository, err := gitexec.Find()
	if err != nil {
		fmt.Fprintf(stderr, "erro: %v\n", err)
		fmt.Fprintln(stderr, "correção: instale o Git e disponibilize-o no PATH")
		return 1
	}
	current, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "erro: não foi possível obter o diretório atual: %v\n", err)
		fmt.Fprintln(stderr, "correção: execute o comando em um diretório acessível")
		return 1
	}
	result, err := workspace.Init(current, args[0], initRepository)
	if err != nil {
		fmt.Fprintf(stderr, "erro: %v\n", err)
		if errors.Is(err, workspace.ErrUnsafeDestination) {
			fmt.Fprintln(stderr, "correção: escolha um destino inexistente ou vazio")
		} else {
			fmt.Fprintln(stderr, "correção: verifique permissões e tente novamente")
		}
		return 1
	}

	fmt.Fprintf(stdout, "Workspace %q criado.\nKnowledge: %s\nSource: %s\n",
		result.Name, result.KnowledgePath, result.SourcePath)
	return 0
}

func runDoctor(args []string, stdout, stderr io.Writer) int {
	if len(args) == 1 && args[0] == "--help" {
		fmt.Fprint(stdout, doctorHelp)
		return 0
	}
	if len(args) != 0 {
		fmt.Fprintf(stderr, "erro: argumento inválido\n")
		fmt.Fprintln(stderr, "uso: cerne doctor")
		return 2
	}

	current, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "erro: não foi possível obter o diretório atual: %v\n", err)
		fmt.Fprintln(stderr, "correção: execute o comando em um diretório acessível")
		return 1
	}
	inspect, err := gitexec.FindInspector()
	if err != nil {
		inspect = nil
	}
	diagnosis := workspace.Doctor(current, adaptGit(inspect), adaptAccess)
	renderDiagnosis(stdout, diagnosis)
	if diagnosis.Status == workspace.Invalid {
		return 1
	}
	return 0
}

func adaptGit(inspect func(string) (gitexec.Repository, error)) workspace.GitInspect {
	if inspect == nil {
		return nil
	}
	return func(path string) (workspace.RepositoryFacts, error) {
		result, err := inspect(path)
		return workspace.RepositoryFacts{
			RequestedRoot: result.RequestedRoot,
			WorktreeRoot:  result.WorktreeRoot,
			CommonDir:     result.CommonDir,
		}, err
	}
}

func adaptAccess(path string) workspace.AccessResult {
	result := filecheck.Access(path)
	return workspace.AccessResult{
		Path:   result.Path,
		Read:   workspace.AccessOutcome(result.Read),
		Write:  workspace.AccessOutcome(result.Write),
		Reason: result.Reason,
	}
}

func renderDiagnosis(stdout io.Writer, diagnosis workspace.Diagnosis) {
	for _, check := range diagnosis.Checks {
		fmt.Fprintf(stdout, "%s %s: %s", symbol(check.Severity), check.Label, check.Detail)
		if check.Correction != "" {
			fmt.Fprintf(stdout, "; correção: %s", check.Correction)
		}
		fmt.Fprintln(stdout)
	}
	switch diagnosis.Status {
	case workspace.Invalid:
		fmt.Fprintln(stdout, "Workspace inválido")
	case workspace.Warnings:
		fmt.Fprintln(stdout, "Workspace com avisos")
	default:
		fmt.Fprintln(stdout, "Workspace saudável")
	}
}

func symbol(severity workspace.Severity) string {
	switch severity {
	case workspace.Error:
		return "✗"
	case workspace.Warning:
		return "!"
	default:
		return "✓"
	}
}

func initUsageError(stderr io.Writer, cause string) int {
	fmt.Fprintf(stderr, "erro: %s\n", cause)
	fmt.Fprintln(stderr, "uso: cerne init <project-name>")
	return 2
}

func commandUsageError(stderr io.Writer, cause string) int {
	fmt.Fprintf(stderr, "erro: %s\n", cause)
	fmt.Fprintln(stderr, "uso: cerne <init|doctor>")
	return 2
}
