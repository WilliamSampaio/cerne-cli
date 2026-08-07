# Tasks: Descoberta de Agente para Spec Kit

**Input**: Design documents from `/specs/009-speckit-agent-discovery/`

**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md,
data-model.md, contracts/, quickstart.md

**Tests**: Domain behavior and regressions require automated tests. Adapters require contract
tests; critical CLI flows require integration tests; authorization, secrets, and destructive
operations require negative tests. Releases require applicable checks on Linux, Windows, and
macOS.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Prepare shared contracts and fixtures used by every story.

- [X] T001 [P] Add CLI regression expectations for `--agent` usage strings in `cmd/cerne/main_test.go`
- [X] T002 [P] Add agent target and bridge command-set contract tests in `internal/workflowexec/setup_test.go`
- [X] T003 [P] Add local discovery bridge domain test scaffolding in `internal/workspace/workflow_test.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core parser, adapter, and domain model that MUST be complete before user stories.

**CRITICAL**: No user story work can begin until this phase is complete.

- [X] T004 Extend `initArguments` and `parseInitArgs` to parse optional `--agent <codex|claude>` only with `--workflow speckit` in `cmd/cerne/main.go`
- [X] T005 Extend `runWorkflow` parsing to accept `cerne workflow setup --agent <codex|claude>` in `cmd/cerne/main.go`
- [X] T006 Add agent target description fields and resolver helpers for Codex and Claude in `internal/workflowexec/setup.go`
- [X] T007 Add local discovery request/result types to workflow setup domain flow in `internal/workspace/workflow.go`
- [X] T008 Add bridge path validation and managed command-set constants without symlink dependency in `internal/workspace/workflow.go`

**Checkpoint**: CLI can parse agent intent and domain/adapter can describe supported local agents.

---

## Phase 3: User Story 1 - Usar Spec Kit no Codex a partir da raiz do workspace (Priority: P1) MVP

**Goal**: `cerne init projeto --workflow speckit --agent codex` leaves Spec Kit valid in
`knowledge` and Codex commands discoverable at the workspace root.

**Independent Test**: Initialize a disposable workspace with a controlled Spec Kit provider and
verify root `.agents/skills/speckit-*/SKILL.md`, `knowledge` as Spec Kit root, unchanged `source`,
no agent field in `knowledge/cerne.json`, and documented stdout/stderr/status.

### Tests for User Story 1

- [X] T009 [P] [US1] Add CLI integration test for `init --workflow speckit --agent codex` success in `cmd/cerne/main_test.go`
- [X] T010 [P] [US1] Add domain test for Codex root bridge creation and no source mutation in `internal/workspace/workflow_test.go`
- [X] T011 [P] [US1] Add adapter invocation test for Spec Kit Codex integration install arguments in `internal/workflowexec/setup_test.go`
- [X] T012 [P] [US1] Add manifest regression assertion that `workflow.agent` is never written in `internal/workspace/init_test.go`

### Implementation for User Story 1

- [X] T013 [US1] Wire `runInit` to pass agent discovery requests through source and no-source workflow paths in `cmd/cerne/main.go`
- [X] T014 [US1] Extend workflow execution adapter to prepare Codex integration in `knowledge` using the official Spec Kit integration in `internal/workflowexec/setup.go`
- [X] T015 [US1] Implement Codex local bridge creation at workspace root `.agents/skills` in `internal/workspace/workflow.go`
- [X] T016 [US1] Render `Agent: codex` and `Descoberta: pronta` success lines and pending correction text in `cmd/cerne/main.go`
- [X] T017 [US1] Preserve provider failure, audit, cleanup, and no-fake-bridge behavior for Codex init in `internal/workspace/workflow.go`

**Checkpoint**: User Story 1 is fully functional and testable independently.

---

## Phase 4: User Story 2 - Restaurar knowledge e escolher outro agente local (Priority: P2)

**Goal**: `cerne workflow setup --agent claude` can prepare a local Claude bridge after restore or
after a previous Codex choice without persisting the agent.

**Independent Test**: Start from a restored or already-ready Spec Kit workspace, run setup with
Claude, and verify root `.claude/skills`, unchanged manifest/source, and correct stdout/stderr/status.

### Tests for User Story 2

- [X] T018 [P] [US2] Add CLI integration test for `workflow setup --agent claude` on ready workflow in `cmd/cerne/main_test.go`
- [X] T019 [P] [US2] Add domain test for changing local discovery from Codex to Claude without modifying manifest or source in `internal/workspace/workflow_test.go`
- [X] T020 [P] [US2] Add adapter invocation test for Claude integration install arguments in `internal/workflowexec/setup_test.go`
- [X] T021 [P] [US2] Add audit test for agent integration subprocess records in `internal/workspace/workflow_test.go`
- [X] T022 [P] [US2] Add missing-provider test proving setup with agent creates no fake bridge in `cmd/cerne/main_test.go`

### Implementation for User Story 2

- [X] T023 [US2] Wire `runWorkflow` to pass optional agent discovery requests to `workspace.SetupWorkflow` in `cmd/cerne/main.go`
- [X] T024 [US2] Extend workflow setup domain flow to refresh only local discovery when provider layout is already ready in `internal/workspace/workflow.go`
- [X] T025 [US2] Extend workflow execution adapter to prepare Claude integration in `knowledge` in `internal/workflowexec/setup.go`
- [X] T026 [US2] Implement Claude local bridge creation at workspace root `.claude/skills` in `internal/workspace/workflow.go`
- [X] T027 [US2] Render setup agent success and operational failure output for `workflow setup --agent` in `cmd/cerne/main.go`

**Checkpoint**: User Stories 1 and 2 work independently.

---

## Phase 5: User Story 3 - Manter comportamento legado sem agente (Priority: P3)

**Goal**: Existing `cerne init --workflow speckit` and `cerne workflow setup` behavior remains
compatible when `--agent` is omitted.

**Independent Test**: Run the existing workflow init/setup tests and verify no workspace-root
agent bridge appears and existing stdout/stderr/status remain unchanged.

### Tests for User Story 3

- [X] T028 [P] [US3] Add regression test that `init --workflow speckit` creates no root `.agents` or `.claude` bridge in `cmd/cerne/main_test.go`
- [X] T029 [P] [US3] Add invalid usage tests for agent without Spec Kit, OpenSpec with agent, unknown agent, repeated agent, and extra workflow args in `cmd/cerne/main_test.go`
- [X] T030 [P] [US3] Add domain regression that setup without agent remains provider-only and idempotent in `internal/workspace/workflow_test.go`

### Implementation for User Story 3

- [X] T031 [US3] Preserve legacy no-agent setup paths and stdout/stderr contracts in `cmd/cerne/main.go`
- [X] T032 [US3] Ensure bridge creation is skipped when agent request is empty in `internal/workspace/workflow.go`
- [X] T033 [US3] Keep generic Spec Kit behavior as internal fallback without exposing `generic` as public agent in `internal/workflowexec/setup.go`

**Checkpoint**: All user stories are independently functional.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, portability, and final validation.

- [X] T034 [P] Update `--agent` help, side effects, examples, and compatibility notes in `README.md`
- [X] T035 [P] Update Portuguese `--agent` documentation in `README.pt-BR.md`
- [X] T036 [P] Update Spanish `--agent` documentation in `README.es.md`
- [X] T037 [P] Add bridge content safety test proving no absolute paths, env, remotes, tokens, provider output, or private knowledge content in `internal/workspace/workflow_test.go`
- [X] T038 Verify Linux, Windows, and macOS bridge path behavior through existing CI matrix in `.github/workflows/test.yml`
- [X] T039 Run `gofmt` on changed Go files
- [X] T040 Run `go test ./...`
- [X] T041 Run `go vet ./...`
- [X] T042 Run `git diff --check`
- [X] T043 Validate quickstart scenarios from `specs/009-speckit-agent-discovery/quickstart.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies.
- **Foundational (Phase 2)**: Depends on Setup completion and blocks all user stories.
- **User Story 1 (Phase 3)**: Depends on Foundational; MVP.
- **User Story 2 (Phase 4)**: Depends on Foundational and can reuse US1 bridge helpers, but remains testable from a ready workspace.
- **User Story 3 (Phase 5)**: Depends on Foundational and can run after or alongside story implementation to protect compatibility.
- **Polish (Phase 6)**: Depends on desired user stories being complete.

### User Story Dependencies

- **US1**: No dependency on other stories after Foundational.
- **US2**: No business dependency on US1, but may reuse bridge helper implementation from US1.
- **US3**: No business dependency; protects legacy behavior while other stories add agent support.

### Within Each User Story

- Tests MUST be written and fail before implementation.
- Adapter contract tests before adapter changes.
- CLI integration tests before CLI output changes.
- Domain bridge tests before bridge creation changes.
- Story complete before treating the next priority as done.

---

## Parallel Opportunities

- T001, T002, T003 can run in parallel.
- T009, T010, T011, T012 can run in parallel.
- T018, T019, T020, T021, T022 can run in parallel.
- T028, T029, T030 can run in parallel.
- T034, T035, T036, T037 can run in parallel.

## Parallel Example: User Story 1

```text
Task: "Add CLI integration test for init --workflow speckit --agent codex in cmd/cerne/main_test.go"
Task: "Add domain test for Codex root bridge creation in internal/workspace/workflow_test.go"
Task: "Add adapter invocation test for Codex integration install arguments in internal/workflowexec/setup_test.go"
Task: "Add manifest regression assertion in internal/workspace/init_test.go"
```

## Parallel Example: User Story 2

```text
Task: "Add CLI integration test for workflow setup --agent claude in cmd/cerne/main_test.go"
Task: "Add domain test for changing local discovery in internal/workspace/workflow_test.go"
Task: "Add adapter invocation test for Claude integration install arguments in internal/workflowexec/setup_test.go"
Task: "Add audit test for agent integration subprocess records in internal/workspace/workflow_test.go"
Task: "Add missing-provider test in cmd/cerne/main_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1 and Phase 2.
2. Complete Phase 3 for Codex init.
3. Stop and validate Scenario 2 from `quickstart.md`.

### Incremental Delivery

1. MVP Codex init.
2. Add `workflow setup --agent` and Claude bridge.
3. Lock legacy no-agent compatibility.
4. Finish documentation and validation.

### Keep It Small

Use existing parser style, existing workflow result flow, and the standard library. Add helper
types/functions only where they avoid duplicating bridge/path logic across init and setup.
