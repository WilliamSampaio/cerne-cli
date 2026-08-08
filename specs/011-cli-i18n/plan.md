# Implementation Plan: CLI Internationalization

**Branch**: `feat/cli-i18n` | **Date**: 2026-08-07 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/011-cli-i18n/spec.md`

## Summary

Add complete English and Brazilian Portuguese presentation catalogs, persistent user language configuration, and temporary language selection while keeping all domain decisions and machine-readable contracts language-neutral. The current `pt-BR` default remains for this minor release and is documented as deprecated in favor of `en` for `1.0`.

## Technical Context

**Language/Version**: Go 1.26.5

**Primary Dependencies**: Go standard library; existing `golang.org/x/sys` only where platform file replacement already requires it

**Storage**: One user-owned JSON preference file at `~/.cerne/config.json`

**Testing**: Go `testing`, temporary home directories, exact stdout/stderr and exit-status assertions, cross-platform CI

**Target Platform**: Linux, Windows, and macOS

**Project Type**: CLI

**Performance Goals**: Language resolution adds no observable delay to an interactive invocation and reads at most one small local configuration file

**Constraints**: Offline operation; no locale service or translation dependency; atomic configuration replacement; no localization of machine-readable values

**Scale/Scope**: Two complete language catalogs covering eight existing command groups plus the new `config` command

## Constitution Check

*GATE: Passed before research and re-checked after design.*

- **Ownership and separation**: PASS. The preference belongs to the current user and never enters knowledge or source repositories.
- **Neutral core**: PASS. Localization is presentation/configuration behavior and remains independent from AI providers and external systems.
- **Minimum context and audit**: PASS. No agent context is added. A language preference is non-sensitive and its explicit configuration command does not require an audit record.
- **Authorization and secrets**: PASS. `config set` and `config unset` explicitly authorize only their named preference change; the file stores no secret.
- **Portability**: PASS. Paths use the resolved home directory and atomic replacement has Linux, Windows, and macOS coverage.
- **Testing**: PASS. Catalog parity, precedence, unsafe paths, exact streams, exit statuses, structured output invariance, and command flows receive automated tests.
- **CLI compatibility and documentation**: PASS. Existing identifiers and machine contracts remain stable. `pt-BR` remains the minor-release default and the `en` transition is announced for `1.0`.
- **Simplicity**: PASS. Source catalogs and standard-library formatting cover the current two-language requirement without a new dependency.

## Project Structure

### Documentation (this feature)

```text
specs/011-cli-i18n/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── config-command.md
│   └── language-selection.md
└── tasks.md
```

### Source Code (repository root)

```text
cmd/cerne/
├── main.go                 # command routing and localized rendering
├── main_test.go            # binary-level language contract tests
├── messages_en.go          # English presentation catalog
└── messages_ptbr.go        # Brazilian Portuguese presentation catalog

internal/localization/
├── language.go             # supported values, precedence, and message lookup
├── language_test.go
├── config.go               # user preference read/write/unset
├── config_test.go
├── atomic_replace_unix.go
└── atomic_replace_windows.go

internal/workspace/         # semantic result/failure codes consumed by renderers
internal/skillinstall/      # semantic failure codes consumed by renderers

docs/{en,pt-BR,es}/         # localized user documentation
README{,.pt-BR,.es}.md      # concise language configuration entry points
```

**Structure Decision**: Keep localization storage and resolution in one small internal package, keep CLI prose catalogs beside the CLI that owns them, and expose semantic domain codes rather than translated domain strings.

## Complexity Tracking

No constitutional violations or additional complexity exceptions are required.
