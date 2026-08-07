# Data Model: Descoberta de Agente para Spec Kit

## Agent Target

Escolha local informada na invocação do CLI.

| Field | Type | Rules |
| --- | --- | --- |
| `name` | string | Obrigatório quando `--agent` é usado; valores aceitos nesta versão: `codex`, `claude` |
| `provider` | string | Deve acompanhar workflow `speckit`; `openspec` e providers desconhecidos recusam agente |
| `discovery_root` | path relativo | `codex` usa `.agents/skills`; `claude` usa `.claude/skills` |
| `integration_options` | string opcional | Codex usa `--skills`; Claude não exige opção pública nesta versão |

`generic` não é `Agent Target`; permanece mecanismo interno/legado do Spec Kit.

## Workflow Declaration

Intenção persistida em `knowledge/cerne.json`.

| Field | Type | Rules |
| --- | --- | --- |
| `provider` | string | Continua sendo o único campo aceito no objeto `workflow` |

Exemplo válido:

```json
{
  "name": "example",
  "source": "../source",
  "workflow": {
    "provider": "speckit"
  }
}
```

O manifesto não armazena agente, estado da ponte, instalação observada, versão do provider ou
preferência local.

## Spec Kit Integration Layout

Layout real mantido em `knowledge`.

| Agent | Provider integration in knowledge | Expected managed files |
| --- | --- | --- |
| none | `generic` | `knowledge/.specify/commands/speckit.*.md` |
| `codex` | `codex` beside `generic` | `knowledge/.agents/skills/speckit-*/SKILL.md` |
| `claude` | `claude` beside `generic` | `knowledge/.claude/skills/speckit-*/SKILL.md` |

O layout do workflow continua pronto quando `knowledge/.specify/init-options.json` existe e o
diretório canônico `knowledge/specs` é válido. Integração de agente em `knowledge` é requisito para
criar ponte local, mas não redefine o estado estrutural do workflow.

## Local Discovery Bridge

Artefatos gerenciados na raiz do workspace Cerne.

| Field | Type | Rules |
| --- | --- | --- |
| `workspace_root` | path | Raiz que contém `knowledge/cerne.json` |
| `agent` | Agent Target | Define path e formato público da ponte |
| `bridge_root` | path | `.agents/skills` para Codex; `.claude/skills` para Claude |
| `command_set` | lista | Conjunto fixo dos comandos Spec Kit suportados |
| `knowledge_root` | path relativo | Deve apontar para `knowledge` sem path absoluto persistido |
| `status` | enum | `absent`, `ready`, `partial`, `conflict` |

Comandos esperados nesta versão:

```text
speckit-analyze
speckit-checklist
speckit-clarify
speckit-constitution
speckit-converge
speckit-implement
speckit-plan
speckit-specify
speckit-tasks
speckit-taskstoissues
```

### Bridge state transitions

```text
absent ── setup --agent succeeds ──> ready
ready ── setup --agent same agent ──> ready
ready ── setup --agent other agent ─> ready for requested agent
partial/conflict ── setup --agent ──> ready or operational failure
```

O Cerne pode atualizar somente os arquivos que pertencem ao conjunto gerenciado do agente
solicitado. Arquivos alheios do usuário no mesmo diretório não fazem parte do modelo.

## Workflow Setup Result

Resultado transitório apresentado ao CLI.

| Field | Type | Meaning |
| --- | --- | --- |
| `provider` | string | `speckit` nesta feature quando agente é usado |
| `knowledge_path` | path | Diretório onde o provider opera |
| `workflow_state` | enum | `pending`, `configured`, `unchanged` |
| `agent` | string opcional | Agente solicitado localmente |
| `discovery_state` | enum opcional | `ready`, `unchanged`, `not-created` ou falha operacional |
| `audit_path` | path opcional | Registro criado quando houve subprocesso real do provider |

`workflow_state` e `discovery_state` são separados para permitir workflow pronto sem ponte local e
troca de agente depois de restore.
