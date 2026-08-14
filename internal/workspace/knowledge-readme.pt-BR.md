# Conhecimento de {{PROJECT_NAME}}

Este repositório armazena o contexto durável necessário para entender e evoluir o projeto. O código
da aplicação pertence ao repositório source separado.

## 1. Valide o workspace

Execute estes comandos na raiz do workspace antes de fazer alterações:

```sh
cerne doctor
cerne status
cerne context
```

- `doctor` informa problemas estruturais e suas correções.
- `status` mostra o estado Git local de knowledge e source sem acessar remotos.
- `context` mostra os caminhos canônicos de Source, Specs, Product, Decisions, Policies e Workflow.

Resolva os problemas bloqueantes informados por `doctor` antes de adicionar conteúdo ao projeto.

## 2. Leia o relatório de contexto

Use `cerne context` como mapa deste workspace:

- **Source** é onde fica o código da aplicação. Ele pode estar dentro ou fora do workspace.
- **Specs** é o local canônico dos requisitos de features; não presuma que seja sempre `specs/`.
- **Workflow pronto** significa que o provider selecionado já pode ser usado.
- **Workflow pendente** significa que a preferência foi registrada, mas o setup está incompleto;
  quando o provider estiver instalado, execute `cerne workflow setup`.
- **Nenhum workflow declarado** significa que você pode gerenciar especificações no caminho Specs
  informado usando seu próprio processo. Um workflow gerenciado é opcional e é selecionado ao criar
  um workspace.

## 3. Escolha seu primeiro resultado útil

Você não precisa preencher todos os diretórios. Escolha o caminho adequado ao trabalho atual.

### Estruture o produto

Crie `product/overview.md` com o problema, usuários-alvo, resultado desejado, restrições conhecidas e
o que explicitamente não faz parte do escopo. Registre evidências e questões em aberto em vez de
apresentar suposições como fatos.

### Continue ou inicie a implementação

Abra o caminho Source informado por `cerne context`. Se este workspace apontar para o source local
errado, leia `cerne link --help` antes de substituir a referência; o Cerne não move nem copia o
conteúdo do source.

### Defina uma mudança

Use o caminho Specs informado por `cerne context`. Se o workflow estiver pronto, invoque seu ponto
de entrada para especificação. Caso contrário, crie um documento pequeno de requisitos que registre
valor para o usuário, limites e resultados verificáveis antes de planejar a implementação.

## 4. Registre orientações duráveis

- `decisions/` — um arquivo por decisão durável de produto ou tecnologia, incluindo justificativa e
  alternativas rejeitadas;
- `policies/` — regras gerais do projeto que trabalhos futuros devem seguir;
- `runs/` — registros sanitizados de execução e auditoria; não use como diretório geral de notas.

Prefira poucos documentos atuais a uma estrutura especulativa.

## 5. Versione cada repositório de forma independente

Knowledge e source são repositórios Git independentes, com históricos, branches, remotos e políticas
de acesso separados. Revise ambos com `cerne status`; depois faça commit ou configure remotos
explicitamente no repositório que pretende alterar. O Cerne não cria commits, pushes, merges ou
remotos para você.

## Documentação no GitHub

- [Primeiros passos](https://github.com/WilliamSampaio/cerne-cli/blob/master/docs/pt-BR/getting-started.md)
- [Referência de comandos](https://github.com/WilliamSampaio/cerne-cli/blob/master/docs/pt-BR/commands.md)
- [Inspeção Git e orientação de workflow](https://github.com/WilliamSampaio/cerne-cli/blob/master/docs/pt-BR/commands.md#cerne-git-inspect)
- [Solução de problemas](https://github.com/WilliamSampaio/cerne-cli/blob/master/docs/pt-BR/troubleshooting.md)

Não armazene segredos ou credenciais em knowledge ou source.
