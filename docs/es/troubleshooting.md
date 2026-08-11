# Solución de problemas

[English](../en/troubleshooting.md) · [Português (Brasil)](../pt-BR/troubleshooting.md) ·
**Español**

[Primeros pasos](getting-started.md) · [Comandos](commands.md)

Cerne envía los errores y la `correção:` sugerida a stderr. Sigue primero esa corrección. Los
mensajes de la CLI se muestran actualmente en portugués, por lo que se citan así a continuación.

## No se encuentra el comando `cerne` o Git

Si el shell no encuentra `cerne`, añade el directorio `GOBIN` o `GOPATH/bin` de Go a `PATH`.
Confirma la instalación con:

```sh
cerne --version
```

Para `erro: Git indisponível`, instala Git y verifica que `git --version` funcione en el mismo shell.

## No se encuentra el workspace

`erro: workspace Cerne não localizado` significa que el directorio actual no está dentro de un
workspace que contenga `knowledge/cerne.json`. Entra en el workspace antes de ejecutar `status`,
`context`, `link` o `workflow setup`.

`cerne doctor` es diferente: ejecútalo desde la raíz del workspace.

```sh
cd mi-proyecto
cerne doctor
cerne context
```

Si `knowledge/cerne.json` se eliminó o dañó, restáuralo desde el repositorio knowledge antes de
continuar.

## `init` rechaza el destino

Cerne nunca reemplaza contenido existente. Usa un nombre de proyecto cuyo destino no exista o sea
un directorio normal vacío. Inspecciona el directorio antes de eliminar cualquier cosa.

Una aplicación local existente debe vincularse en lugar de usarse como destino:

```sh
cerne init mi-proyecto --source ../aplicacion-existente
```

## Se rechaza un source local

La ruta usada en `--source` o `link` debe ser la raíz de un repositorio Git existente, no bare y con
working tree. Debe ser independiente de `knowledge` y no puede solaparse con rutas protegidas del
workspace.

```sh
git -C ../aplicacion-existente status
cerne link ../aplicacion-existente --replace
```

Usa `--replace` solo cuando ya esté configurado otro source. Reemplaza la referencia del manifiesto,
no los repositorios.

## El setup del workflow está pendiente o falla

Si `init` informa que falta el ejecutable `specify` u `openspec`, el workspace sí fue creado.
Instala por separado el provider elegido y ejecuta dentro del workspace:

```sh
cerne workflow setup
```

Para `estrutura do workflow inválida ou parcial`, no repitas el setup indefinidamente. Inspecciona
y corrige el directorio parcial perteneciente al provider y ejecuta `cerne doctor` antes de volver a
intentarlo.

## Un clon de source deja el workspace incompleto

Un fallo después de iniciar el clon puede conservar `knowledge` y su auditoría saneada, dejando el
workspace incompleto. Inspecciona primero el registro:

```text
knowledge/runs/source-clone.json
```

Después, asocia un source local válido o elimina manualmente el workspace incompleto tras confirmar
que no contiene nada necesario. `init` no tiene modo de reanudación.

## Falla `restore`

El destino derivado del manifiesto restaurado no puede existir. La autenticación y el acceso remoto
son responsabilidad de Git; comprueba que los orígenes funcionen con tu configuración normal de Git.

Restore no tiene modo de reanudación. Lee la corrección mostrada y el registro privado en
`~/.cerne/audit`, corrige la causa e inténtalo de nuevo. Cerne no reemplaza un destino existente
durante la recuperación.

## El paquete de la skill del agente no está disponible

`erro: pacote oficial cerne-skills incorporado está inacessível` significa que el paquete
incorporado no pudo materializarse o validarse. Comprueba el acceso al directorio temporal del
sistema y reinstala Cerne antes de repetir `cerne skill install <codex|claude|gemini>`.
