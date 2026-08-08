# Referencia de comandos

[English](../en/commands.md) · [Português (Brasil)](../pt-BR/commands.md) · **Español**

[Primeros pasos](getting-started.md) · [Solución de problemas](troubleshooting.md)

Los mensajes de la CLI se muestran actualmente en portugués. Ejecuta `cerne <comando> --help` para
consultar el contrato completo implementado por tu versión instalada.

<!-- AUTO-GENERATED: mantener sincronizado con cmd/cerne/main.go y los contratos de la CLI. -->

## Opciones globales

- `cerne --help` muestra los comandos disponibles y las opciones globales.
- `cerne --version` muestra la versión instalada como identificador SemVer estable.

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

## `cerne skill install <codex|claude>`

Instala explícitamente la skill oficial `cerne-context` en el perfil del usuario actual:
`~/.codex/skills/cerne-context` para Codex o `~/.claude/skills/cerne-context` para Claude. El
comando usa el paquete oficial `cerne-skills` incorporado en el binario, sin red, valida manifiesto,
adaptador y schema `cerne.context.v1` antes de copiar, y registra una auditoría privada en
`~/.cerne/audit`.

El uso inválido, incluido `generic`, variantes de mayúsculas, agente ausente o argumentos extra,
devuelve estado `2` sin auditoría ni cambios de archivos. Los fallos operativos devuelven
stderr/`1`. Reinstalar la misma versión es no-op; versiones gestionadas distintas se actualizan.
`init`, `restore` y `workflow setup` nunca instalan esta skill por implicación.

## `cerne workflow setup [--agent codex|claude]`

Localiza el workspace ancestro más cercano y materializa el provider declarado en el manifiesto.
No acepta provider, ruta ni opción de fuerza. Con `--agent`, el workflow declarado debe ser Spec Kit
y Cerne prepara o actualiza el puente de descubrimiento en la raíz para el agente local elegido.
Cada subproceso real de provider o integración de agente crea un JSON de auditoría sanitizado en
`knowledge/runs`; no se auditan un ejecutable ausente ni un layout listo sin setup de agente.

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
