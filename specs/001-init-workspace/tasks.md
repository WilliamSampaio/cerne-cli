---
description: "Tarefas para implementar cerne init"
---

# Tasks: Inicialização de Workspace

**Input**: Design documents from `specs/001-init-workspace/`

**Prerequisites**: `plan.md`, `spec.md`, `research.md`, `data-model.md`,
`contracts/init-command.md`, `quickstart.md`

**Tests**: Testes de domínio, contrato Git, integração do CLI, rollback e portabilidade são
obrigatórios pela constituição.

**Organization**: As tarefas são agrupadas por história e mantêm testes antes da implementação.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Pode executar em paralelo por afetar arquivos diferentes e não depender de trabalho
  incompleto.
- **[Story]**: Mapeia a tarefa para US1, US2 ou US3.
- Todas as descrições incluem caminhos exatos.

## Phase 1: Setup

**Purpose**: Estabelecer a validação multiplataforma sem adicionar dependências.

- [X] T001 Configurar jobs `go test ./...` para Linux, Windows e macOS em `.github/workflows/test.yml`

---

## Phase 2: Foundational

**Purpose**: Criar o único adaptador externo compartilhado por todas as histórias.

**⚠️ CRITICAL**: A inicialização de workspace depende deste adaptador.

- [X] T002 Escrever testes de contrato inicialmente falhos para descoberta do Git, ambiente `GIT_*` isolado e repositório vazio em `internal/gitexec/init_test.go`
- [X] T003 Implementar descoberta do executável e `git init --quiet` sem shell em `internal/gitexec/init.go`

**Checkpoint**: O adaptador Git inicializa um repositório local vazio sem remoto ou commit.

---

## Phase 3: User Story 1 - Criar um workspace novo (Priority: P1) 🎯 MVP

**Goal**: Criar a árvore, o manifesto e dois repositórios Git independentes em destino ausente ou
vazio.

**Independent Test**: Executar `cerne init exemplo` em diretório temporário e verificar árvore,
manifesto, raízes Git distintas, `source/` vazio, stdout e status zero.

### Tests for User Story 1

> **NOTE: Escrever e executar estes testes primeiro; eles devem falhar antes da implementação.**

- [X] T004 [P] [US1] Escrever testes de domínio para destino ausente/vazio, árvore, manifesto e duas chamadas do adaptador em `internal/workspace/init_test.go`
- [X] T005 [P] [US1] Escrever teste de integração do binário para sucesso, stdout, manifesto, raízes Git, remotos e commits em `cmd/cerne/main_test.go`

### Implementation for User Story 1

- [X] T006 [US1] Implementar criação para destino ausente/vazio, árvore, `cerne.json` exclusivo e resultado em `internal/workspace/init.go`
- [X] T007 [US1] Implementar despacho `init`, preflight Git e saída de sucesso em `cmd/cerne/main.go`

**Checkpoint**: US1 cria um workspace válido completo e pode ser demonstrada isoladamente.

---

## Phase 4: User Story 2 - Preservar dados existentes em falhas (Priority: P2)

**Goal**: Recusar destinos inseguros, preservar todo conteúdo anterior e desfazer falhas parciais.

**Independent Test**: Exercitar nomes inválidos, arquivo/link/diretório não vazio, Git ausente e
falha na segunda inicialização; comparar o destino antes e depois.

### Tests for User Story 2

> **NOTE: Escrever e executar estes testes primeiro; eles devem falhar antes da implementação.**

- [X] T008 [P] [US2] Adicionar testes de domínio para nomes inválidos, destinos recusados e rollback após falha injetada em `internal/workspace/init_test.go`
- [X] T009 [P] [US2] Adicionar testes do CLI para status 1/2, stderr corretivo, stdout vazio, Git ausente e sentinelas intactas em `cmd/cerne/main_test.go`

### Implementation for User Story 2

- [X] T010 [US2] Implementar validação portátil do nome, inspeção com `Lstat`, recusa sem mutação e rollback reverso em `internal/workspace/init.go`
- [X] T011 [US2] Implementar classificação de uso/erro operacional e diagnósticos com causa e correção em `cmd/cerne/main.go`

**Checkpoint**: US2 falha com segurança e preserva destinos preexistentes em todos os cenários de
aceite.

---

## Phase 5: User Story 3 - Usar o comando de forma previsível (Priority: P3)

**Goal**: Fixar ajuda, streams, códigos de saída e documentação para humanos e scripts.

**Independent Test**: Verificar ajuda, saída exata, status 0/1/2 e execução em caminho com espaços
ou Unicode no mesmo teste executado pela matriz multiplataforma.

### Tests for User Story 3

> **NOTE: Escrever e executar este teste primeiro; ele deve falhar antes da implementação.**

- [X] T012 [US3] Adicionar testes do contrato de ajuda, saídas exatas, status e caminho portátil em `cmd/cerne/main_test.go`

### Implementation for User Story 3

- [X] T013 [P] [US3] Implementar `cerne init --help` e estabilizar o contrato de stdout/stderr/status em `cmd/cerne/main.go`
- [X] T014 [P] [US3] Documentar sintaxe, nome, árvore, efeitos, streams, status, erros e exemplos em `README.md`

**Checkpoint**: US3 é não interativa, documentada e adequada para automação nos três sistemas.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Executar as validações finais exigidas pelo plano e pela constituição.

- [X] T015 Validar todos os cenários e corrigir divergências do guia em `specs/001-init-workspace/quickstart.md`
- [X] T016 Executar `gofmt`, `go vet ./...` e `go test ./...` sobre `cmd/cerne/main.go`, `cmd/cerne/main_test.go`, `internal/gitexec/init.go`, `internal/gitexec/init_test.go`, `internal/workspace/init.go` e `internal/workspace/init_test.go`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: Sem dependências.
- **Foundational (Phase 2)**: Inicia após Setup e bloqueia as histórias.
- **US1 (Phase 3)**: Depende de T003.
- **US2 (Phase 4)**: Depende da US1 funcional para exercitar falhas do comando completo.
- **US3 (Phase 5)**: Depende das US1 e US2 para fixar todos os status e diagnósticos.
- **Polish (Phase 6)**: Depende das três histórias concluídas.

### User Story Dependencies

```text
Setup → Foundational → US1 → US2 → US3 → Polish
```

- **US1** entrega a criação válida.
- **US2** adiciona segurança de falha sobre o fluxo da US1.
- **US3** estabiliza e documenta o contrato completo das US1 e US2.

### Within Each User Story

- Testes MUST ser escritos e observados falhando antes da implementação correspondente.
- T004 e T005 podem ser escritos em paralelo; T006 precede T007.
- T008 e T009 podem ser escritos em paralelo; T010 precede T011.
- T012 precede T013 e T014; T013 e T014 podem executar em paralelo.

## Parallel Opportunities

- **US1**: T004 e T005 escrevem testes em pacotes diferentes.
- **US2**: T008 e T009 estendem arquivos de teste diferentes.
- **US3**: T013 altera o CLI enquanto T014 altera somente documentação.
- A matriz T001 executa os pacotes em três sistemas, mas não substitui as dependências entre fases.

## Parallel Examples

### User Story 1

```text
Task: "T004 [US1] Escrever testes de domínio em internal/workspace/init_test.go"
Task: "T005 [US1] Escrever teste de integração em cmd/cerne/main_test.go"
```

### User Story 2

```text
Task: "T008 [US2] Adicionar testes de domínio negativos em internal/workspace/init_test.go"
Task: "T009 [US2] Adicionar testes de falha do CLI em cmd/cerne/main_test.go"
```

### User Story 3

```text
Task: "T013 [US3] Implementar ajuda e contrato em cmd/cerne/main.go"
Task: "T014 [US3] Documentar o comando em README.md"
```

## Implementation Strategy

### MVP First

1. Concluir T001–T003.
2. Concluir T004–T007.
3. Validar US1 isoladamente.
4. Não publicar até US2, US3 e Polish concluírem os gates de segurança e compatibilidade.

### Incremental Delivery

1. **US1**: workspace válido e repositórios independentes.
2. **US2**: preservação de dados e rollback.
3. **US3**: contrato público, ajuda e documentação.
4. **Polish**: quickstart e validação total.

## Notes

- Não adicionar framework CLI, biblioteca Git, schema extra ou abstração além da função-adaptador.
- Não criar commit, remoto, `.gitignore`, `.gitkeep` ou arquivo em `source/`.
- Não usar shell nem remoto real nos testes.
- Marcar cada tarefa `[x]` somente após sua validação passar.
