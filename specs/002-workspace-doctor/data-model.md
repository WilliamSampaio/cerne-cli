# Data Model: Diagnóstico de Workspace

## Workspace Diagnostic

Agregado transitório produzido por uma execução; nada é persistido.

| Campo | Tipo conceitual | Regras |
|---|---|---|
| `root` | caminho absoluto canônico | Diretório atual; nunca descoberto em ancestral |
| `checks` | sequência de `CheckResult` | Exatamente 10, na ordem pública |
| `status` | `healthy`, `warning` ou `invalid` | Derivado pela maior severidade |

Transição:

```text
Requested → Collecting Facts → Classifying Checks → Complete
```

Falhas de dependência não interrompem a transição; preenchem os checks afetados com erro.

## Check Result

| Campo | Tipo | Regras |
|---|---|---|
| `id` | identificador estável | Um dos dez IDs abaixo |
| `label` | texto estável | Nome público exibido após o símbolo |
| `severity` | `pass`, `warning` ou `error` | Exatamente uma |
| `detail` | texto | Não vazio; não expõe conteúdo privado |
| `correction` | texto opcional | Obrigatório para aviso e erro |

Ordem e IDs:

1. `manifest`
2. `knowledge`
3. `source`
4. `git-independence`
5. `versioning-isolation`
6. `manifest-paths`
7. `knowledge-directories`
8. `git-available`
9. `permissions`
10. `manifest-version`

Precedência do agregado: qualquer `error` → `invalid`; senão qualquer `warning` → `warning`;
senão → `healthy`.

## Workspace Manifest

Representação interna lida de `knowledge/cerne.json`.

| Campo | Tipo | Regras |
|---|---|---|
| `name` | string | Obrigatório e portátil; divergência válida da raiz gera aviso |
| `source` | caminho relativo ou absoluto | Obrigatório; relativo a `knowledge` quando possível; pode apontar para repositório externo |
| `version` | inteiro JSON opcional | Ausente significa 1; explícito aceita somente 1 |

O documento contém um único objeto JSON. Dados posteriores ao objeto, tipos incorretos, caminho
inexistente ou link invalidam o manifesto ou seus caminhos. Campos desconhecidos são tolerados
para evolução compatível. Um `name` portátil que difere do basename da raiz mantém o
manifesto válido, mas torna seu check um aviso com orientação para alinhar os nomes.

## Repository Facts

Fatos retornados pelo adaptador Git para cada caminho esperado.

| Campo | Tipo | Regras |
|---|---|---|
| `requestedRoot` | caminho canônico | `knowledge` ou `source` esperado |
| `worktreeRoot` | caminho canônico | Deve ser o mesmo recurso que `requestedRoot` |
| `commonDir` | caminho canônico | Deve diferir do outro repositório |

As duas raízes precisam ser distintas e nenhuma pode conter a outra. Um top-level ancestral não
prova repositório próprio. Common dir compartilhado identifica linked worktrees do mesmo
repositório e falha a independência.

## Access Assessment

| Campo | Tipo | Regras |
|---|---|---|
| `path` | caminho | Recurso obrigatório avaliado |
| `capabilities` | conjunto | Leitura; escrita; travessia para diretórios |
| `outcome` | `allowed`, `denied` ou `unknown` | `denied` é erro; `unknown` é aviso |
| `reason` | texto sanitizado | Explica negação ou limitação, sem ACL completa |

O diagnóstico agrega manifesto, raízes dos repositórios e cinco diretórios obrigatórios. Todos
permitidos aprovam o check; qualquer negação o torna erro; ausência de negação com ao menos um
resultado desconhecido o torna aviso.

## Invariantes

- Um relatório completo sempre contém dez checks e um único status.
- Erro prevalece sobre aviso; aviso prevalece sobre aprovação.
- Sucesso nunca é emitido para fato não verificado.
- Manifesto legado e `version: 1` representam a mesma versão suportada.
- `name` válido diferente da raiz produz aviso; `name` inválido produz erro.
- Nenhum modelo possui operação de criação, atualização, correção ou remoção.
