# CLI Contract: `cerne link`

## Syntax

```text
cerne link <caminho-do-repositorio>
cerne link <caminho-do-repositorio> --replace
cerne link --help
```

Nenhum outro argumento ou flag é aceito nesta versão. O diretório atual é usado para localizar o
workspace Cerne e para resolver caminhos relativos informados pelo usuário.

## Behavior

`cerne link` vincula o workspace atual a um repositório Git local existente como `source`.

O comando:

- localiza o workspace ancestral mais próximo;
- lê o manifesto em `knowledge/cerne.json`;
- valida o caminho informado;
- valida o repositório Git candidato;
- valida a independência entre knowledge e source;
- recusa substituição sem `--replace`;
- atualiza somente o manifesto, de forma atômica.

## Successful update

Uma atualização bem-sucedida escreve em stdout:

```text
Projeto: <project-name>
Source anterior: <previous-source>
Novo source: <new-source>
Manifesto atualizado.
```

`Source anterior` deve aparecer quando o manifesto possuir uma referência source anterior.

## No-op

Quando o source informado já é o source configurado, o comando escreve em stdout:

```text
Projeto: <project-name>
Source atual: <current-source>
Nenhuma alteração necessária.
```

Esse resultado usa status `0` e não regrava o manifesto.

## Replacement

Se o manifesto aponta para outro source, `cerne link <caminho>` falha por padrão. O usuário deve
executar:

```text
cerne link <caminho> --replace
```

Mesmo com `--replace`, apenas a referência no manifesto pode mudar. O source anterior não é apagado,
movido, limpo, commitado, acessado remotamente ou modificado.

## Path storage

O manifesto armazena o novo source como caminho normalizado. Quando for possível representar o
source de forma portátil em relação ao repositório knowledge, o valor armazenado deve ser relativo.
Quando isso não for possível, o valor pode ser absoluto e normalizado.

## Streams and status

| Resultado | stdout | stderr | Status |
|---|---|---|---:|
| Manifesto atualizado | resumo completo | vazio | 0 |
| Source já configurado | resumo sem alteração | vazio | 0 |
| Ajuda | ajuda | vazio | 0 |
| Uso inválido | vazio | causa e `uso: cerne link <caminho> [--replace]` | 2 |
| Workspace não localizado | vazio | causa, caminho de partida e correção | 1 |
| Manifesto ausente/inválido/incompatível | vazio | causa, caminho do manifesto e correção | 1 |
| Caminho informado inexistente | vazio | causa, caminho informado e correção | 1 |
| Caminho informado não diretório | vazio | causa, caminho informado e correção | 1 |
| Repositório Git inválido | vazio | causa, caminho informado e correção | 1 |
| Repositório bare | vazio | causa, caminho informado e correção | 1 |
| Source igual a knowledge | vazio | causa, caminhos envolvidos e correção | 1 |
| Sobreposição perigosa | vazio | causa, caminhos envolvidos e correção | 1 |
| Substituição sem `--replace` | vazio | causa, source atual, novo source e correção | 1 |
| Manifesto não pode ser atualizado | vazio | causa, caminho do manifesto e correção | 1 |

## Read/write guarantee

O comando pode ler arquivos do workspace, o manifesto e metadados Git locais. Ele pode gravar
somente o manifesto, e somente após todas as validações passarem.

O comando não pode:

- copiar, mover, renomear ou apagar arquivos do source antigo ou novo;
- criar links simbólicos;
- executar checkout, reset, add, commit, clean, fetch, pull ou push;
- acessar GitHub, GitLab ou qualquer remoto;
- alterar remotes, branches, stage, commits ou configuração do source;
- realizar commit automático no knowledge;
- chamar agentes de IA.

## Help

`cerne link --help` documenta finalidade, sintaxe, argumento, `--replace`, resolução de caminhos,
validações, efeitos colaterais, streams, códigos de saída, limitações e exemplos.
