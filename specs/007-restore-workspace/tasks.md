# Tasks: Restauração de Workspace

**Input**: Design documents from `specs/007-restore-workspace/`

**Prerequisites**: `plan.md`, `spec.md`, `research.md`, `data-model.md`, `contracts/`, `quickstart.md`

**Tests**: Testes devem ser escritos primeiro e falhar antes da implementação. Fluxos Git usam
somente repositórios locais temporários ou executáveis controlados, sem rede ou credenciais reais.

**Organization**: Tarefas agrupadas por história para preservar rastreabilidade e validação
independente.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Pode executar em paralelo por tocar arquivos diferentes e não depender de tarefa incompleta.
- **[Story]**: História da especificação atendida pela tarefa.
- Todos os itens possuem path exato do arquivo afetado.

## Phase 1: Setup (Shared Test Infrastructure)

**Purpose**: Preparar fixtures reutilizáveis sem adicionar dependência ou framework.

- [ ] T001 Criar helpers de repositório Git local, manifesto controlado e home temporária para restore em `internal/workspace/restore_test.go` e `cmd/cerne/main_test.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Tornar a promoção sem substituição reutilizável pelo init e pelo restore.

**⚠️ CRITICAL**: Nenhuma história começa antes desta fase.

- [ ] T002 Adicionar testes de regressão para promoção de diretório sem substituição e preservação de target concorrente em `internal/workspace/init_test.go`
- [ ] T003 Generalizar `promoteSource` para promoção de diretório no-replace e atualizar o caller existente em `internal/workspace/source_promote_linux.go`, `internal/workspace/source_promote_darwin.go`, `internal/workspace/source_promote_windows.go` e `internal/workspace/init.go`

**Checkpoint**: A primitiva multiplataforma mantém `init --clone` compatível e pode promover o root do restore.

---

## Phase 3: User Story 1 - Restaurar knowledge e clonar source (Priority: P1) 🎯 MVP

**Goal**: Restaurar, em uma invocação, knowledge e source clonados como repositórios independentes,
nomeados pelo manifesto e auditados sem executar workflow.

**Independent Test**: Usar duas origens Git locais temporárias e verificar nome, checkout,
histórico, remotos, manifesto, workflow preservado, separação, stdout/stderr, audit e origens intactas.

### Tests for User Story 1

> Escrever e executar estes testes primeiro; eles devem falhar antes da implementação.

- [ ] T004 [P] [US1] Adicionar testes de domínio para audit anterior ao Git, dois clones, manifesto regular, nome, versão, diretórios obrigatórios, workflow pronto/pendente aceito, workflow parcial/provider desconhecido recusado sem execução, source portátil, independência, promoção e resultado em `internal/workspace/restore_test.go`
- [ ] T005 [P] [US1] Adicionar testes de integração para help, parsing do modo `--clone`, stdout/stderr/status exatos, ausência de origens e provider não executado em `cmd/cerne/main_test.go`

### Implementation for User Story 1

- [ ] T006 [US1] Implementar request/result/failure, registro global da tentativa e transições knowledge/source sem origem ou fingerprint em `internal/workspace/restore.go`
- [ ] T007 [US1] Implementar staging privado, clone/validação de knowledge, manifesto regular, nome, versão, diretórios obrigatórios, estados de workflow sem executar provider, target clonado, independência, promoção, revalidação e rollback em `internal/workspace/restore.go`
- [ ] T008 [US1] Adicionar dispatch, parser, help, classificação das duas origens, adaptação Git, renderização e códigos do comando `restore --clone` em `cmd/cerne/main.go`

**Checkpoint**: US1 restaura dois repositórios locais controlados e pode ser demonstrada isoladamente.

---

## Phase 4: User Story 2 - Reutilizar source local existente (Priority: P2)

**Goal**: Restaurar knowledge e associar um working tree local sem copiar, mover ou modificar o
source, atualizando somente `manifest.source` quando necessário.

**Independent Test**: Comparar snapshot byte a byte do source antes/depois, validar a referência
final e confirmar preservação dos demais campos do manifesto.

### Tests for User Story 2

> Escrever e executar estes testes primeiro; eles devem falhar antes da implementação.

- [ ] T009 [P] [US2] Adicionar testes de domínio para source local non-bare, snapshot imutável, revalidação concorrente, separação e atualização exclusiva/preservadora do campo `source` em `internal/workspace/restore_test.go`
- [ ] T010 [P] [US2] Adicionar testes de integração para parsing, sucesso, linha opcional de manifesto atualizado e streams/status do modo `--source` em `cmd/cerne/main_test.go`

### Implementation for User Story 2

- [ ] T011 [US2] Implementar preflight, inspeção/reinspeção, separação futura e escrita atômica da referência local calculada contra o knowledge final em `internal/workspace/restore.go`
- [ ] T012 [US2] Integrar `--source` ao runner e renderizar source vinculado e mudança de manifesto em `cmd/cerne/main.go`

**Checkpoint**: US2 funciona isoladamente com source local externo e byte a byte idêntico.

---

## Phase 5: User Story 3 - Recusar restaurações inseguras ou incompletas (Priority: P3)

**Goal**: Recusar entradas/destinos inseguros, impedir vazamento e garantir audit persistente com
rollback limitado a artefatos cuja ownership seja demonstrada.

**Independent Test**: Injetar entradas maliciosas, destino concorrente e falha em cada transição;
verificar ausência de workspace parcial, preservação de conteúdo alheio e exatamente um audit
redigido final ou inconclusivo.

### Tests for User Story 3

> Escrever e executar estes testes primeiro; eles devem falhar antes do hardening.

- [ ] T013 [P] [US3] Adicionar testes de audit para symlinks, ownership/bits POSIX, IDs exclusivos, transições e falhas de persistência em `internal/workspace/restore_audit_test.go`, além de DACL privada e herança permissiva no Windows em `internal/workspace/restore_audit_windows_test.go`
- [ ] T014 [P] [US3] Adicionar testes negativos e de borda para manifesto/path estrutural symlink, nomes, paths Unix/Windows, traversal, sobreposição, target/destino concorrente, bare/worktree/vazio/subdiretório e duas origens iguais conforme o spec em `internal/workspace/restore_security_test.go`
- [ ] T015 [P] [US3] Adicionar testes de falha em cada clone/validação/promoção/cleanup, confirmação de identidade, rollback pós-promoção e audit inconclusivo em `internal/workspace/restore_failure_test.go`
- [ ] T016 [P] [US3] Adicionar testes de integração para uso inválido, allowlist, credenciais/query/fragmento, redaction, ausência de efeitos e erros operacionais seguros em `cmd/cerne/main_test.go`

### Implementation for User Story 3

- [ ] T017 [US3] Endurecer audit com `os.UserHomeDir`, `os.OpenRoot`, criação exclusiva, escrita atômica e permissões verificáveis em `internal/workspace/restore.go`, `internal/workspace/audit_permissions_unix.go` e `internal/workspace/audit_permissions_windows.go`
- [ ] T018 [US3] Implementar validação lexical portátil de source clonado, manifesto regular, sobreposições e destino obrigatoriamente ausente em `internal/workspace/restore.go`
- [ ] T019 [US3] Implementar cleanup por pai/prefixo/tipo/identidade, rollback do root promovido e preservação do último audit durável em `internal/workspace/restore.go`
- [ ] T020 [US3] Completar causas/correções redigidas e garantir status `2` para uso/origem inválida e status `1` para falha operacional em `cmd/cerne/main.go`

**Checkpoint**: Todas as recusas e falhas mantêm conteúdo preexistente, não vazam entradas e deixam
somente o audit quando a limpeza comprovadamente segura termina.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Documentar o contrato, preservar compatibilidade e validar a release nos três sistemas.

- [ ] T021 [P] Documentar sintaxe, rede, autenticação, autorização, audit, rollback, retomada, streams, status e exemplos em `README.md`
- [ ] T022 [P] Atualizar a documentação equivalente em português em `README.pt-BR.md`
- [ ] T023 [P] Atualizar a documentação equivalente em espanhol em `README.es.md`
- [ ] T024 Adicionar a feature compatível e seus efeitos ao changelog e preparar a versão MINOR `0.5.0` em `CHANGELOG.md` e `cmd/cerne/main.go`
- [ ] T025 Executar regressão de `init`, `doctor`, `status`, `link` e `workflow`, além de `gofmt`, `go vet ./...`, `go test -count=1 ./...` e `git diff --check`, corrigindo somente arquivos afetados listados em `specs/007-restore-workspace/plan.md`
- [ ] T026 Validar os cenários end-to-end e registrar resultados Linux/Windows/macOS em `specs/007-restore-workspace/quickstart.md` usando a matriz de `.github/workflows/test.yml`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: Sem dependências.
- **Foundational (Phase 2)**: Depende de Setup e bloqueia as histórias.
- **US1 (Phase 3)**: Depende de Foundational e entrega o núcleo transacional.
- **US2 (Phase 4)**: Depende do núcleo de US1 e acrescenta source local.
- **US3 (Phase 5)**: Depende de US1; testes de segurança do modo clone podem começar em paralelo com
  US2, mas a conclusão inclui recusas do modo local e depende de US2.
- **Polish (Phase 6)**: Depende das três histórias concluídas.

### User Story Dependency Graph

```text
Setup -> Foundational -> US1 -> US2 ----> US3 -> Polish
                            \-----------> US3
```

### Within Each User Story

- Testes devem ser escritos e observados falhando antes da implementação correspondente.
- Modelo/audit e validações de domínio precedem a integração CLI.
- A história termina somente após seu teste independente passar.
- Nenhuma etapa comunica sucesso antes da auditoria final e validação pós-promoção.

### Parallel Opportunities

- T004 e T005 podem ser escritos em paralelo.
- T009 e T010 podem ser escritos em paralelo.
- T013, T014, T015 e T016 podem ser escritos em paralelo em arquivos distintos.
- T021, T022 e T023 podem ser atualizados em paralelo após o contrato estabilizar.
- Depois de US1, os testes clone-only de US3 podem avançar enquanto US2 é implementada.

---

## Parallel Example: User Story 1

```text
Task T004: testes de domínio em internal/workspace/restore_test.go
Task T005: testes CLI em cmd/cerne/main_test.go
```

## Parallel Example: User Story 2

```text
Task T009: testes de domínio do source local em internal/workspace/restore_test.go
Task T010: testes CLI do modo local em cmd/cerne/main_test.go
```

## Parallel Example: User Story 3

```text
Task T013: auditoria e falhas de persistência em internal/workspace/restore_audit_test.go
Task T014: entradas, manifesto, paths e destinos em internal/workspace/restore_security_test.go
Task T015: rollback e ownership em internal/workspace/restore_failure_test.go
Task T016: contrato CLI e redaction em cmd/cerne/main_test.go
```

---

## Implementation Strategy

### MVP First (User Story 1)

1. Concluir Setup e Foundational.
2. Escrever os testes T004/T005 e confirmar falha.
3. Implementar T006–T008.
4. Validar US1 com duas origens locais temporárias.
5. Usar este ponto somente como demonstração de desenvolvimento; release exige o hardening de US3.

### Incremental Delivery

1. US1: dois clones e núcleo transacional auditado.
2. US2: associação imutável de source local.
3. US3: matriz completa de segurança, redaction, ownership e falhas.
4. Polish: documentação, versão, regressão e validação multiplataforma.

### Minimality Rules

- Reutilizar adapter Git, validadores, manifesto, doctor e promoção existentes antes de criar código.
- Manter a orquestração em `internal/workspace/restore.go`; não criar pacote, interface ou config
  sem segundo caso concreto.
- Não implementar resume, retenção, branch/tag/commit, retry, sync, SDK de host ou workflow automático.

## Notes

- `[P]` indica arquivo distinto e ausência de dependência incompleta dentro da fase.
- Cada tarefa deve deixar o menor teste executável que quebra sem o comportamento.
- Commits, push, tag e release permanecem fora deste fluxo e exigem autorização separada.
