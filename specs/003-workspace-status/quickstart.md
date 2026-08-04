# Quickstart: Validar `cerne status`

## Prerequisites

- Go na versão declarada em `go.mod`.
- Git disponível em `PATH`.
- Um diretório temporário fora do repositório do Cerne.

## Build

Na raiz do Cerne:

```text
go build -o <binary-path> ./cmd/cerne
```

Use um caminho de binário adequado ao sistema atual.

## Scenario 1: Workspace limpo

No diretório temporário:

```text
<binary-path> init exemplo
cd exemplo
git -C knowledge add .
git -C knowledge commit -m "init knowledge"
git -C source commit --allow-empty -m "init source"
<binary-path> status
```

Expected:

- stdout segue [o contrato](contracts/status-command.md#report);
- `Projeto` usa o `name` do manifesto;
- `Workspace` usa caminho absoluto;
- `Knowledge` e `Source` exibem `Estado: limpo`;
- as três contagens de cada repositório são `0`;
- stderr fica vazio e o status é `0`.

## Scenario 2: Execução em subdiretório

Execute a partir de `knowledge/product`, `knowledge/specs` e de um subdiretório dentro de
`source`.

Expected: o mesmo workspace ancestral é localizado, o relatório permanece completo e o status é
`0`.

## Scenario 3: Alterações pendentes separadas

Em fixture descartável, crie:

- um arquivo modificado fora do stage;
- um arquivo em stage;
- um arquivo não rastreado.

Execute:

```text
<binary-path> status
```

Expected: o repositório afetado usa `Estado: alterações pendentes`, status `0`, e as contagens
`Modificados`, `Em stage` e `Não rastreados` refletem categorias separadas.

## Scenario 4: Detached HEAD

Em fixture descartável com commit, coloque um dos repositórios em detached HEAD e execute o status.

Expected: `Branch: detached HEAD`, commit abreviado presente, stderr vazio e status `0`.

## Scenario 5: Repositório sem commits

Use o workspace recém-criado por `cerne init`, sem criar commits.

Expected: `Commit: sem commits` para os repositórios sem HEAD existente, sem erro. O repositório
`knowledge` pode exibir alterações pendentes porque o manifesto inicial ainda é um arquivo não
rastreado até o usuário versioná-lo.

## Scenario 6: Workspace não localizado

Execute em um diretório temporário que não contém ancestral Cerne.

Expected: stdout vazio; stderr informa que o workspace não foi localizado e inclui o diretório de
partida; status `1`.

## Scenario 7: Manifesto ou caminho inválido

Em fixture descartável, remova ou corrompa `knowledge/cerne.json`, ou altere `source` para um
caminho inexistente.

Expected: stdout vazio; stderr identifica o manifesto ou caminho afetado; status `1`; o comando
não corrige o fixture.

## Scenario 8: Diretório esperado não é Git

Em fixture descartável, remova os metadados Git de `knowledge` ou `source` e execute o status.

Expected: stdout vazio; stderr identifica o repositório afetado; status `1`; nenhuma operação Git
modificadora é executada.

## Scenario 9: Leitura exclusiva, ajuda e uso inválido

Compare snapshot lógico do workspace antes/depois de um status bem-sucedido. Execute também:

```text
<binary-path> status --help
<binary-path> status extra
```

Expected:

- nenhum arquivo, conteúdo, stage, branch, commit, remoto ou configuração muda;
- ajuda usa stdout e status `0`;
- argumento excedente usa somente stderr, inclui `uso: cerne status` e retorna status `2`;
- todos os cenários completam em até 5 segundos em workspace pequeno.

## Automated validation

Na raiz do Cerne:

```text
gofmt -w <changed-go-files>
go vet ./...
go test -count=1 ./...
git diff --check
```

A matriz existente deve repetir a suíte em Linux, Windows e macOS.
