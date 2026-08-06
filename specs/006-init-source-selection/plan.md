# Implementation Plan: Seleção de Source no Init

**Branch**: `feat/init-source-selection` | **Date**: 2026-08-05 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/006-init-source-selection/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command; its definition describes the execution workflow.

## Summary

Estender `cerne init` com dois modos opt-in: `--source` associa atomicamente um working tree Git
local usando as regras existentes de `link`, enquanto `--clone` obtém uma origem permitida em
`workspace/source`. O caminho sem flag permanece byte a byte compatível. A solução reutiliza o
adapter Git e os validadores atuais, acrescenta somente execução segura de clone e uma auditoria
prévia. O clone ocorre em área temporária privada e só é promovido sem substituição após validação;
falhas preservam knowledge/auditoria e removem apenas essa área controlada.

## Technical Context

**Language/Version**: Go 1.26.5, conforme `go.mod`

**Primary Dependencies**: Biblioteca padrão, `golang.org/x/sys` já instalado e Git local; nenhuma
dependência nova

**Storage**: Sistema de arquivos local, `knowledge/cerne.json`, repositórios Git e um registro JSON
fixo em `knowledge/runs`

**Testing**: `go test ./...`, repositórios Git temporários locais, callbacks de domínio e
executáveis Git falsos; nenhuma rede ou credencial real

**Target Platform**: Linux, Windows e macOS

**Project Type**: CLI de projeto único

**Performance Goals**: Validação local concluída sem percorrer conteúdo do working tree; clone
limitado pelo desempenho normal do Git e da origem

**Constraints**: Modo sem flag exato; `--source` e `--clone` mutuamente exclusivos e combináveis com `--workflow`; nenhum shell; protocolos
limitados; credenciais embutidas recusadas; prompts controláveis pelo Git, templates e hooks
neutralizados, com limitações de helpers externos documentadas; origem e output Git não
registrados; promoção privada não substitui source concorrente; falha pós-clone preserva auditoria;
nenhum comando novo de retomada, branch, depth, submódulo ou autenticação

**Scale/Scope**: Um workspace, um knowledge, um source e no máximo uma tentativa de clone por init;
três modos fechados, sem registry, host SDK ou estado global

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Ownership and separation — PASS**: Source local é apenas referenciado; clone cria um repositório
  independente e knowledge nunca é copiado para ele.
- **Neutral core — PASS**: Não há agente ou IA. Git continua em `internal/gitexec`; o domínio recebe
  funções e fatos definidos por seu contrato.
- **Minimum context and audit — PASS**: Cada clone real cria um JSON `started` antes do processo e
  o finaliza sem origem ou output; falha de finalização permanece inconclusiva e bloqueia sucesso.
- **Authorization and secrets — PASS**: `--clone` identifica a operação e seu destino; transportes
  e efeitos são documentados, credenciais embutidas são recusadas e autenticação permanece externa.
- **Portability — PASS**: Paths, processo sem shell, diretório vazio de templates/hooks e limpeza
  usam recursos portáveis; SCP-like possui validação testada nas três plataformas.
- **Testing — PASS**: Domínio, regressão, adapter, CLI, rollback, auditoria e segredos possuem testes
  determinísticos com repositórios e executáveis temporários.
- **CLI compatibility and documentation — PASS**: O modo padrão não muda; os modos novos definem
  sintaxe, streams, status, efeitos e ajuda como adição MINOR.
- **Simplicity — PASS**: Reutilizar `link`, callbacks, parser manual, JSON e `os/exec` evita framework,
  interface, factory, SDK de host, storage global ou comando de retomada.

**Pre-research gate**: PASS após escolher preservação de knowledge/auditoria em falha pós-clone e
recusar transportes/credenciais incompatíveis com os princípios de autorização e segredos.

**Post-design re-check**: PASS. Contratos limitam cada modo, tornam a transição de rollback
explícita, mantêm source externo imutável, auditam clone antes da execução e preservam o CLI legado.

## Project Structure

### Documentation (this feature)

```text
specs/006-init-source-selection/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── init-source-command.md
│   └── source-clone-audit.md
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
    ├── init_test.go
    ├── link.go
    └── link_test.go

README.md
README.pt-BR.md
README.es.md
CHANGELOG.md
```

**Structure Decision**: Manter parsing, help, streams e adaptação em `cmd/cerne`; ampliar o caso de
uso existente em `internal/workspace/init.go`, extraindo apenas validadores já presentes em
`link.go` quando necessário; acrescentar clone ao adapter existente `internal/gitexec/init.go`.
Uma request struct e um conjunto pequeno de callbacks bastam; nenhum novo pacote ou interface.
`doctor`, `status` e `link` não exigem mudança de produção porque já resolvem sources externos ou
ausentes pelo manifesto.

## Complexity Tracking

Nenhuma violação constitucional exige justificativa.
