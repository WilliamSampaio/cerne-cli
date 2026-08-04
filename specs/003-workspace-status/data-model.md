# Data Model: Status do Workspace

## Workspace Location

Representa o workspace localizado a partir do diretório atual.

| Campo | Tipo conceitual | Regras |
|---|---|---|
| `startPath` | caminho absoluto | Diretório onde o comando foi invocado |
| `root` | caminho absoluto canônico | Ancestral mais próximo que contém `knowledge/cerne.json` |
| `manifestPath` | caminho absoluto | Sempre `root/knowledge/cerne.json` |

Transição:

```text
Requested → Searching Ancestors → Found Workspace
Requested → Searching Ancestors → Not Found
```

## Workspace Status

Relatório transitório de uma execução bem-sucedida; nada é persistido.

| Campo | Tipo conceitual | Regras |
|---|---|---|
| `projectName` | string | Vem do manifesto válido |
| `root` | caminho absoluto | Raiz localizada |
| `repositories` | sequência | Exatamente `knowledge` e `source`, nessa ordem |

## Repository Status

Estado transitório de um repositório Git local.

| Campo | Tipo conceitual | Regras |
|---|---|---|
| `name` | `knowledge` ou `source` | Label público estável |
| `path` | caminho absoluto | Diretório do repositório |
| `branch` | string | Nome da branch ou `detached HEAD` |
| `commit` | string | Hash abreviado de 7 caracteres ou `sem commits` |
| `state` | `limpo` ou `alterações pendentes` | Derivado das contagens |
| `modifiedCount` | inteiro não negativo | Arquivos alterados fora do stage |
| `stagedCount` | inteiro não negativo | Arquivos adicionados/alterados/removidos no stage |
| `untrackedCount` | inteiro não negativo | Arquivos não rastreados |

Regra de estado:

```text
modifiedCount == 0 && stagedCount == 0 && untrackedCount == 0 → limpo
caso contrário → alterações pendentes
```

## Workspace Manifest

Documento usado para identificar o projeto e localizar `source`.

| Campo | Tipo conceitual | Regras |
|---|---|---|
| `name` | string | Obrigatório e válido para exibição do projeto |
| `source` | caminho relativo | Obrigatório, relativo a `knowledge`, existente e dentro do workspace |
| `version` | inteiro opcional | Semântica compatível com o manifesto atual |

## Failure Result

Falha anterior à produção do relatório.

| Campo | Tipo conceitual | Regras |
|---|---|---|
| `cause` | texto | Problema compreensível para o usuário |
| `path` | caminho opcional | Obrigatório quando a falha afeta recurso local específico |
| `correction` | texto | Orientação manual para corrigir |
| `exitStatus` | inteiro | `1` para falha operacional, `2` para uso inválido |

## Invariantes

- Um status bem-sucedido sempre contém exatamente dois repositórios.
- Alterações pendentes nunca tornam o comando inválido por si só.
- Falhas de workspace, manifesto, path ou consulta Git interrompem o relatório e usam stderr.
- Nenhum modelo contém operação de criação, correção, remoção, stage, commit, checkout ou remoto.

