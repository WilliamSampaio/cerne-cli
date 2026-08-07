# Tasks: Instalação de Skills Cerne

**Input**: Design documents from `specs/010-skill-install-command/`

**Prerequisites**: [plan.md](plan.md), [spec.md](spec.md), [research.md](research.md), [data-model.md](data-model.md), [contracts/](contracts/)

**Tests**: Domain behavior and regressions require automated tests. Adapters require contract
tests; critical CLI flows require integration tests; authorization, secrets, and destructive
operations require negative tests. Releases require applicable checks on Linux, Windows, and macOS.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Create the smallest package boundary needed for global skill installation.

- [X] T001 Create `internal/skillinstall/` with empty package files in `internal/skillinstall/install.go`, `internal/skillinstall/package.go`, `internal/skillinstall/resolver.go`, and `internal/skillinstall/targets.go`
- [X] T002 [P] Create test fixture helpers for local `cerne-skills` packages in `internal/skillinstall/package_test.go`
- [X] T003 [P] Add CLI test fixture helper for isolated home/cache paths in `cmd/cerne/main_test.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core validation and filesystem primitives that every story depends on.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [X] T004 [P] Define request, package, manifest, target, managed install, and audit structs in `internal/skillinstall/install.go`
- [X] T005 Implement supported-agent validation for exactly `codex` and `claude` in `internal/skillinstall/targets.go`
- [X] T006 Implement official target resolution for `~/.codex/skills/cerne-context` and `~/.claude/skills/cerne-context` in `internal/skillinstall/targets.go`
- [X] T007 [P] Add target validation tests for supported agents, `generic`, case variants, extra args, and path containment in `internal/skillinstall/targets_test.go`
- [X] T008 Implement companion/cache package placement and lookup contract without network access in `internal/skillinstall/resolver.go`
- [X] T009 [P] Add resolver tests proving the CLI finds only the managed companion/cache package and never scans sibling checkouts or network locations in `internal/skillinstall/resolver_test.go`
- [X] T010 Implement manifest parsing and validation for `cerne-skills`, `cerne-context`, adapter presence, and `contextSchema` v1 in `internal/skillinstall/package.go`
- [X] T011 Add manifest validation tests for valid, malformed, missing skill, missing adapter, and incompatible schema in `internal/skillinstall/package_test.go`
- [X] T012 Implement safe package walk that rejects symlinks and paths escaping the package root in `internal/skillinstall/package.go`
- [X] T013 Add package walk security tests for regular files, directories, symlinks, and path escape entries in `internal/skillinstall/package_test.go`
- [X] T014 Implement private global audit creation/finalization for `skill.install` by reusing or minimally extracting restore audit behavior in `internal/skillinstall/install.go`
- [X] T015 [P] Add audit privacy tests for no skill content, no environment variables, no remotes, and no raw resolver output in `internal/skillinstall/install_test.go`

**Checkpoint**: Foundation ready - user story implementation can now begin.

---

## Phase 3: User Story 1 - Instalar skill Cerne para um agente suportado (Priority: P1) 🎯 MVP

**Goal**: Install or update `cerne-context` for Codex or Claude from the local companion package in the current user's profile.

**Independent Test**: Run `cerne skill install codex` and `cerne skill install claude` with a fixture package and temporary home; verify stdout/status, destination files, idempotency, upgrade, and audit.

### Tests for User Story 1

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation.**

- [X] T016 [US1] Add CLI help and usage tests for `cerne skill --help` and `cerne skill install --help` in `cmd/cerne/main_test.go`
- [X] T017 [US1] Add CLI integration test for successful Codex install stdout, destination, and audit in `cmd/cerne/main_test.go`
- [X] T018 [US1] Add CLI integration test for successful Claude install stdout, destination, and audit in `cmd/cerne/main_test.go`
- [X] T019 [US1] Add domain test for same-version idempotent reinstall in `internal/skillinstall/install_test.go`
- [X] T020 [US1] Add domain test for managed different-version automatic upgrade in `internal/skillinstall/install_test.go`

### Implementation for User Story 1

- [X] T021 [US1] Implement staging copy and final promotion for absent installs in `internal/skillinstall/install.go`
- [X] T022 [US1] Implement managed ownership marker and file list writing in `internal/skillinstall/install.go`
- [X] T023 [US1] Implement same-version idempotency without unnecessary rewrites in `internal/skillinstall/install.go`
- [X] T024 [US1] Implement automatic upgrade for managed different-version installs in `internal/skillinstall/install.go`
- [X] T025 [US1] Add `skill` command parsing, help text, and `install` dispatch in `cmd/cerne/main.go`
- [X] T026 [US1] Wire `cmd/cerne/main.go` to `internal/skillinstall` with stdout/status 0 success messages for installed, already installed, and upgraded outcomes

**Checkpoint**: User Story 1 is independently functional for valid Codex and Claude installs.

---

## Phase 4: User Story 2 - Evitar instalação implícita em fluxos de workspace (Priority: P1)

**Goal**: Preserve the boundary that workspace commands may suggest but never execute global skill installation.

**Independent Test**: Run `init`, `restore`, and `workflow setup` with agent/workflow scenarios and verify global skill destinations remain absent or unchanged.

### Tests for User Story 2

- [X] T027 [US2] Add CLI regression test that `cerne init app --workflow speckit --agent codex` does not create `~/.codex/skills/cerne-context` in `cmd/cerne/main_test.go`
- [X] T028 [US2] Add CLI regression test that `cerne workflow setup --agent claude` does not create `~/.claude/skills/cerne-context` in `cmd/cerne/main_test.go`
- [X] T029 [US2] Add CLI regression test that `cerne restore` does not create Codex or Claude skill destinations in `cmd/cerne/main_test.go`

### Implementation for User Story 2

- [X] T030 [US2] Review and keep `init`, `restore`, and `workflow setup` paths free of calls to `internal/skillinstall` in `cmd/cerne/main.go`
- [X] T031 [US2] Add optional guidance text for missing global skill only where existing command contracts allow it in `cmd/cerne/main.go`

**Checkpoint**: Workspace flows remain independently testable and never perform global skill installation.

---

## Phase 5: User Story 3 - Recusar instalações inseguras ou incompatíveis (Priority: P2)

**Goal**: Refuse invalid usage, unsafe packages, incompatible schemas, unknown destinations, and rollback-risk cases without overwriting user files.

**Independent Test**: Run invalid and unsafe fixture scenarios; verify status/stderr, no audit for usage errors, failed audit for operational errors, and byte-for-byte preservation of existing files.

### Tests for User Story 3

- [X] T032 [P] [US3] Add CLI invalid usage tests for missing agent, `generic`, case variants, repeated agent, and extra args in `cmd/cerne/main_test.go`
- [X] T033 [US3] Add domain test for missing companion package returning operational failure without destination mutation in `internal/skillinstall/install_test.go`
- [X] T034 [US3] Add domain test for unknown destination content preservation in `internal/skillinstall/install_test.go`
- [X] T035 [US3] Add domain test for rollback preserving previous managed install after pre-promotion failure in `internal/skillinstall/install_test.go`
- [X] T036 [US3] Add domain test that audit creation failure stops installation before destination mutation in `internal/skillinstall/install_test.go`
- [X] T037 [US3] Add domain test that audit finalization failure returns status 1 and reports a safe diagnostic in `internal/skillinstall/install_test.go`
- [X] T038 [US3] Add CLI test that operational failures use stderr/status 1 and usage failures use stderr/status 2 with stdout empty in `cmd/cerne/main_test.go`

### Implementation for User Story 3

- [X] T039 [US3] Implement invalid usage handling without audit or filesystem mutation in `cmd/cerne/main.go`
- [X] T040 [US3] Implement unknown destination refusal before overwrite in `internal/skillinstall/install.go`
- [X] T041 [US3] Implement rollback cleanup for staging and preservation of previous managed install in `internal/skillinstall/install.go`
- [X] T042 [US3] Map operational errors to safe Portuguese stderr causes and corrections in `cmd/cerne/main.go`

**Checkpoint**: Unsafe and incompatible install attempts fail safely and preserve user data.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, compatibility, and release checks across all stories.

- [X] T043 [P] Document `cerne skill install <codex|claude>` syntax, destinations, authorization, status codes, audit, and no automatic install in `README.pt-BR.md`
- [X] T044 [P] Update English command documentation for `cerne skill install <codex|claude>` in `README.md`
- [X] T045 [P] Update Spanish command documentation for `cerne skill install <codex|claude>` in `README.es.md`
- [X] T046 [P] Add changelog entry for the new explicit skill install command in `CHANGELOG.md`
- [X] T047 Add companion `cerne-skills` package distribution validation to `.github/workflows/test.yml`
- [X] T048 Add or confirm Linux, Windows, and macOS validation for skill install tests in `.github/workflows/test.yml`
- [X] T049 Run `gofmt` on changed Go files in `cmd/cerne/main.go` and `internal/skillinstall/*.go`
- [X] T050 Run `go test ./...` from the module rooted at `go.mod`
- [X] T051 Run `go vet ./...` from the module rooted at `go.mod`
- [X] T052 Validate quickstart scenarios from `specs/010-skill-install-command/quickstart.md`
- [X] T053 Review `specs/010-skill-install-command/checklists/install-safety.md` and mark passed requirement-quality items

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately.
- **Foundational (Phase 2)**: Depends on Setup completion - blocks all user stories.
- **User Story 1 (Phase 3)**: Depends on Foundational - MVP.
- **User Story 2 (Phase 4)**: Depends on Foundational and can run alongside US1 after CLI parsing is stable.
- **User Story 3 (Phase 5)**: Depends on Foundational and shares install error mapping with US1.
- **Polish (Phase 6)**: Depends on the selected user stories being implemented.

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational; delivers the MVP command.
- **User Story 2 (P1)**: Can start after Foundational; independent regression boundary for workspace flows.
- **User Story 3 (P2)**: Can start after Foundational; hardens failure and refusal paths.

### Within Each User Story

- Tests MUST be written and fail before implementation.
- Domain tests before domain implementation.
- CLI integration tests before command dispatch/output implementation.
- Safe filesystem behavior before documentation claims completion.

### Parallel Opportunities

- T002 and T003 can run in parallel.
- T004 and T007 can run in parallel where they touch separate files.
- T009 and T015 can run in parallel after their target files exist.
- T032 can run in parallel with the US3 domain tests.
- T043 through T046 can run in parallel after command behavior stabilizes.

---

## Execution Example: User Story 1

```bash
Task: "Add CLI integration test for successful Codex install stdout, destination, and audit in cmd/cerne/main_test.go"
Task: "Add CLI integration test for successful Claude install stdout, destination, and audit in cmd/cerne/main_test.go"
Task: "Add domain test for same-version idempotent reinstall in internal/skillinstall/install_test.go"
Task: "Add domain test for managed different-version automatic upgrade in internal/skillinstall/install_test.go"
```

---

## Execution Example: User Story 3

```bash
Task: "Add CLI invalid usage tests for missing agent, generic, case variants, repeated agent, and extra args in cmd/cerne/main_test.go"
Task: "Add domain test for missing companion package returning operational failure without destination mutation in internal/skillinstall/install_test.go"
Task: "Add domain test for unknown destination content preservation in internal/skillinstall/install_test.go"
Task: "Add domain test for rollback preserving previous managed install after pre-promotion failure in internal/skillinstall/install_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup.
2. Complete Phase 2: Foundational.
3. Complete Phase 3: User Story 1.
4. Stop and validate `cerne skill install codex` and `cerne skill install claude` with local fixtures.

### Incremental Delivery

1. Foundation ready.
2. Add US1 valid install/idempotency/upgrade.
3. Add US2 workspace non-installation regression tests.
4. Add US3 refusals, rollback, and safe diagnostics.
5. Finish docs and quickstart validation.

### Lazy Boundaries

- Keep `codex` and `claude` as explicit mappings; add a registry only when a third public agent exists.
- Keep package resolution as one local companion/cache resolver seam; add remote release download only when a feature requires it.
- Keep tests on `t.TempDir()` homes and local package fixtures; no real agent profiles, network, credentials, or releases.

---

## Phase 7: Convergence

- [X] T054 CRITICAL harden `internal/skillinstall` audit directory/file privacy by reusing or extracting restore-style owner/permission and Windows DACL checks, with tests, per Constitution VI and FR-016 (partial)
- [X] T055 Preserve validated package version in failed `skill.install` audit finalization after package load, with regression coverage for post-validation failures, per FR-016 and audit contract (partial)
- [X] T056 Validate existing Cerne ownership marker before idempotency or upgrade, rejecting wrong agent, empty or invalid version, unsafe managed file paths, and malformed ownership evidence, per FR-014 and ManagedInstallation (partial)
- [X] T057 Validate `cerne-skills` package version as semver during manifest loading, with invalid-version tests, per SkillPackage model and FR-007 (partial)
- [X] T058 Add package path-escape fixture coverage for manifest `source` escaping the package root, per T013 and SC-005 (partial)
