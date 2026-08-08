# Quickstart: Validate CLI Internationalization

## Prerequisites

- Go version declared by `go.mod`.
- Git available in `PATH` for existing workspace command tests.
- An isolated temporary home directory for manual configuration checks.

## Automated Validation

```bash
go test -count=1 ./...
go vet ./...
```

The suite must validate catalog parity, preference safety, language precedence, exact localized streams, unchanged exit statuses, and language-neutral JSON.

## Persistent Preference

With `HOME` directed to an isolated directory:

```bash
cerne config set language en
cerne config get language
cerne --help
cerne config set language pt-BR
cerne --help
cerne config unset language
```

Expected results:

- `config get` reports the saved canonical identifier.
- Help switches completely between English and Brazilian Portuguese.
- Unset returns to the compatibility default `pt-BR`.
- The preference file is valid JSON and never partially written.

## Temporary Selection

```bash
cerne config set language pt-BR
cerne --lang en doctor
CERNE_LANG=en cerne doctor
cerne config get language
```

The temporary invocations use English while the final command confirms that `pt-BR` remains saved.

## Automation Invariance

Run `cerne context --json` against the same workspace once with `--lang en` and once with `--lang pt-BR`. The complete JSON bytes and exit status must match.

## Safety Refusal

Replace the isolated `~/.cerne` or `~/.cerne/config.json` with a symbolic link and attempt `config set`. The command must fail without modifying the link target.
