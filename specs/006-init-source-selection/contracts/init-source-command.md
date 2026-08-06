# CLI Contract: `cerne init` Source Selection

## Syntax

```text
cerne init <project-name>
cerne init <project-name> --source <local-path>
cerne init <project-name> --clone <repository-location>
cerne init <project-name> --source <local-path> --workflow <speckit|openspec>
cerne init <project-name> --clone <repository-location> --workflow <speckit|openspec>
cerne init --help
```

As flags aparecem somente depois do nome e exigem um valor. `--source` e `--clone` não podem ser
repetidas nem combinadas entre si; `--workflow` pode acompanhar qualquer uma delas em qualquer ordem.

## Default mode

Sem flag, todo o contrato vigente de `cerne init <project-name>` permanece exato, incluindo
manifesto, `.gitkeep`, stdout, stderr, status, repositórios vazios e ausência de rede.

## Local source mode

`<local-path>` é resolvido a partir do diretório da invocação e precisa ser a raiz de um working
tree Git non-bare independente. Worktrees válidos são aceitos. O init não cria
`<project-name>/source` nem altera qualquer byte ou metadado do source informado.

Manifesto:

```json
{
  "name": "<project-name>",
  "source": "<portable-path-to-existing-source>"
}
```

## Clone mode

`<repository-location>` aceita:

- path local existente;
- `file://...`;
- `https://...` sem userinfo, query ou fragmento;
- `ssh://...` sem senha, query ou fragmento;
- SSH SCP-like, como `git@example.com:org/repo.git`.

HTTP sem TLS, `git://`, `ext::`, helpers desconhecidos, option-like input e credenciais embutidas
são recusados antes do processo. Path local existente tem precedência sobre interpretação de URL.

O clone é completo, não recursivo para submódulos, usa remoto `origin` e checkout padrão. O Cerne
não solicita depth, branch, mirror, bare, LFS, fetch extra ou push. Redirects, autenticação externa
e filtros configurados localmente podem participar do comportamento normal do Git.

Manifesto:

```json
{
  "name": "<project-name>",
  "source": "../source"
}
```

## Success

Status: `0`

Source local:

```text
Workspace "<project-name>" criado.
Knowledge: <absolute-knowledge-path>
Source vinculado: <absolute-existing-source-path>
```

Clone:

```text
Workspace "<project-name>" criado.
Knowledge: <absolute-knowledge-path>
Source clonado: <absolute-workspace-source-path>
```

stderr fica vazio. A localização original do clone não é exibida.

Quando `--workflow` é combinado, o sucesso acrescenta ao stdout:

```text
Workflow: <speckit|openspec>
Setup: <concluído|pendente>
```

Setup pendente mantém status `0` e usa stderr para o aviso e a correção, conforme o contrato de
workflow. Falha do source impede o provider; falha do provider preserva o source já validado.

## Invalid usage

Status: `2`; stdout vazio.

```text
erro: <cause>
uso: cerne init <project-name> [--source <caminho> | --clone <origem>] [--workflow <speckit|openspec>]
```

Nenhum arquivo é criado.

## Operational failure before clone

Status: `1`; stdout vazio. Validação local, Git ausente, destino inseguro, criação de knowledge,
manifesto ou auditoria inicial falham com rollback integral dos artefatos da invocação.

```text
erro: <safe-cause>
correção: <safe-action>
```

## Operational failure after clone starts

Status: `1`; stdout vazio. Knowledge, manifesto e `runs/source-clone.json` permanecem; somente o
staging privado parcial é removido quando a limpeza conclui. Um `source` que apareça antes da
promoção nunca é substituído ou removido. Normalmente o workspace fica incompleto e
`doctor`/`status` diagnosticam o source ausente.

```text
erro: não foi possível concluir o clone do source
correção: inspecione knowledge/runs/source-clone.json e associe um source válido ou remova o workspace incompleto antes de repetir o init
```

Nenhuma URL ou saída Git bruta aparece no diagnóstico.

Se o clone foi promovido mas a finalização da auditoria falhou, o source validado permanece, o
registro fica `started` e o comando ainda retorna status um; remover um source já público seria
mais destrutivo que preservar o resultado inconclusivo.

## Authorization and effects

- `--source` autoriza somente inspeção local read-only do path informado.
- `--clone` autoriza uma operação Git clone, seus transportes/redirects/autenticação/filtros
  normais e escrita somente no novo source.
- Nenhum modo autoriza push, publicação, merge, submódulo, alteração da origem ou conteúdo de
  knowledge dentro de source.
