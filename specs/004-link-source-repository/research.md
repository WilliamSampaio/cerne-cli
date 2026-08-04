# Research: Link de Repositório Source

## Decisão 1: Preservar CLI manual e biblioteca padrão

**Decision**: Adicionar `link` ao despacho manual de `cmd/cerne`, sem framework CLI e sem
dependências novas.

**Rationale**: O projeto já possui comandos simples com contratos de stdout, stderr e status
testados. A nova sintaxe tem apenas um argumento obrigatório e uma flag booleana.

**Alternatives considered**: Framework CLI, registry genérico de comandos ou parser externo; todos
adicionam dependência ou abstração sem requisito atual.

## Decisão 2: Reusar a localização de workspace por ancestral

**Decision**: Localizar o workspace a partir do diretório atual usando a mesma semântica de
`cerne status`: ancestral mais próximo com `knowledge/cerne.json`.

**Rationale**: `link` deve funcionar na raiz, em `knowledge/`, em `source/` e em subdiretórios.
Manter a mesma regra evita contratos diferentes entre comandos.

**Alternatives considered**: Exigir execução na raiz, que contraria a especificação; aceitar
argumento de workspace, fora do escopo; varrer descendentes, ambíguo e mais caro.

## Decisão 3: Resolver caminho informado a partir do diretório de execução

**Decision**: Caminhos relativos passados em `cerne link <caminho>` serão resolvidos em relação ao
diretório atual da invocação, não em relação ao workspace ou ao repositório knowledge.

**Rationale**: Este é o comportamento esperado de CLIs para argumentos de caminho e permite o
exemplo `cerne link ../geo-app` a partir de qualquer diretório onde o usuário executa o comando.

**Alternatives considered**: Resolver sempre a partir do workspace ou de `knowledge`; isso torna o
resultado surpreendente quando o comando é executado em subdiretório.

## Decisão 4: Armazenar source relativo ao knowledge quando portátil

**Decision**: O manifesto continuará armazenando `source` relativo a `knowledge` quando `filepath.Rel`
conseguir produzir caminho relativo seguro e portátil. Quando isso não for possível, usar caminho
absoluto normalizado.

**Rationale**: O manifesto atual já usa `../source`, relativo ao knowledge. Caminho relativo torna
workspaces movíveis no mesmo conjunto de diretórios, enquanto caminho absoluto cobre volumes
distintos no Windows e casos externos inevitáveis.

**Alternatives considered**: Sempre absoluto, menos portável; sempre relativo, falha em volumes
distintos; relativo ao workspace, divergiria do manifesto atual.

## Decisão 5: Validar Git por consultas locais read-only

**Decision**: Validar o source candidato com consultas Git locais sem shell e sem remoto:
identificar worktree root, git common dir, bare/non-bare e presença de árvore de trabalho.

**Rationale**: A especificação exige aceitar worktrees e recusar bare repositories sem modificar
repositórios. Consultas locais são suficientes e preservam o requisito de não acessar remotos.

**Alternatives considered**: Parse direto de `.git`, frágil para worktrees e links; `git status`,
mais amplo que o necessário; `fetch`/`pull`, explicitamente fora do escopo.

## Decisão 6: Comparar repositórios por identidade Git e limites de caminho

**Decision**: Recusar source igual ao knowledge quando as raízes de worktree ou metadados comuns
identificarem o mesmo repositório. Recusar aninhamento quando um diretório ou worktree contém o
outro.

**Rationale**: A separação entre conhecimento e código depende tanto de caminhos quanto de
identidade Git. Worktrees podem ter caminhos diferentes e ainda compartilhar metadados comuns.

**Alternatives considered**: Comparar só strings de caminho, insuficiente com aliases/symlinks;
comparar só common dir, insuficiente para alguns layouts; permitir aninhamento com `.gitignore`,
arriscado e contrário à constituição.

## Decisão 7: Substituição exige `--replace`, mesmo se o source atual estiver quebrado

**Decision**: Se o manifesto possui qualquer source configurado diferente do candidato, `link`
recusa a troca sem `--replace`, inclusive quando o source atual não existe ou está inválido.

**Rationale**: A existência da referência anterior é uma intenção persistida. Trocar essa intenção
sem autorização explícita seria uma alteração sensível do workspace.

**Alternatives considered**: Permitir troca automática quando o source atual está inválido; isso
facilita recuperação, mas enfraquece a proteção contra substituição acidental.

## Decisão 8: Atualização atômica restrita ao manifesto

**Decision**: Escrever o novo manifesto em arquivo temporário no mesmo diretório, sincronizar quando
possível e renomear sobre o manifesto somente após todas as validações concluírem.

**Rationale**: Renomeio no mesmo diretório é o caminho portátil mais simples para preservar o
manifesto anterior em caso de falha antes da troca final.

**Alternatives considered**: Reescrever o arquivo diretamente, pode corromper em falha; criar backup
persistente, adiciona artefato sem requisito; commitar automaticamente, fora do escopo.

## Decisão 9: Saída textual estável

**Decision**: Sucesso, sem alteração e ajuda usam stdout. Falhas e uso inválido usam stderr. Status
`0`, `1` e `2` seguem o padrão dos comandos existentes.

**Rationale**: O Cerne já usa contratos simples e previsíveis para automação por scripts.

**Alternatives considered**: JSON, cores ou prompts interativos; todos estão fora do escopo desta
primeira versão.
