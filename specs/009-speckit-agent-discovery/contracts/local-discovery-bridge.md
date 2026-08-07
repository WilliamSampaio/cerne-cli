# Contract: Ponte Local de Descoberta de Agente

## Purpose

A ponte local torna comandos Spec Kit descobríveis pelo agente iniciado na raiz do workspace Cerne,
enquanto preserva `knowledge` como raiz real do projeto Spec Kit.

## Workspace layout

Codex:

```text
workspace/
├── .agents/
│   └── skills/
│       └── speckit-*/SKILL.md
├── knowledge/
│   ├── .agents/skills/speckit-*/SKILL.md
│   ├── .specify/
│   └── specs/
└── source/
```

Claude:

```text
workspace/
├── .claude/
│   └── skills/
│       └── speckit-*/SKILL.md
├── knowledge/
│   ├── .claude/skills/speckit-*/SKILL.md
│   ├── .specify/
│   └── specs/
└── source/
```

## Required command set

The bridge MUST expose these commands for the selected agent:

```text
speckit-analyze
speckit-checklist
speckit-clarify
speckit-constitution
speckit-converge
speckit-implement
speckit-plan
speckit-specify
speckit-tasks
speckit-taskstoissues
```

## Behavioral requirements

- Each bridge artifact MUST direct the agent to treat `knowledge` as the Spec Kit project root.
- Bridge artifacts MUST use workspace-relative references where possible and MUST NOT persist
  absolute paths.
- Bridge artifacts MUST NOT contain private knowledge content, provider stdout/stderr, environment
  variables, remotes, tokens or credentials.
- The bridge MUST NOT require symlinks.
- The bridge MUST NOT create or modify files under `source`.
- Re-running setup for the same agent MUST be idempotent for the managed command set.
- Preparing another supported agent MAY leave the previous agent's bridge in place; the CLI result
  concerns only the requested agent.
- Existing user files outside the managed command set MUST remain untouched.
- A partial or conflicting managed bridge MUST either be made ready for the requested agent or cause
  an operational failure with a safe correction.
