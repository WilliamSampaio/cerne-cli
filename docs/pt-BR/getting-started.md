# Começando com o Cerne

[English](../en/getting-started.md) · **Português (Brasil)** ·
[Español](../es/getting-started.md)

O Cerne administra um workspace de software com dois repositórios Git independentes:

- `knowledge` guarda contexto de produto, decisões, políticas e registros de execução. Normalmente
  é privado.
- `source` guarda o código-fonte da aplicação.

A raiz do workspace conecta esses repositórios, mas não é um repositório Git. O Cerne não faz
commit, push, publicação ou deploy do seu trabalho.

## Antes de começar

Você precisa do Git no `PATH` e de Linux, Windows ou macOS. Para instalar o Cerne com Go:

```sh
go install github.com/WilliamSampaio/cerne-cli/cmd/cerne@latest
cerne --version
```

O Go instala o binário em `GOBIN` ou em `GOPATH/bin` quando `GOBIN` não está definido. Adicione esse
diretório ao `PATH` caso o comando `cerne` não seja encontrado.

## Crie seu workspace

Escolha o comando que representa onde seu código-fonte está agora:

| Ponto de partida | Comando | O que acontece |
| --- | --- | --- |
| Projeto novo | `cerne init meu-projeto` | Cria repositórios `knowledge` e `source` vazios. |
| Repositório local existente | `cerne init meu-projeto --source ../minha-app` | Vincula o repositório sem movê-lo ou alterá-lo. |
| Repositório remoto existente | `cerne init meu-projeto --clone git@host:org/minha-app.git` | Clona o repositório no workspace como `source`. |

Entre no workspace e verifique sua estrutura:

```sh
cd meu-projeto
cerne doctor
cerne status
cerne context
```

`doctor` valida a estrutura e os limites de segurança do workspace. `status` resume o estado local
do Git nos dois repositórios. `context` mostra os caminhos e o workflow opcional detectado pelo
Cerne. Esses comandos são somente leitura.

## Entenda a estrutura

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

Trabalhe com `knowledge` e `source` como repositórios separados: cada um tem seu próprio histórico,
branches, commits e remotos. Não armazene credenciais ou segredos em nenhum deles.

## Workflow opcional

Você pode preparar um workflow de especificação compatível ao criar o workspace:

```sh
cerne init meu-projeto --workflow speckit
cerne init meu-projeto --workflow openspec
```

O Cerne usa uma instalação local existente do `specify` ou `openspec`; ele nunca instala ou atualiza
essas ferramentas. Se o executável não estiver disponível durante a criação, instale-o separadamente
e depois execute:

```sh
cerne workflow setup
```

## Restaure um workspace existente

A restauração sempre precisa de uma origem para `knowledge` e de uma origem ou caminho local para
`source`:

```sh
cerne restore git@host:org/knowledge.git --clone git@host:org/source.git
cerne restore ../knowledge.git --source ../source-existente
```

O Cerne cria um novo destino e se recusa a substituir um destino existente.

## Detalhes dos comandos

Consulte a [referência dos comandos](commands.md), os
[códigos de saída](../../README.pt-BR.md#códigos-de-saída-e-streams) e as regras de
[segurança e privacidade](../../README.pt-BR.md#segurança-e-privacidade). Para orientações de
recuperação, consulte a [solução de problemas](troubleshooting.md).
