# Research: Descoberta de Agente para Spec Kit

## Decisão 1: `--agent` é escolha local e não entra no manifesto

**Decision**: `--agent codex|claude` seleciona somente a descoberta local da máquina atual. O
manifesto continua persistindo apenas `workflow.provider`.

**Rationale**: `knowledge` viaja entre máquinas e usuários. Um workspace criado com Codex pode ser
restaurado em outra máquina onde Claude é o agente preferido. Persistir agente no manifesto criaria
acoplamento indevido entre conhecimento do projeto e ferramenta local.

**Alternatives considered**: Persistir `workflow.agent`, que impede troca natural após restore;
criar configuração local versionada, que adiciona estado sem necessidade atual; inferir agente pelo
ambiente, que quebra automação e pode escolher errado.

## Decisão 2: Suporte inicial limitado a Spec Kit, Codex e Claude

**Decision**: A primeira versão aceita `--agent` somente com workflow Spec Kit e somente os valores
`codex` e `claude`.

**Rationale**: O defeito observado é específico do Spec Kit com comandos não descobertos a partir da
raiz Cerne. OpenSpec foi integrado com `--tools none` e não tem o mesmo contrato de skills nesta
feature. Codex e Claude são os agentes necessários para o projeto `cerne-skills`.

**Alternatives considered**: Aceitar `--agent` com OpenSpec, que amplia escopo sem caso testado;
expor `generic`, que é detalhe técnico e não alvo de usuário; aceitar qualquer string, que empurra
erros para o provider e enfraquece contratos do CLI.

## Decisão 3: Ponte de descoberta fica na raiz do workspace

**Decision**: O Cerne prepara artefatos de descoberta no root do workspace Cerne:

```text
workspace/.agents/skills/...   # codex
workspace/.claude/skills/...   # claude
```

O projeto Spec Kit real permanece em:

```text
workspace/knowledge/.specify/...
workspace/knowledge/specs/...
```

**Rationale**: O agente normalmente é iniciado na raiz do workspace Cerne. Colocar apenas
`knowledge/.agents` ou `knowledge/.claude` exige que o usuário mude o cwd para `knowledge`, que foi
o problema original.

**Alternatives considered**: Mover `.specify` para a raiz do workspace, que quebra a separação do
repositório de conhecimento; pedir sempre `cd knowledge`, que deixa o fluxo frágil; symlink para a
pasta de `knowledge`, que é menos portátil e pode confundir versionamento.

## Decisão 4: Usar integração oficial do Spec Kit dentro de `knowledge`

**Decision**: Quando um agente é solicitado, o Cerne deve garantir que a integração oficial
correspondente exista em `knowledge` e então criar a ponte local da raiz para essa integração.
Para um layout generic já pronto, a pesquisa local confirmou que `specify integration install
codex --force --integration-options=--skills` cria `knowledge/.agents/skills` mantendo default
generic. O Cerne também aceita `knowledge/skills` como layout compatível para Codex quando esse
conjunto completo já existir. `specify integration install claude --force` cria
`knowledge/.claude/skills` mantendo default generic.

**Rationale**: Reusar a integração oficial mantém o conteúdo das skills alinhado à versão local do
Spec Kit e evita o Cerne empacotar templates de terceiros. O `--force` é necessário para adicionar
uma integração ao lado de `generic`, sem trocar o default.

**Alternatives considered**: Gerar todos os comandos no Cerne, que duplicaria o Spec Kit; trocar o
default de `generic` para o agente, que altera estado de provider sem necessidade; copiar arquivos
de uma versão vendorizada, que envelhece.

## Decisão 5: Ponte local deve ser wrapper, não cópia cega

**Decision**: A ponte na raiz deve conter artefatos gerenciados que direcionam o agente para usar
`knowledge` como raiz Spec Kit. Ela não deve copiar conhecimento privado nem fingir que a raiz Cerne
é um projeto Spec Kit.

**Rationale**: As skills oficiais esperam `.specify` no projeto ativo. Uma cópia cega para a raiz
faria o agente procurar `.specify` no lugar errado. Um wrapper pequeno preserva descoberta no agente
e deixa claro que execução real usa `knowledge`.

**Alternatives considered**: Copiar os `SKILL.md` oficiais para a raiz sem adaptação, que reproduz
o bug em outro formato; criar links simbólicos, que são frágeis no Windows e em repositórios Git;
duplicar `.specify` na raiz, que cria duas fontes de verdade.

## Decisão 6: Pronto do workflow e pronto da ponte são estados separados

**Decision**: O estado `ready` do workflow continua significando que o layout do provider em
`knowledge` é válido. A ponte local é reportada separadamente no stdout quando `--agent` é usado.

**Rationale**: Um workspace restaurado pode ter workflow pronto e nenhuma ponte local. Isso não deve
tornar o workflow inválido; apenas significa que o usuário ainda não preparou um agente naquela
máquina.

**Alternatives considered**: Fazer `doctor/context` bloquearem ponte ausente, que amplia a feature;
persistir estado da ponte no manifesto, que envelhece entre máquinas; considerar workflow pendente
quando só falta descoberta local, que confunde provider com agente.

## Decisão 7: Falhas de ponte não podem mascarar falhas do provider

**Decision**: Se o provider estiver ausente ou falhar, a ponte local não é criada nem reportada como
pronta. Se o provider já estiver pronto e a ponte falhar, o comando falha com status operacional e
correção segura, sem alterar `source`.

**Rationale**: O usuário não deve receber comandos descobríveis que não conseguem operar sobre um
Spec Kit válido. Separar as fases deixa a recuperação explícita.

**Alternatives considered**: Criar ponte antes do provider, que pode gerar comandos quebrados;
ignorar falha de ponte como warning, que deixa o problema original invisível; rollback total do
workspace, que apagaria um workflow válido.

## Decisão 8: Instalação de integração de agente é execução auditável

**Decision**: Qualquer `specify integration install` executado pelo Cerne para preparar `codex` ou
`claude` em `knowledge` MUST criar e finalizar auditoria em `knowledge/runs`, usando o mesmo modelo
seguro de workflow setup.

**Rationale**: A constituição exige rastreabilidade para execução automatizada. Embora a escolha do
agente seja local e não persistida no manifesto, o subprocesso modifica `knowledge` e precisa ser
reconstruível sem expor saída bruta, ambiente ou segredos.

**Alternatives considered**: Tratar como detalhe da ponte local, que viola auditoria; registrar só
stdout/stderr, que pode vazar segredo; não usar subprocesso, que duplicaria assets do Spec Kit.
