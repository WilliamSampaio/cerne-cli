# Referencia de comandos

[English](../en/commands.md) · [Português (Brasil)](../pt-BR/commands.md) · **Español**

[Primeros pasos](getting-started.md) · [Solución de problemas](troubleshooting.md)

La CLI admite inglés y portugués de Brasil para los mensajes dirigidos a personas. Ejecuta
`cerne <comando> --help` para consultar el contrato completo implementado por tu versión instalada.

<!-- AUTO-GENERATED: mantener sincronizado con cmd/cerne/main.go y los contratos de la CLI. -->

## Instalador standalone

Las releases para Linux y macOS publican `install.sh`, `checksums.txt` y binarios para `amd64` y
`arm64`. El instalador acepta `--version <version>`, `--agent <codex|claude|gemini>` y `--help`.
Instala solo `~/.local/bin/cerne`, verifica SHA-256 y que el binario instalado reporte exactamente
la versión solicitada antes de promoverlo, rechaza destinos que sean directorios o symlinks, nunca
usa `sudo` y nunca edita archivos de perfil del shell.

```sh
curl --proto '=https' --tlsv1.2 -fsSL \
  https://github.com/WilliamSampaio/cerne-cli/releases/latest/download/install.sh | sh
```

## Opciones globales

- `cerne --help` muestra los comandos disponibles y las opciones globales.
- `cerne --version` muestra la versión instalada como identificador SemVer estable.
- `cerne --lang <en|pt-BR> ...` selecciona el idioma para una sola ejecución sin guardarlo.

`CERNE_LANG` ofrece la misma sustitución temporal. La precedencia es `--lang`, `CERNE_LANG`, la
preferencia guardada y, por último, `pt-BR`. El valor predeterminado actual sigue siendo `pt-BR`
por compatibilidad y cambiará a `en` en la versión 1.0. La selección solo modifica mensajes para
personas; comandos, flags, campos JSON, identificadores, códigos de salida y `--version` permanecen
estables.

## `cerne config <set language <en|pt-BR>|get language|unset language>`

Administra la preferencia de idioma del usuario actual en `~/.cerne/config.json`. `set` guarda un
idioma compatible, `get` muestra el valor guardado y `unset` elimina la preferencia para volver al
valor predeterminado de compatibilidad. Para reparar una configuración normal inválida, usa una
selección temporal, por ejemplo:

```sh
cerne --lang en config set language en
```

Cerne rechaza enlaces simbólicos y rutas de configuración inseguras en lugar de seguirlos.

## `cerne init <project-name> [--source ... | --clone ...] [--workflow ... [--agent ...]]`

Crea un workspace debajo del directorio actual. El destino debe estar ausente o ser un directorio
normal vacío; los enlaces simbólicos y destinos no vacíos se rechazan. El contenido existente nunca
se reemplaza. Si ocurre un fallo, Cerne revierte solo los artefactos creados por ese intento.

El nombre utiliza entre 1 y 255 caracteres ASCII, comienza con una letra o número y puede continuar
con letras, números, `.`, `_` o `-`. Se rechazan los nombres reservados de Windows y los terminados
en `.`.
Sin `--workflow`, el comportamiento no cambia. Con él, Cerne ejecuta el provider instalado solo en
knowledge y sin interacción. `--agent codex|claude` se acepta solo con `--workflow speckit`;
prepara descubrimiento local para esa invocación sin persistir el agente. Un ejecutable ausente
genera una advertencia; un provider ejecutado que falla conserva el workspace base y devuelve error
operativo.

`--source` valida y vincula un working tree local existente sin modificarlo. `--clone` rechaza
HTTP, `git://`, `ext::`, helpers desconocidos, valores similares a opciones, credenciales
embebidas, query y fragmento antes de ejecutar Git. La autenticación, los redirects y los filtros
de checkout siguen bajo responsabilidad de Git; Cerne desactiva los prompts controlables, pero los
helpers externos aún pueden fallar o actuar fuera del control portable de la CLI. El clon no añade
depth, branch, submódulos, LFS, push ni fetch extra. Cualquier opción de source puede combinarse
con `--workflow`; source y clone siguen siendo exclusivos.

Cada clon iniciado crea primero una auditoría saneada en `knowledge/runs/source-clone.json`. Los
fallos previos al clon revierten el intento. Los posteriores conservan knowledge y la auditoría,
eliminan solo el staging privado de Cerne e informan que el workspace quedó incompleto. La promoción
nunca reemplaza un source concurrente; si falla la auditoría final tras la promoción, el source
válido permanece.

## `cerne restore <origen-knowledge> (--source <ruta> | --clone <origen-source>)`

Clona knowledge, lee el nombre del workspace desde `cerne.json` y luego clona source en la ruta
portable del manifiesto o vincula una raíz Git local no bare sin modificarla. El destino debe estar
ausente. Se rechazan sin reemplazo layouts existentes, concurrentes, solapados, con symlink,
parciales, provider desconocido o Git no independiente. Un workflow listo o pendiente se conserva,
pero nunca se ejecuta.

Cada intento válido inicia antes de Git un registro privado y saneado en `~/.cerne/audit`. Se
excluyen orígenes, credenciales, salida Git, argumentos, entorno y rutas absolutas de repositorios.
La autenticación y los redirects siguen siendo comportamiento del Git externo. Los fallos revierten
solo artefactos cuya identidad aún pertenece al intento; repetir el comando es la recuperación
soportada, sin `--resume`. Éxito/ayuda usan stdout y estado `0`, fallos operativos stderr/`1`, y uso
u origen inválido stderr/`2`. Restore no autoriza workflow, agente, push, merge, fetch extra,
submódulos, instalación, publicación ni deploy.

```sh
cerne restore ../knowledge.git --clone ../source.git
cerne restore git@host:org/knowledge.git --source ../source-existente
```

## `cerne skill install <codex|claude|gemini> [cerne-context|cerne-git-workflow]`

Sin argumento de skill, instala todas las skills oficiales compatibles en el perfil del agente del
usuario actual. Codex y Claude reciben `cerne-context` y `cerne-git-workflow`; Gemini recibe solo
`cerne-git-workflow`. Con argumento de skill, instala exactamente esa skill. Los destinos son
`~/.codex/skills/<skill>`, `~/.claude/skills/<skill>` o
`~/.gemini/skills/cerne-git-workflow`.

El comando usa el paquete oficial `cerne-skills` incorporado en el binario, sin red, valida
manifiesto, adaptador, entrypoint y schema `cerne.context.v1` antes de copiar, y registra una
auditoría privada en `~/.cerne/audit` por cada skill instalada.

El uso inválido, incluido `generic`, variantes de mayúsculas, agente ausente o argumentos extra,
devuelve estado `2` sin auditoría ni cambios de archivos. Los fallos operativos devuelven
stderr/`1`. Reinstalar la misma versión es no-op; versiones gestionadas distintas se actualizan.
`init`, `restore` y `workflow setup` nunca instalan skills por implicación.

## `cerne git inspect`

Proporciona la superficie segura de inspección Git usada por `cerne-git-workflow`. Cerne no ejecuta
efectos Git; el agente usa los datos inspeccionados y pide confirmación antes de branch, commit,
push o Pull Request.

```sh
cerne git inspect --agent codex --task task-1 --json
```

`inspect` es de solo lectura y devuelve schema versión 1 con `state_id` determinístico, remotes
sanitizados, branches locales, paths cambiados literales e id privado de auditoría. Los comandos de
branch, commit, push y Pull Request no están disponibles en Cerne; las operaciones Git destructivas
o fuera de alcance siguen fuera de la skill.

El éxito JSON usa stdout/estado `0`; un snapshot de workspace inválido usa stdout/estado `1`; uso
inválido usa stderr/estado `2`. Las auditorías privadas en `~/.cerne/audit` cubren solo `inspect` y
excluyen conversaciones, salida de Git, URLs remotas, tokens, contenido de archivos, cuerpo de PR y
errores brutos. La evidencia de ejecución pertenece al agente o harness, no a la auditoría de Cerne.

## `cerne workflow setup [--agent codex|claude]`

Localiza el workspace ancestro más cercano y materializa el provider declarado en el manifiesto.
No acepta provider, ruta ni opción de fuerza. Con `--agent`, el workflow declarado debe ser Spec Kit
y Cerne prepara o actualiza el puente de descubrimiento en la raíz para el agente local elegido.
Cada subproceso real de provider o integración de agente crea un JSON de auditoría sanitizado en
`knowledge/runs`; no se auditan un ejecutable ausente ni un layout listo sin setup de agente. Para
que Codex descubra el puente en `.agents/skills`, inicia la sesión en la raíz del workspace Cerne, no
dentro de `source/`.

## `cerne context [--json]`

Localiza el workspace ancestro más cercano e informa rutas canónicas de workspace, knowledge,
product, specs, decisions, policies, source y workflow declarado. `--json` emite el schema estable
versión 1 para skills y scripts. Informes saludables o con advertencias devuelven `0`; informes
estructuralmente inválidos siguen siendo válidos y devuelven `1`; uso inválido devuelve `2` en
stderr.

El comando solo lee metadatos estructurales. No lee contenido de repositorios ni archivos de
agente, ejecuta Git o providers, consulta remotos o ejecutables, accede a la red ni crea auditoría,
caché, instrucciones o cambios en el manifiesto.

```sh
cerne context
cerne context --json
```

## `cerne doctor`

Ejecuta diez verificaciones de solo lectura desde la raíz: manifiesto, directorios de ambos
repositorios, independencia Git, aislamiento de versionado, rutas del manifiesto, directorios
obligatorios de conocimiento, Git, permisos y versión del manifiesto. Nunca repara el workspace.
Un workflow declarado añade una verificación para los estados listo, pendiente, no disponible,
desconocido, parcial o con Git anidado.

## `cerne status`

Localiza el workspace ancestro más cercano desde el directorio actual y lee ambos repositorios.
Reconoce árboles limpios o con cambios, detached HEAD y repositorios sin commits. No ejecuta fetch
ni compara con remotos.

## `cerne link <ruta> [--replace]`

Vincula como `source` un repositorio Git local no bare con árbol de trabajo. Acepta rutas relativas,
absolutas y worktrees válidos. Knowledge y source deben ser distintos y no pueden estar anidados de
forma peligrosa. Cambiar un source ya configurado requiere `--replace`; vincular el mismo source
termina sin reescribir el manifiesto. La sustitución del manifiesto es atómica.

<!-- END AUTO-GENERATED -->
