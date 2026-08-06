# Cerne

[![Tests](https://github.com/WilliamSampaio/cerne-cli/actions/workflows/test.yml/badge.svg)](https://github.com/WilliamSampaio/cerne-cli/actions/workflows/test.yml)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

**English** · [Português (Brasil)](README.pt-BR.md) · [Español](README.es.md)

Cerne is an open-source, cross-platform CLI written in Go for managing software workspaces made of
two independent Git repositories:

- **knowledge** — the project's intent, product information, specifications, decisions, policies,
  and execution records; normally private;
- **source** — the application's source code.

The name *Cerne* means “core”. The project starts with safe local workspace management and is
designed to evolve into a model- and vendor-independent harness for coordinating AI agents across
documentation, product, implementation, validation, and maintenance.

## Why Cerne?

Cerne is built on a few durable rules:

- your knowledge belongs to you and remains accessible as ordinary files and Git history;
- private knowledge and application code stay in separate repositories;
- integrations belong behind adapters instead of leaking into the domain;
- automated work must be traceable and receive only the context it needs;
- push, merge, publication, deployment, and destructive operations require explicit approval;
- secrets and credentials must never be stored in managed repositories.

The current version is intentionally local. It does not call AI agents, manage hosting services,
publish, or deploy. It contacts a Git origin only for an explicit `init --clone`.

## Requirements

- Git available in `PATH`;
- Go 1.26.5 or newer to build from source;
- Linux, Windows, or macOS.

Spec Kit's `specify` or OpenSpec's `openspec` executable is optional and required only when that
workflow is selected. Cerne never installs or updates either tool.

## Installation

Install directly with Go:

```sh
go install github.com/WilliamSampaio/cerne-cli/cmd/cerne@latest
cerne --version
cerne --help
```

Go places the binary in `GOBIN`, or in `GOPATH/bin` when `GOBIN` is unset. Make sure that directory
is in `PATH`.

To build a development copy:

```sh
git clone https://github.com/WilliamSampaio/cerne-cli.git
cd cerne-cli
go build -o cerne ./cmd/cerne
./cerne --version
```

On Windows, the generated binary is `cerne.exe`.

## Quick start

### 1. Create a workspace

```sh
cerne init my-project
cd my-project
```

Cerne creates:

```text
my-project/
├── knowledge/
│   ├── .git/
│   ├── cerne.json
│   ├── product/
│   ├── specs/
│   ├── decisions/
│   ├── policies/
│   └── runs/
└── source/
    └── .git/
```

Both repositories are local, independent, and initially have no commits or remotes. The workspace
root itself is not a Git repository.

Because Git does not track empty directories, Cerne creates a `.gitkeep` file in each required
`knowledge` directory. You can remove it after adding content to that directory. Cerne does not
create commits automatically.

To start with an existing source, choose exactly one source mode:

```sh
cerne init my-project --source ../existing-application
cerne init my-project --clone https://host/organization/application.git
```

`--source` links a local non-bare Git worktree, resolves relative paths from the invocation
directory, and never creates an internal source or changes the external repository. `--clone`
accepts an existing local path, `file`, HTTPS, or SSH (including SCP-like syntax), performs a full
standard clone into internal `source`, and keeps the remote named `origin`. `--source` and
`--clone` are mutually exclusive; either one can be combined with `--workflow`.

To bootstrap an optional specification workflow during creation:

```sh
cerne init my-project --workflow speckit
cerne init my-project --workflow openspec
cerne init my-project --clone https://host/organization/application.git --workflow speckit
```

Spec Kit keeps specifications in `knowledge/specs` and owns `knowledge/.specify`. OpenSpec keeps
them in `knowledge/openspec/specs` and owns `knowledge/openspec`; it does not create top-level
`knowledge/specs`. Common product, decision, policy, and run directories remain in both layouts.

If the executable is missing, `init` still succeeds, records the choice, and warns on stderr.
After installing it, run `cerne workflow setup` from anywhere inside the workspace. Setup is
idempotent; a ready layout is unchanged and a partial or nested-Git layout is refused.

### 2. Validate it

Run this from the workspace root:

```sh
cerne doctor
```

The report marks each check with `✓` (passed), `!` (warning), or `✗` (blocking error).

### 3. Inspect local Git state

```sh
cerne status
```

The command reports the branch, abbreviated commit, worktree state, staged files, modified files,
and untracked files for both repositories. Pending changes are information, not an error.

### 4. Link an existing source repository (optional)

`init` already configures the empty `source` repository. Use `--replace` to point the manifest to
another local repository:

```sh
cerne link ../existing-application --replace
```

Only the manifest reference changes. Cerne never copies, moves, cleans, checks out, commits, or
deletes either source repository.

## Manifest

`knowledge/cerne.json` identifies the project and locates its source repository:

```json
{
  "name": "my-project",
  "source": "../source"
}
```

The absence of `version` means manifest version 1. When present, the only currently supported value
is the JSON integer `1`. Cerne stores a normalized relative source path whenever the platforms and
locations allow it.

With a selected workflow, the manifest also contains `"workflow":{"provider":"speckit"}` or
`"workflow":{"provider":"openspec"}`. Installation state and tool versions are not persisted.

## Command reference

### Global options

- `cerne --help` prints the available commands and global options.
- `cerne --version` prints the stable SemVer identifier, currently `cerne 0.3.0`.

### `cerne init <project-name> [--source ... | --clone ...] [--workflow ...]`

Creates a new workspace below the current directory. The destination must not exist or must be an
empty regular directory; symbolic links and non-empty destinations are rejected. Existing content
is never replaced. If creation fails, Cerne rolls back only artifacts created by that attempt.

Project names use 1–255 ASCII characters, start with a letter or number, and may continue with
letters, numbers, `.`, `_`, or `-`. Windows reserved names and names ending in `.` are rejected.
Without the option, behavior is unchanged. With it, Cerne invokes the installed provider only in
knowledge, non-interactively and without selecting an AI agent. A missing executable is a warning;
an executed provider failure keeps the base workspace and returns an operational error.

`--source` validates and links an existing local worktree without modifying it. `--clone` rejects
HTTP, `git://`, `ext::`, unknown helpers, option-like inputs, embedded credentials, queries, and
fragments before Git runs. Authentication, redirects, and checkout filters remain Git behavior;
Cerne disables controllable prompts, but external helpers may still fail or behave outside the
CLI's portable control. Clone adds no depth, branch, submodule, LFS, push, or extra fetch. Either
source option can be combined with `--workflow`; source and clone remain mutually exclusive.

Every started clone first creates a redacted `knowledge/runs/source-clone.json`. A pre-clone failure
rolls back the attempt. A later failure preserves knowledge and the audit, removes only Cerne's
private staging, and reports an incomplete workspace. Promotion never replaces a concurrent source;
if final audit writing fails after promotion, the valid source remains.

### `cerne workflow setup`

Locates the nearest ancestor workspace and materializes the provider declared in its manifest. It
does not accept a provider, path, or force option. Each real attempt creates one redacted JSON audit
record in `knowledge/runs`; no audit is created for a missing executable or ready layout.

### `cerne doctor`

Performs ten read-only checks from the workspace root: manifest readability, both repository
directories, Git independence, versioning isolation, manifest paths, required knowledge
directories, Git availability, permissions, and manifest version. A declared workflow adds a check
for ready, pending, unavailable, unknown, partial, or nested-Git state. It never repairs the workspace.

### `cerne status`

Locates the nearest workspace from the current directory and reads both repositories. It recognizes
clean and pending worktrees, detached HEAD, and repositories without commits. It does not fetch or
compare with remotes.

### `cerne link <path> [--replace]`

Links a local non-bare Git repository with a worktree as `source`. Relative and absolute paths are
accepted, including valid Git worktrees. Knowledge and source must be distinct and must not be
dangerously nested. Replacing a different configured source requires `--replace`; linking the same
source succeeds without rewriting the manifest. Manifest replacement is atomic.

Use `<command> --help` for the complete contract. CLI output is currently in Portuguese.

## Exit codes and streams

| Code | Meaning |
| --- | --- |
| `0` | Success, help, a healthy workspace, warnings only, or successfully collected pending status |
| `1` | Operational failure or a blocking `doctor` finding |
| `2` | Invalid command usage or invalid project name |

Normal output and help use stdout. Usage and operational failures use stderr. `doctor` reports,
including blocking findings, use stdout so the full diagnosis remains one stable stream.

## Safety and privacy

- `doctor` and `status` are read-only.
- `link` updates only `knowledge/cerne.json` after all validations pass.
- Workflow setup uses fixed arguments, no shell, a minimal environment, and disables OpenSpec
  telemetry. It receives no credentials or source path and does not log raw provider output.
- Clone uses fixed shell-free Git arguments, a protocol allowlist, private staging, and
  non-replacing promotion. The origin and raw Git output are excluded from Cerne output, manifest,
  and audit; authentication remains external and Git retains the origin as remote `origin`.
- A failed attempt preserves the base workspace and audit, removing only a newly created
  provider-owned root.
- Git inspection disables optional locks and terminal prompts and removes redirecting `GIT_*`
  variables from child processes.
- Only explicit `init --clone` may contact an origin or use externally configured credentials.
- Do not place tokens, passwords, private keys, or other secrets in either managed repository.

## Technical design

The codebase keeps responsibilities small and explicit:

```text
cmd/cerne/          command parsing, terminal output, and exit codes
internal/workspace/ domain rules and workspace operations
internal/gitexec/   adapter for the local Git executable
internal/filecheck/ cross-platform permission checks
internal/workflowexec/ adapters for optional local workflow executables
specs/              feature specifications, plans, contracts, and tasks
```

The implementation prefers the Go standard library. Platform-specific filesystem behavior is
isolated with build tags. CI runs the test suite on Linux, Windows, and macOS. Domain behavior is
separate from terminal rendering so it can be reused by future interfaces.

## Development

```sh
go build -o cerne ./cmd/cerne
go test ./...
go test -count=1 ./...
go vet ./...
gofmt -w <changed-go-files>
```

Tests use Go's `testing` package, temporary directories, and local Git repositories only. They do
not require network access or credentials.

## Contributing

Contributions are welcome:

1. Open an issue or discuss the intended behavior before a large change.
2. Create a focused branch and keep domain rules out of terminal rendering.
3. Add or update a test that fails without the behavior change.
4. Run `gofmt`, `go vet ./...`, and `go test -count=1 ./...`.
5. Open a pull request describing the intent, linked issue or `specs/` artifact, validation
   commands, and any CLI compatibility impact.

Use short Conventional Commit-style subjects such as `feat: add command` or `fix: preserve manifest`.
See [AGENTS.md](AGENTS.md) for repository-specific contributor rules and
[the project constitution](.specify/memory/constitution.md) for governance and compatibility rules.
Release history is documented in [CHANGELOG.md](CHANGELOG.md).

## Roadmap and scope

The current scope is workspace creation from an empty, linked, or cloned source, optional workflow
bootstrap, validation, local status, and source linking. Future work may add auditable agent coordination for product, documentation,
implementation, validation, and maintenance while remaining independent of specific AI models,
agents, and providers.

Remote hosting management, automatic commits, push, pull requests, merge, publication,
deployment, GUI, JSON output, and AI execution are not part of the current CLI.

## License

Cerne is distributed under the [MIT License](LICENSE).
