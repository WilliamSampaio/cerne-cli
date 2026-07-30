package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/WilliamSampaio/cerne-cli/internal/gitexec"
	"github.com/WilliamSampaio/cerne-cli/internal/workspace"
)

const initHelp = `Inicializa um workspace Cerne com repositórios Git independentes.

Uso:
  cerne init <project-name>

Nome:
  1 a 255 caracteres ASCII; começa por letra ou número e continua com
  letras, números, ponto, hífen ou sublinhado. Nomes reservados e ponto final
  não são aceitos.

Estrutura:
  <project-name>/
  ├── knowledge/
  │   ├── cerne.json
  │   ├── product/
  │   ├── specs/
  │   ├── decisions/
  │   ├── policies/
  │   └── runs/
  └── source/

Efeitos:
  Cria dois repositórios Git locais vazios, sem commit ou remoto.
  Não acessa a rede e não altera conteúdo existente.

Saídas:
  Sucesso e ajuda usam stdout. Erros usam stderr.
  Status 0: sucesso ou ajuda; 1: falha operacional; 2: uso ou nome inválido.

Erros:
  O destino deve estar ausente ou ser um diretório regular vazio.
  Instale Git, corrija o nome ou escolha outro destino conforme o diagnóstico.

Exemplo:
  cerne init exemplo
`

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) != 2 || args[0] != "init" {
		return usageError(stderr, "informe exatamente um nome de projeto")
	}
	if args[1] == "--help" {
		fmt.Fprint(stdout, initHelp)
		return 0
	}
	if err := workspace.ValidateName(args[1]); err != nil {
		return usageError(stderr, err.Error())
	}

	initRepository, err := gitexec.Find()
	if err != nil {
		fmt.Fprintf(stderr, "erro: %v\n", err)
		fmt.Fprintln(stderr, "correção: instale o Git e disponibilize-o no PATH")
		return 1
	}
	current, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "erro: não foi possível obter o diretório atual: %v\n", err)
		fmt.Fprintln(stderr, "correção: execute o comando em um diretório acessível")
		return 1
	}
	result, err := workspace.Init(current, args[1], initRepository)
	if err != nil {
		fmt.Fprintf(stderr, "erro: %v\n", err)
		if errors.Is(err, workspace.ErrUnsafeDestination) {
			fmt.Fprintln(stderr, "correção: escolha um destino inexistente ou vazio")
		} else {
			fmt.Fprintln(stderr, "correção: verifique permissões e tente novamente")
		}
		return 1
	}

	fmt.Fprintf(stdout, "Workspace %q criado.\nKnowledge: %s\nSource: %s\n",
		result.Name, result.KnowledgePath, result.SourcePath)
	return 0
}

func usageError(stderr io.Writer, cause string) int {
	fmt.Fprintf(stderr, "erro: %s\n", cause)
	fmt.Fprintln(stderr, "uso: cerne init <project-name>")
	return 2
}
