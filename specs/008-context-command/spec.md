# Feature Specification: Contexto Estrutural do Workspace

**Feature Branch**: `feat/context-command`

**Created**: 2026-08-06

**Status**: Draft

**Input**: User description: "Fornecer `cerne context` e `cerne context --json` para que pessoas e skills de Codex ou Claude descubram, de forma neutra e somente leitura, a estrutura validada de um workspace Cerne sem carregar o conteúdo do knowledge."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Entregar contexto estrutural para skills (Priority: P1)

Uma pessoa abre um workspace Cerne e invoca uma skill de IA. A skill solicita um contexto estável e
estruturado para localizar workspace, knowledge, source, coleções de conhecimento e workflow sem
interpretar o manifesto nem depender de mensagens textuais.

**Why this priority**: Este é o contrato que permite ao projeto separado `cerne-skills` carregar o
contexto mínimo correto para Codex e Claude sem duplicar regras do núcleo do Cerne.

**Independent Test**: Executar `cerne context --json` na raiz e em descendentes de workspaces
controlados, analisar a saída como JSON e confirmar paths canônicos, estados, problemas, códigos de
saída, estabilidade byte a byte e ausência de conteúdo dos repositórios.

**Acceptance Scenarios**:

1. **Given** um workspace válido sem workflow declarado, **When** uma skill solicita o contexto em
   formato estruturado, **Then** recebe schema 1, status saudável, paths normalizados do workspace,
   knowledge, coleções e source, workflow não declarado e lista vazia de problemas.
2. **Given** um workspace válido com Spec Kit ou OpenSpec, **When** uma skill solicita o contexto,
   **Then** recebe o provider, o estado estrutural e o path normalizado de especificações correto
   sem conhecer markers ou detalhes internos do provider.
3. **Given** um source dentro ou fora do workspace, **When** o contexto é produzido, **Then** a
   skill recebe seu path absoluto e uma indicação verificável sobre ele estar ou não dentro do
   workspace, sem inferência sobre ter sido clonado ou vinculado.
4. **Given** o mesmo workspace sem alterações, **When** o comando estruturado é repetido, **Then**
   as duas saídas são idênticas.

---

### User Story 2 - Inspecionar contexto no terminal (Priority: P2)

Uma pessoa quer confirmar rapidamente qual workspace está ativo, onde ficam knowledge e source e
qual é o estado do workflow antes de iniciar uma sessão de trabalho.

**Why this priority**: A saída humana torna o mesmo modelo observável e permite diagnosticar a
integração sem exigir que o usuário leia JSON ou conheça a estrutura do manifesto.

**Independent Test**: Executar `cerne context` nos mesmos fixtures e comparar stdout, stderr e
status com o contrato humano, incluindo source interno, source externo, workflow ausente, pendente
e pronto.

**Acceptance Scenarios**:

1. **Given** um workspace saudável, **When** o usuário consulta o contexto, **Then** vê nome,
   status, root, paths de knowledge e suas coleções, source, localização do source e workflow em
   português no stdout com status zero.
2. **Given** um workflow pendente, **When** o usuário consulta o contexto, **Then** recebe o
   relatório disponível, um aviso e uma correção segura sem bloquear a consulta.
3. **Given** a opção de ajuda, **When** o usuário a solicita, **Then** recebe finalidade, sintaxe,
   campos, streams, status, efeitos e exemplos sem inspecionar o workspace.

---

### User Story 3 - Diagnosticar contexto incompleto com segurança (Priority: P3)

Uma pessoa ou skill executa o comando em local incorreto ou em workspace parcialmente inválido e
precisa entender o que está disponível sem receber paths inventados, erros brutos ou correções
automáticas.

**Why this priority**: Contexto incorreto pode levar um agente a ler ou modificar o repositório
errado; falhas precisam ser explícitas, estruturadas e livres de efeitos colaterais.

**Independent Test**: Exercitar ausência de workspace, manifesto inválido, versão não suportada,
coleções ausentes, source inseguro ou inexistente e workflow pendente, parcial ou desconhecido;
verificar contexto parcial, catálogo fechado de problemas, códigos de saída e snapshots idênticos.

**Acceptance Scenarios**:

1. **Given** uma invocação fora de qualquer workspace, **When** o formato estruturado é solicitado,
   **Then** stdout contém JSON válido com status inválido e `workspace-not-found`, status do processo
   é um e nenhum artefato é criado.
2. **Given** um workspace identificável com source inválido, **When** o contexto é solicitado,
   **Then** os objetos comprováveis são preservados na saída, source não é inventado, um problema
   bloqueante é listado e o status é um.
3. **Given** somente problemas não bloqueantes, **When** o contexto é solicitado, **Then** o status
   geral é aviso e o processo retorna zero.
4. **Given** uso inválido, **When** argumentos ou opções fora do contrato são informados, **Then**
   stdout fica vazio, stderr apresenta uso seguro e o status é dois.

### Edge Cases

- A invocação começa na raiz, em knowledge, em source interno ou em outro descendente do workspace.
- Um source externo é resolvido, mas a ferramenta de IA pode não possuir acesso a ele; o comando
  informa somente sua localização e não tenta detectar ou contornar permissões do agente/editor.
- O manifesto está ausente, malformado, possui conteúdo extra, nome inválido, versão não suportada,
  workflow inválido ou source vazio.
- Workspace, knowledge, source, manifesto ou diretório obrigatório contém symlink ou sobreposição
  estrutural insegura.
- O source usa path absoluto, relativo, espaços, Unicode, volume Windows ou diferenças de caixa
  relevantes ao sistema.
- Product, specs, decisions ou policies está ausente ou não é diretório regular.
- Spec Kit ou OpenSpec está pendente, pronto, parcial, contém Git aninhado ou tem provider
  desconhecido; disponibilidade do executável não participa do contexto.
- Mais de um problema ocorre simultaneamente e precisa aparecer em ordem determinística.
- O diretório atual deixa de estar acessível antes que o workspace seja localizado.
- Campos desconhecidos futuros aparecem no manifesto sem alterar o contrato atual.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: O sistema MUST fornecer exatamente `cerne context`, `cerne context --json` e
  `cerne context --help` na primeira versão, sem argumentos de seleção, formatos adicionais ou
  correção automática.
- **FR-002**: O sistema MUST localizar o workspace ancestral mais próximo a partir da raiz ou de
  qualquer descendente do workspace.
- **FR-003**: A descoberta inicial MUST NOT localizar reversamente um workspace quando a invocação
  começar dentro de um source externo.
- **FR-004**: O contexto MUST resolver somente fatos verificáveis de workspace, knowledge, source,
  coleções de conhecimento e workflow, sem ler ou resumir conteúdo dessas coleções.
- **FR-005**: O sistema MUST produzir paths absolutos, canônicos e no formato nativo do sistema
  operacional.
- **FR-006**: O contexto MUST informar nome e root do workspace; path de knowledge; paths de
  product, specs, decisions e policies; path do source; relação interna/externa do source; declaração,
  provider e estado estrutural do workflow quando disponíveis.
- **FR-007**: O sistema MUST NOT afirmar se o source foi clonado ou vinculado, pois esse fato não é
  persistido pelo manifesto atual.
- **FR-008**: O sistema MUST normalizar o path de especificações como `knowledge/specs` sem
  workflow ou com Spec Kit e como `knowledge/openspec/specs` com OpenSpec.
- **FR-009**: Provider desconhecido MUST omitir `specs_path`, produzir problema bloqueante e tornar
  o contexto inválido.
- **FR-010**: O workflow MUST possuir exatamente os estados públicos `not-declared`, `pending`,
  `ready`, `invalid` e `unknown-provider`.
- **FR-011**: O sistema MUST avaliar manifesto, versão, nome, paths, symlinks, sobreposições,
  os diretórios obrigatórios `product`, `specs`, `decisions`, `policies` e `runs` e a estrutura
  declarada do workflow sem executar processo externo; `runs` participa da validação, mas não é
  exposto como contexto de agente.
- **FR-012**: O comando MUST NOT verificar executável de provider, validade Git, branch, remoto,
  histórico ou estado do working tree e MUST NOT substituir `cerne doctor`.
- **FR-013**: O status geral MUST ser `invalid` quando existir qualquer problema de severidade
  `error`, `warning` quando existirem somente warnings e `healthy` quando não existirem problemas.
- **FR-014**: Workflow pendente MUST ser um warning não bloqueante; workflow parcial, inválido ou
  desconhecido MUST ser bloqueante.
- **FR-015**: O catálogo inicial MUST limitar códigos públicos a `workspace-not-found`,
  `manifest-invalid`, `manifest-version-unsupported`, `knowledge-invalid`, `source-invalid`,
  `required-directory-invalid`, `workflow-pending`, `workflow-invalid` e
  `workflow-unknown-provider`.
- **FR-016**: Cada problema estruturado MUST conter código estável, severidade `warning` ou `error`
  e componente afetado, sem erro bruto do sistema ou mensagem localizada.
- **FR-017**: O formato estruturado MUST usar `schema_version` inteiro igual a 1, campos em ordem
  estável, problemas em ordem estável, indentação estável e newline final.
- **FR-018**: Mudanças compatíveis do schema 1 MUST ser somente aditivas; remoção, renomeação ou
  mudança semântica de campo existente MUST exigir uma nova versão de schema.
- **FR-019**: Objetos que não puderem ser comprovados MUST ser omitidos; `problems` MUST sempre
  existir como array, inclusive quando vazio.
- **FR-020**: O formato estruturado MUST NOT conter timestamp, versão do binário, origem ou remoto
  Git, environment, credencial, conteúdo de repositório, prompt ou instrução específica de IA.
- **FR-021**: A saída humana MUST apresentar em português os mesmos fatos disponíveis, além de
  causa e correção seguras para problemas, sem conteúdo decorativo que dificulte automação.
- **FR-022**: Relatórios humano e estruturado MUST usar stdout; contexto saudável ou com warnings
  MUST retornar zero; contexto inválido MUST retornar um; uso inválido MUST usar stderr e retornar
  dois.
- **FR-023**: O formato estruturado MUST produzir JSON válido em stdout com status um mesmo quando
  nenhum workspace for localizado, permitindo que consumidores tratem o problema pelo código.
- **FR-024**: A ajuda MUST usar stdout e status zero sem localizar, validar ou acessar arquivos do
  workspace.
- **FR-025**: O comando MUST ser estritamente somente leitura e MUST NOT criar audit, arquivo de
  instrução, cache, índice ou outro artefato; modificar manifesto; executar Git, workflow ou agente;
  acessar rede; instalar skill; ou corrigir o workspace.
- **FR-026**: O núcleo MUST NOT procurar, interpretar ou validar `AGENTS.md`, `CLAUDE.md` ou outro
  arquivo específico de agente.
- **FR-027**: O núcleo MUST NOT escolher uma especificação ativa; deve fornecer somente o
  `specs_path` normalizado para que a integração selecione contexto conforme a tarefa.
- **FR-028**: Repetições sobre estado inalterado MUST produzir saída estruturada byte a byte
  idêntica.
- **FR-029**: Workspaces e manifestos existentes MUST continuar válidos sem migração, novo campo ou
  alteração automática.
- **FR-030**: O comando e seus diagnósticos MUST possuir comportamento funcional equivalente em
  Linux, Windows e macOS.

### Constitutional Requirements *(include when the feature affects these concerns)*

- **Ownership/Repositories**: Knowledge e source permanecem separados e sob controle do usuário;
  a consulta não copia conteúdo, persiste contexto nem altera nenhum repositório.
- **AI/Integrations**: O núcleo entrega apenas fatos neutros. Descoberta de arquivos específicos,
  seleção progressiva e introdução de contexto pertencem aos adaptadores externos de Codex e
  Claude e não contaminam o domínio.
- **Context/Audit**: Nenhum conteúdo de knowledge é selecionado ou enviado pelo comando. Como a
  operação é uma consulta local sem ação automatizada ou sensível, ela não cria audit; consumidores
  posteriores continuam responsáveis pela rastreabilidade de ações que executarem.
- **Authorization/Secrets**: A consulta explícita autoriza somente leitura estrutural local. Paths
  finais podem ser exibidos localmente, mas origens, remotos, ambiente, credenciais e conteúdo são
  proibidos.
- **Portability**: Descoberta ancestral, canonicalização, contenção, symlinks, volumes, caixa,
  streams e serialização devem ser equivalentes nos três sistemas suportados usando paths nativos.
- **CLI/Compatibility**: O novo comando é uma adição compatível; ajuda, saída humana, schema JSON,
  códigos e ordem determinística são contratos públicos documentados e protegidos por testes.

### Key Entities *(include if feature involves data)*

- **Context Report**: Retrato efêmero e somente leitura dos fatos estruturais comprovados, com
  schema, status, entidades disponíveis e problemas ordenados.
- **Workspace Context**: Nome e root canônico do workspace ancestral identificado.
- **Knowledge Context**: Path canônico de knowledge e das coleções product, specs, decisions e
  policies, normalizadas independentemente do provider.
- **Source Context**: Path canônico do source e relação verificável de contenção com o workspace,
  sem origem ou modo histórico inferido.
- **Workflow Context**: Declaração, provider conhecido e estado estrutural, sem disponibilidade de
  executável ou execução de setup.
- **Context Problem**: Código público fechado, severidade e componente que explicam por que o
  relatório está saudável, com aviso ou inválido.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Em 100% dos fixtures válidos, pessoas e skills obtêm todos os paths estruturais
  corretos em uma única invocação, a partir da raiz ou de qualquer descendente interno.
- **SC-002**: Em 100% das repetições sobre filesystem inalterado, o formato estruturado é byte a
  byte idêntico e analisável como um único documento JSON válido.
- **SC-003**: Em 100% dos fixtures inválidos ou parciais, nenhum objeto não comprovado é inventado,
  todos os problemas aplicáveis usam o catálogo público e o status/código de saída corresponde à
  severidade observada.
- **SC-004**: Em 100% das consultas, snapshots de workspace, knowledge, source e estado local do
  usuário permanecem idênticos antes e depois da execução.
- **SC-005**: Nenhuma saída humana ou estruturada contém origem Git, remoto, environment,
  credencial, conteúdo de repositório ou instrução específica de Codex/Claude.
- **SC-006**: Workspaces sem workflow, com Spec Kit e com OpenSpec apresentam o path de
  especificações e estado esperados em todos os cenários suportados.
- **SC-007**: Os mesmos cenários automatizados produzem resultados funcionais equivalentes em
  Linux, Windows e macOS, respeitando a representação nativa de paths.
- **SC-008**: Em uma validação manual com o manifesto fechado e o cronômetro iniciado somente após
  a saída humana terminar de ser exibida, uma pessoa consegue identificar corretamente workspace,
  knowledge, source, relação interna/externa do source e workflow em menos de 10 segundos.
- **SC-009**: Uma skill consegue distinguir contexto saudável, warning e inválido exclusivamente
  pelo schema e códigos públicos, sem analisar mensagens localizadas.
- **SC-010**: Todos os comandos e workspaces existentes mantêm seus contratos e não exigem
  migração após a adição da feature.

## Assumptions

- O usuário invoca o comando dentro da árvore física do workspace; descoberta reversa a partir de
  source externo pertence a uma feature futura.
- O manifesto atual continua sendo a autoridade para nome, source e workflow e não recebe campos
  novos para esta feature.
- Paths absolutos são dados locais necessários à integração e não são persistidos ou enviados pelo
  Cerne.
- A skill possui responsabilidade própria para verificar se o agente/editor pode acessar um source
  externo e para solicitar ação do usuário quando não puder.
- `cerne-skills` define adaptadores de Codex e Claude, mas seu empacotamento, instalação e
  comportamento ficam fora desta feature.
- Seleção de spec ativa, leitura progressiva, resolução de conflitos entre knowledge e instruções
  locais e auditoria de ações posteriores pertencem às skills.
- Disponibilidade do Git e dos executáveis Spec Kit/OpenSpec não é necessária para produzir o
  contexto estrutural.
