# Quickstart: Validar `cerne doctor`

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

## Scenario 1: Healthy workspace

No diretório temporário:

```text
<binary-path> init exemplo
cd exemplo
<binary-path> doctor
```

Expected:

- dez linhas seguem [o contrato](contracts/doctor-command.md#report);
- o manifesto sem `version` é aprovado como versão 1 implícita;
- o resumo é `Workspace saudável`, stderr fica vazio e o status é `0`;
- nenhum arquivo ou estado Git muda durante `doctor`.

## Scenario 2: Valid name differs from root

Em um fixture descartável, altere `name` para outro nome portátil válido e execute o diagnóstico.

Expected: `Manifesto` usa `!` com orientação para alinhar os nomes, as dez linhas permanecem
presentes, o resumo é `Workspace com avisos` e o status é `0`. O comando não renomeia nem edita.

## Scenario 3: Explicit manifest version

Em fixtures descartáveis, confirme que `version: 1` é aprovado e que `"version": "1"` falha.

Expected: somente o inteiro JSON `1` equivale à versão suportada. Texto, valor fracionário, nulo ou
outra versão fazem `Versão do manifesto` usar `✗`, com resumo inválido e status `1`.

## Scenario 4: Missing required directory

Em uma cópia descartável do workspace, renomeie temporariamente `knowledge/product` e execute:

```text
<binary-path> doctor
```

Expected: `Diretórios obrigatórios` usa `✗`, os demais checks possíveis continuam presentes, o
resumo é `Workspace inválido` e o status é `1`. Restaure o fixture somente após a execução; o
comando não o corrige.

## Scenario 5: External and invalid manifest paths

Em um fixture descartável, vincule um repositório source externo válido e execute o diagnóstico;
depois substitua `source` por caminho inexistente ou link.

Expected: o source externo válido é aprovado sem alterações; caminho inexistente ou link faz
`Caminhos do manifesto` e checks dependentes falharem.

## Scenario 6: Repositories are not independent

Prepare um fixture em que `source` seja reconhecido apenas pelo repositório Git de `knowledge`, ou
em que ambos compartilhem o mesmo common dir, e execute o diagnóstico.

Expected: `Independência Git` usa `✗`, o isolamento aplicável não é omitido, não ocorre operação
Git modificadora ou remota e o status é `1`.

## Scenario 7: Git unavailable

Execute o binário com um `PATH` controlado que não contenha Git.

Expected: as dez linhas são emitidas; `Git` e as verificações dependentes usam `✗`; checks puramente
locais ainda apresentam seus resultados; resumo inválido e status `1`.

## Scenario 8: Permission uncertainty

Em fixture e ambiente descartáveis, use uma plataforma ou filesystem cuja consulta efetiva de
escrita seja declarada não suportada.

Expected: `Permissões` usa `!`, nenhum outro erro existe, o resumo é `Workspace com avisos` e o
status permanece `0`. Negação confirmada usa `✗`, nunca aviso.

## Scenario 9: Help and invalid usage

```text
<binary-path> doctor --help
<binary-path> doctor extra
```

Expected: ajuda usa stdout/status `0`; argumento excedente usa somente stderr/status `2`.

## Automated validation

Na raiz do Cerne:

```text
gofmt -w <changed-go-files>
go vet ./...
go test -count=1 ./...
```

Os testes comparam árvore, conteúdo, mtimes e estado Git antes/depois; `atime` é ignorado porque
pode ser atualizado pela própria leitura. A matriz existente repete a suíte em Linux, Windows e
macOS, sem rede ou credenciais.
