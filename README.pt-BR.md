# Cerne

[![Testes](https://github.com/WilliamSampaio/cerne-cli/actions/workflows/test.yml/badge.svg)](https://github.com/WilliamSampaio/cerne-cli/actions/workflows/test.yml)
[![Licença: MIT](https://img.shields.io/badge/licen%C3%A7a-MIT-blue.svg)](LICENSE)

[English](README.md) · **Português (Brasil)** · [Español](README.es.md)

[Documentação do usuário](docs/pt-BR/getting-started.md)

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
Com Spec Kit, `--agent codex|claude` também pode preparar a descoberta local de comandos na raiz do
workspace; a escolha do agente não é gravada em `knowledge/cerne.json`.

## Instalação

No Linux e macOS, instale o binário standalone estável mais recente sem Go:

```sh
curl --proto '=https' --tlsv1.2 -fsSL \
  https://github.com/WilliamSampaio/cerne-cli/releases/latest/download/install.sh | sh
~/.local/bin/cerne --version
```

Inspecione o instalador antes usando o mesmo `curl` sem `| sh`. Para instalar uma versão fixa, use
a URL da release daquela tag:

```sh
curl --proto '=https' --tlsv1.2 -fsSL \
  https://github.com/WilliamSampaio/cerne-cli/releases/download/vX.Y.Z/install.sh |
  sh -s -- --version vX.Y.Z
```

Para instalar as skills opcionais de um agente no mesmo fluxo explícito:

```sh
curl --proto '=https' --tlsv1.2 -fsSL \
  https://github.com/WilliamSampaio/cerne-cli/releases/latest/download/install.sh |
  sh -s -- --agent codex
```

`--agent codex` e `--agent claude` instalam todas as skills oficiais compatíveis. `--agent gemini`
instala somente a skill de fluxo Git.

O instalador escreve somente `~/.local/bin/cerne`, verifica checksums da release antes de substituir
um arquivo regular, recusa destino como diretório ou symlink, nunca usa `sudo` e nunca edita perfis
de shell. Remova com `rm ~/.local/bin/cerne`. Para instalação manual, baixe o arquivo compatível e
`checksums.txt` na release, verifique o SHA-256, extraia `cerne` e coloque-o no `PATH`.

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

Para instalar o estado local do checkout no binário usado pelo host, sem depender de commits ou do
remoto:

```sh
go install ./cmd/cerne
# ou
make install-local
```

No Windows, o binário gerado é `cerne.exe`.

## Idioma

As mensagens do CLI estão disponíveis em inglês e português do Brasil. Salve uma preferência ou
substitua-a somente em uma execução:

```sh
cerne config set language pt-BR
cerne --lang en doctor
CERNE_LANG=en cerne status
```

A precedência é `--lang`, `CERNE_LANG`, preferência salva e, por fim, `pt-BR`. O padrão atual
permanece `pt-BR` por compatibilidade e mudará para `en` na versão 1.0. Saídas estruturadas,
comandos, flags, identificadores, códigos de saída e a versão são neutros quanto ao idioma.

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
cerne init meu-projeto --workflow speckit --agent codex
cerne init meu-projeto --workflow openspec
cerne init meu-projeto --clone https://host/organizacao/aplicacao.git --workflow speckit
```

O Spec Kit mantém as especificações em `knowledge/specs` e controla `knowledge/.specify`. O
OpenSpec usa `knowledge/openspec/specs` e controla `knowledge/openspec`, sem criar o diretório
superior `knowledge/specs`. Produto, decisões, políticas e execuções permanecem comuns.

Se o executável estiver ausente, o `init` ainda conclui, registra a escolha e avisa em stderr.
Depois de instalar a ferramenta, execute `cerne workflow setup` de qualquer diretório dentro do
workspace. O setup é idempotente; estruturas parciais ou com Git aninhado são recusadas.
Quando `--agent codex` ou `--agent claude` é usado com Spec Kit, o Cerne também pede ao Spec Kit
para criar a integração correspondente dentro de `knowledge` e grava pequenas pontes na raiz do
workspace em `.agents/skills` ou `.claude/skills`. Essas pontes apontam de volta para `knowledge` e
não contêm conhecimento privado, remotos, credenciais, dumps de ambiente ou paths absolutos. Para o
Codex descobrir essas skills locais, inicie a sessão a partir da raiz do workspace Cerne; uma sessão
iniciada dentro de `source/` não sobe para a raiz do workspace porque `source` é um repositório Git
separado.

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

### 4. Coordene uma etapa Git aprovada (opcional)

Instale `cerne-git-workflow` com `cerne skill install <agent>` e deixe a skill inspecionar primeiro
e pedir uma confirmação separada antes de cada mutação:

```sh
cerne git inspect --agent codex --task task-1 --json
cerne git branch create --name feat/exemplo --base knowledge=main --base source=main --state <state-id> --confirm --agent codex --task task-1 --json
cerne git commit source --message "feat: exemplo" --include caminho/arquivo.go --state <state-id> --confirm --agent codex --task task-1 --json
cerne git push source --remote origin --branch feat/exemplo --state <state-id> --confirm --agent codex --task task-1 --json
cerne git pr create source --remote origin --base main --head feat/exemplo --title "feat: exemplo" --body-file pr.md --state <state-id> --confirm --agent codex --task task-1 --json
```

Commit, push e Pull Request são etapas separadas. A criação de PR suporta GitHub.com e lê tokens
somente de `GH_TOKEN` ou `GITHUB_TOKEN`. Operações Git destrutivas ou fora de escopo são recusadas.
Veja a [referência de comandos](docs/pt-BR/commands.md) para o contrato completo de JSON/auditoria.

### 5. Vincule um source existente (opcional)

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
`"workflow":{"provider":"openspec"}`. Estado de instalação, versão e escolha local de agente não
são persistidos.

## Referência dos comandos

Consulte a [referência completa dos comandos do Cerne](docs/pt-BR/commands.md).

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
- `skill install` escreve somente no perfil do agente autorizado, valida o pacote oficial
  incorporado antes da cópia e recusa conteúdo desconhecido no destino.
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
internal/skillinstall/ instalação global explícita das skills oficiais
specs/              especificações, planos, contratos e tarefas
```

A implementação prefere a biblioteca padrão de Go. Comportamentos específicos de sistema de
arquivos são isolados com build tags. O CI executa os testes em Linux, Windows e macOS. O domínio é
separado da impressão no terminal para permitir reutilização em interfaces futuras.

## Desenvolvimento

```sh
go build -o cerne ./cmd/cerne
go install ./cmd/cerne
go test ./...
go test -count=1 ./...
go vet ./...
gofmt -w <arquivos-go-alterados>
```

Atalhos equivalentes estão disponíveis via `make build`, `make install-local`, `make test`,
`make test-fresh`, `make vet`, `make fmt` e `make check`. Use `make install-path` para ver onde
`make install-local` disponibiliza o executável. Para instalar em outro diretório, defina `GOBIN`,
por exemplo `GOBIN=/tmp/bin make install-local`.

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
