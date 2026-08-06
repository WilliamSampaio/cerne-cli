# Quickstart: validar `cerne context`

Este roteiro valida o contrato sem depender de rede, credenciais ou repositórios remotos. Os
detalhes normativos estão em [contracts/context-command.md](contracts/context-command.md) e
[contracts/context-report-schema.md](contracts/context-report-schema.md).

## Prerequisites

- Go 1.26.5
- checkout da branch `feat/context-command`
- Git pode existir na máquina, mas as consultas de contexto não devem invocá-lo

## Automated validation

```sh
gofmt -w cmd/cerne/main.go cmd/cerne/main_test.go internal/workspace/context.go internal/workspace/context_test.go internal/workflowexec/setup.go internal/workflowexec/setup_test.go
go test -count=1 ./...
go vet ./...
go build -o cerne ./cmd/cerne
```

Os testes devem usar `t.TempDir()` e fixtures locais, verificar stdout/stderr/status exatos e passar
com `PATH` vazio nos cenários de contexto.

## End-to-end scenarios

1. **Workspace saudável sem workflow**: executar na raiz, em `knowledge`, em `source` interno e em
   um descendente. Confirmar status 0, todos os paths canônicos e `not-declared`.
2. **JSON determinístico**: executar `cerne context --json` duas vezes sem alterar o fixture e
   comparar os bytes. Confirmar JSON válido, `schema_version: 1`, newline final e `problems: []`.
3. **Source externo**: configurar path relativo ou absoluto fora da raiz. Confirmar
   `inside_workspace: false`; ao executar dentro desse source externo, confirmar
   `workspace-not-found` e status 1.
4. **Workflow Spec Kit/OpenSpec**: cobrir raiz ausente (`pending`, status 0), estrutura completa
   (`ready`) e estrutura parcial, symlink ou `.git` aninhado (`invalid`, status 1). Remover PATH e
   confirmar resultado idêntico.
5. **Contexto parcial**: cobrir manifesto ausente/malformado/symlink, version 2, source inseguro,
   coleções ausentes e provider desconhecido. Confirmar que fatos independentes permanecem e fatos
   não comprovados são omitidos.
6. **Fronteira ancestral**: criar um workspace parcial dentro de um workspace válido e executar no
   descendente parcial. Confirmar que o ancestral válido não é selecionado.
7. **Uso e ajuda**: confirmar que `--help` funciona fora de workspace com status 0; opções extras
   deixam stdout vazio, usam stderr e status 2.
8. **Somente leitura**: fotografar nomes, tipos, conteúdo e mtimes relevantes antes/depois de todos
   os modos. Confirmar ausência de arquivos de audit/cache e zero chamadas aos fakes de Git,
   provider ou processo.

## Manual smoke test

Dentro de um workspace existente:

```sh
./cerne context
./cerne context --json
./cerne context --help
```

Inspecione se a saída humana permite localizar workspace, knowledge, source e workflow sem abrir o
manifesto. Analise a saída JSON com o parser padrão disponível no ambiente; não use comparação de
paths com `/` fixo em testes portáveis.

### Legibilidade humana (SC-008)

1. Mantenha `knowledge/cerne.json` fechado e execute `./cerne context` em um workspace saudável.
2. Inicie o cronômetro somente depois de a saída terminar de ser exibida.
3. Sem abrir outro arquivo, peça que a pessoa identifique workspace, knowledge, source, relação
   interna/externa do source e workflow.
4. Registre sucesso apenas se os cinco fatos estiverem corretos em menos de 10 segundos. Tempo de
   startup ou execução do binário não participa desta medição.

## Release verification

Antes da release minor, executar a suíte em Linux, Windows e macOS e atualizar `README.md`,
`README.pt-BR.md`, `README.es.md`, `CHANGELOG.md`, ajuda global e versão do binário. Nenhum workspace
existente deve exigir migração.
