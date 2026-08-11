# Changelog

All notable changes to Cerne are documented in this file. This project follows
[Semantic Versioning](https://semver.org/).

## Unreleased

### Changed

- Official skills now respond in the active conversation language, falling back to the effective
  Cerne language when available.

## 0.10.0 - 2026-08-09

### Added

- `cerne skill install <agent>` now installs every compatible official skill for Codex, Claude, and
  Gemini, including the new `cerne-git-workflow` skill.
- `cerne skill install <agent> <skill>` installs one explicit supported skill.
- `cerne git inspect` for the `cerne-git-workflow` skill; branch, commit, push, and Pull Request
  effects are delegated to the agent's native Git and GitHub tooling.

### Security

- Every `cerne git` mutation requires `--state`, `--confirm`, `--agent`, and `--task`, starts a
  private redacted audit before effects, reinspects stale state, disables controllable Git prompts,
  and refuses destructive or arbitrary Git operations.
- Pull Request preparation uses no network or credentials and prevents Cerne from handling GitHub
  tokens; the agent uses exactly the authorized PR targets with its own authentication.

## 0.9.0 - 2026-08-09

### Added

- Standalone Linux/macOS installer published with release assets, including checksum-verified
  binary archives for `amd64` and `arm64`.
- Optional `install.sh --agent <codex|claude>` flow that delegates skill installation to the
  installed `cerne skill install` command.

### Security

- The installer writes only `~/.local/bin/cerne`, refuses directory and symlink destinations,
  verifies SHA-256 before promotion, never uses `sudo`, and never edits shell profile files.

## 0.8.1 - 2026-08-09

### Fixed

- `cerne restore --clone` now repairs non-portable restored source paths to `../source` instead
  of rejecting workspaces whose manifest points to a source location from another machine.

## 0.8.0 - 2026-08-08

### Added

- English and Brazilian Portuguese human-readable CLI output, selectable for one invocation with
  `--lang` or `CERNE_LANG`.
- `cerne config set|get|unset language` to manage the current user's preference in
  `~/.cerne/config.json`.

### Changed

- The official `cerne-skills` package is embedded in the CLI binary, so `go install` supports
  `cerne skill install <codex|claude>` without a separately distributed companion directory.
- Brazilian Portuguese remains the compatibility default until 1.0, when English becomes the
  default. Commands, flags, structured output, identifiers, exit statuses, and version output are
  unchanged by language selection.

## 0.7.0 - 2026-08-07

### Added

- `cerne skill install <codex|claude>` to explicitly install the official `cerne-context` skill
  from the local companion `cerne-skills` package into the current user's agent profile.
- `--agent codex|claude` for Spec Kit workspaces during `cerne init` and `cerne workflow setup`,
  making Spec Kit commands discoverable from the workspace root while keeping `knowledge` as the
  real Spec Kit project root.
- Local Codex and Claude discovery bridges through workspace-root `.agents/skills` and
  `.claude/skills`, with the agent choice kept out of `knowledge/cerne.json`.

### Security

- Skill installation validates the companion manifest and adapter before copying, refuses unknown
  destinations, is idempotent for the same managed version, and writes redacted private audit
  records under `~/.cerne/audit`.
- Agent integration subprocesses are audited in `knowledge/runs`, and bridge files avoid symlinks,
  absolute paths, provider output, environment dumps, remotes, credentials, and private knowledge
  content.

## 0.6.0 - 2026-08-06

### Added

- `cerne context` human report and stable `cerne context --json` schema version 1 for structural
  workspace discovery by people and external skills.

### Security

- Context inspection is strictly read-only and does not read repository content or agent files,
  execute Git/providers, access the network, expose credentials/remotes, or require workspace
  migration.

## 0.5.0 - 2026-08-05

### Added

- `cerne restore` to clone knowledge and either clone or link source in one transactional command.
- Private global restore audits with redacted phase transitions and cross-platform access controls.

### Security

- Restore validates portable manifest paths, independent Git roots, absent destinations and
  workflow state before non-replacing promotion, with identity-checked rollback on failure.

## 0.4.0 - 2026-08-05

### Added

- `--workflow` can be combined with either `--source` or `--clone` during initialization.

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
