# Data Model: Inicialização de Workspace

## Workspace

Agregado criado por uma invocação de `cerne init`.

| Campo | Tipo conceitual | Regras |
|---|---|---|
| `name` | string | Nome portátil validado pelo contrato do comando |
| `root` | caminho absoluto | `cwd/name`; ausente ou diretório regular vazio |
| `knowledge` | Repository | Sempre `root/knowledge`, papel `knowledge` |
| `source` | Repository | Sempre `root/source`, papel `source` |
| `manifest` | ProjectManifest | Sempre `knowledge/cerne.json` |

## Repository

| Campo | Tipo conceitual | Regras |
|---|---|---|
| `role` | `knowledge` ou `source` | Exatamente um de cada por workspace |
| `root` | caminho absoluto | Diretórios irmãos, nunca aninhados |
| `gitRoot` | caminho absoluto | Deve resolver para o próprio `root` |
| `remotes` | coleção | Vazia na criação |
| `commits` | contagem | Zero na criação |

O repositório `source` não possui arquivos fora dos metadados Git. O repositório `knowledge`
possui manifesto e os diretórios `product`, `specs`, `decisions`, `policies` e `runs`.

## ProjectManifest

Contrato persistido como objeto JSON em `knowledge/cerne.json`.

| Campo | Tipo | Regras |
|---|---|---|
| `name` | string | Igual ao nome validado do workspace |
| `source` | string | Literal relativo `../source` |

Nenhum campo adicional é obrigatório nesta feature.

## Initialization Result

Valor retornado ao CLI após sucesso.

| Campo | Tipo | Uso |
|---|---|---|
| `name` | string | Linha de resumo |
| `knowledgePath` | caminho absoluto | Linha `Knowledge` em stdout |
| `sourcePath` | caminho absoluto | Linha `Source` em stdout |

## Estado transitório de inicialização

O estado não é persistido; existe apenas durante a execução para controlar rollback.

```text
Requested
  → Validated
  → RootReady
  → KnowledgeReady
  → SourceReady
  → Complete
```

Qualquer erro após `RootReady` transita para `RollingBack` e depois `RolledBack`. O rollback remove
somente caminhos registrados como criados pela execução. Uma raiz preexistente nunca é removida.

## Invariantes

- Workspace completo contém exatamente dois repositórios Git administrados.
- A raiz do workspace não contém metadados Git criados pelo Cerne.
- `knowledge` e `source` não contêm um ao outro.
- O manifesto sempre aponta de `knowledge` para `../source`.
- Sucesso só é retornado após os dois repositórios e todos os artefatos existirem.
- Falha não substitui nem remove conteúdo anterior à invocação.
