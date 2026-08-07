# Implementation Plan: Descoberta de Agente para Spec Kit

**Branch**: `009-speckit-agent-discovery` | **Date**: 2026-08-06 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/009-speckit-agent-discovery/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command; its definition describes the execution workflow.

## Summary

Adicionar `--agent codex|claude` como opção local para `cerne init ... --workflow speckit` e
`cerne workflow setup`, mantendo `knowledge` como raiz real do Spec Kit e expondo uma ponte de
descoberta na raiz do workspace Cerne. Sem `--agent`, o comportamento existente permanece
compatível e `generic` continua sendo detalhe interno. A escolha do agente não entra em
`knowledge/cerne.json`; após restore, o usuário pode escolher outro agente local com
`cerne workflow setup --agent <agent>`.

## Technical Context

**Language/Version**: Go 1.26.5, conforme `go.mod`

**Primary Dependencies**: Biblioteca padrão, Git local obrigatório, executável local opcional
`specify`; nenhuma dependência Go nova

**Storage**: Sistema de arquivos local; `knowledge/cerne.json`; layout Spec Kit em `knowledge`;
ponte local na raiz do workspace (`.agents/skills` para Codex, `.claude/skills` para Claude);
auditorias existentes em `knowledge/runs`

**Testing**: `go test ./...`, testes de domínio com diretórios temporários, testes de contrato do
adapter `workflowexec` e integração CLI com provider controlado

**Target Platform**: Linux, Windows e macOS

**Project Type**: CLI de projeto único

**Performance Goals**: Preparar a descoberta local na mesma invocação de init/setup; operação
limitada ao conjunto fixo de comandos Spec Kit, sem varrer conteúdo de `knowledge` ou `source`

**Constraints**: Sem mudança no modo sem `--agent`; sem persistir agente no manifesto; sem tocar
`source`; sem instalar agentes; sem rede; sem credenciais; sem substituir artefatos de agente do
usuário fora do conjunto gerenciado; paths portáveis; stdout, stderr e status estáveis em português

**Scale/Scope**: Um workspace, um provider Spec Kit, um agente local escolhido por invocação; alvos
iniciais `codex` e `claude`; OpenSpec e diagnósticos automáticos de ponte ausente ficam fora desta
feature

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Ownership and separation — PASS**: Specs, templates e estado do provider continuam em
  `knowledge`; `source` não é alterado; a raiz do workspace recebe somente artefatos locais de
  descoberta.
- **Neutral core — PASS**: O manifesto persiste só `workflow.provider`; agente é escolha local e
  entra por descrição de adapter, não por campo permanente do domínio.
- **Minimum context and audit — PASS**: Nenhum conteúdo de `knowledge` é copiado para `source` ou
  para auditoria. Execução real do provider segue os registros existentes; ponte local é efeito de
  setup explicitamente solicitado.
- **Authorization and secrets — PASS**: `--agent` autoriza apenas criar/atualizar artefatos locais
  de descoberta para o agente informado. Não autoriza instalação, update, rede, credenciais,
  commits, push, merge, deploy ou publicação.
- **Portability — PASS**: A ponte usa arquivos regulares, paths relativos portáveis e geração sem
  symlink obrigatório.
- **Testing — PASS**: O desenho exige regressão do modo legado, cenários positivos para Codex e
  Claude, restore/setup, combinações inválidas, ausência/falha do provider e snapshots de `source`.
- **CLI compatibility and documentation — PASS**: `--agent` é aditivo; sem a flag, stdout, stderr,
  status e efeitos existentes permanecem. Contratos e help serão atualizados na mesma mudança.
- **Simplicity — PASS**: Reusa parser manual, callbacks, JSON e helpers existentes. Sem framework
  CLI, registry dinâmico, nova dependência ou persistência local extra.

**Pre-research gate**: PASS.

**Post-design re-check**: PASS. Os contratos mantêm agente como opção local, provider em
`knowledge`, `source` intocado, sem segredos, sem instalação de agentes e com comportamento
compatível para invocações antigas.

## Project Structure

### Documentation (this feature)

```text
specs/009-speckit-agent-discovery/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── init-agent-command.md
│   ├── workflow-setup-agent-command.md
│   └── local-discovery-bridge.md
└── tasks.md
```

### Source Code (repository root)

```text
cmd/cerne/
├── main.go
└── main_test.go

internal/
├── workflowexec/
│   ├── setup.go
│   └── setup_test.go
└── workspace/
    ├── init.go
    ├── workflow.go
    ├── workflow_test.go
    └── init_test.go

README.md
README.pt-BR.md
README.es.md
```

**Structure Decision**: Manter parsing, help e streams no CLI. Reusar `internal/workflowexec` para
resolver provider e agente em descrições específicas de execução/layout. Reusar `internal/workspace`
para aplicar a ponte local com validação de workspace, rollback conservador e preservação de
`source`. Criar arquivos novos só se isso reduzir tamanho dos arquivos existentes durante a
implementação.

## Complexity Tracking

Nenhuma violação constitucional exige justificativa.
