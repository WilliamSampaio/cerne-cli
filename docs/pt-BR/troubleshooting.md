# Solução de problemas

[English](../en/troubleshooting.md) · **Português (Brasil)** ·
[Español](../es/troubleshooting.md)

[Começando](getting-started.md) · [Comandos](commands.md)

O Cerne envia erros e a `correção:` sugerida para stderr. Siga primeiro essa correção.

## O comando `cerne` ou o Git não é encontrado

Se o shell não encontrar `cerne`, adicione o diretório `GOBIN` ou `GOPATH/bin` do Go ao `PATH`.
Confirme a instalação com:

```sh
cerne --version
```

Para `erro: Git indisponível`, instale o Git e verifique se `git --version` funciona no mesmo shell.

## Workspace não localizado

`erro: workspace Cerne não localizado` significa que o diretório atual não está dentro de um
workspace que contenha `knowledge/cerne.json`. Entre no workspace antes de executar `status`,
`context`, `link` ou `workflow setup`.

O `cerne doctor` é diferente: execute-o na raiz do workspace.

```sh
cd meu-projeto
cerne doctor
cerne context
```

Se `knowledge/cerne.json` foi removido ou danificado, restaure-o pelo repositório knowledge antes de
continuar.

## O `init` recusa o destino

O Cerne nunca substitui conteúdo existente. Use um nome de projeto cujo destino não exista ou seja
um diretório regular vazio. Inspecione o diretório antes de remover qualquer coisa.

Uma aplicação local existente deve ser vinculada em vez de usada como destino:

```sh
cerne init meu-projeto --source ../aplicacao-existente
```

## Um source local é recusado

O caminho usado em `--source` ou `link` deve ser a raiz de um repositório Git existente, não-bare e
com working tree. Ele deve ser independente de `knowledge` e não pode sobrepor caminhos protegidos
do workspace.

```sh
git -C ../aplicacao-existente status
cerne link ../aplicacao-existente --replace
```

Use `--replace` somente quando outro source já estiver configurado. Ele substitui a referência no
manifesto, não os repositórios.

## O setup do workflow está pendente ou falha

Se o `init` informar que o executável `specify` ou `openspec` está ausente, o workspace ainda foi
criado. Instale separadamente o provider escolhido e execute dentro do workspace:

```sh
cerne workflow setup
```

Para `estrutura do workflow inválida ou parcial`, não repita o setup indefinidamente. Inspecione e
corrija o diretório parcial pertencente ao provider e execute `cerne doctor` antes de tentar outra
vez.

## Um clone de source deixa o workspace incompleto

Uma falha depois do início do clone pode preservar `knowledge` e sua auditoria sanitizada, deixando
o workspace incompleto. Primeiro inspecione o registro:

```text
knowledge/runs/source-clone.json
```

Depois, associe um source local válido ou remova manualmente o workspace incompleto após confirmar
que ele não contém nada necessário. O `init` não possui modo de retomada.

## O `restore` falha

O destino derivado do manifesto restaurado não pode existir. Autenticação e acesso remoto são
responsabilidade do Git; confirme que as origens funcionam com sua configuração normal do Git.

O restore não possui modo de retomada. Leia a correção exibida e o registro privado em
`~/.cerne/audit`, corrija a causa e tente novamente. O Cerne não substitui um destino existente
durante a recuperação.

## O pacote da skill do agente não está disponível

`erro: pacote oficial cerne-skills incorporado está inacessível` significa que o pacote incorporado
não pôde ser materializado ou validado. Verifique o acesso ao diretório temporário do sistema e
reinstale o Cerne antes de repetir `cerne skill install <codex|claude>`.
