# Data Contract: Source Clone Audit

Cada processo de clone corresponde ao arquivo `knowledge/runs/source-clone.json`:

```json
{
  "kind": "source-clone",
  "executor": "git",
  "operation": "clone",
  "project": "example",
  "destination": "../source",
  "origin_transport": "https",
  "origin_fingerprint": "<sha256-hex>",
  "authorization": "--clone",
  "status": "succeeded",
  "started_at": "2026-08-05T12:00:00Z",
  "finished_at": "2026-08-05T12:00:01Z"
}
```

## Rules

- Criação exclusiva e durável em `started` ocorre antes do processo.
- O mesmo arquivo é substituído atomicamente para `succeeded` ou `failed`.
- `finished_at` existe somente no estado final.
- `failure`, quando presente, é categoria Cerne estável como `clone-failed`,
  `invalid-result` ou `cleanup-failed`.
- `origin_fingerprint` é SHA-256 da localização exata e serve somente para verificar correlação
  quando o usuário reapresenta o valor.
- URL, host, path da origem, userinfo, credencial, ambiente e output Git são proibidos.
- `knowledge/runs/.gitkeep` não é registro de tentativa.

Se a finalização atômica falhar antes da promoção, o arquivo permanece `started` e somente o
staging privado é removido. Se falhar depois da promoção, o source validado permanece. Nos dois
casos o init retorna falha. Uma interrupção abrupta também pode deixar `started`, sinalizando
tentativa inconclusiva.
