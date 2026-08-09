---
name: cerne-git-workflow
description: Guide Cerne workspace Git work by inspecting first, then delegating Git effects to the agent with explicit confirmation.
---

# Cerne Git Workflow

Use this skill when helping with branch, commit, push, or GitHub Pull Request work in a Cerne
workspace.

## Required Flow

Follow `Detect -> Suggest -> Confirm -> Execute -> Report`.

1. Detect with the Cerne CLI before suggesting any Git action:

   ```text
   cerne git inspect --agent <codex|claude|gemini> --task <short-task-id> --json
   ```

2. Derive the active work unit only in the conversation. Do not persist it in files, audit,
   workspace metadata, or agent memory.
3. Suggest the action, targets, expected effects, and `state_id`.
4. Ask for explicit user confirmation for exactly one effect.
5. Execute only the exact confirmed effect with the agent's native Git tooling or GitHub capability.
   Cerne provides the inspected repository names, paths, branches, remotes, changed paths, and
   `state_id`; it must not execute branch, commit, push, Pull Request creation, or PR preparation.
6. Report the structured result honestly, including blocked or partial outcomes.

Agent skill activation or consent to load this skill is not authorization for Git.

## Branch Guidance

Suggest a new branch only for a new independent work unit. Continue the current branch when the task
is related to it. If the relationship is uncertain, ask for direction and do not run a mutating
command.

Branch names use:

```text
feat/<slug>
fix/<slug>
refactor/<slug>
chore/<slug>
spec/<slug>
```

Use a short lowercase slug with hyphens. Base branches come from the inspected Git state and must be
passed explicitly for every participant.

Dirty repositories block branch creation. Suggest a checkpoint instead.

## Checkpoints

For commit, push, and Pull Request sequences:

- reinspect before each step;
- ask for a fresh confirmation for that step only;
- before a commit, ask the user to confirm they reviewed every included change;
- use explicit changed paths for commit;
- use explicit remote and branch for push;
- create Pull Requests only on GitHub.com using the agent's native capability, such as its GitHub
  integration or authenticated `gh` CLI;
- do not read, request, store, or pass GitHub tokens through Cerne;
- stop after the first refusal, blocked state, failed result, or partial result.

Prefer the repository's documented commit convention. If none is visible, use a short Conventional
Commit-style subject.

Use this exact safety question before executing a commit:

```text
Do you confirm that you reviewed every change included in this commit?
```

## Refusals

Refuse destructive, ambiguous, arbitrary, or out-of-scope Git operations, including merge, rebase,
reset, stash, clean, amend, branch deletion, force push, remote mutation, PR merge, and PR close.

Do not ask Cerne to execute Git effects or pass arbitrary Git flags, URLs, credentials, free
refspecs, absolute paths, path traversal, or pathspec magic to Cerne.

## Reporting

Never claim repositories are aligned unless the CLI result proves it. For partial results, name what
completed, what did not run, and the safe next step: reinspect before deciding.
