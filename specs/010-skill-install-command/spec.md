# Feature Specification: Instalação de Skills Cerne

**Feature Branch**: `010-skill-install-command`

**Created**: 2026-08-07

**Status**: Draft

**Input**: User description: "Adicionar `cerne skill install <codex|claude>` para instalar explicitamente as skills oficiais do Cerne no perfil do usuário atual. A instalação não deve acontecer automaticamente durante `init`, `restore` ou `workflow setup`."

## Clarifications

### Session 2026-08-07

- Q: Nesta primeira versão, de onde o `cerne skill install <agent>` deve obter o pacote oficial `cerne-skills` em uso real? → A: Pacote local/cacheado gerenciado pelo CLI, sem rede nesta primeira versão.
- Q: Quais destinos globais devem ser considerados oficiais para instalar a skill `cerne-context` nesta primeira versão? → A: Codex em `~/.codex/skills/cerne-context` e Claude em `~/.claude/skills/cerne-context`.
- Q: Como o pacote local/cacheado `cerne-skills` deve chegar até o CLI nesta primeira versão? → A: O instalador/distribuição do `cerne-cli` coloca um pacote `cerne-skills` versionado em cache/diretório gerenciado pelo CLI.
- Q: Quando já existir uma instalação gerenciada pelo Cerne em versão diferente, o comando `cerne skill install <agent>` deve atualizá-la para a versão do pacote companheiro? → A: Sim, atualizar automaticamente instalações gerenciadas para a versão do pacote companheiro.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Instalar skill Cerne para um agente suportado (Priority: P1)

Uma pessoa que usa Codex ou Claude quer instalar a skill oficial `cerne-context` uma vez no perfil
do usuário atual, para que o agente consiga carregar contexto Cerne em qualquer workspace.

**Why this priority**: A skill é o ponto de entrada para agentes usarem `cerne context --json` sem
duplicar regras do CLI. Sem instalação explícita, cada workspace dependeria de cópia manual.

**Independent Test**: Executar o comando com um pacote oficial controlado e confirmar que o diretório
de skill do agente escolhido é criado no perfil do usuário atual, que o manifesto é validado antes
da cópia e que a saída informa o destino instalado.

**Acceptance Scenarios**:

1. **Given** um pacote `cerne-skills` oficial e compatível, **When** o usuário executa
   `cerne skill install codex`, **Then** a skill `cerne-context` é instalada no destino de usuário
   atual do Codex e o comando retorna sucesso com destino e versão.
2. **Given** um pacote `cerne-skills` oficial e compatível, **When** o usuário executa
   `cerne skill install claude`, **Then** a skill `cerne-context` é instalada no destino de usuário
   atual do Claude e o comando retorna sucesso com destino e versão.
3. **Given** a skill já instalada e gerenciada pelo Cerne na mesma versão, **When** o usuário executa
   o mesmo comando novamente, **Then** o comando é idempotente e não regrava arquivos
   desnecessariamente.
4. **Given** a skill já instalada e gerenciada pelo Cerne em versão diferente, **When** o usuário
   executa o mesmo comando, **Then** o comando atualiza automaticamente a instalação para a versão
   do pacote companheiro preservando conteúdo não pertencente ao Cerne.

---

### User Story 2 - Evitar instalação implícita em fluxos de workspace (Priority: P1)

Uma pessoa cria, restaura ou prepara um workspace sem querer alterar o perfil global do usuário.
Esses comandos podem sugerir a instalação da skill, mas não podem executá-la sem pedido explícito.

**Why this priority**: Instalar skill escreve fora do workspace e afeta sessões futuras do agente.
Essa autorização é diferente de criar ou preparar um workspace.

**Independent Test**: Executar `init`, `restore` e `workflow setup` em cenários com `--agent` e
confirmar que nenhum destino de skill em perfil de usuário é criado ou alterado.

**Acceptance Scenarios**:

1. **Given** um usuário sem skill Cerne instalada, **When** executa `cerne init app --workflow
   speckit --agent codex`, **Then** o workspace e a ponte local podem ser preparados, mas nenhuma
   skill global é instalada.
2. **Given** um knowledge restaurado que declara workflow, **When** o usuário executa `cerne restore`
   ou `cerne workflow setup --agent claude`, **Then** nenhum arquivo em perfil global de agente é
   criado ou alterado.
3. **Given** um comando de workspace conclui e a skill não está instalada, **When** o CLI imprime
   orientação, **Then** a orientação pode sugerir `cerne skill install <agent>` sem executar o
   comando automaticamente.

---

### User Story 3 - Recusar instalações inseguras ou incompatíveis (Priority: P2)

Uma pessoa executa o instalador em um ambiente com pacote incompatível, destino existente
desconhecido ou agente não suportado. O Cerne deve recusar com diagnóstico claro e sem sobrescrever
arquivos do usuário.

**Why this priority**: O comando altera o perfil do usuário. Uma instalação errada pode quebrar
configurações pessoais do agente ou espalhar instruções incompatíveis.

**Independent Test**: Usar pacotes e destinos controlados para simular agente desconhecido,
manifesto inválido, schema incompatível, destino já ocupado por conteúdo não gerenciado e falha de
escrita, verificando stderr, status e preservação dos arquivos.

**Acceptance Scenarios**:

1. **Given** o usuário informa `generic` ou outro agente não suportado, **When** executa
   `cerne skill install generic`, **Then** o comando falha por uso inválido sem alterar arquivos.
2. **Given** o destino de instalação contém arquivos não gerenciados pelo Cerne, **When** o usuário
   instala a skill, **Then** o comando falha antes de sobrescrever e explica a correção segura.
3. **Given** o pacote requer um schema de contexto não suportado pelo CLI atual, **When** a
   instalação é solicitada, **Then** o comando falha antes de copiar arquivos.
4. **Given** o pacote companheiro oficial está ausente ou inacessível, **When** o usuário executa
   `cerne skill install codex`, **Then** o comando falha como erro operacional sem alterar o destino
   do agente.
5. **Given** o pacote contém manifesto malformado, arquivo fora do pacote ou link simbólico, **When**
   a instalação é solicitada, **Then** o comando falha antes de copiar arquivos e preserva qualquer
   destino existente.

### Edge Cases

- O agente informado está ausente, repetido, vazio, com caixa diferente ou possui argumentos extras.
- O pacote oficial não pode ser obtido, não é íntegro, não possui manifesto ou possui manifesto
  malformado.
- O pacote declara mais de uma skill, mas não declara `cerne-context`.
- O manifesto aponta para arquivo fora do pacote ou para link simbólico.
- O destino de instalação já existe com conteúdo parcial, desconhecido, link simbólico ou arquivo
  regular onde deveria haver diretório.
- O destino oficial do Codex é `~/.codex/skills/cerne-context` e o destino oficial do Claude é
  `~/.claude/skills/cerne-context`; se algum deles resolver para fora do perfil do usuário atual, a
  instalação deve falhar antes de copiar arquivos.
- O destino fica em perfil de usuário inacessível, somente leitura ou com permissões inseguras.
- A instalação é interrompida durante a cópia.
- O usuário executa o comando dentro ou fora de um workspace Cerne.
- O usuário está offline, o pacote companheiro oficial não foi instalado junto com o CLI, o cache
  local gerenciado não existe ou o pacote cacheado está indisponível.
- O sistema operacional resolve diretórios de perfil de forma diferente entre Linux, Windows e
  macOS.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: O CLI MUST fornecer `cerne skill install <agent>` como comando explícito para
  instalar skills oficiais do Cerne.
- **FR-002**: A primeira versão MUST aceitar exatamente `codex` e `claude` como agentes públicos.
- **FR-003**: O comando MUST rejeitar `generic` e qualquer outro agente não suportado com status 2,
  stderr e stdout vazio.
- **FR-004**: `cerne init`, `cerne restore` e `cerne workflow setup` MUST NOT instalar skills em
  perfil de usuário, mesmo quando `--agent codex|claude` for usado.
- **FR-005**: Comandos de workspace MAY sugerir `cerne skill install <agent>` quando útil, mas MUST
  NOT executar a instalação automaticamente.
- **FR-006**: A instalação MUST atingir somente o perfil do usuário atual e MUST NOT exigir
  privilégios administrativos ou alterar diretórios do sistema.
- **FR-007**: Nesta primeira versão, o instalador MUST obter um pacote oficial versionado de
  `cerne-skills` a partir de pacote local/cacheado gerenciado pelo CLI, sem acesso à rede em uso
  real, e validar seu manifesto antes de copiar arquivos.
- **FR-007a**: A distribuição ou instalador do `cerne-cli` MUST disponibilizar esse pacote
  `cerne-skills` versionado como artefato companheiro em cache ou diretório gerenciado pelo CLI.
- **FR-007b**: `cerne skill install` MUST falhar com status 1, sem alterar destinos de agente, quando
  o pacote companheiro oficial estiver ausente ou inacessível.
- **FR-008**: Testes automatizados MUST usar pacote controlado local e MUST NOT depender de rede,
  GitHub, credenciais ou releases reais.
- **FR-009**: O instalador MUST validar compatibilidade entre o `contextSchema` requerido pela skill
  e o schema de `cerne context --json` suportado pelo CLI atual.
- **FR-010**: O instalador MUST instalar a skill `cerne-context` em
  `~/.codex/skills/cerne-context` para `codex` e em `~/.claude/skills/cerne-context` para `claude`,
  sempre resolvidos dentro do perfil do usuário atual.
- **FR-011**: O instalador MUST copiar somente arquivos regulares e diretórios pertencentes ao
  pacote validado e MUST rejeitar links simbólicos ou paths que escapem do pacote.
- **FR-012**: O instalador MUST NOT sobrescrever conteúdo desconhecido no destino.
- **FR-013**: Reinstalar a mesma versão gerenciada pelo Cerne MUST ser idempotente.
- **FR-014**: Atualizar uma instalação gerenciada existente em versão diferente MUST substituir
  automaticamente somente arquivos que o Cerne consiga provar que pertencem à instalação anterior,
  usando a versão do pacote companheiro.
- **FR-015**: Falha antes da promoção final MUST preservar a instalação anterior quando existir e
  MUST NOT deixar destino parcial marcado como pronto.
- **FR-016**: Cada instalação que tenta modificar o perfil do usuário MUST produzir auditoria local
  privada com agente, pacote, versão, destino, resultado e timestamps, sem registrar conteúdo da
  skill, variáveis de ambiente, tokens, remotes ou saída externa bruta.
- **FR-017**: Sucesso MUST usar stdout/status 0 e informar agente, skill, versão e destino.
- **FR-018**: Falhas operacionais MUST usar stderr/status 1 com causa e correção seguras.
- **FR-019**: Uso inválido MUST usar stderr/status 2 e não criar auditoria nem alterar arquivos.
- **FR-020**: A ajuda e a documentação MUST descrever sintaxe, agentes suportados, destino,
  autorização, efeitos colaterais, códigos de saída, segurança e a ausência de instalação
  automática em `init`, `restore` e `workflow setup`.

### Constitutional Requirements *(include when the feature affects these concerns)*

- **Ownership/Repositories**: A instalação não modifica workspaces, knowledge ou source; ela escreve
  apenas no perfil do usuário atual.
- **AI/Integrations**: Codex e Claude entram por adaptadores explícitos. `generic` não é alvo
  público nesta versão.
- **Context/Audit**: A auditoria registra metadados mínimos da instalação sem copiar conteúdo de
  skills ou conhecimento privado.
- **Authorization/Secrets**: `cerne skill install <agent>` é a autorização explícita para escrever
  no perfil daquele agente; nenhum outro comando concede essa autorização por inferência.
- **Portability**: Destinos de perfil e promoção atômica devem funcionar em Linux, Windows e macOS.
- **CLI/Compatibility**: O novo comando é aditivo; contratos existentes de `init`, `restore`,
  `workflow setup` e `context` permanecem compatíveis.

### Key Entities *(include if feature involves data)*

- **Skill Package**: Artefato oficial versionado de `cerne-skills`, contendo manifesto,
  diretórios de skill e metadados de adaptador.
- **Skill Manifest**: Arquivo do pacote que declara skills, versão, adaptadores e compatibilidade de
  schema; os destinos oficiais de instalação ficam definidos pelo CLI nesta versão.
- **Agent Install Target**: Destino oficial no perfil do usuário atual para um agente suportado:
  `~/.codex/skills/cerne-context` para Codex e `~/.claude/skills/cerne-context` para Claude.
- **Managed Installation**: Instalação cuja origem, versão e arquivos são reconhecidos pelo Cerne por
  marcador privado contendo agente, skill, versão do pacote e lista de paths relativos gerenciados.
- **Skill Install Attempt**: Registro auditável local de uma tentativa de instalação.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Em 100% dos cenários válidos com pacote compatível, `cerne skill install codex` e
  `cerne skill install claude` instalam `cerne-context` no destino esperado e retornam status 0.
- **SC-002**: Em 100% dos cenários de `init`, `restore` e `workflow setup`, nenhum arquivo de skill
  global é criado ou alterado sem `cerne skill install`.
- **SC-003**: Em 100% dos cenários com agente inválido, manifesto incompatível ou destino
  desconhecido, arquivos preexistentes permanecem byte a byte inalterados.
- **SC-004**: Cada tentativa operacional de instalação deixa exatamente um registro auditável local
  sem segredos e sem conteúdo integral da skill.
- **SC-005**: A suíte automatizada cobre instalação válida, reinstalação idempotente, atualização de
  instalação gerenciada, agente inválido, pacote inválido, schema incompatível, destino ocupado,
  rollback e ausência de instalação automática em comandos de workspace.

## Assumptions

- `cerne-skills` é o pacote oficial inicial e contém a skill `cerne-context`.
- O CLI atual suporta `cerne context --json` schema v1.
- A instalação inicial não adiciona suporte a outros agentes, registry dinâmico, atualização
  automática remota/de releases ou instalação administrativa.
- A origem oficial versionada será disponibilizada pela distribuição do `cerne-cli` como pacote
  companheiro local/cacheado gerenciado pelo CLI; testes usam pacote local controlado para manter a
  suíte determinística e sem rede.
