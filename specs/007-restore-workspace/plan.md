# Implementation Plan: Restauração de Workspace

**Branch**: `feat/restore-workspace` | **Date**: 2026-08-05 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/007-restore-workspace/spec.md`

## Summary

Adicionar `cerne restore <knowledge-origin>` com uma escolha obrigatória entre `--source` e
`--clone`. A operação cria primeiro uma auditoria privada global, clona knowledge em staging
privado, obtém e valida o nome pelo manifesto, materializa ou associa source e promove o workspace
com rename sem substituição. Qualquer falha reverte somente o staging ou o root comprovadamente
criado pela tentativa; a auditoria global permanece. Origens são transitórias e não entram no
manifesto, na auditoria ou na saída do Cerne.

## Technical Context

**Language/Version**: Go 1.26.5, conforme `go.mod`

**Primary Dependencies**: Biblioteca padrão, `golang.org/x/sys` já instalado e Git local; nenhuma
dependência nova

**Storage**: Sistema de arquivos local, dois repositórios Git, `knowledge/cerne.json` existente e
um JSON exclusivo por tentativa em `~/.cerne/audit` ou equivalente da home do usuário

**Testing**: `go test ./...`, repositórios Git temporários locais, callbacks de domínio e Git falso;
nenhuma rede, credencial ou remoto real

**Target Platform**: Linux, Windows e macOS

**Project Type**: CLI de projeto único

**Performance Goals**: Validações locais sem percorrer conteúdo dos repositórios; no máximo dois
clones completos, limitados pelo desempenho normal do Git e das origens

**Constraints**: Auditoria durável antes do primeiro Git; exatamente um source; destino final
obrigatoriamente ausente; staging privado; promoção sem substituição; rollback integral dos alvos
com ownership confirmada e preservação de alvo ambíguo; source local imutável; nenhuma origem,
fingerprint ou output Git em registros/saídas; nenhum workflow,
agente, push, fetch adicional, submódulo, instalação ou deploy; audit `0700/0600` e owner-only em
POSIX, com DACL sem herança permissiva limitada ao usuário atual e `SYSTEM` no Windows

**Scale/Scope**: Uma tentativa, um knowledge e um source por invocação; um arquivo de auditoria e
no máximo dois subprocessos de clone; sem sync, retry, retenção automática ou seleção de revisão

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Ownership and separation — PASS**: As duas entradas são explícitas; source local não é alterado
  e os clones mantêm repositórios, históricos e remotos independentes.
- **Neutral core — PASS**: Não há IA, agente ou host específico. Git permanece em
  `internal/gitexec` e o domínio recebe operações pequenas já usadas pelo projeto.
- **Minimum context and audit — PASS**: Um JSON privado, redigido e exclusivo existe antes do
  primeiro processo Git, registra knowledge/source separadamente e sobrevive ao rollback.
- **Authorization and secrets — PASS**: A invocação autoriza somente os clones ou inspeção local
  pedidos; credenciais embutidas são recusadas e origens, fingerprints e output não são gravados.
- **Portability — PASS**: `filepath`, `os.UserHomeDir`, `os.Root`, processos sem shell e a promoção
  no-replace já implementada cobrem Linux, Windows e macOS. Bits/ownership POSIX e DACL explícita
  via `golang.org/x/sys/windows` tornam a privacidade verificável sem nova dependência.
- **Testing — PASS**: Domínio, adapter e CLI terão testes determinísticos para parsing, streams,
  status, redaction, corrida, auditoria, rollback e imutabilidade usando apenas fixtures locais.
- **CLI compatibility and documentation — PASS**: `restore` é adição compatível; manifesto versão
  1 e comandos existentes não mudam. Contratos definem ajuda, efeitos, saídas e códigos.
- **Simplicity — PASS**: Reutilizam-se classificador/clone Git, validação de repositório, leitura e
  escrita de manifesto, doctor e promoção existentes. Não há dependency, framework ou abstração
  especulativa nova.

**Pre-research gate**: PASS. Os pontos desconhecidos de auditoria global, transação e contrato CLI
foram enviados à pesquisa sem violação constitucional.

**Post-design re-check**: PASS. O desenho mantém a auditoria fora dos repositórios, torna ownership
da limpeza verificável, recusa destinos ambíguos e deixa workflow apenas preservado/validado.

## Project Structure

### Documentation (this feature)

```text
specs/007-restore-workspace/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── restore-audit-record.md
│   └── restore-command.md
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
    ├── audit_permissions_unix.go
    ├── audit_permissions_windows.go
    ├── doctor.go
    ├── init.go
    ├── init_test.go
    ├── link.go
    ├── restore.go
    ├── restore_audit_test.go
    ├── restore_audit_windows_test.go
    ├── restore_failure_test.go
    ├── restore_security_test.go
    ├── restore_test.go
    ├── source_promote_darwin.go
    ├── source_promote_linux.go
    └── source_promote_windows.go

README.md
README.pt-BR.md
README.es.md
CHANGELOG.md
```

**Structure Decision**: Manter parsing, help, streams e classificação das entradas em
`cmd/cerne`; implementar a transação coesa em um único `internal/workspace/restore.go`; reutilizar
`gitexec.ClassifyCloneOrigin` e `FindClone` sem alterar a política Git. Generalizar apenas o nome da
promoção de diretório existente para knowledge/source compartilharem a mesma primitiva no-replace.
Isolar somente a aplicação de permissões em dois arquivos com build tags, sem interface, para usar
bits/ownership em POSIX e a DACL nativa no Windows. Não chamar `InitWithSource*`: seu manifesto
novo, auditoria interna e rollback parcial têm contrato diferente do restore.

## Complexity Tracking

Nenhuma violação constitucional exige justificativa.
