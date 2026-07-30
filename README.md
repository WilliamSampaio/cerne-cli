# Cerne

Cerne é um CLI open source em Go, sob licença MIT, para administrar workspaces formados por dois
repositórios Git independentes: conhecimento e código-fonte.

## `cerne init`

Inicializa um workspace local:

```text
cerne init <project-name>
cerne init --help
```

`<project-name>` deve ter de 1 a 255 caracteres ASCII. O primeiro deve ser uma letra ou número; os
demais podem ser letras, números, `.`, `_` ou `-`. O nome não pode terminar em ponto nem usar nomes
reservados do Windows, inclusive antes de uma extensão: `CON`, `PRN`, `AUX`, `NUL`, `COM1`–`COM9`
ou `LPT1`–`LPT9`.

O destino é `<diretório-atual>/<project-name>` e precisa estar ausente ou ser um diretório regular
vazio, nunca um link. Nenhum arquivo existente é substituído ou removido.

```text
<project-name>/
├── knowledge/
│   ├── .git/
│   ├── cerne.json
│   ├── product/
│   ├── specs/
│   ├── decisions/
│   ├── policies/
│   └── runs/
└── source/
    └── .git/
```

Os dois diretórios são repositórios Git locais independentes, sem commit ou remoto. O manifesto
`knowledge/cerne.json` identifica o projeto e aponta `source` para `../source`. A raiz do workspace
não é inicializada como repositório.

O comando não usa rede, credenciais ou agentes; não clona, publica, faz push, merge ou deploy. A
invocação autoriza apenas a criação local acima. Se ocorrer uma falha parcial, o Cerne desfaz
somente os artefatos criados pela própria tentativa.

### Saída e status

Em sucesso, o status é `0`, stderr fica vazio e stdout contém:

```text
Workspace "<project-name>" criado.
Knowledge: <absolute-knowledge-path>
Source: <absolute-source-path>
```

Ajuda também usa stdout e status `0`. Falhas operacionais — destino inseguro, permissão, Git
indisponível ou erro de criação — usam stderr, status `1` e informam causa e correção. Uso ou nome
inválido usa stderr, status `2` e inclui `uso: cerne init <project-name>`. Não há prompt, cor ou
texto decorativo.

Exemplo:

```text
cerne init exemplo
```

Para compilar e testar:

```text
go build -o cerne ./cmd/cerne
go test ./...
```

Git deve estar disponível em `PATH`. O comportamento é suportado em Linux, Windows e macOS.
