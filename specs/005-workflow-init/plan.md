# Implementation Plan: Inicialização com Workflow SDD

**Branch**: `feat/workflow-init` | **Date**: 2026-08-04 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/005-workflow-init/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command; its definition describes the execution workflow.

## Summary

Estender `cerne init` com `--workflow speckit|openspec`, persistir a escolha opcional no manifesto
e inicializar o provider instalado dentro de `knowledge`, sem alterar `source`. O modo sem flag
permanece idêntico. A ausência do executável deixa o workflow pendente com aviso e status zero;
`cerne workflow setup` retoma a configuração. Providers entram por uma função-adaptador baseada
em processo local, com argumentos fixos, ambiente mínimo, telemetria desativada e sem shell. O
domínio aplica uma descrição genérica resolvida pelo adapter, valida seu layout, cria a auditoria
antes da execução e limpa somente a raiz descrita quando ela era ausente, preservando o workspace
base e arquivos preexistentes.

## Technical Context

**Language/Version**: Go 1.26.5, conforme `go.mod`

**Primary Dependencies**: Biblioteca padrão, Git local obrigatório e executável local opcional
`specify` ou `openspec`; nenhuma dependência Go nova

**Storage**: Sistema de arquivos local, `knowledge/cerne.json`, layouts do provider e registros em
`knowledge/runs`

**Testing**: `go test ./...`, callbacks falsos de domínio, subprocessos controlados de adaptador e
testes do despacho CLI sem instalações globais reais

**Target Platform**: Linux, Windows e macOS

**Project Type**: CLI de projeto único

**Performance Goals**: Criação e bootstrap concluídos em uma única invocação quando o provider está
disponível; `doctor` e detecção de estado sem varrer conteúdo integral do workspace

**Constraints**: Modo padrão byte a byte compatível, incluindo os cinco `.gitkeep` obrigatórios;
execução não interativa e sem shell; provider restrito por cwd, argumentos e ambiente; nenhuma
instalação, atualização, telemetria, credencial, agente específico, remoto ou mutação de `source`;
stdout/stderr e status estáveis; limpeza limitada a raízes de provider ausentes antes da tentativa;
auditoria obrigatória antes de executar, desconsiderando `runs/.gitkeep` na contagem de registros

**Scale/Scope**: Um workspace, um manifesto, um provider opcional e uma tentativa por invocação;
somente Spec Kit e OpenSpec, sem troca, conversão, sincronização ou execução do workflow

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Ownership and separation — PASS**: O provider recebe `knowledge` como diretório de trabalho e
  não recebe o caminho de `source`; layouts e auditoria ficam no repositório de conhecimento.
- **Neutral core — PASS**: O domínio trata `provider` como identificador opaco e recebe do adapter
  uma descrição genérica de specs path, owned root, marker e função de setup; nomes suportados,
  descoberta, argumentos, ambiente e processo externo ficam em `internal/workflowexec`.
- **Minimum context and audit — PASS**: Nenhum agente é chamado. Um registro iniciado é persistido
  antes do subprocesso e finalizado com resultado e timestamps; se a finalização falhar, permanece
  `started` como tentativa inconclusiva, sem saída externa integral.
- **Authorization and secrets — PASS**: A flag ou o comando de setup autoriza somente o bootstrap
  local. O adaptador usa a allowlist por plataforma definida em `research.md`, remove todo o restante,
  desativa telemetria e não instala dependências.
- **Portability — PASS**: Descoberta e execução nativas, sem shell, substituem scripts; testes cobrem
  Linux, Windows e macOS.
- **Testing — PASS**: O desenho prevê testes de domínio, contrato do adaptador, integração CLI,
  rollback limitado, auditoria, segredo, ausência de mutação e compatibilidade regressiva.
- **CLI compatibility and documentation — PASS**: O modo sem flag preserva seu contrato; novos
  usos, warning, status, manifesto, help e documentação são definidos como adição MINOR.
- **Simplicity — PASS**: O parser manual, callbacks, JSON, processo nativo e helpers existentes
  bastam; não há registry, framework CLI, SDK, instalação automática ou versão persistida.

**Pre-research gate**: PASS após ajustar a falha pós-execução para preservar workspace e auditoria.

**Post-design re-check**: PASS. Contratos e modelo mantêm knowledge/source separados, execução
local explicitamente autorizada, auditoria prévia, ambiente reduzido, limpeza conservadora,
adapters testáveis, portabilidade e compatibilidade do init legado.

## Project Structure

### Documentation (this feature)

```text
specs/005-workflow-init/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── init-workflow-command.md
│   ├── workflow-setup-command.md
│   └── workflow-audit-record.md
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
    ├── init_test.go
    ├── doctor.go
    ├── doctor_test.go
    ├── workflow.go
    └── workflow_test.go

README.md
README.pt-BR.md
README.es.md
.github/workflows/test.yml
```

**Structure Decision**: Manter parsing, help, streams e status em `cmd/cerne`; estender
`internal/workspace` somente com estado genérico, aplicação de uma descrição de layout, auditoria e
retomada; criar `internal/workflowexec` para resolver os dois identificadores em descrições e
funções específicas de execução. O limite usa uma função e uma struct de dados, como os adapters
Git atuais, sem interface, factory ou registry. O manifesto continua lido pelo helper existente e
`link` preserva campos desconhecidos.

## Complexity Tracking

Nenhuma violação constitucional exige justificativa.
