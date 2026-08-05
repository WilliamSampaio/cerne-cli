# CLI Contract: `cerne workflow setup`

## Syntax

```text
cerne workflow setup
cerne workflow --help
```

O comando localiza o workspace ancestral mais próximo pelo manifesto e usa somente o provider já
declarado. Não aceita provider, path, `--force` ou argumento adicional.

## Configured success

Status: `0`; stderr vazio. stdout:

```text
Workflow: <provider>
Knowledge: <absolute-knowledge-path>
Setup concluído.
```

## Already configured

Status: `0`; stderr vazio. stdout:

```text
Workflow: <provider>
Knowledge: <absolute-knowledge-path>
Nenhuma alteração necessária.
```

Nenhum subprocesso nem novo registro de auditoria é criado.

## Operational failures

Status: `1`; stdout vazio. stderr segue:

```text
erro: <safe-cause>
correção: <action>
```

Inclui workspace ausente, manifesto inválido ou sem workflow, provider desconhecido, executável
ausente/incompatível, layout parcial, auditoria indisponível, processo falho ou marker inválido.
Quando houve subprocesso, a auditoria é preservada e somente owned roots comprovadamente novas
podem ser limpas.

## Invalid usage

Status: `2`; stdout vazio. stderr:

```text
erro: argumento inválido
uso: cerne workflow setup
```

## Side effects and authorization

O comando autoriza uma tentativa local de materializar a preferência existente dentro de knowledge
e registrar a execução. Não troca provider, não instala ferramentas, não escolhe agente, não altera
source, não usa credenciais e não executa comandos posteriores do workflow.
