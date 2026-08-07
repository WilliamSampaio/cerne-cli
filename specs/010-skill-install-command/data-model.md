# Data Model: Instalação de Skills Cerne

## SkillInstallRequest

Pedido explícito de instalação.

| Field | Type | Rules |
|-------|------|-------|
| `agent` | enum | Obrigatório; exatamente `codex` ou `claude` |
| `requested_at` | timestamp | Gerado no início da tentativa operacional |

Validation:

- Argumento ausente, repetido, extra, caixa diferente ou agente desconhecido é uso inválido.
- Uso inválido termina antes de resolver pacote, destino, auditoria ou filesystem mutável.

## SkillPackage

Artefato oficial versionado do repositório `cerne-skills`, entregue como pacote companheiro local
ou cacheado gerenciado pelo CLI.

| Field | Type | Rules |
|-------|------|-------|
| `name` | string | Deve ser `cerne-skills` |
| `version` | semver string | Obrigatório para saída, auditoria e ownership |
| `manifest` | SkillManifest | Deve ser válido antes da cópia |
| `root` | path | Raiz controlada do pacote |
| `source` | enum | `companion-cache` nesta primeira versão |

Validation:

- Pacote ausente, sem manifesto, malformado ou não oficial falha com status 1.
- Pacote companheiro ausente ou inacessível falha antes de alterar destinos de agente.
- Entradas do manifesto não podem escapar de `root`.
- Links simbólicos são recusados.

## SkillManifest

Contrato do pacote usado pelo instalador.

| Field | Type | Rules |
|-------|------|-------|
| `schemaVersion` | integer | Versão do manifesto suportada pelo CLI |
| `package` | string | Deve identificar `cerne-skills` |
| `version` | string | Versão do pacote |
| `contextSchema` | integer | Deve ser compatível com `cerne context --json` schema v1 |
| `skills` | list | Deve conter `cerne-context` |
| `adapters` | map | Deve conter entrada para o agente solicitado |

Relationships:

- `SkillPackage` contém exatamente um manifesto validado.
- `AgentInstallTarget` usa o agente solicitado e os destinos oficiais fixos desta versão.

## AgentInstallTarget

Destino oficial de usuário atual para um agente suportado.

| Field | Type | Rules |
|-------|------|-------|
| `agent` | enum | `codex` ou `claude` |
| `skill_name` | string | Deve ser `cerne-context` |
| `destination` | path | `~/.codex/skills/cerne-context` ou `~/.claude/skills/cerne-context`, resolvido dentro da home atual |
| `invocation_name` | string | Nome usado pelo agente para descobrir/invocar a skill |

Validation:

- O destino não pode ser arquivo regular, link simbólico ou path fora da home do usuário.
- A instalação não pode exigir permissão administrativa.

## ManagedInstallation

Estado de uma skill previamente instalada e gerenciada pelo Cerne.

| Field | Type | Rules |
|-------|------|-------|
| `agent` | enum | Agente da instalação |
| `skill_name` | string | `cerne-context` |
| `package_version` | string | Versão instalada |
| `files` | list | Arquivos pertencentes à instalação |
| `ownership_marker` | file | Marcador privado contendo agente, skill, versão do pacote e lista de paths relativos gerenciados |

State transitions:

```text
absent -> staging_validated -> installed
installed(same version) -> unchanged
installed(managed different version) -> staging_validated -> installed(package version)
unknown_existing -> refused
staging_validated -> failed -> previous_state_preserved
```

Validation:

- Mesma versão gerenciada é no-op idempotente.
- Atualização gerenciada em versão diferente acontece automaticamente e só substitui arquivos
  listados/provados como gerenciados.
- Conteúdo desconhecido bloqueia a operação antes de sobrescrever.
- O marcador de ownership contém somente metadados e paths relativos; não contém conteúdo da skill,
  secrets, variáveis de ambiente ou saída externa.

## SkillInstallAttempt

Registro auditável local de tentativa operacional.

| Field | Type | Rules |
|-------|------|-------|
| `schema_version` | integer | Versão do registro de auditoria |
| `operation` | string | `skill.install` |
| `agent` | string | Agente solicitado |
| `skill` | string | `cerne-context` quando conhecido |
| `package` | string | `cerne-skills` quando conhecido |
| `package_version` | string | Versão quando validada |
| `destination` | path | Destino resolvido, sem conteúdo da skill |
| `status` | enum | `started`, `succeeded`, `failed` |
| `error_code` | string | Código seguro para falhas |
| `started_at` | timestamp | Obrigatório |
| `finished_at` | timestamp | Obrigatório ao finalizar |

Validation:

- Uso inválido não cria registro.
- Falha de auditoria impede a operação sensível de continuar.
- O registro não contém conteúdo da skill, variáveis de ambiente, tokens, remotes ou saída externa
  bruta.
