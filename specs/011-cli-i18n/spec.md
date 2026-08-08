# Feature Specification: CLI Internationalization

**Feature Branch**: `feat/cli-i18n`

**Created**: 2026-08-07

**Status**: Draft

**Input**: User description: "Allow users to choose the Cerne CLI language, persist the preference in the user's home directory, use English as the future default, and initially support English and Brazilian Portuguese."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Persist a Preferred Language (Priority: P1)

As a Cerne user, I can select English or Brazilian Portuguese once and have subsequent commands use that language without repeating an option.

**Why this priority**: A persistent preference is the primary user value and avoids making localization cumbersome during everyday use.

**Independent Test**: Set the language to Brazilian Portuguese, run a separate help or diagnostic command, and verify that its human-readable output is in Brazilian Portuguese; then repeat with English.

**Acceptance Scenarios**:

1. **Given** no saved preference, **When** the user sets the language to `pt-BR`, **Then** the preference is stored for that user and later commands render human-readable output in Brazilian Portuguese.
2. **Given** a saved `pt-BR` preference, **When** the user sets the language to `en`, **Then** later commands render human-readable output in English.
3. **Given** a saved preference, **When** the user asks for its current value, **Then** the CLI reports the canonical language identifier.
4. **Given** a saved preference, **When** the user removes it, **Then** later commands return to the release's default language.

---

### User Story 2 - Override the Language Temporarily (Priority: P2)

As a user or automation author, I can select a language for one invocation or environment without changing my saved preference.

**Why this priority**: Temporary selection supports troubleshooting, documentation examples, CI, and shared machines without mutating user configuration.

**Independent Test**: Save `pt-BR`, invoke a command with an English temporary selection, verify English output, and then verify that the saved preference remains `pt-BR`.

**Acceptance Scenarios**:

1. **Given** a saved `pt-BR` preference, **When** a command is invoked with `--lang en`, **Then** that invocation uses English and the saved preference remains unchanged.
2. **Given** a saved `pt-BR` preference, **When** a command runs with `CERNE_LANG=en`, **Then** that invocation uses English and the saved preference remains unchanged.
3. **Given** both an environment selection and a command-line selection, **When** the command runs, **Then** the command-line selection takes precedence.

---

### User Story 3 - Preserve Automation Contracts (Priority: P3)

As an automation author, I can change the human language without changing machine-readable output, command names, flags, or exit statuses.

**Why this priority**: Localization must not make existing automation dependent on a user's language preference.

**Independent Test**: Run the same machine-readable command and failure scenarios in both languages and compare their structured output and exit statuses.

**Acceptance Scenarios**:

1. **Given** any supported language, **When** `cerne context --json` runs against the same workspace, **Then** its JSON is identical across languages.
2. **Given** any supported language, **When** a command succeeds or fails, **Then** its documented exit status is unchanged.
3. **Given** any supported language, **When** a user invokes commands and flags, **Then** their existing English identifiers remain valid and are not localized.

### Edge Cases

- An unsupported language identifier is rejected with the supported values and does not alter the saved preference.
- A missing preference file behaves as an unset preference.
- A malformed, unsafe, or unreadable preference file produces a clear diagnostic and recovery guidance rather than silently selecting an arbitrary language.
- An unavailable or unwritable user home prevents persistent changes without partially replacing an existing preference.
- A temporary language selection remains usable to diagnose or repair an invalid saved preference.
- Language selection does not translate paths, provider names, agent names, schema keys, enum values, or other stable identifiers.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The CLI MUST provide complete human-readable output in English and Brazilian Portuguese.
- **FR-002**: Users MUST be able to set, inspect, and remove their persistent language preference through `cerne config set language <en|pt-BR>`, `cerne config get language`, and `cerne config unset language`.
- **FR-003**: The persistent preference MUST belong to the current user and be stored at `~/.cerne/config.json` without modifying a workspace.
- **FR-004**: The effective language MUST be resolved in this order: `--lang`, `CERNE_LANG`, saved preference, default language.
- **FR-005**: The default language for this release MUST remain `pt-BR`; the release MUST announce that `en` becomes the default in `1.0` after the required deprecation period.
- **FR-006**: `--lang` MUST affect only the current invocation, and `CERNE_LANG` MUST affect only processes that receive that environment value; neither MAY modify the saved preference.
- **FR-007**: Unsupported language values MUST fail without changing configuration and MUST identify `en` and `pt-BR` as supported values.
- **FR-008**: Help, reports, warnings, errors, causes, corrections, and user-facing labels MUST use the effective language consistently within one invocation.
- **FR-009**: Command names, subcommands, flags, paths, agent/provider identifiers, structured field names, structured enum values, and version output MUST remain language-neutral and retain their existing values.
- **FR-010**: Machine-readable output and exit statuses MUST remain identical for equivalent operations in both languages.
- **FR-011**: Persistent configuration writes MUST either complete fully or preserve the previous valid configuration.
- **FR-012**: The CLI MUST reject unsafe preference-file paths and MUST NOT follow a symbolic link when writing the preference.
- **FR-013**: The feature MUST behave consistently on Linux, Windows, and macOS.
- **FR-014**: User documentation and command help MUST document supported languages, selection precedence, persistence, temporary selection, reset behavior, and automation guarantees in English, Brazilian Portuguese, and Spanish documentation.
- **FR-015**: Spanish CLI output and translation of command or flag identifiers are out of scope for this release.

### Constitutional Requirements

- **Ownership/Repositories**: Language preferences are user-level configuration and MUST NOT be stored in or alter knowledge and source repositories.
- **AI/Integrations**: Language selection MUST remain independent of agents, providers, models, and external services.
- **Context/Audit**: Language selection MUST NOT expand agent context. Reading or changing this non-sensitive user preference does not require an audit record.
- **Authorization/Secrets**: The explicit `config set` or `config unset` invocation authorizes only the corresponding user configuration change. The configuration MUST contain no secrets or credentials.
- **Portability**: Preference resolution and storage MUST provide equivalent behavior and diagnostics on Linux, Windows, and macOS.
- **CLI/Compatibility**: Existing command and flag names, structured output, and exit statuses remain stable. The default-language transition MUST follow the project's deprecation and major-version policy.

### Key Entities

- **Language Preference**: The current user's optional saved choice, with one canonical value from `en` or `pt-BR`.
- **Effective Language**: The language selected for one invocation after applying command-line, environment, saved-preference, and default precedence.
- **Message**: A human-readable CLI text identified independently from its translated English and Brazilian Portuguese representations.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A user can set, inspect, use, and remove a language preference using at most one command for each action.
- **SC-002**: Every documented command can render its help, success output, warnings, and errors in both supported languages without mixed-language diagnostic labels.
- **SC-003**: Equivalent machine-readable commands produce byte-for-byte identical structured output in both supported languages.
- **SC-004**: Equivalent scenarios retain 100% of their documented exit statuses across both supported languages.
- **SC-005**: Preference writes interrupted before completion leave either the previous valid preference or no preference, never a partial configuration.
- **SC-006**: All supported operating-system test jobs complete successfully with both language catalogs available.

## Assumptions

- Users select languages through canonical identifiers `en` and `pt-BR`; locale aliases and automatic operating-system locale detection are out of scope.
- The feature ships in a minor release with `pt-BR` as the compatibility default; changing the default to `en` is deferred to `1.0`.
- Human-readable output is intended for people; automation relies on structured output and documented exit statuses.
- The preference file initially stores only the language preference; unrelated configuration capabilities are out of scope.
- Adding another CLI language later requires a complete catalog but does not change language precedence or configuration semantics.
