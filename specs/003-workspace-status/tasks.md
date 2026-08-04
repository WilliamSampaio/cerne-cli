---
description: "Tarefas para implementar cerne status"
---

# Tasks: Status do Workspace

**Input**: Design documents from `specs/003-workspace-status/`

**Prerequisites**: `plan.md`, `spec.md`, `research.md`, `data-model.md`,
`contracts/status-command.md`, `quickstart.md`

**Tests**: Testes de domínio, contratos do adaptador Git, integração do binário, cenários
negativos, leitura exclusiva e portabilidade são obrigatórios pela constituição e pela
especificação.

**Organization**: As tarefas são agrupadas por história e mantêm testes falhando antes da
implementação correspondente.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Pode executar em paralelo após suas dependências, pois afeta arquivos diferentes.
- **[Story]**: Mapeia a tarefa para US1, US2 ou US3.
- Toda descrição inclui caminhos exatos.

## Phase 1: Setup

**Purpose**: Confirmar o caminho mínimo do plano, sem nova dependência ou framework.

- [X] T001 Confirmar que `go.mod` e `go.sum` permanecem sem dependência nova para `cerne status`

---

## Phase 2: Foundational

**Purpose**: Criar os blocos compartilhados para localizar workspace e consultar Git localmente.

**CRITICAL**: Nenhuma história deve iniciar antes de localização, manifesto e adaptador Git local
estarem definidos.

- [X] T002 [P] Escrever testes de contrato para raiz Git, ambiente `GIT_*` hostil, comandos somente-leitura permitidos e ausência de mutação em `internal/gitexec/status_test.go`
- [X] T003 [P] Escrever testes de domínio para localização por ancestral mais próximo, manifesto válido e resolução de `knowledge` e `source` dentro do workspace em `internal/workspace/status_test.go`
- [X] T004 Implementar coletor Git local com ambiente saneado, sem shell, prompts, locks opcionais ou remotos em `internal/gitexec/status.go`
- [X] T005 Implementar localização de workspace, leitura de manifesto e resolução canônica dos dois repositórios em `internal/workspace/status.go`

**Checkpoint**: O workspace pode ser localizado e os repositórios podem ser consultados sem alterar
arquivos ou Git.

---

## Phase 3: User Story 1 - Ver o estado geral do workspace (Priority: P1) MVP

**Goal**: Exibir projeto, workspace e os dois repositórios em um workspace válido, inclusive a
partir de subdiretórios.

**Independent Test**: Criar um workspace válido, executar `cerne status` na raiz e em
subdiretórios, e verificar relatório completo, stderr vazio e status zero.

### Tests for User Story 1

> **NOTE: Escrever e executar estes testes primeiro; eles devem falhar antes da implementação.**

- [X] T006 [P] [US1] Escrever testes de domínio para `Workspace Status` com projeto, raiz absoluta, `knowledge` e `source` limpos em `internal/workspace/status_test.go`
- [X] T007 [P] [US1] Escrever testes de integração do binário para saída exata de workspace limpo, execução em subdiretório, stderr vazio e status zero em `cmd/cerne/main_test.go`

### Implementation for User Story 1

- [X] T008 [US1] Implementar agregação de `Workspace Status` com dois `Repository Status` limpos em `internal/workspace/status.go`
- [X] T009 [US1] Implementar despacho `status`, ligação dos adaptadores e renderização básica do relatório em `cmd/cerne/main.go`

**Checkpoint**: US1 entrega o status mínimo útil de um workspace válido e pode ser demonstrada
isoladamente.

---

## Phase 4: User Story 2 - Distinguir estados Git relevantes (Priority: P2)

**Goal**: Exibir alterações pendentes separadas, detached HEAD e repositórios sem commits sem
tratar esses estados como erro.

**Independent Test**: Preparar repositórios locais com alterações fora do stage, em stage,
arquivos não rastreados, detached HEAD e ausência de commits; executar `cerne status` e confirmar
campos, contagens e status zero.

### Tests for User Story 2

> **NOTE: Escrever e executar estes testes primeiro; eles devem falhar antes da implementação.**

- [X] T010 [P] [US2] Adicionar testes de contrato para `symbolic-ref`, `rev-parse --verify`, `diff --name-only`, `diff --cached --name-only` e `ls-files --others --exclude-standard` em `internal/gitexec/status_test.go`
- [X] T011 [P] [US2] Adicionar testes de domínio para `limpo`, `alterações pendentes`, contagens separadas, `detached HEAD` e `sem commits` em `internal/workspace/status_test.go`
- [X] T012 [P] [US2] Adicionar testes de integração do binário para relatório com modificados, stage, não rastreados, detached HEAD, sem commits e status zero em `cmd/cerne/main_test.go`

### Implementation for User Story 2

- [X] T013 [US2] Implementar coleta de branch, commit, detached HEAD, ausência de commits e contagens separadas em `internal/gitexec/status.go`
- [X] T014 [US2] Implementar classificação `limpo` versus `alterações pendentes` a partir das contagens em `internal/workspace/status.go`
- [X] T015 [US2] Completar renderização de branch, commit, estado e contagens estáveis em `cmd/cerne/main.go`

**Checkpoint**: US2 distingue todos os estados Git previstos e mantém alterações pendentes como
resultado bem-sucedido.

---

## Phase 5: User Story 3 - Receber falhas claras sem efeitos colaterais (Priority: P3)

**Goal**: Falhar de forma clara e segura quando workspace, manifesto, caminhos ou Git não puderem
ser consultados.

**Independent Test**: Executar fora de workspace e em fixtures com manifesto, caminhos ou Git
inválidos; confirmar stdout vazio, stderr corretivo, caminho afetado, status diferente de zero e
snapshot sem mutação.

### Tests for User Story 3

> **NOTE: Escrever e executar estes testes primeiro; eles devem falhar antes da implementação.**

- [X] T016 [P] [US3] Adicionar testes de domínio para workspace não localizado, manifesto ausente/malformado, `source` ausente/escape/link e repositório inválido em `internal/workspace/status_test.go`
- [X] T017 [P] [US3] Adicionar testes de integração para erros em stderr com caminho afetado, stdout vazio, status um, ajuda stdout/status zero e uso inválido status dois em `cmd/cerne/main_test.go`
- [X] T018 [P] [US3] Adicionar teste de snapshot sem criação, remoção, alteração de arquivos, stage, branch, commit, remoto ou segredo exposto em `cmd/cerne/main_test.go`

### Implementation for User Story 3

- [X] T019 [US3] Implementar falhas operacionais sanitizadas para workspace, manifesto, paths e consulta Git em `internal/workspace/status.go`
- [X] T020 [US3] Implementar `cerne status --help`, uso inválido, streams stdout/stderr e status 0/1/2 em `cmd/cerne/main.go`
- [X] T021 [P] [US3] Documentar sintaxe, localização por ancestral, campos, estados especiais, streams, status, leitura exclusiva, limitações e exemplos em `README.md`

**Checkpoint**: US3 estabiliza contrato de erro, ajuda, segurança observável e documentação.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Validar cenários finais, portabilidade e gates do projeto.

- [X] T022 Validar os nove cenários de `specs/003-workspace-status/quickstart.md`, incluindo workspace limpo, subdiretórios, alterações separadas, detached HEAD, sem commits, falhas, ajuda, uso inválido e conclusão em até 5 segundos
- [X] T023 Executar `gofmt`, `go vet ./...`, `go test -count=1 ./...` e `git diff --check`, confirmando a matriz Linux/Windows/macOS em `.github/workflows/test.yml`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: Sem dependências.
- **Foundational (Phase 2)**: Depende de T001 e bloqueia todas as histórias.
- **US1 (Phase 3)**: Depende de T004 e T005.
- **US2 (Phase 4)**: Depende da US1 funcional para estender o relatório com estados Git.
- **US3 (Phase 5)**: Depende das US1 e US2 para estabilizar falhas, ajuda e segurança.
- **Polish (Phase 6)**: Depende das três histórias concluídas.

### User Story Dependencies

```text
Setup → Foundational → US1 → US2 → US3 → Polish
```

- **US1** entrega o status saudável mínimo.
- **US2** adiciona estados Git ricos sem alterar o fluxo saudável.
- **US3** adiciona falhas, ajuda, contrato seguro e documentação.

Cada história possui critério de teste isolado e deve ser validada antes da próxima prioridade.

### Within Each User Story

- Testes MUST ser escritos, executados e observados falhando antes da implementação correspondente.
- Contratos do adaptador Git precedem código em `internal/gitexec`.
- Domínio precede a integração do CLI na mesma história.
- O contrato de `cerne init` e `cerne doctor` MUST continuar passando após cada checkpoint.

## Parallel Opportunities

- T002 e T003 podem ser escritos em paralelo.
- T006 e T007 podem ser escritos em paralelo.
- T010, T011 e T012 podem ser escritos em paralelo.
- T016, T017, T018 e T021 podem avançar em paralelo após US2.
- T004 e T005 alteram arquivos diferentes, mas devem ser integrados antes de US1.

## Parallel Example: Foundational

```text
Task: "T002 Testar o coletor Git em internal/gitexec/status_test.go"
Task: "T003 Testar localização de workspace em internal/workspace/status_test.go"
```

## Parallel Example: User Story 1

```text
Task: "T006 Testar domínio do status limpo em internal/workspace/status_test.go"
Task: "T007 Testar binário para status limpo em cmd/cerne/main_test.go"
```

## Parallel Example: User Story 2

```text
Task: "T010 Testar adaptador Git para estados ricos em internal/gitexec/status_test.go"
Task: "T011 Testar domínio para classificação em internal/workspace/status_test.go"
Task: "T012 Testar binário para alterações pendentes em cmd/cerne/main_test.go"
```

## Parallel Example: User Story 3

```text
Task: "T016 Testar falhas de domínio em internal/workspace/status_test.go"
Task: "T017 Testar contrato de erros do CLI em cmd/cerne/main_test.go"
Task: "T018 Testar leitura exclusiva em cmd/cerne/main_test.go"
Task: "T021 Documentar status em README.md"
```

## Implementation Strategy

### MVP First

1. Concluir T001–T005.
2. Concluir T006–T009.
3. Validar US1 isoladamente em workspace criado por `cerne init`.
4. Não avançar para entrega final até US2, US3 e os gates finais cobrirem segurança e
   compatibilidade.

### Incremental Delivery

1. **US1**: projeto, workspace e dois repositórios limpos.
2. **US2**: alterações separadas, detached HEAD e sem commits.
3. **US3**: falhas claras, ajuda, status 0/1/2 e prova de leitura exclusiva.
4. **Polish**: quickstart, documentação, gates e matriz multiplataforma.

## Notes

- Não adicionar framework CLI, formato JSON, modo watch, remoto ou dependência nova.
- Não usar `git add`, `commit`, `checkout`, `reset`, `fetch`, `pull` ou comandos que alterem os
  repositórios.
- Não exibir nomes de arquivos alterados, conteúdo do manifesto, remotos, credenciais ou segredos.
- Marcar cada tarefa `[X]` somente após sua validação passar.
