package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/WilliamSampaio/cerne-cli/internal/filecheck"
	"github.com/WilliamSampaio/cerne-cli/internal/gitexec"
	"github.com/WilliamSampaio/cerne-cli/internal/workflowexec"
	"github.com/WilliamSampaio/cerne-cli/internal/workspace"
)

const version = "0.6.0"

const globalHelp = `Cerne administra workspaces com repositórios Git independentes de conhecimento e código-fonte.

Uso:
  cerne <comando> [argumentos]
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

Opções:
  --help       Exibe esta ajuda
  --version    Exibe a versão do Cerne

Use "cerne <comando> --help" para detalhes de um comando.
`

const contextHelp = `Exibe o contexto estrutural comprovado do workspace Cerne.

Uso:
  cerne context
  cerne context --json
  cerne context --help

Localização:
  Procura o workspace ancestral mais próximo. Um source externo não realiza
  descoberta reversa do workspace.

Relatório:
  Informa workspace, knowledge, coleções, source e workflow sem ler conteúdo.
  --json usa schema_version 1 estável para automação.

Saídas:
  Relatórios e ajuda usam stdout. Uso inválido usa stderr.
  Status 0: saudável, avisos ou ajuda; 1: contexto inválido; 2: uso inválido.

Efeitos:
  Leitura estrutural exclusiva. Não executa Git, provider ou agente; não acessa
  rede, ambiente ou credenciais; não cria auditoria, cache ou instruções.

Exemplos:
  cerne context
  cerne context --json
`

const restoreHelp = `Restaura um workspace Cerne a partir de um repositório knowledge.

Uso:
  cerne restore <origem-knowledge> --clone <origem-source>
  cerne restore <origem-knowledge> --source <caminho-local>
  cerne restore --help

Comportamento:
  Knowledge é sempre clonado. --clone também clona source; --source vincula a
  raiz de um working tree Git local sem copiá-lo nem alterá-lo. O nome e o
  caminho portátil do source são lidos de knowledge/cerne.json.

Segurança:
  O destino deve estar ausente e nunca é substituído. A restauração usa staging
  privado, valida repositórios independentes e não executa setup de workflow.
  Cada tentativa válida cria um registro redigido em ~/.cerne/audit antes do Git.

Saídas:
  Sucesso e ajuda usam stdout. Falhas usam stderr.
  Status 0: sucesso ou ajuda; 1: falha operacional; 2: uso ou origem inválida.

Efeitos:
  Não faz push, merge, fetch adicional, instalação, deploy ou execução de
  provider. Autenticação e redirects seguem a configuração externa do Git.

Exemplos:
  cerne restore git@host:org/knowledge.git --clone git@host:org/source.git
  cerne restore ../knowledge.git --source ../source-existente
`

const initHelp = `Inicializa um workspace Cerne com repositórios Git independentes.

Uso:
  cerne init <project-name>
  cerne init <project-name> --source <caminho>
  cerne init <project-name> --clone <origem>
  cerne init <project-name> --workflow <speckit|openspec>
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
  cerne init exemplo --clone https://host/organizacao/aplicacao.git --workflow speckit
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
	case "restore":
		return runRestore(args[1:], stdout, stderr)
	case "doctor":
		return runDoctor(args[1:], stdout, stderr)
	case "status":
		return runStatus(args[1:], stdout, stderr)
	case "link":
		return runLink(args[1:], stdout, stderr)
	case "workflow":
		return runWorkflow(args[1:], stdout, stderr)
	case "context":
		return runContext(args[1:], stdout, stderr)
	default:
		return commandUsageError(stderr, "comando desconhecido")
	}
}

func runContext(args []string, stdout, stderr io.Writer) int {
	if len(args) == 1 && args[0] == "--help" {
		fmt.Fprint(stdout, contextHelp)
		return 0
	}
	jsonOutput := len(args) == 1 && args[0] == "--json"
	if len(args) != 0 && !jsonOutput {
		fmt.Fprintln(stderr, "erro: argumento inválido")
		fmt.Fprintln(stderr, "uso: cerne context [--json]")
		return 2
	}
	current, _ := currentDirectory()
	report := workspace.Context(current, workflowexec.Describe)
	if jsonOutput {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(report); err != nil {
			return 1
		}
	} else {
		renderContext(stdout, report)
	}
	if report.Status == workspace.Invalid {
		return 1
	}
	return 0
}

func renderContext(stdout io.Writer, report workspace.ContextReport) {
	if report.Workspace != nil && report.Workspace.Name != "" {
		fmt.Fprintf(stdout, "Workspace: %s\n", report.Workspace.Name)
	}
	fmt.Fprintf(stdout, "Status: %s\n", contextStatus(report.Status))
	if report.Workspace != nil {
		fmt.Fprintf(stdout, "Root: %s\n", report.Workspace.Root)
	}
	if report.Knowledge != nil {
		fmt.Fprintf(stdout, "\nKnowledge: %s\n", report.Knowledge.Path)
		contextPathLine(stdout, "Product", report.Knowledge.ProductPath)
		contextPathLine(stdout, "Specs", report.Knowledge.SpecsPath)
		contextPathLine(stdout, "Decisions", report.Knowledge.DecisionsPath)
		contextPathLine(stdout, "Policies", report.Knowledge.PoliciesPath)
	}
	if report.Source != nil {
		fmt.Fprintf(stdout, "\nSource: %s\n", report.Source.Path)
		location := "externo ao workspace"
		if report.Source.InsideWorkspace {
			location = "interno ao workspace"
		}
		fmt.Fprintf(stdout, "Localização: %s\n", location)
	}
	if report.Workflow != nil {
		fmt.Fprintln(stdout)
		if !report.Workflow.Declared {
			fmt.Fprintln(stdout, "Workflow: não declarado")
		} else {
			fmt.Fprintf(stdout, "Workflow: %s (%s)\n", report.Workflow.Provider, contextWorkflowState(report.Workflow.State))
		}
	}
	if len(report.Problems) > 0 {
		fmt.Fprintln(stdout)
	}
	for _, problem := range report.Problems {
		symbol := "✗"
		if problem.Severity == "warning" {
			symbol = "!"
		}
		fmt.Fprintf(stdout, "%s %s: %s\n", symbol, contextComponent(problem.Component), contextProblemDetail(problem.Code))
		fmt.Fprintf(stdout, "  Correção: %s\n", contextProblemCorrection(problem.Code))
	}
}

func contextPathLine(output io.Writer, label, path string) {
	if path != "" {
		fmt.Fprintf(output, "%s: %s\n", label, path)
	}
}

func contextStatus(status workspace.Status) string {
	switch status {
	case workspace.Invalid:
		return "inválido"
	case workspace.Warnings:
		return "aviso"
	default:
		return "saudável"
	}
}

func contextWorkflowState(state workspace.ContextWorkflowState) string {
	switch state {
	case workspace.ContextWorkflowPending:
		return "pendente"
	case workspace.ContextWorkflowReady:
		return "pronto"
	case workspace.ContextWorkflowInvalid:
		return "inválido"
	default:
		return "provider desconhecido"
	}
}

func contextComponent(component string) string {
	switch component {
	case "workspace":
		return "Workspace"
	case "manifest":
		return "Manifesto"
	case "knowledge":
		return "Knowledge"
	case "source":
		return "Source"
	case "workflow":
		return "Workflow"
	default:
		return component
	}
}

func contextProblemDetail(code string) string {
	details := map[string]string{"workspace-not-found": "não localizado", "manifest-invalid": "ausente ou inválido", "manifest-version-unsupported": "versão não suportada", "knowledge-invalid": "ausente ou inseguro", "source-invalid": "ausente ou inseguro", "required-directory-invalid": "diretório obrigatório ausente ou inválido", "workflow-pending": "pendente", "workflow-invalid": "estrutura inválida ou parcial", "workflow-unknown-provider": "provider não suportado"}
	return details[code]
}

func contextProblemCorrection(code string) string {
	corrections := map[string]string{"workspace-not-found": "execute o comando dentro de um workspace Cerne", "manifest-invalid": "corrija ou restaure knowledge/cerne.json", "manifest-version-unsupported": "use uma versão compatível do Cerne", "knowledge-invalid": "restaure o diretório knowledge", "source-invalid": "corrija o caminho source no manifesto", "required-directory-invalid": "restaure o diretório indicado pelo componente", "workflow-pending": "execute cerne workflow setup quando o provider estiver disponível", "workflow-invalid": "corrija a estrutura antes de continuar", "workflow-unknown-provider": "use speckit ou openspec no manifesto"}
	return corrections[code]
}

type restoreArguments struct {
	KnowledgeOrigin string
	SourceMode      workspace.SourceMode
	SourceValue     string
}

func parseRestoreArgs(args []string) (restoreArguments, bool) {
	if len(args) != 3 || strings.HasPrefix(args[0], "--") || args[0] == "" || args[2] == "" || strings.HasPrefix(args[2], "--") {
		return restoreArguments{}, false
	}
	parsed := restoreArguments{KnowledgeOrigin: args[0], SourceValue: args[2]}
	switch args[1] {
	case "--clone":
		parsed.SourceMode = workspace.SourceClone
	case "--source":
		parsed.SourceMode = workspace.SourceLocal
	default:
		return restoreArguments{}, false
	}
	return parsed, true
}

func runRestore(args []string, stdout, stderr io.Writer) int {
	if len(args) == 1 && args[0] == "--help" {
		fmt.Fprint(stdout, restoreHelp)
		return 0
	}
	parsed, ok := parseRestoreArgs(args)
	if !ok {
		fmt.Fprintln(stderr, "erro: argumento inválido")
		fmt.Fprintln(stderr, "uso: cerne restore <origem-knowledge> (--source <caminho> | --clone <origem-source>)")
		return 2
	}
	current, err := currentDirectory()
	if err != nil {
		fmt.Fprintln(stderr, "erro: não foi possível obter o diretório atual")
		fmt.Fprintln(stderr, "correção: execute o comando em um diretório acessível")
		return 1
	}
	knowledgeOrigin, err := gitexec.ClassifyCloneOrigin(current, parsed.KnowledgeOrigin)
	if err != nil {
		return restoreUsageError(stderr, "origem do knowledge inválida")
	}
	sourceInput := parsed.SourceValue
	if parsed.SourceMode == workspace.SourceClone {
		sourceOrigin, err := gitexec.ClassifyCloneOrigin(current, parsed.SourceValue)
		if err != nil {
			return restoreUsageError(stderr, "origem de clone do source inválida")
		}
		sourceInput = sourceOrigin.Location
	}
	clone, err := gitexec.FindClone()
	if err != nil {
		fmt.Fprintln(stderr, "erro: Git indisponível")
		fmt.Fprintln(stderr, "correção: instale o Git e disponibilize-o no PATH")
		return 1
	}
	inspect, err := gitexec.FindLinkInspector()
	if err != nil {
		fmt.Fprintln(stderr, "erro: Git indisponível")
		fmt.Fprintln(stderr, "correção: instale o Git e disponibilize-o no PATH")
		return 1
	}
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintln(stderr, "erro: não foi possível localizar o diretório pessoal")
		fmt.Fprintln(stderr, "correção: configure um diretório pessoal acessível")
		return 1
	}
	result, err := workspace.Restore(current, home, workspace.RestoreRequest{
		KnowledgeOrigin: knowledgeOrigin.Location, SourceMode: parsed.SourceMode, SourceInput: sourceInput,
	}, adaptLink(inspect), clone, workflowexec.Resolve)
	if err != nil {
		var failure workspace.RestoreFailure
		if errors.As(err, &failure) {
			fmt.Fprintf(stderr, "erro: %s\ncorreção: %s\n", failure.Cause, failure.Correction)
		} else {
			fmt.Fprintln(stderr, "erro: não foi possível restaurar o workspace")
			fmt.Fprintln(stderr, "correção: verifique a auditoria e tente novamente")
		}
		return 1
	}
	fmt.Fprintf(stdout, "Workspace %q restaurado.\nKnowledge: %s\n", result.Name, result.KnowledgePath)
	if result.SourceMode == workspace.SourceClone {
		fmt.Fprintf(stdout, "Source clonado: %s\n", result.SourcePath)
	} else {
		fmt.Fprintf(stdout, "Source vinculado: %s\n", result.SourcePath)
		if result.ManifestChanged {
			fmt.Fprintln(stdout, "Manifesto: referência de source atualizada.")
		}
	}
	return 0
}

func restoreUsageError(stderr io.Writer, cause string) int {
	fmt.Fprintf(stderr, "erro: %s\n", cause)
	fmt.Fprintln(stderr, "uso: cerne restore <origem-knowledge> (--source <caminho> | --clone <origem-source>)")
	return 2
}

func runInit(args []string, stdout, stderr io.Writer) int {
	if len(args) == 1 && args[0] == "--help" {
		fmt.Fprint(stdout, initHelp)
		return 0
	}
	parsed, ok := parseInitArgs(args)
	if !ok {
		return initUsageError(stderr, "argumento inválido")
	}
	if err := workspace.ValidateName(parsed.Name); err != nil {
		return initUsageError(stderr, err.Error())
	}
	var definition workspace.WorkflowDefinition
	if parsed.Workflow != "" {
		var err error
		definition, err = workflowexec.Resolve(parsed.Workflow)
		if err != nil {
			return initUsageError(stderr, "workflow inválido: use speckit ou openspec")
		}
	}
	var current string
	var origin gitexec.CloneOrigin
	var err error
	if parsed.SourceMode == workspace.SourceClone {
		current, err = currentDirectory()
		if err != nil {
			fmt.Fprintln(stderr, "erro: não foi possível obter o diretório atual")
			fmt.Fprintln(stderr, "correção: execute o comando em um diretório acessível")
			return 1
		}
		origin, err = gitexec.ClassifyCloneOrigin(current, parsed.SourceValue)
		if err != nil {
			return initUsageError(stderr, "origem de clone inválida")
		}
	}

	initRepository, err := gitexec.Find()
	if err != nil {
		fmt.Fprintf(stderr, "erro: %v\n", err)
		fmt.Fprintln(stderr, "correção: instale o Git e disponibilize-o no PATH")
		return 1
	}
	if current == "" {
		current, err = currentDirectory()
		if err != nil {
			fmt.Fprintf(stderr, "erro: não foi possível obter o diretório atual: %v\n", err)
			fmt.Fprintln(stderr, "correção: execute o comando em um diretório acessível")
			return 1
		}
	}
	if parsed.SourceMode != "" {
		inspect, inspectErr := gitexec.FindLinkInspector()
		if inspectErr != nil {
			fmt.Fprintf(stderr, "erro: %v\n", inspectErr)
			fmt.Fprintln(stderr, "correção: instale o Git e disponibilize-o no PATH")
			return 1
		}
		request := workspace.SourceInitRequest{Mode: parsed.SourceMode, Input: parsed.SourceValue}
		var clone workspace.CloneSource
		if parsed.SourceMode == workspace.SourceClone {
			clone, err = gitexec.FindClone()
			if err != nil {
				fmt.Fprintf(stderr, "erro: %v\n", err)
				fmt.Fprintln(stderr, "correção: instale o Git e disponibilize-o no PATH")
				return 1
			}
			request.Input = origin.Location
			request.OriginTransport = origin.Transport
			request.OriginFingerprint = origin.Fingerprint
		}
		result, workflow, err := workspace.InitWithSourceAndWorkflow(current, parsed.Name, request, definition, initRepository, adaptLink(inspect), clone)
		if err != nil {
			var workflowFailure workspace.WorkflowFailure
			if errors.As(err, &workflowFailure) {
				fmt.Fprintf(stderr, "erro: não foi possível inicializar workflow %s: %s\n", parsed.Workflow, workflowCause(err))
				fmt.Fprintf(stderr, "correção: corrija ou atualize %s e execute \"cerne workflow setup\" dentro de %s\n", parsed.Workflow, filepath.Dir(result.KnowledgePath))
				return 1
			}
			renderSourceInitFailure(stderr, err)
			return 1
		}
		fmt.Fprintf(stdout, "Workspace %q criado.\nKnowledge: %s\n", result.Name, result.KnowledgePath)
		if result.SourceMode == workspace.SourceLocal {
			fmt.Fprintf(stdout, "Source vinculado: %s\n", result.SourcePath)
		} else {
			fmt.Fprintf(stdout, "Source clonado: %s\n", result.SourcePath)
		}
		if parsed.Workflow != "" {
			fmt.Fprintf(stdout, "Workflow: %s\nSetup: %s\n", parsed.Workflow, map[workspace.WorkflowState]string{workspace.WorkflowConfigured: "concluído", workspace.WorkflowPending: "pendente"}[workflow.State])
			if workflow.State == workspace.WorkflowPending {
				fmt.Fprintf(stderr, "aviso: executável %q não encontrado; workflow %s não inicializado\n", definition.Executor, parsed.Workflow)
				fmt.Fprintf(stderr, "correção: instale %s e execute \"cerne workflow setup\" dentro do workspace\n", parsed.Workflow)
			}
		}
		return 0
	}
	if parsed.Workflow != "" {
		result, workflow, err := workspace.InitWithWorkflow(current, parsed.Name, definition, initRepository)
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
			fmt.Fprintf(stderr, "erro: não foi possível inicializar workflow %s: %s\n", parsed.Workflow, workflowCause(err))
			fmt.Fprintf(stderr, "correção: corrija ou atualize %s e execute \"cerne workflow setup\" dentro de %s\n", parsed.Workflow, filepath.Dir(result.KnowledgePath))
			return 1
		}
		fmt.Fprintf(stdout, "Workspace %q criado.\nKnowledge: %s\nSource: %s\nWorkflow: %s\nSetup: %s\n",
			result.Name, result.KnowledgePath, result.SourcePath, parsed.Workflow, map[workspace.WorkflowState]string{workspace.WorkflowConfigured: "concluído", workspace.WorkflowPending: "pendente"}[workflow.State])
		if workflow.State == workspace.WorkflowPending {
			fmt.Fprintf(stderr, "aviso: executável %q não encontrado; workflow %s não inicializado\n", definition.Executor, parsed.Workflow)
			fmt.Fprintf(stderr, "correção: instale %s e execute \"cerne workflow setup\" dentro do workspace\n", parsed.Workflow)
		}
		return 0
	}

	result, err := workspace.Init(current, parsed.Name, initRepository)
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

type initArguments struct {
	Name        string
	Workflow    string
	SourceMode  workspace.SourceMode
	SourceValue string
}

func parseInitArgs(args []string) (initArguments, bool) {
	if len(args) == 1 && !strings.HasPrefix(args[0], "--") {
		return initArguments{Name: args[0]}, true
	}
	if (len(args) != 3 && len(args) != 5) || strings.HasPrefix(args[0], "--") {
		return initArguments{}, false
	}
	parsed := initArguments{Name: args[0]}
	for index := 1; index < len(args); index += 2 {
		value := args[index+1]
		if value == "" || strings.HasPrefix(value, "--") {
			return initArguments{}, false
		}
		switch args[index] {
		case "--workflow":
			if parsed.Workflow != "" || value != "speckit" && value != "openspec" {
				return initArguments{}, false
			}
			parsed.Workflow = value
		case "--source":
			if parsed.SourceMode != "" {
				return initArguments{}, false
			}
			parsed.SourceMode, parsed.SourceValue = workspace.SourceLocal, value
		case "--clone":
			if parsed.SourceMode != "" {
				return initArguments{}, false
			}
			parsed.SourceMode, parsed.SourceValue = workspace.SourceClone, value
		default:
			return initArguments{}, false
		}
	}
	return parsed, true
}

func renderSourceInitFailure(stderr io.Writer, err error) {
	var cloneFailure workspace.SourceInitFailure
	if errors.As(err, &cloneFailure) {
		fmt.Fprintf(stderr, "erro: %s\n", cloneFailure.Cause)
		fmt.Fprintf(stderr, "correção: %s\n", cloneFailure.Correction)
		return
	}
	var linkFailure workspace.LinkFailure
	if errors.As(err, &linkFailure) {
		if linkFailure.Path == "" {
			fmt.Fprintf(stderr, "erro: %s\n", linkFailure.Cause)
		} else {
			fmt.Fprintf(stderr, "erro: %s: %s\n", linkFailure.Cause, linkFailure.Path)
		}
		fmt.Fprintf(stderr, "correção: %s\n", linkFailure.Correction)
		return
	}
	fmt.Fprintf(stderr, "erro: %v\n", err)
	if errors.Is(err, workspace.ErrUnsafeDestination) {
		fmt.Fprintln(stderr, "correção: escolha um destino inexistente ou vazio")
	} else {
		fmt.Fprintln(stderr, "correção: verifique permissões e tente novamente")
	}
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
	fmt.Fprintln(stderr, "uso: cerne init <project-name> [--source <caminho> | --clone <origem>] [--workflow <speckit|openspec>]")
	return 2
}

func commandUsageError(stderr io.Writer, cause string) int {
	fmt.Fprintf(stderr, "erro: %s\n", cause)
	fmt.Fprintln(stderr, "uso: cerne <init|doctor|status|link|workflow>")
	return 2
}
