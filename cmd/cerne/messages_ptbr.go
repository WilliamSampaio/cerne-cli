package main

var portugueseBrazilMessages = map[messageID]string{
	messageSkillHelp:    skillHelp,
	messageContextHelp:  contextHelp,
	messageRestoreHelp:  restoreHelp,
	messageInitHelp:     initHelp,
	messageWorkflowHelp: workflowHelp,
	messageDoctorHelp:   doctorHelp,
	messageStatusHelp:   statusHelp,
	messageLinkHelp:     linkHelp,
	messageGlobalHelp: `Cerne administra workspaces com repositórios Git independentes de conhecimento e código-fonte.

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
`,
	messageConfigHelp: `Administra preferências do usuário do Cerne.

Uso:
  cerne config set language <en|pt-BR>
  cerne config get language
  cerne config unset language
  cerne config --help

Armazenamento:
  A preferência é salva em ~/.cerne/config.json para o usuário atual. O comando
  recusa symlinks e substitui o arquivo de forma atômica.

Precedência:
  --lang, CERNE_LANG, preferência salva e pt-BR.

Compatibilidade:
  pt-BR permanece o padrão nesta versão. en será o padrão no Cerne 1.0.
`,
	messageInvalidLanguage:                                    "erro: idioma inválido: %q\ncorreção: use en ou pt-BR\n",
	messageInvalidGlobalOption:                                "erro: opção global inválida\nuso: cerne [--lang <en|pt-BR>] <comando> [argumentos]\n",
	messageConfigUsage:                                        "erro: argumento inválido\nuso: cerne config <set language <en|pt-BR>|get language|unset language>\n",
	messageConfigSet:                                          "Idioma salvo: %s\n",
	messageConfigGet:                                          "Idioma salvo: %s\n",
	messageConfigGetUnset:                                     "Idioma não definido. Padrão atual: pt-BR\n",
	messageConfigUnset:                                        "Preferência de idioma removida.\n",
	messageConfigUnsafe:                                       "erro: configuração do usuário insegura\ncorreção: remova symlinks de ~/.cerne ou ~/.cerne/config.json e tente novamente\n",
	messageConfigRead:                                         "erro: não foi possível ler a configuração do usuário\ncorreção: verifique as permissões de ~/.cerne/config.json\n",
	messageConfigInvalid:                                      "erro: configuração de idioma inválida\ncorreção: use cerne --lang pt-BR config set language <en|pt-BR> para repará-la\n",
	messageConfigWrite:                                        "erro: não foi possível atualizar a configuração do usuário\ncorreção: verifique as permissões de ~/.cerne e tente novamente\n",
	messageHomeUnavailable:                                    "erro: não foi possível localizar o diretório pessoal\ncorreção: configure um diretório pessoal acessível\n",
	"skill.usage":                                             "erro: argumento inválido\nuso: cerne skill install <codex|claude|gemini> [cerne-context|cerne-git-workflow]\n",
	"skill.installed":                                         "Skill instalada: %s\n",
	"skill.already":                                           "Skill já instalada: %s\n",
	"skill.upgraded":                                          "Skill atualizada: %s\n",
	"skill.result":                                            "Agente: %s\nVersão: %s\nDestino: %s\n",
	"skill.failure.default":                                   "erro: não foi possível instalar a skill\ncorreção: verifique permissões e tente novamente\n",
	"skill.failure.home-unavailable":                          "erro: não foi possível localizar o diretório pessoal\ncorreção: configure um diretório pessoal acessível\n",
	"skill.failure.destination-invalid":                       "erro: destino do agente inválido\ncorreção: configure um diretório pessoal acessível\n",
	"skill.failure.audit-start-failed":                        "erro: não foi possível registrar a tentativa de instalação\ncorreção: verifique a segurança e as permissões de ~/.cerne/audit\n",
	"skill.failure.install-failed":                            "erro: não foi possível instalar a skill\ncorreção: verifique permissões e tente novamente\n",
	"skill.failure.audit-finalization-failed":                 "erro: não foi possível finalizar a auditoria da instalação\ncorreção: verifique ~/.cerne/audit antes de tentar novamente\n",
	"skill.failure.package-unavailable":                       "erro: pacote oficial cerne-skills incorporado está inacessível\ncorreção: verifique o diretório temporário e reinstale o Cerne\n",
	"skill.failure.manifest-invalid":                          "erro: manifesto do pacote cerne-skills inválido\ncorreção: reinstale o Cerne\n",
	"skill.failure.manifest-incompatible":                     "erro: pacote cerne-skills incompatível\ncorreção: atualize ou reinstale o Cerne\n",
	"skill.failure.adapter-missing":                           "erro: adaptador do agente ausente no pacote cerne-skills\ncorreção: atualize ou reinstale o Cerne\n",
	"skill.failure.skill-missing":                             "erro: skill ausente no pacote cerne-skills\ncorreção: use uma skill oficial suportada ou atualize o Cerne\n",
	"skill.failure.unsafe-package":                            "erro: pacote cerne-skills contém conteúdo inseguro\ncorreção: reinstale o Cerne\n",
	"skill.failure.destination-inaccessible":                  "erro: destino do agente inacessível\ncorreção: verifique permissões no perfil do agente\n",
	"skill.failure.unknown-destination":                       "erro: destino existente não é gerenciado pelo Cerne\ncorreção: mova o conteúdo existente antes de instalar ou atualizar\n",
	"skill.failure.promotion-failed":                          "erro: não foi possível promover a instalação\ncorreção: verifique permissões no perfil do agente\n",
	"context.usage":                                           "erro: argumento inválido\nuso: cerne context [--json]\n",
	"context.workspace":                                       "Workspace: %s\n",
	"context.status":                                          "Status: %s\n",
	"context.root":                                            "Root: %s\n",
	"context.knowledge":                                       "\nKnowledge: %s\n",
	"context.source":                                          "\nSource: %s\n",
	"context.location":                                        "Localização: %s\n",
	"context.workflow.not-declared":                           "Workflow: não declarado\n",
	"context.workflow":                                        "Workflow: %s (%s)\n",
	"context.problem":                                         "%s %s: %s\n",
	"context.correction":                                      "  Correção: %s\n",
	"context.label.product":                                   "Product",
	"context.label.specs":                                     "Specs",
	"context.label.decisions":                                 "Decisions",
	"context.label.policies":                                  "Policies",
	"context.location.inside":                                 "interno ao workspace",
	"context.location.outside":                                "externo ao workspace",
	"context.status.healthy":                                  "saudável",
	"context.status.warning":                                  "aviso",
	"context.status.invalid":                                  "inválido",
	"context.workflow.pending":                                "pendente",
	"context.workflow.ready":                                  "pronto",
	"context.workflow.invalid":                                "inválido",
	"context.workflow.unknown-provider":                       "provider desconhecido",
	"context.component.workspace":                             "Workspace",
	"context.component.manifest":                              "Manifesto",
	"context.component.knowledge":                             "Knowledge",
	"context.component.source":                                "Source",
	"context.component.workflow":                              "Workflow",
	"context.problem.workspace-not-found.detail":              "não localizado",
	"context.problem.workspace-not-found.correction":          "execute o comando dentro de um workspace Cerne",
	"context.problem.manifest-invalid.detail":                 "ausente ou inválido",
	"context.problem.manifest-invalid.correction":             "corrija ou restaure knowledge/cerne.json",
	"context.problem.manifest-version-unsupported.detail":     "versão não suportada",
	"context.problem.manifest-version-unsupported.correction": "use uma versão compatível do Cerne",
	"context.problem.knowledge-invalid.detail":                "ausente ou inseguro",
	"context.problem.knowledge-invalid.correction":            "restaure o diretório knowledge",
	"context.problem.source-invalid.detail":                   "ausente ou inseguro",
	"context.problem.source-invalid.correction":               "corrija o caminho source no manifesto",
	"context.problem.required-directory-invalid.detail":       "diretório obrigatório ausente ou inválido",
	"context.problem.required-directory-invalid.correction":   "restaure o diretório indicado pelo componente",
	"context.problem.workflow-pending.detail":                 "pendente",
	"context.problem.workflow-pending.correction":             "execute cerne workflow setup quando o provider estiver disponível",
	"context.problem.workflow-invalid.detail":                 "estrutura inválida ou parcial",
	"context.problem.workflow-invalid.correction":             "corrija a estrutura antes de continuar",
	"context.problem.workflow-unknown-provider.detail":        "provider não suportado",
	"context.problem.workflow-unknown-provider.correction":    "use speckit ou openspec no manifesto",
	messageGitHelp:                                            gitHelp,
	"git.usage":                                               "erro: argumento inválido\nuso: cerne git inspect --agent <codex|claude|gemini> --task <task-id> --json\n     cerne git branch create --name <branch> --base knowledge=<base> --base source=<base> --state <state-id> --confirm --agent <codex|claude|gemini> --task <task-id> --json\n",
	"git.inspect.usage":                                       "erro: argumento inválido\nuso: cerne git inspect --agent <codex|claude|gemini> --task <task-id> --json\n",
	"git.branch.usage":                                        "erro: argumento inválido\nuso: cerne git branch create --name <branch> --base knowledge=<base> --base source=<base> --state <state-id> --confirm --agent <codex|claude|gemini> --task <task-id> --json\n",
	"git.commit.usage":                                        "erro: argumento inválido\nuso: cerne git commit <repository> --message <subject> --include <path> --state <state-id> --confirm --agent <codex|claude|gemini> --task <task-id> --json\n",
	"git.push.usage":                                          "erro: argumento inválido\nuso: cerne git push <repository> --remote <name> --branch <branch> --state <state-id> --confirm --agent <codex|claude|gemini> --task <task-id> --json\n",
	"git.pr.usage":                                            "erro: argumento inválido\nuso: cerne git pr create <repository> --remote <name> --base <branch> --head <branch> --title <title> --body-file <path> --state <state-id> --confirm --agent <codex|claude|gemini> --task <task-id> --json\n",
	"git.failure":                                             "erro: não foi possível consultar o Git do workspace\ncorreção: verifique o workspace e tente novamente\n",
	"command.missing":                                         "erro: informe um comando\nuso: cerne <init|restore|doctor|status|link|workflow|context|skill|git|config>\n",
	"command.unknown":                                         "erro: comando desconhecido\nuso: cerne <init|restore|doctor|status|link|workflow|context|skill|git|config>\n",
	"common.cwd":                                              "erro: não foi possível obter o diretório atual\ncorreção: execute o comando em um diretório acessível\n",
	"common.git":                                              "erro: Git indisponível\ncorreção: instale o Git e disponibilize-o no PATH\n",
	"common.home":                                             "erro: não foi possível localizar o diretório pessoal\ncorreção: configure um diretório pessoal acessível\n",
	"restore.usage":                                           "uso: cerne restore <origem-knowledge> (--source <caminho> | --clone <origem-source>)\n",
	"restore.invalid-argument":                                "erro: argumento inválido\n",
	"restore.invalid-knowledge-origin":                        "erro: origem do knowledge inválida\n",
	"restore.invalid-source-origin":                           "erro: origem de clone do source inválida\n",
	"restore.failure.default":                                 "erro: não foi possível restaurar o workspace\ncorreção: verifique a auditoria e tente novamente\n",
	"restore.result":                                          "Workspace %q restaurado.\nKnowledge: %s\n",
	"restore.source.cloned":                                   "Source clonado: %s\n",
	"restore.source.linked":                                   "Source vinculado: %s\n",
	"restore.manifest.changed":                                "Manifesto: referência de source atualizada.\n",
	"init.usage":                                              "uso: cerne init <project-name> [--source <caminho> | --clone <origem>] [--workflow <speckit|openspec> [--agent <codex|claude>]]\n",
	"init.invalid-argument":                                   "erro: argumento inválido\n",
	"init.invalid-name":                                       "erro: nome de projeto inválido; use de 1 a 255 caracteres ASCII, comece por letra ou número e não use nomes reservados\n",
	"init.invalid-workflow":                                   "erro: workflow inválido: use speckit ou openspec\n",
	"init.invalid-clone-origin":                               "erro: origem de clone inválida\n",
	"init.destination-unsafe":                                 "erro: destino inseguro\ncorreção: escolha um destino inexistente ou vazio\n",
	"init.failure.default":                                    "erro: não foi possível criar o workspace\ncorreção: verifique permissões e tente novamente\n",
	"init.workflow.failure":                                   "erro: não foi possível inicializar workflow %s: %s\n",
	"init.workflow.correction":                                "correção: corrija ou atualize %s e execute %q dentro de %s\n",
	"init.result":                                             "Workspace %q criado.\nKnowledge: %s\nSource: %s\n",
	"init.result.knowledge":                                   "Workspace %q criado.\nKnowledge: %s\n",
	"init.source.linked":                                      "Source vinculado: %s\n",
	"init.source.cloned":                                      "Source clonado: %s\n",
	"init.workflow.result":                                    "Workflow: %s\nSetup: %s\n",
	"init.result.workflow":                                    "Workspace %q criado.\nKnowledge: %s\nSource: %s\nWorkflow: %s\nSetup: %s\n",
	"workflow.state.configured":                               "concluído",
	"workflow.state.pending":                                  "pendente",
	"agent.discovery":                                         "Agent: %s\nDescoberta: pronta\n",
	"workflow.pending.warning":                                "aviso: executável %q não encontrado; workflow %s não inicializado\n",
	"workflow.pending.correction":                             "correção: instale %s e execute %q dentro do workspace\n",
	"workflow.usage":                                          "erro: argumento inválido\nuso: cerne workflow setup [--agent <codex|claude>]\n",
	"workflow.failure.default":                                "erro: não foi possível inicializar o workflow\ncorreção: verifique o workspace e tente novamente\n",
	"workflow.executor.missing":                               "erro: executável %q não encontrado\ncorreção: instale %s e execute novamente\n",
	"workflow.result":                                         "Workflow: %s\nKnowledge: %s\n",
	"workflow.unchanged":                                      "Nenhuma alteração necessária.\n",
	"workflow.completed":                                      "Setup concluído.\n",
	"doctor.usage":                                            "erro: argumento inválido\nuso: cerne doctor\n",
	"diagnosis.line":                                          "%s %s: %s",
	"diagnosis.correction":                                    "; correção: %s",
	"diagnosis.invalid":                                       "Workspace inválido\n",
	"diagnosis.warning":                                       "Workspace com avisos\n",
	"diagnosis.healthy":                                       "Workspace saudável\n",
	"status.usage":                                            "erro: argumento inválido\nuso: cerne status\n",
	"status.failure.default":                                  "erro: não foi possível consultar o workspace\ncorreção: verifique o workspace e tente novamente\n",
	"status.project":                                          "Projeto: %s\n",
	"status.workspace":                                        "Workspace: %s\n\n",
	"status.path":                                             "  Caminho: %s\n",
	"status.branch":                                           "  Branch: %s\n",
	"status.commit":                                           "  Commit: %s\n",
	"status.state":                                            "  Estado: %s\n",
	"status.modified":                                         "  Modificados: %d\n",
	"status.staged":                                           "  Em stage: %d\n",
	"status.untracked":                                        "  Não rastreados: %d\n",
	"status.repository.knowledge":                             "Knowledge",
	"status.repository.source":                                "Source",
	"status.state.clean":                                      "limpo",
	"status.state.pending":                                    "alterações pendentes",
	"status.branch.detached-head":                             "detached HEAD",
	"status.commit.no-commits":                                "sem commits",
	"link.usage":                                              "erro: argumento inválido\nuso: cerne link <caminho> [--replace]\n",
	"link.failure.default":                                    "erro: não foi possível vincular o source\ncorreção: verifique o workspace e tente novamente\n",
	"link.project":                                            "Projeto: %s\n",
	"link.current":                                            "Source atual: %s\n",
	"link.unchanged":                                          "Nenhuma alteração necessária.\n",
	"link.previous":                                           "Source anterior: %s\n",
	"link.new":                                                "Novo source: %s\n",
	"link.updated":                                            "Manifesto atualizado.\n",
	"failure.cause":                                           "erro: %s\n",
	"failure.cause.path":                                      "erro: %s: %s\n",
	"failure.correction":                                      "correção: %s\n",
	"failure.operational":                                     "falha operacional",
	"failure.check-and-retry":                                 "verifique o workspace e tente novamente",
}
