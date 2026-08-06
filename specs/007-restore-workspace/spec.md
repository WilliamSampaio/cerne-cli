# Feature Specification: Restauração de Workspace

**Feature Branch**: `feat/restore-workspace`

**Created**: 2026-08-05

**Status**: Draft

**Input**: User description: "Restaurar em uma máquina nova um workspace Cerne a partir de um
repositório knowledge remoto ou local e de um source local ou clonado, sem persistir as origens no
manifesto. O nome do workspace deve vir do `cerne.json`."

## Clarifications

### Session 2026-08-05

- Q: Como informar e materializar o knowledge remoto ou local? → A: Uma única origem Git aceita URL ou path local e sempre é materializada por clone.
- Q: Onde registrar a auditoria antes de o knowledge existir? → A: Em diretório privado global do usuário, sob `~/.cerne/audit` ou equivalente portável.
- Q: O que fazer se o source falhar depois que o knowledge for obtido? → A: Reverter integralmente os artefatos do workspace e deixar somente a auditoria global.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Restaurar knowledge e clonar source (Priority: P1)

Um usuário em uma máquina nova quer informar explicitamente as origens Git do knowledge e do
source e obter um workspace completo, com os dois repositórios independentes e o mesmo nome
declarado no manifesto.

**Why this priority**: Este é o fluxo principal de recuperação e elimina a preparação manual da
estrutura sem introduzir origens persistentes ou integrações específicas de host.

**Independent Test**: Usar dois repositórios Git locais temporários como origens controladas,
restaurar o workspace em um diretório vazio e verificar nome, histórico, remotos, manifesto,
separação, saída, auditoria e ausência de alteração nas origens.

**Acceptance Scenarios**:

1. **Given** um knowledge Git válido com manifesto suportado e uma origem de source válida,
   **When** o usuário solicita a restauração com ambas as origens, **Then** o sistema cria um
   workspace com o nome do manifesto, knowledge restaurado e source clonado.
2. **Given** origens válidas, **When** a restauração termina, **Then** os dois repositórios mantêm
   históricos, remotos e ciclos de vida independentes.
3. **Given** um manifesto que declara workflow, **When** a restauração termina, **Then** a
   declaração é preservada sem executar automaticamente o provider.

---

### User Story 2 - Reutilizar source local existente (Priority: P2)

Um usuário que já possui o source na nova máquina quer restaurar apenas o knowledge e associar o
working tree existente sem copiá-lo ou modificá-lo.

**Why this priority**: Evita clone redundante e reutiliza as garantias atuais de associação de
source local.

**Independent Test**: Restaurar um knowledge controlado com um source local temporário, comparar o
source byte a byte antes e depois e verificar que o manifesto resultante resolve o repositório
informado.

**Acceptance Scenarios**:

1. **Given** um source local Git válido, **When** ele é escolhido na restauração, **Then** o
   workspace o referencia sem copiar, mover ou modificar seu conteúdo e metadados.
2. **Given** que o caminho local difere do caminho registrado no manifesto restaurado, **When** a
   associação é autorizada, **Then** somente a referência do manifesto restaurado é atualizada e
   essa alteração é informada ao usuário.

---

### User Story 3 - Recusar restaurações inseguras ou incompletas (Priority: P3)

Um usuário precisa confiar que conteúdo controlado por uma origem Git não escolherá destinos,
conexões ou operações adicionais fora do que foi explicitamente solicitado.

**Why this priority**: A restauração combina duas entradas externas e um nome vindo do manifesto;
falhas ou dados maliciosos não podem causar sobrescrita, vazamento de segredo ou mistura dos
repositórios.

**Independent Test**: Exercitar nomes inseguros, manifesto inválido, credenciais embutidas,
transportes recusados, destinos concorrentes, repositórios parciais e falhas em cada etapa,
verificando efeitos e registros persistentes.

**Acceptance Scenarios**:

1. **Given** um manifesto com nome inválido ou caminho inseguro, **When** ele é obtido, **Then** a
   restauração falha sem criar ou substituir o destino final.
2. **Given** um destino final existente e não vazio, **When** a restauração é solicitada, **Then**
   nenhum conteúdo preexistente é alterado ou removido.
3. **Given** uma origem com credenciais embutidas ou transporte recusado, **When** a restauração é
   solicitada, **Then** ela falha antes de contactar a origem e não exibe o valor sensível.
4. **Given** qualquer tentativa de restauração, **When** ela inicia ou falha, **Then** existe um
   único registro privado no diretório global de auditoria e nenhum workspace parcial permanece.

### Edge Cases

- O knowledge é válido, mas não contém `cerne.json`, ou o manifesto está malformado.
- O `name` contém travessia, separador, nome reservado, Unicode incompatível ou diverge do nome
  esperado pelo usuário.
- Knowledge ou source são repositórios vazios, bare, worktrees, subpastas Git ou possuem symlinks.
- As duas origens apontam para o mesmo repositório ou criam aninhamento entre knowledge e source.
- O destino aparece depois da validação e antes da conclusão.
- Uma clonagem é interrompida depois de criar conteúdo parcial.
- O diretório global de auditoria está ausente, inacessível, é link simbólico ou não permite
  finalizar o registro.
- O manifesto declara source absoluto específico de outra máquina.
- A origem local do knowledge possui arquivos não versionados, segredos ou alterações pendentes.
- O workflow declarado está pronto, pendente, parcial ou depende de executável ausente.
- O caminho da invocação contém espaços, Unicode, aliases ou diferenças de caixa.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: O sistema MUST fornecer `cerne restore <knowledge-origin>` com exatamente uma opção
  `--source <local-path>` ou `--clone <source-origin>`.
- **FR-002**: `<knowledge-origin>` MUST aceitar URL ou path de repositório Git local pelos mesmos
  transportes permitidos e MUST sempre materializar knowledge por clone, sem cópia recursiva do
  diretório informado.
- **FR-003**: A seleção de source local e a origem de clone MUST ser mutuamente exclusiva, exigir
  valor e aparecer explicitamente na invocação.
- **FR-004**: O sistema MUST obter o nome do workspace do manifesto restaurado, validar todas as
  regras portáteis de nome antes de escolher o destino final e MUST NOT aceitar travessia ou path.
- **FR-005**: A restauração MUST recusar destino final inseguro, link simbólico ou diretório não
  vazio e MUST NOT substituir conteúdo preexistente ou concorrente.
- **FR-006**: Origens Git MUST ser fornecidas somente na invocação e MUST NOT ser persistidas no
  manifesto, auditoria, diagnóstico ou saída integral.
- **FR-007**: As regras vigentes para transportes permitidos, credenciais embutidas, interatividade,
  clone padrão e redaction MUST valer igualmente para knowledge e source.
- **FR-008**: Um source local MUST ser working tree Git non-bare, informado pela raiz, separado do
  knowledge e preservado byte a byte.
- **FR-009**: Quando um source local autorizado divergir da referência restaurada, o sistema MUST
  atualizar somente o campo `source`, preservar os demais campos e informar a alteração.
- **FR-010**: Um source clonado MUST ocupar o caminho portátil declarado pelo manifesto; referência
  absoluta ou externa incompatível MUST ser recusada com correção segura.
- **FR-011**: Knowledge e source restaurados MUST permanecer repositórios Git independentes, sem
  repositório Git na raiz do workspace e sem cópia entre eles.
- **FR-012**: A restauração MUST validar manifesto, versão, workflow, estrutura obrigatória,
  repositórios e separação antes de comunicar sucesso.
- **FR-013**: A restauração MUST NOT executar workflow, agente, push, fetch adicional, submódulo,
  merge, publicação, instalação ou deploy.
- **FR-014**: Cada tentativa MUST criar, antes de qualquer clone, um registro único no diretório
  privado de auditoria do usuário, localizado sob `~/.cerne/audit` ou equivalente portável, e MUST
  registrar separadamente as fases knowledge e source sem armazenar origens ou saída integral.
- **FR-015**: Falha ao criar ou iniciar o registro global MUST impedir qualquer processo externo;
  falha ao finalizá-lo MUST tornar a operação malsucedida, reverter o workspace e preservar o
  registro inconclusivo quando ele já existir.
- **FR-016**: Qualquer falha anterior à conclusão integral MUST remover somente os artefatos de
  workspace comprovadamente criados pela tentativa e deixar o destino final ausente, preservando
  apenas a auditoria global.
- **FR-017**: A limpeza automática MUST limitar-se a artefatos novos cuja propriedade pela tentativa
  seja demonstrável e MUST recusar alvos ambíguos.
- **FR-018**: Sucesso MUST usar stdout e status zero; falha operacional MUST usar stderr e status
  um; uso inválido MUST usar stderr e status dois.
- **FR-019**: A saída MUST identificar nome e caminhos finais sem reproduzir origens potencialmente
  sensíveis.
- **FR-020**: O comando MUST possuir comportamento funcional equivalente em Linux, Windows e macOS.
- **FR-021**: Ajuda e documentação MUST explicar sintaxe, autorizações, rede, autenticação,
  efeitos, auditoria, falhas, retomada, streams, status e exemplos.
- **FR-022**: Workspaces e manifestos existentes MUST continuar válidos sem migração e comandos
  anteriores MUST manter seus contratos.
- **FR-023**: Testes MUST usar repositórios temporários e processos controlados, sem rede,
  credenciais ou remotos reais.

### Constitutional Requirements *(include when the feature affects these concerns)*

- **Ownership/Repositories**: O usuário escolhe explicitamente ambas as entradas; knowledge e
  source permanecem independentes e nenhuma restauração copia conteúdo entre eles.
- **AI/Integrations**: A feature não exige host, provider de IA ou agente; integrações externas
  continuam opcionais e limitadas por contratos substituíveis.
- **Context/Audit**: Cada tentativa possui registro privado no estado local do usuário, iniciado
  antes do primeiro processo e suficiente para reconstruir as fases executadas sem armazenar
  origem ou output integral; o registro sobrevive ao rollback do workspace.
- **Authorization/Secrets**: Cada origem é autorizada explicitamente na invocação; credenciais
  permanecem externas e nunca entram em repositório, auditoria ou diagnóstico.
- **Portability**: Nome, paths, aliases, processo, promoção, limpeza e saídas possuem comportamento
  equivalente nos três sistemas suportados.
- **CLI/Compatibility**: O novo comando é adição compatível; nenhum manifesto existente muda e o
  contrato completo será documentado e protegido por testes de integração.

### Key Entities *(include if feature involves data)*

- **Knowledge Origin**: URL ou localização local explicitamente autorizada para obter o
  repositório de conhecimento; nunca persistida pelo Cerne.
- **Source Selection**: Escolha exclusiva entre working tree local existente e origem Git a ser
  clonada.
- **Restore Manifest**: Manifesto obtido do knowledge que determina nome, referência local do
  source e workflow declarado, sem determinar conexões externas.
- **Restore Attempt**: Registro global e privado que acompanha obtenção, validação,
  materialização, fases knowledge/source e resultado sem persistir valores sensíveis.
- **Restored Workspace**: Raiz promovida somente quando satisfaz nome, manifesto, repositórios,
  separação e seleção de source autorizada.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Em 100% dos cenários válidos, o usuário restaura knowledge e source em uma única
  invocação, sem editar manualmente o manifesto.
- **SC-002**: Em 100% dos fixtures com source local, o repositório informado permanece byte a byte
  idêntico após sucesso ou falha.
- **SC-003**: Em 100% dos fixtures clonados, knowledge e source preservam histórico e remoto
  esperados e permanecem Git independentes.
- **SC-004**: 100% dos nomes, destinos, manifestos, origens e estruturas inseguros são recusados sem
  alteração de conteúdo preexistente.
- **SC-005**: Nenhum cenário persiste ou exibe origem integral, credencial, segredo ou saída externa
  integral.
- **SC-006**: Cada tentativa deixa exatamente um registro global auditável, final ou inconclusivo,
  criado antes de qualquer processo e preservado mesmo quando o workspace é revertido.
- **SC-007**: Os mesmos cenários automatizados produzem resultados funcionais equivalentes em
  Linux, Windows e macOS.
- **SC-008**: Workspaces anteriores continuam aprovados pelos comandos existentes sem migração ou
  alteração automática.

## Assumptions

- O usuário possui Git e credenciais externas adequadas para as origens autorizadas.
- O diretório global de auditoria pertence ao usuário, usa permissões privadas e não sofre remoção
  automática nesta versão; retenção e limpeza permanecem sob controle do usuário.
- O knowledge informado é um repositório Git; importar diretório comum sem histórico permanece
  fora do escopo.
- A restauração recupera o checkout padrão atual das origens; reprodução de commit, tag ou branch
  específica permanece fora do escopo.
- Múltiplos remotos, submódulos, Git LFS explícito, sincronização e atualização de workspaces
  existentes permanecem fora do escopo.
- O provider declarado, quando houver, será tratado posteriormente por `cerne workflow setup`.
