# Primeros pasos con Cerne

[English](../en/getting-started.md) · [Português (Brasil)](../pt-BR/getting-started.md) ·
**Español**

Cerne administra un espacio de trabajo de software con dos repositorios Git independientes:

- `knowledge` guarda contexto del producto, decisiones, políticas y registros de ejecución.
  Normalmente es privado.
- `source` guarda el código fuente de la aplicación.

La raíz del espacio de trabajo conecta estos repositorios, pero no es un repositorio Git. Cerne no
hace commit, push, publicación ni despliegue de tu trabajo.

## Antes de comenzar

Necesitas Git en `PATH` y Linux, Windows o macOS. Para instalar Cerne con Go:

```sh
go install github.com/WilliamSampaio/cerne-cli/cmd/cerne@latest
cerne --version
```

Go instala el binario en `GOBIN`, o en `GOPATH/bin` cuando `GOBIN` no está definido. Añade ese
directorio a `PATH` si no se encuentra el comando `cerne`.

## Crea tu espacio de trabajo

Elige el comando que represente dónde está ahora tu código fuente:

| Punto de partida | Comando | Qué sucede |
| --- | --- | --- |
| Proyecto nuevo | `cerne init mi-proyecto` | Crea repositorios `knowledge` y `source` vacíos. |
| Repositorio local existente | `cerne init mi-proyecto --source ../mi-app` | Vincula el repositorio sin moverlo ni modificarlo. |
| Repositorio remoto existente | `cerne init mi-proyecto --clone git@host:org/mi-app.git` | Clona el repositorio en el espacio de trabajo como `source`. |

Entra en el espacio de trabajo y comprueba su estructura:

```sh
cd mi-proyecto
cerne doctor
cerne status
cerne context
```

`doctor` valida la estructura y los límites de seguridad. `status` resume el estado local de Git en
ambos repositorios. `context` muestra las rutas y el workflow opcional detectado por Cerne. Estos
comandos son de solo lectura. Los mensajes de la CLI se muestran actualmente en portugués.

## Entiende la estructura

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

Trabaja con `knowledge` y `source` como repositorios separados: cada uno tiene su propio historial,
ramas, commits y remotos. No guardes credenciales ni secretos en ninguno de ellos.

## Workflow opcional

Puedes preparar un workflow de especificación compatible al crear el espacio de trabajo:

```sh
cerne init mi-proyecto --workflow speckit
cerne init mi-proyecto --workflow openspec
```

Cerne usa una instalación local existente de `specify` u `openspec`; nunca instala ni actualiza
estas herramientas. Si el ejecutable no está disponible durante la creación, instálalo por separado
y después ejecuta:

```sh
cerne workflow setup
```

## Restaura un espacio de trabajo existente

La restauración siempre necesita un origen para `knowledge` y un origen o ruta local para `source`:

```sh
cerne restore git@host:org/knowledge.git --clone git@host:org/source.git
cerne restore ../knowledge.git --source ../source-existente
```

Cerne crea un destino nuevo y se niega a reemplazar uno existente.

## Detalles de los comandos

Consulta la [referencia de comandos](commands.md), los
[códigos de salida](../../README.es.md#códigos-de-salida-y-streams) y las reglas de
[seguridad y privacidad](../../README.es.md#seguridad-y-privacidad).
