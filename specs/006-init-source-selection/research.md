# Research: Seleção de Source no Init

## Decisão 1: Manter três modos explícitos e fechados

**Decision**: Aceitar `cerne init <name>`, `<name> --source <path>` e `<name> --clone <origin>`.
`--source` e `--clone` são mutuamente exclusivos; `--workflow` pode acompanhar qualquer modo e,
quando combinado com uma opção de source, os pares podem aparecer em qualquer ordem depois do nome.
O modo sem flag mantém seu contrato byte a byte.

**Rationale**: O parser manual atual é suficiente e a posição fixa evita ambiguidade antes de
qualquer efeito. Os modos representam necessidades presentes, não um registry de estratégias.

**Alternatives considered**: Prompt interativo; flag genérica que tenta adivinhar path versus URL;
framework CLI; permitir opções antes do nome.

## Decisão 2: Reusar integralmente as regras de `cerne link`

**Decision**: `--source` resolve o input a partir do diretório de invocação e reutiliza
canonicalização, inspeção non-bare, exigência de raiz própria, common dir, sobreposição e
serialização portátil do manifesto já usadas por `link`. O source é validado antes dos efeitos e
revalidado antes do sucesso, sem lock.

**Rationale**: Essas regras já protegem sources externos e worktrees nas três plataformas. O
source pertence ao usuário e nunca deve ser bloqueado ou modificado pelo init.

**Alternatives considered**: Executar `init` seguido de `link --replace`, que cria source órfão;
duplicar validação; mover/copiar o repositório; adicionar lock ao source externo.

## Decisão 3: Restringir localizações de clone

**Decision**: Aceitar paths locais, `file`, HTTPS, SSH e SCP-like. Recusar HTTP sem TLS, protocolo
`git`, `ext`, helpers desconhecidos, URL com userinfo/credencial HTTP(S), senha, query ou fragmento.
Para inputs existentes no filesystem, path local tem precedência sobre interpretação de URL.

**Rationale**: `ext` e helpers podem executar programas. Credencial embutida seria persistida pelo
próprio Git em `remote.origin.url`. Os transportes escolhidos cobrem Git local e hosts modernos sem
parser ou SDK específico.

**Alternatives considered**: Repassar qualquer string ao Git; aceitar e redigir o remoto depois,
que alteraria o clone e ainda gravaria segredo; permitir apenas um host; adicionar biblioteca de URL.

## Decisão 4: Executar um clone completo, previsível e sem shell

**Decision**: O adapter usa argv separado e a forma equivalente a:

```text
git -c credential.interactive=false
    -c protocol.allow=never
    -c protocol.file.allow=always
    -c protocol.https.allow=always
    -c protocol.ssh.allow=always
    -c core.hooksPath=<empty-dir>
    clone --quiet --origin=origin --no-local --template=<empty-dir> -- <origin> <private-staging>
```

O ambiente remove redirecionamentos `GIT_*`, define `GIT_TERMINAL_PROMPT=0` e controles conhecidos
de interatividade, e não fornece stdin. O diretório vazio é temporário e controlado pelo Cerne.

**Rationale**: `--` impede option injection; `--no-local` evita hardlinks em clones por path;
protocol allowlist bloqueia helpers arbitrários; template e hooks path vazios impedem hooks locais
como `post-checkout`. Não usar depth, branch, filter ou submodule preserva o clone completo esperado.

**Alternatives considered**: Shell; ambiente integral sem overrides; ignorar toda configuração
global, quebrando proxy/CA/credential helper; clone raso; clonar diretamente em `source`, que
mistura resultado parcial com o path público e amplia o risco de limpeza indevida.

**Official references**:

- Git clone: <https://git-scm.com/docs/git-clone>
- Git environment: <https://git-scm.com/docs/git>
- Credential interactivity: <https://git-scm.com/docs/git-config#Documentation/git-config.txt-credentialinteractive>
- Protocol policy: <https://git-scm.com/docs/git-config#Documentation/git-config.txt-protocolallow>
- Hooks: <https://git-scm.com/docs/githooks#_post_checkout>
- Templates: <https://git-scm.com/docs/git-init#_template_directory>

## Decisão 5: Aceitar efeitos normais do checkout, sem prometer sandbox

**Decision**: A autorização cobre a cadeia normal do clone, incluindo redirects HTTPS,
autenticação externa e filtros de checkout configurados localmente. O Cerne não executa comandos
adicionais de submódulo, LFS, fetch ou push. Helpers externos que ignorem controles de
interatividade podem falhar ou apresentar UI própria; essa limitação é documentada.

**Rationale**: Checkout pode invocar filtros smudge, inclusive LFS, e o Git não oferece bloqueio
genérico de todos os filtros mantendo uma árvore pronta. `--no-checkout` entregaria um source
incompleto para o objetivo da feature.

**Alternatives considered**: `--no-checkout`; descobrir e neutralizar filtros arbitrários;
desabilitar configuração global e, com ela, autenticação, proxy e certificados esperados.

**Official reference**: Git attributes/filter: <https://git-scm.com/docs/gitattributes#_filter>

## Decisão 6: Mudar o rollback somente após auditoria durável

**Decision**: Antes de uma tentativa real de clone, falhas mantêm o rollback integral vigente.
Depois que `source-clone.json` em estado `started` é persistido, qualquer falha preserva knowledge,
manifesto e auditoria. O clone usa uma área temporária privada de acesso restrito no mesmo
filesystem. Falha de processo ou validação remove somente essa área. Após validação, a promoção
para `source` recusa substituição; se outro source apareceu, ele permanece intacto. Falha ao
finalizar a auditoria depois da promoção deixa `started`, preserva o source válido e impede
comunicar sucesso.

**Rationale**: A fronteira satisfaz rastreabilidade sem estado global. O nome imprevisível e o
acesso restrito estabelecem ownership da área removível; a promoção separa clone parcial do path
público e nunca autoriza apagar um source concorrente.

**Alternatives considered**: Rollback integral perdendo auditoria; storage global; manter source
parcial no path final; clone direto; snapshot de todo workspace; comando de retomada nesta versão.

## Decisão 7: Registrar a origem somente por impressão digital

**Decision**: A auditoria contém tipo de transporte e SHA-256 da localização exata, nunca o valor
bruto. Também contém operação, executor, projeto, destino relativo, autorização, status e
timestamps. A saída Git não é persistida nem reproduzida; erros usam categorias e correções do
Cerne.

**Rationale**: A impressão digital permite correlacionar uma origem fornecida posteriormente sem
expor host, path ou credencial. O remoto real já pertence ao clone bem-sucedido em `.git/config`.

**Alternatives considered**: URL integral; somente host; output completo; nenhuma correlação.

## Decisão 8: Não criar retomada nem opções de clone

**Decision**: Falha pós-clone deixa workspace incompleto diagnosticável pelos comandos atuais. O
usuário pode inspecionar a auditoria e remover o workspace para repetir ou associar um source
válido. Branch, depth, mirror, bare, submódulos, múltiplos remotos e retry ficam fora desta versão.

**Rationale**: Cada opção amplia contratos, estados e testes sem ser necessária para os dois fluxos
pedidos. `link` já oferece uma rota de recuperação local.

**Alternatives considered**: `cerne source clone`; retry automático; modo resume; repassar opções
arbitrárias do Git.
