# Contract: Context Report JSON Schema v1

## Encoding

- UTF-8 JSON, indentação de dois espaços e exatamente um newline final.
- Ordem top-level: `schema_version`, `status`, `workspace`, `knowledge`, `source`, `workflow`,
  `problems`; objetos opcionais ausentes são omitidos.
- Ordem interna segue a definição abaixo; `problems` sempre é array.
- Nenhum campo pode conter erro bruto ou mensagem localizada.

## Shape

```text
ContextReport {
  schema_version: 1,
  status: "healthy" | "warning" | "invalid",
  workspace?: { name?: string, root: native-absolute-path },
  knowledge?: {
    path: native-absolute-path,
    product_path?: native-absolute-path,
    specs_path?: native-absolute-path,
    decisions_path?: native-absolute-path,
    policies_path?: native-absolute-path
  },
  source?: { path: native-absolute-path, inside_workspace: boolean },
  workflow?: {
    declared: boolean,
    provider?: string,
    state: "not-declared" | "pending" | "ready" | "invalid" | "unknown-provider"
  },
  problems: [{
    code: ProblemCode,
    severity: "warning" | "error",
    component: Component
  }]
}
```

`ProblemCode` é exatamente um de:

```text
workspace-not-found
manifest-invalid
manifest-version-unsupported
knowledge-invalid
source-invalid
required-directory-invalid
workflow-pending
workflow-invalid
workflow-unknown-provider
```

`Component` é exatamente um dos componentes definidos em [data-model.md](../data-model.md).

## Healthy example

```json
{
  "schema_version": 1,
  "status": "healthy",
  "workspace": {
    "name": "exemplo",
    "root": "/work/exemplo"
  },
  "knowledge": {
    "path": "/work/exemplo/knowledge",
    "product_path": "/work/exemplo/knowledge/product",
    "specs_path": "/work/exemplo/knowledge/specs",
    "decisions_path": "/work/exemplo/knowledge/decisions",
    "policies_path": "/work/exemplo/knowledge/policies"
  },
  "source": {
    "path": "/work/exemplo/source",
    "inside_workspace": true
  },
  "workflow": {
    "declared": false,
    "state": "not-declared"
  },
  "problems": []
}
```

## Invalid example without workspace

```json
{
  "schema_version": 1,
  "status": "invalid",
  "problems": [
    {
      "code": "workspace-not-found",
      "severity": "error",
      "component": "workspace"
    }
  ]
}
```

Este documento continua válido mesmo com status de processo 1: o JSON inteiro é emitido em stdout.

## Omission and status rules

- Um objeto só aparece quando sua identidade mínima foi comprovada.
- Campos filhos não comprovados são omitidos sem `null`, string vazia ou fallback.
- Provider desconhecido aparece em `workflow.provider`, usa estado `unknown-provider` e omite
  `knowledge.specs_path`.
- `pending` adiciona `workflow-pending`/`warning`; qualquer outro problema estrutural listado usa
  `error`.
- Specs ausente, irregular ou symlink depois de a raiz do provider existir produz, nesta ordem,
  `required-directory-invalid` em `knowledge.specs` e `workflow-invalid` em `workflow`.
- Problemas repetidos de diretórios usam a ordem fixa product, specs, decisions, policies, runs.

## Schema evolution

Schema 1 aceita somente adições compatíveis de campos opcionais. Remoção, renomeação, mudança de
tipo ou semântica de campo existente — inclusive ampliar enums ou o catálogo fechado de códigos —
exige incremento de `schema_version` e a política de compatibilidade da CLI.
