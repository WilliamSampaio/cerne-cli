# Contract: `cerne skill install`

## Syntax

```text
cerne skill install <agent>
cerne skill install --help
cerne skill --help
```

`<agent>` MUST be exactly one of:

- `codex`
- `claude`

`generic`, uppercase variants, empty values, repeated agents and extra arguments are invalid usage.

## Authorization

Running `cerne skill install <agent>` is the explicit authorization to modify only the current
user's profile for that agent. No workspace command grants this authorization by implication.

## Success

Status: `0`  
Stream: stdout only  
Side effects:

- Resolves the official versioned companion/cache `cerne-skills` package without network access.
- Validates the package manifest before destination mutation.
- Installs `cerne-context` in `~/.codex/skills/cerne-context` for Codex or
  `~/.claude/skills/cerne-context` for Claude, resolved inside the current user's home.
- Creates or finalizes one private global audit record.

Required stdout fields:

```text
Skill instalada: cerne-context
Agente: <codex|claude>
Versão: <package-version>
Destino: <display-path>
```

For idempotent reinstall of the same managed version, status remains `0` and stdout MUST make the
no-op clear without rewriting files unnecessarily:

```text
Skill já instalada: cerne-context
Agente: <codex|claude>
Versão: <package-version>
Destino: <display-path>
```

stderr MUST be empty.

For a managed installation in a different version, status remains `0` and the command MUST upgrade
to the companion package version, replacing only files proven to belong to the previous Cerne
installation.

## Operational failures

Status: `1`  
Stream: stderr only  
stdout MUST be empty.

Operational failures include:

- companion/cache package unavailable;
- invalid package or manifest;
- missing `cerne-context`;
- unsupported manifest or context schema;
- destination inaccessible;
- destination contains unknown content;
- symlink/path escape detected;
- audit cannot be created or finalized;
- promotion or rollback failure.

stderr MUST contain:

```text
erro: <safe cause>
correção: <safe recovery instruction>
```

stderr MUST NOT include secrets, environment variables, embedded credentials, remotes or raw output
from external tools.

## Invalid usage

Status: `2`  
Stream: stderr only  
stdout MUST be empty.  
No audit is created and no file is changed.

Required form:

```text
erro: argumento inválido
uso: cerne skill install <codex|claude>
```

## Workspace command boundary

The following commands MUST NOT install skills or alter global agent profiles:

```text
cerne init <project-name> --workflow speckit --agent <codex|claude>
cerne restore <origem-knowledge> (--source <caminho> | --clone <origem-source>)
cerne workflow setup --agent <codex|claude>
```

They MAY print safe guidance suggesting:

```text
cerne skill install <agent>
```

but MUST NOT invoke it.
