# CLI Contract: `cerne doctor`

## Syntax

```text
cerne doctor
cerne doctor --help
```

Nenhum outro argumento ou flag é aceito nesta versão. O diretório atual é a raiz analisada.

## Report

Um diagnóstico iniciado com sucesso escreve em stdout exatamente dez linhas, nesta ordem:

1. `Manifesto`
2. `Repositório de conhecimento`
3. `Repositório de código-fonte`
4. `Independência Git`
5. `Isolamento de versionamento`
6. `Caminhos do manifesto`
7. `Diretórios obrigatórios`
8. `Git`
9. `Permissões`
10. `Versão do manifesto`

Formato de cada linha:

```text
<symbol> <label>: <detail>[; correção: <action>]
```

Símbolos:

- `✓`: aprovação;
- `✗`: erro bloqueante;
- `!`: aviso não bloqueante.

Aviso e erro sempre incluem `correção`. Não há cor, prompt, progresso ou logging.

Após as dez linhas, stdout contém exatamente um resumo:

```text
Workspace saudável
```

ou:

```text
Workspace com avisos
```

ou:

```text
Workspace inválido
```

Símbolos, labels, ordem, formato, resumos, streams e status são estáveis para automação. O detalhe
e a ação corretiva são destinados a humanos e podem ser esclarecidos sem mudar sua semântica.

## Healthy example

Um workspace criado pelo `cerne init` atual usa versão 1 implícita:

```text
✓ Manifesto: legível
✓ Repositório de conhecimento: encontrado
✓ Repositório de código-fonte: encontrado
✓ Independência Git: raízes e históricos distintos
✓ Isolamento de versionamento: nenhum repositório contém o outro
✓ Caminhos do manifesto: válidos
✓ Diretórios obrigatórios: product, specs, decisions, policies e runs encontrados
✓ Git: disponível
✓ Permissões: leitura e escrita confirmadas
✓ Versão do manifesto: versão 1 implícita e suportada
Workspace saudável
```

## Warning example

```text
! Manifesto: name válido difere do nome da raiz; correção: alinhe o manifesto ou renomeie o workspace
Workspace com avisos
```

Os outros nove resultados continuam presentes. Incerteza ao confirmar escrita também pode gerar
aviso em `Permissões`. Aviso nunca é apresentado como aprovação.

## Manifest identity and version

`name` inválido é erro. `name` válido diferente do basename da raiz gera aviso em `Manifesto` e,
na ausência de erros, status `0`. A ausência de `version` significa versão 1 implícita; quando o
campo existe, somente o inteiro JSON `1` é aceito. Valores como `"1"`, `1.0`, `null` ou outra versão
são erros bloqueantes.

## Errors and dependencies

Erros de manifesto, estrutura, Git, permissão, versão e isolamento usam `✗`. Quando uma dependência
impede uma verificação, a linha dependente também usa `✗` e identifica a dependência; ela não é
omitida. Falhas simultâneas permanecem visíveis e qualquer uma torna o resumo inválido.

Mensagens não reproduzem o conteúdo do manifesto, configuração Git, credenciais ou outros dados
privados. Saída de subprocesso é sanitizada e limitada.

## Streams and status

| Resultado | stdout | stderr | Status |
|---|---|---|---:|
| Saudável | relatório completo | vazio | 0 |
| Somente avisos | relatório completo | vazio | 0 |
| Erro bloqueante | relatório completo | vazio | 1 |
| Ajuda | ajuda | vazio | 0 |
| Uso inválido | vazio | causa e `uso: cerne doctor` | 2 |
| Falha antes de iniciar relatório | vazio | causa e correção | 1 |

Git indisponível é um resultado diagnosticável: produz relatório completo, check `Git` com erro,
checks Git dependentes com erro e status 1; não é falha anterior ao relatório.

## Read-only guarantee

O comando pode abrir recursos existentes somente para leitura ou consulta de acesso. Ele não usa
flags de criação, truncamento, remoção ou alteração; não cria arquivo-sonda. Inspeções Git são
locais, sem shell, locks opcionais, prompts ou remotos.

Não há acesso a GitHub, rede, credenciais ou agentes de IA.

## Help

`cerne doctor --help` documenta finalidade, sintaxe, raiz avaliada, dez verificações, símbolos,
resumos, streams, status, leitura exclusiva, limitações de permissão, erros e ao menos um exemplo.
