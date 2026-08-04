# Implementation Plan: Inicialização de Workspace

**Branch**: `feat/init-workspace` | **Date**: 2026-07-28 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/001-init-workspace/spec.md`

## Summary

Adicionar `cerne init <nome-do-projeto>` ao entrypoint existente. A implementação usará somente a
biblioteca padrão do Go e o executável Git local. O caso de uso de workspace validará o destino,
criará a estrutura e o manifesto e receberá a inicialização Git como uma função-adaptador. Em
falhas, removerá em ordem inversa somente os caminhos criados pela execução.

## Technical Context

**Language/Version**: Go 1.26.5, conforme `go.mod`

**Primary Dependencies**: Biblioteca padrão do Go e executável Git disponível em `PATH`

**Storage**: Sistema de arquivos local, dois repositórios Git e `knowledge/cerne.json`

**Testing**: `go test ./...`, diretórios temporários, Git local real e subprocesso do CLI

**Target Platform**: Linux, Windows e macOS

**Project Type**: CLI de projeto único

**Performance Goals**: Sem meta de latência adicional; uma execução cria um workspace local sem
rede e realiza exatamente duas inicializações Git

**Constraints**: Zero dependências Go novas; nenhuma rede; nenhum commit ou remoto; não seguir
links no destino; não substituir conteúdo; rollback limitado a caminhos criados pela execução

**Scale/Scope**: Um workspace por invocação, dois repositórios e seis artefatos de conhecimento

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Ownership and separation — PASS**: O conhecimento fica local; `knowledge/` e `source/` são
  irmãos com metadados Git próprios, e a raiz do workspace não é um repositório.
- **Neutral core — PASS**: Não há IA, rede ou fornecedor. Git entra por uma função-adaptador
  fornecida ao caso de uso, sem `os/exec` no domínio.
- **Minimum context and audit — PASS**: Nenhum agente é executado; `runs/` nasce vazio e nenhuma
  informação externa ao destino é coletada.
- **Authorization and secrets — PASS**: A invocação autoriza somente criação local. Não há
  operação sensível remota, credencial, log ou segredo.
- **Portability — PASS**: Nome validado por uma regra comum, caminhos via `filepath`, Git sem shell
  e testes em Linux, Windows e macOS.
- **Testing — PASS**: Há testes de domínio, rollback, contrato do adaptador Git e integração do CLI
  real em diretórios temporários.
- **CLI compatibility and documentation — PASS**: O contrato fixa sintaxe, streams, status,
  manifesto, estrutura e ajuda; README e help serão atualizados juntos.
- **Simplicity — PASS**: Parsing manual, biblioteca padrão e uma função-adaptador; nenhum framework,
  factory, container ou schema especulativo.

**Post-design re-check**: PASS. `data-model.md`, `contracts/init-command.md` e `quickstart.md`
preservam todos os gates acima e não introduzem exceções constitucionais.

## Project Structure

### Documentation (this feature)

```text
specs/001-init-workspace/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── init-command.md
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
│   └── init_test.go
└── workspace/
    ├── init.go
    └── init_test.go

.github/workflows/test.yml
README.md
```

**Structure Decision**: Manter o entrypoint existente em `cmd/cerne`, concentrar a regra e o
rollback em `internal/workspace` e isolar a única integração externa em `internal/gitexec`.
Testes permanecem ao lado dos pacotes; o teste do entrypoint cobre o fluxo completo do binário.

## Complexity Tracking

Nenhuma violação ou complexidade excepcional.
