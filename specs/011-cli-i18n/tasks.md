# Tasks: CLI Internationalization

**Input**: Design documents from `/specs/011-cli-i18n/`

## Phase 1: Setup

- [x] T001 Define the language/message API and empty parity-checked catalogs in `internal/localization/language.go`, `cmd/cerne/messages_en.go`, and `cmd/cerne/messages_ptbr.go`
- [x] T002 Create the user-configuration package and platform replacement files in `internal/localization/config.go`, `internal/localization/atomic_replace_unix.go`, and `internal/localization/atomic_replace_windows.go`

---

## Phase 2: Foundational

- [x] T003 [P] Add supported-language parsing and catalog parity tests in `internal/localization/language_test.go` and `cmd/cerne/messages_test.go`
- [x] T004 Implement immutable per-invocation message lookup and formatting in `internal/localization/language.go` and `cmd/cerne/messages.go`
- [x] T005 Replace user-visible prose as domain identity with stable failure/check codes in `internal/workspace/*.go` and `internal/skillinstall/install.go`

**Checkpoint**: Domain results and presentation messages can vary independently.

---

## Phase 3: User Story 1 - Persist a Preferred Language (Priority: P1)

**Goal**: Let users set, inspect, use, and remove a language preference in their home directory.

**Independent Test**: Set each supported language in an isolated home, run help in a separate invocation, inspect the preference, and unset it back to the compatibility default.

- [x] T006 [P] [US1] Add valid, absent, malformed, atomic-write, permission, and symlink refusal tests in `internal/localization/config_test.go`
- [x] T007 [US1] Implement safe read, set, and unset behavior for `~/.cerne/config.json` in `internal/localization/config.go` and platform replacement files
- [x] T008 [P] [US1] Add exact stdout, stderr, exit-status, and filesystem integration tests for `cerne config` in `cmd/cerne/main_test.go`
- [x] T009 [US1] Implement `cerne config set|get|unset language` routing and localized output in `cmd/cerne/main.go`
- [x] T010 [US1] Localize global and config help in `cmd/cerne/messages_en.go` and `cmd/cerne/messages_ptbr.go`

**Checkpoint**: Persistent language selection works independently with localized global/config help.

---

## Phase 4: User Story 2 - Override the Language Temporarily (Priority: P2)

**Goal**: Select a language for one invocation or environment without changing the saved preference.

**Independent Test**: Save `pt-BR`, invoke commands through `--lang en` and `CERNE_LANG=en`, and confirm the preference remains `pt-BR`.

- [x] T011 [P] [US2] Add precedence, invalid-value, and non-persistence tests in `internal/localization/language_test.go`
- [x] T012 [P] [US2] Add global `--lang` and `CERNE_LANG` CLI integration tests in `cmd/cerne/main_test.go`
- [x] T013 [US2] Implement global option parsing and flag/environment/config/default resolution in `cmd/cerne/main.go` and `internal/localization/language.go`

**Checkpoint**: Temporary selection follows `--lang > CERNE_LANG > saved preference > pt-BR`.

---

## Phase 5: User Story 3 - Preserve Automation Contracts (Priority: P3)

**Goal**: Localize every human-facing command path while preserving machine-readable behavior.

**Independent Test**: Exercise help, success, warning, usage, and operational failures for every command in both languages and compare JSON plus exit statuses.

- [x] T014 [P] [US3] Add English and Brazilian Portuguese rendering tests for workspace reports and failures in `cmd/cerne/main_test.go`
- [x] T015 [P] [US3] Add byte-identical `context --json`, neutral version, command-name, flag-name, and exit-status tests in `cmd/cerne/main_test.go`
- [x] T016 [US3] Complete all command help and presentation messages in `cmd/cerne/messages_en.go` and `cmd/cerne/messages_ptbr.go`
- [x] T017 [US3] Route skill, context, restore, init, workflow, doctor, status, and link rendering through the effective catalog in `cmd/cerne/main.go`
- [x] T018 [US3] Remove remaining localized prose from domain decisions and expose semantic values in `internal/workspace/*.go`, `internal/gitexec/*.go`, and `internal/skillinstall/*.go`
- [x] T019 [US3] Verify catalog completeness and formatting-argument compatibility in `cmd/cerne/messages_test.go`

**Checkpoint**: All human output is bilingual and all automation-facing contracts are language-neutral.

---

## Phase 6: Polish & Cross-Cutting Concerns

- [x] T020 [P] Document configuration, precedence, compatibility default, and the `1.0` transition in `README.md`, `README.pt-BR.md`, `README.es.md`, and `docs/*/commands.md`
- [x] T021 [P] Add migration and compatibility notes under Unreleased in `CHANGELOG.md`
- [x] T022 Run the quickstart and full validation with `go test -count=1 ./...`, `go test -race ./...`, `go vet ./...`, and `git diff --check`

---

## Dependencies & Execution Order

- Phase 1 precedes all implementation.
- Phase 2 blocks all user stories because catalogs and semantic domain values are shared.
- User Story 1 establishes persisted configuration before User Story 2 resolves temporary precedence.
- User Story 3 consumes the selection behavior from Stories 1 and 2 and completes command coverage.
- Documentation and full cross-platform validation follow all stories.

## Parallel Opportunities

- T003 can proceed while configuration file scaffolding is prepared.
- T006 and T008 cover separate package and CLI boundaries after the foundational API exists.
- T011 and T012 cover separate resolution and integration boundaries.
- T014 and T015 divide localized prose from automation invariance.
- T020 and T021 touch independent documentation files after behavior stabilizes.

## Implementation Strategy

1. Deliver the persisted-language vertical slice with global/config help.
2. Add deterministic temporary selection and precedence.
3. Convert all remaining presentation paths and lock down automation invariance.
4. Update multilingual documentation and run the complete portability-oriented test suite.
