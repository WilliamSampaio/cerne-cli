# Cerne

[![Pruebas](https://github.com/WilliamSampaio/cerne-cli/actions/workflows/test.yml/badge.svg)](https://github.com/WilliamSampaio/cerne-cli/actions/workflows/test.yml)
[![Licencia: MIT](https://img.shields.io/badge/licencia-MIT-blue.svg)](LICENSE)

[English](README.md) · [Português (Brasil)](README.pt-BR.md) · **Español**

Cerne es una CLI open source y multiplataforma, escrita en Go, para administrar espacios de trabajo
de software formados por dos repositorios Git independientes:

- **knowledge** — intención del proyecto, información del producto, especificaciones, decisiones,
  políticas y registros de ejecución; normalmente privado;
- **source** — código fuente de la aplicación.

El nombre *Cerne* significa “núcleo”. El proyecto comienza con la administración local y segura del
workspace y fue diseñado para evolucionar hacia un harness independiente de modelos y proveedores,
capaz de coordinar agentes de IA en documentación, producto, implementación, validación y
mantenimiento.

## ¿Por qué Cerne?

Cerne sigue algunas reglas duraderas:

- tu conocimiento te pertenece y permanece accesible como archivos comunes e historial Git;
- el conocimiento privado y el código de la aplicación permanecen en repositorios separados;
- las integraciones se implementan mediante adaptadores, sin contaminar el dominio;
- el trabajo automatizado debe ser trazable y recibir solo el contexto necesario;
- push, merge, publicación, despliegue y operaciones destructivas requieren autorización explícita;
- los secretos y credenciales nunca deben almacenarse en los repositorios administrados.

La versión actual es deliberadamente local. No llama agentes de IA, accede a GitHub, clona remotos,
publica ni despliega nada.

## Requisitos

- Git disponible en `PATH`;
- Go 1.24.6 o posterior para compilar el proyecto;
- Linux, Windows o macOS.

## Instalación

Instala directamente con Go:

```sh
go install github.com/WilliamSampaio/cerne-cli/cmd/cerne@latest
cerne --version
cerne --help
```

Go coloca el binario en `GOBIN` o en `GOPATH/bin` cuando `GOBIN` no está definido. Asegúrate de que
ese directorio esté en `PATH`.

Para compilar una copia de desarrollo:

```sh
git clone https://github.com/WilliamSampaio/cerne-cli.git
cd cerne-cli
go build -o cerne ./cmd/cerne
./cerne --version
```

En Windows, el binario generado es `cerne.exe`.

## Inicio rápido

### 1. Crea un workspace

```sh
cerne init mi-proyecto
cd mi-proyecto
```

Cerne crea:

```text
mi-proyecto/
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

Los dos repositorios son locales, independientes y comienzan sin commits ni remotos. La raíz del
workspace no es un repositorio Git.

Git no versiona directorios vacíos. Añade conocimiento del proyecto antes del primer commit de
knowledge para preservar los directorios necesarios; Cerne no crea placeholders ni commits
automáticos de forma intencional.

### 2. Valida la estructura

Ejecuta desde la raíz del workspace:

```sh
cerne doctor
```

El informe identifica cada verificación con `✓` (aprobada), `!` (advertencia) o `✗` (error
bloqueante).

### 3. Consulta el estado local de Git

```sh
cerne status
```

El comando muestra branch, commit abreviado, estado del árbol, archivos en stage, modificados y no
rastreados en ambos repositorios. Los cambios pendientes son información, no un error.

### 4. Vincula un source existente (opcional)

`init` ya configura el repositorio `source` vacío. Usa `--replace` para apuntar el manifiesto a otro
repositorio local:

```sh
cerne link ../aplicacion-existente --replace
```

Solo cambia la referencia del manifiesto. Cerne nunca copia, mueve, limpia, hace checkout, commit o
elimina el source anterior ni el nuevo.

## Manifiesto

El archivo `knowledge/cerne.json` identifica el proyecto y localiza el repositorio source:

```json
{
  "name": "mi-proyecto",
  "source": "../source"
}
```

La ausencia de `version` representa la versión 1 del manifiesto. Cuando está presente, el único
valor admitido actualmente es el entero JSON `1`. Cerne almacena una ruta source relativa y
normalizada siempre que las plataformas y ubicaciones lo permitan.

## Referencia de comandos

### Opciones globales

- `cerne --help` muestra los comandos disponibles y las opciones globales.
- `cerne --version` muestra el identificador SemVer estable, actualmente `cerne 0.1.0`.

### `cerne init <project-name>`

Crea un workspace debajo del directorio actual. El destino debe estar ausente o ser un directorio
normal vacío; los enlaces simbólicos y destinos no vacíos se rechazan. El contenido existente nunca
se reemplaza. Si ocurre un fallo, Cerne revierte solo los artefactos creados por ese intento.

El nombre utiliza entre 1 y 255 caracteres ASCII, comienza con una letra o número y puede continuar
con letras, números, `.`, `_` o `-`. Se rechazan los nombres reservados de Windows y los terminados
en `.`.

### `cerne doctor`

Ejecuta diez verificaciones de solo lectura desde la raíz: manifiesto, directorios de ambos
repositorios, independencia Git, aislamiento de versionado, rutas del manifiesto, directorios
obligatorios de conocimiento, Git, permisos y versión del manifiesto. Nunca repara el workspace.

### `cerne status`

Localiza el workspace ancestro más cercano desde el directorio actual y lee ambos repositorios.
Reconoce árboles limpios o con cambios, detached HEAD y repositorios sin commits. No ejecuta fetch
ni compara con remotos.

### `cerne link <ruta> [--replace]`

Vincula como `source` un repositorio Git local no bare con árbol de trabajo. Acepta rutas relativas,
absolutas y worktrees válidos. Knowledge y source deben ser distintos y no pueden estar anidados de
forma peligrosa. Cambiar un source ya configurado requiere `--replace`; vincular el mismo source
termina sin reescribir el manifiesto. La sustitución del manifiesto es atómica.

Usa `<comando> --help` para consultar el contrato completo. La salida de la CLI está actualmente en
portugués.

## Códigos de salida y streams

| Código | Significado |
| --- | --- |
| `0` | Éxito, ayuda, workspace saludable, solo advertencias o estado pendiente consultado con éxito |
| `1` | Fallo operacional o error bloqueante encontrado por `doctor` |
| `2` | Uso inválido del comando o nombre de proyecto inválido |

La salida normal y la ayuda usan stdout. Los errores de uso y fallos operacionales usan stderr. Los
informes de `doctor`, incluidos los errores bloqueantes, usan stdout para mantener el diagnóstico en
un único stream.

## Seguridad y privacidad

- `doctor` y `status` son de solo lectura.
- `link` actualiza únicamente `knowledge/cerne.json` después de completar todas las validaciones.
- La inspección Git desactiva locks opcionales y prompts y elimina variables `GIT_*` capaces de
  redirigir los procesos hijos.
- Ningún comando actual accede a remotos ni necesita credenciales.
- No guardes tokens, contraseñas, claves privadas u otros secretos en los repositorios administrados.

## Diseño técnico

El código mantiene responsabilidades pequeñas y explícitas:

```text
cmd/cerne/          argumentos, salida del terminal y códigos de salida
internal/workspace/ reglas de dominio y operaciones del workspace
internal/gitexec/   adaptador para el ejecutable Git local
internal/filecheck/ verificaciones multiplataforma de permisos
specs/              especificaciones, planes, contratos y tareas
```

La implementación prefiere la biblioteca estándar de Go. Los comportamientos específicos del
sistema de archivos se aíslan con build tags. CI ejecuta las pruebas en Linux, Windows y macOS. El
dominio está separado de la impresión en terminal para poder reutilizarlo en futuras interfaces.

## Desarrollo

```sh
go build -o cerne ./cmd/cerne
go test ./...
go test -count=1 ./...
go vet ./...
gofmt -w <archivos-go-modificados>
```

Las pruebas usan el paquete `testing`, directorios temporales y solamente repositorios Git locales.
No necesitan red ni credenciales.

## Cómo contribuir

Las contribuciones son bienvenidas:

1. Abre un issue o discute el comportamiento antes de un cambio grande.
2. Crea una branch enfocada y mantén las reglas de dominio fuera de la impresión del terminal.
3. Añade o actualiza una prueba que falle sin el cambio propuesto.
4. Ejecuta `gofmt`, `go vet ./...` y `go test -count=1 ./...`.
5. Abre un pull request explicando el objetivo, el issue o artefacto en `specs/`, los comandos de
   validación y cualquier impacto de compatibilidad en la CLI.

Usa asuntos cortos al estilo Conventional Commits, como `feat: add command` o
`fix: preserve manifest`. Consulta [AGENTS.md](AGENTS.md) para las reglas de contribución y la
[constitución del proyecto](.specify/memory/constitution.md) para gobernanza y compatibilidad.

## Roadmap y alcance

El alcance actual incluye creación y validación del workspace, estado local y vínculo de un
repositorio source local existente. En el futuro, Cerne podrá coordinar agentes auditables para
producto, documentación, implementación, validación y mantenimiento, sin depender de modelos,
agentes o proveedores específicos.

La administración de repositorios remotos, commits automáticos, push, pull requests, merge,
publicación, despliegue, interfaz gráfica, salida JSON y ejecución de IA no forman parte de la CLI
actual.

## Licencia

Cerne se distribuye bajo la [Licencia MIT](LICENSE).
