# Troubleshooting

**English** · [Português (Brasil)](../pt-BR/troubleshooting.md) ·
[Español](../es/troubleshooting.md)

[Getting started](getting-started.md) · [Commands](commands.md)

Cerne prints errors and their suggested `correção:` to stderr. Follow that correction first. The
CLI messages are currently in Portuguese, so the relevant messages are quoted below.

## `cerne` or Git is not found

If your shell cannot find `cerne`, add Go's `GOBIN` or `GOPATH/bin` directory to `PATH`. Confirm the
installation with:

```sh
cerne --version
```

For `erro: Git indisponível`, install Git and make sure `git --version` works in the same shell.

## Workspace not found

`erro: workspace Cerne não localizado` means the current directory is not inside a workspace that
contains `knowledge/cerne.json`. Enter the workspace before running `status`, `context`, `link`, or
`workflow setup`.

`cerne doctor` is different: run it from the workspace root.

```sh
cd my-project
cerne doctor
cerne context
```

If `knowledge/cerne.json` was removed or damaged, restore it from the knowledge repository before
continuing.

## `init` refuses the destination

Cerne never replaces existing content. Use a project name whose destination does not exist or is an
empty regular directory. Inspect the directory yourself before removing anything.

An existing local application should be linked instead of used as the destination:

```sh
cerne init my-project --source ../existing-application
```

## A local source is rejected

The path passed to `--source` or `link` must be the root of an existing, non-bare Git repository
with a working tree. It must be independent from `knowledge` and must not overlap protected
workspace paths.

```sh
git -C ../existing-application status
cerne link ../existing-application --replace
```

Use `--replace` only when another source is already configured. It replaces the manifest reference,
not either repository.

## Workflow setup is pending or fails

If `init` reports a missing `specify` or `openspec` executable, the workspace was still created.
Install the selected provider separately and run this inside the workspace:

```sh
cerne workflow setup
```

For `estrutura do workflow inválida ou parcial`, do not rerun setup repeatedly. Inspect and repair
the partial provider-owned directory, then run `cerne doctor` before trying again.

## A source clone leaves an incomplete workspace

A failure after cloning starts may preserve `knowledge` and its redacted audit while leaving the
workspace incomplete. Inspect the record first:

```text
knowledge/runs/source-clone.json
```

Then either associate a valid local source or remove the incomplete workspace manually after
confirming it contains nothing you need. `init` has no resume mode.

## `restore` fails

The destination derived from the restored manifest must not already exist. Authentication and
remote access are handled by Git, so verify that the origins work with your normal Git setup.

Restore has no resume mode. Read the displayed correction and the private record under
`~/.cerne/audit`, fix the cause, and retry. Cerne does not overwrite an existing destination during
recovery.

## The agent skill package is unavailable

`erro: pacote oficial cerne-skills incorporado está inacessível` means the embedded package could
not be materialized or validated. Check access to the system temporary directory and reinstall
Cerne before retrying `cerne skill install <codex|claude|gemini>`.
