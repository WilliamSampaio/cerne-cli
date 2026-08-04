# Quickstart: Validar `cerne init`

## Prerequisites

- Go 1.26.5 ou a versão declarada em `go.mod`.
- Git disponível em `PATH`.
- Um diretório temporário vazio fora do repositório do Cerne.

## Build

Na raiz do Cerne, compile `./cmd/cerne` para um caminho executável adequado ao sistema:

```text
go build -o <binary-path> ./cmd/cerne
```

Use `<binary-path>` nos cenários abaixo.

## Scenario 1: New workspace

Em um diretório temporário vazio:

```text
<binary-path> init exemplo
git -C exemplo/knowledge rev-parse --show-toplevel
git -C exemplo/source rev-parse --show-toplevel
git -C exemplo/source remote
git -C exemplo/source rev-list --all --count
```

Expected:

- o comando retorna `0` e stdout segue [o contrato](contracts/init-command.md#success);
- as duas consultas `--show-toplevel` retornam raízes diferentes;
- `remote` não lista entradas e a contagem de commits é `0`;
- `knowledge/cerne.json` contém `name: exemplo` e `source: ../source`;
- a árvore corresponde ao contrato e `source/` contém somente `.git`.

## Scenario 2: Existing empty directory

Crie `vazio/` sem entradas e execute:

```text
<binary-path> init vazio
```

Expected: status `0`, a raiz preexistente é preservada e recebe a estrutura completa.

## Scenario 3: Existing content is protected

Crie `ocupado/` com um arquivo sentinela e execute:

```text
<binary-path> init ocupado
```

Expected: status `1`, stdout vazio, stderr contém causa e correção, e o arquivo sentinela permanece
inalterado.

## Scenario 4: Invalid name

```text
<binary-path> init ../invalido
```

Expected: status `2`, nenhum destino é criado e stderr inclui a sintaxe correta.

## Scenario 5: Help

```text
<binary-path> init --help
```

Expected: status `0`, stderr vazio e ajuda completa em stdout.

## Automated validation

Na raiz do Cerne:

```text
go test ./...
```

O mesmo comando deve passar nos jobs Linux, Windows e macOS. Os testes usam diretórios temporários,
Git local e nenhum remoto real.
