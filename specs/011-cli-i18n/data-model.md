# Data Model: CLI Internationalization

## Language

A canonical identifier for a complete human-readable CLI catalog.

| Value | Meaning |
|-------|---------|
| `en` | English |
| `pt-BR` | Brazilian Portuguese |

Validation rules:

- Values are exact and case-sensitive.
- Unsupported values are rejected before command execution.
- Command names, flags, structured values, and paths are not language values and are never translated.

## User Configuration

The current user's optional persisted preferences.

| Field | Required | Validation |
|-------|----------|------------|
| `language` | No | When present, one supported Language value |

Storage contract:

```json
{
  "language": "pt-BR"
}
```

States:

- **Absent**: no saved preference; use the release default unless a temporary selection exists.
- **Valid**: contains a supported language.
- **Invalid**: malformed, unsafe, unreadable, or contains an unsupported language; normal resolution fails with recovery guidance.

Transitions:

- `config set language VALUE`: Absent, Valid, or malformed regular file -> Valid.
- `config unset language`: Valid or malformed regular file -> Absent.
- Failed write: previous state remains unchanged.
- Unsafe path: no transition is attempted.

## Effective Language

The immutable language selected for one process invocation.

Resolution order:

1. Global `--lang` value.
2. `CERNE_LANG` environment value.
3. User Configuration language.
4. Compatibility default `pt-BR`.

Once selected, every human-readable message in that invocation uses the same catalog.

## Message

A semantic identifier and its localized template.

| Attribute | Meaning |
|-----------|---------|
| ID | Stable internal meaning independent of wording |
| English template | Complete English rendering |
| Portuguese template | Complete Brazilian Portuguese rendering |
| Arguments | Language-neutral runtime values inserted into the template |

Catalog validity requires every message ID to exist in both catalogs with compatible formatting arguments.
