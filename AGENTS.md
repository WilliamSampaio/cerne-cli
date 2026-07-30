# Repository Guidelines

## Project Structure & Module Organization

The CLI entry point is `cmd/cerne/main.go`; its binary-level integration tests live beside it in
`cmd/cerne/main_test.go`. Domain behavior belongs in `internal/workspace`, while interaction with
the local Git executable is isolated in `internal/gitexec`. Keep tests next to their package using
`*_test.go`.

Feature artifacts are stored under `specs/<number>-<feature>/` and normally include `spec.md`,
`plan.md`, `tasks.md`, contracts, and a quickstart. Project governance is defined in
`.specify/memory/constitution.md`; reusable Spec Kit workflows and templates live in `.specify/`.
CI configuration is in `.github/workflows/test.yml`.

## Build, Test, and Development Commands

- `go build -o cerne ./cmd/cerne` — build the local CLI binary.
- `go test ./...` — run domain, adapter, and CLI integration tests.
- `go vet ./...` — detect common Go correctness problems.
- `gofmt -w <files>` — format changed Go files before committing.
- `go test -count=1 ./...` — rerun the suite without cached results.

Git must be available in `PATH`. Tests require no network or credentials.

## Coding Style & Naming Conventions

Target the Go version declared in `go.mod` and prefer the standard library. Use `gofmt` formatting,
lowercase package names, `CamelCase` exported identifiers, and concise filenames such as `init.go`.
Keep domain rules independent of `os/exec`; external tools belong behind adapters. Do not add
frameworks, interfaces, or configuration for hypothetical future requirements. Preserve the
documented Portuguese CLI diagnostics and stable stdout, stderr, and exit-code contracts.

## Testing Guidelines

Use Go’s `testing` package with descriptive `TestBehavior` names and `t.TempDir()` isolation.
Behavior changes and bug fixes need a test that fails without the change. Exercise Git adapters
with temporary local repositories only—never real remotes. Critical CLI tests must verify arguments,
exact streams, exit status, filesystem effects, rollback, and Linux/Windows/macOS portability.

## Commit & Pull Request Guidelines

History follows Conventional Commit-style subjects, for example `feat: implement cerne init` and
`docs: amend Cerne constitution`. Use `feat`, `fix`, `docs`, `test`, or `chore` with a short
imperative summary.

Pull requests should explain intent, link the relevant issue or `specs/` artifact, list validation
commands, and call out public CLI compatibility changes. Include before/after terminal output when
stdout, stderr, help, or status codes change.

## Security & Workspace Safety

Never commit secrets, credentials, or private knowledge repositories. Generated workspaces are not
project fixtures unless explicitly documented. Push, merge, publication, deployment, and
destructive operations require explicit user authorization.
