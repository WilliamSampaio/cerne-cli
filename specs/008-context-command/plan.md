# Implementation Plan: Contexto Estrutural do Workspace

**Branch**: `feat/context-command` | **Date**: 2026-08-06 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/008-context-command/spec.md`

## Summary

Adicionar `cerne context` e `cerne context --json` como uma consulta estrutural, determinística e
estritamente somente leitura. O domínio produzirá um relatório parcial próprio a partir do
filesystem e do manifesto, reutilizando apenas parsing e helpers puros existentes. A saída JSON
usará structs tipadas e `encoding/json`; a saída humana projetará os mesmos fatos em português.
O comando não chamará `doctor`, Git, providers, rede ou auditoria.

## Technical Context

**Language/Version**: Go 1.26.5

**Primary Dependencies**: Biblioteca padrão do Go; dependências existentes permanecem inalteradas

**Storage**: Filesystem e `knowledge/cerne.json` existentes, somente leitura; nenhuma persistência

**Testing**: `testing`, `t.TempDir()`, testes unitários de domínio/adaptador e integração do binário

**Target Platform**: Linux, Windows e macOS

**Project Type**: CLI Go de projeto único

**Performance Goals**: Consulta limitada aos ancestrais e à pequena raiz estrutural do workflow,
sem varrer conteúdo de knowledge/source. SC-008 mede legibilidade humana depois de a saída estar
completa, não a duração de execução do comando

**Constraints**: Sem processos externos, Git, rede, credenciais, timestamps, cache, escrita ou
dependência de `PATH`; JSON byte a byte estável; paths absolutos, canônicos e nativos

**Scale/Scope**: Um workspace, um source, quatro coleções públicas, dois providers conhecidos,
cinco estados de workflow e nove códigos públicos de problema

## Constitution Check

*GATE: aprovado antes da pesquisa e reaprovado após o design da Fase 1.*

- **Ownership and separation — PASS**: a consulta não copia, persiste nem modifica knowledge ou
  source e não mistura seus ciclos de vida.
- **Neutral core — PASS**: o domínio expõe fatos neutros; Codex, Claude, prompts e seleção de
  contexto permanecem no futuro repositório de skills.
- **Minimum context and audit — PASS**: somente metadados estruturais solicitados são lidos. Não há
  execução automatizada nem operação sensível, portanto nenhuma auditoria é criada; consumidores
  continuam responsáveis por auditar ações posteriores.
- **Authorization and secrets — PASS**: a invocação autoriza apenas leitura local e proíbe origem,
  remoto, environment, credenciais e conteúdo nas saídas.
- **Portability — PASS**: `filepath`, `os` e `encoding/json` fornecem paths e serialização portáveis;
  testes evitam expectativas dependentes de separador ou caixa.
- **Testing — PASS**: regras do relatório, descrição estática de providers e contratos críticos do
  CLI terão testes determinísticos, negativos e de ausência de efeitos.
- **CLI compatibility and documentation — PASS**: o comando é aditivo, possui contrato explícito
  de argumentos, streams, status, JSON, ajuda e documentação nos três READMEs e changelog.
- **Simplicity — PASS**: nenhum pacote ou dependency novo; helpers puros existentes são reutilizados
  e o modelo de `doctor` não é adaptado artificialmente para um contrato diferente.

## Phase 0: Research Outcome

As decisões e alternativas estão consolidadas em [research.md](research.md). Não restam itens
`NEEDS CLARIFICATION`.

## Phase 1: Design Outcome

- Entidades, dependências de validação e estados: [data-model.md](data-model.md)
- Contrato do comando e da saída humana: [contracts/context-command.md](contracts/context-command.md)
- Schema JSON público v1: [contracts/context-report-schema.md](contracts/context-report-schema.md)
- Validação executável de ponta a ponta: [quickstart.md](quickstart.md)

### Post-design Constitution Check

**PASS**. O design não adiciona dependência, storage, interface especulativa ou efeito colateral.
A única separação nova é necessária: descrição estática de provider sem `exec.LookPath`, reutilizada
pelo resolver operacional atual. O relatório de contexto possui política própria porque converter
`Doctor` violaria os requisitos de neutralidade, contexto mínimo e ausência de processos.

## Project Structure

### Documentation (this feature)

```text
specs/008-context-command/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── context-command.md
│   └── context-report-schema.md
└── tasks.md              # criado posteriormente por /speckit-tasks
```

### Source Code (repository root)

```text
cmd/cerne/
├── main.go               # parsing, help e renderização CLI
└── main_test.go          # contratos do binário

internal/workspace/
├── context.go            # descoberta, modelo e diagnóstico estrutural
├── context_test.go       # regras de domínio e filesystem
├── doctor.go             # parsing/path helpers existentes reutilizados
├── status.go             # path helpers existentes reutilizados
└── workflow.go           # layout helpers existentes reutilizados

internal/workflowexec/
├── setup.go              # descrição pura + resolução operacional existente
└── setup_test.go         # contrato sem PATH/processo da descrição

README.md
README.pt-BR.md
README.es.md
CHANGELOG.md
```

**Structure Decision**: manter o projeto único e os testes junto aos pacotes. `context.go` concentra
o único modelo novo; nenhuma camada, interface ou pacote adicional é necessário.

## Implementation Boundaries

1. Criar um localizador específico que pare na primeira evidência estrutural Cerne. Não alterar o
   contrato fail-fast dos localizadores usados por `status`, `link` e `workflow`.
2. Compartilhar `readManifest`, validação de source, canonicalização e helpers de workflow, mas não
   chamar nem traduzir `Doctor`, `CurrentStatus` ou `workflowCheck`.
3. Extrair `workflowexec.Describe(provider)` como função pura; `Resolve` a reutiliza e continua sendo
   o único caminho que consulta `PATH` e prepara execução. Para evitar ciclo de imports,
   `workspace.Context(start, resolver)` recebe um `WorkflowResolver`, o pacote `workspace` não
   importa `workflowexec` e `cmd/cerne` injeta `workflowexec.Describe` na chamada ao domínio.
4. Exigir diretório de specs regular para estado `ready`, inclusive OpenSpec. Um provider pendente
   pode referenciar o path normativo ainda ausente sem publicá-lo como comprovado.
5. Rejeitar symlinks nos artefatos governados; um alias no caminho de invocação pode ser
   canonicalizado para a árvore física antes da validação.
6. Inicializar `problems` como slice vazio e montar structs em ordem declarada; não usar maps na
   serialização pública.

## Complexity Tracking

Nenhuma violação constitucional ou complexidade excepcional a justificar.
