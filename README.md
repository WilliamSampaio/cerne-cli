# Cerne

Cerne é um CLI open source em Go, sob licença MIT, para administrar workspaces formados por dois
repositórios Git independentes: conhecimento e código-fonte.

## `cerne init`

Inicializa um workspace local:

```text
cerne init <project-name>
cerne init --help
```

`<project-name>` deve ter de 1 a 255 caracteres ASCII. O primeiro deve ser uma letra ou número; os
demais podem ser letras, números, `.`, `_` ou `-`. O nome não pode terminar em ponto nem usar nomes
reservados do Windows, inclusive antes de uma extensão: `CON`, `PRN`, `AUX`, `NUL`, `COM1`–`COM9`
ou `LPT1`–`LPT9`.

O destino é `<diretório-atual>/<project-name>` e precisa estar ausente ou ser um diretório regular
vazio, nunca um link. Nenhum arquivo existente é substituído ou removido.

```text
<project-name>/
├── knowledge/
│   ├── .git/
│   ├── cerne.json
│   ├── product/
│   ├── specs/
│   ├── decisions/
│   ├── policies/
│   └── runs/
└── source/
    └── .git/
```

Os dois diretórios são repositórios Git locais independentes, sem commit ou remoto. O manifesto
`knowledge/cerne.json` identifica o projeto e aponta `source` para `../source`. A raiz do workspace
não é inicializada como repositório.

O comando não usa rede, credenciais ou agentes; não clona, publica, faz push, merge ou deploy. A
invocação autoriza apenas a criação local acima. Se ocorrer uma falha parcial, o Cerne desfaz
somente os artefatos criados pela própria tentativa.

### Saída e status

Em sucesso, o status é `0`, stderr fica vazio e stdout contém:

```text
Workspace "<project-name>" criado.
Knowledge: <absolute-knowledge-path>
Source: <absolute-source-path>
```

Ajuda também usa stdout e status `0`. Falhas operacionais — destino inseguro, permissão, Git
indisponível ou erro de criação — usam stderr, status `1` e informam causa e correção. Uso ou nome
inválido usa stderr, status `2` e inclui `uso: cerne init <project-name>`. Não há prompt, cor ou
texto decorativo.

Exemplo:

```text
cerne init exemplo
```

## `cerne doctor`

Analisa o workspace Cerne no diretório atual e imprime um relatório estável, sem modificar
arquivos, manifesto ou repositórios:

```text
cerne doctor
cerne doctor --help
```

O relatório sempre contém dez verificações, nesta ordem: manifesto, repositório de conhecimento,
repositório de código-fonte, independência Git, isolamento de versionamento, caminhos do
manifesto, diretórios obrigatórios, Git, permissões e versão do manifesto.

Cada linha começa com `✓` para aprovado, `✗` para erro bloqueante ou `!` para aviso não
bloqueante. Erros e avisos incluem `correção:`. Ao final, o resumo é exatamente um destes textos:
`Workspace saudável`, `Workspace com avisos` ou `Workspace inválido`.

Status e streams: relatórios e ajuda usam stdout; uso inválido usa stderr e status `2`; erro
bloqueante no diagnóstico usa stdout e status `1`; workspace saudável ou somente com avisos usa
status `0`. Uma falha antes de iniciar o relatório usa stderr, status `1` e não imprime resumo.

O manifesto atual fica em `knowledge/cerne.json`. A ausência de `version` significa versão 1
implícita; quando o campo existe, somente o inteiro JSON `1` é aceito. `name` inválido é erro;
`name` válido diferente do nome da raiz gera aviso.

O comando é somente de leitura: não cria diretórios, não corrige problemas, não altera Git, não
usa remotos, GitHub, rede, credenciais ou agentes de IA. Quando a plataforma não permite confirmar
permissões efetivas com segurança, a verificação de permissões emite aviso em vez de aprovação.

Exemplo:

```text
cerne doctor
```

## `cerne status`

Apresenta o estado local do workspace Cerne a partir do diretório atual:

```text
cerne status
cerne status --help
```

O comando sobe pelos ancestrais até encontrar `knowledge/cerne.json`, carrega o manifesto e consulta
os repositórios `knowledge` e `source`. Para o projeto, exibe o nome e o caminho absoluto do
workspace. Para cada repositório, exibe caminho, branch, commit abreviado, estado e contagens de
arquivos modificados, em stage e não rastreados.

Um repositório sem alterações aparece como `Estado: limpo`. Qualquer modificação fora do stage,
alteração em stage ou arquivo não rastreado aparece como `Estado: alterações pendentes`; isso não é
erro e mantém status `0` quando a consulta for concluída. Em estados especiais, `Branch: detached
HEAD` indica HEAD destacado e `Commit: sem commits` indica repositório ainda sem commit.

Relatório e ajuda usam stdout. Uso inválido usa stderr e status `2`. Falhas operacionais — workspace
não localizado, manifesto ausente ou inválido, caminho inexistente, diretório sem Git ou falha de
consulta Git — usam stderr, status `1`, incluem o caminho afetado quando houver e uma orientação de
correção.

O comando é somente de leitura: não cria, corrige, modifica arquivos, altera stage, troca branch,
cria commit, executa reset, acessa remotos, usa rede, credenciais ou agentes de IA. Não há JSON,
watch, comparação com GitHub ou exibição de nomes de arquivos alterados nesta versão.

Exemplo:

```text
cerne status
```

## `cerne link`

Vincula o workspace atual a um repositório Git local existente como `source`:

```text
cerne link <caminho>
cerne link <caminho> --replace
cerne link --help
```

O caminho pode ser relativo ao diretório atual ou absoluto. Ele deve apontar para a raiz de um
repositório Git local com árvore de trabalho; worktrees válidos são aceitos e repositórios bare são
recusados. O comando localiza o workspace por ancestral, lê `knowledge/cerne.json`, valida o
manifesto, normaliza o novo caminho e grava somente o campo `source`. Quando possível, o manifesto
armazena o caminho relativo ao diretório `knowledge`.

Se o manifesto já aponta para outro source, a troca falha por padrão. Use `--replace` para autorizar
explicitamente a substituição da referência:

```text
cerne link ../geo-app --replace
```

Mesmo com `--replace`, o Cerne não copia, move, apaga, limpa, faz checkout, reset, add, commit,
fetch, pull ou push no source anterior ou no novo. Também não acessa remotos, rede, credenciais ou
agentes de IA. Source e knowledge precisam ser repositórios independentes e não podem estar
aninhados de forma perigosa.

Sucesso e ajuda usam stdout e status `0`. Se o source informado já estiver configurado, o comando
informa `Nenhuma alteração necessária.` e não regrava o manifesto. Uso inválido usa stderr, status
`2` e inclui `uso: cerne link <caminho> [--replace]`. Falhas operacionais usam stderr, status `1`,
incluem causa, caminho afetado quando houver e uma orientação de correção. A atualização do
manifesto é feita por arquivo temporário e substituição final.

Para compilar e testar:

```text
go build -o cerne ./cmd/cerne
go test ./...
```

Git deve estar disponível em `PATH`. O comportamento é suportado em Linux, Windows e macOS.
