# Contract: `cerne context`

## Syntax

```text
cerne context
cerne context --json
cerne context --help
```

Nenhum argumento posicional, alias, combinação de flags ou formato adicional é aceito.

## Streams and exit status

| Caso | stdout | stderr | Status |
|---|---|---|---|
| contexto healthy | relatório | vazio | 0 |
| contexto warning | relatório | vazio | 0 |
| contexto invalid | relatório | vazio | 1 |
| `--help` | ajuda | vazio | 0 |
| uso inválido | vazio | erro e uso | 2 |

Uso inválido:

```text
erro: argumento inválido
uso: cerne context [--json]
```

## Human report

A saída humana projeta somente campos presentes no mesmo `ContextReport`. Rótulos, ordem e
vocabulário são estáveis; paths usam a representação nativa.

Exemplo saudável sem workflow:

```text
Workspace: exemplo
Status: saudável
Root: /work/exemplo

Knowledge: /work/exemplo/knowledge
Product: /work/exemplo/knowledge/product
Specs: /work/exemplo/knowledge/specs
Decisions: /work/exemplo/knowledge/decisions
Policies: /work/exemplo/knowledge/policies

Source: /work/exemplo/source
Localização: interno ao workspace

Workflow: não declarado
```

Para source externo, `Localização` é `externo ao workspace`. Estados declarados usam provider e um
dos termos `pendente`, `pronto`, `inválido` ou `provider desconhecido`.

Problemas aparecem depois dos fatos comprovados. A apresentação humana associa a cada código uma
causa e correção fixa em português; nunca imprime `error.Error()`, paths não comprovados ou texto
do provider.

| Código | Causa | Correção |
|---|---|---|
| `workspace-not-found` | Workspace Cerne não localizado | execute o comando dentro de um workspace Cerne |
| `manifest-invalid` | Manifesto ausente ou inválido | corrija ou restaure `knowledge/cerne.json` |
| `manifest-version-unsupported` | Versão do manifesto não suportada | use uma versão compatível do Cerne |
| `knowledge-invalid` | Knowledge ausente ou inseguro | restaure o diretório `knowledge` |
| `source-invalid` | Source ausente ou inseguro | corrija o caminho `source` no manifesto |
| `required-directory-invalid` | Diretório obrigatório ausente ou inválido | restaure o diretório indicado pelo componente |
| `workflow-pending` | Workflow ainda não materializado | execute `cerne workflow setup` quando o provider estiver disponível |
| `workflow-invalid` | Estrutura do workflow inválida ou parcial | corrija a estrutura antes de continuar |
| `workflow-unknown-provider` | Provider declarado não suportado | use `speckit` ou `openspec` no manifesto |

Exemplo não bloqueante:

```text
! Workflow: pendente
  Correção: execute cerne workflow setup quando o provider estiver disponível
```

Exemplo fora de workspace:

```text
Status: inválido

✗ Workspace: não localizado
  Correção: execute o comando dentro de um workspace Cerne
```

## Help contract

A ajuda documenta finalidade, sintaxe, descoberta ancestral, campos, `--json`, streams, status,
ausência de efeitos e exemplos. Ela não consulta o diretório atual nem qualquer workspace.

## Effects and authorization

O comando possui leitura estrutural local exclusiva. Não cria auditoria, cache ou instruções; não
modifica manifesto; não executa Git, provider ou agente; não consulta PATH, rede ou credenciais; não
lê conteúdo de knowledge/source. Nenhuma autorização adicional é solicitada.

## Compatibility

O comando é uma adição compatível. Texto humano, sintaxe, streams e status são contratos públicos.
O formato de automação normativo é o JSON descrito em
[context-report-schema.md](context-report-schema.md).
