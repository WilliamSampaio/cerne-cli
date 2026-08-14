# Command reference

**English** · [Português (Brasil)](../pt-BR/commands.md) · [Español](../es/commands.md)

[Getting started](getting-started.md) · [Troubleshooting](troubleshooting.md)

Cerne supports English and Brazilian Portuguese for human-readable output. Run
`cerne <command> --help` for the complete contract implemented by your installed version.

<!-- AUTO-GENERATED: keep synchronized with cmd/cerne/main.go and the CLI contracts. -->

## Standalone installer

Linux and macOS releases publish `install.sh`, `checksums.txt`, and binary archives for `amd64` and
`arm64`. The installer accepts `--version <version>`, `--agent <codex|claude|gemini>`, and `--help`.
It installs only `~/.local/bin/cerne`, verifies SHA-256 and the binary-reported version before
promotion, refuses directory or symlink destinations, never uses `sudo`, and never edits shell
profile files.

```sh
curl --proto '=https' --tlsv1.2 -fsSL \
  https://github.com/WilliamSampaio/cerne-cli/releases/latest/download/install.sh | sh
```

## Global options

- `cerne --help` prints the available commands and global options.
- `cerne --version` prints the installed version as a stable SemVer identifier.
- `cerne --lang <en|pt-BR> ...` selects the language for one invocation without saving it.

`CERNE_LANG` provides the same temporary override. Language precedence is `--lang`,
`CERNE_LANG`, the saved preference, then `en`. Language selection changes only human-readable text;
commands, flags, JSON fields, identifiers, exit statuses, and `--version` remain stable.

## `cerne config <set language <en|pt-BR>|get language|unset language>`

Manages the language preference stored in `~/.cerne/config.json` for the current user.
`set` persists a supported language, `get` prints the saved value, and `unset` removes the
preference so the default applies. Use a temporary override to repair an invalid
regular configuration, for example:

```sh
cerne --lang en config set language en
```

Cerne refuses symbolic links and unsafe configuration paths instead of following them.

## `cerne init <project-name> [--source ... | --clone ...] [--workflow ... [--agent ...]]`

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

## `cerne restore <knowledge-origin> (--source <path> | --clone <source-origin>)`

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

## `cerne skill install <codex|claude|gemini> [cerne-context|cerne-git-workflow]`

Without a skill argument, installs every compatible official skill for the current user's agent
profile. Codex and Claude receive `cerne-context` and `cerne-git-workflow`; Gemini receives
`cerne-git-workflow` only. With a skill argument, installs exactly that skill. Destinations are
`~/.codex/skills/<skill>`, `~/.claude/skills/<skill>`, or `~/.gemini/skills/cerne-git-workflow`.

The command uses the official `cerne-skills` package embedded in the binary, without network access,
validates the manifest, adapter, entrypoint, and `cerne.context.v1` schema before copying, and writes
a private audit record under `~/.cerne/audit` for each installed skill.

Invalid usage, including `generic`, case variants, missing agents, or extra arguments, returns
status `2` without audit or filesystem mutation. Operational failures return stderr/status `1`.
Reinstalling the same version is a no-op; different managed versions are upgraded. `init`,
`restore`, and `workflow setup` never install skills by implication.

## `cerne git inspect`

Provides the safe Git inspection surface used by `cerne-git-workflow`. Cerne does not execute Git
effects; the agent uses the inspected data and asks for confirmation before branch, commit, push,
or Pull Request work.

```sh
cerne git inspect --agent codex --task task-1 --json
```

`inspect` is read-only and returns schema version 1 with a deterministic `state_id`, sanitized
remotes, local branches, changed literal paths, and a private audit id. Branch, commit, push, and
Pull Request commands are intentionally not available through Cerne; unsupported or destructive Git
operations remain out of scope for the skill.

JSON success returns stdout/status `0`; an invalid workspace snapshot returns stdout/status `1`;
invalid usage returns stderr/status `2`. Private audit records under `~/.cerne/audit` cover only
`inspect` and exclude conversations, Git output, remote URLs, tokens, file contents, PR body, and
raw errors. Execution evidence belongs to the agent or harness, not to the Cerne audit.

## `cerne workflow setup [--agent codex|claude]`

Locates the nearest ancestor workspace and materializes the provider declared in its manifest. It
does not accept a provider, path, or force option. With `--agent`, the declared workflow must be
Spec Kit and Cerne prepares or refreshes the workspace-root discovery bridge for the selected local
agent. Each real provider or agent-integration subprocess creates one redacted JSON audit record in
`knowledge/runs`; no audit is created for a missing executable or ready layout without agent setup.
For Codex to discover the bridge in `.agents/skills`, start the session at the Cerne workspace root,
not inside `source/`.

## `cerne context [--json]`

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

## `cerne doctor`

Performs ten read-only checks from the workspace root: manifest readability, both repository
directories, Git independence, versioning isolation, manifest paths, required knowledge
directories, Git availability, permissions, and manifest version. A declared workflow adds a check
for ready, pending, unavailable, unknown, partial, or nested-Git state. It never repairs the workspace.

## `cerne status`

Locates the nearest workspace from the current directory and reads both repositories. It recognizes
clean and pending worktrees, detached HEAD, and repositories without commits. It does not fetch or
compare with remotes.

## `cerne link <path> [--replace]`

Links a local non-bare Git repository with a worktree as `source`. Relative and absolute paths are
accepted, including valid Git worktrees. Knowledge and source must be distinct and must not be
dangerously nested. Replacing a different configured source requires `--replace`; linking the same
source succeeds without rewriting the manifest. Manifest replacement is atomic.

<!-- END AUTO-GENERATED -->
