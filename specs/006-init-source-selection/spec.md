# Feature Specification: Seleção de Source no Init

**Feature Branch**: `feat/init-source-selection`

**Created**: 2026-08-05

**Status**: Draft

**Input**: User description: "Permitir que `cerne init` associe um repositório source local
existente com `--source` ou clone um repositório com `--clone`; o modo padrão continua criando um
source vazio. Falhas de clone devem ser tratadas com segurança, sem implementar ainda."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Inicializar com source local existente (Priority: P1)

Um usuário que já possui um repositório de código quer criar somente o knowledge e associá-lo ao
source existente, sem criar um diretório source órfão e sem executar depois `link --replace`.

**Why this priority**: Reutiliza a capacidade local já existente, elimina um passo e não exige
rede nem copia código.

**Independent Test**: Criar um repositório Git local temporário com histórico, executar
`cerne init exemplo --source <caminho>` e verificar manifesto, saída, ausência de `exemplo/source`,
independência Git e preservação byte a byte do repositório informado.

**Acceptance Scenarios**:

1. **Given** um repositório Git local válido, **When** o usuário executa `cerne init exemplo --source ../projeto-existente`, **Then** o workspace é criado com knowledge próprio e o manifesto referencia o repositório existente sem modificá-lo.
2. **Given** um caminho relativo de source, **When** o init é executado, **Then** o caminho é resolvido a partir do diretório em que o comando foi invocado e persistido segundo as mesmas regras de portabilidade de `cerne link`.
3. **Given** um repositório inválido, bare, uma subpasta do working tree ou um caminho incompatível com a separação dos repositórios, **When** o init é solicitado, **Then** a operação falha antes de criar o workspace.

---

### User Story 2 - Inicializar clonando um repositório (Priority: P2)

Um usuário quer criar o workspace Cerne e obter o código de um repositório existente em uma única
invocação explícita.

**Why this priority**: Evita preparação manual do diretório source, preservando o fluxo padrão
para projetos novos e o fluxo local para repositórios já disponíveis.

**Independent Test**: Usar um repositório Git local temporário como origem controlada, executar
`cerne init exemplo --clone <origem>` e verificar knowledge independente, clone em `source`,
histórico, branch, remoto, manifesto, saída e ausência de alterações na origem.

**Acceptance Scenarios**:

1. **Given** uma origem Git acessível, **When** o usuário executa `cerne init exemplo --clone <origem>`, **Then** o knowledge é criado e a origem é clonada em `exemplo/source` com seu histórico e remotos.
2. **Given** um repositório remoto vazio, **When** o clone conclui conforme o Git, **Then** o workspace é aceito com source Git válido mesmo sem commit ou arquivo versionado.
3. **Given** `--clone`, **When** a operação ocorre, **Then** o Cerne executa somente um clone padrão pelos transportes permitidos, sem push, recursão de submódulos, publicação ou segundo remoto; redirects, autenticação e filtros de checkout configurados localmente são tratados como efeitos do próprio Git e documentados ao usuário.

---

### User Story 3 - Manter compatibilidade e falhar com segurança (Priority: P3)

Um usuário precisa distinguir erros de uso, validação local e clone, sem perda de dados, vazamento
de credenciais ou mudança no comportamento tradicional do init.

**Why this priority**: A nova opção introduz rede e uma operação Git que pode deixar arquivos
parciais, exigindo limites explícitos antes da implementação.

**Independent Test**: Exercitar combinações inválidas, Git ausente, origem inacessível, falha
parcial controlada e diagnóstico com valores semelhantes a segredo, comparando todos os caminhos
preexistentes antes e depois.

**Acceptance Scenarios**:

1. **Given** nenhuma flag de source, **When** o usuário executa `cerne init exemplo`, **Then** árvore, manifesto, stdout, stderr, status e efeitos permanecem idênticos ao contrato vigente.
2. **Given** `--source` e `--clone` juntos, repetidos, sem valor, fora da posição documentada ou com argumentos extras, **When** o comando é executado, **Then** retorna erro de uso antes de qualquer efeito.
3. **Given** uma origem de clone que exige autenticação não disponível, **When** o clone é tentado, **Then** o Cerne desabilita os prompts e a interatividade controláveis pelo Git, não expõe credenciais e informa uma correção segura; helpers externos permanecem sujeitos à limitação documentada.
4. **Given** qualquer falha, **When** o comando termina, **Then** nenhum arquivo preexistente fora do destino recém-autorizado é alterado ou removido.
5. **Given** exatamente uma opção de source, **When** ela é combinada com `--workflow` em qualquer ordem depois do nome, **Then** ambas as escolhas são aplicadas sem enfraquecer as garantias do source.

### Edge Cases

- O caminho local ou a origem contém espaços, Unicode ou é informado por alias/symlink.
- O source local aponta para a raiz do workspace, para knowledge, para uma subpasta Git ou para um
  repositório bare.
- A origem do clone está vazia, indisponível, é inválida, redireciona ou encerra após criar arquivos.
- A URL contém credencial embutida, query, fragmento ou usa transporte não permitido.
- Git está ausente, não é executável ou uma configuração local tentaria abrir prompt.
- O destino do workspace já existe vazio, não vazio, como arquivo ou como link simbólico.
- A origem local muda entre validação e persistência do manifesto.
- O clone produz finais de linha ou permissões diferentes conforme a plataforma.
- A origem declara submódulos ou arquivos que ativariam ferramentas auxiliares configuradas no Git
  local; o Cerne não solicita recursão nem instalação dessas ferramentas.
- Um processo concorrente cria `workspace/source` antes da promoção do clone validado.
- A limpeza encontra alteração inesperada dentro da área temporária privada criada pelo Cerne.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: O sistema MUST aceitar `cerne init <project-name> --source <local-path>` e `cerne init <project-name> --clone <repository-location>` nas posições documentadas.
- **FR-002**: A ausência das novas flags MUST preservar integralmente o contrato público vigente de `cerne init <project-name>`.
- **FR-003**: `--source` e `--clone` MUST ser mutuamente exclusivos e o sistema MUST rejeitar flags ausentes de valor, repetidas, deslocadas ou acompanhadas de argumentos excedentes antes de criar arquivos.
- **FR-004**: `--source` MUST resolver caminhos relativos a partir do diretório de invocação, antes de criar o workspace.
- **FR-005**: O source local MUST satisfazer as mesmas regras de segurança de `cerne link`: working tree Git existente, não bare, informado pela raiz e independente de knowledge.
- **FR-006**: O modo `--source` MUST NOT criar `workspace/source`, copiar conteúdo, alterar configuração, working tree, index, histórico, remotos ou permissões do repositório informado.
- **FR-007**: O manifesto MUST persistir o source local em representação canônica e portátil compatível com o contrato existente de `cerne link`.
- **FR-008**: O modo `--clone` MUST obter e validar a origem em uma área temporária privada pertencente ao novo workspace e promovê-la a `workspace/source` somente depois de um clone válido.
- **FR-009**: Um clone concluído MUST preservar o histórico, o checkout padrão e o remoto `origin`, inclusive quando a origem está vazia.
- **FR-010**: O Cerne MUST NOT solicitar clone raso, branch específica, recursão de submódulos, execução explícita de Git LFS, segundo remoto ou qualquer extensão não pedida pelo usuário; filtros configurados no Git local podem participar do checkout padrão e esse efeito MUST ser documentado.
- **FR-011**: A flag `--clone` MUST autorizar uma única operação de clone e sua cadeia normal de transporte, redirects, autenticação e checkout; ela MUST NOT autorizar comandos adicionais de push, fetch, submódulo, publicação, merge ou modificação da origem.
- **FR-012**: O sistema MUST usar somente o Git já instalado, sem shell, e aceitar apenas caminhos locais, `file`, HTTPS e SSH, incluindo a forma SCP-like; HTTP sem TLS, protocolo `git`, `ext` e helpers de transporte desconhecidos MUST ser recusados antes da execução.
- **FR-013**: O Cerne MUST rejeitar localizações com credenciais embutidas, query ou fragmento; MUST NOT aceitar credenciais em opção separada, persistir a localização original no manifesto ou registrar URL, usuário, token, senha, chave ou saída Git integral em diagnóstico ou auditoria.
- **FR-014**: Autenticação MUST permanecer sob responsabilidade dos mecanismos externos já configurados para o Git; o Cerne MUST desabilitar prompts e interatividade controláveis pelo Git e documentar que helpers externos podem falhar ou possuir comportamento próprio fora do controle portátil do CLI.
- **FR-015**: O destino do workspace MUST continuar ausente ou ser um diretório regular vazio; nenhum caminho preexistente fora dele pode ser removido.
- **FR-016**: Validações de uso, destino, source local e disponibilidade do Git MUST ocorrer antes de efeitos evitáveis.
- **FR-017**: Se a criação local falhar antes do clone, o sistema MUST aplicar o rollback transacional vigente somente aos artefatos criados pela invocação.
- **FR-018**: Se uma tentativa real de clone ocorrer, o sistema MUST criar previamente um registro auditável mínimo com operação, executor, contexto, autorização, resultado e timestamps, sem segredos nem saída externa integral.
- **FR-019**: Se o clone falhar após iniciar, o sistema MUST preservar o knowledge e a auditoria, remover somente a área temporária privada criada para a tentativa, retornar falha operacional e diagnosticar que o workspace permanece incompleto.
- **FR-020**: A promoção MUST recusar substituir qualquer `workspace/source` que apareça durante a operação; a limpeza MUST limitar-se à área temporária criada pelo Cerne e MUST ser recusada quando sua ownership não puder ser demonstrada.
- **FR-021**: Sucesso MUST usar stdout; falhas operacionais MUST usar stderr e status um; uso inválido MUST usar stderr e status dois; nenhum caminho de falha pode produzir sucesso parcial silencioso.
- **FR-022**: A saída de sucesso MUST identificar projeto, knowledge e source; no modo clone, MUST também indicar que o source foi clonado sem expor a localização original potencialmente sensível.
- **FR-023**: `cerne link`, `doctor` e `status` MUST aceitar os workspaces resultantes sem mudar seus contratos existentes para workspaces anteriores.
- **FR-024**: Ajuda e documentação MUST explicar os três modos, exclusividade das flags, resolução de caminhos, rede, autenticação, remotos, rollback, auditoria, streams, status e exemplos.
- **FR-025**: O comportamento funcional MUST ser equivalente em Linux, Windows e macOS, incluindo paths, aliases, execução sem shell, diagnóstico e rollback.
- **FR-026**: Testes MUST usar repositórios e origens locais temporários, sem rede, credenciais ou remotos reais, e MUST verificar argumentos, streams, status, efeitos, histórico, remotos, auditoria e preservação byte a byte.
- **FR-027**: `--workflow` MAY acompanhar `--source` ou `--clone` em qualquer ordem depois do nome; `--source` e `--clone` MUST continuar mutuamente exclusivos e o workflow MUST NOT alterar o source selecionado.

### Constitutional Requirements *(include when the feature affects these concerns)*

- **Ownership/Repositories**: Knowledge e source permanecem Git independentes; source local nunca é copiado e clone nunca recebe conteúdo privado de knowledge.
- **AI/Integrations**: A feature não envolve agentes ou modelos; Git continua isolado atrás do adapter existente e substituível nos testes.
- **Context/Audit**: Cada clone real precisa de registro mínimo persistente, criado antes do processo e suficiente para reconstruir a ação sem armazenar a origem ou output integral.
- **Authorization/Secrets**: `--clone <origem>` autoriza apenas a obtenção daquela origem; credenciais permanecem externas e valores sensíveis são recusados ou mascarados.
- **Portability**: Caminhos locais, processos, aliases e rollback mantêm comportamento equivalente nas três plataformas.
- **CLI/Compatibility**: O modo sem flag não muda; as novas flags são adições compatíveis de versão MINOR e seus contratos completos devem aparecer na ajuda e documentação.

### Key Entities *(include if feature involves data)*

- **Source Selection**: Escolha transitória entre criar source vazio, associar working tree local ou clonar uma origem.
- **Local Source**: Working tree Git existente, externo ao knowledge, validado e nunca modificado pelo init.
- **Clone Origin**: Localização fornecida somente para a operação, não persistida pelo Cerne e potencialmente sensível.
- **Clone Attempt**: Registro auditável de uma execução real com identidade, autorização, contexto, resultado e timestamps seguros.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Em 100% dos fixtures válidos de source local, o workspace é criado em uma invocação e o repositório informado permanece byte a byte idêntico.
- **SC-002**: Em 100% dos fixtures válidos de clone, o workspace contém knowledge e source Git independentes, e o source preserva histórico e remotos esperados da origem.
- **SC-003**: 100% das execuções sem `--source` ou `--clone` preservam árvore, manifesto, stdout, stderr, status e efeitos do init vigente.
- **SC-004**: 100% dos usos inválidos e sources locais inválidos são recusados sem criar o workspace nem alterar qualquer repositório preexistente.
- **SC-005**: 100% das falhas de clone preservam todos os arquivos preexistentes e não deixam source parcial aceito como workspace saudável.
- **SC-006**: Nenhum fixture persiste ou exibe credencial, localização sensível da origem ou saída Git integral em manifesto, auditoria, stdout ou stderr.
- **SC-007**: Cada clone realmente iniciado deixa exatamente um registro auditável final ou inconclusivo conforme o contrato de falha escolhido.
- **SC-008**: Todos os cenários automatizados passam sem rede ou credenciais reais em Linux, Windows e macOS.

## Assumptions

- `--source` é o atalho atômico para o resultado hoje obtido com `init` seguido de `link --replace`, mas não cria o diretório source interno órfão.
- `--clone` aceita caminhos locais, `file`, HTTPS, SSH e a forma SCP-like; esta versão usa apenas
  validação sintática dos transportes permitidos e não cria integração própria com hosts ou
  autenticação.
- O clone usa o comportamento padrão do Git para branch e checkout e não oferece `--branch`, `--depth`, submódulos ou outras opções nesta versão.
- A localização usada no clone não é necessária no manifesto porque o próprio repositório source mantém seus remotos conforme o Git.
- Repositórios privados funcionam somente quando o Git local consegue autenticá-los sem prompt por mecanismos externos adequados.
- Redirects HTTPS e filtros de checkout configurados localmente, inclusive filtros capazes de
  obter conteúdo adicional, fazem parte dos efeitos normais do Git e devem ser explicitados na
  ajuda; o Cerne não inicia esses mecanismos separadamente.
- Após uma falha de clone, esta versão não oferece comando próprio de retomada; o usuário inspeciona
  a auditoria e pode remover o workspace incompleto para repetir o init ou associar posteriormente
  um source válido pelos meios documentados.
- A área temporária de clone usa nome imprevisível, acesso restrito e o mesmo filesystem do
  workspace. Escritas deliberadas por outro processo da mesma identidade dentro dessa área privada
  não podem ser distinguidas portavelmente e ficam fora do modelo de concorrência suportado.
- Migração entre os três modos depois do init permanece responsabilidade de `cerne link` e de comandos Git explícitos do usuário.
- Clone espelhado, bare, parcial, múltiplos remotos, criação de repositório remoto e importação de knowledge estão fora de escopo.
