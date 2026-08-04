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

A versão atual é deliberadamente local. Ela não chama agentes de IA, acessa GitHub, clona remotos,
publica ou realiza deploy.

## Requisitos

- Git disponível no `PATH`;
- Go 1.24.6 ou mais recente para compilar o projeto;
- Linux, Windows ou macOS.

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

O Git não versiona diretórios vazios. Adicione conhecimento do projeto antes do primeiro commit de
knowledge para preservar os diretórios necessários; o Cerne não cria placeholders nem commits
automáticos de propósito.

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

## Referência dos comandos

### Opções globais

- `cerne --help` exibe os comandos disponíveis e as opções globais.
- `cerne --version` exibe o identificador SemVer estável, atualmente `cerne 0.1.0`.

### `cerne init <project-name>`

Cria um workspace abaixo do diretório atual. O destino deve estar ausente ou ser um diretório
regular vazio; links simbólicos e destinos não vazios são recusados. Conteúdo existente nunca é
substituído. Em caso de falha, o Cerne desfaz somente os artefatos criados naquela tentativa.

O nome usa de 1 a 255 caracteres ASCII, começa com letra ou número e pode continuar com letras,
números, `.`, `_` ou `-`. Nomes reservados do Windows e nomes terminados em `.` são recusados.

### `cerne doctor`

Executa dez verificações somente de leitura a partir da raiz: manifesto, diretórios dos dois
repositórios, independência Git, isolamento de versionamento, caminhos do manifesto, diretórios
obrigatórios de conhecimento, Git, permissões e versão do manifesto. O comando nunca corrige o
workspace.

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
- A inspeção Git desativa locks opcionais e prompts e remove variáveis `GIT_*` capazes de
  redirecionar os processos filhos.
- Nenhum comando atual acessa remotos ou precisa de credenciais.
- Não coloque tokens, senhas, chaves privadas ou outros segredos nos repositórios administrados.

## Projeto técnico

O código mantém responsabilidades pequenas e explícitas:

```text
cmd/cerne/          argumentos, saída do terminal e códigos de saída
internal/workspace/ regras de domínio e operações do workspace
internal/gitexec/   adaptador para o executável Git local
internal/filecheck/ verificações multiplataforma de permissões
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

## Roadmap e escopo

O escopo atual inclui criação e validação do workspace, status local e vínculo de um repositório
source local existente. Futuramente, o Cerne poderá coordenar agentes auditáveis para produto,
documentação, implementação, validação e manutenção, sem depender de modelos, agentes ou
fornecedores específicos.

Administração de repositórios remotos, commits automáticos, push, pull request, merge, publicação,
deploy, interface gráfica, saída JSON e execução de IA não fazem parte do CLI atual.

## Licença

O Cerne é distribuído sob a [Licença MIT](LICENSE).
