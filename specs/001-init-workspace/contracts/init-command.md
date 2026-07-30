# CLI Contract: `cerne init`

## Syntax

```text
cerne init <project-name>
cerne init --help
```

`<project-name>` é obrigatório e não aceita flags nesta versão.

## Project name

- Comprimento: 1 a 255 caracteres ASCII.
- Primeiro caractere: `A-Z`, `a-z` ou `0-9`.
- Demais caracteres: alfanuméricos, `.`, `_` ou `-`.
- O nome não pode terminar em `.`.
- O segmento anterior ao primeiro `.` não pode ser, sem diferenciar maiúsculas:
  `CON`, `PRN`, `AUX`, `NUL`, `COM1`–`COM9` ou `LPT1`–`LPT9`.

## Destination

O destino é o caminho absoluto resultante de `<diretório-atual>/<project-name>`. Ele precisa estar
ausente ou ser um diretório regular vazio. Arquivos, links e diretórios com qualquer entrada são
recusados sem modificação.

## Success

Status: `0`

stdout:

```text
Workspace "<project-name>" criado.
Knowledge: <absolute-knowledge-path>
Source: <absolute-source-path>
```

stderr fica vazio.

## Help

`cerne init --help` retorna status `0`, escreve em stdout e documenta:

- finalidade e sintaxe;
- regra do nome;
- árvore criada;
- dois repositórios Git locais, sem commit ou remoto;
- stdout, stderr e status `0`, `1` e `2`;
- destinos recusados e ações corretivas;
- ao menos um exemplo.

## Errors

### Operational failure

Status: `1`

stderr:

```text
erro: <cause>
correção: <action>
```

Aplica-se a destino inseguro, permissão, Git indisponível, criação, manifesto, inicialização Git ou
rollback. stdout fica vazio.

### Invalid usage

Status: `2`

stderr:

```text
erro: <cause>
uso: cerne init <project-name>
```

Aplica-se a argumento ausente, excedente ou nome inválido. stdout fica vazio.

## Workspace layout

```text
<project-name>/
├── knowledge/
│   ├── .git/
│   ├── cerne.json
│   ├── product/
│   ├── specs/
│   ├── decisions/
│   ├── policies/
│   └── runs/
└── source/
    └── .git/
```

A raiz `<project-name>/` não recebe `.git`.

## Manifest

`knowledge/cerne.json` é JSON UTF-8 válido, terminado por newline:

```json
{
  "name": "<project-name>",
  "source": "../source"
}
```

## Git guarantees

- `knowledge` e `source` possuem raízes Git distintas.
- Nenhum remoto é configurado.
- Nenhum commit é criado.
- O nome do branch inicial não faz parte deste contrato.
- Variáveis ambientais que redirecionam internamente o Git não podem unir os repositórios.

## Side effects and authorization

A invocação autoriza somente a criação local descrita. O comando não acessa rede, credenciais,
agentes, repositórios remotos, push, merge, publicação ou deploy. Em falha, somente artefatos
criados pela invocação podem ser removidos.
