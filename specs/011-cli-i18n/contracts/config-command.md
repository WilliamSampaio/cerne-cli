# Contract: Language Configuration Command

## Set

```text
cerne config set language <en|pt-BR>
```

- Validates the language before writing.
- Stores the canonical value at `~/.cerne/config.json`.
- Replaces an existing regular configuration atomically.
- Refuses an unsafe `.cerne` directory or configuration target.
- Success writes a localized confirmation to stdout and exits `0`.
- Invalid usage exits `2`; operational failure exits `1`.

## Get

```text
cerne config get language
```

- Reports the saved canonical value when present.
- Reports that the preference is unset and identifies the current compatibility default when absent.
- Does not modify the filesystem.
- Success exits `0`; invalid usage exits `2`; unreadable or invalid configuration exits `1`.

## Unset

```text
cerne config unset language
```

- Removes the saved language preference.
- An already absent preference succeeds without creating files.
- A malformed regular configuration may be removed as recovery.
- An unsafe path is refused.
- Success writes a localized confirmation to stdout and exits `0`.
- Invalid usage exits `2`; operational failure exits `1`.

## File Format

```json
{
  "language": "en"
}
```

- UTF-8 JSON with one trailing newline.
- Only the `language` field is defined in this version.
- The file contains no secrets and is scoped to the current user.
