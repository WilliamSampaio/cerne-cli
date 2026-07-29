<!--
Sync Impact Report
- Version change: template (unratified) → 1.0.0
- Modified principles:
  - Template Principle 1 → I. Specification Is the Source of Truth
  - Template Principle 2 → II. Stable CLI Contracts
  - Template Principle 3 → III. Simplicity First
  - Template Principle 4 → IV. Proportionate Verification
  - Template Principle 5 → V. Safe, Explicit Failure
- Added sections:
  - Engineering Constraints
  - Development Workflow
- Removed sections: None
- Templates:
  - ✅ updated: .specify/templates/plan-template.md
  - ✅ compatible, no change: .specify/templates/spec-template.md
  - ✅ updated: .specify/templates/tasks-template.md
  - ✅ compatible, no change: .specify/templates/checklist-template.md
- Runtime guidance:
  - ✅ reviewed, no change: README.md
  - ✅ updated: .agents/skills/speckit-tasks/SKILL.md
  - ✅ reviewed, no change: remaining .agents/skills/speckit-*/SKILL.md
- Follow-up TODOs: None
-->
# cerne-cli Constitution

## Core Principles

### I. Specification Is the Source of Truth
Every feature MUST have testable requirements and acceptance scenarios before implementation.
Plans and tasks MUST trace work to a requirement, user story, acceptance scenario, or explicit
constitution rule. Implementation outside that scope MUST be removed or separately specified.
This keeps intent reviewable and prevents undocumented behavior from becoming accidental policy.

### II. Stable CLI Contracts
User-facing commands MUST define accepted arguments and input, stdout output, stderr diagnostics,
exit status, and side effects. Existing contracts MUST remain compatible unless the specification
identifies the break and provides a migration path. Human-readable output is the default; a
machine-readable format MUST be stable when provided. Explicit contracts make the CLI scriptable
and predictable.

### III. Simplicity First
Changes MUST use the smallest design that satisfies current requirements. Existing project code,
the Go standard library, and native platform behavior MUST be preferred, in that order, before a
new dependency or abstraction. New dependencies and complexity MUST be justified in the plan's
Complexity Tracking section. Speculative extension points are prohibited.

### IV. Proportionate Verification
Every change with branching, parsing, state mutation, security impact, or bug-regression risk MUST
include the smallest automated check that would fail if the behavior regressed. CLI contract
changes MUST include an end-to-end or integration check of output and exit status. Trivial
documentation or formatting-only changes MAY omit tests when the plan records why no behavioral
check applies.

### V. Safe, Explicit Failure
Inputs from users, files, environment variables, and external processes MUST be validated at their
trust boundary. Failures MUST return a non-zero status, place actionable diagnostics on stderr,
and avoid partial or silent state changes. Destructive or irreversible actions MUST require
explicit user intent and MUST identify their exact target before execution.

## Engineering Constraints

- The Go version declared in `go.mod` is authoritative.
- Code MUST pass the repository's formatter, static checks, and relevant automated checks.
- Secrets and sensitive values MUST NOT be committed, logged, or emitted in diagnostics.
- Dependencies MUST be pinned through Go modules and added only when Principle III is satisfied.
- Documentation MUST describe user-visible command or compatibility changes in the same change.

## Development Workflow

Work MUST proceed through specification, planning, dependency-ordered tasks, and implementation.
The Constitution Check in `plan.md` MUST pass before research and MUST be repeated after design.
Tasks MUST preserve requirement or user-story traceability and place required verification before
the implementation it protects. Reviews MUST confirm scope, CLI compatibility, simplicity,
verification, and safe failure behavior before merge.

## Governance

This constitution supersedes conflicting project guidance. Amendments MUST document the reason,
affected principles or sections, migration impact, and synchronized template changes in the Sync
Impact Report. Maintainers approve amendments through normal review.

Versions follow semantic versioning: MAJOR for incompatible governance changes or removed or
redefined principles, MINOR for new principles or materially expanded obligations, and PATCH for
clarifications without changed obligations. Every feature plan and code review MUST verify
compliance. A justified exception MUST be recorded in the plan's Complexity Tracking section with
the simpler alternative considered; unexplained violations block implementation.

**Version**: 1.0.0 | **Ratified**: 2026-07-28 | **Last Amended**: 2026-07-28
