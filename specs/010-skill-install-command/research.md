# Research: Instalação de Skills Cerne

## Decisão 1: Instalação somente por comando explícito

**Decision**: Implementar apenas `cerne skill install <agent>` como operação que escreve no perfil
global do usuário. `init`, `restore` e `workflow setup` podem sugerir o comando, mas não o executam.

**Rationale**: Instalar skill altera sessões futuras do agente e escreve fora do workspace. Essa
autorização é diferente da autorização para criar, restaurar ou preparar um workspace.

**Alternatives considered**: Instalar automaticamente durante `init --agent` ou `workflow setup
--agent`, rejeitado porque reutiliza uma autorização local para efeito global; perguntar
interativamente nesses comandos, rejeitado porque quebra automação e ainda mistura responsabilidades.

## Decisão 2: Agentes públicos iniciais são somente Codex e Claude

**Decision**: Aceitar exatamente `codex` e `claude`; rejeitar `generic` e qualquer outro valor com
status 2, stderr e stdout vazio.

**Rationale**: Os requisitos iniciais cobrem dois adaptadores reais. `generic` seria uma promessa de
contrato sem destino de instalação, empacotamento e invocação definidos.

**Alternatives considered**: Expor `generic` como alias experimental, rejeitado porque aumentaria o
contrato público; criar registry configurável, rejeitado por não haver terceiro agente no escopo.

## Decisão 3: Pacote companheiro local/cacheado validado antes da cópia

**Decision**: O instalador consome um pacote versionado `cerne-skills` entregue pela
distribuição/instalador do `cerne-cli` como artefato companheiro em cache ou diretório gerenciado
pelo CLI. A instalação real não acessa rede nesta versão e valida manifesto, skill
`cerne-context`, adaptador do agente e `contextSchema` antes de copiar qualquer arquivo para o
destino final.

**Rationale**: A skill deve ser atualizável e auditável sem duplicar assets no `cerne-cli`. A
validação antecipada evita destino parcial e impede pacote incompatível. O artefato companheiro
mantém o comando simples, offline e sem flag pública extra.

**Alternatives considered**: Embutir arquivos da skill no binário do CLI, rejeitado porque acopla
releases do CLI e do repositório de skills; baixar sempre do GitHub em teste, rejeitado porque a
suíte precisa ser offline e determinística; exigir `--package <path>`, rejeitado porque transfere
ao usuário uma decisão que pertence à distribuição do CLI; procurar checkout irmão `../cerne-skills`,
rejeitado porque depende de layout de desenvolvimento.

## Decisão 4: Testes usam pacote local controlado

**Decision**: Automatizar cenários com fixture local de pacote e home temporária. Nenhum teste deve
depender de rede, GitHub, credenciais, releases reais ou diretórios reais de Codex/Claude.

**Rationale**: A constituição exige testes determinísticos e sem credenciais reais. Fixtures também
permitem simular manifesto inválido, schema incompatível, symlink e destino ocupado.

**Alternatives considered**: Teste end-to-end contra release pública, rejeitado por flake,
credenciais e acoplamento externo; mocks extensos de filesystem, rejeitados porque `t.TempDir()`
cobre melhor os efeitos observáveis com menos código.

## Decisão 5: Destino é perfil do usuário atual e nunca diretório administrativo

**Decision**: Resolver os destinos oficiais a partir da home do usuário atual:
`~/.codex/skills/cerne-context` para Codex e `~/.claude/skills/cerne-context` para Claude, usando
caminhos portáveis e sem privilégios administrativos. O comando pode rodar dentro ou fora de
workspace.

**Rationale**: Skills são configuração do agente no perfil do usuário, não artefato do workspace.
Essa escolha mantém knowledge/source intocados e evita escalada de permissão.

**Alternatives considered**: Instalação por workspace, rejeitada porque repete skill em cada
projeto; instalação do sistema, rejeitada por exigir permissão administrativa e afetar outros
usuários.

## Decisão 6: Overwrite seguro com instalação gerenciada

**Decision**: Escrever em staging, validar o pacote inteiro, recusar destino desconhecido, promover
somente quando seguro e tornar reinstalação da mesma versão idempotente. Instalação gerenciada em
versão diferente é atualizada automaticamente para a versão do pacote companheiro, substituindo
somente arquivos que o Cerne consiga provar que gerencia.

**Rationale**: A operação mexe em diretórios pessoais do usuário. A prova de ownership é a linha que
separa manutenção segura de sobrescrever trabalho alheio.

**Alternatives considered**: `rm -rf` do destino antes de copiar, rejeitado por risco de perda de
dados; sobrescrever arquivo a arquivo, rejeitado porque deixa instalação parcialmente misturada;
exigir `--upgrade` ou prompt interativo, rejeitado porque complica um fluxo scriptável sem aumentar
a segurança quando a instalação já é gerenciada.

## Decisão 7: Auditoria global privada e redigida

**Decision**: Registrar cada tentativa operacional em auditoria local privada com agente, pacote,
versão, destino, resultado e timestamps. Uso inválido não cria auditoria.

**Rationale**: A instalação modifica o perfil do usuário e deve ser reconstruível sem vazar conteúdo
da skill, ambiente, tokens, remotes ou saída externa bruta.

**Alternatives considered**: Registrar no workspace, rejeitado porque o comando pode rodar fora de
workspace e não deve alterar knowledge/source; não auditar, rejeitado pela constituição.

## Decisão 8: Contrato CLI segue padrões existentes

**Decision**: Sucesso usa stdout/status 0; falha operacional usa stderr/status 1; uso inválido usa
stderr/status 2 e stdout vazio. Diagnósticos permanecem em português com causa e correção segura.

**Rationale**: Mantém previsibilidade para scripts e compatibilidade com os comandos existentes do
Cerne.

**Alternatives considered**: Saída decorativa ou prompts interativos, rejeitados porque dificultam
automação e testes exatos.
