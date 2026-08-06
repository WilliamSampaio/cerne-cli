# CLI Contract: `cerne restore`

## Syntax

```text
cerne restore <knowledge-origin> --source <local-path>
cerne restore <knowledge-origin> --clone <source-origin>
cerne restore --help
```

`<knowledge-origin>` é posicional. A opção de source aparece depois dela, exatamente uma vez, e
exige valor. Não há forma sem source, combinação das duas flags, reordenação, repetição ou
argumento extra nesta versão.

## Inputs

Knowledge sempre é clonado. Knowledge e source clonado aceitam a mesma política de
`init --clone`: path Git local existente, `file`, HTTPS, SSH ou SCP-like. HTTP, `git://`, `ext`,
helper desconhecido, option-like, credencial embutida, query e fragmento são recusados antes do
processo. O clone é completo, não recursivo para submódulos, usa checkout padrão e remoto `origin`.

`--source` resolve path relativo a partir do diretório da invocação e exige a raiz de um working
tree Git non-bare independente. O repositório é somente inspecionado.

O workspace final é criado no diretório da invocação com o `name` de `cerne.json`. O destino deve
estar ausente; até diretório vazio preexistente é preservado e recusado.

## Success

Status `0`. A auditoria está finalizada, stderr fica vazio e nenhuma origem aparece.

Source clonado:

```text
Workspace "<name>" restaurado.
Knowledge: <absolute-knowledge-path>
Source clonado: <absolute-source-path>
```

Source local sem mudança da referência:

```text
Workspace "<name>" restaurado.
Knowledge: <absolute-knowledge-path>
Source vinculado: <absolute-local-source-path>
```

Se somente `manifest.source` foi atualizado, acrescentar:

```text
Manifesto: referência de source atualizada.
```

O path do source final é saída exigida e não é origem de clone. O path global da auditoria não faz
parte do stdout estável.

## Invalid usage or origin

Status `2`, stdout vazio, nenhum audit, processo ou artefato.

```text
erro: <causa segura>
uso: cerne restore <origem-knowledge> (--source <caminho> | --clone <origem-source>)
```

Causas incluem `argumento inválido`, `origem do knowledge inválida` e `origem de clone do source
inválida`. O valor recebido nunca é reproduzido.

## Operational failure

Status `1`, stdout vazio, workspace final ausente e único audit global preservado:

```text
erro: <causa segura>
correção: <ação segura>
```

Causas públicas cobrem:

- auditoria privada inacessível ou insegura;
- Git indisponível;
- falha ao restaurar knowledge ou source;
- manifesto, versão, workflow, nome ou layout inválido;
- source local inválido, sobreposto ou alterado durante a tentativa;
- source clonado com path absoluto, externo ou não portátil;
- destino já existente ou criado concorrentemente;
- validação, promoção, rollback ou finalização da auditoria incompleta.

Erros do Git viram categorias e correções do Cerne; argv, origem e output bruto nunca são exibidos.
Falha de finalização da auditoria também retorna `1`, remove o workspace promovido quando ownership
é confirmado e deixa o registro no último estado durável inconclusivo. Se limpeza segura não puder
ser confirmada, o diagnóstico não afirma que ela terminou e não remove alvo ambíguo.

## Workflow

Uma declaração válida é preservada. Layout pendente e executável ausente não bloqueiam a
restauração; layout parcial ou provider desconhecido bloqueiam. `restore` nunca executa setup. O
usuário pode chamar `cerne workflow setup` separadamente depois do sucesso.

## Authorization and effects

- A origem posicional autoriza um clone Git de knowledge para staging privado.
- `--clone` autoriza um segundo clone para o target declarado no manifesto.
- `--source` autoriza inspeção Git read-only e alteração exclusiva de `manifest.source` no clone de
  knowledge.
- Autenticação, redirects e filtros normais do Git podem participar do clone, conforme o contrato
  já documentado de `init --clone`.
- Nenhuma autorização cobre workflow, agente, push, merge, fetch adicional, submódulo, instalação,
  publicação ou deploy.

Após falha, uma nova invocação é a retomada suportada; não existe `--resume`.
