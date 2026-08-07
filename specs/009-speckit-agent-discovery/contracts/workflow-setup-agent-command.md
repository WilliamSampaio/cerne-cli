# CLI Contract: `cerne workflow setup --agent`

## Syntax

```text
cerne workflow setup
cerne workflow setup --agent <codex|claude>
cerne workflow --help
```

Sem `--agent`, o contrato atual permanece. Com `--agent`, o comando localiza o workspace ancestral,
exige workflow declarado como Spec Kit e prepara a descoberta local do agente escolhido.

## Configured or refreshed success

Status: `0`; stderr vazio.

stdout:

```text
Workflow: speckit
Knowledge: <absolute-knowledge-path>
Setup concluído.
Agent: <codex|claude>
Descoberta: pronta
```

Se o workflow já estava pronto e somente a ponte foi criada ou atualizada, `Setup concluído.` pode
ser substituído pelo texto existente:

```text
Nenhuma alteração necessária.
```

As linhas `Agent` e `Descoberta` indicam o resultado local da ponte, não um campo persistido.

## Missing provider

Status: `1`; stdout vazio.

stderr:

```text
erro: executável "specify" não encontrado
correção: instale speckit e execute novamente
```

Nenhuma ponte local é criada quando o provider necessário para materializar ou validar integração
de agente está indisponível.

## Operational failures

Status: `1`; stdout vazio.

stderr:

```text
erro: <safe-cause>
correção: <action>
```

Inclui workspace ausente, manifesto inválido ou sem workflow, workflow diferente de Spec Kit,
provider desconhecido, layout parcial, auditoria indisponível, processo falho, marker inválido,
falha ao preparar integração de agente em `knowledge` ou falha ao criar ponte local.

## Invalid usage

Status: `2`; stdout vazio.

stderr:

```text
erro: argumento inválido
uso: cerne workflow setup [--agent <codex|claude>]
```

Inclui agente ausente, desconhecido, repetido ou argumentos extras.

## Side effects and authorization

`workflow setup --agent <agent>` autoriza uma tentativa local de materializar o workflow declarado
em `knowledge`, instalar a integração Spec Kit correspondente dentro de `knowledge` quando
necessário e criar ou atualizar a ponte local na raiz do workspace. Não altera source, não troca o
provider, não persiste agente no manifesto, não instala agentes externos, não usa credenciais e não
executa comandos posteriores do workflow.
