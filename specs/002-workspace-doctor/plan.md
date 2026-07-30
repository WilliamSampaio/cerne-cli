# Implementation Plan: Diagnóstico de Workspace

**Branch**: `feat/doctor` | **Date**: 2026-07-29 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/002-workspace-doctor/spec.md`

## Summary

Adicionar `cerne doctor` ao despacho manual existente. O caso de uso produzirá uma lista fixa de
dez resultados ao inspecionar manifesto, estrutura, permissões e limites Git do diretório atual.
Git e consultas nativas de acesso efetivo entram por funções-adaptadoras; o domínio agrega
severidades e o CLI renderiza o contrato público. Nenhuma verificação cria, corrige ou altera
artefatos. Nome válido divergente da raiz gera aviso; versão explícita aceita somente o inteiro
JSON `1`, enquanto a ausência do campo representa v1 implícita.

## Technical Context

**Language/Version**: Go 1.24.6, conforme `go.mod`

**Primary Dependencies**: Biblioteca padrão, executável Git local e `golang.org/x/sys` para
consultas nativas de acesso efetivo sem mutação

**Storage**: Sistema de arquivos local, `knowledge/cerne.json` e metadados dos dois repositórios
Git; nenhuma escrita

**Testing**: `go test ./...`, testes table-driven, diretórios e repositórios temporários, teste do
binário real e matriz CI existente

**Target Platform**: Linux, Windows e macOS

**Project Type**: CLI de projeto único

**Performance Goals**: Diagnóstico completo de um workspace mínimo em até 5 segundos, sem rede

**Constraints**: Dez resultados em ordem fixa; leitura exclusiva; zero prompts, remotos,
credenciais ou agentes; caminhos sem links ou escape; nenhum shell; stdout/stderr estáveis;
status `0` para válido, `1` para erro bloqueante e `2` para uso inválido

**Scale/Scope**: Um workspace por invocação, um manifesto, dois repositórios, cinco diretórios
obrigatórios e dez verificações

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Ownership and separation — PASS**: O comando lê somente metadados indispensáveis e valida que
  conhecimento e código têm raízes próprias, distintas e sem contenção.
- **Neutral core — PASS**: Não há IA ou fornecedor. Git e acesso nativo entram por
  funções-adaptadoras; detalhes de processo e sistema operacional ficam fora das regras de
  classificação.
- **Minimum context and audit — PASS**: Nenhum agente ou tarefa automatizada é executado. O
  relatório completo é observável em stdout; persistir auditoria violaria o requisito explícito de
  leitura exclusiva.
- **Authorization and secrets — PASS**: A invocação autoriza apenas inspeção local. Não há operação
  sensível, rede ou credencial, e diagnósticos não reproduzem conteúdo privado do manifesto.
- **Portability — PASS**: Caminhos usam semântica portátil; permissões efetivas usam adaptadores
  nativos; incerteza vira aviso. A CI existente executa nos três sistemas.
- **Testing — PASS**: O plano cobre domínio, contratos dos adaptadores, CLI real, cenários
  negativos, ausência de mutação e regressão do `init`.
- **CLI compatibility and documentation — PASS**: Ordem, símbolos, labels, resumos, streams,
  status `0`/`1`/`2` e ajuda são definidos em contrato e serão documentados junto ao comando.
- **Simplicity — PASS**: O parsing manual e a função-adaptadora existentes são preservados. A única
  dependência nova é necessária para consultar ACLs e identidade efetiva sem arquivo-sonda.

**Pre-research gate**: PASS. Não há violação constitucional ou esclarecimento pendente.

**Post-design re-check**: PASS. O modelo mantém exatamente dez resultados; os contratos impedem
rede e mutação; o desenho não adiciona framework, interface de implementação única, filesystem
virtual ou alteração no manifesto produzido por `cerne init`. Nome válido divergente da raiz
permanece seguro como aviso; manifesto sem versão preserva compatibilidade como v1 e versão
explícita exige inteiro JSON `1`; os status auditáveis são `0`, `1` e `2`.

## Project Structure

### Documentation (this feature)

```text
specs/002-workspace-doctor/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── doctor-command.md
└── tasks.md
```

### Source Code (repository root)

```text
cmd/cerne/
├── main.go
└── main_test.go

internal/
├── filecheck/
│   ├── access.go
│   ├── access_test.go
│   ├── access_unix.go
│   └── access_windows.go
├── gitexec/
│   ├── init.go
│   ├── init_test.go
│   ├── inspect.go
│   └── inspect_test.go
└── workspace/
    ├── doctor.go
    ├── doctor_test.go
    ├── init.go
    └── init_test.go

.github/workflows/test.yml
README.md
go.mod
go.sum
```

**Structure Decision**: Estender `internal/workspace` com o caso de uso e os tipos do relatório,
estender o adaptador Git já existente para inspeção e isolar somente a consulta de permissão
específica de plataforma em `internal/filecheck`. O caso de uso recebe funções, não interfaces.
O entrypoint atual continua responsável por despacho, ajuda, streams e status.

## Design Decisions

### Fluxo

1. O CLI valida a sintaxe, obtém o diretório atual e descobre o Git.
2. O caso de uso resolve a raiz absoluta, coleta fatos de filesystem, manifesto, Git e permissão
   sem interromper verificações ainda confiáveis.
3. Dependências ausentes produzem um resultado de erro na posição correspondente, nunca omissão.
4. O relatório agrega erro acima de aviso e aviso acima de aprovação.
5. O CLI renderiza dez linhas e um resumo em stdout; somente falhas anteriores ao relatório usam
   stderr. O status é `0` sem erro, `1` com erro bloqueante e `2` para uso inválido.

### Limites mínimos

- `workspace` contém manifesto interno, severidade, resultado e diagnóstico; reutiliza
  `ValidateName` e a lista existente de diretórios de conhecimento. Nome inválido é erro; nome
  válido diferente da raiz é aviso não bloqueante.
- `gitexec` retorna fatos de raiz de worktree e diretório Git comum por comandos locais e
  não modificadores, com ambiente saneado.
- `filecheck` retorna `aprovado`, `negado` ou `inconclusivo` para acesso efetivo. Arquivos de
  plataforma usam build tags, sem arquivo temporário.
- `main` separa `runInit` e `runDoctor` para não alterar o contrato já testado do `init`.

Nenhuma camada adicional, schema externo, modo de correção, formato de máquina ou descoberta em
diretórios ancestrais é planejado.
