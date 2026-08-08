# Cerne Context Contract

Expected command:

```bash
cerne context --json
```

The Cerne CLI owns workspace discovery, manifest interpretation, source resolution, workflow detection, status classification, and path normalization. Skills consume the contract; they do not duplicate those rules.

Schema v1 shape:

```json
{
  "schema_version": 1,
  "status": "healthy",
  "workspace": {
    "name": "my-project",
    "root": "/path/to/workspace"
  },
  "knowledge": {
    "path": "/path/to/workspace/knowledge",
    "product_path": "/path/to/workspace/knowledge/product",
    "specs_path": "/path/to/workspace/knowledge/specs",
    "decisions_path": "/path/to/workspace/knowledge/decisions",
    "policies_path": "/path/to/workspace/knowledge/policies"
  },
  "source": {
    "path": "/path/to/workspace/source",
    "inside_workspace": true
  },
  "workflow": {
    "declared": true,
    "provider": "speckit",
    "state": "ready"
  },
  "problems": []
}
```

Required semantics:

- `schema_version` is the integer compatibility version. This skill supports `1`.
- `status` is `healthy`, `warning`, or `invalid`.
- `workspace.root`, `knowledge.path`, and `source.path` are absolute paths when present.
- `knowledge.*_path` fields are present only when the CLI proved those directories.
- `source.inside_workspace` tells the agent whether source is inside or outside the workspace.
- `workflow.declared` tells whether the manifest declares a workflow.
- `workflow.provider` is present when declared and supported by the report.
- `workflow.state` is `not-declared`, `pending`, `ready`, `invalid`, or `unknown-provider`.
- `problems` contains public diagnostic objects with `code`, `severity`, and `component`.

Process semantics:

- Exit status `0` can return `healthy` or `warning`.
- Exit status `1` can still return valid JSON with `status: "invalid"` on stdout.
- Exit status `2`, command failure, malformed JSON, or unsupported schema stops the skill before source work.

The payload does not list documents or local instruction files. The agent inventories the proven directories progressively and reads `AGENTS.md` or `CLAUDE.md` from source when relevant.
