# Implementation Plan: Instalação de Skills Cerne

**Branch**: `010-skill-install-command` | **Date**: 2026-08-07 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/010-skill-install-command/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command; its definition describes the execution workflow.

## Summary

Adicionar `cerne skill install <codex|claude>` como autorização explícita para instalar a skill
oficial `cerne-context` do pacote versionado `cerne-skills` entregue como artefato companheiro da
distribuição do `cerne-cli`. A implementação deve resolver esse pacote local/cacheado sem rede,
validar manifesto e compatibilidade com `cerne context --json` schema v1 antes de copiar, instalar
nos destinos oficiais de Codex e Claude, preservar conteúdo desconhecido, atualizar instalações
gerenciadas em versão diferente, registrar auditoria global privada para tentativas operacionais e
manter `init`, `restore` e `workflow setup` sem instalação implícita.

## Technical Context

**Language/Version**: Go 1.26.5, conforme `go.mod`

**Primary Dependencies**: Biblioteca padrão e `golang.org/x/sys` já instalado; nenhuma dependência
nova

**Storage**: Sistema de arquivos local; pacote companheiro `cerne-skills` local/cacheado gerenciado
pelo CLI; destinos `~/.codex/skills/cerne-context` e `~/.claude/skills/cerne-context`; auditoria
JSON exclusiva em `~/.cerne/audit` ou equivalente da home do usuário

**Testing**: `go test ./...`, fixtures locais de pacote, home temporária isolada e testes de CLI
para stdout, stderr, status e efeitos de filesystem; sem rede, GitHub, credenciais ou releases reais

**Target Platform**: Linux, Windows e macOS

**Project Type**: CLI de projeto único

**Performance Goals**: Operação local limitada ao tamanho do pacote oficial; instalação deve
percorrer somente o manifesto e os arquivos declarados da skill

**Constraints**: Comando explícito obrigatório; agentes públicos exatamente `codex` e `claude`;
`generic` recusado; nenhuma instalação em `init`, `restore` ou `workflow setup`; perfil do usuário
atual apenas; sem diretórios administrativos; sem rede nesta versão; pacote companheiro ausente ou
inacessível falha sem alterar destinos; manifesto validado antes de cópia; `contextSchema`
compatível com schema v1; links simbólicos e paths fora do pacote recusados; destino desconhecido
preservado; reinstalação da mesma versão idempotente; instalação gerenciada em versão diferente
atualizada automaticamente; falha antes da promoção final preserva instalação anterior; auditoria
redigida sem conteúdo de skill, variáveis de ambiente, tokens, remotes ou saída externa bruta

**Scale/Scope**: Um pacote oficial inicial (`cerne-skills`), uma skill inicial (`cerne-context`),
dois agentes públicos e uma tentativa de instalação por invocação

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Ownership and separation — PASS**: A instalação não modifica workspaces, knowledge ou source;
  escreve somente no perfil do usuário atual.
- **Neutral core — PASS**: O núcleo continua independente de IA. Diferenças de Codex e Claude
  ficam em adaptadores de destino e empacotamento.
- **Minimum context and audit — PASS**: O comando não lê conteúdo de workspace; a auditoria global
  registra apenas metadados mínimos da tentativa.
- **Authorization and secrets — PASS**: Só `cerne skill install <agent>` autoriza escrita no perfil
  do agente. Uso inválido não audita nem altera arquivos; auditoria e diagnósticos não registram
  segredos ou saída externa bruta.
- **Portability — PASS**: O desenho usa `os.UserHomeDir`, `filepath`, escrita em staging e promoção
  sem shell. Regras específicas de permissão reaproveitam o modelo já existente de auditoria global.
- **Testing — PASS**: A implementação terá testes de domínio, contrato de adaptadores, CLI e
  recusas de segurança com pacote local controlado e home temporária.
- **CLI compatibility and documentation — PASS**: O comando é aditivo. Contratos existentes de
  `init`, `restore`, `workflow setup` e `context` permanecem; ajuda e README devem documentar a
  ausência de instalação automática.
- **Simplicity — PASS**: Sem registry genérico, sem dependência nova e sem abstração especulativa.
  Reutilizar parsing/streams do CLI, validações de filesystem e auditoria global existente.

Any failed gate MUST be resolved before proceeding. Necessary complexity MUST be recorded below.

**Pre-research gate**: PASS. As decisões pendentes foram reduzidas a contrato de CLI, pacote
companheiro local/cacheado, destinos oficiais por agente, política de overwrite/upgrade, auditoria
e testes offline.

**Post-design re-check**: PASS. Os artefatos de design mantêm instalação explícita, pacote sem rede,
adapters limitados a `codex`/`claude`, auditoria redigida, testes determinísticos, upgrade seguro de
instalações gerenciadas e nenhum efeito colateral nos fluxos de workspace.

## Project Structure

### Documentation (this feature)

```text
specs/010-skill-install-command/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── checklists/
│   └── requirements.md
├── contracts/
│   ├── skill-install-audit-record.md
│   └── skill-install-command.md
└── tasks.md
```

### Source Code (repository root)

```text
cmd/cerne/
├── main.go
└── main_test.go

internal/
├── skillinstall/
│   ├── install.go
│   ├── install_test.go
│   ├── package.go
│   ├── package_test.go
│   ├── resolver.go
│   ├── resolver_test.go
│   ├── targets.go
│   └── targets_test.go
└── workspace/
    ├── audit_permissions_unix.go
    ├── audit_permissions_windows.go
    └── restore.go

README.md
README.pt-BR.md
README.es.md
CHANGELOG.md
```

**Structure Decision**: Manter sintaxe, ajuda, streams e códigos de saída em `cmd/cerne`. Criar
`internal/skillinstall` porque a feature não pertence ao domínio de workspace: ela instala pacote
global no perfil do usuário e pode rodar fora de um workspace Cerne. Resolver o pacote companheiro
por uma função pequena e testável, sem registry ou rede. Reaproveitar o modelo de auditoria global e
permissões já usado pelo restore, movendo somente o mínimo compartilhável se a implementação exigir.
Não adicionar registry dinâmico, flag pública `--package` ou agente `generic`; `codex` e `claude`
ficam como mapeamento explícito enquanto são os únicos requisitos atuais.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

Nenhuma violação constitucional exige justificativa.
