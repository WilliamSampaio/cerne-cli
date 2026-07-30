---
description: "Tarefas para implementar cerne doctor"
---

# Tasks: Diagnóstico de Workspace

**Input**: Design documents from `specs/002-workspace-doctor/`

**Prerequisites**: `plan.md`, `spec.md`, `research.md`, `data-model.md`,
`contracts/doctor-command.md`, `quickstart.md`

**Tests**: Testes de domínio, contratos dos adaptadores, integração do binário, cenários negativos,
leitura exclusiva e portabilidade são obrigatórios pela constituição e pela especificação.

**Organization**: As tarefas são agrupadas por história e mantêm testes falhando antes da
implementação correspondente.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Pode executar em paralelo após suas dependências, pois afeta arquivos diferentes.
- **[Story]**: Mapeia a tarefa para US1, US2 ou US3.
- Toda descrição inclui caminhos exatos.

## Phase 1: Setup

**Purpose**: Adicionar somente a dependência nativa necessária à verificação de acesso efetivo.

- [ ] T001 Adicionar `golang.org/x/sys` compatível com Go 1.24.6 em `go.mod` e `go.sum`

---

## Phase 2: Foundational

**Purpose**: Criar os dois adaptadores de leitura compartilhados por todas as histórias.

**⚠️ CRITICAL**: Nenhuma história pode produzir diagnóstico confiável antes desta fase.

- [ ] T002 [P] Escrever testes de contrato inicialmente falhos para acesso permitido, negado, inconclusivo e sem mutação em `internal/filecheck/access_test.go`
- [ ] T003 [P] Escrever testes de contrato inicialmente falhos para raiz própria, repositório ancestral, common dir compartilhado, ambiente `GIT_*` hostil e ausência de mutação em `internal/gitexec/inspect_test.go`
- [ ] T004 [P] Implementar resultado comum e sondas efetivas sem criação ou escrita em `internal/filecheck/access.go`, `internal/filecheck/access_unix.go` e `internal/filecheck/access_windows.go`
- [ ] T005 [P] Implementar descoberta e inspeção local com top-level, common dir, ambiente saneado, locks e prompts desabilitados em `internal/gitexec/inspect.go`

**Checkpoint**: Permissões e identidade Git podem ser consultadas localmente sem alterar o
workspace.

---

## Phase 3: User Story 1 - Confirmar a saúde do workspace (Priority: P1) 🎯 MVP

**Goal**: Produzir dez aprovações, resumo saudável e status zero para um workspace válido criado
pelo `cerne init`.

**Independent Test**: Criar um workspace temporário, executar `cerne doctor` em sua raiz e verificar
dez linhas aprovadas, versão 1 implícita, repositórios independentes, resumo saudável e status zero.

### Tests for User Story 1

> **NOTE: Escrever e executar estes testes primeiro; eles devem falhar antes da implementação.**

- [ ] T006 [P] [US1] Escrever testes de domínio para manifesto legado e inteiro JSON `"version": 1`, caminhos canônicos, cinco diretórios, dez checks aprovados e resumo saudável em `internal/workspace/doctor_test.go`
- [ ] T007 [P] [US1] Escrever teste de integração do binário para workspace saudável, ordem e stdout exatos, stderr vazio e status zero em `cmd/cerne/main_test.go`

### Implementation for User Story 1

- [ ] T008 [US1] Implementar manifesto interno com v1 implícita e inteiro JSON `1` explícito, `CheckResult`, `Diagnosis`, coleta saudável e agregação das dez verificações em `internal/workspace/doctor.go`
- [ ] T009 [US1] Implementar despacho `doctor`, diretório atual, ligação dos adaptadores e relatório saudável em `cmd/cerne/main.go`

**Checkpoint**: US1 diagnostica de forma completa um workspace válido e pode ser demonstrada
isoladamente.

---

## Phase 4: User Story 2 - Localizar problemas bloqueantes (Priority: P2)

**Goal**: Exibir todos os erros aplicáveis, suas correções, resumo inválido e status um sem ocultar
checks dependentes.

**Independent Test**: Corromper isoladamente manifesto, caminhos, estrutura, Git, versão ou
permissões em fixtures temporários e confirmar linha de erro, correção, dez checks e status um.

### Tests for User Story 2

> **NOTE: Escrever e executar estes testes primeiro; eles devem falhar antes da implementação.**

- [ ] T010 [P] [US2] Adicionar testes table-driven para manifesto ausente/malformado, `name` inválido, versões `"1"`, `1.0`, `null` ou diferente de 1, caminho absoluto/escape/link, diretório ausente/tipo incorreto, Git ancestral/common dir, acesso negado e precedência de erro em `internal/workspace/doctor_test.go`
- [ ] T011 [P] [US2] Adicionar testes do binário para workspace inválido, Git ausente, correções, dez linhas em stdout, stderr vazio, status um e ausência de conteúdo privado em `cmd/cerne/main_test.go`

### Implementation for User Story 2

- [ ] T012 [US2] Implementar validações bloqueantes, propagação explícita de dependências e correções sanitizadas sem short-circuit em `internal/workspace/doctor.go`
- [ ] T013 [US2] Implementar renderização e status um para relatórios inválidos sem regressão do `init` em `cmd/cerne/main.go`

**Checkpoint**: US2 localiza falhas simultâneas, preserva dez resultados e nunca apresenta
aprovação para fato não verificado.

---

## Phase 5: User Story 3 - Consumir um diagnóstico previsível e seguro (Priority: P3)

**Goal**: Estabilizar avisos, ajuda, streams, status e a garantia observável de leitura exclusiva
para humanos e scripts.

**Independent Test**: Executar cenários saudável, somente com aviso, inválido, ajuda e uso incorreto
em caminho com espaços/Unicode, comparando saída, status e snapshots antes/depois.

### Tests for User Story 3

> **NOTE: Escrever e executar estes testes primeiro; eles devem falhar antes da implementação.**

- [ ] T014 [P] [US3] Adicionar testes de domínio para `name` válido divergente da raiz, acesso inconclusivo, resumo com avisos e precedência erro sobre aviso mantendo ordem e total dez em `internal/workspace/doctor_test.go`
- [ ] T015 [P] [US3] Adicionar testes do binário para aviso de `name`, ajuda, uso inválido, saída exata dos três resumos, status 0/1/2, caminho com espaços/Unicode e snapshot sem criação, alteração, rede, prompt ou segredo em `cmd/cerne/main_test.go`

### Implementation for User Story 3

- [ ] T016 [US3] Implementar aviso não bloqueante para `name` válido divergente, acesso inconclusivo, precedência final e dados estáveis de apresentação em `internal/workspace/doctor.go`
- [ ] T017 [US3] Implementar `cerne doctor --help`, símbolos, labels, resumos e status 0/1/2 preservando os contratos existentes em `cmd/cerne/main.go`
- [ ] T018 [P] [US3] Documentar sintaxe, dez checks, divergência de `name`, versão implícita e formato explícito, símbolos, streams, status 0/1/2, leitura exclusiva, limitações e exemplos em `README.md`

**Checkpoint**: US3 entrega contrato não interativo, documentado e estável nos três sistemas.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Validar todos os cenários, dependências, documentação e gates multiplataforma.

- [ ] T019 Validar os nove cenários de `specs/002-workspace-doctor/quickstart.md`, incluindo divergência de `name`, versão explícita, Git ausente, leitura exclusiva, ajuda e status 0/1/2
- [ ] T020 Executar `gofmt`, `go mod tidy`, `go vet ./...`, `go test -count=1 ./...` e `git diff --check`, confirmando a matriz Linux/Windows/macOS em `.github/workflows/test.yml` para `cmd/cerne`, `internal/workspace`, `internal/gitexec` e `internal/filecheck`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: Sem dependências.
- **Foundational (Phase 2)**: Depende de T001 e bloqueia todas as histórias.
- **US1 (Phase 3)**: Depende de T004 e T005.
- **US2 (Phase 4)**: Depende da US1 funcional para estender o mesmo relatório com erros.
- **US3 (Phase 5)**: Depende das US1 e US2 para estabilizar os três estados e todos os status.
- **Polish (Phase 6)**: Depende das três histórias concluídas.

### User Story Dependencies

```text
Setup → Foundational → US1 → US2 → US3 → Polish
```

- **US1** entrega o diagnóstico saudável mínimo.
- **US2** adiciona falhas bloqueantes sem alterar o fluxo saudável.
- **US3** adiciona avisos e fixa o contrato público completo.

Cada história possui critério de teste isolado, embora compartilhem o mesmo relatório e entrypoint e
devam ser integradas na ordem acima.

### Within Each User Story

- Testes MUST ser escritos, executados e observados falhando antes da implementação correspondente.
- Implementações de adaptador só começam após seus respectivos testes de contrato.
- Domínio precede a integração do CLI na mesma história.
- O contrato do `cerne init` MUST continuar passando após cada checkpoint.

## Parallel Opportunities

- T002 e T003 podem ser escritos em paralelo.
- T004 e T005 podem ser implementados em paralelo após os respectivos testes.
- T006 e T007 podem ser escritos em paralelo.
- T010 e T011 podem ser escritos em paralelo.
- T014 e T015 podem ser escritos em paralelo.
- T018 pode avançar em paralelo à implementação da US3 por alterar somente documentação.

## Parallel Example: Foundational

```text
Task: "T002 Testar o adaptador de acesso em internal/filecheck/access_test.go"
Task: "T003 Testar o adaptador Git em internal/gitexec/inspect_test.go"
```

Após os testes falharem:

```text
Task: "T004 Implementar internal/filecheck/access*.go"
Task: "T005 Implementar internal/gitexec/inspect.go"
```

## Parallel Example: User Story 1

```text
Task: "T006 Testar diagnóstico saudável em internal/workspace/doctor_test.go"
Task: "T007 Testar binário saudável em cmd/cerne/main_test.go"
```

## Parallel Example: User Story 2

```text
Task: "T010 Testar falhas de domínio em internal/workspace/doctor_test.go"
Task: "T011 Testar falhas do CLI em cmd/cerne/main_test.go"
```

## Parallel Example: User Story 3

```text
Task: "T014 Testar avisos e precedência em internal/workspace/doctor_test.go"
Task: "T015 Testar contrato e leitura exclusiva em cmd/cerne/main_test.go"
```

## Implementation Strategy

### MVP First

1. Concluir T001–T005.
2. Concluir T006–T009.
3. Validar US1 isoladamente em workspace criado por `cerne init`.
4. Não publicar até US2, US3 e os gates finais concluírem segurança e compatibilidade.

### Incremental Delivery

1. **US1**: dez aprovações e resumo saudável.
2. **US2**: erros completos, corretivos e status bloqueante.
3. **US3**: avisos, ajuda, contrato estável e prova de leitura exclusiva.
4. **Polish**: quickstart, documentação, dependências e matriz multiplataforma.

## Notes

- Não alterar o manifesto gerado por `cerne init`; ausência de `version` é versão 1 aprovada.
- Não adicionar framework CLI, interface de implementação única, arquivo-sonda ou acesso remoto.
- Não usar `git status`, shell, prompts, cores ou logging persistente.
- Marcar cada tarefa `[X]` somente após sua validação passar.
