# Feature Specification: Diagnóstico de Workspace

**Feature Branch**: `feat/doctor`

**Created**: 2026-07-29

**Status**: Draft

**Input**: User description: "Adicionar `cerne doctor` para diagnosticar, sem modificações, a
validade e a segurança do workspace atual."

## Clarifications

### Session 2026-07-29

- Q: Como `cerne doctor` deve classificar um manifesto cujo `name` é válido, mas diferente do nome
  do diretório raiz do workspace? → A: Aviso não bloqueante; o workspace continua válido.
- Q: Quando o campo `version` estiver presente no manifesto, qual formato deve ser aceito como
  versão 1? → A: Somente o inteiro JSON `1`.
- Q: Quais códigos exatos `cerne doctor` deve retornar para erro bloqueante e uso inválido? →
  A: Status `1` para erro bloqueante e `2` para uso inválido.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Confirmar a saúde do workspace (Priority: P1)

Como usuário, quero verificar o workspace atual para saber se conhecimento, código, manifesto e
repositórios estão íntegros e corretamente separados.

**Why this priority**: Confirmar a estrutura fundamental do Cerne é o valor principal do comando e
permite detectar riscos antes de outras operações.

**Independent Test**: Executar `cerne doctor` na raiz de um workspace válido e verificar que todas
as verificações são aprovadas, o resumo informa saúde e o status é zero.

**Acceptance Scenarios**:

1. **Given** um workspace criado corretamente, **When** o usuário executa `cerne doctor`, **Then**
   cada verificação obrigatória é exibida como aprovada e o resumo informa `Workspace saudável`.
2. **Given** repositórios `knowledge` e `source` válidos e distintos, **When** o diagnóstico é
   executado, **Then** confirma que nenhum deles pertence à árvore de versionamento do outro.
3. **Given** um manifesto legível cujos caminhos existem e cuja versão é suportada, **When** o
   diagnóstico é executado, **Then** as verificações de manifesto, caminhos e versão são aprovadas.

---

### User Story 2 - Localizar problemas bloqueantes (Priority: P2)

Como usuário, quero ver cada problema separadamente e receber um resumo inválido para identificar
o que preciso corrigir manualmente.

**Why this priority**: Um diagnóstico só é confiável se não ocultar falhas e se distinguir
claramente problemas que impedem o uso seguro do workspace.

**Independent Test**: Remover ou tornar inválido um requisito por vez, executar o comando e
verificar a linha com erro, o resumo inválido, a orientação aplicável e o status um.

**Acceptance Scenarios**:

1. **Given** manifesto ausente, ilegível ou inválido, **When** o diagnóstico é executado, **Then**
   a verificação correspondente falha sem impedir a apresentação das demais verificações possíveis.
2. **Given** `knowledge` ou `source` ausente, inacessível ou sem repositório Git próprio, **When**
   o diagnóstico é executado, **Then** cada defeito aplicável é marcado como erro bloqueante.
3. **Given** os dois caminhos resolvendo para o mesmo repositório Git, ou um repositório estando
   sob a raiz de versionamento do outro, **When** o diagnóstico é executado, **Then** o isolamento
   falha e o workspace é declarado inválido.
4. **Given** Git indisponível, versão de manifesto não suportada, caminho registrado inexistente,
   diretório obrigatório ausente ou permissão insuficiente, **When** o diagnóstico é executado,
   **Then** a verificação correspondente falha e informa uma ação corretiva.

---

### User Story 3 - Consumir um diagnóstico previsível e seguro (Priority: P3)

Como usuário ou script, quero símbolos, resumo e status estáveis sem efeitos colaterais para poder
executar diagnósticos repetidamente e automatizar decisões.

**Why this priority**: A previsibilidade habilita automação, enquanto a leitura exclusiva preserva
a confiança do usuário e os limites constitucionais do Cerne.

**Independent Test**: Capturar estado e conteúdo do workspace antes e depois de diagnósticos
saudáveis, com avisos e inválidos; comparar os estados, a saída e o código de status.

**Acceptance Scenarios**:

1. **Given** qualquer workspace analisável, **When** o comando termina, **Then** apresenta
   exatamente um resultado para cada verificação obrigatória, seguido de um único resumo.
2. **Given** somente avisos não bloqueantes, **When** o comando termina, **Then** usa `!`, informa
   `Workspace com avisos` e retorna status zero.
3. **Given** ao menos um erro bloqueante, **When** o comando termina, **Then** usa `✗`, informa
   `Workspace inválido` e retorna status um.
4. **Given** qualquer resultado, **When** o estado anterior e posterior é comparado, **Then** nenhum
   diretório, arquivo, manifesto ou repositório foi criado, alterado ou removido.
5. **Given** um manifesto com `name` portátil diferente do nome da raiz, **When** o diagnóstico é
   executado, **Then** informa aviso, resumo com avisos e status zero.

### Edge Cases

- O comando é executado fora de uma raiz de workspace.
- O manifesto existe, mas está vazio, malformado, com campos obrigatórios ausentes ou caminhos
  inválidos.
- O `name` do manifesto é válido, mas diverge do nome do diretório raiz.
- Um caminho do manifesto é absoluto, escapa da raiz esperada, aponta para link, arquivo ou para o
  mesmo local do outro repositório.
- Um repositório Git existe acima do workspace e pode fazer um diretório ausente parecer
  versionado.
- `knowledge` e `source` existem, mas somente um possui raiz Git própria.
- Um repositório está aninhado no outro ou ambos compartilham a mesma raiz Git.
- Um diretório obrigatório existe como arquivo ou link em vez de diretório regular.
- Parte do workspace pode ser lida, mas não escrita, ou as permissões não podem ser determinadas
  de forma confiável na plataforma atual.
- Git está disponível no início, mas uma inspeção local falha durante o diagnóstico.
- O caminho do workspace contém espaços ou caracteres Unicode válidos.
- Há simultaneamente avisos e erros; erros prevalecem no resumo e no status.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: O sistema MUST oferecer `cerne doctor` sem argumentos obrigatórios para analisar o
  diretório atual como raiz de um workspace Cerne.
- **FR-002**: O sistema MUST oferecer `cerne doctor --help` com finalidade, sintaxe, verificações,
  símbolos, streams, status, efeitos colaterais, limitações, erros e exemplo.
- **FR-003**: O diagnóstico MUST executar e apresentar, em ordem estável, estas verificações:
  manifesto legível; repositório de conhecimento existente; repositório de código-fonte existente;
  repositórios Git independentes; isolamento de versionamento; caminhos do manifesto existentes;
  diretórios obrigatórios existentes; Git disponível; permissões de leitura e escrita; versão do
  manifesto suportada.
- **FR-004**: Cada verificação MUST produzir exatamente uma linha iniciada por `✓` quando aprovada,
  `✗` quando houver erro bloqueante ou `!` quando houver aviso não bloqueante.
- **FR-005**: Cada linha MUST identificar a verificação e apresentar informação suficiente para o
  usuário entender o resultado; erros e avisos MUST incluir orientação corretiva.
- **FR-006**: O sistema MUST continuar todas as verificações que ainda puderem produzir resultado
  confiável após uma falha e MUST identificar como erro as verificações inviabilizadas por uma
  dependência ausente ou inválida.
- **FR-007**: O sistema MUST considerar o manifesto legível somente quando ele existir, puder ser
  lido e possuir os dados obrigatórios válidos para identificar o projeto e o repositório de
  código-fonte. `name` inválido MUST ser erro; `name` válido diferente do nome da raiz MUST ser
  aviso não bloqueante.
- **FR-008**: O sistema MUST resolver os caminhos registrados no manifesto a partir da localização
  definida pelo próprio contrato do manifesto e MUST rejeitar caminhos inexistentes, absolutos,
  que escapem da raiz do workspace, sejam links ou não representem o tipo de recurso esperado.
- **FR-009**: O sistema MUST considerar `knowledge` válido somente quando o diretório esperado
  existir como diretório acessível e representar o repositório de conhecimento do workspace.
- **FR-010**: O sistema MUST considerar `source` válido somente quando o caminho registrado no
  manifesto existir como diretório acessível e representar o repositório de código-fonte.
- **FR-011**: O sistema MUST confirmar que `knowledge` e `source` possuem raízes Git próprias,
  distintas e correspondentes aos caminhos esperados.
- **FR-012**: O sistema MUST falhar o isolamento quando um repositório contiver o outro, quando os
  dois compartilharem a mesma raiz Git ou quando um deles for reconhecido apenas por um repositório
  Git ancestral.
- **FR-013**: O sistema MUST verificar em `knowledge` os diretórios regulares `product`, `specs`,
  `decisions`, `policies` e `runs`; ausência ou tipo incorreto MUST ser erro bloqueante.
- **FR-014**: O sistema MUST verificar a disponibilidade local do executável Git antes de concluir
  as verificações dependentes dele; indisponibilidade MUST ser erro bloqueante com orientação para
  disponibilizá-lo.
- **FR-015**: O sistema MUST verificar se manifesto, diretórios obrigatórios e raízes dos dois
  repositórios possuem as permissões de leitura e escrita necessárias ao uso normal do workspace.
- **FR-016**: Falta confirmada de leitura ou escrita em um recurso obrigatório MUST ser erro
  bloqueante; uma limitação da plataforma que impeça conclusão confiável MUST ser aviso explícito,
  sem alegar aprovação.
- **FR-017**: A versão inicial suportada do manifesto MUST ser a versão `1`. Um manifesto atual sem
  versão explícita MUST ser interpretado como versão `1` implícita e aprovado. Quando presente,
  `version` MUST ser o inteiro JSON `1`; qualquer outro tipo ou valor MUST ser erro bloqueante.
- **FR-018**: Ao final, o sistema MUST apresentar exatamente um dos resumos: `Workspace saudável`
  quando todas as verificações forem aprovadas, `Workspace com avisos` quando não houver erro e
  existir aviso, ou `Workspace inválido` quando houver qualquer erro bloqueante.
- **FR-019**: O comando MUST retornar status zero para workspace saudável ou somente com avisos e
  status `1` quando houver qualquer erro bloqueante.
- **FR-020**: Resultados e resumo MUST usar stdout. Uso inválido ou impossibilidade de iniciar o
  próprio comando MUST usar stderr e não MUST produzir um resumo enganoso. Uso inválido MUST
  retornar status `2`; falha operacional anterior ao relatório MUST retornar status `1`.
- **FR-021**: O comando MUST ser não interativo e adequado para scripts, com ordem, símbolos,
  streams, textos de resumo e códigos de saída documentados como contrato público.
- **FR-022**: O diagnóstico MUST ser estritamente de leitura: MUST NOT criar, corrigir, substituir,
  alterar ou remover arquivos, diretórios, manifesto, configuração ou estado dos repositórios.
- **FR-023**: Toda inspeção Git MUST ser local e não modificadora; o sistema MUST NOT acessar
  remotos, GitHub ou qualquer outro serviço de rede.
- **FR-024**: O comando MUST NOT solicitar credenciais, expor segredos, chamar agentes de IA ou
  registrar conteúdo privado do workspace.
- **FR-025**: O comportamento funcional, os resultados e os status MUST ser equivalentes em Linux,
  Windows e macOS; limitações reais de permissão MUST ser indicadas sem falsos resultados.

### Constitutional Requirements

- **Ownership/Repositories**: O diagnóstico preserva o controle do usuário, valida a separação entre
  conhecimento e código e não lê conteúdo além do necessário para as verificações declaradas.
- **AI/Integrations**: A funcionalidade não depende de IA, fornecedor, Git host ou rede; Git é
  tratado somente como limite externo local e substituível.
- **Context/Audit**: Nenhum agente é executado e nenhum registro é persistido, pois a garantia de
  leitura exclusiva proíbe mutações; a saída observável descreve todas as verificações realizadas.
- **Authorization/Secrets**: A invocação autoriza apenas inspeção local. Não há operação sensível,
  credencial ou exposição do conteúdo privado encontrado.
- **Portability**: Caminhos, repositórios e permissões mantêm significado consistente nos três
  sistemas suportados; incerteza de plataforma gera aviso, não aprovação falsa.
- **CLI/Compatibility**: Símbolos, ordem das verificações, resumos, streams e status tornam-se
  contrato público e MUST ser documentados junto com o comando.

### Key Entities

- **Workspace Diagnostic**: Resultado completo de uma execução, contendo dez verificações
  individuais e um estado final.
- **Check Result**: Resultado nomeado com severidade `aprovado`, `aviso` ou `erro`, explicação e,
  quando aplicável, orientação corretiva.
- **Workspace Manifest**: Registro que identifica o projeto, a localização do código-fonte e a
  versão de formato usada para interpretação.
- **Repository Boundary**: Relação entre as raízes de conhecimento e código usada para comprovar
  que são distintas e não versionam uma à outra.

## Out of Scope

- Criar, remover, mover ou corrigir diretórios e arquivos.
- Atualizar, migrar ou reescrever o manifesto.
- Inicializar ou reparar repositórios Git.
- Criar commits, configurar remotos ou executar qualquer operação Git modificadora.
- Acessar GitHub, Git hosts, rede, credenciais ou serviços externos.
- Chamar agentes de IA, gerar recomendações por IA ou executar tarefas de manutenção.
- Diagnosticar conteúdo ou qualidade da aplicação armazenada em `source`.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Em 100% dos workspaces saudáveis de aceite, sem erros ou avisos, o usuário recebe dez
  resultados individuais, um resumo saudável e status zero.
- **SC-002**: Em 100% dos defeitos bloqueantes previstos, ao menos uma linha identifica o defeito,
  o resumo informa workspace inválido e o status é `1`.
- **SC-003**: Em 100% dos casos contendo apenas avisos, o resumo informa workspace com avisos e o
  status permanece zero.
- **SC-004**: Em 100% das execuções, comparações antes e depois confirmam que nenhum conteúdo ou
  configuração do workspace foi criado, alterado ou removido.
- **SC-005**: Em 100% das linhas com erro ou aviso, a saída identifica a verificação problemática e
  inclui uma ação corretiva, sem depender de registros externos.
- **SC-006**: Para um workspace local com a estrutura mínima, o diagnóstico completo termina em até
  5 segundos, sem depender de rede.
- **SC-007**: O mesmo conjunto de cenários produz classificação, resumo e status funcionalmente
  equivalentes em Linux, Windows e macOS.

## Assumptions

- O usuário executa o comando na raiz que deseja diagnosticar; descoberta em diretórios ancestrais
  ou seleção por argumento fica fora desta primeira versão.
- O manifesto atual fica em `knowledge/cerne.json`, identifica o projeto e registra o caminho de
  `source` relativo a `knowledge`.
- O formato atual sem versão explícita representa implicitamente a versão `1`, preservando como
  saudáveis os workspaces já criados por `cerne init`.
- Ausência de permissão de escrita torna o workspace inadequado ao uso normal do Cerne e, portanto,
  é bloqueante mesmo que o diagnóstico em si consiga apenas ler.
- Uma verificação dependente de outra não é omitida: ela aparece como erro com a dependência que
  impediu sua conclusão.
- Entradas adicionais no workspace não são problema por si só, desde que não violem limites entre
  repositórios, caminhos declarados ou requisitos obrigatórios.
