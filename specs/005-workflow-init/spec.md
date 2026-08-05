# Feature Specification: Inicialização com Workflow SDD

**Feature Branch**: `feat/workflow-init`

**Created**: 2026-08-04

**Status**: Draft

**Input**: User description: "Permitir que `cerne init` selecione Spec Kit ou OpenSpec, adapte a
estrutura de knowledge ao workflow, registre a escolha no manifesto e trate a ausência da
ferramenta como aviso recuperável, sem instalar dependências."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Criar workspace com workflow escolhido (Priority: P1)

Um usuário quer criar um workspace Cerne já preparado para Spec Kit ou OpenSpec, mantendo o
conhecimento separado do código-fonte e sem precisar conhecer os detalhes do comando de
inicialização da ferramenta escolhida.

**Why this priority**: Esta é a entrega principal da feature: transformar a escolha explícita do
workflow em uma estrutura local utilizável e registrada.

**Independent Test**: Disponibilizar um substituto local determinístico para cada ferramenta,
executar `cerne init exemplo --workflow <provider>` e verificar manifesto, estrutura condicional,
invocação do provider, repositórios Git independentes, saída e registro auditável.

**Acceptance Scenarios**:

1. **Given** que Spec Kit está disponível, **When** o usuário executa `cerne init exemplo --workflow speckit`, **Then** o workspace é criado, o manifesto registra `speckit`, a estrutura comum do knowledge é preservada e os artefatos do Spec Kit são inicializados dentro de knowledge.
2. **Given** que OpenSpec está disponível, **When** o usuário executa `cerne init exemplo --workflow openspec`, **Then** o workspace é criado, o manifesto registra `openspec`, a estrutura comum do knowledge é preservada e os artefatos do OpenSpec são inicializados dentro de knowledge.
3. **Given** que nenhuma opção de workflow é informada, **When** o usuário executa `cerne init exemplo`, **Then** o comando mantém a sintaxe, estrutura, manifesto, saídas, status e ausência de execução externa do comportamento anterior.
4. **Given** um workflow selecionado, **When** o bootstrap é executado, **Then** ele é não interativo, não configura um agente de IA específico em nome do usuário e não instala nem atualiza ferramentas.

---

### User Story 2 - Concluir init quando a ferramenta estiver ausente (Priority: P2)

Um usuário quer preservar o workspace criado e a escolha do workflow mesmo quando ainda não
instalou a ferramenta correspondente.

**Why this priority**: A ausência de uma dependência opcional não deve impedir o usuário de começar,
mas a diferença entre intenção e estrutura materializada precisa ser explícita e recuperável.

**Independent Test**: Executar o init com PATH controlado sem o provider, verificar status zero,
workspace válido, preferência persistida, aviso em stderr e orientação para retomar o bootstrap.

**Acceptance Scenarios**:

1. **Given** que o provider selecionado não está disponível, **When** o usuário executa o init com workflow, **Then** o workspace base é criado, a preferência é registrada, o comando retorna sucesso e stderr informa que o workflow permanece pendente.
2. **Given** um workspace com workflow pendente e a ferramenta posteriormente disponível, **When** o usuário executa `cerne workflow setup`, **Then** o bootstrap é concluído sem recriar o workspace nem alterar o source.
3. **Given** um workspace com workflow já inicializado, **When** o usuário executa `cerne workflow setup`, **Then** o comando conclui sem duplicar ou substituir artefatos e informa que nenhuma ação é necessária.
4. **Given** um workspace com workflow pendente, **When** o usuário executa `cerne doctor`, **Then** o relatório mantém o workspace utilizável e apresenta aviso não bloqueante com a correção.

---

### User Story 3 - Receber falhas seguras e previsíveis (Priority: P3)

Um usuário precisa de diagnósticos claros quando a opção, o manifesto, o workspace ou a execução
do provider é inválida, sem deixar estruturas parciais ou misturar conhecimento e código.

**Why this priority**: A feature executa um processo externo que pode criar vários arquivos; falhas
precisam preservar as garantias transacionais e de separação existentes.

**Independent Test**: Usar providers locais que retornam falha ou criam artefatos parciais e
verificar códigos, streams, rollback, auditoria, ausência de segredos e preservação do source.

**Acceptance Scenarios**:

1. **Given** um valor de workflow desconhecido, **When** o usuário executa o init, **Then** o comando retorna erro de uso antes de criar arquivos e lista os valores aceitos.
2. **Given** que o provider está disponível mas falha durante `cerne init`, **When** o comando termina, **Then** o comando retorna falha operacional, preserva o workspace base e a auditoria, remove somente artefatos novos atribuíveis ao provider e deixa o workflow pendente para retomada.
3. **Given** que o provider falha durante `cerne workflow setup`, **When** o comando termina, **Then** o workspace base e arquivos anteriores permanecem intactos, artefatos parciais da tentativa não são aceitos como configuração concluída e o erro orienta a correção.
4. **Given** um manifesto sem workflow configurado, **When** o usuário executa `cerne workflow setup`, **Then** o comando recusa a operação sem modificar o workspace e explica como selecionar um workflow em um novo workspace.

### Edge Cases

- `--workflow` é informado sem valor, repetido, em posição inválida ou acompanhado de argumentos extras.
- O nome do executável correto existe, mas não é executável ou não pode ser iniciado.
- O provider solicita interação apesar da invocação não interativa ou não oferece a capacidade neutra esperada.
- O provider cria parte da estrutura e encerra com falha.
- O provider retorna sucesso sem produzir a estrutura mínima esperada.
- A estrutura esperada já existe integral ou parcialmente antes de `cerne workflow setup`.
- O manifesto possui workflow desconhecido, malformado ou incompatível com sua versão.
- O usuário remove o executável depois de concluir o bootstrap.
- O caminho do workspace contém espaços ou Unicode.
- Arquivos produzidos pelo provider contêm referência ao source ou tentam criar Git aninhado.
- Um diagnóstico ou registro externo inclui texto que se parece com segredo.
- A inicialização ocorre em Linux, Windows ou macOS com nomes de executáveis e regras de processo diferentes.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: O sistema MUST aceitar `cerne init <project-name> --workflow speckit` e `cerne init <project-name> --workflow openspec`.
- **FR-002**: A ausência de `--workflow` MUST preservar integralmente o contrato público anterior de `cerne init <project-name>`.
- **FR-003**: O sistema MUST rejeitar valores de workflow diferentes de `speckit` e `openspec`, flag sem valor, repetição da flag e argumentos excedentes como uso inválido antes de criar arquivos.
- **FR-004**: Quando um workflow for selecionado, o manifesto MUST registrar a escolha em um campo opcional e estruturado sem registrar disponibilidade transitória do executável.
- **FR-005**: Workspaces sem o campo de workflow MUST continuar válidos e equivaler ao modo sem workflow.
- **FR-006**: Leitores e escritores existentes do manifesto MUST preservar o campo de workflow quando não estiverem alterando essa configuração.
- **FR-007**: A estrutura comum do knowledge MUST continuar contendo manifesto, informações de produto, decisões, políticas e registros de execução independentemente do provider.
- **FR-008**: A localização canônica das especificações MUST ser determinada pelo workflow: `knowledge/specs` sem workflow ou com Spec Kit, e `knowledge/openspec/specs` com OpenSpec.
- **FR-009**: O bootstrap do provider MUST ocorrer somente dentro do repositório knowledge e MUST NOT criar, copiar ou modificar conteúdo no repositório source.
- **FR-010**: O sistema MUST localizar somente um executável já instalado correspondente ao provider selecionado.
- **FR-011**: O sistema MUST NOT instalar, atualizar ou baixar Spec Kit, OpenSpec ou seus gerenciadores de pacote.
- **FR-012**: A invocação do provider MUST ser não interativa e MUST NOT selecionar ou configurar um agente de IA específico em nome do usuário; quando o provider exigir uma integração, o sistema MUST usar sua opção genérica em um diretório neutro do knowledge.
- **FR-013**: O código e as regras de domínio compartilhadas MUST permanecer independentes de Spec Kit e OpenSpec; comportamento específico de cada provider MUST permanecer em limites substituíveis e testáveis.
- **FR-014**: Se o executável não for encontrado durante `cerne init`, o sistema MUST concluir a criação do workspace com status zero, manter a preferência no manifesto e emitir aviso não bloqueante com instrução de instalação e retomada.
- **FR-015**: O sistema MUST fornecer `cerne workflow setup` para retomar o bootstrap declarado no manifesto sem recriar o workspace.
- **FR-016**: `cerne workflow setup` MUST localizar o workspace Cerne a partir do diretório atual e carregar um manifesto válido antes de executar qualquer provider.
- **FR-017**: Se o workflow já estiver corretamente materializado, `cerne workflow setup` MUST ser idempotente, concluir com status zero e não reexecutar o provider.
- **FR-018**: Se não houver workflow declarado, `cerne workflow setup` MUST falhar sem alteração e informar a ausência da configuração.
- **FR-019**: Se o provider estiver instalado mas falhar durante o init, o sistema MUST retornar falha operacional, preservar o workspace base válido, a preferência e a auditoria, remover somente artefatos novos comprovadamente criados pelo provider e manter o workflow pendente para retomada.
- **FR-020**: Se o provider falhar durante `cerne workflow setup`, o sistema MUST preservar todos os arquivos anteriores à tentativa e MUST NOT considerar o workflow concluído.
- **FR-021**: O sistema MUST validar a estrutura mínima produzida pelo provider antes de comunicar bootstrap concluído.
- **FR-022**: O sistema MUST recusar como inválida qualquer estrutura de workflow que crie repositório Git próprio dentro de knowledge ou altere a independência entre knowledge e source.
- **FR-023**: `cerne doctor` MUST validar o campo e a estrutura esperada para o workflow declarado.
- **FR-024**: Workflow declarado com executável ausente ou estrutura ainda não materializada MUST produzir aviso não bloqueante no doctor.
- **FR-025**: Campo de workflow desconhecido, malformado, estrutura incoerente ou violação da separação Git MUST produzir diagnóstico bloqueante no doctor.
- **FR-026**: Cada tentativa real de executar um provider MUST produzir registro local auditável com provider, operação solicitada, contexto limitado ao knowledge, resultado e timestamps, sem persistir segredos nem a saída externa integral.
- **FR-027**: A opção explícita `--workflow` e o comando explícito `cerne workflow setup` MUST autorizar somente a execução local do bootstrap no knowledge; nenhuma autorização para agentes, source, Git remoto, publicação ou deploy pode ser inferida.
- **FR-028**: O sistema MUST NOT fornecer credenciais ao provider nem persistir variáveis de ambiente, tokens, chaves ou segredos em diagnóstico, manifesto ou auditoria.
- **FR-029**: A saída de sucesso com workflow MUST identificar o workspace, os caminhos knowledge/source, o provider selecionado e se o bootstrap foi concluído ou ficou pendente.
- **FR-030**: Avisos não bloqueantes MUST usar stderr sem alterar o status zero; falhas operacionais MUST usar stderr e status um; uso inválido MUST usar stderr e status dois.
- **FR-031**: A ajuda e a documentação MUST explicar sintaxe, valores aceitos, valor padrão, estrutura condicional, dependências externas, retomada, streams, status, efeitos externos e exemplos.
- **FR-032**: O comportamento funcional MUST ser consistente em Linux, Windows e macOS, incluindo descoberta do executável, argumentos sem shell, caminhos, rollback e códigos de saída.
- **FR-033**: Testes MUST usar providers locais controlados e MUST NOT depender de rede, credenciais, instalações globais reais ou repositórios remotos.
- **FR-034**: O bootstrap MUST usar somente recursos locais empacotados pelo provider e MUST desabilitar telemetria ou atualização remota quando o provider oferecer esse comportamento.

### Constitutional Requirements *(include when the feature affects these concerns)*

- **Ownership/Repositories**: Todo artefato de especificação e configuração do workflow permanece no knowledge; source continua um repositório independente e não recebe cópias automáticas.
- **AI/Integrations**: Spec Kit e OpenSpec entram por adaptadores locais opcionais; nenhum provider ou agente específico se torna requisito do núcleo ou do modo padrão.
- **Context/Audit**: O provider recebe somente o caminho knowledge e argumentos de bootstrap; cada execução real gera registro mínimo suficiente para reconstruir provider, operação, resultado e momento.
- **Authorization/Secrets**: A escolha explícita autoriza apenas o bootstrap local documentado; o Cerne não fornece credenciais, instala ferramentas, acessa remotos nem registra saída externa potencialmente sensível.
- **Portability**: Descoberta, execução sem shell, paths, estrutura esperada e rollback possuem comportamento equivalente nos três sistemas suportados.
- **CLI/Compatibility**: O init sem flag permanece byte a byte compatível; novos contratos de flag, manifesto, warning, comando de retomada, ajuda e doctor são documentados e testados.

### Key Entities *(include if feature involves data)*

- **Workflow Configuration**: Preferência opcional persistida no manifesto, contendo o provider selecionado e nenhuma informação transitória de instalação.
- **Workflow Provider**: Ferramenta local suportada que pode materializar sua estrutura no knowledge mediante solicitação explícita.
- **Workflow Layout**: Conjunto mínimo de caminhos que confirma que o provider foi inicializado no knowledge e define a localização canônica das especificações.
- **Workflow Setup Attempt**: Tentativa auditável de bootstrap com provider, operação, contexto autorizado, resultado e timestamps.
- **Workflow Diagnosis**: Resultado do doctor que distingue configuração ausente, pendente, válida ou inválida.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Em 100% dos cenários válidos com provider disponível, usuários conseguem criar o workspace e materializar o workflow em uma única execução, sem etapa manual intermediária.
- **SC-002**: 100% das execuções sem `--workflow` mantêm manifesto, árvore, stdout, stderr, status e efeitos observáveis do init anterior.
- **SC-003**: 100% dos cenários com provider ausente preservam um workspace base saudável, registram a escolha e retornam orientação de retomada sem status de erro.
- **SC-004**: Após instalar o provider ausente, usuários conseguem concluir o bootstrap com um único `cerne workflow setup`, sem recriar ou editar manualmente o workspace.
- **SC-005**: 100% das falhas de provider durante init ou retomada preservam byte a byte os arquivos preexistentes antes da tentativa, exceto o registro de auditoria criado pela própria tentativa; exatamente um registro permanece como `failed` ou, se sua finalização também falhar, como `started` inconclusivo, e artefatos parciais não são aceitos como workflow concluído.
- **SC-006**: Em todos os cenários, nenhum arquivo do source, remoto Git, credencial ou instalação global é criado, alterado ou removido pelo Cerne.
- **SC-007**: O doctor distingue corretamente, em 100% dos fixtures, workflows ausentes, pendentes, válidos, desconhecidos e estruturalmente inválidos.
- **SC-008**: Cada execução real de provider deixa exatamente um registro auditável sem segredos e sem reproduzir integralmente stdout ou stderr externos.
- **SC-009**: Os cenários de contrato e recuperação produzem os mesmos resultados funcionais em Linux, Windows e macOS.

## Assumptions

- O modo sem workflow permanece o padrão e não grava configuração adicional no manifesto.
- Falhas ocorridas antes de qualquer execução de provider conservam o rollback transacional do init atual; depois de uma execução externa, o workspace base é preservado para manter auditoria e permitir retomada.
- Somente Spec Kit e OpenSpec fazem parte desta versão; registro dinâmico de providers é futuro e está fora de escopo.
- A estrutura geral do knowledge é combinada com a estrutura nativa do provider; apenas a localização canônica das especificações varia.
- O executável esperado para Spec Kit é `specify` e para OpenSpec é `openspec`.
- O bootstrap utiliza opções não interativas e neutras oferecidas pela versão suportada do provider: integração genérica em diretório neutro para Spec Kit e nenhuma ferramenta de IA para OpenSpec; incompatibilidade de versão é tratada como falha diagnosticável, não como instalação automática.
- A estrutura mínima materializada é comprovada por arquivos persistentes de configuração do provider; diretórios vazios que não sobrevivem ao Git não são suficientes para esse diagnóstico.
- A disponibilidade do executável é observada em tempo de execução e não é persistida no manifesto.
- O Cerne inicializa somente a estrutura do workflow. Execução de comandos de especificação, planejamento, agentes ou implementação permanece fora do escopo.
- Conversão, sincronização ou migração entre Spec Kit e OpenSpec permanece fora do escopo.
- Alterar o workflow de um workspace existente permanece fora do escopo; `cerne workflow setup` apenas materializa a preferência já declarada.
- A saída externa integral não é persistida porque pode conter dados desnecessários ou sensíveis; o diagnóstico imediato pode apresentar uma síntese segura.
