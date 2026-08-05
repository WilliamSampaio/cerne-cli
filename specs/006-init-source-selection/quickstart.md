# Quickstart: Validar Seleção de Source no Init

## Prerequisites

- Binary built from this feature branch: `go build -o cerne ./cmd/cerne`.
- Git available in PATH.
- Repositories and parent directories disposable and local.
- No real network, credential or valuable repository in failure scenarios.

Use `<binary-path>` for the absolute Cerne binary path.

## Scenario 1: Default init is unchanged

```text
<binary-path> init default
```

Expected: exact legacy three-line stdout, empty stderr, status zero, two empty independent Git
repositories, legacy manifest and five `.gitkeep`; no audit or network.

## Scenario 2: Existing local source

Create a temporary non-bare Git repository with one commit outside the target workspace, record a
byte snapshot, and run:

```text
<binary-path> init local --source <existing-repository>
```

Expected: status zero, `Source vinculado`, portable manifest path, no `local/source`, independent
knowledge Git, and an identical before/after source snapshot.

## Scenario 3: Relative path and worktree

Invoke init from a directory with spaces and Unicode using a relative path to a temporary linked
worktree.

Expected: canonical working-tree root is persisted, source remains untouched and doctor/status
resolve it successfully.

## Scenario 4: Invalid local sources

Try missing path, file, bare repository, Git subdirectory, workspace ancestor and knowledge path.

Expected: status one, safe correction, empty stdout and no target workspace.

## Scenario 5: Clone a populated local origin

Create a temporary origin with commits, branch and tag, then run:

```text
<binary-path> init cloned --clone <local-origin>
```

Expected: status zero, `Source clonado`, empty stderr, independent knowledge/source roots,
complete history, checkout, tag, remote named `origin`, unchanged origin, no staging remnant and one
succeeded audit.

## Scenario 6: Clone an empty origin

Use a temporary empty bare origin.

Expected: status zero; source is a valid Git root with `origin` and no commits; the workspace is
accepted without inventing a branch or commit.

## Scenario 7: Reject unsafe origins before effects

Try option-like values, HTTP, `git://`, `ext::`, unknown helper transport, HTTPS userinfo/query and
URL-embedded password.

Expected: status two for invalid syntax/location, no Git process, audit or workspace, and no secret
in stdout/stderr.

## Scenario 8: Clone failure remains auditable

Use a fake Git executable that validates exact argv/environment, creates partial source and exits
non-zero.

Expected: status one; knowledge, manifest and exactly one failed audit remain; private partial
staging is removed and final source remains absent; URL and fake token-like stderr appear nowhere.

## Scenario 9: Audit failure blocks clone or success

Inject failure first when creating `started`, then when finalizing a successful clone.

Expected: initial audit failure runs no clone and rolls back fully. Finalization failure after
promotion returns one and preserves both the validated source and knowledge with a `started` audit.

Also inject a concurrent `source` immediately before promotion. Expected: it is neither replaced
nor removed; staging is cleaned, audit becomes failed and the command returns one.

## Scenario 10: Invalid CLI shapes

Try missing values, repeated flags, both flags, reordered flags and extras.

Expected: status two, exact usage on stderr, empty stdout, no Git lookup and no filesystem effect.

## Scenario 11: Existing commands accept results

Run `doctor`, `status` and `link` against successful default, local and clone workspaces, then
against a failed-clone workspace.

Expected: successful workspaces keep existing command contracts; failed clone receives the
existing blocking missing-source diagnosis without a new state machine.

## Scenario 12: Portability and regression

Run all automated fixtures in CI on Linux, Windows and macOS with local origins and fake Git only.

Expected: equivalent functional results, no shell, network, real credentials, external remotes or
platform-only path assumptions.

## Automated validation

```text
gofmt -w <changed-go-files>
go vet ./...
go test -count=1 ./...
git diff --check
```
