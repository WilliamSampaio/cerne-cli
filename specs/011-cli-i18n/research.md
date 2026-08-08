# Research: CLI Internationalization

## Source Catalogs

**Decision**: Keep one compile-time source catalog per supported language and identify messages with semantic keys.

**Rationale**: Two languages need formatting and long help text but no plural rules, date formatting, or runtime catalog loading. Source catalogs are validated at test time, ship inside the binary, and use no new dependency.

**Alternatives considered**:

- `golang.org/x/text/message`: useful when plural and locale-sensitive formatting become requirements, but adds machinery not needed now.
- Embedded JSON/YAML catalogs: friendlier to non-Go tooling, but add runtime parsing and awkward multiline help text.
- Translating existing Portuguese strings directly: rejected because wording changes would silently break lookup and domain text would remain presentation-coupled.

## Domain and Presentation Boundary

**Decision**: Domain failures and report states expose stable semantic codes and values; only CLI renderers select human text.

**Rationale**: A semantic code can be tested, logged, and rendered in any language without comparing prose. This also prevents English and Portuguese from leaking into structured output.

**Alternatives considered**:

- Returning already translated errors from domain packages: rejected because it requires locale state inside domain logic.
- Keeping both Portuguese and English fields in domain results: rejected because every new language would expand domain models.

## Language Resolution

**Decision**: Resolve the first available value in this order: global `--lang`, `CERNE_LANG`, saved preference, compatibility default `pt-BR`.

**Rationale**: Explicit invocation wins, environment supports CI and temporary sessions, persistence supports normal use, and the current default remains compatible. Operating-system locale is deliberately ignored so output does not vary unexpectedly between shells, CI, and containers.

**Alternatives considered**:

- Automatic operating-system locale: rejected because it makes output dependent on ambient host configuration.
- Persistent preference only: rejected because users need a recovery and automation override.

## Preference Storage

**Decision**: Store only the canonical language value in `~/.cerne/config.json`, write through a private temporary file, and atomically replace the destination after validating the `.cerne` directory and target path.

**Rationale**: The location matches the existing user-level Cerne area. Atomic replacement prevents partial JSON, and refusing symbolic links avoids writing outside the intended user path.

**Alternatives considered**:

- `~/.config/cerne`: rejected because Cerne already owns `~/.cerne` and one user root is easier to explain.
- A workspace preference: rejected because language belongs to a user and must not alter repositories.
- Shell-profile edits: rejected because they are invasive and platform-specific.

## Compatibility Rollout

**Decision**: Keep `pt-BR` as the default in this minor release, provide full opt-in English, and announce `en` as the `1.0` default.

**Rationale**: This follows the constitution's deprecation requirement while making the intended future behavior available immediately.

**Alternatives considered**:

- Change to English in the minor release: rejected as an incompatible default change.
- Keep Portuguese permanently: rejected because it does not meet the product's international direction.

## Automation Contract

**Decision**: Localize prose only. Commands, flags, JSON keys and values, provider/agent identifiers, version output, and exit statuses remain unchanged.

**Rationale**: Scripts must not depend on the user's language. Existing `context --json` output must remain byte-for-byte invariant.

**Alternatives considered**:

- Translate structured enum values: rejected because consumers would need locale-specific parsers.
- Add JSON variants to every command in this feature: rejected as unrelated scope; current documented machine contracts remain unchanged.
