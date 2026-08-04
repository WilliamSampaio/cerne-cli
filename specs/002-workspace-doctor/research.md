# Research: Diagnóstico de Workspace

## Decisão 1: Preservar parsing manual e biblioteca padrão

**Decision**: Estender `run` com despacho para `doctor`, sem framework CLI. Usar biblioteca padrão
para JSON, caminhos, arquivos e processos.

**Rationale**: Existem dois comandos e uma única opção de ajuda. O padrão atual já é testável e
suficiente.

**Alternatives considered**: Cobra, urfave/cli e uma árvore genérica de comandos; todos adicionam
estrutura e dependência sem requisito atual.

## Decisão 2: Identidade e versão preservam compatibilidade

**Decision**: Aceitar `version` ausente como versão 1 implícita e aprovada. Quando presente,
aceitar somente o inteiro JSON `1`; valor nulo, textual, fracionário ou diferente é bloqueante.
Campos desconhecidos permanecem tolerados. `name` inválido é bloqueante; `name` válido diferente
do nome da raiz produz aviso, sem invalidar o workspace.

**Rationale**: `cerne init` atualmente grava somente `name` e `source`. Considerar esse manifesto
saudável preserva compatibilidade e evita ampliar o escopo para alterar `init`.
Tratar a divergência de identidade válida como aviso mantém o diagnóstico útil sem rejeitar um
workspace que pode ter sido renomeado externamente.

**Alternatives considered**: Avisar em todo workspace atual, o que contradizia a jornada saudável;
adicionar `version` ao `init`, o que mudaria um contrato público sem necessidade; rejeitar campos
desconhecidos, o que impediria extensões compatíveis; tornar divergência de nome bloqueante ou
ignorá-la, opções mais rígida e menos informativa, respectivamente.

## Decisão 3: Caminhos canônicos e sem links

**Decision**: Avaliar o diretório atual como raiz. Inspecionar recursos com semântica que não segue
links, resolver `source` relativo a partir de `knowledge`, aceitar caminho absoluto quando
necessário e canonicalizar os caminhos existentes.

**Rationale**: `cerne link` permite source externo e pode usar caminho absoluto entre volumes.
Comparação por identidade do filesystem ainda detecta links, aliases e repositórios compartilhados.

**Alternatives considered**: `filepath.Clean` isolado, insuficiente contra links; busca ascendente
de workspace, fora do escopo; manter o limite antigo, incompatível com `cerne link`.

## Decisão 4: Identidade Git usa worktree e common dir

**Decision**: Para cada repositório esperado, consultar localmente a raiz da worktree e o diretório
Git comum. Comparar identidades reais com o caminho esperado e entre si; linked worktrees com
`common dir` compartilhado falham. Contenção canônica entre as raízes também é bloqueante.

**Rationale**: Procurar `.git` ou comparar apenas top-level aceita repositório ancestral e linked
worktrees que compartilham histórico. A documentação oficial define `--show-toplevel` e
`--git-common-dir` para esses fatos:
[git-rev-parse](https://git-scm.com/docs/git-rev-parse).

**Alternatives considered**: Ler internamente `.git`, formato que varia para worktrees; usar
`git status`, que pode tomar locks; consultar remotos, proibido; `ls-files`, desnecessário porque a
política já invalida qualquer contenção, rastreada ou não.

## Decisão 5: Processo Git estritamente local e saneado

**Decision**: Executar Git diretamente, sem shell, remover variáveis `GIT_*` herdadas que alterem
descoberta ou configuração e definir apenas `GIT_OPTIONAL_LOCKS=0` e
`GIT_TERMINAL_PROMPT=0`. Usar somente inspeções locais que não escrevem.

**Rationale**: Evita redirecionamento ambiental, prompts, rede e operações opcionais com lock. A
documentação oficial descreve que `GIT_OPTIONAL_LOCKS=0` desabilita operações opcionais que exigem
lock: [git](https://git-scm.com/docs/git).

**Alternatives considered**: Preservar todo ambiente, vulnerável a `GIT_DIR`, config injetada e
ceiling; shell, menos portátil; `status`, `fetch`, `remote` ou `submodule`, desnecessários ou
potencialmente modificadores/remotos.

## Decisão 6: Permissão efetiva sem arquivo-sonda

**Decision**: Confirmar leitura por operações reais não modificadoras. Para escrita, usar
`golang.org/x/sys`: `Faccessat` com identidade efetiva em Linux/macOS e abertura de recurso
existente com direitos mínimos e sem criação/truncamento no Windows. Acesso negado é erro; API ou
filesystem incapaz de concluir é aviso.

**Rationale**: Bits de modo não representam ACLs, grupos, privilégios ou DACLs do Windows. Criar um
arquivo seria a prova mais direta, mas violaria leitura exclusiva. Os pacotes oficiais expõem os
primitivos necessários:
[x/sys/unix](https://pkg.go.dev/golang.org/x/sys/unix) e
[x/sys/windows](https://pkg.go.dev/golang.org/x/sys/windows).

**Alternatives considered**: `FileMode.Perm`, sujeito a falsos resultados; `syscall.Access`, que
não cobre a identidade efetiva e Windows de modo uniforme; arquivo temporário, proibido; implementar
ACLs manualmente, mais código e risco que `x/sys`.

## Decisão 7: Dez resultados, sem short-circuit

**Decision**: Montar sempre dez slots na ordem contratada. Coletar dependências antes quando
necessário, mas renderizá-las em sua posição fixa. Falta de uma dependência torna o check dependente
um erro explícito.

**Rationale**: O usuário precisa ver cada verificação, e scripts precisam de ordem previsível.
Omitir checks produziria relatórios incomparáveis e poderia parecer aprovação.

**Alternatives considered**: Parar no primeiro erro, que oculta defeitos; executar estritamente na
ordem visual, que repetiria trabalho; gerar checks dinamicamente, que enfraquece o contrato.

## Decisão 8: Separar fatos, classificação e apresentação

**Decision**: `workspace` classifica fatos em `CheckResult`; adaptadores retornam somente fatos ou
falhas; `main` renderiza símbolos, linhas, resumo e status.

**Rationale**: Mantém detalhes Git e de plataforma fora do domínio e permite testar a precedência
erro > aviso > aprovação sem subprocessos.

**Alternatives considered**: Interface para cada check, abstração especulativa; imprimir dentro dos
adaptadores, que mistura streams com integração; um registry extensível, não exigido.

## Decisão 9: Provar leitura exclusiva por snapshot lógico

**Decision**: Testes de integração capturam árvore, tipos, conteúdo, tamanhos, modos, mtimes e estado
Git relevante antes e depois. `atime` não faz parte da comparação porque a própria leitura pode ser
registrada pelo filesystem.

**Rationale**: Demonstra ausência de criação, escrita e configuração sem exigir comportamento de
atime idêntico entre plataformas.

**Alternatives considered**: Apenas revisar os comandos usados, prova insuficiente; comparar todos
os metadados, instável por atime; usar um remoto sentinela, desnecessário e proibido.

## Riscos aceitos

- A consulta de permissão é um retrato do processo atual e não autoriza uma escrita futura.
- Filesystems remotos ou APIs sem suporte confiável podem produzir aviso, nunca aprovação falsa.
- Mudança concorrente entre validações produz erro, não tentativa de correção.
- Mensagens do Git são sanitizadas e limitadas; conteúdo do manifesto não é reproduzido.
