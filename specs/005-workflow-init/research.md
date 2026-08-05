# Research: Inicialização com Workflow SDD

## Decisão 1: Manter o modo padrão inalterado

**Decision**: `cerne init <project-name>` continua sem workflow, sem campo adicional, sem processo
externo e com a mesma árvore e saídas. O workflow exige `--workflow speckit|openspec`.

**Rationale**: Sintaxe, manifesto, stdout, stderr, status e efeitos do init atual são contratos
públicos. Uma flag opt-in permite a funcionalidade como adição compatível de versão MINOR.

**Alternatives considered**: Prompt durante todo init, que quebra automação; workflow padrão, que
amplia dependências e efeitos; flag `--no-workflow`, que inverte o contrato atual.

## Decisão 2: Combinar a estrutura Cerne com o layout nativo

**Decision**: `product`, `decisions`, `policies` e `runs` permanecem comuns. Sem workflow e com
Spec Kit, especificações ficam em `knowledge/specs`; com OpenSpec, a fonte canônica fica em
`knowledge/openspec/specs`. Esses caminhos pertencem à descrição resolvida pelo adapter; o domínio
apenas aplica os caminhos genéricos recebidos. Arquivos nativos do provider não são convertidos nem
duplicados.

**Rationale**: Os diretórios comuns representam responsabilidades do Cerne não cobertas
integralmente pelos providers. Manter o formato nativo evita sincronização bidirecional e lock-in
criado pelo próprio wrapper.

**Alternatives considered**: Substituir todo knowledge, perdendo conceitos Cerne; copiar specs
entre layouts, criando duas fontes de verdade; links simbólicos, não portáveis.

## Decisão 3: Persistir somente a intenção do workflow

**Decision**: Adicionar ao manifesto versão 1 o objeto opcional `workflow` com apenas `provider`.
Ausência equivale ao modo legado. O domínio trata o valor como identificador opaco; o adapter
decide se ele é suportado e devolve sua descrição. Disponibilidade, versão instalada e conclusão
são inferidas em tempo de execução e pela estrutura persistente.

**Rationale**: `installed` e versão observada envelhecem quando o ambiente muda. Um campo opcional
é compatível com leitores antigos, e `cerne link` já preserva propriedades desconhecidas.

**Alternatives considered**: Incrementar a versão do manifesto, desnecessário para campo opcional;
persistir `pending/ready`, que pode divergir do filesystem; armazenar versão observada.

## Decisão 4: Usar o init oficial neutro de cada provider

**Decision**: Para Spec Kit, executar em `knowledge` `specify init --here --force --integration
generic --integration-options="--commands-dir .specify/commands" --ignore-agent-tools` com variante
de script explícita por plataforma. Para OpenSpec, executar `openspec init <knowledge> --tools none
--profile core --no-animation`. Executável, argumentos, variante de script, raiz própria, marcador
e caminho de specs ficam exclusivamente em `internal/workflowexec`; `internal/workspace` recebe uma
descrição genérica e uma função de setup.

**Rationale**: Spec Kit não oferece init sem integração; `generic` em diretório próprio é sua opção
menos acoplada. OpenSpec oferece `--tools none`. Ambas evitam prompt e não selecionam agente
específico. Assets de bootstrap são empacotados localmente e nenhuma extensão ou preset remoto é
solicitado.

**Alternatives considered**: Omitir integração do Spec Kit, que pergunta ou assume Copilot;
escolher Codex, contrariando neutralidade; replicar templates no Cerne; permitir interação.

**Official references**:

- Spec Kit core: <https://github.github.com/spec-kit/reference/core.html>
- Spec Kit integrations: <https://github.github.com/spec-kit/reference/integrations.html>
- OpenSpec CLI: <https://github.com/Fission-AI/OpenSpec/blob/main/docs/cli.md>
- OpenSpec init: <https://github.com/Fission-AI/OpenSpec/blob/main/src/core/init.ts>

## Decisão 5: Desabilitar efeitos remotos e reduzir o ambiente

**Decision**: Executar por caminho absoluto, sem shell, copiando somente esta allowlist, com nomes
comparados de forma adequada à plataforma:

- todas as plataformas: `PATH`, `TMPDIR`, `TMP`, `TEMP`, `LANG`, `LC_ALL`, `LC_CTYPE`;
- Unix: `HOME`;
- Windows: `SystemRoot`, `WINDIR`, `ComSpec`, `PATHEXT`, `USERPROFILE`, `APPDATA`, `LOCALAPPDATA`;
- valores definidos pelo Cerne para OpenSpec: `OPENSPEC_TELEMETRY=0`, `DO_NOT_TRACK=1` e
  `NO_COLOR=1`.

Toda variável não listada é removida, incluindo opções de runtime e variáveis com tokens, chaves,
senhas, secrets ou credenciais. Nenhum provider recebe opção de update, preset remoto ou instalação.

**Rationale**: OpenSpec habilita telemetria e pode escrever configuração global na primeira
execução. Desativá-la evita rede e efeito fora do workspace. A lista mantém resolução do
executável/runtime, diretórios temporários, locale e variáveis básicas de usuário/sistema exigidas
pelas plataformas sem herdar credenciais ou opções de injeção dos runtimes.

**Alternatives considered**: Herdar todo ambiente; sandbox específico de SO, incompatível com a
portabilidade; confiar apenas na documentação sem neutralizar telemetria conhecida.

**Official references**:

- OpenSpec telemetry: <https://github.com/Fission-AI/OpenSpec/blob/main/src/telemetry/index.ts>
- OpenSpec config: <https://github.com/Fission-AI/OpenSpec/blob/main/src/telemetry/config.ts>

## Decisão 6: Não impor faixa de versão no manifesto

**Decision**: Descobrir `specify` ou `openspec` com PATH e executar a superfície de argumentos
testada. Opção desconhecida ou capacidade ausente é falha operacional com correção para instalar
versão compatível; nenhuma faixa semver ou versão observada é persistida.

**Rationale**: OpenSpec não publica política estável de versão mínima para esse init. Uma sondagem
adicional não elimina a necessidade de tratar falha do comando real.

**Alternatives considered**: Fixar versões, transferindo manutenção ao Cerne; checks de update com
rede; armazenar versão observada, que não representa o ambiente atual.

## Decisão 7: Ausência é warning; execução falha é erro recuperável

**Decision**: Executável ausente durante init mantém workspace e preferência, retorna zero e avisa
em stderr. Se um provider realmente executado falhar, o comando retorna um, preserva workspace,
preferência e auditoria, limpa apenas raízes novas conhecidas e deixa setup pendente. O comando
`cerne workflow setup` retoma ambos os casos.

**Rationale**: Ausência é condição opcional prevista. Depois de iniciar processo externo, rollback
total apagaria o registro constitucionalmente obrigatório. Preservar o workspace base permite
diagnóstico e retomada sem aceitar artefatos parciais.

**Alternatives considered**: Falhar antes de criar quando ausente; retornar zero para falha real;
rollback total pós-execução; manter arquivos parciais como concluídos.

## Decisão 8: Auditoria é criada antes do subprocesso

**Decision**: Criar exatamente um registro em `knowledge/runs` no estado iniciado antes de chamar o
provider. Se a criação falhar, não executar. Ao término, atualizar o mesmo registro atomicamente
com provider, operação, contexto knowledge, resultado e timestamps. Não persistir ambiente nem
stdout/stderr externos.

**Rationale**: Toda execução real mantém rastro mesmo em falha ou interrupção. A saída externa pode
conter informação sensível e não é necessária para reconstruir a ação pedida e seu resultado.

**Alternatives considered**: Registrar somente depois; usar logs globais; armazenar output completo.

## Decisão 9: Limpeza conservadora e idempotência por marcador

**Decision**: Antes da execução, classificar o layout descrito pelo adapter como ausente, válido ou
parcial. Válido torna setup no-op. Parcial é recusado com correção, sem `--force` destrutivo.
Ausente permite executar e, em falha, remover somente a raiz própria recebida que não existia —
`.specify` e `openspec` nas duas descrições desta versão. Arquivos comuns e preexistentes nunca são
removidos.

**Rationale**: Ambos os inits podem deixar arquivos em falhas e não prometem transação completa no
diretório existente. Limitar ownership a uma raiz conhecida evita um motor genérico de snapshots.

**Alternatives considered**: Snapshot de todo knowledge; staging que produz paths errados;
reexecutar sobre layout parcial; ignorar parcial.

## Decisão 10: Validar arquivos persistentes, não diretórios vazios

**Decision**: Cada adapter informa seu marcador persistente: arquivos essenciais em `.specify`
para Spec Kit e `openspec/config.yaml` com configuração reconhecível para OpenSpec. O domínio
valida o marcador recebido sem ramificações por provider. Diretórios vazios de specs/changes não
são exigidos pelo doctor.

**Rationale**: OpenSpec cria diretórios vazios que não sobrevivem a Git. Um clone saudável deve ser
diagnosticado por arquivos versionáveis, sem placeholders alheios.

**Alternatives considered**: Exigir a árvore física inicial; criar placeholders Cerne; aceitar
somente existência da pasta raiz.

## Decisão 11: Reusar localização e contratos existentes

**Decision**: `cerne workflow setup` localiza o ancestral mais próximo com
`knowledge/cerne.json`, como status/link. CLI mantém parser manual e códigos zero, um e dois.

**Rationale**: Reuso reduz regras divergentes e dependências. A sintaxe é pequena demais para
framework ou registry.

**Alternatives considered**: Exigir raiz; argumento de workspace; framework CLI.
