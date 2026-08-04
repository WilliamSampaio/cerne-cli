# Feature Specification: Status do Workspace

**Feature Branch**: `003-workspace-status`

**Created**: 2026-08-03

**Status**: Draft

**Input**: User description: "Especifique o comando `cerne status` para apresentar o estado atual
de um workspace Cerne composto por repositórios Git independentes de knowledge e source."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Ver o estado geral do workspace (Priority: P1)

Como usuário, quero executar `cerne status` em qualquer diretório dentro de um workspace Cerne para
ver o projeto identificado e o estado atual dos repositórios de conhecimento e código-fonte.

**Why this priority**: Este é o valor central do comando: responder rapidamente "onde estou e qual
é o estado dos dois repositórios?" sem exigir que o usuário rode comandos Git separados.

**Independent Test**: Criar um workspace válido com manifesto e dois repositórios Git, executar
`cerne status` a partir da raiz e de subdiretórios do workspace, e verificar projeto, caminho
absoluto do workspace, dados de knowledge e dados de source.

**Acceptance Scenarios**:

1. **Given** um workspace válido com manifesto legível, **When** o usuário executa `cerne status`,
   **Then** a saída apresenta o nome do projeto e o caminho absoluto da raiz do workspace.
2. **Given** `knowledge` e `source` são repositórios Git válidos, **When** o status é consultado,
   **Then** a saída apresenta para cada repositório nome, caminho, branch, commit, estado e
   contagens de modificados, em stage e não rastreados.
3. **Given** o usuário está em um subdiretório dentro do workspace, **When** executa
   `cerne status`, **Then** o comando localiza o workspace ancestral correto e apresenta o mesmo
   relatório da raiz.

---

### User Story 2 - Distinguir estados Git relevantes (Priority: P2)

Como usuário, quero distinguir repositórios limpos, alterações fora do stage, alterações em stage,
arquivos não rastreados, detached HEAD e ausência de commits para decidir o próximo passo manual.

**Why this priority**: O comando só é útil se condensar os estados Git mais comuns sem transformar
alterações pendentes em erro.

**Independent Test**: Preparar repositórios locais com cada estado Git previsto e confirmar que o
relatório classifica cada caso corretamente, preservando status de saída zero quando a consulta
for concluída.

**Acceptance Scenarios**:

1. **Given** um repositório sem alterações, **When** `cerne status` é executado, **Then** esse
   repositório é exibido como `limpo` e suas contagens são zero.
2. **Given** um repositório com alterações fora do stage, em stage e arquivos não rastreados,
   **When** o status é executado, **Then** as três contagens são apresentadas separadamente e o
   estado geral é `alterações pendentes`.
3. **Given** um repositório em detached HEAD, **When** o status é executado, **Then** o campo de
   branch informa `detached HEAD` em vez de nome de branch.
4. **Given** um repositório ainda sem commits, **When** o status é executado, **Then** o campo de
   commit informa claramente `sem commits`.

---

### User Story 3 - Receber falhas claras sem efeitos colaterais (Priority: P3)

Como usuário ou script, quero que falhas de localização, manifesto, caminhos ou Git sejam
informadas com o caminho afetado e status diferente de zero, sem que o comando corrija ou altere o
workspace.

**Why this priority**: A previsibilidade de erro mantém o comando seguro para automação e respeita
a separação entre diagnóstico e correção.

**Independent Test**: Executar `cerne status` fora de um workspace e em workspaces com manifesto,
caminhos ou repositórios corrompidos, verificando mensagem, caminho afetado, status diferente de
zero e ausência de mutação.

**Acceptance Scenarios**:

1. **Given** nenhum workspace Cerne pode ser localizado a partir do diretório atual, **When** o
   comando é executado, **Then** ele falha com status diferente de zero e informa o diretório de
   partida.
2. **Given** o manifesto está ausente, ilegível ou inválido, **When** o comando é executado,
   **Then** ele falha com mensagem clara e caminho do manifesto afetado.
3. **Given** um caminho registrado no manifesto não existe ou não é um diretório esperado, **When**
   o comando é executado, **Then** ele falha com o caminho afetado.
4. **Given** `knowledge` ou `source` não é um repositório Git consultável, **When** o comando é
   executado, **Then** ele falha com mensagem clara, caminho do repositório e status diferente de
   zero.

### Edge Cases

- O comando é executado na raiz, em `knowledge`, em `source` ou em subdiretórios internos.
- Há mais de um ancestral com estrutura parecida; o workspace mais próximo do diretório atual é
  usado.
- O manifesto existe, mas contém JSON inválido, campos obrigatórios ausentes ou `source` inválido.
- O caminho de `source` é relativo, mas resolve para fora do workspace ou para recurso inexistente.
- Um dos repositórios não possui commits.
- Um dos repositórios está em detached HEAD.
- Arquivos renomeados, removidos ou copiados aparecem nos estados Git e precisam contribuir para
  as contagens apropriadas.
- O Git está disponível, mas a consulta de um repositório específico falha.
- Caminhos contêm espaços, acentos ou separadores próprios de Linux, Windows ou macOS.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: O sistema MUST oferecer `cerne status` sem argumentos obrigatórios para apresentar o
  estado atual do workspace Cerne localizado a partir do diretório atual.
- **FR-002**: O sistema MUST localizar o workspace procurando o ancestral mais próximo que possua o
  manifesto Cerne esperado.
- **FR-003**: O sistema MUST carregar o manifesto do projeto e exibir o nome do projeto e o caminho
  absoluto da raiz do workspace.
- **FR-004**: O sistema MUST resolver o repositório de conhecimento como `knowledge` dentro do
  workspace localizado.
- **FR-005**: O sistema MUST resolver o repositório de código-fonte pelo caminho registrado no
  manifesto, mantendo-o dentro dos limites do workspace.
- **FR-006**: Para cada repositório, o sistema MUST exibir nome público, caminho absoluto, branch
  atual, commit atual, estado geral e as quantidades de arquivos modificados, em stage e não
  rastreados.
- **FR-007**: Um repositório sem alterações MUST ser apresentado como `limpo`.
- **FR-008**: Um repositório com qualquer alteração em stage, fora do stage ou não rastreada MUST
  ser apresentado como `alterações pendentes`.
- **FR-009**: Alterações em stage, alterações fora do stage e arquivos não rastreados MUST ser
  contadas separadamente.
- **FR-010**: Um repositório em detached HEAD MUST informar `detached HEAD` no campo de branch.
- **FR-011**: Um repositório sem commits MUST informar claramente `sem commits` no campo de commit.
- **FR-012**: Alterações pendentes MUST NOT ser tratadas como erro; quando o estado dos dois
  repositórios for obtido com sucesso, o comando MUST retornar status zero.
- **FR-013**: O sistema MUST retornar status diferente de zero quando o workspace não puder ser
  localizado, o manifesto estiver ausente ou inválido, um caminho registrado não existir, um
  diretório esperado não for repositório Git ou o estado Git não puder ser consultado.
- **FR-014**: Mensagens de erro MUST identificar o problema e o caminho afetado quando houver um
  caminho aplicável.
- **FR-015**: O comando MUST ser estritamente de leitura: MUST NOT criar, corrigir, remover,
  versionar, alternar branch, alterar stage, alterar arquivos ou acessar remotos.
- **FR-016**: A coleta do estado dos repositórios MUST produzir dados reutilizáveis separados da
  apresentação no terminal.
- **FR-017**: O comando MUST produzir resultados funcionalmente consistentes em Linux, Windows e
  macOS.
- **FR-018**: O comando MUST NOT acessar GitHub, remotos, rede, credenciais, agentes de IA ou
  conteúdo além do necessário para identificar manifesto, caminhos e estado Git local.
- **FR-019**: O formato textual, labels principais, stdout, stderr e códigos de saída do comando
  MUST ser documentados como contrato público.

### Constitutional Requirements *(include when the feature affects these concerns)*

- **Ownership/Repositories**: O status informa knowledge e source separadamente e não mistura
  conteúdo privado de conhecimento com código-fonte.
- **AI/Integrations**: A funcionalidade não depende de modelos, agentes ou fornecedores de IA.
- **Context/Audit**: O comando apenas relata estado local observável; nenhuma execução automatizada
  ou registro persistente é criado nesta funcionalidade.
- **Authorization/Secrets**: A invocação autoriza somente leitura local. Nenhuma credencial,
  segredo ou conteúdo privado de arquivos deve ser exibido.
- **Portability**: Caminhos, branch, detached HEAD, repositório sem commits e contagens devem ter
  significado equivalente em Linux, Windows e macOS.
- **CLI/Compatibility**: `cerne status`, sua saída textual, streams e códigos de saída tornam-se
  contrato público e devem ser documentados com exemplos.

### Key Entities *(include if feature involves data)*

- **Workspace Status**: Relatório transitório com nome do projeto, caminho absoluto da raiz e os
  dois estados de repositório.
- **Repository Status**: Estado transitório de `knowledge` ou `source`, contendo nome público,
  caminho, branch, commit, estado geral e contagens de alterações.
- **Workspace Manifest**: Documento do workspace usado para identificar o projeto e localizar o
  repositório de código-fonte.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Em 100% dos workspaces válidos de aceite, `cerne status` apresenta projeto,
  workspace e os dois repositórios com todos os campos obrigatórios.
- **SC-002**: Em 100% dos repositórios limpos de aceite, o estado exibido é `limpo` e as três
  contagens de alterações são zero.
- **SC-003**: Em 100% dos casos com alterações em stage, fora do stage e não rastreadas, as três
  contagens aparecem separadas e corretas.
- **SC-004**: Em 100% dos casos de detached HEAD ou repositório sem commits, a saída comunica o
  estado especial sem tratar isso como erro quando a consulta Git for bem-sucedida.
- **SC-005**: Em 100% das falhas previstas de workspace, manifesto, caminho ou Git, o comando
  retorna status diferente de zero e a mensagem identifica o caminho afetado quando aplicável.
- **SC-006**: Em 100% das execuções de aceite, comparações antes e depois confirmam que nenhum
  arquivo, diretório, stage, branch, commit ou configuração remota foi criado, removido ou alterado.
- **SC-007**: Para um workspace local com dois repositórios pequenos, o status completo é exibido
  em até 5 segundos, sem depender de rede.
- **SC-008**: A mesma matriz de cenários produz resultados equivalentes em Linux, Windows e macOS.

## Assumptions

- O manifesto Cerne esperado fica em `knowledge/cerne.json`.
- A raiz do workspace é o diretório ancestral mais próximo que contém o manifesto esperado.
- O caminho de `source` registrado no manifesto é relativo ao repositório `knowledge`, como nos
  workspaces criados por `cerne init`.
- Contagens são por arquivo afetado em cada categoria visível ao usuário, não por número de hunks
  ou linhas alteradas.
- O comando textual exibe as contagens mesmo quando forem zero, para manter a saída previsível.
- Branch, commit e contagens representam o estado local no momento da consulta; comparação com
  remotos fica fora do escopo.
