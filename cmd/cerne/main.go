package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/WilliamSampaio/cerne-cli/internal/filecheck"
	"github.com/WilliamSampaio/cerne-cli/internal/gitexec"
	"github.com/WilliamSampaio/cerne-cli/internal/workflowexec"
	"github.com/WilliamSampaio/cerne-cli/internal/workspace"
)

const version = "0.2.0"

const globalHelp = `Cerne administra workspaces com repositórios Git independentes de conhecimento e código-fonte.

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

const initHelp = `Inicializa um workspace Cerne com repositórios Git independentes.

Uso:
  cerne init <project-name> [--workflow <speckit|openspec>]

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

Workflow:
  Sem a opção, mantém o layout padrão em knowledge/specs. Spec Kit também usa
  specs e cria .specify. OpenSpec usa openspec/specs e cria openspec.
  O Cerne usa somente uma instalação local existente, sem instalar, atualizar,
  selecionar agente ou fornecer credenciais. Se ausente, o setup fica pendente.

Efeitos:
  Cria dois repositórios Git locais vazios, sem commit ou remoto.
  Com --workflow, executa o init local do provider somente em knowledge e cria
  uma auditoria em runs. Não altera source nem autoriza operações Git remotas.

Saídas:
  Sucesso e ajuda usam stdout. Erros usam stderr.
  Status 0: sucesso ou ajuda; 1: falha operacional; 2: uso ou nome inválido.

Erros:
  O destino deve estar ausente ou ser um diretório regular vazio.
  Instale Git, corrija o nome ou escolha outro destino conforme o diagnóstico.

Exemplo:
  cerne init exemplo
  cerne init exemplo --workflow speckit
`

const workflowHelp = `Inicializa o workflow já declarado no manifesto do workspace.

Uso:
  cerne workflow setup
  cerne workflow --help

Localização:
  Procura o workspace ancestral mais próximo por knowledge/cerne.json.

Comportamento:
  Executa somente o provider declarado e já instalado dentro de knowledge.
  Um layout pronto não é alterado. Um layout parcial é recusado.

Saídas:
  Sucesso e ajuda usam stdout. Falhas usam stderr.
  Status 0: concluído, já pronto ou ajuda; 1: falha operacional; 2: uso inválido.

Efeitos:
  Registra tentativas em knowledge/runs. Não instala ferramentas, não troca o
  provider, não altera source e não autoriza rede, Git remoto ou agentes.

Exemplo:
  cerne workflow setup
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
  Quando declarado: provider, materialização e disponibilidade do workflow.

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

const statusHelp = `Apresenta o estado atual de um workspace Cerne sem modificar arquivos.

Uso:
  cerne status
  cerne status --help

Localização:
  Procura, a partir do diretório atual, o workspace ancestral mais próximo
  identificado por knowledge/cerne.json.

Relatório:
  Exibe projeto, caminho absoluto do workspace e, para knowledge e source,
  caminho, branch, commit, estado e contagens de modificados, stage e não
  rastreados.

Estados especiais:
  Branch: detached HEAD quando não há branch simbólica.
  Commit: sem commits quando o repositório ainda não possui commit.
  Estado: limpo sem alterações; alterações pendentes com qualquer contagem.

Saídas:
  Relatório e ajuda usam stdout. Uso inválido e falhas usam stderr.
  Status 0: consulta obtida ou ajuda; 1: falha operacional; 2: uso inválido.

Efeitos:
  Leitura exclusiva. Não cria, corrige, altera stage, faz commit, checkout,
  reset, fetch, pull, acessa remotos, credenciais ou agentes de IA.

Exemplo:
  cerne status
`

const linkHelp = `Vincula o workspace Cerne atual a um repositório Git local existente como source.

Uso:
  cerne link <caminho>
  cerne link <caminho> --replace
  cerne link --help

Caminho:
  Pode ser relativo ao diretório atual ou absoluto. Deve apontar para a raiz
  de um repositório Git local com árvore de trabalho. Worktrees válidos são
  aceitos; repositórios bare não são aceitos.

Substituição:
  Se outro source já estiver configurado, a troca exige --replace. Mesmo com
  --replace, o Cerne atualiza somente knowledge/cerne.json e não altera o
  source anterior nem o novo.

Validações:
  Workspace, manifesto, versão, caminho informado, Git local, repositório
  non-bare, independência entre knowledge/source e sobreposição perigosa.

Saídas:
  Sucesso e ajuda usam stdout. Uso inválido e falhas usam stderr.
  Status 0: atualizado, nenhuma alteração ou ajuda; 1: falha operacional;
  2: uso inválido.

Efeitos:
  Lê o workspace e metadados Git locais. Não copia, move, apaga, faz checkout,
  reset, add, commit, clean, fetch, pull, push, acessa remotos, credenciais ou
  agentes de IA.

Exemplo:
  cerne link ../aplicacao-existente --replace
`

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

var currentDirectory = os.Getwd

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 1 && args[0] == "--help" {
		fmt.Fprint(stdout, globalHelp)
		return 0
	}
	if len(args) == 1 && args[0] == "--version" {
		fmt.Fprintf(stdout, "cerne %s\n", version)
		return 0
	}
	if len(args) == 0 {
		return commandUsageError(stderr, "informe um comando")
	}
	switch args[0] {
	case "init":
		return runInit(args[1:], stdout, stderr)
	case "doctor":
		return runDoctor(args[1:], stdout, stderr)
	case "status":
		return runStatus(args[1:], stdout, stderr)
	case "link":
		return runLink(args[1:], stdout, stderr)
	case "workflow":
		return runWorkflow(args[1:], stdout, stderr)
	default:
		return commandUsageError(stderr, "comando desconhecido")
	}
}

func runInit(args []string, stdout, stderr io.Writer) int {
	if len(args) == 1 && args[0] == "--help" {
		fmt.Fprint(stdout, initHelp)
		return 0
	}
	name, provider, ok := parseInitArgs(args)
	if !ok {
		return initUsageError(stderr, "argumento inválido")
	}
	if err := workspace.ValidateName(name); err != nil {
		return initUsageError(stderr, err.Error())
	}
	var definition workspace.WorkflowDefinition
	if provider != "" {
		var err error
		definition, err = workflowexec.Resolve(provider)
		if err != nil {
			return initUsageError(stderr, "workflow inválido: use speckit ou openspec")
		}
	}

	initRepository, err := gitexec.Find()
	if err != nil {
		fmt.Fprintf(stderr, "erro: %v\n", err)
		fmt.Fprintln(stderr, "correção: instale o Git e disponibilize-o no PATH")
		return 1
	}
	current, err := currentDirectory()
	if err != nil {
		fmt.Fprintf(stderr, "erro: não foi possível obter o diretório atual: %v\n", err)
		fmt.Fprintln(stderr, "correção: execute o comando em um diretório acessível")
		return 1
	}
	if provider != "" {
		result, workflow, err := workspace.InitWithWorkflow(current, name, definition, initRepository)
		if err != nil {
			if result.KnowledgePath == "" {
				fmt.Fprintf(stderr, "erro: %v\n", err)
				if errors.Is(err, workspace.ErrUnsafeDestination) {
					fmt.Fprintln(stderr, "correção: escolha um destino inexistente ou vazio")
				} else {
					fmt.Fprintln(stderr, "correção: verifique permissões e tente novamente")
				}
				return 1
			}
			fmt.Fprintf(stderr, "erro: não foi possível inicializar workflow %s: %s\n", provider, workflowCause(err))
			fmt.Fprintf(stderr, "correção: corrija ou atualize %s e execute \"cerne workflow setup\" dentro de %s\n", provider, filepath.Dir(result.KnowledgePath))
			return 1
		}
		fmt.Fprintf(stdout, "Workspace %q criado.\nKnowledge: %s\nSource: %s\nWorkflow: %s\nSetup: %s\n",
			result.Name, result.KnowledgePath, result.SourcePath, provider, map[workspace.WorkflowState]string{workspace.WorkflowConfigured: "concluído", workspace.WorkflowPending: "pendente"}[workflow.State])
		if workflow.State == workspace.WorkflowPending {
			fmt.Fprintf(stderr, "aviso: executável %q não encontrado; workflow %s não inicializado\n", definition.Executor, provider)
			fmt.Fprintf(stderr, "correção: instale %s e execute \"cerne workflow setup\" dentro do workspace\n", provider)
		}
		return 0
	}

	result, err := workspace.Init(current, name, initRepository)
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

func parseInitArgs(args []string) (string, string, bool) {
	if len(args) == 1 && args[0] != "--workflow" {
		return args[0], "", true
	}
	if len(args) == 3 && args[0] != "--workflow" && args[1] == "--workflow" && (args[2] == "speckit" || args[2] == "openspec") {
		return args[0], args[2], true
	}
	return "", "", false
}

func runWorkflow(args []string, stdout, stderr io.Writer) int {
	if len(args) == 1 && args[0] == "--help" {
		fmt.Fprint(stdout, workflowHelp)
		return 0
	}
	if len(args) != 1 || args[0] != "setup" {
		fmt.Fprintln(stderr, "erro: argumento inválido")
		fmt.Fprintln(stderr, "uso: cerne workflow setup")
		return 2
	}
	current, err := currentDirectory()
	if err != nil {
		fmt.Fprintln(stderr, "erro: não foi possível obter o diretório atual")
		fmt.Fprintln(stderr, "correção: execute o comando em um diretório acessível")
		return 1
	}
	result, err := workspace.SetupWorkflow(current, workflowexec.Resolve)
	if err != nil {
		var failure workspace.WorkflowFailure
		if errors.As(err, &failure) {
			fmt.Fprintf(stderr, "erro: %s\n", failure.Cause)
			fmt.Fprintf(stderr, "correção: %s\n", failure.Correction)
		} else {
			fmt.Fprintln(stderr, "erro: não foi possível inicializar o workflow")
			fmt.Fprintln(stderr, "correção: verifique o workspace e tente novamente")
		}
		return 1
	}
	if result.State == workspace.WorkflowPending {
		fmt.Fprintf(stderr, "erro: executável %q não encontrado\n", result.Executor)
		fmt.Fprintf(stderr, "correção: instale %s e execute novamente\n", result.Provider)
		return 1
	}
	fmt.Fprintf(stdout, "Workflow: %s\nKnowledge: %s\n", result.Provider, result.KnowledgePath)
	if result.State == workspace.WorkflowUnchanged {
		fmt.Fprintln(stdout, "Nenhuma alteração necessária.")
	} else {
		fmt.Fprintln(stdout, "Setup concluído.")
	}
	return 0
}

func workflowCause(err error) string {
	var failure workspace.WorkflowFailure
	if errors.As(err, &failure) {
		return failure.Cause
	}
	return "falha operacional"
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

	current, err := currentDirectory()
	if err != nil {
		fmt.Fprintf(stderr, "erro: não foi possível obter o diretório atual: %v\n", err)
		fmt.Fprintln(stderr, "correção: execute o comando em um diretório acessível")
		return 1
	}
	inspect, err := gitexec.FindInspector()
	if err != nil {
		inspect = nil
	}
	diagnosis := workspace.DoctorWithWorkflow(current, adaptGit(inspect), adaptAccess, workflowexec.Resolve)
	renderDiagnosis(stdout, diagnosis)
	if diagnosis.Status == workspace.Invalid {
		return 1
	}
	return 0
}

func runStatus(args []string, stdout, stderr io.Writer) int {
	if len(args) == 1 && args[0] == "--help" {
		fmt.Fprint(stdout, statusHelp)
		return 0
	}
	if len(args) != 0 {
		fmt.Fprintf(stderr, "erro: argumento inválido\n")
		fmt.Fprintln(stderr, "uso: cerne status")
		return 2
	}

	current, err := currentDirectory()
	if err != nil {
		fmt.Fprintf(stderr, "erro: não foi possível obter o diretório atual: %v\n", err)
		fmt.Fprintln(stderr, "correção: execute o comando em um diretório acessível")
		return 1
	}
	collect, err := gitexec.FindStatus()
	if err != nil {
		fmt.Fprintf(stderr, "erro: %v\n", err)
		fmt.Fprintln(stderr, "correção: instale o Git e disponibilize-o no PATH")
		return 1
	}
	report, err := workspace.CurrentStatus(current, adaptStatus(collect))
	if err != nil {
		var failure workspace.StatusFailure
		if errors.As(err, &failure) {
			if failure.Path != "" {
				fmt.Fprintf(stderr, "erro: %s: %s\n", failure.Cause, failure.Path)
			} else {
				fmt.Fprintf(stderr, "erro: %s\n", failure.Cause)
			}
			if failure.Correction != "" {
				fmt.Fprintf(stderr, "correção: %s\n", failure.Correction)
			}
		} else {
			fmt.Fprintf(stderr, "erro: %v\n", err)
			fmt.Fprintln(stderr, "correção: verifique o workspace e tente novamente")
		}
		return 1
	}
	renderStatus(stdout, report)
	return 0
}

func runLink(args []string, stdout, stderr io.Writer) int {
	if len(args) == 1 && args[0] == "--help" {
		fmt.Fprint(stdout, linkHelp)
		return 0
	}
	source, replace, ok := parseLinkArgs(args)
	if !ok {
		fmt.Fprintf(stderr, "erro: argumento inválido\n")
		fmt.Fprintln(stderr, "uso: cerne link <caminho> [--replace]")
		return 2
	}

	current, err := currentDirectory()
	if err != nil {
		fmt.Fprintf(stderr, "erro: não foi possível obter o diretório atual: %v\n", err)
		fmt.Fprintln(stderr, "correção: execute o comando em um diretório acessível")
		return 1
	}
	inspect, err := gitexec.FindLinkInspector()
	if err != nil {
		fmt.Fprintf(stderr, "erro: %v\n", err)
		fmt.Fprintln(stderr, "correção: instale o Git e disponibilize-o no PATH")
		return 1
	}
	result, err := workspace.Link(current, workspace.LinkRequest{SourceInput: source, Replace: replace}, adaptLink(inspect))
	if err != nil {
		var failure workspace.LinkFailure
		if errors.As(err, &failure) {
			if failure.Path != "" {
				fmt.Fprintf(stderr, "erro: %s: %s\n", failure.Cause, failure.Path)
			} else {
				fmt.Fprintf(stderr, "erro: %s\n", failure.Cause)
			}
			if failure.Correction != "" {
				fmt.Fprintf(stderr, "correção: %s\n", failure.Correction)
			}
		} else {
			fmt.Fprintf(stderr, "erro: %v\n", err)
			fmt.Fprintln(stderr, "correção: verifique o workspace e tente novamente")
		}
		return 1
	}
	renderLink(stdout, result)
	return 0
}

func parseLinkArgs(args []string) (string, bool, bool) {
	if len(args) == 1 && args[0] != "--replace" {
		return args[0], false, true
	}
	if len(args) == 2 && args[0] != "--replace" && args[1] == "--replace" {
		return args[0], true, true
	}
	return "", false, false
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

func adaptStatus(collect func(string) (gitexec.RepositoryStatus, error)) workspace.GitStatus {
	return func(path string) (workspace.GitRepositoryStatus, error) {
		result, err := collect(path)
		return workspace.GitRepositoryStatus{
			Path:           result.Path,
			Branch:         result.Branch,
			Commit:         result.Commit,
			ModifiedCount:  result.ModifiedCount,
			StagedCount:    result.StagedCount,
			UntrackedCount: result.UntrackedCount,
		}, err
	}
}

func adaptLink(inspect func(string) (gitexec.LinkRepository, error)) workspace.LinkGitInspect {
	return func(path string) (workspace.LinkRepositoryFacts, error) {
		result, err := inspect(path)
		return workspace.LinkRepositoryFacts{
			RequestedPath: result.RequestedPath,
			WorktreeRoot:  result.WorktreeRoot,
			CommonDir:     result.CommonDir,
			IsBare:        result.IsBare,
			HasWorktree:   result.HasWorktree,
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

func renderStatus(stdout io.Writer, report workspace.WorkspaceReport) {
	fmt.Fprintf(stdout, "Projeto: %s\n", report.ProjectName)
	fmt.Fprintf(stdout, "Workspace: %s\n\n", report.Root)
	for index, repository := range report.Repositories {
		if index > 0 {
			fmt.Fprintln(stdout)
		}
		fmt.Fprintf(stdout, "%s\n", repositoryTitle(repository.Name))
		fmt.Fprintf(stdout, "  Caminho: %s\n", repository.Path)
		fmt.Fprintf(stdout, "  Branch: %s\n", repository.Branch)
		fmt.Fprintf(stdout, "  Commit: %s\n", repository.Commit)
		fmt.Fprintf(stdout, "  Estado: %s\n", repository.State)
		fmt.Fprintf(stdout, "  Modificados: %d\n", repository.ModifiedCount)
		fmt.Fprintf(stdout, "  Em stage: %d\n", repository.StagedCount)
		fmt.Fprintf(stdout, "  Não rastreados: %d\n", repository.UntrackedCount)
	}
}

func renderLink(stdout io.Writer, result workspace.LinkResult) {
	fmt.Fprintf(stdout, "Projeto: %s\n", result.ProjectName)
	if !result.Changed {
		fmt.Fprintf(stdout, "Source atual: %s\n", result.NewSource)
		fmt.Fprintln(stdout, "Nenhuma alteração necessária.")
		return
	}
	if result.PreviousSource != "" {
		fmt.Fprintf(stdout, "Source anterior: %s\n", result.PreviousSource)
	}
	fmt.Fprintf(stdout, "Novo source: %s\n", result.NewSource)
	fmt.Fprintln(stdout, "Manifesto atualizado.")
}

func repositoryTitle(name string) string {
	if name == "source" {
		return "Source"
	}
	return "Knowledge"
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
	fmt.Fprintln(stderr, "uso: cerne init <project-name> [--workflow <speckit|openspec>]")
	return 2
}

func commandUsageError(stderr io.Writer, cause string) int {
	fmt.Fprintf(stderr, "erro: %s\n", cause)
	fmt.Fprintln(stderr, "uso: cerne <init|doctor|status|link|workflow>")
	return 2
}
