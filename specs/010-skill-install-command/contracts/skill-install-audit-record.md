# Contract: Skill Install Audit Record

## Location

Each operational install attempt writes exactly one private global audit record under
`~/.cerne/audit` or the operating-system equivalent of the current user's home.

Invalid usage does not create an audit record.

## Privacy

The audit record MUST NOT contain:

- skill file contents;
- knowledge or source contents;
- environment variables;
- tokens, keys or credentials;
- raw external command output;
- Git remotes or package resolver output.

## JSON shape

```json
{
  "schema_version": 1,
  "operation": "skill.install",
  "agent": "codex",
  "skill": "cerne-context",
  "package": "cerne-skills",
  "package_version": "1.2.3",
  "destination": "/home/user/.codex/skills/cerne-context",
  "status": "succeeded",
  "error_code": "",
  "started_at": "2026-08-07T12:00:00Z",
  "finished_at": "2026-08-07T12:00:01Z"
}
```

## Status transitions

```text
started -> succeeded
started -> failed
```

Failure to create the initial audit record MUST stop the installation before package copy or
destination mutation.

Failure to finalize the audit record MUST return status 1. If finalization fails after filesystem
promotion, implementation MUST preserve the safest recoverable state and report the audit failure.
