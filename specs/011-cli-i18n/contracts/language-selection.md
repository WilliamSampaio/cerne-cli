# Contract: Language Selection

## Supported Values

- `en`
- `pt-BR`

Values are canonical and case-sensitive. Other values are usage errors.

## Global Option

```text
cerne --lang <en|pt-BR> <command> [arguments]
```

- The option applies only to that invocation.
- It does not modify `~/.cerne/config.json`.
- Missing or unsupported values write a diagnostic to stderr and exit with status `2`.

## Environment

```text
CERNE_LANG=<en|pt-BR>
```

- The value applies to Cerne processes receiving that environment.
- It does not modify `~/.cerne/config.json`.
- An unsupported value writes a diagnostic to stderr and exits with status `2`.

## Precedence

```text
--lang > CERNE_LANG > ~/.cerne/config.json > pt-BR
```

Resolution stops at the first source that is present. An invalid value is reported; it does not fall through silently.

## Stable Values

Language selection MUST NOT alter:

- command, subcommand, and flag names;
- `cerne --version` output;
- JSON keys, enum values, ordering, indentation, or line endings;
- agent and provider identifiers;
- filesystem paths;
- exit statuses.

## Compatibility Notice

`pt-BR` remains the default for this minor release. `en` is planned as the default for `1.0`; users who require deterministic Portuguese output should persist `pt-BR` explicitly.
