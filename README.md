# Cerne

[![Tests](https://github.com/WilliamSampaio/cerne-cli/actions/workflows/test.yml/badge.svg)](https://github.com/WilliamSampaio/cerne-cli/actions/workflows/test.yml)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

**English** · [Português (Brasil)](README.pt-BR.md) · [Español](README.es.md)

[User documentation](docs/en/getting-started.md)

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
With Spec Kit, `--agent codex|claude` can also prepare local command discovery at the workspace
root; the agent choice is not stored in `knowledge/cerne.json`.

## Installation

On Linux and macOS, install the latest stable standalone binary without Go:

```sh
curl --proto '=https' --tlsv1.2 -fsSL \
  https://github.com/WilliamSampaio/cerne-cli/releases/latest/download/install.sh | sh
~/.local/bin/cerne --version
```

Inspect the installer first with the same `curl` command without `| sh`. To install a fixed
version, use the release URL for that tag:

```sh
curl --proto '=https' --tlsv1.2 -fsSL \
  https://github.com/WilliamSampaio/cerne-cli/releases/download/vX.Y.Z/install.sh |
  sh -s -- --version vX.Y.Z
```

To install the optional skills for an agent in the same explicit flow:

```sh
curl --proto '=https' --tlsv1.2 -fsSL \
  https://github.com/WilliamSampaio/cerne-cli/releases/latest/download/install.sh |
  sh -s -- --agent codex
```

`--agent codex` and `--agent claude` install every compatible official skill. `--agent gemini`
installs the Git workflow skill only.

The installer writes only `~/.local/bin/cerne`, verifies release checksums before replacing a
regular file, refuses directory or symlink destinations, never uses `sudo`, and never edits shell
profiles. Remove it with `rm ~/.local/bin/cerne`. For manual installation, download the matching
archive and `checksums.txt` from the release, verify SHA-256, extract `cerne`, and place it on
`PATH`.

Install directly with Go instead:

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

To install the current checkout for host-wide testing without depending on commits or the remote:

```sh
go install ./cmd/cerne
# or
make install-local
```

On Windows, the generated binary is `cerne.exe`.

## Language

Human-readable CLI output is available in English and Brazilian Portuguese. Save a preference or
override it for one invocation:

```sh
cerne config set language en
cerne --lang pt-BR doctor
CERNE_LANG=en cerne status
```

Precedence is `--lang`, `CERNE_LANG`, the saved preference, then `pt-BR`. The current default
remains `pt-BR` for compatibility and will change to `en` in 1.0. Structured output, commands,
flags, identifiers, exit statuses, and version output are language-neutral.

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
cerne init my-project --workflow speckit --agent codex
cerne init my-project --workflow openspec
cerne init my-project --clone https://host/organization/application.git --workflow speckit
```

Spec Kit keeps specifications in `knowledge/specs` and owns `knowledge/.specify`. OpenSpec keeps
them in `knowledge/openspec/specs` and owns `knowledge/openspec`; it does not create top-level
`knowledge/specs`. Common product, decision, policy, and run directories remain in both layouts.

If the executable is missing, `init` still succeeds, records the choice, and warns on stderr.
After installing it, run `cerne workflow setup` from anywhere inside the workspace. Setup is
idempotent; a ready layout is unchanged and a partial or nested-Git layout is refused.
When `--agent codex` or `--agent claude` is used with Spec Kit, Cerne also asks Spec Kit to create
the matching integration inside `knowledge` and writes small workspace-root bridge files in
`.agents/skills` or `.claude/skills`. Those bridge files point back to `knowledge` and contain no
private knowledge content, remotes, credentials, environment dumps, or absolute paths. For Codex to
discover these local skills, start the session from the Cerne workspace root; a session started
inside `source/` does not walk up to the workspace root because `source` is a separate Git
repository.

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

### 4. Coordinate an approved Git step (optional)

Install `cerne-git-workflow` with `cerne skill install <agent>`, then let the skill inspect first
and ask for a separate confirmation before the agent runs any Git effect:

```sh
cerne git inspect --agent codex --task task-1 --json
```

Branch, commit, push, and Pull Request are separate agent-run steps. Cerne only provides the
sanitized snapshot (`state_id`, repositories, branches, remotes, and changed paths). Unsupported or destructive Git operations are refused by the skill. See
[the command reference](docs/en/commands.md) for the full JSON/audit contract.
The Cerne audit covers inspection only; native effect evidence belongs to the agent or harness.

### 5. Link an existing source repository (optional)

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
`"workflow":{"provider":"openspec"}`. Installation state, tool versions, and local agent choice are
not persisted.

## Command reference

See the complete [Cerne command reference](docs/en/commands.md).

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
- `restore` keeps its audit under the user's private `~/.cerne/audit`, validates both repository
  boundaries, and uses identity-checked rollback with non-replacing promotion.
- `skill install` writes only to the authorized agent profile, validates the embedded official
  package before copying, and refuses unknown destination content.
- A failed attempt preserves the base workspace and audit, removing only a newly created
  provider-owned root.
- Git inspection disables optional locks and terminal prompts and removes redirecting `GIT_*`
  variables from child processes.
- Only explicit `init --clone` or `restore` may contact an origin or use externally configured credentials.
- Do not place tokens, passwords, private keys, or other secrets in either managed repository.

## Technical design

The codebase keeps responsibilities small and explicit:

```text
cmd/cerne/          command parsing, terminal output, and exit codes
internal/workspace/ domain rules and workspace operations
internal/gitexec/   adapter for the local Git executable
internal/filecheck/ cross-platform permission checks
internal/workflowexec/ adapters for optional local workflow executables
internal/skillinstall/ explicit global installation of official skills
specs/              feature specifications, plans, contracts, and tasks
```

The implementation prefers the Go standard library. Platform-specific filesystem behavior is
isolated with build tags. CI runs the test suite on Linux, Windows, and macOS. Domain behavior is
separate from terminal rendering so it can be reused by future interfaces.

## Development

```sh
go build -o cerne ./cmd/cerne
go install ./cmd/cerne
go test ./...
go test -count=1 ./...
go vet ./...
gofmt -w <changed-go-files>
```

Equivalent shortcuts are available through `make build`, `make install-local`, `make test`,
`make test-fresh`, `make vet`, `make fmt`, and `make check`. Use `make install-path` to show where
`make install-local` makes the executable available. To install somewhere else, set `GOBIN`, for
example `GOBIN=/tmp/bin make install-local`.

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
