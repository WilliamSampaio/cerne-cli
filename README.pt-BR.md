# Cerne

[![Testes](https://github.com/WilliamSampaio/cerne-cli/actions/workflows/test.yml/badge.svg)](https://github.com/WilliamSampaio/cerne-cli/actions/workflows/test.yml)
[![Licença: MIT](https://img.shields.io/badge/licen%C3%A7a-MIT-blue.svg)](LICENSE)

[English](README.md) · **Português (Brasil)** · [Español](README.es.md)

O Cerne é um CLI open source e multiplataforma, escrito em Go, para administrar workspaces de
software formados por dois repositórios Git independentes:

- **knowledge** — intenção do projeto, informações do produto, especificações, decisões, políticas
  e registros de execução; normalmente privado;
- **source** — código-fonte da aplicação.

O nome *Cerne* representa o núcleo. O projeto começa pela administração local e segura do workspace
e foi concebido para evoluir para um harness independente de modelos e fornecedores, capaz de
coordenar agentes de IA em documentação, produto, implementação, validação e manutenção.

## Por que Cerne?

O Cerne segue algumas regras duradouras:

- seu conhecimento pertence a você e permanece acessível como arquivos comuns e histórico Git;
- conhecimento privado e código da aplicação ficam em repositórios separados;
- integrações ficam atrás de adaptadores, sem contaminar o domínio;
- trabalho automatizado deve ser rastreável e receber apenas o contexto necessário;
- push, merge, publicação, deploy e operações destrutivas exigem autorização explícita;
- segredos e credenciais nunca devem ser armazenados nos repositórios administrados.

A versão atual é deliberadamente local. Ela não chama agentes de IA, administra serviços de
hospedagem, publica ou realiza deploy. Só acessa uma origem Git com `init --clone` explícito.

## Requisitos

- Git disponível no `PATH`;
- Go 1.26.5 ou mais recente para compilar o projeto;
- Linux, Windows ou macOS.

O executável `specify`, do Spec Kit, ou `openspec`, do OpenSpec, é opcional e necessário somente
quando esse workflow é selecionado. O Cerne nunca instala nem atualiza essas ferramentas.

## Instalação

Instale diretamente com Go:

```sh
go install github.com/WilliamSampaio/cerne-cli/cmd/cerne@latest
cerne --version
cerne --help
```

O Go coloca o binário em `GOBIN` ou em `GOPATH/bin` quando `GOBIN` não está definido. Certifique-se
de que esse diretório esteja no `PATH`.

Para compilar uma cópia de desenvolvimento:

```sh
git clone https://github.com/WilliamSampaio/cerne-cli.git
cd cerne-cli
go build -o cerne ./cmd/cerne
./cerne --version
```

No Windows, o binário gerado é `cerne.exe`.

## Início rápido

### 1. Crie um workspace

```sh
cerne init meu-projeto
cd meu-projeto
```

O Cerne cria:

```text
meu-projeto/
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

Os dois repositórios são locais, independentes e começam sem commits ou remotos. A raiz do
workspace não é um repositório Git.

Como o Git não registra diretórios vazios, o Cerne cria um arquivo `.gitkeep` em cada diretório
obrigatório de `knowledge`. Você pode removê-lo quando adicionar conteúdo ao diretório. O Cerne não
cria commits automaticamente.

Para começar com um source existente, escolha exatamente um modo de source:

```sh
cerne init meu-projeto --source ../aplicacao-existente
cerne init meu-projeto --clone https://host/organizacao/aplicacao.git
```

`--source` vincula um working tree Git local non-bare, resolve caminhos relativos a partir do
diretório da invocação e nunca cria source interno nem altera o repositório externo. `--clone`
aceita path local existente, `file`, HTTPS ou SSH (inclusive SCP-like), faz um clone padrão completo
no `source` interno e mantém o remoto `origin`. `--source` e `--clone` são mutuamente exclusivas;
qualquer uma pode ser combinada com `--workflow`.

Para inicializar um workflow opcional de especificação durante a criação:

```sh
cerne init meu-projeto --workflow speckit
cerne init meu-projeto --workflow openspec
cerne init meu-projeto --clone https://host/organizacao/aplicacao.git --workflow speckit
```

O Spec Kit mantém as especificações em `knowledge/specs` e controla `knowledge/.specify`. O
OpenSpec usa `knowledge/openspec/specs` e controla `knowledge/openspec`, sem criar o diretório
superior `knowledge/specs`. Produto, decisões, políticas e execuções permanecem comuns.

Se o executável estiver ausente, o `init` ainda conclui, registra a escolha e avisa em stderr.
Depois de instalar a ferramenta, execute `cerne workflow setup` de qualquer diretório dentro do
workspace. O setup é idempotente; estruturas parciais ou com Git aninhado são recusadas.

### 2. Valide a estrutura

Execute a partir da raiz do workspace:

```sh
cerne doctor
```

O relatório identifica cada verificação com `✓` (aprovada), `!` (aviso) ou `✗` (erro bloqueante).

### 3. Consulte o estado local do Git

```sh
cerne status
```

O comando mostra branch, commit abreviado, estado da árvore, arquivos em stage, modificados e não
rastreados nos dois repositórios. Alterações pendentes são informação, não erro.

### 4. Vincule um source existente (opcional)

O `init` já configura o repositório `source` vazio. Use `--replace` para apontar o manifesto para
outro repositório local:

```sh
cerne link ../aplicacao-existente --replace
```

Somente a referência do manifesto é alterada. O Cerne nunca copia, move, limpa, faz checkout,
commit ou exclui o source anterior ou o novo.

## Manifesto

O arquivo `knowledge/cerne.json` identifica o projeto e localiza o repositório source:

```json
{
  "name": "meu-projeto",
  "source": "../source"
}
```

A ausência de `version` representa a versão 1 do manifesto. Quando presente, o único valor aceito
atualmente é o inteiro JSON `1`. O Cerne armazena um caminho source relativo e normalizado sempre
que as plataformas e localizações permitirem.

Com um workflow selecionado, o manifesto também contém `"workflow":{"provider":"speckit"}` ou
`"workflow":{"provider":"openspec"}`. Estado de instalação e versão não são persistidos.

## Referência dos comandos

### Opções globais

- `cerne --help` exibe os comandos disponíveis e as opções globais.
- `cerne --version` exibe o identificador SemVer estável, atualmente `cerne 0.5.0`.

### `cerne init <project-name> [--source ... | --clone ...] [--workflow ...]`

Cria um workspace abaixo do diretório atual. O destino deve estar ausente ou ser um diretório
regular vazio; links simbólicos e destinos não vazios são recusados. Conteúdo existente nunca é
substituído. Em caso de falha, o Cerne desfaz somente os artefatos criados naquela tentativa.

O nome usa de 1 a 255 caracteres ASCII, começa com letra ou número e pode continuar com letras,
números, `.`, `_` ou `-`. Nomes reservados do Windows e nomes terminados em `.` são recusados.
Sem a opção, o comportamento não muda. Com ela, o Cerne executa o provider instalado somente em
knowledge, sem interação e sem selecionar agente de IA. Executável ausente gera aviso; falha de um
provider executado preserva o workspace base e retorna erro operacional.

`--source` valida e vincula um working tree local existente sem modificá-lo. `--clone` recusa HTTP,
`git://`, `ext::`, helpers desconhecidos, valores semelhantes a opções, credenciais embutidas,
query e fragmento antes de executar Git. Autenticação, redirects e filtros de checkout continuam
sob responsabilidade do Git; o Cerne desabilita prompts controláveis, mas helpers externos ainda
podem falhar ou agir fora do controle portátil do CLI. O clone não acrescenta depth, branch,
submódulos, LFS, push ou fetch extra. Qualquer opção de source pode ser combinada com
`--workflow`; source e clone continuam exclusivos.

Cada clone iniciado cria antes uma auditoria sanitizada em `knowledge/runs/source-clone.json`.
Falhas anteriores ao clone revertem a tentativa. Falhas posteriores preservam knowledge e
auditoria, removem somente o staging privado do Cerne e informam que o workspace ficou incompleto.
A promoção nunca substitui um source concorrente; se a auditoria final falhar após a promoção, o
source válido permanece.

### `cerne restore <origem-knowledge> (--source <caminho> | --clone <origem-source>)`

Clona knowledge, lê o nome do workspace em `cerne.json` e então clona source no caminho portátil do
manifesto ou vincula a raiz de um Git local não-bare sem modificá-la. O destino deve estar ausente.
Layouts existentes, concorrentes, sobrepostos, com symlink, parciais, provider desconhecido ou Git
não independente são recusados sem substituição. Workflow pronto ou pendente é preservado, nunca
executado.

Cada tentativa válida inicia antes do Git um registro privado e sanitizado em `~/.cerne/audit`.
Origens, credenciais, saída Git, argumentos, ambiente e paths absolutos dos repositórios são
excluídos. Autenticação e redirects continuam sob responsabilidade do Git externo. Falhas revertem
somente artefatos cuja identidade ainda pertença à tentativa; repetir o comando é a retomada
suportada, sem `--resume`. Sucesso/ajuda usam stdout e status `0`, falhas operacionais stderr/status
`1` e uso/origem inválida stderr/status `2`. Restore não autoriza workflow, agente, push, merge,
fetch extra, submódulo, instalação, publicação ou deploy.

```sh
cerne restore ../knowledge.git --clone ../source.git
cerne restore git@host:org/knowledge.git --source ../source-existente
```

### `cerne workflow setup`

Localiza o workspace ancestral mais próximo e materializa o provider declarado no manifesto. Não
aceita provider, caminho nem opção de força. Cada tentativa real cria um JSON de auditoria
sanitizado em `knowledge/runs`; executável ausente e layout pronto não criam auditoria.

### `cerne doctor`

Executa dez verificações somente de leitura a partir da raiz: manifesto, diretórios dos dois
repositórios, independência Git, isolamento de versionamento, caminhos do manifesto, diretórios
obrigatórios de conhecimento, Git, permissões e versão do manifesto. O comando nunca corrige o
workspace. Um workflow declarado acrescenta uma verificação para os estados pronto, pendente,
indisponível, desconhecido, parcial ou com Git aninhado.

### `cerne status`

Localiza o workspace ancestral mais próximo a partir do diretório atual e lê os dois repositórios.
Reconhece árvore limpa ou com alterações, detached HEAD e repositórios sem commits. Não executa
fetch nem compara com remotos.

### `cerne link <caminho> [--replace]`

Vincula como `source` um repositório Git local não-bare com árvore de trabalho. Aceita caminhos
relativos, absolutos e worktrees válidos. Knowledge e source devem ser distintos e não podem ter
aninhamento perigoso. Trocar um source já configurado exige `--replace`; vincular o mesmo source
conclui sem regravar o manifesto. A substituição do manifesto é atômica.

Use `<comando> --help` para consultar o contrato completo. A saída do CLI está atualmente em
português.

## Códigos de saída e streams

| Código | Significado |
| --- | --- |
| `0` | Sucesso, ajuda, workspace saudável, somente avisos ou status pendente consultado com sucesso |
| `1` | Falha operacional ou erro bloqueante encontrado pelo `doctor` |
| `2` | Uso inválido do comando ou nome de projeto inválido |

Saídas normais e ajuda usam stdout. Erros de uso e falhas operacionais usam stderr. Os relatórios do
`doctor`, inclusive com erros bloqueantes, usam stdout para manter o diagnóstico em um único stream.

## Segurança e privacidade

- `doctor` e `status` são somente de leitura.
- `link` atualiza apenas `knowledge/cerne.json`, depois que todas as validações passam.
- O setup usa argumentos fixos, não usa shell, recebe ambiente mínimo e desabilita a telemetria do
  OpenSpec. Não recebe credenciais nem o caminho de source e não registra a saída bruta do provider.
- O clone usa argumentos Git fixos sem shell, allowlist de protocolos, staging privado e promoção
  sem substituição. Origem e saída Git bruta não entram na saída do Cerne, manifesto ou auditoria;
  a autenticação permanece externa e o Git mantém a origem como remoto `origin`.
- `restore` mantém o audit privado em `~/.cerne/audit`, valida os dois limites Git e usa rollback
  por identidade com promoção sem substituição.
- Uma tentativa falha preserva o workspace base e a auditoria, removendo somente uma nova raiz
  pertencente ao provider.
- A inspeção Git desativa locks opcionais e prompts e remove variáveis `GIT_*` capazes de
  redirecionar os processos filhos.
- Somente `init --clone` ou `restore` explícito pode acessar uma origem ou usar credenciais externas.
- Não coloque tokens, senhas, chaves privadas ou outros segredos nos repositórios administrados.

## Projeto técnico

O código mantém responsabilidades pequenas e explícitas:

```text
cmd/cerne/          argumentos, saída do terminal e códigos de saída
internal/workspace/ regras de domínio e operações do workspace
internal/gitexec/   adaptador para o executável Git local
internal/filecheck/ verificações multiplataforma de permissões
internal/workflowexec/ adaptadores para executáveis locais opcionais de workflow
specs/              especificações, planos, contratos e tarefas
```

A implementação prefere a biblioteca padrão de Go. Comportamentos específicos de sistema de
arquivos são isolados com build tags. O CI executa os testes em Linux, Windows e macOS. O domínio é
separado da impressão no terminal para permitir reutilização em interfaces futuras.

## Desenvolvimento

```sh
go build -o cerne ./cmd/cerne
go test ./...
go test -count=1 ./...
go vet ./...
gofmt -w <arquivos-go-alterados>
```

Os testes usam o pacote `testing`, diretórios temporários e apenas repositórios Git locais. Eles não
dependem de rede nem credenciais.

## Como contribuir

Contribuições são bem-vindas:

1. Abra uma issue ou discuta o comportamento antes de uma alteração grande.
2. Crie uma branch focada e mantenha as regras de domínio fora da impressão do terminal.
3. Adicione ou atualize um teste que falhe sem a mudança proposta.
4. Execute `gofmt`, `go vet ./...` e `go test -count=1 ./...`.
5. Abra um pull request explicando o objetivo, a issue ou artefato em `specs/`, os comandos de
   validação e qualquer impacto de compatibilidade no CLI.

Use assuntos curtos no estilo Conventional Commits, como `feat: add command` ou
`fix: preserve manifest`. Consulte [AGENTS.md](AGENTS.md) para regras de contribuição e a
[constituição do projeto](.specify/memory/constitution.md) para governança e compatibilidade.
O histórico de versões está documentado em [CHANGELOG.md](CHANGELOG.md).

## Roadmap e escopo

O escopo atual inclui criação com source vazio, vinculado ou clonado, bootstrap opcional de
workflow, validação, status local e vínculo de source. Futuramente, o Cerne poderá coordenar agentes auditáveis para produto,
documentação, implementação, validação e manutenção, sem depender de modelos, agentes ou
fornecedores específicos.

Administração de hospedagem remota, commits automáticos, push, pull request, merge, publicação,
deploy, interface gráfica, saída JSON e execução de IA não fazem parte do CLI atual.

## Licença

O Cerne é distribuído sob a [Licença MIT](LICENSE).
