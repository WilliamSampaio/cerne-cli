# Quickstart: Validar Restauração de Workspace

## Prerequisites

- Binary da branch: `go build -o cerne ./cmd/cerne`.
- Git disponível no `PATH`.
- Diretório temporário vazio para cada cenário.
- Home temporária isolada para observar somente `<home>/.cerne/audit` da tentativa.
- Repositórios Git locais descartáveis; nenhuma rede, credencial ou remoto real.

Use `<cerne>` como path absoluto do binário. Os fixtures de knowledge precisam conter
`cerne.json`, diretórios obrigatórios e commits Git. Consulte [o contrato do comando](contracts/restore-command.md)
e [o contrato da auditoria](contracts/restore-audit-record.md) para saídas e campos exatos.

## Scenario 1: Restore com dois clones locais

Crie origens locais independentes e execute no parent vazio:

```text
<cerne> restore <knowledge-origin> --clone <source-origin>
```

Expected: status zero, três linhas no stdout, stderr vazio, root nomeado pelo manifesto, dois Git
roots com histórico/remoto `origin`, nenhum `.git` no root, nenhuma origem no manifesto/output e um
audit `succeeded`.

## Scenario 2: Restore com source local

Registre um snapshot byte a byte de um working tree fora do parent e execute:

```text
<cerne> restore <knowledge-origin> --source <local-source>
```

Expected: source vinculado, nenhum source interno, snapshot idêntico e somente `manifest.source`
alterado quando necessário. `name`, `version`, `workflow` e campos desconhecidos permanecem iguais.

## Scenario 3: Origem local dirty não vira cópia

Adicione arquivos não rastreados e mudanças sem commit na origem local de knowledge.

Expected: o clone contém apenas o checkout Git obtido normalmente; origem e mudanças locais ficam
intactas e nenhum arquivo extra é copiado recursivamente.

## Scenario 4: Workflow preservado

Restaure knowledge com provider conhecido em layout pronto e, separadamente, pendente.

Expected: ambos restauram sem executar provider. Layout parcial ou provider desconhecido retorna
um, faz rollback e deixa somente audit global.

## Scenario 5: Manifesto ou nome inválido

Teste manifesto ausente, symlink, JSON malformado, conteúdo extra, version diferente de `1` e nomes
com traversal, separador, reservado Windows, Unicode ou ponto final.

Expected: status um, causa/correção seguras, destino ausente, staging removido e audit sem origem.

## Scenario 6: Path declarado para source clonado

Teste `../source` e um path portátil aninhado aceito. Depois teste absoluto Unix/Windows, volume,
backslash, `../../escape`, root, knowledge, descendente de knowledge, symlink e target concorrente.

Expected: casos válidos clonam exatamente no path declarado; casos incompatíveis falham antes do
segundo clone quando detectáveis e nunca escrevem fora do staging.

## Scenario 7: Destino preexistente ou concorrente

Antes da execução, crie `<parent>/<manifest-name>` vazio, não vazio, arquivo e symlink. Em outro
fixture, faça o destino aparecer imediatamente antes da promoção.

Expected: todos são preservados sem alteração; status um, nenhum replace/remove e audit final ou
inconclusivo conforme a falha injetada.

## Scenario 8: Auditoria bloqueia processos

Use home inacessível e depois `.cerne`/`audit` como arquivo, symlink ou diretório POSIX aberto a
grupo/outros. Injete falhas de criação, write, sync, close e transição.

Expected: falha ao iniciar/transicionar audit impede o próximo Git; nenhum workspace final existe.
Quando o arquivo inicial foi criado, ele permanece como único registro inconclusivo.

## Scenario 9: Falha em cada clone e validação

Use Git falso que cria conteúdo parcial e falha no clone de knowledge, clone de source, inspeção e
validação final. Inclua output com URL e token fictícios.

Expected: status um, stdout vazio, root/staging ausentes, exatamente um audit e nenhum valor/output
sensível em streams ou JSON.

## Scenario 10: Finalização da auditoria

Faça a montagem e promoção concluírem, mas falhe a escrita final do audit.

Expected: status um, root promovido removido após confirmação de identidade e audit no último
estado durável inconclusivo. Substituição concorrente do root torna limpeza ambígua e deve ser
recusada, nunca removida pelo nome.

## Scenario 11: Source local permanece imutável

Falhe cada etapa depois da primeira inspeção e compare snapshots. Troque também os fatos Git do
source entre inspeções e teste source contendo o parent ou a home/auditoria.

Expected: source permanece byte a byte idêntico; mudança concorrente falha; sobreposição é recusada
no preflight sem processo ou audit.

## Scenario 12: Contrato CLI e redaction

Teste help, origem ausente, flag ausente, ambas as flags, flag repetida, valor ausente/option-like,
ordem diferente, extras, transporte recusado e credencial/query/fragmento embutido.

Expected: help/status zero em stdout; usos/origens inválidos status dois em stderr; nenhum efeito.
As mensagens exatas seguem o contrato e nunca repetem a entrada.

## Scenario 13: Compatibilidade dos workspaces

Nos restores válidos, execute `cerne doctor`, `cerne status`, `cerne link` e, quando declarado,
`cerne workflow setup` separadamente. Rode também a suíte dos comandos anteriores.

Expected: workspaces restaurados são aceitos e contratos existentes permanecem sem migração.

## Scenario 14: Portabilidade

Execute os testes automatizados em Linux, Windows e macOS com parent, home e sources contendo
espaços, Unicode e diferenças de caixa relevantes ao sistema.

Expected: resultados funcionais equivalentes, processos sem shell e nenhum pressuposto de
separador/volume de uma única plataforma.

## Automated validation

```text
gofmt -w <changed-go-files>
go vet ./...
go test -count=1 ./...
git diff --check
```

A CI existente já executa `go test ./...` nos três sistemas; os novos testes devem usar apenas
homes, repositórios, executáveis e diretórios temporários controlados.
