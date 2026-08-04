# CLI Contract: `cerne status`

## Syntax

```text
cerne status
cerne status --help
```

Nenhum outro argumento ou flag é aceito nesta versão. O diretório atual é usado para localizar o
workspace ancestral mais próximo.

## Report

Uma consulta bem-sucedida escreve em stdout:

```text
Projeto: <project-name>
Workspace: <absolute-workspace-path>

Knowledge
  Caminho: <absolute-knowledge-path>
  Branch: <branch-name|detached HEAD>
  Commit: <short-hash|sem commits>
  Estado: <limpo|alterações pendentes>
  Modificados: <count>
  Em stage: <count>
  Não rastreados: <count>

Source
  Caminho: <absolute-source-path>
  Branch: <branch-name|detached HEAD>
  Commit: <short-hash|sem commits>
  Estado: <limpo|alterações pendentes>
  Modificados: <count>
  Em stage: <count>
  Não rastreados: <count>
```

Labels, ordem, indentação, streams e códigos de saída são contrato público. Os valores de caminho,
branch, commit e contagem refletem o estado local no momento da consulta.

## Clean repository

Um repositório é `limpo` somente quando:

- `Modificados` é `0`;
- `Em stage` é `0`;
- `Não rastreados` é `0`.

## Pending changes

Um repositório é `alterações pendentes` quando qualquer contagem é maior que zero. Isso não é erro
e não altera o código de saída quando todos os dados forem obtidos com sucesso.

## Detached HEAD and no commits

- Quando não houver branch simbólica atual, `Branch` deve ser `detached HEAD`.
- Quando não houver commit atual, `Commit` deve ser `sem commits`.

Esses estados são informativos e não são erro se a consulta Git do repositório concluir.

## Streams and status

| Resultado | stdout | stderr | Status |
|---|---|---|---:|
| Consulta obtida | relatório completo | vazio | 0 |
| Ajuda | ajuda | vazio | 0 |
| Uso inválido | vazio | causa e `uso: cerne status` | 2 |
| Workspace não localizado | vazio | causa, caminho de partida e correção | 1 |
| Manifesto inválido | vazio | causa, caminho do manifesto e correção | 1 |
| Caminho registrado inválido | vazio | causa, caminho afetado e correção | 1 |
| Repositório Git inválido | vazio | causa, caminho do repositório e correção | 1 |
| Consulta Git falhou | vazio | causa, caminho do repositório e correção | 1 |

## Read-only guarantee

O comando pode consultar arquivos existentes e metadados Git locais. Ele não cria, remove,
reescreve, corrige, altera stage, troca branch, cria commit, executa reset, acessa remotos ou usa
rede.

Mensagens não reproduzem conteúdo de arquivos, segredos, credenciais ou detalhes privados de
remotos.

## Help

`cerne status --help` documenta finalidade, sintaxe, localização do workspace, campos exibidos,
estados especiais, streams, códigos de saída, leitura exclusiva, limitações e exemplo.

