# Research: Status do Workspace

## Decisão 1: Preservar CLI manual e biblioteca padrão

**Decision**: Adicionar `status` ao despacho manual de `cmd/cerne`, sem framework CLI e sem
dependências novas.

**Rationale**: O projeto já possui três comandos simples e contratos testáveis com biblioteca
padrão. A nova funcionalidade não exige roteamento complexo.

**Alternatives considered**: Cobra, urfave/cli e registries genéricos de comandos; todos adicionam
dependência ou abstração sem requisito atual.

## Decisão 2: Localizar workspace por ancestral mais próximo

**Decision**: A partir do diretório atual, procurar subindo ancestrais até encontrar
`knowledge/cerne.json`; usar o primeiro encontrado como raiz do workspace.

**Rationale**: O usuário pediu localização a partir do diretório atual. O ancestral mais próximo
permite executar o comando na raiz, em `knowledge`, em `source` ou em subdiretórios, sem argumento
extra.

**Alternatives considered**: Exigir execução na raiz, simples demais para o requisito; aceitar
argumento de caminho, fora do escopo inicial; varrer descendentes, mais caro e ambíguo.

## Decisão 3: Reutilizar regras de manifesto e path do domínio

**Decision**: Reusar a semântica existente de `knowledge/cerne.json`, `name`, `source` relativo a
`knowledge`, paths canônicos dentro do workspace e rejeição de caminhos inválidos.

**Rationale**: `status` e `doctor` interpretam o mesmo workspace. Duplicar uma semântica diferente
criaria inconsistência pública.

**Alternatives considered**: Parser próprio para status, que poderia divergir; aceitar `source`
externo, incompatível com o modelo atual; seguir links, arriscado para limites do workspace.

## Decisão 4: Consultas Git locais, separadas e somente-leitura

**Decision**: Coletar dados por comandos locais sem shell: `rev-parse --show-toplevel`,
`rev-parse --git-common-dir`, `symbolic-ref --quiet --short HEAD`,
`rev-parse --verify --short=7 HEAD`, `diff --name-only`, `diff --cached --name-only` e
`ls-files --others --exclude-standard`.

**Rationale**: Esses comandos consultam fatos locais suficientes para branch, detached HEAD,
commit, ausência de commits e contagens separadas sem depender de remotos. O padrão existente de
ambiente saneado, `GIT_OPTIONAL_LOCKS=0` e `GIT_TERMINAL_PROMPT=0` continua válido.

**Alternatives considered**: `git status --porcelain`, compacto, mas pode fazer mais trabalho que
o necessário; parsing de `.git`, frágil com worktrees; `fetch`, `pull` ou comparação com remoto,
explicitamente fora do escopo.

## Decisão 5: Detached HEAD e repositório sem commits não são erros

**Decision**: Falha de `symbolic-ref --quiet --short HEAD` com repositório válido classifica a
branch como `detached HEAD`. Falha de `rev-parse --verify --short=7 HEAD` por HEAD inexistente
classifica commit como `sem commits`.

**Rationale**: A especificação exige comunicar esses estados especiais sem invalidar a consulta
quando o Git local foi acessado com sucesso.

**Alternatives considered**: Tratar como erro, contradiz o requisito; ocultar branch ou commit,
menos claro para humanos e scripts.

## Decisão 6: Contagens por arquivo visível ao usuário

**Decision**: Contar nomes distintos retornados para alterações fora do stage, alterações em stage
e arquivos não rastreados. O mesmo arquivo pode contar em mais de uma categoria quando possuir
mudanças staged e unstaged.

**Rationale**: O usuário quer as categorias separadas, não uma contagem deduplicada global. Isso
explica melhor o próximo passo manual.

**Alternatives considered**: Contagem total única, perde informação; contar hunks ou linhas,
detalhe demais e fora do escopo; exibir nomes de arquivos, pode expor conteúdo privado.

## Decisão 7: Saída textual estável com contagens sempre presentes

**Decision**: Renderizar os mesmos campos para `knowledge` e `source`, incluindo contagens zero.
Sucesso e ajuda usam stdout; falhas e uso inválido usam stderr; status `0`, `1` e `2` seguem o
padrão dos comandos existentes.

**Rationale**: Saída estável é mais previsível para scripts e documentação. O projeto já usa
status `2` para uso inválido.

**Alternatives considered**: Omitir contagens zero, mais compacto mas menos estável; JSON, fora do
escopo; cores ou ícones, decoração sem requisito.

## Decisão 8: Provar leitura exclusiva por snapshot lógico

**Decision**: Testes de integração comparam árvore, tipos, conteúdo, tamanho, mtimes de arquivos e
estado Git relevante antes e depois. `atime` e mtimes de diretórios ficam fora da comparação.

**Rationale**: Leituras podem atualizar `atime`, e alguns sistemas ajustam metadados de diretório
sem alteração de conteúdo. A garantia relevante é não criar, remover, alterar arquivos, stage,
branch, commits ou configuração remota.

**Alternatives considered**: Comparar todos os metadados, instável em Windows/macOS; confiar apenas
em revisão de comandos, prova fraca; criar arquivo-sonda, proibido.

