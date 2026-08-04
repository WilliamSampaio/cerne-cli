# Feature Specification: Link de Repositório Source

**Feature Branch**: `004-link-source-repository`

**Created**: 2026-08-04

**Status**: Draft

**Input**: User description: "Especifique o comando `cerne link` para permitir que um workspace existente utilize um repositório Git local já existente como seu repositório source."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Vincular um source local válido (Priority: P1)

Um usuário que já possui um workspace Cerne quer substituir o source inicial por um repositório
local de código-fonte existente, sem copiar nem mover arquivos.

**Why this priority**: Este é o valor principal do comando: reaproveitar um repositório de aplicação
real em um workspace Cerne existente.

**Independent Test**: Criar um workspace Cerne, criar um repositório Git local com árvore de
trabalho em outro caminho, executar `cerne link <caminho> --replace` e verificar que o manifesto
passa a referenciar esse source, sem alterações no repositório vinculado.

**Acceptance Scenarios**:

1. **Given** um workspace Cerne válido e um repositório Git local não-bare, **When** o usuário executa `cerne link ../app --replace`, **Then** o manifesto é atualizado para apontar para esse repositório e o comando exibe o projeto, o source anterior, o novo source e a confirmação da atualização.
2. **Given** um workspace Cerne válido e um caminho absoluto para um repositório Git local válido, **When** o usuário executa `cerne link <caminho-absoluto>`, **Then** o comando aceita o caminho, armazena a referência normalizada e mantém o repositório vinculado intacto.
3. **Given** um repositório Git worktree válido, **When** o usuário executa `cerne link <worktree>`, **Then** o comando aceita o worktree como source.

---

### User Story 2 - Evitar substituição acidental de source (Priority: P2)

Um usuário quer segurança ao apontar para outro source quando o manifesto já possui um source
configurado.

**Why this priority**: A troca de source altera a referência central do workspace e pode confundir
automação e documentação se ocorrer por engano.

**Independent Test**: Preparar um manifesto que já aponta para um source válido, executar `cerne link`
com outro repositório e verificar recusa sem `--replace`; repetir com `--replace` e verificar apenas
a atualização do manifesto.

**Acceptance Scenarios**:

1. **Given** um workspace cujo manifesto já aponta para um source diferente, **When** o usuário executa `cerne link <novo-source>` sem `--replace`, **Then** o comando falha, mantém o manifesto anterior e informa que `--replace` é necessário.
2. **Given** um workspace cujo manifesto já aponta para outro source, **When** o usuário executa `cerne link <novo-source> --replace`, **Then** o manifesto passa a apontar para o novo source e o source anterior não é removido, alterado ou acessado remotamente.
3. **Given** um workspace cujo manifesto já aponta para o mesmo source informado, **When** o usuário executa `cerne link <source-atual>`, **Then** o comando conclui com sucesso e informa que nenhuma alteração foi necessária.

---

### User Story 3 - Receber falhas claras e seguras (Priority: P3)

Um usuário precisa entender por que o vínculo não pode ser criado quando o workspace, manifesto,
caminho informado ou relação entre repositórios é inválida.

**Why this priority**: Falhas de link devem ser seguras e compreensíveis, pois o comando altera uma
referência persistente do workspace.

**Independent Test**: Executar o comando fora de um workspace e com fixtures inválidas para caminho,
manifesto, repositório Git, bare repository, source igual a knowledge e repositórios aninhados,
verificando erro, caminho afetado, orientação e ausência de alterações.

**Acceptance Scenarios**:

1. **Given** um diretório que não pertence a um workspace Cerne, **When** o usuário executa `cerne link ../app`, **Then** o comando falha, informa que o workspace não foi localizado e não altera arquivos.
2. **Given** um caminho inexistente ou que não é diretório, **When** o usuário executa `cerne link <caminho>`, **Then** o comando falha com o caminho afetado e mantém o manifesto inalterado.
3. **Given** um repositório bare, **When** o usuário executa `cerne link <repo-bare>`, **Then** o comando falha informando que repositórios bare não são aceitos.
4. **Given** um source que é o mesmo repositório do knowledge ou está aninhado de forma perigosa, **When** o usuário executa `cerne link <caminho>`, **Then** o comando falha e explica que knowledge e source devem permanecer independentes.

### Edge Cases

- Caminho informado contém espaços, Unicode, `.` ou `..`.
- Caminho informado é relativo a partir de um subdiretório dentro do workspace.
- Source informado fica fora do workspace, mas no mesmo volume ou em volume diferente.
- Source atual do manifesto está ausente ou inválido antes do link.
- Manifesto existe mas não pode ser lido, está malformado ou usa versão incompatível.
- Repositório informado é Git válido, mas não possui árvore de trabalho.
- Source e knowledge usam worktrees ou metadados Git que poderiam compartilhar o mesmo histórico.
- Falha ocorre durante a gravação do manifesto.
- `--replace` é informado em posição inválida ou junto de argumento extra.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: O comando MUST aceitar a sintaxe `cerne link <caminho-do-repositorio>` e `cerne link <caminho-do-repositorio> --replace`.
- **FR-002**: O comando MUST localizar o workspace Cerne a partir do diretório atual antes de validar ou gravar qualquer alteração.
- **FR-003**: O comando MUST carregar o manifesto do workspace e falhar quando ele estiver ausente, ilegível, malformado ou incompatível.
- **FR-004**: O caminho informado MUST ser aceito como relativo ou absoluto.
- **FR-005**: Caminhos relativos informados pelo usuário MUST ser resolvidos a partir do diretório onde o comando foi executado.
- **FR-006**: O caminho informado MUST existir e apontar para um diretório.
- **FR-007**: O diretório informado MUST representar um repositório Git local com árvore de trabalho.
- **FR-008**: Repositórios Git bare MUST NOT ser aceitos como source.
- **FR-009**: Git worktrees válidos MUST ser aceitos como source.
- **FR-010**: O comando MUST NOT copiar, mover, renomear, criar link simbólico, apagar ou modificar arquivos do repositório vinculado.
- **FR-011**: O comando MUST NOT executar operações Git que alterem o repositório vinculado, incluindo checkout, reset, add, commit, clean, fetch, pull ou push.
- **FR-012**: O source informado MUST NOT ser o mesmo repositório usado como knowledge.
- **FR-013**: Knowledge e source MUST NOT estar aninhados de forma que um repositório possa versionar acidentalmente o conteúdo do outro.
- **FR-014**: O caminho persistido no manifesto MUST ser normalizado.
- **FR-015**: Sempre que o source puder ser representado de forma portável em relação ao workspace, o manifesto MUST armazenar caminho relativo.
- **FR-016**: Quando uma representação relativa portátil não for possível, o manifesto MAY armazenar caminho absoluto normalizado.
- **FR-017**: Se o repositório informado já for o source configurado, o comando MUST concluir com sucesso e informar que nenhuma alteração foi necessária.
- **FR-018**: Se o manifesto já apontar para outro source válido, o comando MUST recusar a substituição por padrão.
- **FR-019**: A substituição de outro source MUST ocorrer somente quando `--replace` for informado.
- **FR-020**: Mesmo com `--replace`, o comando MUST NOT apagar, mover, modificar ou acessar remotamente o source anteriormente configurado.
- **FR-021**: Todas as validações MUST ser concluídas antes de qualquer tentativa de gravação do manifesto.
- **FR-022**: A atualização do manifesto MUST ser atômica; se a gravação falhar, o manifesto anterior MUST permanecer válido e utilizável.
- **FR-023**: Ao concluir com atualização, o comando MUST exibir nome do projeto, source anteriormente configurado quando existir, novo caminho do source e confirmação de que o manifesto foi atualizado.
- **FR-024**: Ao concluir sem alteração por source idêntico, o comando MUST exibir nome do projeto, source atual e mensagem clara de que nenhuma alteração foi necessária.
- **FR-025**: Erros bloqueantes MUST usar mensagens claras, informar o problema, o caminho afetado quando houver e a ação necessária quando aplicável.
- **FR-026**: O comando MUST falhar sem alterar o manifesto para workspace não localizado, manifesto inválido, caminho inexistente, caminho não diretório, repositório Git inválido, repositório bare, source igual a knowledge, sobreposição perigosa, substituição sem `--replace` e falha de gravação segura.
- **FR-027**: O comando MUST rejeitar argumentos extras, flags desconhecidas e uso de `--replace` sem caminho de repositório.
- **FR-028**: O comando MUST produzir comportamento consistente em Linux, Windows e macOS, incluindo normalização de caminhos, aliases de filesystem e volumes distintos.

### Constitutional Requirements *(include when the feature affects these concerns)*

- **Ownership/Repositories**: O comando MUST preservar a separação entre knowledge e source, sem copiar conhecimento para o source nem incorporar um repositório dentro do outro.
- **AI/Integrations**: A funcionalidade MUST operar sem modelos, agentes ou fornecedores de IA; Git local é apenas uma dependência de workspace, não uma integração remota.
- **Context/Audit**: Nenhum agente é executado e nenhum contexto de projeto é compartilhado; a saída do comando deve ser suficiente para auditoria humana imediata da alteração de manifesto.
- **Authorization/Secrets**: `--replace` é autorização explícita para trocar apenas a referência no manifesto; o comando MUST NOT manipular segredos, credenciais, remotos ou publicação.
- **Portability**: O comportamento de path, worktree, repositório bare, comparação de repositórios e atualização atômica MUST ser validado nos sistemas suportados.
- **CLI/Compatibility**: Sintaxe, stdout, stderr, mensagens de ajuda, códigos de saída e formato do manifesto são contratos públicos e MUST ser documentados junto ao comando.

### Key Entities *(include if feature involves data)*

- **Workspace Cerne**: Diretório localizado a partir do ponto de execução; contém o repositório knowledge e um manifesto.
- **Manifesto do Projeto**: Arquivo que identifica o projeto e registra a referência atual para o source.
- **Repositório Knowledge**: Repositório Git do conhecimento do projeto; deve permanecer independente do source.
- **Source Candidato**: Caminho informado pelo usuário que deve ser validado como repositório Git local com árvore de trabalho.
- **Resultado do Link**: Estado final comunicado ao usuário: atualizado, sem alteração necessária ou falha bloqueante.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Usuários conseguem vincular um repositório Git local válido a partir da raiz, de `knowledge/` ou de subdiretório do workspace em até 5 segundos em um workspace pequeno.
- **SC-002**: 100% dos cenários de falha bloqueante preservam o manifesto anterior byte a byte.
- **SC-003**: 100% dos cenários bem-sucedidos preservam arquivos, branch, stage, commits e remotos dos repositórios source antigo e novo.
- **SC-004**: O comando recusa substituição de source sem `--replace` em todos os casos em que o source atual e o source informado são diferentes.
- **SC-005**: A saída de sucesso permite identificar, sem abrir arquivos, o projeto, o source anterior, o novo source e se houve atualização ou nenhuma alteração.
- **SC-006**: Mensagens de erro para entradas inválidas indicam causa e caminho afetado em todos os casos aplicáveis.
- **SC-007**: O mesmo conjunto de cenários de aceitação passa em Linux, Windows e macOS.

## Assumptions

- O manifesto atual registra o source como caminho relativo ao repositório knowledge.
- Um caminho relativo portátil é preferido quando source e workspace puderem ser expressos sem depender de volume, raiz ou prefixo absoluto específico do sistema.
- Quando o campo `source` do manifesto existe, mas aponta para caminho ausente ou repositório inválido, ele ainda conta como configuração existente; trocar para outro source exige `--replace`. Manifesto ausente, malformado, sem campo `source` ou com versão incompatível continua bloqueando o comando antes de qualquer substituição.
- O comando pode consultar metadados Git locais necessários para distinguir repositório com árvore de trabalho, repositório bare, worktree e identidade dos repositórios.
- Códigos de saída seguirão o padrão atual do CLI: sucesso ou ajuda com zero, falha operacional com não-zero e uso inválido com código específico de uso.
