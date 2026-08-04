# Implementation Plan: Status do Workspace

**Branch**: `003-workspace-status` | **Date**: 2026-08-03 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/003-workspace-status/spec.md`

## Summary

Adicionar `cerne status` ao despacho manual existente. O comando localizará o workspace a partir
do diretório atual, carregará o manifesto, coletará um status transitório para `knowledge` e
`source` e renderizará uma saída textual estável. A consulta Git permanece em `internal/gitexec`,
o domínio monta dados reutilizáveis em `internal/workspace` e `cmd/cerne` fica responsável por
sintaxe, streams, ajuda e códigos de saída. Nenhuma etapa escreve arquivos, altera stage, troca
branch ou acessa remotos.

## Technical Context

**Language/Version**: Go 1.24.6, conforme `go.mod`

**Primary Dependencies**: Biblioteca padrão e executável Git local; nenhuma dependência nova

**Storage**: Sistema de arquivos local, `knowledge/cerne.json` e metadados Git locais; somente
leitura

**Testing**: `go test ./...`, testes table-driven, repositórios temporários, teste do binário real
e matriz CI existente

**Target Platform**: Linux, Windows e macOS

**Project Type**: CLI de projeto único

**Performance Goals**: Status completo de um workspace pequeno em até 5 segundos, sem rede

**Constraints**: Localização por ancestral mais próximo; saída textual estável; stdout para
relatório e ajuda; stderr para uso inválido e falhas; status `0` para consulta obtida, `1` para
falha operacional e `2` para uso inválido; nenhum shell, prompt, remoto, credencial, IA ou mutação

**Scale/Scope**: Um workspace por invocação, um manifesto, dois repositórios e três contagens Git
por repositório

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Ownership and separation — PASS**: O comando apresenta `knowledge` e `source` separadamente e
  não copia conteúdo entre repositórios.
- **Neutral core — PASS**: Não há IA, agente ou fornecedor. Git entra por adaptador local.
- **Minimum context and audit — PASS**: Nenhum agente é executado e nenhuma auditoria persistente é
  criada, pois a operação autorizada é apenas leitura local observável.
- **Authorization and secrets — PASS**: A invocação autoriza somente consulta. O relatório não deve
  expor conteúdo de arquivos, credenciais, remotos ou configuração privada.
- **Portability — PASS**: Caminhos usam semântica portátil; consultas Git locais e parsing de
  saída devem produzir significado equivalente em Linux, Windows e macOS.
- **Testing — PASS**: O plano cobre domínio, adaptador Git, CLI real, cenários negativos, ausência
  de mutação e estados Git especiais.
- **CLI compatibility and documentation — PASS**: Saída textual, labels, streams, ajuda e status
  são definidos em contrato e serão documentados junto ao comando.
- **Simplicity — PASS**: O despacho manual existente é preservado; não há framework CLI,
  interface especulativa, formato JSON ou dependência nova.

**Pre-research gate**: PASS. Não há violação constitucional ou esclarecimento pendente.

**Post-design re-check**: PASS. Os artefatos mantêm leitura exclusiva, adaptador Git substituível,
saída estável e testes nos três sistemas. A complexidade adicionada limita-se ao caso de uso, ao
adaptador Git de status e à renderização do comando público.

## Project Structure

### Documentation (this feature)

```text
specs/003-workspace-status/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── status-command.md
└── tasks.md
```

### Source Code (repository root)

```text
cmd/cerne/
├── main.go
└── main_test.go

internal/
├── gitexec/
│   ├── init.go
│   ├── init_test.go
│   ├── inspect.go
│   ├── inspect_test.go
│   ├── status.go
│   └── status_test.go
└── workspace/
    ├── init.go
    ├── init_test.go
    ├── doctor.go
    ├── doctor_test.go
    ├── status.go
    └── status_test.go

README.md
go.mod
go.sum
.github/workflows/test.yml
```

**Structure Decision**: Estender `internal/gitexec` com uma função de status Git local e
`internal/workspace` com localização do workspace e agregação do relatório. Se `doctor` e `status`
precisarem dos mesmos detalhes de manifesto ou path, mover helpers privados dentro de
`internal/workspace` em vez de criar uma camada nova. `cmd/cerne` continua pequeno: valida
argumentos, chama o caso de uso e renderiza stdout/stderr.

## Complexity Tracking

Nenhuma violação constitucional exige justificativa.

