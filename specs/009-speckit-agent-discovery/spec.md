# Feature Specification: Descoberta de Agente para Spec Kit

**Feature Branch**: `009-speckit-agent-discovery`

**Created**: 2026-08-06

**Status**: Draft

**Input**: User description: "Adicionar suporte explícito a agentes locais no workflow Spec Kit do Cerne. O Spec Kit continua materializado em knowledge, mas comandos ou skills devem ficar disponíveis a partir da raiz do workspace quando o usuário selecionar um agente como Codex ou Claude. A escolha do agente é local, opcional e não deve ser persistida no manifesto do knowledge."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Usar Spec Kit no Codex a partir da raiz do workspace (Priority: P1)

Uma pessoa cria um workspace Cerne com Spec Kit e escolhe Codex como agente local. Ao abrir o Codex
na raiz do workspace, os comandos do workflow Spec Kit aparecem sem que a pessoa precise entrar
manualmente em `knowledge`.

**Why this priority**: Este é o defeito observado em uso real: o workflow fica pronto dentro de
`knowledge`, mas o agente iniciado na raiz do workspace não encontra os comandos.

**Independent Test**: Inicializar um workspace com Spec Kit e agente Codex, abrir a inspeção de
arquivos a partir da raiz do workspace e confirmar que os comandos esperados do Spec Kit estão
descobríveis no local usado pelo Codex, enquanto `knowledge` continua sendo a raiz do projeto Spec Kit.

**Acceptance Scenarios**:

1. **Given** um diretório pai vazio e Spec Kit disponível, **When** o usuário executa
   `cerne init projeto --workflow speckit --agent codex`, **Then** o workspace é criado com
   `knowledge` e `source` independentes, o workflow é materializado em `knowledge` e os comandos
   Spec Kit ficam disponíveis para Codex na raiz do workspace.
2. **Given** um workspace criado com `--agent codex`, **When** uma sessão Codex começa na raiz do
   workspace, **Then** o agente encontra comandos ou skills Spec Kit sem depender de `cd knowledge`.
3. **Given** comandos expostos para Codex na raiz, **When** eles forem usados, **Then** operam sobre
   `knowledge` como raiz do projeto Spec Kit e não sobre a raiz Cerne nem sobre `source`.

---

### User Story 2 - Restaurar knowledge e escolher outro agente local (Priority: P2)

Uma pessoa restaura o mesmo knowledge em outra máquina e usa um agente diferente do usado na máquina
original. Ela quer preparar a descoberta local desse agente sem alterar a intenção persistida do
workflow.

**Why this priority**: Knowledge viaja entre máquinas, mas agente é preferência local. Persistir o
agente no manifesto faria uma restauração puxar uma escolha que pode não existir no novo ambiente.

**Independent Test**: Restaurar ou preparar um workspace que declara `workflow.provider=speckit`,
executar `cerne workflow setup --agent claude` e confirmar que a ponte local muda para Claude sem
gravar agente em `knowledge/cerne.json`.

**Acceptance Scenarios**:

1. **Given** um workspace restaurado com workflow Spec Kit declarado, **When** o usuário executa
   `cerne workflow setup --agent claude`, **Then** a descoberta local para Claude é criada ou
   atualizada na raiz do workspace e o manifesto permanece sem agente persistido.
2. **Given** um workspace já preparado para Codex, **When** o usuário executa
   `cerne workflow setup --agent claude`, **Then** o Cerne deixa a descoberta local coerente com
   Claude, sem exigir recriação do workspace e sem alterar `source`.
3. **Given** um workflow Spec Kit já materializado em `knowledge`, **When** o usuário solicita
   setup com agente, **Then** o Cerne não precisa tratar o workflow como pendente apenas porque a
   ponte local ainda não existia.

---

### User Story 3 - Manter comportamento legado sem agente (Priority: P3)

Uma pessoa ou automação já usa `cerne init --workflow speckit` sem selecionar agente. Esse fluxo deve
continuar funcionando sem nova escolha obrigatória e sem exposição acidental de comandos de agente.

**Why this priority**: `--workflow speckit` já é contrato público. Tornar agente obrigatório ou mudar
efeitos padrão seria uma quebra desnecessária.

**Independent Test**: Executar os cenários existentes de workflow Spec Kit sem `--agent` e comparar
stdout, stderr, status, manifesto, auditoria e layout com o comportamento compatível esperado.

**Acceptance Scenarios**:

1. **Given** o usuário não informa agente, **When** executa `cerne init projeto --workflow speckit`,
   **Then** o Cerne mantém o comportamento compatível e não cria comandos específicos de Codex ou
   Claude na raiz do workspace.
2. **Given** um workspace sem agente local configurado, **When** `cerne workflow setup` é executado
   sem `--agent`, **Then** somente o workflow declarado é materializado ou reconhecido conforme o
   contrato atual.
3. **Given** uma automação existente que não conhece `--agent`, **When** ela roda os comandos
   antigos, **Then** não precisa alterar argumentos nem tratar novos prompts.

### Edge Cases

- `--agent` é informado sem `--workflow speckit` durante `init`.
- `--agent` é informado em um workspace cujo manifesto declara OpenSpec ou provider desconhecido.
- O agente informado é desconhecido, vazio, repetido ou combinado com argumentos extras.
- O Spec Kit está ausente, falha ou deixa layout parcial antes de qualquer ponte local ser criada.
- O workflow já está pronto em `knowledge`, mas a ponte local está ausente, incompleta ou voltada a
  outro agente.
- A raiz do workspace já contém artefatos de agente criados pelo usuário ou por outro agente.
- O workspace foi restaurado em outra máquina, com paths diferentes e agente local diferente.
- `knowledge` e `source` são repositórios Git independentes e a raiz do workspace não é repositório.
- Paths contêm espaços, diferenças de caixa, separadores Windows ou caracteres Unicode.
- O usuário executa `workflow setup --agent` a partir da raiz, de `knowledge`, de `source` interno
  ou de um descendente do workspace.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: O sistema MUST aceitar a opção opcional `--agent <agent>` em `cerne init` somente
  quando combinada com `--workflow speckit`.
- **FR-002**: O sistema MUST aceitar a opção opcional `--agent <agent>` em `cerne workflow setup`
  para preparar descoberta local de agente quando o workflow declarado for Spec Kit.
- **FR-003**: A primeira versão MUST aceitar exatamente os agentes públicos `codex` e `claude`.
- **FR-004**: O sistema MUST rejeitar agente ausente, desconhecido, repetido ou usado em combinação
  incompatível sem criar ou alterar artefatos de workspace.
- **FR-005**: A ausência de `--agent` MUST preservar o comportamento compatível atual de
  `cerne init --workflow speckit` e `cerne workflow setup`.
- **FR-006**: O sistema MUST NOT expor `generic` como valor público de `--agent`; qualquer uso
  neutro ou legado do provider permanece detalhe interno.
- **FR-007**: O workflow Spec Kit MUST continuar materializado dentro de `knowledge`, mantendo
  `knowledge` como raiz do projeto Spec Kit e `knowledge/specs` como diretório canônico de specs.
- **FR-008**: Quando `--agent` for informado, o sistema MUST criar ou atualizar uma ponte de
  descoberta na raiz do workspace para o agente escolhido.
- **FR-009**: A ponte local MUST permitir que uma sessão iniciada na raiz do workspace encontre os
  comandos ou skills Spec Kit do agente escolhido.
- **FR-010**: Comandos expostos pela ponte local MUST operar contra `knowledge` como raiz Spec Kit,
  não contra a raiz Cerne nem contra `source`.
- **FR-011**: A escolha do agente MUST NOT ser persistida em `knowledge/cerne.json`.
- **FR-012**: O manifesto MUST continuar declarando somente a intenção de workflow necessária ao
  workspace, por exemplo `workflow.provider=speckit`.
- **FR-013**: `cerne workflow setup --agent <agent>` MUST conseguir preparar descoberta local para
  um agente suportado diferente depois de restore ou depois de uma escolha local anterior.
- **FR-014**: Preparar ou trocar ponte local MUST NOT modificar `source`, remotos Git, branches,
  stage, commits, credenciais ou origens do provider.
- **FR-015**: Se o workflow Spec Kit ainda estiver pendente e o executável do provider estiver
  disponível, setup com `--agent` MUST materializar o workflow em `knowledge` e depois preparar
  descoberta local para o agente selecionado.
- **FR-016**: Se o executável do provider estiver ausente, setup com `--agent` MUST preservar o
  comportamento de workflow pendente e MUST NOT criar ponte local que finja que comandos são
  utilizáveis.
- **FR-017**: Se a materialização do workflow falhar, artefatos de descoberta local MUST NOT ser
  reportados como prontos.
- **FR-018**: Se o workflow já estiver estruturalmente pronto em `knowledge`, setup com `--agent`
  MUST poder criar ou atualizar somente a ponte de descoberta local.
- **FR-019**: O estado pronto reportado para workflow MUST continuar se referindo ao layout do
  provider em `knowledge`; descoberta local de agente é uma capacidade local separada.
- **FR-020**: O CLI MUST reportar resultados de descoberta de agente em português com stdout,
  stderr e códigos de saída estáveis.
- **FR-021**: A ajuda de `init` e `workflow` MUST documentar `--agent`, valores suportados,
  limites de compatibilidade, efeitos colaterais e exemplos.
- **FR-022**: Uso inválido MUST retornar status 2, usar stderr, deixar stdout vazio e evitar
  efeitos em arquivos.
- **FR-023**: Falhas operacionais após uma solicitação válida MUST retornar status 1 com causa e
  correção seguras, preservando as regras de auditoria existentes para subprocessos de workflow.
- **FR-024**: Artefatos de descoberta de agente MUST NOT conter segredos, tokens, dumps de
  ambiente, origens remotas, saída bruta do provider ou conteúdo privado de knowledge.
- **FR-025**: O comportamento MUST ser funcionalmente portável entre Linux, Windows e macOS.

### Constitutional Requirements *(include when the feature affects these concerns)*

- **Ownership/Repositories**: Knowledge permanece dono das specs e do estado do workflow; source
  permanece separado e intocado; descoberta local de agente não copia knowledge privado para source.
- **AI/Integrations**: Seleção de agente é responsabilidade de adaptador. O domínio armazena
  intenção de provider, não uma preferência permanente de agente de IA, e novos agentes suportados
  podem ser adicionados sem mudar entidades centrais.
- **Context/Audit**: Execução do provider permanece auditável pelos registros existentes de workflow
  setup. Criar ou atualizar descoberta local é um efeito local de setup e deve ser observável pela
  saída do CLI sem expor contexto desnecessário.
- **Authorization/Secrets**: `--agent` é autorização explícita para criar artefatos de descoberta
  local somente para esse agente. Não autoriza instalações, updates, rede, credenciais, alterações
  de source, commits, pushes, merges, publicação ou deploy.
- **Portability**: Paths de descoberta, nomes de comandos e comportamento de setup devem funcionar
  com a semântica nativa de filesystem em Linux, Windows e macOS.
- **CLI/Compatibility**: `--agent` é um contrato opcional aditivo. Comandos existentes sem ele
  permanecem compatíveis, e novos stdout, stderr, status e textos de ajuda viram comportamento
  público do CLI.

### Key Entities *(include if feature involves data)*

- **Agent Target**: Agente local suportado solicitado pelo usuário para descoberta de comandos,
  inicialmente `codex` ou `claude`.
- **Workflow Declaration**: Intenção de provider manifestada em `knowledge/cerne.json`, mantida
  independente da escolha local de agente.
- **Local Discovery Bridge**: Artefatos na raiz do workspace que tornam comandos Spec Kit visíveis
  ao agente selecionado enquanto direcionam o trabalho para `knowledge`.
- **Workflow Setup Result**: Resultado apresentado ao usuário que distingue materialização do
  provider em `knowledge` de descoberta local de agente.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Em 100% dos cenários válidos de init com Codex, uma sessão iniciada na raiz do
  workspace consegue descobrir os comandos Spec Kit esperados sem mudar de diretório para
  `knowledge`.
- **SC-002**: Em 100% dos cenários válidos de restore/setup, um agente suportado diferente do
  agente da máquina original pode ser selecionado localmente sem adicionar campo de agente a
  `knowledge/cerne.json`.
- **SC-003**: 100% dos testes existentes de workflow Spec Kit sem agente continuam passando sem
  argumentos alterados ou novos prompts obrigatórios.
- **SC-004**: 100% das combinações inválidas de `--agent` deixam snapshots de workspace, knowledge e
  source inalterados e retornam o status documentado de uso inválido.
- **SC-005**: Em Linux, Windows e macOS, setup válido com `--agent` produz resultados observáveis
  equivalentes de workspace, workflow e descoberta para o mesmo agente suportado.

## Assumptions

- Codex e Claude são os únicos alvos públicos iniciais porque são os consumidores imediatos
  discutidos para Cerne skills.
- Descoberta específica de agente pertence à raiz local do workspace, não a `source` e não ao
  manifesto persistido do Cerne.
- O setup neutro ou generic do Spec Kit pode permanecer como mecanismo interno de compatibilidade,
  mas usuários não o selecionam por `--agent`.
- Uma feature futura pode adicionar diagnósticos para descoberta local ausente, mas esta feature
  cobre somente setup explícito por `--agent`.
