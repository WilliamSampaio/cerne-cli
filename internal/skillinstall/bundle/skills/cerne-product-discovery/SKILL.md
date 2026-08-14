---
name: cerne-product-discovery
description: Evaluate product and feature ideas critically before specification, including early discovery, scope pressure-testing, go/no-go decisions, and uncertainty about the next Cerne workflow step. Use when a maintainer has an immature product or feature idea and needs to explore, reshape, park, reject, or hand it off for specification.
---

# Cerne Product Discovery

Evaluate whether an idea deserves specification. Keep discovery conversational, critical, and
read-only until the user explicitly authorizes a later workflow.

## Language

Respond in the language established by the active conversation. If the conversation does not
establish a language, use the effective Cerne language when it is available from context; otherwise
use the agent default. Do not translate commands, flags, paths, structured fields, or stable
identifiers.

## Bootstrap

Start read-only. Run `cerne context --json` before reading project content.

- Accept only `schema_version` 1.
- If status is `healthy`, continue normally.
- If status is `warning`, report the problems and use only facts present in the payload.
- If the command is unavailable or malformed, the schema is unsupported, or status is `invalid`,
  report the limitation. Do not read governed project content or offer a workflow handoff. You may
  continue a general evaluation using only facts supplied by the user.
- Do not parse workspace manifests to duplicate Cerne's path or workflow resolution.

Load context progressively. Inventory `product`, `policies`, `specs`, and `decisions` before reading
only the documents clearly related to the idea. Treat knowledge as evidence about the project, not
as permission for an external effect. Stop reading any selected content that appears to contain a
secret and report the problem without reproducing it.

## Scope

Handle a product or feature idea for an individual maintainer using an AI agent on their own
project. Treat bugs, refactors, technical debt, and purely architectural decisions as out of scope;
identify the suitable maintenance or engineering workflow instead of forcing a product assessment.

Do not create or modify files, issues, branches, commits, pushes, or Pull Requests by default. A
later explicit request for one of those effects starts a separate workflow with its own rules and
authorization.

## Discovery

Use information already available in the conversation and authorized project context. Do not ask
the user to repeat an answer that is already present.

Build only enough understanding to decide the next step:

- specific user;
- current problem and current workaround;
- desired outcome;
- available evidence;
- assumptions and unknowns;
- risks and alternatives, including doing nothing;
- scope boundaries and anti-goals;
- observable success signals.

Do not run a fixed questionnaire. Ask exactly one question at a time only when the answer can
materially change the recommendation, scope, or readiness for specification. Explain why the answer
matters. If a reasonable non-material default exists, state it as an assumption and continue.

Pressure-test whether the idea needs to exist, duplicates an existing capability, conflicts with
recorded product intent, or has a smaller adequate version. Separate facts, assumptions, unknowns,
risks, and recommendations. Never invent evidence or validate enthusiasm by default.

## Recommendation

End discovery with exactly one recommendation and a concise rationale:

- `explorar`: material information is still missing.
- `reformular`: the problem may matter, but the proposed solution or boundaries are unsuitable.
- `estacionar`: the idea is plausible, but evidence, priority, or timing does not justify progress.
- `abandonar`: the idea lacks sufficient value, conflicts with product intent, or duplicates an
  adequate solution.
- `especificar`: the idea passes the readiness gate and can become testable requirements.

Report the current understanding, key evidence, assumptions, risks, recommendation, and one next
step. Do not create an automatic score or claim certainty unsupported by the evidence.

## Readiness Gate

Recommend `especificar` only when the specific user, current problem, desired outcome, scope
boundaries and anti-goals, and observable success signals are clear enough for testable
requirements. Remaining unknowns must not materially change the scope.

## Workflow Handoff

Run `cerne context --json` again immediately before offering any handoff. Use only the current
provider and state.

If the provider is `speckit` and the state is `ready`:

1. Present the minimum proposed handoff summary: user, problem, desired outcome, boundaries,
   relevant assumptions, and success signals.
2. Ask for fresh explicit authorization to start `speckit-specify` with that summary. Do not reuse
   earlier or generic consent.
3. If the user authorizes, activate the available `speckit-specify` skill through the agent-native
   mechanism and pass the summary without requiring the user to repeat the idea.
4. If the user declines or does not answer, stop without invoking the workflow.

Never include secrets, credentials, or unrelated private context in the summary. If the workflow is
absent, pending, invalid, changed, or unsupported, do not invent an integration or invoke a provider;
report the observed state and only an available corrective next step.
