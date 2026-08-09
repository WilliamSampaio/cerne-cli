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
	"github.com/WilliamSampaio/cerne-cli/internal/localization"
	"github.com/WilliamSampaio/cerne-cli/internal/skillinstall"
	"github.com/WilliamSampaio/cerne-cli/internal/workflowexec"
	"github.com/WilliamSampaio/cerne-cli/internal/workspace"
)

var version = "0.8.1"

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

const skillHelp = `Instala skills oficiais do Cerne no perfil do agente.

Uso:
  cerne skill install <codex|claude>
  cerne skill install --help
  cerne skill --help

Autorização:
  Este comando é a autorização explícita para modificar somente o perfil do
  usuário atual do agente informado. init, restore e workflow setup não
  instalam skills por implicação.

Pacote:
  Usa o pacote oficial cerne-skills incorporado ao binário, sem rede. O
  manifesto, a skill cerne-context, o adaptador do agente e o schema
  cerne.context.v1 são validados antes de alterar o destino.

Destinos:
  codex:  ~/.codex/skills/cerne-context
  claude: ~/.claude/skills/cerne-context

Saídas:
  Sucesso e ajuda usam stdout. Falhas usam stderr.
  Status 0: instalada, já instalada ou atualizada; 1: falha operacional;
  2: uso inválido.

Efeitos:
  Cria auditoria privada em ~/.cerne/audit. Recusa destino desconhecido,
  symlinks e pacote incompatível; reinstalação da mesma versão é no-op.

Exemplos:
  cerne skill install codex
  cerne skill install claude
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

const workflowHelp = `Inicializa o workflow já declarado no manifesto do workspace.

Uso:
  cerne workflow setup
  cerne workflow setup --agent <codex|claude>
  cerne workflow --help

Localização:
  Procura o workspace ancestral mais próximo por knowledge/cerne.json.

Comportamento:
  Executa somente o provider declarado e já instalado dentro de knowledge.
  Um layout pronto não é alterado. Um layout parcial é recusado.
  --agent codex|claude prepara descoberta local para Spec Kit na raiz do
  workspace sem persistir agente no manifesto.

Saídas:
  Sucesso e ajuda usam stdout. Falhas usam stderr.
  Status 0: concluído, já pronto ou ajuda; 1: falha operacional; 2: uso inválido.

Efeitos:
  Registra tentativas em knowledge/runs. Não instala agentes externos, não troca
  o provider, não altera source e não autoriza rede, Git remoto ou credenciais.
  Para instalar a skill global, use cerne skill install <agent>.

Exemplo:
  cerne workflow setup
  cerne workflow setup --agent claude
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
	languageValue, commandArgs, valid := parseGlobalLanguage(args)
	if !valid {
		fmt.Fprint(stderr, (localizer{language: localization.Default}).text(messageInvalidGlobalOption))
		return 2
	}
	home, _ := os.UserHomeDir()
	environment, environmentSet := os.LookupEnv("CERNE_LANG")
	language, err := localization.Resolve(languageValue, environment, environmentSet, home)
	if err != nil {
		selected := localization.Default
		if parsed, parseErr := localization.Parse(languageValue); parseErr == nil && languageValue != "" {
			selected = parsed
		} else if parsed, parseErr := localization.Parse(environment); parseErr == nil && environmentSet {
			selected = parsed
		}
		messages := localizer{language: selected}
		var invalid localization.InvalidLanguageError
		if errors.As(err, &invalid) {
			fmt.Fprint(stderr, messages.text(messageInvalidLanguage, invalid.Value))
			return 2
		}
		renderConfigFailure(stderr, messages, err)
		return 1
	}
	return runLocalized(commandArgs, stdout, stderr, localizer{language: language}, home)
}

func runLocalized(args []string, stdout, stderr io.Writer, messages localizer, home string) int {
	if len(args) == 1 && args[0] == "--help" {
		fmt.Fprint(stdout, messages.text(messageGlobalHelp))
		return 0
	}
	if len(args) == 1 && args[0] == "--version" {
		fmt.Fprintf(stdout, "cerne %s\n", version)
		return 0
	}
	if len(args) == 0 {
		return commandUsageError(stderr, messages, "command.missing")
	}
	switch args[0] {
	case "init":
		return runInit(args[1:], stdout, stderr, messages)
	case "restore":
		return runRestore(args[1:], stdout, stderr, messages)
	case "doctor":
		return runDoctor(args[1:], stdout, stderr, messages)
	case "status":
		return runStatus(args[1:], stdout, stderr, messages)
	case "link":
		return runLink(args[1:], stdout, stderr, messages)
	case "workflow":
		return runWorkflow(args[1:], stdout, stderr, messages)
	case "context":
		return runContext(args[1:], stdout, stderr, messages)
	case "skill":
		return runSkill(args[1:], stdout, stderr, messages)
	case "config":
		return runConfig(args[1:], stdout, stderr, messages, home)
	default:
		return commandUsageError(stderr, messages, "command.unknown")
	}
}

func parseGlobalLanguage(args []string) (string, []string, bool) {
	if len(args) == 0 || args[0] != "--lang" {
		return "", args, true
	}
	if len(args) < 2 || args[1] == "" {
		return "", nil, false
	}
	return args[1], args[2:], true
}

func runConfig(args []string, stdout, stderr io.Writer, messages localizer, home string) int {
	if len(args) == 1 && args[0] == "--help" {
		fmt.Fprint(stdout, messages.text(messageConfigHelp))
		return 0
	}
	if home == "" {
		fmt.Fprint(stderr, messages.text(messageHomeUnavailable))
		return 1
	}
	if len(args) == 3 && args[0] == "set" && args[1] == "language" {
		language, err := localization.Parse(args[2])
		if err != nil {
			fmt.Fprint(stderr, messages.text(messageInvalidLanguage, args[2]))
			return 2
		}
		if err := localization.Set(home, language); err != nil {
			renderConfigFailure(stderr, messages, err)
			return 1
		}
		fmt.Fprint(stdout, messages.text(messageConfigSet, language))
		return 0
	}
	if len(args) == 2 && args[0] == "get" && args[1] == "language" {
		config, present, err := localization.Load(home)
		if err != nil {
			renderConfigFailure(stderr, messages, err)
			return 1
		}
		if !present {
			fmt.Fprint(stdout, messages.text(messageConfigGetUnset))
		} else {
			fmt.Fprint(stdout, messages.text(messageConfigGet, config.Language))
		}
		return 0
	}
	if len(args) == 2 && args[0] == "unset" && args[1] == "language" {
		if err := localization.Unset(home); err != nil {
			renderConfigFailure(stderr, messages, err)
			return 1
		}
		fmt.Fprint(stdout, messages.text(messageConfigUnset))
		return 0
	}
	fmt.Fprint(stderr, messages.text(messageConfigUsage))
	return 2
}

func renderConfigFailure(stderr io.Writer, messages localizer, err error) {
	var failure localization.ConfigFailure
	if !errors.As(err, &failure) {
		fmt.Fprint(stderr, messages.text(messageConfigRead))
		return
	}
	id := map[string]messageID{
		localization.ConfigUnsafe:  messageConfigUnsafe,
		localization.ConfigRead:    messageConfigRead,
		localization.ConfigInvalid: messageConfigInvalid,
		localization.ConfigWrite:   messageConfigWrite,
	}[failure.Code]
	if id == "" {
		id = messageConfigRead
	}
	fmt.Fprint(stderr, messages.text(id))
}

func runSkill(args []string, stdout, stderr io.Writer, messages localizer) int {
	if len(args) == 1 && args[0] == "--help" || len(args) == 2 && args[0] == "install" && args[1] == "--help" {
		fmt.Fprint(stdout, messages.text(messageSkillHelp))
		return 0
	}
	if len(args) != 2 || args[0] != "install" || !skillinstall.SupportedAgent(args[1]) {
		fmt.Fprint(stderr, messages.text("skill.usage"))
		return 2
	}
	result, err := skillinstall.Install(args[1], skillinstall.Options{})
	if err != nil {
		var failure skillinstall.Failure
		if errors.As(err, &failure) {
			id := messageID("skill.failure." + failure.Code)
			if _, ok := messages.find(id); !ok {
				id = "skill.failure.default"
			}
			fmt.Fprint(stderr, messages.text(id))
		} else {
			fmt.Fprint(stderr, messages.text("skill.failure.default"))
		}
		return 1
	}
	switch result.Outcome {
	case "already":
		fmt.Fprint(stdout, messages.text("skill.already", skillinstall.SkillName))
	case "upgraded":
		fmt.Fprint(stdout, messages.text("skill.upgraded", skillinstall.SkillName))
	default:
		fmt.Fprint(stdout, messages.text("skill.installed", skillinstall.SkillName))
	}
	fmt.Fprint(stdout, messages.text("skill.result", result.Agent, result.Version, result.Destination))
	return 0
}

func runContext(args []string, stdout, stderr io.Writer, messages localizer) int {
	if len(args) == 1 && args[0] == "--help" {
		fmt.Fprint(stdout, messages.text(messageContextHelp))
		return 0
	}
	jsonOutput := len(args) == 1 && args[0] == "--json"
	if len(args) != 0 && !jsonOutput {
		fmt.Fprint(stderr, messages.text("context.usage"))
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
		renderContext(stdout, report, messages)
	}
	if report.Status == workspace.Invalid {
		return 1
	}
	return 0
}

func renderContext(stdout io.Writer, report workspace.ContextReport, messages localizer) {
	if report.Workspace != nil && report.Workspace.Name != "" {
		fmt.Fprint(stdout, messages.text("context.workspace", report.Workspace.Name))
	}
	fmt.Fprint(stdout, messages.text("context.status", contextStatus(messages, report.Status)))
	if report.Workspace != nil {
		fmt.Fprint(stdout, messages.text("context.root", report.Workspace.Root))
	}
	if report.Knowledge != nil {
		fmt.Fprint(stdout, messages.text("context.knowledge", report.Knowledge.Path))
		contextPathLine(stdout, messages, "context.label.product", report.Knowledge.ProductPath)
		contextPathLine(stdout, messages, "context.label.specs", report.Knowledge.SpecsPath)
		contextPathLine(stdout, messages, "context.label.decisions", report.Knowledge.DecisionsPath)
		contextPathLine(stdout, messages, "context.label.policies", report.Knowledge.PoliciesPath)
	}
	if report.Source != nil {
		fmt.Fprint(stdout, messages.text("context.source", report.Source.Path))
		location := messages.text("context.location.outside")
		if report.Source.InsideWorkspace {
			location = messages.text("context.location.inside")
		}
		fmt.Fprint(stdout, messages.text("context.location", location))
	}
	if report.Workflow != nil {
		fmt.Fprintln(stdout)
		if !report.Workflow.Declared {
			fmt.Fprint(stdout, messages.text("context.workflow.not-declared"))
		} else {
			fmt.Fprint(stdout, messages.text("context.workflow", report.Workflow.Provider, contextWorkflowState(messages, report.Workflow.State)))
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
		fmt.Fprint(stdout, messages.text("context.problem", symbol, contextComponent(messages, problem.Component), contextProblemDetail(messages, problem.Code)))
		fmt.Fprint(stdout, messages.text("context.correction", contextProblemCorrection(messages, problem.Code)))
	}
}

func contextPathLine(output io.Writer, messages localizer, label messageID, path string) {
	if path != "" {
		fmt.Fprintf(output, "%s: %s\n", messages.text(label), path)
	}
}

func contextStatus(messages localizer, status workspace.Status) string {
	return messages.text(messageID("context.status." + string(status)))
}

func contextWorkflowState(messages localizer, state workspace.ContextWorkflowState) string {
	return messages.text(messageID("context.workflow." + string(state)))
}

func contextComponent(messages localizer, component string) string {
	if translated, ok := messages.find(messageID("context.component." + component)); ok {
		return translated
	}
	return component
}

func contextProblemDetail(messages localizer, code string) string {
	return messages.text(messageID("context.problem." + code + ".detail"))
}

func contextProblemCorrection(messages localizer, code string) string {
	return messages.text(messageID("context.problem." + code + ".correction"))
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

func runRestore(args []string, stdout, stderr io.Writer, messages localizer) int {
	if len(args) == 1 && args[0] == "--help" {
		fmt.Fprint(stdout, messages.text(messageRestoreHelp))
		return 0
	}
	parsed, ok := parseRestoreArgs(args)
	if !ok {
		fmt.Fprint(stderr, messages.text("restore.invalid-argument"), messages.text("restore.usage"))
		return 2
	}
	current, err := currentDirectory()
	if err != nil {
		fmt.Fprint(stderr, messages.text("common.cwd"))
		return 1
	}
	knowledgeOrigin, err := gitexec.ClassifyCloneOrigin(current, parsed.KnowledgeOrigin)
	if err != nil {
		return restoreUsageError(stderr, messages, "restore.invalid-knowledge-origin")
	}
	sourceInput := parsed.SourceValue
	if parsed.SourceMode == workspace.SourceClone {
		sourceOrigin, err := gitexec.ClassifyCloneOrigin(current, parsed.SourceValue)
		if err != nil {
			return restoreUsageError(stderr, messages, "restore.invalid-source-origin")
		}
		sourceInput = sourceOrigin.Location
	}
	clone, err := gitexec.FindClone()
	if err != nil {
		fmt.Fprint(stderr, messages.text("common.git"))
		return 1
	}
	inspect, err := gitexec.FindLinkInspector()
	if err != nil {
		fmt.Fprint(stderr, messages.text("common.git"))
		return 1
	}
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprint(stderr, messages.text("common.home"))
		return 1
	}
	result, err := workspace.Restore(current, home, workspace.RestoreRequest{
		KnowledgeOrigin: knowledgeOrigin.Location, SourceMode: parsed.SourceMode, SourceInput: sourceInput,
	}, adaptLink(inspect), clone, workflowexec.Resolve)
	if err != nil {
		var failure workspace.RestoreFailure
		if errors.As(err, &failure) {
			renderFailure(stderr, messages, "restore", failure.Code, failure.Cause, "", failure.Correction)
		} else {
			fmt.Fprint(stderr, messages.text("restore.failure.default"))
		}
		return 1
	}
	fmt.Fprint(stdout, messages.text("restore.result", result.Name, result.KnowledgePath))
	if result.SourceMode == workspace.SourceClone {
		fmt.Fprint(stdout, messages.text("restore.source.cloned", result.SourcePath))
	} else {
		fmt.Fprint(stdout, messages.text("restore.source.linked", result.SourcePath))
		if result.ManifestChanged {
			fmt.Fprint(stdout, messages.text("restore.manifest.changed"))
		}
	}
	return 0
}

func restoreUsageError(stderr io.Writer, messages localizer, cause messageID) int {
	fmt.Fprint(stderr, messages.text(cause), messages.text("restore.usage"))
	return 2
}

func runInit(args []string, stdout, stderr io.Writer, messages localizer) int {
	if len(args) == 1 && args[0] == "--help" {
		fmt.Fprint(stdout, messages.text(messageInitHelp))
		return 0
	}
	parsed, ok := parseInitArgs(args)
	if !ok {
		return initUsageError(stderr, messages, "init.invalid-argument")
	}
	if err := workspace.ValidateName(parsed.Name); err != nil {
		return initUsageError(stderr, messages, "init.invalid-name")
	}
	var definition workspace.WorkflowDefinition
	if parsed.Workflow != "" {
		var err error
		definition, err = workflowexec.Resolve(parsed.Workflow)
		if err != nil {
			return initUsageError(stderr, messages, "init.invalid-workflow")
		}
	}
	var current string
	var origin gitexec.CloneOrigin
	var err error
	if parsed.SourceMode == workspace.SourceClone {
		current, err = currentDirectory()
		if err != nil {
			fmt.Fprint(stderr, messages.text("common.cwd"))
			return 1
		}
		origin, err = gitexec.ClassifyCloneOrigin(current, parsed.SourceValue)
		if err != nil {
			return initUsageError(stderr, messages, "init.invalid-clone-origin")
		}
	}

	initRepository, err := gitexec.Find()
	if err != nil {
		fmt.Fprint(stderr, messages.text("common.git"))
		return 1
	}
	if current == "" {
		current, err = currentDirectory()
		if err != nil {
			fmt.Fprint(stderr, messages.text("common.cwd"))
			return 1
		}
	}
	if parsed.SourceMode != "" {
		inspect, inspectErr := gitexec.FindLinkInspector()
		if inspectErr != nil {
			fmt.Fprint(stderr, messages.text("common.git"))
			return 1
		}
		request := workspace.SourceInitRequest{Mode: parsed.SourceMode, Input: parsed.SourceValue}
		var clone workspace.CloneSource
		if parsed.SourceMode == workspace.SourceClone {
			clone, err = gitexec.FindClone()
			if err != nil {
				fmt.Fprint(stderr, messages.text("common.git"))
				return 1
			}
			request.Input = origin.Location
			request.OriginTransport = origin.Transport
			request.OriginFingerprint = origin.Fingerprint
		}
		result, workflow, err := workspace.InitWithSourceAndWorkflowAndAgent(current, parsed.Name, request, definition, parsed.Agent, initRepository, adaptLink(inspect), clone)
		if err != nil {
			var workflowFailure workspace.WorkflowFailure
			if errors.As(err, &workflowFailure) {
				fmt.Fprint(stderr, messages.text("init.workflow.failure", parsed.Workflow, workflowCause(messages, err)))
				fmt.Fprint(stderr, messages.text("init.workflow.correction", parsed.Workflow, workflowSetupCommand(parsed.Agent), filepath.Dir(result.KnowledgePath)))
				return 1
			}
			renderSourceInitFailure(stderr, messages, err)
			return 1
		}
		fmt.Fprint(stdout, messages.text("init.result.knowledge", result.Name, result.KnowledgePath))
		if result.SourceMode == workspace.SourceLocal {
			fmt.Fprint(stdout, messages.text("init.source.linked", result.SourcePath))
		} else {
			fmt.Fprint(stdout, messages.text("init.source.cloned", result.SourcePath))
		}
		if parsed.Workflow != "" {
			fmt.Fprint(stdout, messages.text("init.workflow.result", parsed.Workflow, workflowState(messages, workflow.State)))
			if parsed.Agent != "" && workflow.State != workspace.WorkflowPending {
				fmt.Fprint(stdout, messages.text("agent.discovery", parsed.Agent))
			}
			if workflow.State == workspace.WorkflowPending {
				fmt.Fprint(stderr, messages.text("workflow.pending.warning", definition.Executor, parsed.Workflow))
				fmt.Fprint(stderr, messages.text("workflow.pending.correction", parsed.Workflow, workflowSetupCommand(parsed.Agent)))
			}
		}
		return 0
	}
	if parsed.Workflow != "" {
		result, workflow, err := workspace.InitWithWorkflowAndAgent(current, parsed.Name, definition, parsed.Agent, initRepository)
		if err != nil {
			if result.KnowledgePath == "" {
				if errors.Is(err, workspace.ErrUnsafeDestination) {
					fmt.Fprint(stderr, messages.text("init.destination-unsafe"))
				} else {
					fmt.Fprint(stderr, messages.text("init.failure.default"))
				}
				return 1
			}
			fmt.Fprint(stderr, messages.text("init.workflow.failure", parsed.Workflow, workflowCause(messages, err)))
			fmt.Fprint(stderr, messages.text("init.workflow.correction", parsed.Workflow, workflowSetupCommand(parsed.Agent), filepath.Dir(result.KnowledgePath)))
			return 1
		}
		fmt.Fprint(stdout, messages.text("init.result.workflow", result.Name, result.KnowledgePath, result.SourcePath, parsed.Workflow, workflowState(messages, workflow.State)))
		if parsed.Agent != "" && workflow.State != workspace.WorkflowPending {
			fmt.Fprint(stdout, messages.text("agent.discovery", parsed.Agent))
		}
		if workflow.State == workspace.WorkflowPending {
			fmt.Fprint(stderr, messages.text("workflow.pending.warning", definition.Executor, parsed.Workflow))
			fmt.Fprint(stderr, messages.text("workflow.pending.correction", parsed.Workflow, workflowSetupCommand(parsed.Agent)))
		}
		return 0
	}

	result, err := workspace.Init(current, parsed.Name, initRepository)
	if err != nil {
		if errors.Is(err, workspace.ErrUnsafeDestination) {
			fmt.Fprint(stderr, messages.text("init.destination-unsafe"))
		} else {
			fmt.Fprint(stderr, messages.text("init.failure.default"))
		}
		return 1
	}

	fmt.Fprint(stdout, messages.text("init.result", result.Name, result.KnowledgePath, result.SourcePath))
	return 0
}

type initArguments struct {
	Name        string
	Workflow    string
	Agent       string
	SourceMode  workspace.SourceMode
	SourceValue string
}

func parseInitArgs(args []string) (initArguments, bool) {
	if len(args) == 1 && !strings.HasPrefix(args[0], "--") {
		return initArguments{Name: args[0]}, true
	}
	if (len(args) != 3 && len(args) != 5 && len(args) != 7) || strings.HasPrefix(args[0], "--") {
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
		case "--agent":
			if parsed.Agent != "" || value != "codex" && value != "claude" {
				return initArguments{}, false
			}
			parsed.Agent = value
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
	if parsed.Agent != "" && parsed.Workflow != "speckit" {
		return initArguments{}, false
	}
	return parsed, true
}

func renderSourceInitFailure(stderr io.Writer, messages localizer, err error) {
	var cloneFailure workspace.SourceInitFailure
	if errors.As(err, &cloneFailure) {
		renderFailure(stderr, messages, "source-init", cloneFailure.Code, cloneFailure.Cause, "", cloneFailure.Correction)
		return
	}
	var linkFailure workspace.LinkFailure
	if errors.As(err, &linkFailure) {
		renderFailure(stderr, messages, "link", linkFailure.Code, linkFailure.Cause, linkFailure.Path, linkFailure.Correction)
		return
	}
	if errors.Is(err, workspace.ErrUnsafeDestination) {
		fmt.Fprint(stderr, messages.text("init.destination-unsafe"))
	} else {
		fmt.Fprint(stderr, messages.text("init.failure.default"))
	}
}

func runWorkflow(args []string, stdout, stderr io.Writer, messages localizer) int {
	if len(args) == 1 && args[0] == "--help" {
		fmt.Fprint(stdout, messages.text(messageWorkflowHelp))
		return 0
	}
	agent, ok := parseWorkflowSetupArgs(args)
	if !ok {
		fmt.Fprint(stderr, messages.text("workflow.usage"))
		return 2
	}
	current, err := currentDirectory()
	if err != nil {
		fmt.Fprint(stderr, messages.text("common.cwd"))
		return 1
	}
	result, err := workspace.SetupWorkflowWithAgent(current, workflowexec.Resolve, agent)
	if err != nil {
		var failure workspace.WorkflowFailure
		if errors.As(err, &failure) {
			renderFailure(stderr, messages, "workflow", failure.Code, failure.Cause, "", failure.Correction)
		} else {
			fmt.Fprint(stderr, messages.text("workflow.failure.default"))
		}
		return 1
	}
	if result.State == workspace.WorkflowPending {
		fmt.Fprint(stderr, messages.text("workflow.executor.missing", result.Executor, result.Provider))
		return 1
	}
	fmt.Fprint(stdout, messages.text("workflow.result", result.Provider, result.KnowledgePath))
	if result.State == workspace.WorkflowUnchanged {
		fmt.Fprint(stdout, messages.text("workflow.unchanged"))
	} else {
		fmt.Fprint(stdout, messages.text("workflow.completed"))
	}
	if result.Agent != "" && result.Discovery == workspace.WorkflowDiscoveryReady {
		fmt.Fprint(stdout, messages.text("agent.discovery", result.Agent))
	}
	return 0
}

func parseWorkflowSetupArgs(args []string) (string, bool) {
	if len(args) == 1 && args[0] == "setup" {
		return "", true
	}
	if len(args) == 3 && args[0] == "setup" && args[1] == "--agent" && (args[2] == "codex" || args[2] == "claude") {
		return args[2], true
	}
	return "", false
}

func workflowSetupCommand(agent string) string {
	if agent == "" {
		return "cerne workflow setup"
	}
	return "cerne workflow setup --agent " + agent
}

func workflowState(messages localizer, state workspace.WorkflowState) string {
	return messages.text(messageID("workflow.state." + string(state)))
}

func workflowCause(messages localizer, err error) string {
	var failure workspace.WorkflowFailure
	if errors.As(err, &failure) {
		return localizedFailureCause(messages, "workflow", failure.Code, failure.Cause)
	}
	return messages.text("failure.operational")
}

func runDoctor(args []string, stdout, stderr io.Writer, messages localizer) int {
	if len(args) == 1 && args[0] == "--help" {
		fmt.Fprint(stdout, messages.text(messageDoctorHelp))
		return 0
	}
	if len(args) != 0 {
		fmt.Fprint(stderr, messages.text("doctor.usage"))
		return 2
	}

	current, err := currentDirectory()
	if err != nil {
		fmt.Fprint(stderr, messages.text("common.cwd"))
		return 1
	}
	inspect, err := gitexec.FindInspector()
	if err != nil {
		inspect = nil
	}
	diagnosis := workspace.DoctorWithWorkflow(current, adaptGit(inspect), adaptAccess, workflowexec.Resolve)
	renderDiagnosis(stdout, diagnosis, messages)
	if diagnosis.Status == workspace.Invalid {
		return 1
	}
	return 0
}

func runStatus(args []string, stdout, stderr io.Writer, messages localizer) int {
	if len(args) == 1 && args[0] == "--help" {
		fmt.Fprint(stdout, messages.text(messageStatusHelp))
		return 0
	}
	if len(args) != 0 {
		fmt.Fprint(stderr, messages.text("status.usage"))
		return 2
	}

	current, err := currentDirectory()
	if err != nil {
		fmt.Fprint(stderr, messages.text("common.cwd"))
		return 1
	}
	collect, err := gitexec.FindStatus()
	if err != nil {
		fmt.Fprint(stderr, messages.text("common.git"))
		return 1
	}
	report, err := workspace.CurrentStatus(current, adaptStatus(collect))
	if err != nil {
		var failure workspace.StatusFailure
		if errors.As(err, &failure) {
			renderFailure(stderr, messages, "status", failure.Code, failure.Cause, failure.Path, failure.Correction)
		} else {
			fmt.Fprint(stderr, messages.text("status.failure.default"))
		}
		return 1
	}
	renderStatus(stdout, report, messages)
	return 0
}

func runLink(args []string, stdout, stderr io.Writer, messages localizer) int {
	if len(args) == 1 && args[0] == "--help" {
		fmt.Fprint(stdout, messages.text(messageLinkHelp))
		return 0
	}
	source, replace, ok := parseLinkArgs(args)
	if !ok {
		fmt.Fprint(stderr, messages.text("link.usage"))
		return 2
	}

	current, err := currentDirectory()
	if err != nil {
		fmt.Fprint(stderr, messages.text("common.cwd"))
		return 1
	}
	inspect, err := gitexec.FindLinkInspector()
	if err != nil {
		fmt.Fprint(stderr, messages.text("common.git"))
		return 1
	}
	result, err := workspace.Link(current, workspace.LinkRequest{SourceInput: source, Replace: replace}, adaptLink(inspect))
	if err != nil {
		var failure workspace.LinkFailure
		if errors.As(err, &failure) {
			renderFailure(stderr, messages, "link", failure.Code, failure.Cause, failure.Path, failure.Correction)
		} else {
			fmt.Fprint(stderr, messages.text("link.failure.default"))
		}
		return 1
	}
	renderLink(stdout, result, messages)
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

func renderDiagnosis(stdout io.Writer, diagnosis workspace.Diagnosis, messages localizer) {
	for _, check := range diagnosis.Checks {
		label, detail, correction := localizedCheck(messages, check)
		fmt.Fprint(stdout, messages.text("diagnosis.line", symbol(check.Severity), label, detail))
		if check.Correction != "" {
			fmt.Fprint(stdout, messages.text("diagnosis.correction", correction))
		}
		fmt.Fprintln(stdout)
	}
	switch diagnosis.Status {
	case workspace.Invalid:
		fmt.Fprint(stdout, messages.text("diagnosis.invalid"))
	case workspace.Warnings:
		fmt.Fprint(stdout, messages.text("diagnosis.warning"))
	default:
		fmt.Fprint(stdout, messages.text("diagnosis.healthy"))
	}
}

func renderStatus(stdout io.Writer, report workspace.WorkspaceReport, messages localizer) {
	fmt.Fprint(stdout, messages.text("status.project", report.ProjectName))
	fmt.Fprint(stdout, messages.text("status.workspace", report.Root))
	for index, repository := range report.Repositories {
		if index > 0 {
			fmt.Fprintln(stdout)
		}
		fmt.Fprintln(stdout, messages.text(messageID("status.repository."+repository.Name)))
		fmt.Fprint(stdout, messages.text("status.path", repository.Path))
		branch := repository.Branch
		if branch == gitexec.DetachedHEAD {
			branch = messages.text("status.branch.detached-head")
		}
		commit := repository.Commit
		if commit == gitexec.NoCommits {
			commit = messages.text("status.commit.no-commits")
		}
		fmt.Fprint(stdout, messages.text("status.branch", branch))
		fmt.Fprint(stdout, messages.text("status.commit", commit))
		fmt.Fprint(stdout, messages.text("status.state", messages.text(messageID("status.state."+repository.State))))
		fmt.Fprint(stdout, messages.text("status.modified", repository.ModifiedCount))
		fmt.Fprint(stdout, messages.text("status.staged", repository.StagedCount))
		fmt.Fprint(stdout, messages.text("status.untracked", repository.UntrackedCount))
	}
}

func renderLink(stdout io.Writer, result workspace.LinkResult, messages localizer) {
	fmt.Fprint(stdout, messages.text("link.project", result.ProjectName))
	if !result.Changed {
		fmt.Fprint(stdout, messages.text("link.current", result.NewSource))
		fmt.Fprint(stdout, messages.text("link.unchanged"))
		return
	}
	if result.PreviousSource != "" {
		fmt.Fprint(stdout, messages.text("link.previous", result.PreviousSource))
	}
	fmt.Fprint(stdout, messages.text("link.new", result.NewSource))
	fmt.Fprint(stdout, messages.text("link.updated"))
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

func initUsageError(stderr io.Writer, messages localizer, cause messageID) int {
	fmt.Fprint(stderr, messages.text(cause), messages.text("init.usage"))
	return 2
}

func commandUsageError(stderr io.Writer, messages localizer, id messageID) int {
	fmt.Fprint(stderr, messages.text(id))
	return 2
}
