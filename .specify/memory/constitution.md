<!--
Sync Impact Report
- Version change: 1.0.0 → 2.0.0
- Modified principles:
  - I. Specification Is the Source of Truth → I. O Conhecimento Pertence ao Usuário
  - II. Stable CLI Contracts → II. Conhecimento e Código Permanecem Separados
  - III. Simplicity First → III. Neutralidade de IA
  - IV. Proportionate Verification → IV. Integrações por Adaptadores
  - V. Safe, Explicit Failure → V. Contexto Mínimo para Agentes
- Added principles:
  - VI. Execuções Rastreáveis e Auditáveis
  - VII. Operações Sensíveis Exigem Autorização
  - VIII. Segredos Nunca Entram nos Repositórios
  - IX. Portabilidade entre Sistemas Operacionais
  - X. Domínio Protegido por Testes
  - XI. CLI Simples, Previsível e Automatizável
  - XII. Somente a Complexidade Necessária
- Added sections:
  - Missão e Escopo
  - Padrões Mínimos de Testes
  - Compatibilidade e Versionamento
  - Documentação de Comandos
- Removed sections:
  - Engineering Constraints (requirements redistributed)
- Templates:
  - ✅ updated: .specify/templates/plan-template.md
  - ✅ updated: .specify/templates/spec-template.md
  - ✅ updated: .specify/templates/tasks-template.md
  - ✅ compatible, no change: .specify/templates/checklist-template.md
- Runtime guidance:
  - ✅ reviewed, no change: README.md
  - ✅ updated: .agents/skills/speckit-tasks/SKILL.md
  - ✅ updated: .agents/skills/speckit-specify/SKILL.md
  - ✅ reviewed, no change: remaining .agents/skills/speckit-*/SKILL.md
- Follow-up TODOs: None
-->
# Constituição do Cerne

## Core Principles

### I. O Conhecimento Pertence ao Usuário
Todo conhecimento administrado pelo Cerne MUST permanecer sob controle do usuário. O Cerne MUST
permitir que o usuário inspecione, versione, exporte e remova esse conteúdo sem dependência de um
serviço proprietário. Formatos e fluxos MUST evitar aprisionamento tecnológico.

### II. Conhecimento e Código Permanecem Separados
Cada workspace MUST associar dois repositórios Git independentes: um repositório de conhecimento,
normalmente privado, e um repositório de código-fonte. O Cerne MUST preservar históricos, remotos,
permissões e ciclos de vida separados. Nenhuma automação MAY copiar conhecimento privado para o
repositório de código sem autorização explícita.

### III. Neutralidade de IA
O núcleo do Cerne MUST ser independente de modelos, agentes, protocolos e fornecedores de IA
específicos. Regras de domínio MUST operar sem exigir um provedor de IA. A troca de fornecedor
MUST NOT exigir alteração das entidades ou políticas centrais.

### IV. Integrações por Adaptadores
Git hosts, modelos, agentes, sistemas de arquivos remotos e demais serviços externos MUST entrar
por adaptadores com contratos definidos pelo núcleo. Código específico de fornecedor MUST NOT
contaminar entidades, casos de uso ou políticas do domínio. Cada adaptador MUST poder ser
substituído ou omitido sem reescrever o núcleo.

### V. Contexto Mínimo para Agentes
Cada agente MUST receber somente os arquivos, trechos, metadados e credenciais estritamente
necessários à tarefa autorizada. A seleção de contexto MUST ser explícita e rastreável. O Cerne
MUST NOT compartilhar um workspace completo por conveniência quando um escopo menor for
suficiente.

### VI. Execuções Rastreáveis e Auditáveis
Toda execução automatizada MUST registrar identidade do executor, tarefa, contexto selecionado,
ações solicitadas, decisões de autorização, resultado e timestamps. Registros MUST permitir
reconstruir o que ocorreu sem expor segredos ou conteúdo desnecessário. Falhas de auditoria MUST
impedir operações sensíveis de prosseguir.

### VII. Operações Sensíveis Exigem Autorização
Push, merge, publicação, deploy e operações destrutivas MUST exigir autorização explícita e
específica do usuário antes da execução. A solicitação MUST identificar operação, destino e
impacto. Autorizações genéricas, presumidas ou obtidas para outra operação MUST NOT ser
reutilizadas.

### VIII. Segredos Nunca Entram nos Repositórios
Segredos, tokens, chaves e credenciais MUST NOT ser gravados nos repositórios administrados pelo
Cerne, em commits, artefatos de auditoria ou logs. Credenciais MUST vir de mecanismos externos
apropriados ao sistema operacional ou ao ambiente de execução e MUST permanecer mascaradas em
diagnósticos.

### IX. Portabilidade entre Sistemas Operacionais
O CLI MUST oferecer comportamento funcional consistente em Linux, Windows e macOS. Caminhos,
processos, permissões, finais de linha e invocações de shell MUST usar abstrações portáveis. Uma
funcionalidade limitada por plataforma MUST ser documentada, detectada e recusada com diagnóstico
claro.

### X. Domínio Protegido por Testes
Todo comportamento do domínio MUST possuir testes automatizados determinísticos. Correções de
defeitos MUST incluir teste de regressão. Adaptadores MUST possuir testes de contrato e fluxos
críticos do CLI MUST possuir testes de integração cobrindo saída e código de status.

### XI. CLI Simples, Previsível e Automatizável
Comandos MUST usar nomes, argumentos, flags, stdout, stderr e códigos de saída consistentes.
Execuções não interativas MUST ser suportadas quando a operação puder ser autorizada previamente.
Saída destinada a scripts MUST possuir formato estável e não depender de texto decorativo.

### XII. Somente a Complexidade Necessária
O projeto MUST implementar apenas abstrações exigidas por requisitos atuais. Código existente,
biblioteca padrão do Go e recursos nativos da plataforma MUST ser preferidos antes de novas
dependências. Extensões especulativas, interfaces com uma única necessidade hipotética e
configurações sem caso de uso atual são proibidas.

## Missão e Escopo

O Cerne é um CLI open source escrito em Go e distribuído sob a licença MIT. Seu objetivo inicial
é administrar workspaces formados por um repositório Git de conhecimento e um repositório Git de
código-fonte independentes.

O projeto poderá evoluir para um harness que coordena agentes de IA em atividades de documentação,
produto, implementação, validação e manutenção de software. Essa evolução MUST preservar todos os
princípios desta constituição; a visão futura não autoriza abstrações sem requisito atual.

## Padrões Mínimos de Testes

- Regras, entidades e casos de uso do domínio MUST possuir testes unitários.
- Toda mudança de comportamento e toda correção de defeito MUST incluir um teste que falhe sem a
  mudança.
- Adaptadores MUST possuir testes de contrato para o limite que implementam.
- Comandos críticos MUST possuir testes de integração para argumentos, stdout, stderr, códigos de
  saída e efeitos observáveis.
- Fluxos Git MUST usar repositórios temporários ou simulados e MUST NOT depender de remotos reais.
- Autorização, segredos e operações destrutivas MUST incluir cenários negativos e de recusa.
- Releases MUST passar pelos testes aplicáveis em Linux, Windows e macOS.
- Testes MUST ser determinísticos, isolados e executáveis sem credenciais reais.

## Compatibilidade e Versionamento

Releases do Cerne MUST seguir `MAJOR.MINOR.PATCH`. PATCH corrige defeitos sem alterar contratos
públicos. MINOR adiciona funcionalidade compatível. MAJOR permite alterações incompatíveis e MUST
incluir guia de migração.

São contratos públicos: nomes de comandos e flags, valores padrão, formatos de entrada e saída
para automação, códigos de saída, arquivos de configuração, metadados de workspace e
comportamentos documentados. Dentro da mesma versão MAJOR, esses contratos MUST permanecer
compatíveis. Funcionalidade experimental MAY quebrar compatibilidade apenas quando estiver
explicitamente identificada como experimental na documentação e na saída de ajuda.

Constituem alterações incompatíveis: remover ou renomear comando ou flag; mudar semântica ou valor
padrão; alterar saída consumida por scripts ou códigos de status; exigir migração de configuração
ou workspace; reduzir plataformas suportadas; ou ampliar permissões e efeitos colaterais de uma
operação existente. Toda alteração incompatível MUST ter justificativa, versão MAJOR, aviso de
depreciação em pelo menos uma release MINOR anterior e instruções de migração. O aviso prévio MAY
ser omitido somente para corrigir vulnerabilidade ativa ou risco imediato de perda de dados; a
exceção MUST ser registrada nas notas da release.

## Documentação de Comandos

Todo comando público MUST documentar finalidade, sintaxe, argumentos, flags e valores padrão,
entradas, stdout, stderr, códigos de saída, efeitos colaterais, autorizações exigidas e exemplos
executáveis. Restrições de plataforma, formatos estáveis para automação e impactos de
compatibilidade MUST ser explícitos. A documentação e o texto de ajuda MUST ser atualizados na
mesma mudança que altera o comando.

## Development Workflow

Cada mudança MUST partir de requisitos testáveis e manter rastreabilidade entre especificação,
plano, tarefas, testes e implementação. O Constitution Check MUST passar antes da pesquisa e após
o design. Revisões MUST verificar limites entre repositórios, neutralidade de IA, adaptadores,
contexto mínimo, auditoria, autorização, segredos, portabilidade, testes, contratos do CLI,
documentação e simplicidade.

## Governance

Esta constituição prevalece sobre orientações conflitantes do projeto. Uma alteração MUST ser
proposta com motivação, princípios e artefatos afetados, impacto de compatibilidade, plano de
migração e atualização dos templates dependentes. A alteração MUST ser aprovada por um mantenedor
em revisão e registrada no Sync Impact Report.

A constituição usa versionamento semântico próprio. MAJOR remove ou redefine obrigações existentes;
MINOR adiciona princípios, seções ou obrigações materiais; PATCH esclarece texto sem mudar
obrigações. A data de ratificação original MUST ser preservada e a data da última alteração MUST
ser atualizada. Violações de MUST bloqueiam implementação e release; exceções exigem primeiro uma
emenda constitucional aprovada.

**Version**: 2.0.0 | **Ratified**: 2026-07-28 | **Last Amended**: 2026-07-28
