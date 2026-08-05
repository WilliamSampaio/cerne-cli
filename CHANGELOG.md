# Changelog

All notable changes to Cerne are documented in this file. This project follows
[Semantic Versioning](https://semver.org/).

## 0.3.0 - 2026-08-05

### Added

- `cerne init --source` to atomically link an existing local Git worktree without modifying it.
- `cerne init --clone` to securely clone allowed local, file, HTTPS, or SSH origins into source.
- Redacted pre-execution clone audits with private staging and non-replacing promotion.

### Security

- Clone rejects unsafe transports and embedded credentials, disables Git prompts, uses fixed
  shell-free arguments, and preserves only owned artifacts across failure boundaries.

## 0.2.0 - 2026-08-05

### Added

- Optional Spec Kit or OpenSpec bootstrap through `cerne init --workflow`.
- Idempotent `cerne workflow setup` recovery for pending workspaces.
- Redacted workflow audit records and workflow-aware `cerne doctor` diagnostics.

### Security

- Workflow providers run without a shell, with fixed arguments, a minimal environment, and no
  inherited credentials; OpenSpec telemetry is disabled.

## 0.1.0 - 2026-08-03

### Added

- `cerne init` to create workspaces with independent knowledge and source repositories.
- `cerne doctor` to validate workspace structure, safety, permissions, and manifest compatibility.
- `cerne status` to report local Git state for both repositories without modifying them.
- `cerne link` to atomically link an existing local Git worktree as source.
- Global `--help` and `--version` options.
- English, Brazilian Portuguese, and Spanish documentation.
- Automated tests on Linux, Windows, and macOS.

### Security

- Git inspection disables prompts and optional locks and sanitizes redirecting `GIT_*` variables.
- Workspace operations reject unsafe repository overlap and preserve existing files by default.
- Build baseline updated to Go 1.26.5 and `golang.org/x/sys` v0.47.0.
