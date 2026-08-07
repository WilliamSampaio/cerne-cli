# Quickstart: Validar Descoberta de Agente para Spec Kit

## Prerequisites

- Binary built from this feature branch: `go build -o cerne ./cmd/cerne`.
- Git available in PATH.
- For configured-success scenarios, `specify` available in PATH.
- Disposable parent directories. Do not run provider-failure scenarios against valuable workspaces.

Use `<binary-path>` below for the absolute path of the built Cerne binary.

## Scenario 1: Legacy Spec Kit setup remains unchanged

```text
<binary-path> init legacy --workflow speckit
```

Expected: status zero, existing stdout/stderr contract, `knowledge/cerne.json` has only
`workflow.provider`, workflow is materialized in `knowledge`, and no workspace-root `.agents` or
`.claude` bridge is created by Cerne.

## Scenario 2: Codex discovery during init

```text
<binary-path> init codex-project --workflow speckit --agent codex
```

Expected: status zero, stderr empty, stdout includes `Agent: codex` and `Descoberta: pronta`,
`knowledge/.specify/init-options.json` exists, `knowledge/.agents/skills/speckit-*/SKILL.md` exists,
workspace-root `.agents/skills/speckit-*/SKILL.md` exists, `source` is unchanged, and
`knowledge/cerne.json` has no agent field.

## Scenario 3: Claude discovery during init

```text
<binary-path> init claude-project --workflow speckit --agent claude
```

Expected: status zero, stderr empty, stdout includes `Agent: claude` and `Descoberta: pronta`,
`knowledge/.claude/skills/speckit-*/SKILL.md` exists, workspace-root
`.claude/skills/speckit-*/SKILL.md` exists, and no agent field is persisted in the manifest.

## Scenario 4: Restore or pending workspace chooses local agent later

Create or restore a workspace whose manifest declares `speckit`, then run from the workspace root or
any descendant:

```text
<binary-path> workflow setup --agent codex
```

Expected: status zero when Spec Kit is available, workflow ready in `knowledge`, Codex bridge ready
at the workspace root, no source mutation, and manifest unchanged except for preexisting
workflow/provider data.

## Scenario 5: Switch local agent after previous choice

In a workspace prepared for Codex:

```text
<binary-path> workflow setup --agent claude
```

Expected: status zero, Claude bridge ready at workspace root, `knowledge` remains the Spec Kit
root, `source` is unchanged, and `knowledge/cerne.json` still has no agent field. Existing Codex
bridge may remain, but the command reports only Claude readiness.

## Scenario 6: Missing Spec Kit does not create fake bridge

Run with PATH containing Git but not `specify`:

```text
<binary-path> init pending --workflow speckit --agent codex
```

Expected: status zero with setup pending warning, no audit, no workspace-root bridge, and correction
mentions `cerne workflow setup --agent codex`.

## Scenario 7: Invalid agent usage has no effects

In disposable empty parents, try:

```text
<binary-path> init invalid --agent codex
<binary-path> init invalid --workflow openspec --agent codex
<binary-path> init invalid --workflow speckit --agent other
<binary-path> workflow setup --agent
<binary-path> workflow setup --agent codex extra
```

Expected: every command returns status two, stdout empty, stderr prints the documented usage and no
workspace or bridge artifacts are created.

## Scenario 8: Discovery failure remains recoverable

Use a controlled fixture where workflow materialization succeeds but creating the workspace-root
bridge fails because the managed bridge path is unsafe or unwritable.

Expected: status one, safe stderr cause and correction, workflow audit preserved if provider ran,
`knowledge` and `source` intact, and no bridge reported as ready.
