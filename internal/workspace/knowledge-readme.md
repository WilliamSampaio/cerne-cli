# {{PROJECT_NAME}} knowledge

This repository stores the durable context needed to understand and evolve the project. Application
code belongs in the separate source repository.

## 1. Validate the workspace

Run these commands from the workspace root before making changes:

```sh
cerne doctor
cerne status
cerne context
```

- `doctor` reports structural problems and their corrections.
- `status` shows the local Git state of knowledge and source without contacting remotes.
- `context` shows the canonical Source, Specs, Product, Decisions, Policies, and Workflow paths.

Resolve blocking `doctor` findings before adding project content.

## 2. Read the context report

Use `cerne context` as the map for this workspace:

- **Source** is where application code lives. It may be inside or outside the workspace.
- **Specs** is the canonical location for feature requirements; do not assume it is always `specs/`.
- **Workflow ready** means the selected provider can be used now.
- **Workflow pending** means its preference is recorded but setup is incomplete; when the provider
  is installed, run `cerne workflow setup`.
- **No workflow declared** means you can manage specifications at the reported Specs path with your
  own process. A managed workflow is optional and is selected when creating a workspace.

## 3. Choose your first useful outcome

You do not need to fill every directory. Pick the path that matches the work in front of you.

### Shape the product

Create `product/overview.md` with the problem, target users, desired outcome, known constraints, and
explicit non-goals. Add evidence and open questions instead of presenting assumptions as facts.

### Continue or start implementation

Open the Source path reported by `cerne context`. If this workspace points to the wrong local source,
read `cerne link --help` before replacing the reference; Cerne does not move or copy source content.

### Define a change

Use the Specs path reported by `cerne context`. If the workflow is ready, invoke its specification
entry point. Otherwise, create a small requirement document that states user value, boundaries, and
verifiable outcomes before planning implementation.

## 4. Record durable guardrails

- `decisions/` — one file per durable product or technical decision, including rationale and rejected
  alternatives;
- `policies/` — project-wide rules that future work must follow;
- `runs/` — sanitized execution and audit records; do not use it as a general notes directory.

Prefer a few current documents over speculative scaffolding.

## 5. Version each repository independently

Knowledge and source are independent Git repositories with separate histories, branches, remotes,
and access policies. Review both with `cerne status`, then commit or configure remotes explicitly in
the repository you intend to change. Cerne does not create commits, pushes, merges, or remotes for you.

## Documentation on GitHub

- [Getting started](https://github.com/WilliamSampaio/cerne-cli/blob/master/docs/en/getting-started.md)
- [Command reference](https://github.com/WilliamSampaio/cerne-cli/blob/master/docs/en/commands.md)
- [Git inspection and workflow guidance](https://github.com/WilliamSampaio/cerne-cli/blob/master/docs/en/commands.md#cerne-git-inspect)
- [Troubleshooting](https://github.com/WilliamSampaio/cerne-cli/blob/master/docs/en/troubleshooting.md)

Do not store secrets or credentials in knowledge or source.
