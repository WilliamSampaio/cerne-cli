# Data Model: Restauração de Workspace

## Restore Request

Dado transitório criado pela CLI e descartado ao fim da invocação.

| Field | Rules |
| --- | --- |
| `knowledge_origin` | URL/path classificado pela allowlist vigente; nunca persistido ou exibido |
| `source_mode` | Exatamente `local` ou `clone` |
| `source_input` | Path local ou origem classificada; nunca entra na auditoria |
| `parent` | Diretório atual canônico onde o novo workspace será promovido |
| `audit_dir` | `<user-home>/.cerne/audit`; não configurável nesta versão |

Help e uso inválido não produzem request. Uma request só vira tentativa após classificação estática
e preflight de paths, antes de qualquer processo Git.

## Restore Attempt

Um JSON global e privado por tentativa, com filename opaco `restore-<id>.json`.

| Field | Type | Rules |
| --- | --- | --- |
| `kind` | string | Sempre `workspace-restore` |
| `executor` | string | Sempre `cerne` |
| `operation` | string | Sempre `restore` |
| `authorization` | string | `restore --source` ou `restore --clone`, sem valores |
| `source_mode` | string | `local` ou `clone` |
| `workspace_name` | string | Omitido até o manifesto ser validado |
| `status` | string | `started`, `succeeded` ou `failed` |
| `started_at` | string | UTC RFC 3339 obrigatório |
| `finished_at` | string | UTC RFC 3339 somente em estado final durável |
| `failure` | string | Categoria segura opcional |
| `phases.knowledge` | Phase | Fase obrigatória de clone/validação |
| `phases.source` | Phase | Fase obrigatória de associação ou clone/validação |

Não há campo de origem, host, fingerprint, remoto, argumento, ambiente ou output. Um arquivo cujo
último estado durável é `started` depois que o comando terminou é inconclusivo; `inconclusive` não é
gravado por uma finalização que falhou.

### Phase

| Field | Rules |
| --- | --- |
| `operation` | `clone` para knowledge; `link` ou `clone` para source |
| `status` | `pending`, `running`, `validating`, `succeeded` ou `failed` |
| `started_at` | Presente a partir de `running` |
| `finished_at` | Presente em `succeeded` ou `failed` |
| `failure` | Categoria Cerne opcional; nunca mensagem externa |

### State transitions

```text
attempt:   started ──────────────────────────────> succeeded
              └─────────────────────────────────> failed
              └── audit update/finalize failure -> inconclusive (derived)

knowledge: pending -> running -> validating -> succeeded
                        └──────────┴────────────> failed

source:    pending -> running -> validating -> succeeded
                        └──────────┴────────────> failed
```

Cada transição que autoriza um processo é persistida antes dele. Falha de persistência impede o
próximo processo. Retenção e remoção do JSON pertencem ao usuário.

## Restore Manifest

É o `knowledge/cerne.json` clonado, no schema vigente.

| Field | Restore behavior |
| --- | --- |
| `name` | Obrigatório, string e nome portátil válido; define o destino final |
| `source` | Obrigatório; preservado no clone ou substituído no modo local |
| `version` | Ausente significa 1 implícito; somente inteiro JSON `1` é suportado |
| `workflow` | Opcional; provider conhecido e layout não parcial, sem execução |
| campos desconhecidos | Preservados; source local muda apenas a chave `source` |

`cerne.json` precisa ser arquivo regular e não symlink. Nenhuma origem é acrescentada ao schema.

## Source Selection

### Local

Working tree Git non-bare já existente, informado pela raiz e preservado byte a byte. Seus fatos Git
são comparados antes/depois e ele não pode sobrepor knowledge, destino, staging, pai ou audit.

Se necessário, `manifest.source` é recalculado em relação ao futuro `knowledge` e escrito
atomicamente no manifesto staged. O resultado informa se a referência mudou.

### Clone

A referência restaurada deve ter forma canônica `../<segmento-portátil>[/<segmento-portátil>...]`,
usar `/`, resolver dentro do workspace e fora de knowledge. Absoluto, volume, backslash, segmento
vazio/ponto, traversal adicional, nome reservado ou target existente/symlink é recusado.

O clone ocupa exatamente o target declarado e mantém `origin`, histórico e checkout padrão.

## Restore Staging

Diretório regular privado, imprevisível e irmão do destino final.

```text
absent -> private -> knowledge-ready -> source-ready -> validated -> promoted -> absent(staging)
             └──────────── any failure ───────────────> removed
promoted ── audit finalization failure ───────────────> root removed
```

Ownership é demonstrado por pai, prefixo, tipo regular e identidade do filesystem. Alvo ambíguo não
é removido. O destino final precisa estar ausente e a promoção nunca o substitui.

## Restored Workspace

| Field | Meaning |
| --- | --- |
| `name` | Nome validado do manifesto |
| `root` | `<parent>/<name>` promovido |
| `knowledge_path` | `<root>/knowledge`, Git root clonado |
| `source_path` | Path local autorizado ou target clonado declarado |
| `source_mode` | `local` ou `clone` |
| `manifest_changed` | Verdadeiro somente se source local substituiu a referência |

Invariantes finais: root não é repositório Git; knowledge/source são roots Git independentes;
manifesto, versão, workflow e diretórios são válidos; nenhum provider foi executado.
