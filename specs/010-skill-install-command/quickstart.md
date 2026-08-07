# Quickstart: Instalação de Skills Cerne

## Prerequisites

- Build the CLI from the feature branch.
- Use a temporary user home for every scenario.
- Use a controlled local `cerne-skills` package fixture through the same companion/cache resolver
  seam used by the CLI; do not use network, GitHub, credentials or real releases in automated
  validation.

```text
go test ./...
```

Contracts:

- [Command contract](contracts/skill-install-command.md)
- [Audit record contract](contracts/skill-install-audit-record.md)

## Scenario 1: Install Codex skill

```text
<cerne> skill install codex
```

Expected: status zero, stderr empty, stdout reports `cerne-context`, `codex`, package version and
destination `<home>/.codex/skills/cerne-context`. Exactly one audit record exists with
`operation: "skill.install"` and `status: "succeeded"`.

## Scenario 2: Install Claude skill

```text
<cerne> skill install claude
```

Expected: same as Codex, but destination is `<home>/.claude/skills/cerne-context`.

## Scenario 3: Reinstall same managed version

Run the same command twice with the same package fixture.

Expected: first run installs, second run returns status zero as an idempotent no-op. Managed files
are not rewritten unnecessarily, stdout says the skill is already installed and a new operational
audit record is finalized successfully.

## Scenario 4: Reject invalid agents

```text
<cerne> skill install generic
<cerne> skill install Codex
<cerne> skill install codex extra
```

Expected: status two, stdout empty, stderr follows the usage contract, no audit record and no
filesystem mutation.

## Scenario 5: Upgrade managed different version

Start from a valid Cerne-managed install with a different version than the companion package.

Expected: status zero, stderr empty, stdout reports the companion package version, only previously
managed files are replaced, unknown content is preserved and audit is finalized successfully.

## Scenario 6: Reject missing companion package

Run without the companion/cache package available to the CLI resolver.

Expected: status one, stdout empty, stderr explains that the official companion package is missing
or inaccessible, no destination is altered and audit is finalized as failed.

## Scenario 7: Reject unsafe package

Use fixtures with missing manifest, malformed manifest, missing `cerne-context`, incompatible
`contextSchema`, path escape and symlink.

Expected: status one, stdout empty, stderr includes safe cause and correction, existing destination
is unchanged and audit is finalized as failed.

## Scenario 8: Preserve unknown destination

Create files at the intended destination that do not include a Cerne ownership marker.

Expected: status one, stdout empty, stderr explains that existing content is not managed by Cerne,
files remain byte-for-byte unchanged and audit is finalized as failed.

## Scenario 9: Preserve previous managed install on failure

Start from a valid older managed install, then inject a package or filesystem failure before final
promotion.

Expected: status one, previous managed install remains usable, no partial destination is marked as
ready and audit is finalized as failed.

## Scenario 10: Workspace flows do not install skills

Run these commands with a temporary home that starts without skill directories:

```text
<cerne> init app --workflow speckit --agent codex
<cerne> workflow setup --agent claude
<cerne> restore <knowledge-origin> --source <local-source>
```

Expected: none of them creates or alters a global skill destination. They may suggest
`cerne skill install <agent>` in guidance, but never execute it.
