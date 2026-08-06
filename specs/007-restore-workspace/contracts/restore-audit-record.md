# Data Contract: Restore Audit Record

Cada tentativa válida cria exatamente um JSON em `<user-home>/.cerne/audit`, antes do primeiro
processo Git. O filename `restore-<opaque-id>.json` não contém nome ou origem.

Exemplo de source clonado concluído:

```json
{
  "kind": "workspace-restore",
  "executor": "cerne",
  "operation": "restore",
  "authorization": "restore --clone",
  "source_mode": "clone",
  "workspace_name": "example",
  "status": "succeeded",
  "started_at": "2026-08-05T22:00:00Z",
  "finished_at": "2026-08-05T22:00:03Z",
  "phases": {
    "knowledge": {
      "operation": "clone",
      "status": "succeeded",
      "started_at": "2026-08-05T22:00:00Z",
      "finished_at": "2026-08-05T22:00:01Z"
    },
    "source": {
      "operation": "clone",
      "status": "succeeded",
      "started_at": "2026-08-05T22:00:01Z",
      "finished_at": "2026-08-05T22:00:03Z"
    }
  }
}
```

No modo local, `authorization` é `restore --source` e `phases.source.operation` é `link`.

## Guarantees

- `.cerne` e `audit` são diretórios regulares privados; symlinks e arquivos são recusados.
- Em POSIX, diretórios pertencem ao usuário atual, usam `0700` e o registro exclusivo usa `0600`.
- No Windows, herança permissiva é desabilitada e a DACL concede acesso somente ao usuário atual e
  `SYSTEM`; estado incompatível é corrigido quando owned com segurança ou recusado.
- O primeiro estado `started` é sincronizado antes de Git. Falha nessa criação impede processos.
- Transições substituem atomicamente o mesmo JSON; uma transição não durável não autoriza a próxima
  fase.
- `knowledge` e `source` são registrados separadamente, inclusive no modo local.
- `failure` usa somente categoria fechada do Cerne, como `clone-failed`, `invalid-manifest`,
  `unsafe-destination`, `invalid-source`, `promotion-failed` ou `cleanup-failed`.
- Um registro que permanece `started` após o comando é inconclusivo por interrupção ou falha de
  atualização/finalização.
- Audit nunca é removido pelo rollback, sucesso, nova tentativa ou retenção automática.

## Forbidden content

O registro não contém:

- origem integral, parcial ou fingerprint;
- host, username, path de origem ou URL remota;
- argumentos completos, environment, stdout ou stderr de processo;
- credencial, token, chave ou conteúdo dos repositórios;
- path absoluto de knowledge/source/staging.

`workspace_name` só aparece depois de validado e é metadado local não sensível. Paths finais são
informados pelo CLI no sucesso, não duplicados no audit.

## Retention

O usuário controla inspeção e remoção manual de `<user-home>/.cerne/audit`. Rotação, expiração,
upload e comando de cleanup estão fora desta versão.
