# Data Model: Contexto Estrutural do Workspace

## ContextReport

Retrato efêmero que alimenta as duas apresentações do comando.

| Campo | Tipo | Regra |
|---|---|---|
| `schema_version` | inteiro | sempre `1` |
| `status` | enum | `healthy`, `warning` ou `invalid` |
| `workspace` | WorkspaceContext opcional | somente com root comprovado |
| `knowledge` | KnowledgeContext opcional | somente com diretório base comprovado |
| `source` | SourceContext opcional | somente com manifesto compatível e path seguro existente |
| `workflow` | WorkflowContext opcional | quando a declaração ou ausência puder ser comprovada |
| `problems` | lista de ContextProblem | sempre presente, inclusive vazia |

Derivação de status: qualquer `error` produz `invalid`; na ausência de erros, qualquer `warning`
produz `warning`; lista vazia produz `healthy`.

## WorkspaceContext

| Campo | Tipo | Regra |
|---|---|---|
| `name` | string opcional | manifesto v1 válido e igual ao nome da raiz |
| `root` | path | absoluto, físico, canônico e nativo |

O primeiro candidato ancestral Cerne é a fronteira da consulta. Um candidato parcial não é
atravessado para procurar outro workspace.

## KnowledgeContext

| Campo | Tipo | Regra |
|---|---|---|
| `path` | path | diretório `knowledge` comprovado |
| `product_path` | path opcional | diretório regular comprovado |
| `specs_path` | path opcional | diretório canônico comprovado para o provider |
| `decisions_path` | path opcional | diretório regular comprovado |
| `policies_path` | path opcional | diretório regular comprovado |

`knowledge/runs` continua obrigatório pela estrutura do workspace, mas não é exposto como contexto
de agente. Se inválido, gera `required-directory-invalid` no componente `knowledge.runs`.

## SourceContext

| Campo | Tipo | Regra |
|---|---|---|
| `path` | path | source existente, regular, seguro e canônico |
| `inside_workspace` | booleano | contenção física de `path` em `workspace.root` |

Não existem campos de origem, remote, clone, link ou acesso do editor. Source não pode coincidir,
conter ou ficar dentro de knowledge, nem coincidir ou conter a raiz do workspace.

## WorkflowContext

| Campo | Tipo | Regra |
|---|---|---|
| `declared` | booleano | fato do manifesto compatível |
| `provider` | string opcional | valor declarado quando comprovado |
| `state` | enum | `not-declared`, `pending`, `ready`, `invalid`, `unknown-provider` |

### State derivation

```text
sem declaração ───────────────────────────────> not-declared
provider conhecido + raiz ausente ───────────> pending
provider conhecido + layout/specs válidos ───> ready
provider conhecido + estrutura parcial ──────> invalid
provider desconhecido ────────────────────────> unknown-provider
```

São estados de uma fotografia, não transições persistidas. Disponibilidade do executável não
participa da derivação.

### Workflow problem matrix

| Situação | Estado | Problemas |
|---|---|---|
| Sem declaração | `not-declared` | nenhum |
| Raiz do provider conhecido ausente | `pending` | `workflow-pending` |
| Layout e specs regulares | `ready` | nenhum |
| Marker ausente, vazio, irregular ou symlink | `invalid` | `workflow-invalid` |
| Symlink dentro da raiz governada | `invalid` | `workflow-invalid` |
| Repositório `.git` aninhado | `invalid` | `workflow-invalid` |
| Specs ausente, irregular ou symlink após materialização | `invalid` | `required-directory-invalid` em `knowledge.specs` e `workflow-invalid` |
| Provider desconhecido | `unknown-provider` | `workflow-unknown-provider` |

## ContextProblem

| Campo | Tipo | Valores/Regra |
|---|---|---|
| `code` | string | catálogo fechado da spec |
| `severity` | string | `warning` ou `error` |
| `component` | string | componente lógico estável, nunca path ou erro bruto |

Componentes iniciais: `workspace`, `manifest`, `knowledge`, `knowledge.product`,
`knowledge.specs`, `knowledge.decisions`, `knowledge.policies`, `knowledge.runs`, `source` e
`workflow`.

## Validation and dependency order

1. Resolver o diretório inicial e localizar o primeiro candidato ancestral.
2. Comprovar root e `knowledge`; sem candidato, emitir apenas `workspace-not-found`.
3. Comprovar manifesto regular não-symlink, JSON único, nome e schema.
4. Se a versão for compatível, validar identidade, source e declaração de workflow.
5. Validar coleções independentes em ordem `product`, `decisions`, `policies`, `runs`.
6. Resolver provider e estrutura do workflow; então validar o path canônico de specs.
7. Ordenar problemas pela ordem pública abaixo, preservando a ordem fixa dos componentes repetidos.
8. Derivar status sem remover os fatos independentes já comprovados.

Ordem pública: `workspace-not-found`, `knowledge-invalid`, `manifest-invalid`,
`manifest-version-unsupported`, `source-invalid`, `required-directory-invalid`,
`workflow-pending`, `workflow-invalid`, `workflow-unknown-provider`.

### Dependency gates

- Sem root: omitir workspace, knowledge, source e workflow.
- Knowledge inválido: preservar apenas root; não ler manifesto ou coleções.
- Manifesto inválido: preservar root, knowledge e coleções comprováveis; omitir fatos derivados.
- Versão não suportada: omitir nome, source e workflow derivados.
- Source inválido: omitir somente source.
- Workflow desconhecido: preservar provider/state, omitir `specs_path`.
- Workflow pendente: warning; o path normativo ausente não é publicado como comprovado.
- Workflow inválido: error; specs só é publicado se seu diretório for comprovado. Specs inválido
  após materialização emite `required-directory-invalid` e `workflow-invalid`, nessa ordem pública.

## Determinism and compatibility

- O modelo não contém timestamp, versão do binário ou dados de runtime.
- Structs e slices substituem maps na fronteira JSON.
- Campos opcionais só são adicionados no schema v1; remoção, renomeação ou mudança semântica exige
  novo `schema_version`.
- Campos futuros desconhecidos do manifesto são ignorados, preservando compatibilidade v1.
