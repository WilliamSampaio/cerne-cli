# Quickstart: Validar Inicialização com Workflow SDD

## Prerequisites

- Binary built from this feature branch: `go build -o cerne ./cmd/cerne`.
- Git available in PATH.
- For configured-success scenarios, an officially supported `specify` or `openspec` executable.
- Disposable parent directories. Do not run provider-failure scenarios against valuable workspaces.

Use `<binary-path>` below for the absolute path of the built Cerne binary.

## Scenario 1: Legacy init remains unchanged

```text
<binary-path> init legacy
```

Expected: status zero, empty stderr, exact three-line legacy stdout, legacy manifest without
`workflow`, `knowledge/specs`, and no `.specify` or `openspec` root. Compare with
[the original init contract](../../001-init-workspace/contracts/init-command.md).

## Scenario 2: Spec Kit setup succeeds

With `specify` installed:

```text
<binary-path> init spec-kit-project --workflow speckit
```

Expected: status zero, `workflow.provider` equals `speckit`, stdout says `Setup: concluído`,
`knowledge/.specify/init-options.json` exists, `knowledge/specs` is canonical, source remains an
empty independent Git repository, and one successful workflow audit exists.

## Scenario 3: OpenSpec setup succeeds without telemetry

With `openspec` installed:

```text
<binary-path> init openspec-project --workflow openspec
```

Expected: status zero, provider `openspec`, `knowledge/openspec/config.yaml` exists,
`knowledge/openspec/specs` is canonical, no agent-specific directory is selected by Cerne, source
is unchanged, and one successful audit exists. The run must not create telemetry/global config as
a consequence of the wrapper.

## Scenario 4: Missing provider is non-blocking

Execute in a controlled environment whose PATH contains Git but not the selected provider:

```text
<binary-path> init pending --workflow openspec
```

Expected: status zero; stdout says `Setup: pendente`; stderr contains warning and correction;
manifest records `openspec`; base workspace is healthy apart from a workflow warning; no audit or
provider root exists.

## Scenario 5: Resume a pending workflow

Install the provider required by Scenario 4, enter any directory inside `pending`, and run:

```text
<binary-path> workflow setup
```

Expected: status zero; the provider marker and one successful audit exist; manifest and source are
unchanged; output identifies provider, knowledge path and completion.

## Scenario 6: Setup is idempotent

Repeat:

```text
<binary-path> workflow setup
```

Expected: status zero and `Nenhuma alteração necessária.` No provider process runs and no audit is
added or changed.

## Scenario 7: Invalid workflow usage has no effects

In a disposable empty parent, try missing, unknown, repeated and misplaced values:

```text
<binary-path> init invalid --workflow
<binary-path> init invalid --workflow other
<binary-path> init invalid --workflow speckit --workflow openspec
<binary-path> init --workflow speckit invalid
```

Expected: every command returns status two, uses stderr, prints the documented usage and leaves no
`invalid` directory.

## Scenario 8: Partial provider layout is refused

Create a pending OpenSpec workspace as in Scenario 4, then add
`knowledge/openspec/unrecognized.txt` without a valid config and run setup.

Expected: status one; no provider process or audit is created; the preexisting file remains byte
for byte; the correction asks the user to resolve the partial layout explicitly.

## Scenario 9: Provider execution failure remains auditable

Run the automated integration fixture whose controlled provider creates its owned root and exits
non-zero.

Expected: status one; base workspace, manifest and source remain; provider-owned new partial files
are removed; exactly one failed audit remains; stderr contains only a safe Cerne-owned cause, not
raw provider output.

## Scenario 10: Doctor classifies workflow states

Run `doctor` against fixtures with no workflow, pending setup, ready provider, missing executable,
unknown provider and partial layout.

Expected: legacy workspace keeps its previous report; pending or missing executable produces a
warning; ready produces a pass; unknown or partial produces a blocking error.

## Scenario 11: Audit contains no secrets

Execute a controlled provider with token-like environment variables and token-like stderr, then
inspect the audit JSON and CLI output.

Expected: the record matches [the audit contract](contracts/workflow-audit-record.md), contains no
environment or raw output, and no token value appears in knowledge, stdout or stderr.

## Scenario 12: Source and Git boundaries remain intact

Before configured setup, hash all source entries and record both Git roots/remotes. Repeat after
setup and after a controlled failure.

Expected: source hashes, Git roots, commit counts and remotes are identical; no nested `.git`
exists inside a provider-owned root.

## Automated validation

```text
gofmt -w <changed-go-files>
go vet ./...
go test -count=1 ./...
git diff --check
```

The CI matrix must pass the same suite on Linux, Windows and macOS without real provider installs,
network access or credentials.
