# CLI Contract: `cerne init` com `--agent`

## Syntax

```text
cerne init <project-name> --workflow speckit --agent <codex|claude>
cerne init <project-name> --source <caminho> --workflow speckit --agent <codex|claude>
cerne init <project-name> --clone <origem> --workflow speckit --agent <codex|claude>
cerne init --help
```

`--agent` é opcional e só é aceito quando `--workflow speckit` também está presente. A ordem das
opções segue o contrato atual de `init`: opções aparecem depois do nome e podem ser combinadas em
pares. `--agent` não aceita `generic`.

## Manifest

Mesmo com agente local, `knowledge/cerne.json` persiste somente:

```json
"workflow": {
  "provider": "speckit"
}
```

Nenhum campo de agente, descoberta local, instalação observada ou versão é gravado.

## Configured success with agent

Status: `0`; stderr vazio.

stdout para source interno:

```text
Workspace "<project-name>" criado.
Knowledge: <absolute-knowledge-path>
Source: <absolute-source-path>
Workflow: speckit
Setup: concluído
Agent: <codex|claude>
Descoberta: pronta
```

stdout para `--source` ou `--clone` preserva as linhas existentes de source vinculado/clonado e
acrescenta as linhas de workflow, agente e descoberta depois delas.

## Pending success because Spec Kit is absent

Status: `0`.

stdout preserva o sucesso pendente existente e não imprime `Agent` nem `Descoberta`.

stderr:

```text
aviso: executável "specify" não encontrado; workflow speckit não inicializado
correção: instale speckit e execute "cerne workflow setup --agent <codex|claude>" dentro do workspace
```

Nenhum subprocesso é iniciado, nenhuma auditoria é criada e nenhuma ponte local é criada.

## Provider or discovery failure

Status: `1`; stdout vazio.

stderr:

```text
erro: não foi possível inicializar workflow speckit: <safe-cause>
correção: corrija ou atualize speckit e execute "cerne workflow setup --agent <codex|claude>" dentro de <workspace-path>
```

Se o provider executou, a auditoria existente em `knowledge/runs` é preservada. Source permanece
intocado. Artefatos de descoberta local não podem ser reportados como prontos quando provider ou
ponte falham.

## Invalid usage

Status: `2`; stdout vazio.

stderr:

```text
erro: argumento inválido
uso: cerne init <project-name> [--source <caminho> | --clone <origem>] [--workflow <speckit|openspec> [--agent <codex|claude>]]
```

Inclui agente sem workflow, agente com OpenSpec, valor desconhecido, valor ausente, flag repetida,
argumento extra ou combinação com source/clone inválida.

## Side effects and authorization

`--agent` autoriza somente criar ou atualizar descoberta local para o agente informado. Não autoriza
instalação de agente, update do Spec Kit, rede, credenciais, alteração de source, Git remoto,
commit, push, merge, publicação ou deploy.
