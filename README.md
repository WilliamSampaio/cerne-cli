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

The current version is intentionally local. It does not call AI agents, access GitHub, clone
remotes, publish, or deploy anything.

## Requirements

- Git available in `PATH`;
- Go 1.26.5 or newer to build from source;
- Linux, Windows, or macOS.

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

Git does not track empty directories. Add project knowledge before the first knowledge commit so
the directories that matter are preserved; Cerne intentionally creates no placeholder files or
automatic commits.

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

## Command reference

### Global options

- `cerne --help` prints the available commands and global options.
- `cerne --version` prints the stable SemVer identifier, currently `cerne 0.1.0`.

### `cerne init <project-name>`

Creates a new workspace below the current directory. The destination must not exist or must be an
empty regular directory; symbolic links and non-empty destinations are rejected. Existing content
is never replaced. If creation fails, Cerne rolls back only artifacts created by that attempt.

Project names use 1–255 ASCII characters, start with a letter or number, and may continue with
letters, numbers, `.`, `_`, or `-`. Windows reserved names and names ending in `.` are rejected.

### `cerne doctor`

Performs ten read-only checks from the workspace root: manifest readability, both repository
directories, Git independence, versioning isolation, manifest paths, required knowledge
directories, Git availability, permissions, and manifest version. It never repairs the workspace.

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
- Git inspection disables optional locks and terminal prompts and removes redirecting `GIT_*`
  variables from child processes.
- No current command contacts remotes or needs credentials.
- Do not place tokens, passwords, private keys, or other secrets in either managed repository.

## Technical design

The codebase keeps responsibilities small and explicit:

```text
cmd/cerne/          command parsing, terminal output, and exit codes
internal/workspace/ domain rules and workspace operations
internal/gitexec/   adapter for the local Git executable
internal/filecheck/ cross-platform permission checks
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

The current scope is workspace creation, validation, local status, and linking an existing local
source repository. Future work may add auditable agent coordination for product, documentation,
implementation, validation, and maintenance while remaining independent of specific AI models,
agents, and providers.

Remote repository management, automatic commits, push, pull requests, merge, publication,
deployment, GUI, JSON output, and AI execution are not part of the current CLI.

## License

Cerne is distributed under the [MIT License](LICENSE).
