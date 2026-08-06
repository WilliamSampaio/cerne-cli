# Research: Restauração de Workspace

## Decisão 1: Definir quando uma tentativa começa

**Decision**: Help, erro de parsing e origem recusada pela classificação estática não criam
auditoria nem executam Git. Depois que a sintaxe, as duas seleções e suas origens passam pelas
checagens puras, a invocação torna-se uma tentativa: cria um registro global durável antes da
primeira inspeção ou clone Git.

**Rationale**: Preserva o contrato de uso inválido sem efeitos, mas garante auditoria antes de todo
processo externo de uma restauração válida. A fronteira também permite recusar sobreposição do
source local com a home, a auditoria ou o pai do workspace antes que o próprio registro o altere.

**Alternatives considered**: Auditar `--help` e erros sintáticos; iniciar auditoria somente antes do
primeiro clone, deixando inspeções Git sem registro; criar audit antes de validar sobreposição e
violar a imutabilidade do source quando ele contém a home.

## Decisão 2: Manter a auditoria em estado global mínimo

**Decision**: Usar `filepath.Join(os.UserHomeDir(), ".cerne", "audit")` nos três sistemas. Criar
diretórios regulares privados e recusar `.cerne` ou `audit` como symlink/arquivo. Confinar o acesso
à home com `os.OpenRoot`, criar um `restore-<id-opaco>.json` exclusivo com modo `0600` e manter
diretórios em `0700`. Em Windows, a privacidade equivalente depende da ACL herdada do perfil.

**Rationale**: Corresponde ao local escolhido pelo usuário, existe antes de knowledge e sobrevive ao
rollback. `os.Root`, biblioteca padrão do Go alvo, limita operações à home sem novo pacote. Nome
opaco evita colisão e não depende do nome ainda desconhecido do workspace.

**Alternatives considered**: `os.UserConfigDir`, que mudaria para XDG/AppData sem necessidade;
arquivo nomeado pelo projeto, impossível antes do clone e sujeito a colisão; audit dentro de
knowledge, que não existe a tempo; SQLite, daemon ou journal, sem requisito atual.

## Decisão 3: Não armazenar qualquer representação das origens

**Decision**: O registro contém somente operação, modo de source, autorização sem valores, nome do
workspace após validação, fases, estados, timestamps e categoria de falha controlada. Não contém
URL, path de origem, host, fingerprint, argumentos, ambiente ou output Git. A URL preservada pelo
próprio Git como remoto `origin` pertence ao repositório clonado e não é duplicada pelo Cerne.

**Rationale**: Atende à minimização e evita correlação/dicionário sobre fingerprints. As fases e
categorias são suficientes para reconstruir o ponto atingido pela operação.

**Alternatives considered**: Reutilizar o fingerprint SHA-256 do audit de `init --clone`; armazenar
host ou URL redigida; persistir origem para retry. Todos ampliam exposição sem serem necessários à
restauração ou à auditoria pedida.

## Decisão 4: Usar uma única transação em staging irmão

**Decision**: Criar `<parent>/.cerne-restore-*` privado, clonar knowledge em `knowledge`, obter o nome
do manifesto e preparar o source dentro da mesma árvore. Depois da validação integral, promover o
staging inteiro para `<parent>/<name>` com a primitiva existente de rename sem substituição.

**Rationale**: Antes da promoção existe exatamente um artefato de workspace comprovadamente criado
pela tentativa. O nome não precisa ser conhecido antes do clone, source clonado não precisa de um
segundo staging e resultado parcial nunca ocupa o path público.

**Alternatives considered**: Clonar direto no destino, cujo nome ainda é desconhecido; criar
knowledge e source publicamente em etapas; copiar recursivamente knowledge local; usar
`InitWithSource*`, que cria manifesto novo e possui rollback/audit incompatíveis.

## Decisão 5: Recusar qualquer destino final existente

**Decision**: Depois de obter e validar `manifest.name`, exigir que `<parent>/<name>` esteja ausente.
Diretório vazio, não vazio, arquivo e symlink existentes são recusados e preservados. A promoção
repete a garantia de no-replace contra criação concorrente.

**Rationale**: FR-016 exige destino ausente após falha, enquanto FR-017 proíbe remover algo sem
ownership. Um diretório vazio preexistente não pertence à tentativa; aceitá-lo tornaria as duas
garantias simultaneamente impossíveis. A regra mais estrita resolve a ambiguidade de FR-005.

**Alternatives considered**: Aceitar vazio e preservá-lo no rollback, violando “destino ausente”;
removê-lo, violando ownership; adicionar lock/snapshot do diretório, que não transfere ownership.

## Decisão 6: Reutilizar políticas, não o caso de uso de init

**Decision**: Reutilizar `gitexec.ClassifyCloneOrigin`, `FindClone`, `validLinkRepository`,
`validateLinkSeparation`, `readManifest`, `ValidateName`, escrita atômica do manifesto, validações do
doctor e a promoção multiplataforma existente. Implementar a orquestração em `restore.go`; não
chamar `initWorkspaceMode`, `initWithClonedSource` ou `rollbackInitializedWorkspace`.

**Rationale**: Os componentes existentes já concentram allowlist, clone sem shell, Git roots,
separação, schema e comportamento portável. O caso de uso atual, porém, cria knowledge e preserva
workspace incompleto após clone, exatamente o oposto do rollback integral do restore.

**Alternatives considered**: Duplicar adapter e validadores; compor comandos públicos `init` e
`link`; criar interfaces/factories genéricas; mover toda a lógica existente para novos pacotes.

## Decisão 7: Tratar source local e clonado conforme ownership

**Decision**: Source local é resolvido e checado lexicalmente antes da auditoria; depois é
inspecionado como Git root non-bare antes/depois da montagem e promoção. Ele não pode conter nem ser
contido por knowledge, destino, staging, pai da restauração ou audit global. Se o manifesto aponta
outro source, atualizar somente o campo `source` no manifesto staged, calculado em relação ao
knowledge final. Source clonado exige referência relativa portátil que resolva dentro do workspace
e fora de knowledge; clonar exatamente nesse path.

**Rationale**: Preflight impede que audit/staging alterem um source que os contenha. Revalidação
detecta troca concorrente sem lock. Atualização por mapa JSON preserva campos conhecidos e
desconhecidos; validação lexical rejeita absolutos/traversal de outro sistema antes do filesystem.

**Alternatives considered**: Copiar ou mover source local; persistir sempre `../source`; aceitar
absoluto externo para clone; usar lock; atualizar o manifesto depois da promoção, ampliando o
rollback público.

## Decisão 8: Tornar a auditoria parte da transação

**Decision**: Persistir cada transição antes do processo que ela autoriza. Estados duráveis cobrem
knowledge e source separadamente. Depois da validação, promover o root, revalidar paths finais e
então finalizar `succeeded`. Se essa finalização falhar, remover o root promovido somente após
confirmar sua identidade e deixar o último JSON durável inconclusivo. Falhas normais finalizam
`failed`; falha ao escrever o estado final também deixa registro inconclusivo.

**Rationale**: Nunca se comunica sucesso sem audit final e nunca se perde a evidência global no
rollback. Identidade, pai, prefixo e tipo regular limitam limpeza automática a artefatos próprios.

**Alternatives considered**: Marcar sucesso antes da promoção, podendo mentir se rename falhar;
preservar workspace quando audit final falha, contrariando rollback integral; apagar audit junto;
remover paths apenas pelo nome.

## Decisão 9: Preservar workflow sem executá-lo

**Decision**: Validar schema/version do manifesto, provider conhecido e layout com as regras
existentes. Estado pendente ou executável ausente não bloqueia restore; layout parcial/inválido
bloqueia. Nenhum provider é executado e nenhuma ferramenta é instalada.

**Rationale**: Um workspace pode ser restaurado em máquina nova antes de instalar o provider. O
comando existente `cerne workflow setup` continua sendo a única autorização para materialização.

**Alternatives considered**: Executar setup automaticamente; ignorar workflow; exigir executável
instalado; adicionar flag de workflow ao restore.

## Decisão 10: Não criar retomada, retenção ou seleção de revisão

**Decision**: Uma falha é retomada por nova invocação depois da correção. Audits não são removidos
ou rotacionados automaticamente. Clone usa checkout padrão atual, sem branch, tag, commit, depth,
submódulo, LFS explícito ou retry.

**Rationale**: O rollback deixa nenhum workspace parcial a retomar. As opções extras criariam
novos estados e contratos sem requisito atual.

**Alternatives considered**: `restore --resume`, cleanup/retention configurável, pins de revisão,
retry automático e sync de workspace existente.
