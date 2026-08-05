# Tasks: Inicialização com Workflow SDD

**Input**: Design documents from `specs/005-workflow-init/`

**Prerequisites**: `plan.md`, `spec.md`, `research.md`, `data-model.md`, `contracts/`, `quickstart.md`

**Tests**: Domain behavior and regressions require automated tests. The provider adapter requires
contract tests; critical CLI flows require integration tests; audit, secrets, cleanup and source
isolation require negative tests. CI must remain deterministic without real providers or network.

**Organization**: Tasks are grouped by user story so each increment can be tested independently.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel because it changes different files and has no incomplete dependency
- **[Story]**: User story served by the task
- Every task names its exact file path

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Reuse the existing Go module, manual CLI dispatcher and three-platform CI.

No setup task is required: the project already has every build and test facility needed, and the
plan adds no dependency or generated scaffold.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Establish the provider-neutral domain state, audit contract, manifest parsing and
local process adapter used by every story.

**⚠️ CRITICAL**: Complete this phase before any user story.

- [ ] T001 [P] Write failing domain tests for applying a generic workflow definition, its canonical specs path, absent/ready/partial layouts and started/final audit records without provider-specific branches in `internal/workspace/workflow_test.go`
- [ ] T002 Implement generic workflow-definition consumption, layout classification and atomic redacted audit lifecycle without Spec Kit/OpenSpec constants in `internal/workspace/workflow.go`
- [ ] T003 [P] Write failing adapter contract tests for resolving Spec Kit/OpenSpec/unknown identifiers into exact definitions, executable discovery, exact arguments, platform script choice, cwd, allowlisted environment, disabled telemetry and sanitized failures in `internal/workflowexec/setup_test.go`
- [ ] T004 Implement the closed provider resolver, generic definitions, PATH discovery and shell-free execution with fixed arguments and minimal environment in `internal/workflowexec/setup.go`
- [ ] T005 [P] Write failing manifest tests for absent, valid, malformed and opaque optional workflow identifiers, acceptance by status and preservation by link in `internal/workspace/doctor_test.go`, `internal/workspace/status_test.go` and `internal/workspace/link_test.go`
- [ ] T006 Extend manifest decoding with an optional nonempty opaque workflow identifier, deferring provider support validation to the adapter resolver while keeping legacy manifests valid in `internal/workspace/doctor.go`

**Checkpoint**: Providers, layouts, audit and manifests have deterministic contracts without a user-facing workflow yet.

---

## Phase 3: User Story 1 - Criar workspace com workflow escolhido (Priority: P1) 🎯 MVP

**Goal**: Initialize a new workspace with Spec Kit or OpenSpec already materialized in knowledge,
while legacy init remains byte-for-byte unchanged.

**Independent Test**: Inject successful local provider callbacks, run both documented init forms and
verify manifest, conditional layout, invocation, audit, Git separation, exact streams and legacy regression.

### Tests for User Story 1

- [ ] T007 [P] [US1] Write failing domain tests for legacy init and two opaque identifiers resolved to fake generic definitions, conditional knowledge directories, successful setup audit and untouched source in `internal/workspace/init_test.go`
- [ ] T008 [P] [US1] Write failing CLI integration tests for both workflow values, exact success output, argument position and exact legacy init regression in `cmd/cerne/main_test.go`

### Implementation for User Story 1

- [ ] T009 [US1] Extend workspace creation to persist the opaque identifier, apply its resolved generic definition only after the base Git repositories exist, validate the supplied marker and represent configured or pending state in `internal/workspace/init.go`
- [ ] T010 [US1] Parse `--workflow`, connect the process adapter, render configured success and preserve the no-flag path in `cmd/cerne/main.go`

**Checkpoint**: With an available controlled provider, each workflow initializes in one invocation; legacy init remains unchanged.

---

## Phase 4: User Story 2 - Concluir init quando a ferramenta estiver ausente (Priority: P2)

**Goal**: Preserve a pending workspace when the provider is absent and allow one-command,
idempotent setup after installation, with doctor warnings.

**Independent Test**: Initialize with provider discovery returning absent, then supply a successful
provider and run setup from nested workspace directories twice; verify warning, completion, no-op and doctor states.

### Tests for User Story 2

- [ ] T011 [P] [US2] Write failing domain tests for init with absent executable, locating pending workspaces, successful resume and ready no-op without a second audit in `internal/workspace/init_test.go` and `internal/workspace/workflow_test.go`
- [ ] T012 [P] [US2] Write failing doctor tests for pending, ready and ready-with-executable-missing states while preserving the exact legacy ten-check report in `internal/workspace/doctor_test.go`
- [ ] T013 [P] [US2] Write failing CLI tests for non-blocking init warning, `workflow setup` success/no-op/help, nested invocation and exact status/stream contracts in `cmd/cerne/main_test.go`

### Implementation for User Story 2

- [ ] T014 [US2] Preserve the initialized workspace and pending result when provider discovery is absent, then implement ancestor location and idempotent resume using the declared provider in `internal/workspace/init.go` and `internal/workspace/workflow.go`
- [ ] T015 [US2] Add the conditional workflow check through the generic resolver and availability callback without changing legacy doctor output in `internal/workspace/doctor.go`
- [ ] T016 [US2] Add `workflow setup` dispatch, help, pending init warning, success/no-op rendering and corrective errors in `cmd/cerne/main.go`

**Checkpoint**: Missing optional tools no longer block workspace creation and setup can be resumed exactly once.

---

## Phase 5: User Story 3 - Receber falhas seguras e previsíveis (Priority: P3)

**Goal**: Reject invalid usage and partial layouts, preserve prior data after provider failures,
keep an audit trail and prevent source, secrets or raw provider output from leaking.

**Independent Test**: Use controlled providers that fail, emit token-like output, create partial
owned roots or invalid Git metadata; compare all preexisting files and source before/after.

### Tests for User Story 3

- [ ] T017 [P] [US3] Write failing domain tests for init provider failure and conditional rollback, missing marker after success, partial preexisting layout, nested Git, audit-finalization failure leaving `started`, conservative cleanup and byte-preserved prior files in `internal/workspace/init_test.go` and `internal/workspace/workflow_test.go`
- [ ] T018 [P] [US3] Write failing adapter tests proving credential-like environment variables and raw token-like output never reach the subprocess result or diagnostics in `internal/workflowexec/setup_test.go`
- [ ] T019 [P] [US3] Write failing doctor tests for malformed workflow identifier, identifier unresolved by the adapter, partial generic marker and nested Git as blocking findings in `internal/workspace/doctor_test.go`
- [ ] T020 [P] [US3] Write failing CLI tests for invalid init flag forms, setup without declaration, missing/incompatible executable, provider failure, safe stderr and preserved base workspace in `cmd/cerne/main_test.go`

### Implementation for User Story 3

- [ ] T021 [US3] Preserve the base workspace after an executed provider fails while retaining pre-execution rollback, then complete safe failure categorization, marker validation, nested-Git refusal, audit finalization and owned-root-only cleanup in `internal/workspace/init.go` and `internal/workspace/workflow.go`
- [ ] T022 [US3] Complete blocking workflow diagnoses and corrections without exposing provider output in `internal/workspace/doctor.go`
- [ ] T023 [US3] Complete usage parsing and operational error rendering for all invalid and failed workflow paths in `cmd/cerne/main.go`

**Checkpoint**: All refusal and provider-failure scenarios are recoverable, auditable and leave source untouched.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Stabilize public documentation, release identity and cross-platform acceptance.

- [ ] T024 [P] Document workflow selection, layouts, optional dependencies, recovery, audit and effects in `README.md`
- [ ] T025 [P] Mirror the command and safety documentation in `README.pt-BR.md` and `README.es.md`
- [ ] T026 Update global/init/doctor/workflow help contracts and bump the compatible MINOR version with exact regression expectations in `cmd/cerne/main.go` and `cmd/cerne/main_test.go`
- [ ] T027 Validate all twelve scenarios from `specs/005-workflow-init/quickstart.md`, using controlled providers for failure/security cases and recording any corrections in that file
- [ ] T028 Run `gofmt` on changed Go files, `go vet ./...`, `go test -count=1 ./...` and `git diff --check`, confirming `.github/workflows/test.yml` covers Linux, Windows and macOS without real providers

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1**: No work; existing infrastructure is reused.
- **Phase 2**: Starts immediately and blocks every story.
- **US1 (Phase 3)**: Depends on Phase 2 and is the MVP.
- **US2 (Phase 4)**: Depends on Phase 2 and integrates the init state produced by US1.
- **US3 (Phase 5)**: Depends on Phase 2; may proceed beside US1/US2 after shared contracts stabilize.
- **Polish (Phase 6)**: Depends on all selected stories.

### User Story Dependencies

- **US1**: No story dependency after the foundation.
- **US2**: Independently testable with a manifest fixture, but final integration extends US1's init output.
- **US3**: Independently testable with workspace/provider fixtures; hardens both earlier stories.

### Within Each User Story

- Write the listed tests first and confirm they fail for the intended behavior.
- Implement domain behavior before connecting CLI rendering.
- Do not run a real provider in automated tests.
- Complete the story checkpoint before moving to the next priority in a single-developer flow.

## Requirement Coverage

| Requirements | Tasks |
| --- | --- |
| FR-001–FR-013 | T001–T010 |
| FR-014–FR-018, FR-023–FR-024 | T011–T016 |
| FR-019–FR-022, FR-025–FR-030, FR-033–FR-034 | T017–T023 |
| FR-031–FR-032 and SC-001–SC-009 | T024–T028 |

## Parallel Opportunities

- T001, T003 and T005 touch independent packages/files and can begin together.
- T007 and T008 can run together after Phase 2.
- T011, T012 and T013 can run together; so can T017–T020.
- US1, US2 fixtures and US3 hardening can be developed in parallel after Phase 2 when merge order is coordinated.
- T024 and T025 can run in parallel after public contracts stabilize.

## Parallel Example: User Story 2

```text
Task: "T011 domain tests for pending init/resume/idempotency in internal/workspace/init_test.go and internal/workspace/workflow_test.go"
Task: "T012 doctor workflow warning tests in internal/workspace/doctor_test.go"
Task: "T013 CLI workflow setup contract tests in cmd/cerne/main_test.go"
```

## Implementation Strategy

### MVP First

1. Complete Phase 2.
2. Complete US1.
3. Verify Spec Kit, OpenSpec and legacy init independently.
4. Stop before recovery/hardening if an early review is desired.

### Incremental Delivery

1. Foundation → provider-neutral types, audit, manifest and adapter contracts.
2. US1 → successful opt-in workflow initialization.
3. US2 → missing-provider recovery and idempotent setup.
4. US3 → failure cleanup, blocking diagnostics and secret safety.
5. Polish → documentation, version and full three-platform validation.

## Notes

- No task installs Spec Kit, OpenSpec, Node, Python, uv or package managers.
- No task adds a Go dependency, provider registry, workflow conversion or agent selection.
- Existing user changes must be preserved; implementation tasks may touch only the listed files.
- Do not commit, push, publish or deploy as part of this task list unless separately authorized.
