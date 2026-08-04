# Quickstart: Validar `cerne link`

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

## Scenario 1: Vincular source local válido

No diretório temporário:

```text
<binary-path> init exemplo
git init app
cd exemplo
<binary-path> link ../app --replace
```

Expected:

- stdout segue [o contrato](contracts/link-command.md#successful-update);
- `Projeto` usa o nome do manifesto;
- `Source anterior` mostra o source configurado antes do link;
- `Novo source` mostra o source normalizado;
- `knowledge/cerne.json` passa a apontar para o novo source;
- o repositório `app` não muda;
- stderr fica vazio e o status é `0`.

## Scenario 2: Caminho absoluto

Execute `cerne link` com caminho absoluto para um repositório Git local válido.

Expected:

- o caminho é aceito;
- o manifesto armazena caminho relativo quando isso for portátil;
- se relativo não for possível, o manifesto armazena caminho absoluto normalizado;
- nenhuma operação Git modificadora é executada.

## Scenario 3: Execução em subdiretório

Execute a partir de `knowledge/product` ou de outro subdiretório dentro do workspace:

```text
<binary-path> link ../../app --replace
```

Expected: o mesmo workspace ancestral é localizado e o caminho relativo informado é resolvido a
partir do diretório de execução.

## Scenario 4: Source já configurado

Execute novamente o link para o mesmo source já registrado.

Expected:

- stdout segue [o contrato](contracts/link-command.md#no-op);
- o status é `0`;
- o manifesto não é regravado;
- stderr fica vazio.

## Scenario 5: Substituição recusada sem `--replace`

Prepare dois repositórios source locais e tente trocar do primeiro para o segundo sem `--replace`.

Expected:

- stdout fica vazio;
- stderr informa que outro source já está configurado;
- stderr orienta usar `--replace`;
- status é `1`;
- o manifesto anterior permanece byte a byte.

## Scenario 6: Substituição explícita

Repita o cenário anterior com:

```text
<binary-path> link <novo-source> --replace
```

Expected:

- manifesto atualizado;
- source antigo preservado sem alterações;
- source novo preservado sem alterações;
- status `0`.

## Scenario 7: Worktree válido

Crie um repositório Git local com commit e adicione um worktree descartável. Execute:

```text
<binary-path> link <caminho-do-worktree> --replace
```

Expected: o worktree é aceito como source e nenhuma operação Git modificadora ocorre.

## Scenario 8: Bare repository recusado

Crie um repositório bare e execute:

```text
<binary-path> link <repo-bare> --replace
```

Expected: stdout vazio, stderr informa que bare repositories não são aceitos, status `1` e
manifesto preservado.

## Scenario 9: Caminho inválido

Execute com caminho inexistente e com caminho que aponta para arquivo regular.

Expected: stdout vazio, stderr identifica o caminho afetado, status `1` e manifesto preservado.

## Scenario 10: Independência entre knowledge e source

Tente vincular o próprio `knowledge`, um ancestral de `knowledge` ou um descendente que gere
aninhamento perigoso.

Expected: o comando recusa a operação, informa que knowledge e source devem permanecer
independentes e preserva o manifesto.

## Scenario 11: Manifesto inválido ou falha de escrita

Corrompa o manifesto ou torne o manifesto indisponível para escrita em fixture descartável.

Expected:

- manifesto inválido bloqueia antes de qualquer gravação;
- falha de escrita mantém o manifesto anterior válido;
- stderr informa causa, caminho e correção;
- status é `1`.

## Scenario 12: Ajuda e uso inválido

Execute:

```text
<binary-path> link --help
<binary-path> link
<binary-path> link <source> extra
<binary-path> link --replace
```

Expected:

- ajuda usa stdout e status `0`;
- usos inválidos usam stderr, incluem `uso: cerne link <caminho> [--replace]` e retornam status
  `2`.

## Automated validation

Na raiz do Cerne:

```text
gofmt -w <changed-go-files>
go vet ./...
go test -count=1 ./...
git diff --check
```

A matriz existente deve repetir a suíte em Linux, Windows e macOS.
