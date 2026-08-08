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
private knowledge content, remotes, credentials, environment dumps, or absolute paths.

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
`"workflow":{"provider":"openspec"}`. Installation state, tool versions, and local agent choice are
not persisted.

## Command reference

### Global options

- `cerne --help` prints the available commands and global options.
- `cerne --version` prints the stable SemVer identifier, currently `cerne 0.7.0`.

### `cerne init <project-name> [--source ... | --clone ...] [--workflow ... [--agent ...]]`

Creates a new workspace below the current directory. The destination must not exist or must be an
empty regular directory; symbolic links and non-empty destinations are rejected. Existing content
is never replaced. If creation fails, Cerne rolls back only artifacts created by that attempt.

Project names use 1–255 ASCII characters, start with a letter or number, and may continue with
letters, numbers, `.`, `_`, or `-`. Windows reserved names and names ending in `.` are rejected.
Without `--workflow`, behavior is unchanged. With it, Cerne invokes the installed provider only in
knowledge and non-interactively. `--agent codex|claude` is accepted only with `--workflow speckit`;
it prepares local discovery for that invocation without persisting the agent. A missing executable
is a warning; an executed provider failure keeps the base workspace and returns an operational
error.

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

### `cerne restore <knowledge-origin> (--source <path> | --clone <source-origin>)`

Clones knowledge, reads the workspace name from `cerne.json`, then either clones source to the
manifest's portable path or links an existing local non-bare Git root without modifying it. The
destination must be absent. Existing, concurrent, overlapping, symlinked, partial, unknown-workflow,
or non-independent layouts are rejected and never replaced. A declared ready or pending workflow is
preserved but never executed.

Each valid attempt starts one private, redacted record in `~/.cerne/audit` before Git. Origins,
credentials, Git output, arguments, environment, and absolute repository paths are excluded. Clone
authentication and redirects remain external Git behavior. Failure rolls back only artifacts whose
identity still belongs to the attempt; retrying the command is the supported recovery—there is no
resume mode. Success/help use stdout and status `0`, operational failures use stderr/status `1`, and
invalid usage or origins use stderr/status `2`. Restore authorizes only the stated clone(s) or local
inspection: no workflow setup, agent, push, merge, extra fetch, submodule, install, publish, or deploy.

```sh
cerne restore ../knowledge.git --clone ../source.git
cerne restore git@host:org/knowledge.git --source ../existing-source
```

### `cerne skill install <codex|claude>`

Explicitly installs the official `cerne-context` skill in the current user's agent profile:
`~/.codex/skills/cerne-context` for Codex or `~/.claude/skills/cerne-context` for Claude. The
command uses only the local companion `cerne-skills` package, without network access, validates the
manifest, adapter, and `cerne.context.v1` schema before copying, and writes a private audit record
under `~/.cerne/audit`.

Invalid usage, including `generic`, case variants, missing agents, or extra arguments, returns
status `2` without audit or filesystem mutation. Operational failures return stderr/status `1`.
Reinstalling the same version is a no-op; different managed versions are upgraded. `init`,
`restore`, and `workflow setup` never install this skill by implication.

### `cerne workflow setup [--agent codex|claude]`

Locates the nearest ancestor workspace and materializes the provider declared in its manifest. It
does not accept a provider, path, or force option. With `--agent`, the declared workflow must be
Spec Kit and Cerne prepares or refreshes the workspace-root discovery bridge for the selected local
agent. Each real provider or agent-integration subprocess creates one redacted JSON audit record in
`knowledge/runs`; no audit is created for a missing executable or ready layout without agent setup.

### `cerne context [--json]`

Locates the nearest ancestor workspace and reports canonical paths for workspace, knowledge,
product, specs, decisions, policies, source, and the declared workflow. `--json` emits stable
schema version 1 for skills and scripts. Healthy and warning reports return `0`; structurally
invalid reports remain valid output and return `1`; invalid usage returns `2` on stderr.

The command reads only structural metadata. It does not read repository content or agent files,
run Git or workflow providers, inspect remotes or executables, access the network, or create audit,
cache, instruction, or manifest changes.

```sh
cerne context
cerne context --json
```

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
- `restore` keeps its audit under the user's private `~/.cerne/audit`, validates both repository
  boundaries, and uses identity-checked rollback with non-replacing promotion.
- `skill install` writes only to the authorized agent profile, validates the local companion package
  before copying, and refuses unknown destination content.
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
