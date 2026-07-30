# Research: Inicialização de Workspace

## Decisão 1: Biblioteca padrão e parsing manual

**Decision**: Manter zero dependências Go. `main` delega para uma função `run` que recebe
argumentos, stdout e stderr; o despacho reconhece ajuda e o subcomando `init`.

**Rationale**: Existe um comando sem flags. Parsing manual é menor, testável e suficiente.

**Alternatives considered**: Cobra, urfave/cli e `flag.FlagSet`. Frameworks adicionam dependência e
estrutura sem requisito; `flag.FlagSet` passa a valer quando surgir a primeira flag real.

## Decisão 2: Contrato portátil para o nome

**Decision**: Aceitar de 1 a 255 caracteres ASCII. O primeiro é alfanumérico; os demais podem ser
alfanuméricos, `.`, `_` ou `-`. Rejeitar ponto final e, sem diferenciar maiúsculas, `CON`, `PRN`,
`AUX`, `NUL`, `COM1`–`COM9` e `LPT1`–`LPT9`, inclusive antes de uma extensão.

**Rationale**: Uma regra única evita que um workspace válido em um sistema falhe em outro. A
validação não depende das regras do host atual.

**Alternatives considered**: Somente `filepath.IsLocal`, insuficiente entre sistemas; Unicode
irrestrito, que exige regras adicionais de normalização e equivalência sem requisito atual.

## Decisão 3: Inspeção segura do destino

**Decision**: Resolver `cwd + nome` com `filepath.Join`, usar `os.Lstat` para não seguir links e
aceitar apenas destino ausente ou diretório regular sem entradas. Verificar Git antes de criar
qualquer caminho.

**Rationale**: Atende a proteção de dados e torna falhas de nome, destino e dependência totalmente
anteriores à mutação.

**Alternatives considered**: `os.Stat`, que segue links; leitura e ordenação completas do diretório,
desnecessárias para saber se existe ao menos uma entrada.

## Decisão 4: Criação com rollback de propriedade explícita

**Decision**: Registrar se a raiz foi criada e cada filho criado pela execução. Em falha, remover
esses caminhos em ordem inversa; remover a raiz somente quando ela também foi criada. Nunca usar
remoção recursiva sobre uma raiz preexistente.

**Rationale**: É o menor mecanismo que preserva um destino vazio preexistente e desfaz falhas
controladas durante manifesto ou Git.

**Alternatives considered**: Staging e publicação por renames. Para um destino vazio preexistente,
isso exige dois renames e compensação específica entre sistemas, sem eliminar todos os riscos.

## Decisão 5: Git como adaptador-função

**Decision**: `internal/gitexec` resolve o executável Git e expõe inicialização local. O caso de uso
recebe apenas `func(path string) error`. Cada repositório executa `git init --quiet` com o próprio
diretório como `Cmd.Dir`, sem shell.

**Rationale**: Satisfaz o limite de adaptador sem criar uma interface de implementação única. A
função também permite simular falha na segunda inicialização para provar rollback.

**Alternatives considered**: Biblioteca go-git, dependência desnecessária; `git worktree`, que
compartilha histórico; executar via shell, menos portátil e sujeito a interpretação de argumentos.

## Decisão 6: Isolamento do processo Git

**Decision**: Remover do ambiente do subprocesso variáveis `GIT_*` que redirecionam diretório,
worktree, índice, objetos ou namespace. Não forçar branch inicial, commit, remoto ou template.

**Rationale**: Impede configuração ambiental de unir os repositórios. O nome do branch ainda não é
contrato e pode seguir a configuração local do Git.

**Alternatives considered**: Forçar `main`, requisito inexistente; criar commit vazio ou
`.gitkeep`, ambos incompatíveis com `source/` vazio.

## Decisão 7: Manifesto mínimo

**Decision**: Gravar `knowledge/cerne.json` como JSON UTF-8 indentado, terminado por newline e
criado de forma exclusiva. O objeto contém somente `name` e `source`, com `source` igual ao literal
portátil `../source`.

**Rationale**: `encoding/json` já existe na biblioteca padrão e o contrato não pede versão,
timestamp, identificador ou campos extensíveis.

**Alternatives considered**: YAML, que exigiria dependência ou parser; campos futuros
especulativos; escrita que permita substituição.

## Decisão 8: Streams e códigos de saída

**Decision**: Sucesso e ajuda usam stdout e status `0`; falha operacional usa stderr e status `1`;
uso inválido usa stderr, inclui a sintaxe e retorna `2`. Não há prompt, cor ou logging.

**Rationale**: Separa resultado consumível de diagnóstico e permite automação previsível.

**Alternatives considered**: Um único status não zero, que perde a distinção entre uso e ambiente;
mensagens em stdout, que contaminam consumo por scripts.

## Decisão 9: Estratégia mínima de testes

**Decision**: Testes unitários cobrem validação e rollback; o adaptador usa Git real em
`t.TempDir`; o teste do CLI compila o binário e cobre stdout, stderr, status e efeitos. CI executa
`go test ./...` em Linux, Windows e macOS.

**Rationale**: Prova domínio, limite externo e contrato público sem mocks extensos ou remotos.

**Alternatives considered**: Somente mocks, que não provam independência Git; scripts shell, que
não são portáveis; remotos reais, proibidos pela constituição.

## Riscos aceitos

- A execução pressupõe uso exclusivo do destino durante `init`; concorrência externa não faz parte
  desta feature.
- Falha do próprio sistema operacional durante rollback pode exigir correção manual; o erro deve
  preservar conteúdo desconhecido e identificar os caminhos afetados.
- Um repositório ancestral pode observar os repositórios internos, mas o Cerne não o altera e os
  dois repositórios administrados continuam independentes entre si.
- Configuração local do Git pode mudar detalhes internos de `.git`; testes verificam contratos Git,
  não o layout interno.
