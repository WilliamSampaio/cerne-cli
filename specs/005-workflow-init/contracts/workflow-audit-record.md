# Data Contract: Workflow Setup Audit Record

Cada subprocesso de workflow corresponde a exatamente um JSON em `knowledge/runs`. O arquivo
comum `knowledge/runs/.gitkeep` não representa uma tentativa e deve ser ignorado em enumerações e
contagens:

```json
{
  "kind": "workflow-setup",
  "provider": "speckit",
  "executor": "specify",
  "operation": "setup",
  "context": "knowledge",
  "authorization": "workflow setup",
  "status": "succeeded",
  "started_at": "2026-08-04T15:00:00Z",
  "finished_at": "2026-08-04T15:00:02Z"
}
```

## Guarantees

- O registro `started` existe antes do subprocesso.
- A transição final atualiza o mesmo arquivo de forma segura; se essa atualização falhar, o arquivo
  permanece `started`, a operação falha e o workflow não é considerado concluído.
- `failure`, quando existir, usa categoria controlada pelo Cerne.
- Não inclui path absoluto, argumentos completos, ambiente, stdout, stderr, token ou credencial.
- Ausência do executável, layout já válido e recusa prévia não criam registro porque não executam provider.
