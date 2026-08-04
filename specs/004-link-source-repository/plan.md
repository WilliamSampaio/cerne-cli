# Implementation Plan: Link de Repositório Source

**Branch**: `004-link-source-repository` | **Date**: 2026-08-04 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/004-link-source-repository/spec.md`

## Summary

Adicionar `cerne link <caminho> [--replace]` ao CLI existente para atualizar o `source` registrado
no manifesto de um workspace Cerne. O comando localizará o workspace a partir do diretório atual,
validará o manifesto, validará o repositório source candidato como Git local com árvore de trabalho,
recusará bare repositories e sobreposições perigosas, e gravará o manifesto de forma atômica. A
operação não copia, move, apaga, altera Git, acessa remotos ou modifica o source antigo ou novo.

## Technical Context

**Language/Version**: Go 1.24.6, conforme `go.mod`

**Primary Dependencies**: Biblioteca padrão e executável Git local; nenhuma dependência nova

**Storage**: Sistema de arquivos local e `knowledge/cerne.json`; atualização atômica do manifesto

**Testing**: `go test ./...`, testes table-driven, repositórios temporários, worktrees locais e
teste do binário real

**Target Platform**: Linux, Windows e macOS

**Project Type**: CLI de projeto único

**Performance Goals**: Link completo de um workspace pequeno em até 5 segundos, sem rede

**Constraints**: Localização por ancestral; validações antes de gravar; gravação atômica; saída
textual estável; stdout para sucesso/ajuda; stderr para uso inválido/falhas; status `0` para
sucesso ou sem alteração, `1` para falha operacional e `2` para uso inválido; nenhum shell, remoto,
credencial, IA ou operação Git modificadora

**Scale/Scope**: Um workspace por invocação, um manifesto, um knowledge e um source candidato

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Ownership and separation — PASS**: O comando altera somente a referência do manifesto e
  preserva repositórios knowledge/source separados, sem copiar conteúdo entre eles.
- **Neutral core — PASS**: Não há IA, agente ou fornecedor. Git local entra por adaptador.
- **Minimum context and audit — PASS**: Nenhum agente recebe contexto; a saída do comando comunica
  projeto, source anterior, novo source e resultado para auditoria humana imediata.
- **Authorization and secrets — PASS**: `--replace` é a autorização explícita para trocar a
  referência; a operação não manipula remotos, credenciais, push, merge, publicação ou deploy.
- **Portability — PASS**: O plano cobre normalização de paths, volumes distintos, aliases do
  filesystem, worktrees e atualização atômica em Linux, Windows e macOS.
- **Testing — PASS**: O plano exige testes de domínio, adaptador Git, integração CLI, cenários
  negativos, ausência de mutação e matriz multiplataforma.
- **CLI compatibility and documentation — PASS**: Sintaxe, flags, streams, status, mensagens,
  ajuda e documentação serão tratados como contrato público.
- **Simplicity — PASS**: O despacho manual existente será preservado; não há framework CLI,
  dependência nova, formato JSON, remoto ou camada especulativa.

**Pre-research gate**: PASS. Não há violação constitucional ou esclarecimento pendente.

**Post-design re-check**: PASS. Os artefatos mantêm separação de repositórios, autorização
explícita para substituição, escrita atômica restrita ao manifesto, consultas Git locais
read-only, testes e documentação.

## Project Structure

### Documentation (this feature)

```text
specs/004-link-source-repository/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── link-command.md
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
│   ├── inspect.go
│   ├── status.go
│   ├── link.go
│   └── *_test.go
└── workspace/
    ├── init.go
    ├── doctor.go
    ├── status.go
    ├── link.go
    └── *_test.go

README.md
go.mod
go.sum
.github/workflows/test.yml
```

**Structure Decision**: Estender `internal/workspace` com o caso de uso de link e helpers de
manifesto/path já existentes quando possível. Estender `internal/gitexec` apenas com consultas Git
locais necessárias para validar source, bare repository, worktree e identidade dos repositórios.
`cmd/cerne` continua responsável por argumentos, help, streams, status e renderização.

## Complexity Tracking

Nenhuma violação constitucional exige justificativa.
