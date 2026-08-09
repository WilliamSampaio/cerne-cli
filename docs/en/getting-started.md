# Getting started with Cerne

**English** · [Português (Brasil)](../pt-BR/getting-started.md) ·
[Español](../es/getting-started.md)

Cerne manages a software workspace containing two independent Git repositories:

- `knowledge` stores product context, decisions, policies, and execution records. It is normally
  private.
- `source` stores the application source code.

The workspace root connects these repositories but is not itself a Git repository. Cerne does not
commit, push, publish, or deploy your work.

## Before you begin

You need Git in `PATH` and Linux, Windows, or macOS.

On Linux and macOS, install the latest stable standalone binary without Go:

```sh
curl --proto '=https' --tlsv1.2 -fsSL \
  https://github.com/WilliamSampaio/cerne-cli/releases/latest/download/install.sh | sh
~/.local/bin/cerne --version
```

To inspect the installer first:

```sh
curl --proto '=https' --tlsv1.2 -fsSL \
  https://github.com/WilliamSampaio/cerne-cli/releases/latest/download/install.sh
```

Install a fixed version or install the optional Codex/Claude skill explicitly:

```sh
curl --proto '=https' --tlsv1.2 -fsSL \
  https://github.com/WilliamSampaio/cerne-cli/releases/download/vX.Y.Z/install.sh |
  sh -s -- --version vX.Y.Z

curl --proto '=https' --tlsv1.2 -fsSL \
  https://github.com/WilliamSampaio/cerne-cli/releases/latest/download/install.sh |
  sh -s -- --agent codex
```

The installer writes only `~/.local/bin/cerne`, never uses `sudo`, and never edits shell profile
files. If `~/.local/bin` is not in `PATH`, add it yourself. Remove Cerne with:

```sh
rm ~/.local/bin/cerne
```

For manual installation, download the matching archive and `checksums.txt` from the GitHub release,
verify the SHA-256 checksum, extract `cerne`, and place it in a directory on `PATH`.

To install Cerne with Go instead:

```sh
go install github.com/WilliamSampaio/cerne-cli/cmd/cerne@latest
cerne --version
```

Go installs the binary in `GOBIN`, or `GOPATH/bin` when `GOBIN` is unset. Add that directory to
`PATH` if the `cerne` command is not found.

## Create your workspace

Choose the command that matches where your source code is now:

| Starting point | Command | What happens |
| --- | --- | --- |
| New project | `cerne init my-project` | Creates empty `knowledge` and `source` repositories. |
| Existing local repository | `cerne init my-project --source ../my-app` | Links the repository without moving or changing it. |
| Existing remote repository | `cerne init my-project --clone git@host:org/my-app.git` | Clones it into the workspace as `source`. |

Then enter the workspace and check it:

```sh
cd my-project
cerne doctor
cerne status
cerne context
```

`doctor` validates the workspace structure and security boundaries. `status` summarizes local Git
state for both repositories. `context` shows the paths and optional workflow detected by Cerne.
These commands are read-only. CLI messages are currently displayed in Portuguese.

## Understand the layout

```text
my-project/
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

Work with `knowledge` and `source` as separate repositories: each has its own history, branches,
commits, and remotes. Do not store credentials or secrets in either repository.

## Optional workflow

You can prepare a supported specification workflow while creating the workspace:

```sh
cerne init my-project --workflow speckit
cerne init my-project --workflow openspec
```

Cerne uses an existing local `specify` or `openspec` installation; it never installs or updates
these tools. If the executable is unavailable during creation, install it separately and then run:

```sh
cerne workflow setup
```

## Restore an existing workspace

Restoration always needs a knowledge origin and either a source origin or a local source:

```sh
cerne restore git@host:org/knowledge.git --clone git@host:org/source.git
cerne restore ../knowledge.git --source ../existing-source
```

Cerne creates a new destination and refuses to replace an existing one.

## Command details

See the [command reference](commands.md),
[exit codes](../../README.md#exit-codes-and-streams), and
[safety and privacy rules](../../README.md#safety-and-privacy). For recovery guidance, see
[troubleshooting](troubleshooting.md).
