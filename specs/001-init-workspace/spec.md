# Feature Specification: Inicialização de Workspace

**Feature Branch**: Não criada (somente especificação)

**Created**: 2026-07-28

**Status**: Draft

**Input**: Comando `cerne init <nome-do-projeto>` para criar um workspace com repositórios
independentes de conhecimento e código-fonte.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Criar um workspace novo (Priority: P1)

Como usuário, quero inicializar um workspace informando apenas o nome do projeto para começar com
os repositórios de conhecimento e código-fonte corretamente separados.

**Why this priority**: Esta é a capacidade mínima que torna o Cerne útil e estabelece a estrutura
fundamental de todos os workspaces futuros.

**Independent Test**: Executar `cerne init exemplo` em um diretório temporário e verificar a
estrutura criada, o manifesto e a independência dos dois repositórios Git.

**Acceptance Scenarios**:

1. **Given** que `exemplo/` não existe, **When** o usuário executa `cerne init exemplo`, **Then**
   o workspace é criado com `knowledge/` e `source/` como repositórios Git independentes.
2. **Given** que `exemplo/` existe e está vazio, **When** o usuário executa `cerne init exemplo`,
   **Then** o diretório existente é usado sem perda de dados e recebe a estrutura completa.
3. **Given** uma inicialização bem-sucedida, **When** o usuário inspeciona `knowledge/`, **Then**
   encontra `cerne.json`, `product/`, `specs/`, `decisions/`, `policies/` e `runs/`.
4. **Given** uma inicialização bem-sucedida, **When** o usuário lê `cerne.json`, **Then** o
   manifesto identifica o projeto como `exemplo` e o repositório de código como `../source`.
5. **Given** uma inicialização bem-sucedida, **When** o usuário inspeciona `source/`, **Then** o
   diretório contém somente os metadados necessários do repositório Git, sem arquivos de aplicação
   ou commits.
6. **Given** uma inicialização bem-sucedida, **When** o comando termina, **Then** retorna status
   zero e exibe em stdout um resumo com os caminhos resolvidos de `knowledge/` e `source/`.

---

### User Story 2 - Preservar dados existentes em falhas (Priority: P2)

Como usuário, quero que a inicialização recuse destinos inseguros e explique como corrigir o
problema para que nenhum arquivo existente seja apagado ou alterado.

**Why this priority**: A confiança no CLI depende de garantir que uma tentativa de inicialização
jamais danifique trabalho existente.

**Independent Test**: Executar o comando contra destinos inválidos, não vazios e ambientes sem Git,
comparando o conteúdo antes e depois de cada tentativa.

**Acceptance Scenarios**:

1. **Given** que `exemplo/` já contém qualquer arquivo ou diretório, **When** o usuário executa
   `cerne init exemplo`, **Then** o comando falha sem alterar o destino, retorna status diferente
   de zero e informa que o usuário deve escolher um destino inexistente ou vazio.
2. **Given** um nome vazio, reservado ou contendo separador de caminho, **When** o usuário tenta
   inicializar o workspace, **Then** o comando falha antes de criar arquivos e explica como fornecer
   um nome portátil válido.
3. **Given** que Git não está disponível, **When** o usuário executa o comando, **Then** nenhum
   workspace parcial permanece e a mensagem orienta a instalar ou disponibilizar Git.
4. **Given** uma falha após o início da criação, **When** o comando encerra, **Then** somente os
   artefatos criados pela própria tentativa são revertidos e todo conteúdo preexistente permanece
   inalterado.

---

### User Story 3 - Usar o comando de forma previsível (Priority: P3)

Como usuário ou script, quero uma interface documentada e consistente para identificar sucesso,
falha e os caminhos criados sem depender de interação adicional.

**Why this priority**: Uma interface previsível permite adoção em terminais, scripts e futuras
automações sem ampliar o escopo da primeira versão.

**Independent Test**: Consultar a ajuda e executar o mesmo cenário válido e inválido em Linux,
Windows e macOS, verificando contrato de entrada, saídas e status.

**Acceptance Scenarios**:

1. **Given** que o usuário solicita ajuda para `cerne init`, **When** a ajuda é exibida, **Then**
   ela documenta sintaxe, argumento, estrutura criada, efeitos, stdout, stderr e situações de erro.
2. **Given** o mesmo nome portátil e um destino vazio, **When** o comando é executado em Linux,
   Windows e macOS, **Then** cria a mesma organização conceitual e usa caminhos válidos para o
   sistema atual.
3. **Given** qualquer falha prevista, **When** o comando termina, **Then** a causa e a orientação
   aparecem em stderr e o status é diferente de zero, sem prompt interativo.

### Edge Cases

- O nome do projeto é vazio, `.` ou `..`.
- O nome contém `/`, `\`, caracteres inválidos ou um nome reservado em algum sistema suportado.
- O destino existe como arquivo, link ou diretório não vazio.
- O destino existe como diretório vazio e possui permissões insuficientes.
- Git não está instalado, não é executável ou falha ao inicializar um dos repositórios.
- A criação falha depois de apenas parte da estrutura ter sido produzida.
- O caminho atual contém espaços ou caracteres Unicode válidos.
- Um repositório Git existe acima do diretório de destino.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: O sistema MUST aceitar exatamente um nome de projeto em
  `cerne init <nome-do-projeto>`.
- **FR-002**: O nome MUST ser um único nome de diretório portátil, não vazio, diferente de `.`
  e `..` e sem separadores de caminho.
- **FR-003**: O sistema MUST resolver o destino como `<diretório-atual>/<nome-do-projeto>`.
- **FR-004**: O sistema MUST criar o destino quando ele não existir e MUST aceitar um destino
  existente somente quando ele for um diretório regular vazio, nunca um link.
- **FR-005**: O sistema MUST criar `knowledge/` e `source/` como diretórios irmãos dentro do
  workspace e MUST NOT inicializar o diretório raiz do workspace como repositório Git.
- **FR-006**: `knowledge/` e `source/` MUST ser repositórios Git independentes, cada um com sua
  própria raiz e metadados, sem incluir o outro em seu escopo de versionamento.
- **FR-007**: O repositório `knowledge/` MUST conter o manifesto `cerne.json` e os diretórios
  `product/`, `specs/`, `decisions/`, `policies/` e `runs/`.
- **FR-008**: `cerne.json` MUST ser um objeto JSON válido com os campos string `name`, contendo o
  nome informado, e `source`, contendo a localização relativa `../source`.
- **FR-009**: O repositório `source/` MUST iniciar sem arquivos de aplicação, remotos ou commits.
- **FR-010**: A primeira versão MUST NOT clonar, vincular ou criar repositórios remotos.
- **FR-011**: Se o destino contiver qualquer entrada, o sistema MUST falhar antes de modificá-lo.
- **FR-012**: O sistema MUST NOT apagar, substituir ou alterar conteúdo existente.
- **FR-013**: Uma falha parcial MUST reverter somente os artefatos criados pela execução corrente
  e MUST NOT deixar um workspace parcialmente inicializado.
- **FR-014**: Erros MUST ser escritos em stderr, retornar status diferente de zero, explicar a
  causa em linguagem compreensível e indicar uma ação corretiva.
- **FR-015**: O sucesso MUST retornar status zero e escrever em stdout um resumo com o nome do
  projeto e os caminhos resolvidos dos dois repositórios.
- **FR-016**: O comando MUST ser não interativo e adequado para execução por scripts.
- **FR-017**: O comportamento funcional e a organização conceitual MUST ser consistentes em Linux,
  Windows e macOS.
- **FR-018**: A ajuda e a documentação do comando MUST cobrir finalidade, sintaxe, argumento,
  estrutura criada, efeitos colaterais, stdout, stderr, status de saída, erros e exemplos.

### Constitutional Requirements *(include when the feature affects these concerns)*

- **Ownership/Repositories**: Todo conteúdo de conhecimento permanece local e sob controle do
  usuário; `knowledge/` e `source/` têm históricos e ciclos de vida Git independentes.
- **AI/Integrations**: A funcionalidade não depende de IA, provedor externo, rede ou Git host.
- **Context/Audit**: Nenhum agente ou execução automatizada de agente ocorre nesta funcionalidade;
  `runs/` é criado vazio para registros futuros.
- **Authorization/Secrets**: A invocação explícita autoriza somente a criação no destino indicado.
  Nenhuma operação remota, destrutiva ou manipulação de credenciais faz parte do comando.
- **Portability**: Nomes, caminhos, diagnósticos e resultados preservam o mesmo significado em
  Linux, Windows e macOS.
- **CLI/Compatibility**: Sintaxe, stdout, stderr, status e estrutura criada passam a constituir
  contrato público; alterações incompatíveis futuras exigem o processo de versionamento do Cerne.

### Key Entities *(include if feature involves data)*

- **Workspace**: Diretório nomeado pelo usuário que associa exatamente um repositório de
  conhecimento e um repositório de código-fonte.
- **Knowledge Repository**: Repositório privado por padrão conceitual que armazena manifesto,
  informações de produto, especificações, decisões, políticas e registros de execução.
- **Source Repository**: Repositório independente reservado para a implementação da aplicação,
  criado vazio nesta versão.
- **Project Manifest**: Registro `cerne.json` com a identidade do projeto e o caminho relativo
  para o repositório de código-fonte.

## Out of Scope

- Integração com Gemini, Codex ou outros agentes.
- Geração de especificações por IA.
- Execução de testes da aplicação.
- Clonagem ou vinculação de repositórios existentes ou remotos.
- Criação automática de repositórios no GitHub ou outro Git host.
- Push, pull request, merge, publicação ou deploy.
- Interface gráfica.
- Configuração de visibilidade remota.
- Criação de arquivos, código ou estrutura da aplicação em `source/`.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Em todos os cenários válidos de aceite, uma única execução cria 100% dos artefatos
  exigidos e dois repositórios Git independentes.
- **SC-002**: Em 100% dos cenários com destino não vazio ou nome inválido, nenhum conteúdo
  preexistente é alterado.
- **SC-003**: Em 100% das falhas previstas, o usuário recebe causa, ação corretiva e status
  diferente de zero.
- **SC-004**: Em 100% das execuções bem-sucedidas, o resumo apresenta os caminhos dos dois
  repositórios e permite localizá-los sem inspeção adicional.
- **SC-005**: O mesmo conjunto de cenários de aceite produz resultados funcionalmente equivalentes
  em Linux, Windows e macOS.
- **SC-006**: Um usuário consegue compreender a estrutura criada, os efeitos e os erros esperados
  consultando apenas a ajuda e a documentação do comando.

## Assumptions

- O comando é executado a partir do diretório que deve receber o workspace.
- Um diretório de destino existente e vazio pode ser reutilizado; qualquer entrada o torna
  inelegível.
- O nome informado é o identificador exibido do projeto e também o nome do diretório.
- O contrato público do manifesto da primeira versão é `knowledge/cerne.json` com os campos
  `name` e `source`; nenhum outro campo é obrigatório nesta funcionalidade.
- Os nomes públicos dos diretórios de conhecimento são `product`, `specs`, `decisions`, `policies`
  e `runs`.
- Git é uma dependência local obrigatória e nenhuma conexão de rede é necessária.
- `runs/` começa vazio porque esta funcionalidade não executa agentes ou tarefas automatizadas.
- A existência de um repositório Git ancestral não muda a independência dos repositórios internos;
  o Cerne não altera esse repositório ancestral.
