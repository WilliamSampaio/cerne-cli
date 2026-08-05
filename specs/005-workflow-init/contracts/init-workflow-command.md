# CLI Contract: `cerne init` com workflow

## Syntax

```text
cerne init <project-name>
cerne init <project-name> --workflow speckit
cerne init <project-name> --workflow openspec
cerne init --help
```

A flag é opcional, aceita somente a posição documentada e não pode ser repetida. O uso sem flag
mantém integralmente o contrato anterior em `specs/001-init-workspace/contracts/init-command.md`.

## Manifest

Com workflow, `knowledge/cerne.json` acrescenta:

```json
"workflow": {
  "provider": "<speckit|openspec>"
}
```

O campo continua ausente no modo padrão. Nenhum estado de instalação ou versão é gravado.

## Configured success

Status: `0`

stdout:

```text
Workspace "<project-name>" criado.
Knowledge: <absolute-knowledge-path>
Source: <absolute-source-path>
Workflow: <provider>
Setup: concluído
```

stderr fica vazio. O provider foi executado, o marker foi validado e a auditoria finalizada.

## Pending success because executable is absent

Status: `0`

stdout:

```text
Workspace "<project-name>" criado.
Knowledge: <absolute-knowledge-path>
Source: <absolute-source-path>
Workflow: <provider>
Setup: pendente
```

stderr:

```text
aviso: executável "<specify|openspec>" não encontrado; workflow <provider> não inicializado
correção: instale <provider> e execute "cerne workflow setup" dentro do workspace
```

Nenhum subprocesso é iniciado e nenhum registro de tentativa é criado.

## Provider execution failure

Status: `1`

stdout fica vazio. stderr:

```text
erro: não foi possível inicializar workflow <provider>: <safe-cause>
correção: corrija ou atualize <provider> e execute "cerne workflow setup" dentro de <workspace-path>
```

O workspace base, manifesto e auditoria permanecem. Somente a owned root ausente antes da tentativa
pode ser removida. Source e arquivos preexistentes permanecem intactos.

## Invalid usage

Status: `2`; stdout vazio. stderr:

```text
erro: <cause>
uso: cerne init <project-name> [--workflow <speckit|openspec>]
```

Aplica-se a valor desconhecido, flag ausente de valor, repetida, fora de posição ou argumento extra.

## Side effects and authorization

`--workflow` autoriza criar o workspace, executar localmente o provider já instalado dentro de
knowledge e registrar a tentativa. Não autoriza instalação, update, telemetria, agente específico,
uso de credenciais, alteração de source, Git remoto, publicação ou deploy.
