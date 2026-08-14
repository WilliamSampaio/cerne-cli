# Referência dos comandos

[English](../en/commands.md) · **Português (Brasil)** · [Español](../es/commands.md)

[Começando](getting-started.md) · [Solução de problemas](troubleshooting.md)

Execute `cerne <comando> --help` para consultar o contrato completo implementado pela sua versão
instalada.

<!-- AUTO-GENERATED: manter sincronizado com cmd/cerne/main.go e os contratos do CLI. -->

## Instalador standalone

Releases para Linux e macOS publicam `install.sh`, `checksums.txt` e binários para `amd64` e
`arm64`. O instalador aceita `--version <version>`, `--agent <codex|claude|gemini>` e `--help`. Ele
instala somente `~/.local/bin/cerne`, verifica SHA-256 e a versão reportada pelo binário antes da
promoção, recusa destino como diretório ou symlink, nunca usa `sudo` e nunca edita arquivos de
perfil do shell.

```sh
curl --proto '=https' --tlsv1.2 -fsSL \
  https://github.com/WilliamSampaio/cerne-cli/releases/latest/download/install.sh | sh
```

## Opções globais

- `cerne --help` exibe os comandos disponíveis e as opções globais.
- `cerne --version` exibe a versão instalada como identificador SemVer estável.
- `cerne --lang <en|pt-BR> ...` seleciona o idioma somente naquela execução, sem salvá-lo.

`CERNE_LANG` oferece a mesma substituição temporária. A precedência é `--lang`, `CERNE_LANG`,
preferência salva e, por fim, `en`. A seleção altera apenas textos destinados a pessoas; comandos,
flags, campos JSON, identificadores, códigos de saída e `--version` permanecem estáveis.

## `cerne config <set language <en|pt-BR>|get language|unset language>`

Administra a preferência de idioma do usuário atual em `~/.cerne/config.json`. `set` persiste um
idioma suportado, `get` exibe o valor salvo e `unset` remove a preferência para voltar ao padrão de
compatibilidade. Para reparar uma configuração regular inválida, use uma seleção temporária, por
exemplo:

```sh
cerne --lang pt-BR config set language pt-BR
```

O Cerne recusa links simbólicos e caminhos de configuração inseguros em vez de segui-los.

## `cerne init <project-name> [--source ... | --clone ...] [--workflow ... [--agent ...]]`

Cria um workspace abaixo do diretório atual. O destino deve estar ausente ou ser um diretório
regular vazio; links simbólicos e destinos não vazios são recusados. Conteúdo existente nunca é
substituído. Em caso de falha, o Cerne desfaz somente os artefatos criados naquela tentativa.

O nome usa de 1 a 255 caracteres ASCII, começa com letra ou número e pode continuar com letras,
números, `.`, `_` ou `-`. Nomes reservados do Windows e nomes terminados em `.` são recusados.
Sem `--workflow`, o comportamento não muda. Com ele, o Cerne executa o provider instalado somente
em knowledge e sem interação. `--agent codex|claude` é aceito somente com `--workflow speckit`;
ele prepara descoberta local para aquela invocação sem persistir o agente. Executável ausente gera
aviso; falha de um provider executado preserva o workspace base e retorna erro operacional.

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

## `cerne restore <origem-knowledge> (--source <caminho> | --clone <origem-source>)`

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

## `cerne skill install <codex|claude|gemini> [cerne-context|cerne-git-workflow]`

Sem argumento de skill, instala todas as skills oficiais compatíveis no perfil do agente do usuário
atual. Codex e Claude recebem `cerne-context` e `cerne-git-workflow`; Gemini recebe somente
`cerne-git-workflow`. Com argumento de skill, instala exatamente aquela skill. Os destinos são
`~/.codex/skills/<skill>`, `~/.claude/skills/<skill>` ou
`~/.gemini/skills/cerne-git-workflow`.

O comando usa o pacote oficial `cerne-skills` incorporado ao binário, sem rede, valida manifesto,
adaptador, entrypoint e schema `cerne.context.v1` antes de copiar e registra auditoria privada em
`~/.cerne/audit` para cada skill instalada.

Uso inválido, incluindo `generic`, caixa diferente, agente ausente ou argumentos extras, retorna
status `2` sem criar auditoria nem alterar arquivos. Falhas operacionais retornam status `1` em
stderr. Reinstalar a mesma versão é no-op; versões gerenciadas diferentes são atualizadas.
`init`, `restore` e `workflow setup` nunca instalam skills por implicação.

## `cerne git inspect`

Fornece a superfície segura de inspeção Git usada por `cerne-git-workflow`. O Cerne não executa
efeitos Git; o agente usa os dados inspecionados e pede confirmação antes de branch, commit, push ou
Pull Request.

```sh
cerne git inspect --agent codex --task task-1 --json
```

`inspect` é somente leitura e retorna schema versão 1 com `state_id` determinístico, remotes
sanitizados, branches locais, paths alterados literais e id privado de auditoria. Comandos de
branch, commit, push e Pull Request não estão disponíveis pelo Cerne; operações Git destrutivas ou
fora de escopo continuam fora da skill.

Sucesso JSON usa stdout/status `0`; snapshot de workspace inválido usa stdout/status `1`; uso
inválido usa stderr/status `2`. Auditorias privadas em `~/.cerne/audit` cobrem somente `inspect` e
excluem conversas, output Git, URLs remotas, tokens, conteúdo de arquivos, corpo de PR e erros
brutos. A evidência da execução pertence ao agente ou harness, não ao audit do Cerne.

## `cerne workflow setup [--agent codex|claude]`

Localiza o workspace ancestral mais próximo e materializa o provider declarado no manifesto. Não
aceita provider, caminho nem opção de força. Com `--agent`, o workflow declarado deve ser Spec Kit
e o Cerne prepara ou atualiza a ponte de descoberta na raiz para o agente local escolhido. Cada
subprocesso real de provider ou integração de agente cria um JSON de auditoria sanitizado em
`knowledge/runs`; executável ausente e layout pronto sem setup de agente não criam auditoria. Para
o Codex descobrir a ponte em `.agents/skills`, inicie a sessão na raiz do workspace Cerne, não dentro
de `source/`.

## `cerne context [--json]`

Localiza o workspace ancestral mais próximo e informa paths canônicos de workspace, knowledge,
product, specs, decisions, policies, source e workflow declarado. `--json` emite o schema estável
versão 1 para skills e scripts. Relatórios saudáveis ou com avisos retornam `0`; relatórios
estruturalmente inválidos continuam válidos e retornam `1`; uso inválido retorna `2` em stderr.

O comando lê somente metadados estruturais. Não lê conteúdo dos repositórios nem arquivos de
agente, executa Git ou providers, consulta remotos ou executáveis, acessa rede ou cria auditoria,
cache, instrução ou alteração no manifesto.

```sh
cerne context
cerne context --json
```

## `cerne doctor`

Executa dez verificações somente de leitura a partir da raiz: manifesto, diretórios dos dois
repositórios, independência Git, isolamento de versionamento, caminhos do manifesto, diretórios
obrigatórios de conhecimento, Git, permissões e versão do manifesto. O comando nunca corrige o
workspace. Um workflow declarado acrescenta uma verificação para os estados pronto, pendente,
indisponível, desconhecido, parcial ou com Git aninhado.

## `cerne status`

Localiza o workspace ancestral mais próximo a partir do diretório atual e lê os dois repositórios.
Reconhece árvore limpa ou com alterações, detached HEAD e repositórios sem commits. Não executa
fetch nem compara com remotos.

## `cerne link <caminho> [--replace]`

Vincula como `source` um repositório Git local não-bare com árvore de trabalho. Aceita caminhos
relativos, absolutos e worktrees válidos. Knowledge e source devem ser distintos e não podem ter
aninhamento perigoso. Trocar um source já configurado exige `--replace`; vincular o mesmo source
conclui sem regravar o manifesto. A substituição do manifesto é atômica.

<!-- END AUTO-GENERATED -->
