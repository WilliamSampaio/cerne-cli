# Data Model: Inicialização com Workflow SDD

## Workflow Configuration

Preferência opcional persistida em `knowledge/cerne.json`.

| Field | Type | Rules |
| --- | --- | --- |
| `provider` | string | Obrigatório e não vazio quando `workflow` existe; o adapter desta versão resolve somente `speckit` ou `openspec` |

Exemplo:

```json
{
  "name": "example",
  "source": "../source",
  "workflow": {
    "provider": "speckit"
  }
}
```

Ausência de `workflow` representa o modo legado sem provider. O manifesto não armazena
disponibilidade, versão observada nem estado de materialização.

## Workflow Definition

Descrição genérica resolvida por `internal/workflowexec`. O domínio recebe os campos abaixo e não
contém switch, constantes de layout ou validação específica por provider.

| Provider | Executable | Canonical specs | Owned root | Persistent marker |
| --- | --- | --- | --- | --- |
| `speckit` | `specify` | `knowledge/specs` | `knowledge/.specify` | `knowledge/.specify/init-options.json` |
| `openspec` | `openspec` | `knowledge/openspec/specs` | `knowledge/openspec` | `knowledge/openspec/config.yaml` |

O adapter mantém o mapeamento fechado, redescobre o executável a cada operação e entrega também a
função de setup. Não há registro dinâmico, factory ou interface de provider.

## Knowledge Layout

Todos os modos preservam:

```text
knowledge/
├── cerne.json
├── product/
│   └── .gitkeep
├── decisions/
│   └── .gitkeep
├── policies/
│   └── .gitkeep
└── runs/
    └── .gitkeep
```

O modo legado acrescenta `specs/.gitkeep`. Spec Kit preserva `specs/.gitkeep` e acrescenta
`.specify/`. OpenSpec acrescenta `openspec/` e não cria `knowledge/specs`; diretórios nativos vazios
do provider não participam da validação persistente porque não sobrevivem ao versionamento Git.

### Layout state

| State | Meaning | Transition |
| --- | --- | --- |
| `undeclared` | Manifesto não possui workflow | Estado final legado; setup é recusado |
| `pending` | Provider declarado e owned root ausente | Pode executar setup quando o binário existir |
| `ready` | Marker válido existe dentro do owned root | Setup é no-op e doctor aprova |
| `partial` | Owned root existe sem marker válido | Setup recusa alteração; doctor bloqueia |
| `invalid` | Campo malformado, identificador não resolvido pelo adapter ou layout viola limites Git | Doctor bloqueia; nenhuma execução ocorre |

Disponibilidade do executável é uma observação separada. `ready` com executável ausente continua
materializado, mas recebe warning operacional no doctor.

## Workflow Setup Attempt

Um arquivo JSON criado em `knowledge/runs` antes de cada subprocesso real do provider. O
`runs/.gitkeep` comum não é um registro e MUST ser ignorado ao enumerar ou contar tentativas.

| Field | Type | Rules |
| --- | --- | --- |
| `kind` | string | Sempre `workflow-setup` |
| `provider` | string | `speckit` ou `openspec` |
| `executor` | string | Nome esperado do executável, sem path sensível |
| `operation` | string | `init` ou `setup` |
| `context` | string | Sempre `knowledge` |
| `authorization` | string | `--workflow` ou `workflow setup` |
| `status` | string | `started`, `succeeded` ou `failed` |
| `started_at` | string | Timestamp UTC RFC 3339 obrigatório |
| `finished_at` | string | Timestamp UTC RFC 3339 somente no estado final |
| `failure` | string | Categoria segura opcional; nunca saída integral do provider |

O nome do arquivo é único e não contém nome do projeto, path absoluto, token ou argumento externo.
O registro começa em `started`; após execução e validação, o mesmo arquivo passa atomicamente a
`succeeded` ou `failed`. Interrupção abrupta pode deixá-lo em `started`, evidenciando tentativa
inconclusiva.

### Attempt transitions

```text
created(started) ── provider+validation succeed ──> succeeded
        └──────── provider/validation/cleanup fail ──> failed
        └──────── audit finalization fails ──────────> started
        └──────── process interrupted ───────────────> started
```

Falha ao criar o registro impede executar o provider. Falha ao finalizá-lo impede declarar setup
concluído, mantém o registro `started` como tentativa inconclusiva e aciona limpeza conservadora
do owned root novo.

## Workflow Setup Result

Resultado transitório retornado ao CLI.

| Field | Type | Meaning |
| --- | --- | --- |
| `project_name` | string | Nome validado do manifesto |
| `knowledge_path` | path | Diretório canônico onde o provider opera |
| `provider` | string | Provider declarado |
| `state` | string | `pending`, `configured` ou `unchanged` |
| `audit_path` | path opcional | Registro criado quando houve subprocesso |

O resultado não persiste stdout, stderr, ambiente ou versão do provider.
