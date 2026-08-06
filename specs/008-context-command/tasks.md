# Tasks: Contexto Estrutural do Workspace

**Input**: Design documents from `/specs/008-context-command/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md

**Tests**: Testes de domínio devem preceder a implementação; o adaptador de workflow requer teste
de contrato; o comando crítico requer integração cobrindo argumentos, streams, status, JSON e
efeitos. A matriz existente valida Linux, Windows e macOS.

**Organization**: As tarefas estão agrupadas por história. US1 entrega o MVP JSON, US2 acrescenta
a apresentação humana e US3 completa o diagnóstico seguro de estados inválidos e parciais.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: pode executar em paralelo por tocar arquivos diferentes e não depender de tarefa aberta
- **[Story]**: história da especificação (`US1`, `US2`, `US3`)

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: preservar a linha de base e confirmar que a feature não exige infraestrutura nova

- [ ] T001 Executar `go test -count=1 ./...` como baseline e confirmar que nenhuma dependência ou alteração de módulo é necessária em `go.mod`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: separar descrição estrutural de provider da resolução operacional que consulta PATH

**⚠️ CRITICAL**: Esta fase bloqueia todas as histórias.

- [ ] T002 Criar testes de contrato para descrições estáticas de Spec Kit/OpenSpec, provider desconhecido e ausência de consulta a PATH em `internal/workflowexec/setup_test.go`
- [ ] T003 Extrair `Describe(provider)` puro e fazer `Resolve(provider)` reutilizá-lo sem mudar setup, environment ou erros existentes em `internal/workflowexec/setup.go`

**Checkpoint**: providers podem ser descritos sem processo, environment ou disponibilidade local.

---

## Phase 3: User Story 1 - Entregar contexto estrutural para skills (Priority: P1) 🎯 MVP

**Goal**: entregar JSON schema 1 determinístico com paths comprovados, source interno/externo e
estado estrutural de workflow, sem ler conteúdo ou executar integrações.

**Independent Test**: executar `cerne context --json` na raiz e em descendentes de fixtures
saudáveis sem workflow, com Spec Kit e com OpenSpec; analisar JSON, comparar bytes repetidos e
confirmar paths canônicos, status 0 e ausência de conteúdo/Git/processos.

### Tests for User Story 1

> Escrever e observar falha antes da implementação.

- [ ] T004 [P] [US1] Criar testes de domínio para descoberta ancestral saudável, paths canônicos, coleções, source interno/externo, estados not-declared/pending/ready e determinismo em `internal/workspace/context_test.go`
- [ ] T005 [P] [US1] Criar testes de integração do binário para contrato JSON v1, ordem/indentação/newline, execução em descendentes, repetição byte a byte e stdout/stderr/status em `cmd/cerne/main_test.go`

### Implementation for User Story 1

- [ ] T006 [US1] Implementar structs tipadas `ContextReport`, workspace, knowledge, source, workflow e problems, mais descoberta ancestral e montagem saudável somente filesystem em `internal/workspace/context.go`
- [ ] T007 [US1] Receber `WorkflowResolver` em `workspace.Context`, usar o resolver injetado para normalizar specs e derivar not-declared/pending/ready sem importar `workflowexec` ou consultar executáveis em `internal/workspace/context.go`
- [ ] T008 [US1] Adicionar dispatch `context --json`, injetar `workflowexec.Describe` na chamada ao domínio e serializar com `encoding/json` usando status healthy/warning/invalid em `cmd/cerne/main.go`
- [ ] T009 [US1] Executar os testes focados de JSON/domínio que cobrem `internal/workspace/context_test.go` e `cmd/cerne/main_test.go`

**Checkpoint**: uma skill já consegue consumir o contexto saudável e de workflow conhecido pelo
schema público, sem depender da apresentação humana.

---

## Phase 4: User Story 2 - Inspecionar contexto no terminal (Priority: P2)

**Goal**: apresentar os mesmos fatos em português e fornecer ajuda completa sem acessar workspace.

**Independent Test**: executar `cerne context` em fixtures com source interno/externo e workflow
ausente/pendente/pronto; comparar stdout/status e executar `--help` fora de workspace.

### Tests for User Story 2

> Escrever e observar falha antes da implementação.

- [ ] T010 [US2] Criar testes de integração para relatório humano saudável, source interno/externo, workflow pendente/pronto e ajuda sem acesso ao filesystem em `cmd/cerne/main_test.go`

### Implementation for User Story 2

- [ ] T011 [US2] Implementar renderização humana, correções fixas em português, `context --help` e entrada do comando na ajuda global em `cmd/cerne/main.go`
- [ ] T012 [US2] Executar os testes focados da apresentação humana e ajuda definidos em `cmd/cerne/main_test.go`

**Checkpoint**: a inspeção humana funciona sem exigir leitura manual do manifesto e não altera o
contrato JSON do MVP.

---

## Phase 5: User Story 3 - Diagnosticar contexto incompleto com segurança (Priority: P3)

**Goal**: preservar fatos independentes, omitir objetos não comprovados e produzir problemas
ordenados para ausência de workspace e estruturas inseguras, sem qualquer correção automática.

**Independent Test**: exercitar ausência de workspace, candidato parcial, manifesto inválido,
versão futura, symlinks, sobreposições, coleções ausentes, source inválido e workflow parcial ou
desconhecido; confirmar catálogo, ordem, omissões, status 0/1/2 e filesystem inalterado.

### Tests for User Story 3

> Escrever e observar falha antes da implementação.

- [ ] T013 [P] [US3] Criar testes de domínio para candidato parcial mais próximo; manifesto ausente, malformado, symlink ou version 2; identidade divergente; source inseguro; `product`, `specs`, `decisions`, `policies` e `runs` inválidos; provider desconhecido; marker inválido; symlink interno; `.git` aninhado; specs ausente após materialização; gating e ordem de problemas; e ausência de leitura de `AGENTS.md`/`CLAUDE.md` inacessíveis ou com sentinelas em `internal/workspace/context_test.go`
- [ ] T014 [P] [US3] Criar testes de integração para JSON/humano inválidos, catálogo e ordem de problemas, uso inválido, status 1/2, ausência de erros brutos/segredos e snapshot somente leitura em `cmd/cerne/main_test.go`

### Implementation for User Story 3

- [ ] T015 [US3] Implementar agregação parcial, dependency gates, validação física de tipos/symlinks/sobreposições, catálogo fechado e ordem determinística de problemas em `internal/workspace/context.go`
- [ ] T016 [US3] Completar renderização de problemas, saída JSON válida com status 1 e erro de uso exato com stdout vazio/status 2 em `cmd/cerne/main.go`
- [ ] T017 [US3] Executar os testes focados de segurança e confirmar zero chamadas a Git/provider/processo e zero writes em `internal/workspace/context_test.go` e `cmd/cerne/main_test.go`

**Checkpoint**: todas as histórias funcionam e relatórios parciais nunca inventam paths ou fatos.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: documentação pública, release minor, compatibilidade e validação multiplataforma

- [ ] T018 [P] Documentar finalidade, sintaxe, JSON, streams, status, efeitos e exemplos em `README.md`
- [ ] T019 [P] Documentar finalidade, sintaxe, JSON, streams, status, efeitos e exemplos em `README.pt-BR.md`
- [ ] T020 [P] Documentar finalidade, sintaxe, JSON, streams, status, efeitos e exemplos em `README.es.md`
- [ ] T021 [P] Registrar a feature aditiva, schema público e ausência de migração em `CHANGELOG.md`
- [ ] T022 Atualizar a versão minor do binário para 0.6.0 e verificar que ajuda/versionamento permanecem compatíveis em `cmd/cerne/main.go`
- [ ] T023 Executar `gofmt`, `go test -count=1 ./...`, `go vet ./...`, build, cenários e protocolo manual SC-008 de `specs/008-context-command/quickstart.md`, confirmando a matriz Linux/Windows/macOS de `.github/workflows/test.yml`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: inicia imediatamente.
- **Foundational (Phase 2)**: depende de T001 e bloqueia as histórias.
- **US1 (Phase 3)**: depende de T003 e entrega o modelo/JSON compartilhado.
- **US2 (Phase 4)**: depende de US1 porque projeta o mesmo `ContextReport` em formato humano.
- **US3 (Phase 5)**: depende de US1; pode avançar em paralelo com US2 após o MVP.
- **Polish (Phase 6)**: depende das histórias desejadas; T022–T023 devem ocorrer após todas.

### User Story Dependency Graph

```text
Setup → Foundation → US1 (MVP) ─┬→ US2
                                └→ US3
US2 + US3 → Polish
```

### Within Each User Story

- Testes são escritos e falham antes da implementação.
- Modelo e domínio precedem o dispatch/renderizador CLI.
- A história termina somente após seu teste independente passar.
- Mudanças em `cmd/cerne/main.go` são sequenciais para evitar conflito; testes em pacotes distintos
  podem avançar em paralelo quando marcados `[P]`.

### Parallel Opportunities

- US1: T004 e T005 podem ser escritos em paralelo.
- Após US1: US2 e US3 podem avançar em paralelo por objetivos independentes, coordenando somente
  as alterações finais em `cmd/cerne/main.go`.
- US3: T013 e T014 podem ser escritos em paralelo.
- Polish: T018, T019, T020 e T021 podem executar em paralelo.

---

## Parallel Example: User Story 1

```text
Task T004: testes do modelo e filesystem em internal/workspace/context_test.go
Task T005: testes do contrato do binário em cmd/cerne/main_test.go
```

## Parallel Example: User Story 2 + User Story 3

```text
Fluxo A: T010 → T011 → T012
Fluxo B: T013 + T014 → T015 → T016 → T017
```

---

## Implementation Strategy

### MVP First: User Story 1

1. Executar T001–T003.
2. Escrever T004–T005 antes do código.
3. Implementar T006–T008.
4. Executar T009 e validar o JSON independentemente.
5. Parar nesse checkpoint se o consumidor `cerne-skills` precisar ser desbloqueado primeiro.

### Incremental Delivery

1. **US1**: contrato JSON neutro para skills.
2. **US2**: projeção humana e ajuda, sem segundo modelo.
3. **US3**: diagnóstico parcial e hardening de segurança.
4. **Polish**: documentação, versão minor e suíte multiplataforma.

## Notes

- Nenhuma tarefa adiciona dependency, storage, interface especulativa ou migração de manifesto.
- `Doctor`, `CurrentStatus` e execução de workflow não são reutilizados como agregadores.
- Nenhuma auditoria é criada: esta feature apenas consulta estrutura local.
- Commit, push, tag e release não fazem parte destas tarefas.
