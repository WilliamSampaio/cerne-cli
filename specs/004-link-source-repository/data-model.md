# Data Model: Link de Repositório Source

## Workspace Location

Representa o workspace encontrado a partir do diretório atual.

| Campo | Tipo conceitual | Regras |
|---|---|---|
| `startPath` | caminho absoluto | Diretório onde `cerne link` foi invocado |
| `root` | caminho absoluto canônico | Ancestral mais próximo com `knowledge/cerne.json` |
| `knowledgePath` | caminho absoluto canônico | Sempre `root/knowledge` |
| `manifestPath` | caminho absoluto | Sempre `knowledgePath/cerne.json` |

Transições:

```text
Requested → Searching Ancestors → Found Workspace
Requested → Searching Ancestors → Not Found
```

## Project Manifest

Documento persistente atualizado pelo comando.

| Campo | Tipo conceitual | Regras |
|---|---|---|
| `name` | string | Obrigatório, válido e preservado |
| `source` | caminho | Obrigatório; atualizado para o source candidato normalizado |
| `version` | inteiro opcional | Ausente significa versão 1 implícita; versões incompatíveis bloqueiam link |
| outros campos | valores existentes | Devem ser preservados quando possível |

Transições:

```text
Loaded → Candidate Validated → Unchanged
Loaded → Candidate Validated → Updated Atomically
Loaded → Validation Failed → Preserved
Loaded → Write Failed → Preserved
```

## Repository Facts

Fatos locais usados para comparar knowledge, source atual e source candidato.

| Campo | Tipo conceitual | Regras |
|---|---|---|
| `requestedPath` | caminho absoluto | Caminho validado informado ao coletor |
| `worktreeRoot` | caminho absoluto canônico | Deve existir para source candidato |
| `commonDir` | caminho absoluto canônico | Usado para detectar identidade/compartilhamento Git |
| `isBare` | booleano | Source candidato deve ser `false` |
| `hasWorktree` | booleano | Source candidato deve ser `true` |

## Source Candidate

Repositório local informado pelo usuário.

| Campo | Tipo conceitual | Regras |
|---|---|---|
| `inputPath` | string | Argumento original do usuário |
| `resolvedPath` | caminho absoluto canônico | Resolvido a partir do diretório de execução |
| `repositoryFacts` | Repository Facts | Deve representar repositório Git local não-bare com worktree |
| `manifestSource` | caminho normalizado | Relativo ao knowledge quando portátil; absoluto quando necessário |

Validações:

- Deve existir.
- Deve ser diretório regular.
- Deve ser Git local válido com árvore de trabalho.
- Não pode ser bare.
- Pode ser worktree.
- Não pode ser o mesmo repositório do knowledge.
- Não pode conter knowledge nem estar contido por knowledge.

## Link Request

Entrada processada do comando.

| Campo | Tipo conceitual | Regras |
|---|---|---|
| `sourceInput` | string | Exatamente um caminho informado |
| `replace` | booleano | Verdadeiro somente com `--replace` |
| `startPath` | caminho | Diretório atual da invocação |

Validações de uso:

- Sem caminho: uso inválido.
- Argumento extra: uso inválido.
- Flag desconhecida: uso inválido.
- `--replace` sem caminho: uso inválido.
- `--replace` aceito apenas como flag da operação de link.

## Link Result

Resultado comunicado ao usuário.

| Campo | Tipo conceitual | Regras |
|---|---|---|
| `projectName` | string | Vem do manifesto |
| `previousSource` | caminho | Exibido quando existir |
| `newSource` | caminho | Caminho normalizado do source candidato |
| `changed` | booleano | `false` quando o source já era o configurado |
| `message` | texto | "manifesto atualizado" ou "nenhuma alteração necessária" |

## Failure Result

Falha anterior à conclusão do link.

| Campo | Tipo conceitual | Regras |
|---|---|---|
| `cause` | texto | Problema compreensível para o usuário |
| `path` | caminho opcional | Obrigatório quando a falha envolve recurso local específico |
| `correction` | texto opcional | Ação necessária quando aplicável |
| `exitStatus` | inteiro | `1` para falha operacional; `2` para uso inválido |

## Invariantes

- Um link bem-sucedido modifica somente o manifesto.
- Todas as falhas bloqueantes preservam o manifesto anterior.
- `--replace` autoriza apenas trocar a referência no manifesto.
- O source antigo e o novo source nunca são apagados, movidos, versionados, limpos ou acessados
  remotamente.
- Knowledge e source permanecem repositórios Git independentes.
