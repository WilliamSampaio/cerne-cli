---
description: "Tarefas para implementar cerne link"
---

# Tasks: Link de Repositório Source

**Input**: Design documents from `specs/004-link-source-repository/`

**Prerequisites**: `plan.md`, `spec.md`, `research.md`, `data-model.md`,
`contracts/link-command.md`, `quickstart.md`

**Tests**: Testes de domínio, contrato do adaptador Git, integração do binário, cenários negativos,
autorização explícita, ausência de mutação, escrita atômica e portabilidade são obrigatórios pela
constituição e pela especificação.

**Organization**: As tarefas são agrupadas por história e mantêm testes falhando antes da
implementação correspondente.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Pode executar em paralelo após suas dependências, pois afeta arquivos diferentes.
- **[Story]**: Mapeia a tarefa para US1, US2 ou US3.
- Toda descrição inclui caminhos exatos.

## Phase 1: Setup

**Purpose**: Confirmar o caminho mínimo do plano, sem dependência nova ou framework CLI.

- [X] T001 Confirmar que `go.mod` e `go.sum` permanecem sem dependência nova para `cerne link`
- [X] T002 [P] Revisar os helpers existentes de manifesto, path e localização em `internal/workspace/doctor.go` e `internal/workspace/status.go`
- [X] T003 [P] Revisar os adaptadores Git existentes e o saneamento de ambiente em `internal/gitexec/inspect.go` e `internal/gitexec/status.go`

---

## Phase 2: Foundational

**Purpose**: Criar os blocos compartilhados para localizar workspace, validar Git local e gravar o
manifesto de forma segura.

**CRITICAL**: Nenhuma história deve iniciar antes de localização, validação Git read-only e escrita
atômica estarem definidos.

- [X] T004 [P] Escrever testes de contrato para consulta Git read-only de repositório não-bare, bare, worktree e ambiente `GIT_*` hostil em `internal/gitexec/link_test.go`
- [X] T005 [P] Escrever testes de domínio para `Link Request`, localização por ancestral, leitura de manifesto e resolução do caminho informado em `internal/workspace/link_test.go`
- [X] T006 Implementar coletor Git local de link com worktree, common dir e bare detection em `internal/gitexec/link.go`
- [X] T007 Implementar modelos e erros de domínio para `Link Request`, `Link Result` e `Failure Result` em `internal/workspace/link.go`
- [X] T008 Implementar resolução canônica de workspace, knowledge, source atual e source candidato em `internal/workspace/link.go`
- [X] T009 Implementar gravação atômica do manifesto preservando campos existentes em `internal/workspace/link.go`

**Checkpoint**: O domínio consegue carregar manifesto, validar caminhos, consultar Git localmente e
preparar uma atualização segura sem expor o comando público.

---

## Phase 3: User Story 1 - Vincular um source local válido (Priority: P1) MVP

**Goal**: Atualizar o manifesto para apontar para um repositório Git local existente, sem alterar o
repositório vinculado.

**Independent Test**: Criar workspace e repositório local válido, executar `cerne link <caminho>
--replace`, verificar stdout, manifesto atualizado e snapshot sem mutação no source.

### Tests for User Story 1

> **NOTE: Escrever e executar estes testes primeiro; eles devem falhar antes da implementação.**

- [X] T010 [P] [US1] Escrever testes de domínio para link de caminho relativo, absoluto, espaços/Unicode, aliases canônicos de filesystem e fallback absoluto quando caminho relativo portátil não for possível em `internal/workspace/link_test.go`
- [X] T011 [P] [US1] Escrever testes de contrato Git para aceitar repositório não-bare e worktree válido em `internal/gitexec/link_test.go`
- [X] T012 [P] [US1] Escrever testes de integração do binário para saída de sucesso, stderr vazio, status zero e manifesto atualizado em `cmd/cerne/main_test.go`
- [X] T013 [P] [US1] Escrever teste de snapshot provando que o source vinculado não muda em `cmd/cerne/main_test.go`

### Implementation for User Story 1

- [X] T014 [US1] Implementar validação do source candidato válido e geração de caminho normalizado para manifesto em `internal/workspace/link.go`
- [X] T015 [US1] Implementar aceitação de worktree válido e recusa de bare repository no adaptador em `internal/gitexec/link.go`
- [X] T016 [US1] Implementar atualização do campo `source` no manifesto após todas as validações em `internal/workspace/link.go`
- [X] T017 [US1] Implementar despacho `link`, parsing de `<caminho>` e `--replace`, adaptação Git e renderização de sucesso em `cmd/cerne/main.go`

**Checkpoint**: US1 entrega o MVP de link seguro para um source válido e pode ser demonstrada
isoladamente.

---

## Phase 4: User Story 2 - Evitar substituição acidental de source (Priority: P2)

**Goal**: Recusar troca de source sem autorização explícita e permitir substituição somente com
`--replace`, preservando source antigo e novo.

**Independent Test**: Preparar source atual e novo source, confirmar recusa sem `--replace`,
sucesso com `--replace` e no-op quando o source informado já é o configurado.

### Tests for User Story 2

> **NOTE: Escrever e executar estes testes primeiro; eles devem falhar antes da implementação.**

- [X] T018 [P] [US2] Escrever testes de domínio para recusa sem `--replace`, substituição com `--replace` e source atual inválido exigindo `--replace` em `internal/workspace/link_test.go`
- [X] T019 [P] [US2] Escrever testes de domínio para no-op quando source candidato já é o configurado em `internal/workspace/link_test.go`
- [X] T020 [P] [US2] Escrever testes de integração do binário para stderr/status um sem `--replace`, stdout/status zero com `--replace` e stdout/status zero em no-op em `cmd/cerne/main_test.go`
- [X] T021 [P] [US2] Escrever teste de snapshot provando que source anterior e novo source não mudam durante substituição em `cmd/cerne/main_test.go`

### Implementation for User Story 2

- [X] T022 [US2] Implementar comparação entre source atual e source candidato por caminho normalizado e fatos Git em `internal/workspace/link.go`
- [X] T023 [US2] Implementar bloqueio de substituição sem `--replace` e erro corretivo em `internal/workspace/link.go`
- [X] T024 [US2] Implementar fluxo no-op sem regravar manifesto quando o source já está configurado em `internal/workspace/link.go`
- [X] T025 [US2] Completar renderização CLI para substituição, recusa e no-op em `cmd/cerne/main.go`

**Checkpoint**: US2 protege contra substituição acidental e mantém todos os repositórios intactos.

---

## Phase 5: User Story 3 - Receber falhas claras e seguras (Priority: P3)

**Goal**: Falhar de forma clara e sem efeitos colaterais quando workspace, manifesto, caminho,
repositório Git ou separação entre knowledge/source forem inválidos.

**Independent Test**: Executar fixtures inválidas e confirmar stdout vazio, stderr corretivo,
status não-zero, manifesto preservado e nenhum acesso remoto ou mutação Git.

### Tests for User Story 3

> **NOTE: Escrever e executar estes testes primeiro; eles devem falhar antes da implementação.**

- [X] T026 [P] [US3] Escrever testes de domínio para workspace não localizado, manifesto ausente, malformado e versão incompatível em `internal/workspace/link_test.go`
- [X] T027 [P] [US3] Escrever testes de domínio para caminho inexistente, arquivo regular, repo Git inválido, bare repo, source igual ao knowledge e aninhamento perigoso em `internal/workspace/link_test.go`
- [X] T028 [P] [US3] Escrever testes de integração para erros em stderr com caminho afetado, stdout vazio, status um, ajuda stdout/status zero e uso inválido status dois em `cmd/cerne/main_test.go`
- [X] T029 [P] [US3] Escrever teste de falha de gravação segura preservando manifesto anterior em `internal/workspace/link_test.go`
- [X] T030 [P] [US3] Escrever teste de contrato garantindo que o adaptador Git não chama fetch, pull, push, checkout, reset, add, commit ou clean em `internal/gitexec/link_test.go`

### Implementation for User Story 3

- [X] T031 [US3] Implementar validações bloqueantes e mensagens sanitizadas para workspace, manifesto e paths em `internal/workspace/link.go`
- [X] T032 [US3] Implementar validação de separação entre knowledge e source por caminho, worktree root e common dir em `internal/workspace/link.go`
- [X] T033 [US3] Implementar rollback natural da escrita atômica e limpeza segura de temporário em falha em `internal/workspace/link.go`
- [X] T034 [US3] Implementar `cerne link --help`, uso inválido, streams stdout/stderr e status 0/1/2 em `cmd/cerne/main.go`
- [X] T035 [P] [US3] Documentar sintaxe, `--replace`, validações, streams, status, leitura/escrita permitida, limitações e exemplos em `README.md`

**Checkpoint**: US3 estabiliza contrato de erro, ajuda, segurança observável e documentação.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Validar cenários finais, portabilidade e gates do projeto.

- [X] T036 Validar os doze cenários de `specs/004-link-source-repository/quickstart.md`, incluindo link relativo, absoluto, subdiretório, no-op, replace, worktree, bare, paths inválidos, aninhamento, escrita segura, ajuda, uso inválido e conclusão em até 5 segundos em workspace pequeno
- [X] T037 Executar `gofmt`, `go vet ./...`, `go test -count=1 ./...` e `git diff --check`, confirmando a matriz Linux/Windows/macOS em `.github/workflows/test.yml`
- [X] T038 Revisar se `cerne init`, `cerne doctor` e `cerne status` continuam com contratos existentes em `cmd/cerne/main_test.go`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: Sem dependências.
- **Foundational (Phase 2)**: Depende de T001–T003 e bloqueia todas as histórias.
- **US1 (Phase 3)**: Depende de T004–T009.
- **US2 (Phase 4)**: Depende da US1 funcional para trocar ou reconhecer source configurado.
- **US3 (Phase 5)**: Depende das US1 e US2 para estabilizar falhas, ajuda e segurança.
- **Polish (Phase 6)**: Depende das três histórias concluídas.

### User Story Dependencies

```text
Setup → Foundational → US1 → US2 → US3 → Polish
```

- **US1** entrega o MVP de vínculo para source válido.
- **US2** adiciona autorização explícita para substituição e no-op.
- **US3** adiciona falhas claras, ajuda, contrato seguro e documentação.

Cada história possui critério de teste isolado e deve ser validada antes da próxima prioridade.

### Within Each User Story

- Testes MUST ser escritos, executados e observados falhando antes da implementação correspondente.
- Contratos do adaptador Git precedem código em `internal/gitexec`.
- Domínio precede integração do CLI na mesma história.
- O contrato de `cerne init`, `cerne doctor` e `cerne status` MUST continuar passando após cada
  checkpoint.

## Parallel Opportunities

- T002 e T003 podem ser feitos em paralelo.
- T004 e T005 podem ser escritos em paralelo.
- T010, T011, T012 e T013 podem ser escritos em paralelo.
- T018, T019, T020 e T021 podem ser escritos em paralelo.
- T026, T027, T028, T029, T030 e T035 podem avançar em paralelo após US2.
- T006 e T007 alteram arquivos diferentes, mas devem ser integrados antes das histórias.

## Parallel Example: Foundational

```text
Task: "T004 Testar adaptador Git de link em internal/gitexec/link_test.go"
Task: "T005 Testar domínio base do link em internal/workspace/link_test.go"
```

## Parallel Example: User Story 1

```text
Task: "T010 Testar domínio de link válido em internal/workspace/link_test.go"
Task: "T011 Testar Git não-bare e worktree em internal/gitexec/link_test.go"
Task: "T012 Testar binário para sucesso em cmd/cerne/main_test.go"
Task: "T013 Testar ausência de mutação no source em cmd/cerne/main_test.go"
```

## Parallel Example: User Story 2

```text
Task: "T018 Testar recusa e replace no domínio em internal/workspace/link_test.go"
Task: "T019 Testar no-op no domínio em internal/workspace/link_test.go"
Task: "T020 Testar binário para replace/no-op em cmd/cerne/main_test.go"
Task: "T021 Testar snapshots dos sources em cmd/cerne/main_test.go"
```

## Parallel Example: User Story 3

```text
Task: "T026 Testar falhas de workspace e manifesto em internal/workspace/link_test.go"
Task: "T027 Testar falhas de path, Git e separação em internal/workspace/link_test.go"
Task: "T028 Testar contrato CLI de erros e ajuda em cmd/cerne/main_test.go"
Task: "T030 Testar comandos Git proibidos em internal/gitexec/link_test.go"
Task: "T035 Documentar cerne link em README.md"
```

## Implementation Strategy

### MVP First

1. Concluir T001–T009.
2. Concluir T010–T017.
3. Validar US1 isoladamente com repositório local válido e snapshot sem mutação.
4. Não avançar para entrega final até US2, US3 e gates finais cobrirem autorização e segurança.

### Incremental Delivery

1. **US1**: link válido e atualização atômica do manifesto.
2. **US2**: recusa sem `--replace`, substituição explícita e no-op.
3. **US3**: falhas claras, ajuda, status 0/1/2, preservação do manifesto e documentação.
4. **Polish**: quickstart, portabilidade, gates e matriz multiplataforma.

## Notes

- Não adicionar framework CLI, formato JSON, remoto, dependência nova ou prompt interativo.
- Não usar `git add`, `commit`, `checkout`, `reset`, `clean`, `fetch`, `pull` ou `push`.
- Não exibir conteúdo do manifesto, nomes de arquivos do source, remotos, credenciais ou segredos.
- Marcar cada tarefa `[X]` somente após sua validação passar.
