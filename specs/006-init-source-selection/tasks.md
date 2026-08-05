# Tasks: Seleção de Source no Init

**Input**: Design documents from `specs/006-init-source-selection/`

**Prerequisites**: `plan.md`, `spec.md`, `research.md`, `data-model.md`, `contracts/`, `quickstart.md`

**Tests**: Domain behavior and regressions require automated tests. The Git clone adapter requires
contract tests; critical CLI flows require integration tests; authorization, secrets, cleanup and
external-source preservation require negative tests. Tests use local temporary repositories and
fake Git executables only.

**Organization**: Tasks are grouped by user story so each source-selection mode can be validated
independently.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel because it changes different files and has no incomplete dependency
- **[Story]**: User story served by the task
- Every task names its exact file path

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Reuse the current Go module, manual CLI dispatcher, Git adapter and CI matrix.

No setup task is required: the repository already contains all required build, test and adapter
infrastructure, and the plan adds no dependency or package.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Establish exact parsing for the three init shapes before any filesystem or Git work.

**⚠️ CRITICAL**: Complete this phase before any user story.

- [ ] T001 Write failing table-driven tests for default, `--source`, `--clone`, missing values, mutual exclusion, repetition, ordering and extra arguments before Git lookup or filesystem effects in `cmd/cerne/main_test.go`
- [ ] T002 Implement the exact-shape manual init parser and source-mode request without changing legacy dispatch in `cmd/cerne/main.go`

**Checkpoint**: Every accepted or rejected invocation has a deterministic request and status-two contract.

---

## Phase 3: User Story 1 - Inicializar com source local existente (Priority: P1) 🎯 MVP

**Goal**: Create knowledge and link an existing local working tree atomically without creating an
internal source or changing the external repository.

**Independent Test**: Initialize from a temporary repository with history through relative,
absolute and worktree paths; verify manifest, streams, Git separation, no internal source and an
identical external snapshot.

### Tests for User Story 1

- [ ] T003 [P] [US1] Write failing domain tests for local path resolution, non-bare root/worktree acceptance, unsafe overlap refusal, portable manifest path, pre-success revalidation, no internal source, byte-preserved external repository and unchanged `link` behavior in `internal/workspace/init_test.go` and `internal/workspace/link_test.go`
- [ ] T004 [P] [US1] Write failing CLI integration tests for `--source`, exact success/failure streams, spaces/Unicode, Git absence and exact no-flag regression in `cmd/cerne/main_test.go`

### Implementation for User Story 1

- [ ] T005 [US1] Refactor only the reusable path, repository and separation checks from `internal/workspace/link.go` and extend init request handling for default/local modes in `internal/workspace/init.go`
- [ ] T006 [US1] Connect the existing link inspector to init, render `Source vinculado` and preserve byte-for-byte legacy output in `cmd/cerne/main.go`

**Checkpoint**: Local source init works in one invocation and default init remains exact.

---

## Phase 4: User Story 2 - Inicializar clonando um repositório (Priority: P2)

**Goal**: Clone one allowed origin into internal source with complete history, stable remote and a
redacted audit created before Git.

**Independent Test**: Clone populated and empty temporary local origins using real Git, then use a
fake Git executable to verify exact argv, environment, protocol controls, hooks/templates,
noninteraction and absence of origin/output in returned data.

### Tests for User Story 2

- [ ] T007 [P] [US2] Write failing adapter contract tests for allowed/rejected origin forms, embedded credentials, exact clone argv, `--` injection protection, `origin`, `--no-local`, protocol allowlist, empty hooks/templates, prompt controls and redacted errors in `internal/gitexec/init_test.go`
- [ ] T008 [P] [US2] Write failing domain tests for populated/empty clone success through private staging, promotion without remnants, unchanged origin, manifest, independent Git roots, complete history/remotes, `started` before callback and atomic succeeded audit excluding `runs/.gitkeep` in `internal/workspace/init_test.go`
- [ ] T009 [P] [US2] Write failing CLI integration tests for successful populated/empty clone, exact output, absent origin in streams/manifest/audit and acceptance by doctor/status in `cmd/cerne/main_test.go`

### Implementation for User Story 2

- [ ] T010 [US2] Implement standard-library origin classification and secure shell-free `FindClone` execution beside the existing Git init adapter in `internal/gitexec/init.go`
- [ ] T011 [US2] Implement clone mode happy path with private staging, non-replacing promotion, SHA-256 origin fingerprint, fixed atomic audit lifecycle, cloned-repository validation and transient result fields in `internal/workspace/init.go`
- [ ] T012 [US2] Connect the clone adapter, render `Source clonado` and map safe operational failures in `cmd/cerne/main.go`

**Checkpoint**: Both populated and empty local origins clone in one invocation without network fixtures or secret exposure.

---

## Phase 5: User Story 3 - Manter compatibilidade e falhar com segurança (Priority: P3)

**Goal**: Reject unsafe inputs before effects and preserve knowledge/audit while removing only
private staging after a real clone failure, without replacing a concurrent source.

**Independent Test**: Inject process, validation, audit-finalization and cleanup failures; compare
all preexisting bytes, exact streams/status and final filesystem/audit state.

### Tests for User Story 3

- [ ] T013 [P] [US3] Write failing adapter tests proving disallowed transports never execute and token-like origin, environment and Git output never enter errors or results in `internal/gitexec/init_test.go`
- [ ] T014 [P] [US3] Write failing domain tests for pre-clone full rollback, clone failure preserving knowledge/failed audit, staging-only cleanup, private-directory cleanup failure, invalid clone result, concurrent source blocking promotion and post-promotion audit-finalization failure preserving valid source with `started` in `internal/workspace/init_test.go`
- [ ] T015 [P] [US3] Write failing CLI tests for unsafe origin, authentication/process failure, incomplete-workspace correction, missing source diagnosis, unchanged existing commands and exact legacy help/output/status in `cmd/cerne/main_test.go`

### Implementation for User Story 3

- [ ] T016 [US3] Complete the post-audit rollback boundary with staging-only failure cleanup, concurrent-source promotion refusal, safe failure categories and post-promotion inconclusive-audit handling in `internal/workspace/init.go`
- [ ] T017 [US3] Complete CLI failure rendering and corrective guidance without exposing origin or Git output in `cmd/cerne/main.go`

**Checkpoint**: Every refusal and failure path is noninteractive, auditable when Git ran, secret-safe and limited to owned paths.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Stabilize documentation, release identity and three-platform acceptance.

- [ ] T018 [P] Document all three init modes, source layouts, transports, authentication, filters, remotes, rollback, audit, streams and examples in `README.md`
- [ ] T019 [P] Mirror the complete init source-selection contract in `README.pt-BR.md` and `README.es.md`
- [ ] T020 Update global/init help, exact help tests, compatible MINOR version and release notes in `cmd/cerne/main.go`, `cmd/cerne/main_test.go` and `CHANGELOG.md`
- [ ] T021 Validate all twelve scenarios from `specs/006-init-source-selection/quickstart.md` with local origins/fake Git and record any corrections in that file
- [ ] T022 Run `gofmt` on changed Go files, `go vet ./...`, `go test -count=1 ./...` and `git diff --check`, confirming `.github/workflows/test.yml` exercises Linux, Windows and macOS without network or credentials

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1**: No work; existing infrastructure is reused.
- **Phase 2**: Starts immediately and blocks every story.
- **US1 (Phase 3)**: Depends on Phase 2 and is the MVP.
- **US2 (Phase 4)**: Depends on Phase 2; extends the same init request but is independently testable.
- **US3 (Phase 5)**: Depends on the clone path from US2 for post-execution hardening.
- **Polish (Phase 6)**: Depends on all selected stories.

### User Story Dependencies

- **US1**: No story dependency after the parser foundation.
- **US2**: No behavioral dependency on US1, but shares the request shape and init creation helper.
- **US3**: Requires US2 because it hardens failures after clone actually starts.

### Within Each User Story

- Write the listed tests first and confirm they fail for the intended behavior.
- Implement domain behavior before CLI rendering.
- Run real Git only against temporary local repositories.
- Complete each checkpoint before continuing in a single-developer flow.

## Requirement Coverage

| Requirements | Tasks |
| --- | --- |
| FR-001–FR-007 | T001–T006 |
| FR-008–FR-014, FR-018, FR-022 | T007–T012 |
| FR-015–FR-017, FR-019–FR-021, FR-023, FR-025–FR-026 | T013–T017, T021–T022 |
| FR-024 | T018–T020 |
| SC-001–SC-008 | T003–T022 |

## Parallel Opportunities

- T003 and T004 can run together after Phase 2.
- T007, T008 and T009 can run together after their contracts stabilize.
- T013, T014 and T015 can run together after US2.
- T018 and T019 can run in parallel after public behavior stabilizes.

## Parallel Example: User Story 2

```text
Task: "T007 clone adapter contract tests in internal/gitexec/init_test.go"
Task: "T008 clone domain and audit tests in internal/workspace/init_test.go"
Task: "T009 clone CLI integration tests in cmd/cerne/main_test.go"
```

## Implementation Strategy

### MVP First

1. Complete Phase 2.
2. Complete US1.
3. Verify local source, worktree and exact legacy init independently.
4. Stop before clone if an early review is desired.

### Incremental Delivery

1. Parser foundation → exact requests without effects.
2. US1 → atomic local-source init.
3. US2 → successful audited clone.
4. US3 → refusal, cleanup and failure hardening.
5. Polish → documentation, version and cross-platform acceptance.

## Notes

- No task adds a Go dependency, CLI framework, host SDK, global state, interface, provider registry or retry command.
- No task accepts arbitrary Git flags, embedded credentials or unknown transports.
- Existing user changes must be preserved; implementation may touch only the listed files.
- Do not commit, push, publish or deploy as part of this task list unless separately authorized.
