# Research: Contexto Estrutural do Workspace

## 1. Relatório próprio em vez de adaptação do Doctor

**Decision**: criar um agregador `ContextReport` somente filesystem e reutilizar apenas helpers
puros de manifesto, paths e layout.

**Rationale**: `Doctor` consulta Git, permissões e disponibilidade de executável, aplica fallbacks e
falha com uma taxonomia diferente. Filtrar seus resultados ainda executaria verificações proibidas
e acoplaria o schema público a mensagens internas.

**Alternatives considered**: chamar `DoctorWithWorkflow` e filtrar checks; converter `Diagnosis`;
reutilizar `CurrentStatus`. Todas foram rejeitadas por efeitos, fail-fast ou fatos inventados.

## 2. Descoberta segura do workspace parcial

**Decision**: localizar a primeira raiz ancestral com evidência Cerne (`knowledge` ou
`knowledge/cerne.json`) e parar nessa fronteira, mesmo que parcial. Um source externo nunca inicia
descoberta reversa.

**Rationale**: atravessar um workspace parcial e selecionar um ancestral válido pode entregar à IA
o projeto errado. O localizador atual perde a raiz quando o manifesto falta e não deve mudar porque
outros comandos dependem de seu contrato fail-fast.

**Alternatives considered**: reutilizar `locateWorkspace` sem mudança; refatorar todos os
localizadores agora; manter busca acima de candidato parcial. Rejeitadas por segurança ou escopo.

## 3. Gating de fatos e diagnóstico parcial

**Decision**: publicar somente objetos comprovados e validar em ordem fixa. Root pode existir sem
manifesto; manifesto inválido não gera source fictício; versão não suportada bloqueia todos os
campos derivados do manifesto; falhas independentes continuam sendo coletadas.

**Rationale**: dependências explícitas evitam cascatas e impedem que semântica de schema futuro seja
tratada como conhecida. A lista determinística atende humanos e consumidores automatizados.

**Alternatives considered**: fail-fast; fallbacks como `root/source`; publicar dados parcialmente
decodificados de versão futura. Rejeitadas por FR-019 e risco de contexto incorreto.

## 4. JSON tipado e determinístico

**Decision**: usar structs públicas dedicadas, campos opcionais com `omitempty`, `problems`
inicializado como slice vazio e `json.Encoder` com indentação de dois espaços e newline final.

**Rationale**: `encoding/json` preserva a ordem de campos de structs e evita dependência nova. Uma
única montagem do relatório alimenta as renderizações JSON e humana.

**Alternatives considered**: `map[string]any`, template JSON e biblioteca de schema. Rejeitadas por
ordem menos clara, risco de escaping e complexidade desnecessária.

## 5. Descrição estática de workflow

**Decision**: extrair `workflowexec.Describe(provider)`, pura e baseada na tabela estática já
presente em `Resolve`. `Resolve` passa a compor essa descrição com `exec.LookPath` e setup.

**Rationale**: contexto precisa conhecer paths normativos de Spec Kit/OpenSpec sem verificar se o
executável está instalado. A extração elimina duplicação e preserva comandos operacionais.

**Alternatives considered**: chamar `Resolve` com PATH vazio; duplicar a tabela no domínio; criar
registro configurável de providers. Rejeitadas por efeito global, divergência e YAGNI.

## 6. Estados e specs normalizados

**Decision**: mapear ausência de workflow para `not-declared`, raiz conhecida ausente para
`pending`, layout completo com specs regular para `ready`, estrutura parcial para `invalid` e
provider não descrito para `unknown-provider`. Specs é `knowledge/specs` para ausência/Spec Kit e
`knowledge/openspec/specs` para OpenSpec.

**Rationale**: são os cinco estados públicos da spec. `workflowSpecsValid` atual aceita por
contenção um path OpenSpec ausente; contexto adicionará a comprovação de diretório regular sem
alterar contratos anteriores.

**Alternatives considered**: expor estados internos; tratar provider indisponível como pending;
considerar specs lexicalmente contido como existente. Rejeitadas por contrato ou falta de prova.

## 7. Paths, symlinks e contenção

**Decision**: canonicalizar somente depois de comprovar tipo/existência, rejeitar symlinks nos
artefatos governados e calcular `inside_workspace` após canonicalização com os helpers existentes.
Um alias de invocação é aceito e resolvido para a árvore física.

**Rationale**: `filepath` mantém representação nativa e a validação física evita escapar por
symlink. Rejeitar aliases legítimos acima do workspace seria frágil entre plataformas.

**Alternatives considered**: paths lexicais; rejeitar qualquer symlink desde a raiz do volume;
resolver source externo para um workspace. Rejeitadas por segurança, portabilidade ou FR-003.

## 8. Sem conteúdo, Git, auditoria ou ação corretiva

**Decision**: limitar leituras a metadados de filesystem, manifesto e estrutura mínima do workflow.
Não ler coleções, instruções de agente, repositórios Git, ambiente ou rede; não gravar auditoria.

**Rationale**: o comando descreve uma consulta local, não uma execução automatizada. Skills fazem
seleção progressiva e auditam suas ações posteriores.

**Alternatives considered**: carregar resumo de knowledge; detectar AGENTS/CLAUDE; escolher spec
ativa; registrar toda consulta. Rejeitadas por neutralidade, mínimo contexto e escopo.

## 9. Contrato e compatibilidade

**Decision**: aceitar somente nenhum argumento, `--json` ou `--help`; usar status 0 para
healthy/warning, 1 para invalid e 2 para uso inválido. Schema v1 evolui apenas aditivamente e a
feature entra em release minor.

**Rationale**: combina automação estável com relatório humano e mantém workspaces v1 sem migração.

**Alternatives considered**: múltiplas flags combináveis, stderr para diagnóstico estrutural,
schema experimental e campo novo no manifesto. Rejeitadas por superfície desnecessária e quebra de
contrato.
