# Data Model: Seleção de Source no Init

## Source Selection

Escolha transitória da invocação; não adiciona campo novo ao manifesto.

| Mode | Input | Materialized source | Manifest source |
| --- | --- | --- | --- |
| `empty` | nenhum | `workspace/source`, Git vazio | `../source` |
| `local` | path local | working tree existente, fora do workspace | path relativo portátil ou absoluto limpo |
| `clone` | localização permitida | clone privado promovido para `workspace/source` | `../source` |

### Validation

- Exatamente um mode; ausência de flag seleciona `empty`.
- `local` exige path existente, diretório, raiz Git non-bare e separado do futuro knowledge.
- `clone` exige localização sintaticamente permitida e sem credencial, query ou fragmento.
- O destino do workspace permanece ausente ou diretório regular vazio.

## Workspace Manifest

O schema continua versão 1 e sem informação sobre modo ou origem:

```json
{
  "name": "example",
  "source": "../source"
}
```

Para source local externo, `source` usa a mesma serialização de `cerne link`. A escolha transitória
não é necessária depois que o manifesto resolve o working tree.

## Clone Origin

Dado transitório usado somente pelo adapter.

| Field | Persistence | Rules |
| --- | --- | --- |
| raw location | nenhuma pelo Cerne | Nunca entra em manifesto, auditoria ou diagnóstico |
| transport | auditoria | `local`, `file`, `https` ou `ssh` |
| fingerprint | auditoria | SHA-256 hexadecimal da localização exata |

O clone bem-sucedido mantém a localização em `source/.git/config` como parte do remoto `origin`
criado pelo Git. Localizações com credenciais embutidas são recusadas antes desse efeito.

## Clone Staging

Diretório transitório criado com nome imprevisível, acesso restrito e no mesmo filesystem do
workspace. Ele é o único alvo do processo Git e da limpeza automática.

```text
absent ── create private staging ──> cloning ── validate ──> ready
ready ── promote without replacement ──> source
  └──── failure ── remove owned staging ──> absent
```

Se `source` aparecer antes da promoção, a operação falha sem substituí-lo ou removê-lo. Depois da
promoção, o source validado não volta a ser alvo de limpeza automática.

## Clone Attempt

Arquivo fixo `knowledge/runs/source-clone.json`, criado antes do subprocesso real.

| Field | Type | Rules |
| --- | --- | --- |
| `kind` | string | Sempre `source-clone` |
| `executor` | string | `git` |
| `operation` | string | `clone` |
| `project` | string | Nome validado do projeto |
| `destination` | string | Sempre `../source` |
| `origin_transport` | string | Transporte permitido, sem host/path |
| `origin_fingerprint` | string | SHA-256 hexadecimal |
| `authorization` | string | Sempre `--clone` |
| `status` | string | `started`, `succeeded` ou `failed` |
| `started_at` | string | UTC RFC 3339 obrigatório |
| `finished_at` | string | UTC RFC 3339 somente no estado final |
| `failure` | string | Categoria segura opcional, sem output Git |

`runs/.gitkeep` não é tentativa e não participa de contagens.

### Attempt transitions

```text
created(started) ── clone+validation+promotion succeed ──> succeeded
        ├──────── clone/validation/promotion fail ──────> failed
        ├──────── audit finalization fails ─────────────> started
        └──────── process interrupted ──────────────────> started
```

Falha ao criar `started` impede o clone e permite rollback integral. Depois de `started`, qualquer
falha preserva knowledge/auditoria e remove somente o staging privado. Se a limpeza falhar, o
registro final usa categoria segura de cleanup e o diagnóstico não afirma que o staging foi
removido. Se a finalização falhar depois da promoção, o source validado permanece e o registro
`started` sinaliza resultado inconclusivo.

## Initialization Result

Resultado transitório entregue ao CLI.

| Field | Meaning |
| --- | --- |
| `project_name` | Nome validado |
| `knowledge_path` | Knowledge criado |
| `source_path` | Working tree local ou interno |
| `source_mode` | `empty`, `local` ou `clone` |
| `audit_path` | Presente somente quando houve clone real |

O resultado não contém a origem do clone.
