# Cerne

[![Pruebas](https://github.com/WilliamSampaio/cerne-cli/actions/workflows/test.yml/badge.svg)](https://github.com/WilliamSampaio/cerne-cli/actions/workflows/test.yml)
[![Licencia: MIT](https://img.shields.io/badge/licencia-MIT-blue.svg)](LICENSE)

[English](README.md) · [Português (Brasil)](README.pt-BR.md) · **Español**

[Documentación para usuarios](docs/es/getting-started.md)

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

La versión actual es deliberadamente local. No llama agentes de IA, administra servicios de
alojamiento, publica ni despliega. Solo accede a un origen Git con `init --clone` explícito.

## Requisitos

- Git disponible en `PATH`;
- Go 1.26.5 o posterior para compilar el proyecto;
- Linux, Windows o macOS.

El ejecutable `specify` de Spec Kit o `openspec` de OpenSpec es opcional y solo es necesario al
seleccionar ese workflow. Cerne nunca instala ni actualiza estas herramientas.
Con Spec Kit, `--agent codex|claude` también puede preparar el descubrimiento local de comandos en
la raíz del workspace; la elección del agente no se guarda en `knowledge/cerne.json`.

## Instalación

En Linux y macOS, instala el binario standalone estable más reciente sin Go:

```sh
curl --proto '=https' --tlsv1.2 -fsSL \
  https://github.com/WilliamSampaio/cerne-cli/releases/latest/download/install.sh | sh
~/.local/bin/cerne --version
```

Inspecciona primero el instalador usando el mismo `curl` sin `| sh`. Para instalar una versión
fija, usa la URL de la release de esa tag:

```sh
curl --proto '=https' --tlsv1.2 -fsSL \
  https://github.com/WilliamSampaio/cerne-cli/releases/download/vX.Y.Z/install.sh |
  sh -s -- --version vX.Y.Z
```

Para instalar las skills opcionales de un agente en el mismo flujo explícito:

```sh
curl --proto '=https' --tlsv1.2 -fsSL \
  https://github.com/WilliamSampaio/cerne-cli/releases/latest/download/install.sh |
  sh -s -- --agent codex
```

`--agent codex` y `--agent claude` instalan todas las skills oficiales compatibles. `--agent gemini`
instala solo la skill de flujo Git.

El instalador escribe solo `~/.local/bin/cerne`, verifica checksums de la release antes de
reemplazar un archivo regular, rechaza destinos que sean directorios o symlinks, nunca usa `sudo` y
nunca edita perfiles del shell. Elimínalo con `rm ~/.local/bin/cerne`. Para instalación manual,
descarga el archivo compatible y `checksums.txt` desde la release, verifica el SHA-256, extrae
`cerne` y colócalo en `PATH`.

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

Para instalar el estado local del checkout para pruebas desde todo el host, sin depender de commits
ni del remoto:

```sh
go install ./cmd/cerne
# o
make install-local
```

En Windows, el binario generado es `cerne.exe`.

## Idioma

Los mensajes de la CLI están disponibles en inglés y portugués de Brasil. Guarda una preferencia o
sustitúyela para una sola ejecución:

```sh
cerne config set language en
cerne --lang pt-BR doctor
CERNE_LANG=en cerne status
```

La precedencia es `--lang`, `CERNE_LANG`, la preferencia guardada y, por último, `en`. Las salidas
estructuradas, comandos, flags, identificadores, códigos de salida y la versión son neutros respecto
al idioma.

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
│   ├── README.md
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

`knowledge/README.md` usa el idioma efectivo `en` o `pt-BR` y explica las colecciones, los límites
entre los repositorios y los primeros comandos seguros sin depender del workflow o agente seleccionado.

Como Git no registra directorios vacíos, Cerne crea un archivo `.gitkeep` en cada directorio
obligatorio de `knowledge`. Puedes eliminarlo después de añadir contenido al directorio. Cerne no
crea commits automáticamente.

Para empezar con un source existente, elige exactamente un modo de source:

```sh
cerne init mi-proyecto --source ../aplicacion-existente
cerne init mi-proyecto --clone https://host/organizacion/aplicacion.git
```

`--source` vincula un working tree Git local non-bare, resuelve rutas relativas desde el directorio
de invocación y nunca crea un source interno ni modifica el repositorio externo. `--clone` acepta
una ruta local existente, `file`, HTTPS o SSH (incluida la sintaxis SCP-like), realiza un clon
estándar completo en el `source` interno y mantiene el remoto `origin`. `--source` y `--clone`
son mutuamente excluyentes; cualquiera puede combinarse con `--workflow`.

Para inicializar un workflow opcional de especificación durante la creación:

```sh
cerne init mi-proyecto --workflow speckit
cerne init mi-proyecto --workflow speckit --agent codex
cerne init mi-proyecto --workflow openspec
cerne init mi-proyecto --clone https://host/organizacion/aplicacion.git --workflow speckit
```

Spec Kit mantiene las especificaciones en `knowledge/specs` y controla `knowledge/.specify`.
OpenSpec usa `knowledge/openspec/specs` y controla `knowledge/openspec`, sin crear el directorio
superior `knowledge/specs`. Producto, decisiones, políticas y ejecuciones siguen siendo comunes.

Si falta el ejecutable, `init` termina correctamente, registra la elección y advierte por stderr.
Después de instalarlo, ejecuta `cerne workflow setup` desde cualquier directorio del workspace. El
setup es idempotente; se rechazan estructuras parciales o con Git anidado.
Cuando `--agent codex` o `--agent claude` se usa con Spec Kit, Cerne también pide a Spec Kit que
cree la integración correspondiente dentro de `knowledge` y escribe pequeños puentes en la raíz del
workspace en `.agents/skills` o `.claude/skills`. Esos puentes apuntan de vuelta a `knowledge` y no
contienen conocimiento privado, remotos, credenciales, dumps de entorno ni rutas absolutas. Para que
Codex descubra esas skills locales, inicia la sesión desde la raíz del workspace Cerne; una sesión
iniciada dentro de `source/` no sube hasta la raíz del workspace porque `source` es un repositorio
Git separado.

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

### 4. Evalúa una idea de producto (opcional)

Instala `cerne-product-discovery` para Codex o Claude y usa la skill antes de convertir una idea
inmadura en una especificación:

```sh
cerne skill install codex cerne-product-discovery
```

La skill cuestiona el problema, la evidencia, el alcance y el valor sin crear archivos por defecto.
Si la idea queda lista y Spec Kit está configurado, solicita una autorización nueva antes de iniciar
`speckit-specify` con el resumen del descubrimiento.

### 5. Coordina una etapa Git aprobada (opcional)

Instala `cerne-git-workflow` con `cerne skill install <agent>` y deja que la skill inspeccione
primero y pida una confirmación separada antes de que el agente ejecute cualquier efecto Git:

```sh
cerne git inspect --agent codex --task task-1 --json
```

Branch, commit, push y Pull Request son etapas separadas ejecutadas por el agente. Cerne solo
proporciona el snapshot sanitizado (`state_id`, repositorios, branches, remotes y paths cambiados). Las operaciones Git destructivas o fuera de alcance las rechaza la skill.
Consulta la [referencia de comandos](docs/es/commands.md) para el contrato completo de JSON/auditoría.
La auditoría de Cerne cubre solo la inspección; la evidencia de efectos nativos pertenece al agente o harness.

### 6. Vincula un source existente (opcional)

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

Con un workflow seleccionado, el manifiesto también contiene `"workflow":{"provider":"speckit"}`
o `"workflow":{"provider":"openspec"}`. El estado de instalación, la versión y la elección local de
agente no se guardan.

## Referencia de comandos

Consulta la [referencia completa de comandos de Cerne](docs/es/commands.md).

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
- El setup usa argumentos fijos, sin shell, recibe un entorno mínimo y desactiva la telemetría de
  OpenSpec. No recibe credenciales ni la ruta source y no registra la salida bruta del provider.
- El clon usa argumentos Git fijos sin shell, una lista de protocolos permitidos, staging privado y
  promoción sin reemplazo. El origen y la salida Git bruta no aparecen en la salida de Cerne, el
  manifiesto ni la auditoría; la autenticación sigue siendo externa y Git conserva el origen como
  remoto `origin`.
- `restore` mantiene su auditoría privada en `~/.cerne/audit`, valida ambos límites Git y usa
  rollback por identidad con promoción sin reemplazo.
- `skill install` escribe solo en el perfil del agente autorizado, valida el paquete oficial
  incorporado antes de copiar y rechaza contenido desconocido en el destino.
- Un intento fallido conserva el workspace base y la auditoría, y solo elimina una nueva raíz
  perteneciente al provider.
- La inspección Git desactiva locks opcionales y prompts y elimina variables `GIT_*` capaces de
  redirigir los procesos hijos.
- Solo `init --clone` o `restore` explícito puede acceder a un origen o usar credenciales externas.
- No guardes tokens, contraseñas, claves privadas u otros secretos en los repositorios administrados.

## Diseño técnico

El código mantiene responsabilidades pequeñas y explícitas:

```text
cmd/cerne/          argumentos, salida del terminal y códigos de salida
internal/workspace/ reglas de dominio y operaciones del workspace
internal/gitexec/   adaptador para el ejecutable Git local
internal/filecheck/ verificaciones multiplataforma de permisos
internal/workflowexec/ adaptadores para ejecutables locales opcionales de workflow
internal/skillinstall/ instalación global explícita de las skills oficiales
specs/              especificaciones, planes, contratos y tareas
```

La implementación prefiere la biblioteca estándar de Go. Los comportamientos específicos del
sistema de archivos se aíslan con build tags. CI ejecuta las pruebas en Linux, Windows y macOS. El
dominio está separado de la impresión en terminal para poder reutilizarlo en futuras interfaces.

## Desarrollo

```sh
go build -o cerne ./cmd/cerne
go install ./cmd/cerne
go test ./...
go test -count=1 ./...
go vet ./...
gofmt -w <archivos-go-modificados>
```

Hay atajos equivalentes con `make build`, `make install-local`, `make test`, `make test-fresh`,
`make vet`, `make fmt` y `make check`. Usa `make install-path` para ver dónde `make install-local`
deja disponible el ejecutable. Para instalar en otro directorio, define `GOBIN`, por ejemplo
`GOBIN=/tmp/bin make install-local`.

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
El historial de versiones está documentado en [CHANGELOG.md](CHANGELOG.md).

## Roadmap y alcance

El alcance actual incluye creación con source vacío, vinculado o clonado, bootstrap opcional de
workflow, validación, estado local, vínculo de source y descubrimiento de producto opcional para
Codex y Claude. En el futuro, Cerne podrá ampliar la coordinación auditable de agentes para
documentación, implementación, validación y mantenimiento, sin depender de modelos, agentes o
proveedores específicos.

La administración de alojamiento remoto, commits automáticos, push, pull requests, merge,
publicación, despliegue, interfaz gráfica, salida JSON y ejecución de IA no forman parte de la CLI
actual.

## Licencia

Cerne se distribuye bajo la [Licencia MIT](LICENSE).
