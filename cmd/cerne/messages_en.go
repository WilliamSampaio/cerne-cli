package main

var englishMessages = map[messageID]string{
	messageSkillHelp: `Installs official Cerne skills in an agent profile.

Usage:
  cerne skill install <codex|claude|gemini>
  cerne skill install <codex|claude|gemini> <cerne-context|cerne-git-workflow>
  cerne skill install --help
  cerne skill --help

Authorization:
  This command explicitly authorizes changes only to the current user's profile
  for the selected agent. init, restore, and workflow setup never install skills
  implicitly.

Package:
  Uses the official cerne-skills package embedded in the binary, without network
  access. The manifest, requested skill, agent adapter, and cerne.context.v1
  schema are validated before the destination is changed.

Destinations:
  codex:  ~/.codex/skills/cerne-context
  claude: ~/.claude/skills/cerne-context
  codex:  ~/.codex/skills/cerne-git-workflow
  claude: ~/.claude/skills/cerne-git-workflow
  gemini: ~/.gemini/skills/cerne-git-workflow

Output:
  Success and help use stdout. Failures use stderr.
  Status 0: installed, already installed, or upgraded; 1: operational failure;
  2: invalid usage.

Effects:
  Creates a private audit in ~/.cerne/audit. Refuses unknown destinations,
  symlinks, and incompatible packages; reinstalling the same version is a no-op.

Examples:
  cerne skill install codex
  cerne skill install gemini cerne-git-workflow
`,
	messageContextHelp: `Shows the verified structural context of a Cerne workspace.

Usage:
  cerne context
  cerne context --json
  cerne context --help

Location:
  Finds the nearest ancestor workspace. An external source does not perform
  reverse workspace discovery.

Report:
  Reports workspace, knowledge, collections, source, and workflow without reading
  content. --json uses stable schema_version 1 for automation.

Output:
  Reports and help use stdout. Invalid usage uses stderr.
  Status 0: healthy, warnings, or help; 1: invalid context; 2: invalid usage.

Effects:
  Structural read only. Does not run Git, providers, or agents; does not access
  the network, environment, or credentials; creates no audit, cache, or instructions.

Examples:
  cerne context
  cerne context --json
`,
	messageRestoreHelp: `Restores a Cerne workspace from a knowledge repository.

Usage:
  cerne restore <knowledge-origin> --clone <source-origin>
  cerne restore <knowledge-origin> --source <local-path>
  cerne restore --help

Behavior:
  Knowledge is always cloned. --clone also clones source; --source links the root
  of an existing local Git working tree without copying or changing it. The name
  and portable source path are read from knowledge/cerne.json.

Safety:
  The destination must be absent and is never replaced. Restore uses private
  staging, validates independent repositories, and does not run workflow setup.
  Every valid attempt creates a redacted record in ~/.cerne/audit before Git runs.

Output:
  Success and help use stdout. Failures use stderr.
  Status 0: success or help; 1: operational failure; 2: invalid usage or origin.

Effects:
  Does not push, merge, perform an additional fetch, install, deploy, or run a
  provider. Authentication and redirects follow external Git configuration.

Examples:
  cerne restore git@host:org/knowledge.git --clone git@host:org/source.git
  cerne restore ../knowledge.git --source ../existing-source
`,
	messageInitHelp: `Initializes a Cerne workspace with independent Git repositories.

Usage:
  cerne init <project-name>
  cerne init <project-name> --source <path>
  cerne init <project-name> --clone <origin>
  cerne init <project-name> --workflow <speckit|openspec>
  cerne init <project-name> --workflow speckit --agent <codex|claude>
  cerne init <project-name> --source <path> --workflow <speckit|openspec>
  cerne init <project-name> --clone <origin> --workflow <speckit|openspec>

Name:
  1 to 255 ASCII characters; starts with a letter or number and continues with
  letters, numbers, dot, hyphen, or underscore. Reserved names and a trailing dot
  are rejected.

Knowledge:
  Creates README.md with collection purposes, repository boundaries, and the
  first safe commands for inspecting the workspace, using the invocation's
  effective language.

Source:
  Without a flag, creates source as an empty Git repository. --source links the
  root of a local non-bare working tree. --clone accepts local paths, file, HTTPS,
  SSH, and SCP-like origins and creates remote origin. --source and --clone are
  mutually exclusive; --workflow may be used with any mode.

Clone:
  HTTP, git://, ext::, unknown helpers, options, embedded credentials, queries,
  and fragments are refused. Clone is complete, without submodules, and uses the
  default checkout. Git-controlled prompts are disabled.

Workflow:
  Without the option, keeps knowledge/specs. Spec Kit uses specs and .specify.
  OpenSpec uses openspec/specs and openspec. --agent codex|claude is accepted only
  with Spec Kit and creates local discovery without changing the manifest.
  Cerne uses existing local installations and does not install agents or providers.
  Use cerne skill install <agent> to install the global skill.

Effects:
  Always creates knowledge as an independent Git repository. The destination must
  be absent or empty and is never replaced. Clone attempts are audited before Git.
  No mode authorizes push, merge, or publication.

Output:
  Success and help use stdout. Errors use stderr.
  Status 0: success or help; 1: operational failure; 2: invalid usage or name.

Errors:
  A status 1 after clone can preserve a promoted source if final audit fails.
  Inspect the displayed correction and knowledge/runs/source-clone.json before
  removing or linking source.

Examples:
  cerne init example
  cerne init example --source ../application
  cerne init example --clone https://host/organization/application.git
  cerne init example --workflow speckit
  cerne init example --workflow speckit --agent codex
`,
	messageWorkflowHelp: `Initializes the workflow already declared in the workspace manifest.

Usage:
  cerne workflow setup
  cerne workflow setup --agent <codex|claude>
  cerne workflow --help

Location:
  Finds the nearest ancestor workspace through knowledge/cerne.json.

Behavior:
  Runs only the declared provider already installed inside knowledge. A ready
  layout is unchanged and a partial layout is refused. --agent prepares local
  Spec Kit discovery without persisting the agent in the manifest.

Output:
  Success and help use stdout. Failures use stderr.
  Status 0: completed, already ready, or help; 1: operational failure; 2: invalid usage.

Effects:
  Records attempts in knowledge/runs. Does not install agents, change providers,
  alter source, or authorize network, remote Git, or credentials.

Examples:
  cerne workflow setup
  cerne workflow setup --agent claude
`,
	messageDoctorHelp: `Analyzes the current Cerne workspace without modifying files or repositories.

Usage:
  cerne doctor
  cerne doctor --help

Root:
  The current directory is treated as the workspace root.

Checks:
  Manifest; knowledge and source repositories; Git independence; versioning
  isolation; manifest paths; required directories; Git; permissions; manifest
  version; and, when declared, workflow provider, layout, and availability.

Symbols and summaries:
  ✓ passed
  ✗ blocking error
  ! non-blocking warning

Output:
  Reports, summaries, and help use stdout. Invalid usage or pre-report failures
  use stderr. Status 0: healthy, warnings, or help; 1: blocking or operational
  failure; 2: invalid usage.

Effects:
  Read only. Does not create, repair, modify, remove, use network or remote Git,
  access credentials, or invoke AI agents.

Example:
  cerne doctor
`,
	messageStatusHelp: `Shows the current state of a Cerne workspace without modifying files.

Usage:
  cerne status
  cerne status --help

Location:
  Finds the nearest ancestor workspace identified by knowledge/cerne.json.

Report:
  Shows project, absolute workspace path, and each repository's path, branch,
  commit, state, and modified, staged, and untracked counts.

Output:
  Report and help use stdout. Invalid usage and failures use stderr.
  Status 0: report or help; 1: operational failure; 2: invalid usage.

Effects:
  Read only. Does not stage, commit, checkout, reset, fetch, pull, access remotes,
  credentials, or AI agents.

Example:
  cerne status
`,
	messageLinkHelp: `Links an existing local Git repository as the current Cerne workspace source.

Usage:
  cerne link <path>
  cerne link <path> --replace
  cerne link --help

Path:
  May be relative or absolute and must identify the root of a local Git working
  tree. Valid worktrees are accepted; bare repositories are not.

Replacement:
  Replacing an existing source requires --replace. Cerne changes only
  knowledge/cerne.json and does not alter either repository.

Output:
  Success and help use stdout. Invalid usage and failures use stderr.
  Status 0: updated, unchanged, or help; 1: operational failure; 2: invalid usage.

Effects:
  Reads local workspace and Git metadata. Does not copy, move, delete, checkout,
  reset, add, commit, clean, fetch, pull, push, access remotes or credentials.

Example:
  cerne link ../existing-application --replace
`,
	messageGlobalHelp: `Cerne manages workspaces with independent Git repositories for knowledge and source code.

Usage:
  cerne [--lang <en|pt-BR>] <command> [arguments]
  cerne --help
  cerne --version

Commands:
  init      Creates a Cerne workspace
  restore   Restores an existing Cerne workspace
  doctor    Validates workspace structure and safety
  status    Shows the local repository state
  link      Links a local Git repository as source
  workflow  Initializes the workflow declared by the workspace
  context   Shows the workspace structural context
  skill     Installs Cerne skills in an agent profile
  git       Coordinates safe Git inspection
  config    Manages user preferences

Options:
  --lang       Uses en or pt-BR for this invocation only
  --help       Shows this help
  --version    Shows the Cerne version

Language:
  CERNE_LANG temporarily selects the language. Precedence is --lang, CERNE_LANG,
  saved preference, then en. Use "cerne config --help" to persist the choice.

Use "cerne <command> --help" for command details.
`,
	messageConfigHelp: `Manages Cerne user preferences.

Usage:
  cerne config set language <en|pt-BR>
  cerne config get language
  cerne config unset language
  cerne config --help

Storage:
  The preference is saved in ~/.cerne/config.json for the current user. The
  command refuses symlinks and replaces the file atomically.

Precedence:
  --lang, CERNE_LANG, saved preference, then en.
`,
	messageInvalidLanguage:                                    "error: invalid language: %q\ncorrection: use en or pt-BR\n",
	messageInvalidGlobalOption:                                "error: invalid global option\nusage: cerne [--lang <en|pt-BR>] <command> [arguments]\n",
	messageConfigUsage:                                        "error: invalid argument\nusage: cerne config <set language <en|pt-BR>|get language|unset language>\n",
	messageConfigSet:                                          "Saved language: %s\n",
	messageConfigGet:                                          "Saved language: %s\n",
	messageConfigGetUnset:                                     "Language is not set. Current default: en\n",
	messageConfigUnset:                                        "Language preference removed.\n",
	messageConfigUnsafe:                                       "error: unsafe user configuration\ncorrection: remove symlinks from ~/.cerne or ~/.cerne/config.json and try again\n",
	messageConfigRead:                                         "error: could not read the user configuration\ncorrection: check the permissions of ~/.cerne/config.json\n",
	messageConfigInvalid:                                      "error: invalid language configuration\ncorrection: use cerne --lang en config set language <en|pt-BR> to repair it\n",
	messageConfigWrite:                                        "error: could not update the user configuration\ncorrection: check the permissions of ~/.cerne and try again\n",
	messageHomeUnavailable:                                    "error: could not locate the home directory\ncorrection: configure an accessible home directory\n",
	"skill.usage":                                             "error: invalid argument\nusage: cerne skill install <codex|claude|gemini> [cerne-context|cerne-git-workflow]\n",
	"skill.installed":                                         "Installed skill: %s\n",
	"skill.already":                                           "Skill already installed: %s\n",
	"skill.upgraded":                                          "Upgraded skill: %s\n",
	"skill.result":                                            "Agent: %s\nVersion: %s\nDestination: %s\n",
	"skill.failure.default":                                   "error: could not install the skill\ncorrection: check permissions and try again\n",
	"skill.failure.home-unavailable":                          "error: could not locate the home directory\ncorrection: configure an accessible home directory\n",
	"skill.failure.destination-invalid":                       "error: invalid agent destination\ncorrection: configure an accessible home directory\n",
	"skill.failure.audit-start-failed":                        "error: could not record the installation attempt\ncorrection: check the safety and permissions of ~/.cerne/audit\n",
	"skill.failure.install-failed":                            "error: could not install the skill\ncorrection: check permissions and try again\n",
	"skill.failure.audit-finalization-failed":                 "error: could not finalize the installation audit\ncorrection: inspect ~/.cerne/audit before trying again\n",
	"skill.failure.package-unavailable":                       "error: the embedded official cerne-skills package is unavailable\ncorrection: check the temporary directory and reinstall Cerne\n",
	"skill.failure.manifest-invalid":                          "error: invalid cerne-skills package manifest\ncorrection: reinstall Cerne\n",
	"skill.failure.manifest-incompatible":                     "error: incompatible cerne-skills package\ncorrection: update or reinstall Cerne\n",
	"skill.failure.adapter-missing":                           "error: the agent adapter is missing from cerne-skills\ncorrection: update or reinstall Cerne\n",
	"skill.failure.skill-missing":                             "error: the skill is missing from cerne-skills\ncorrection: use a supported official skill or update Cerne\n",
	"skill.failure.unsafe-package":                            "error: cerne-skills contains unsafe content\ncorrection: reinstall Cerne\n",
	"skill.failure.destination-inaccessible":                  "error: the agent destination is inaccessible\ncorrection: check permissions in the agent profile\n",
	"skill.failure.unknown-destination":                       "error: the existing destination is not managed by Cerne\ncorrection: move the existing content before installing or upgrading\n",
	"skill.failure.promotion-failed":                          "error: could not promote the installation\ncorrection: check permissions in the agent profile\n",
	"context.usage":                                           "error: invalid argument\nusage: cerne context [--json]\n",
	"context.workspace":                                       "Workspace: %s\n",
	"context.status":                                          "Status: %s\n",
	"context.root":                                            "Root: %s\n",
	"context.knowledge":                                       "\nKnowledge: %s\n",
	"context.source":                                          "\nSource: %s\n",
	"context.location":                                        "Location: %s\n",
	"context.workflow.not-declared":                           "Workflow: not declared\n",
	"context.workflow":                                        "Workflow: %s (%s)\n",
	"context.problem":                                         "%s %s: %s\n",
	"context.correction":                                      "  Correction: %s\n",
	"context.label.product":                                   "Product",
	"context.label.specs":                                     "Specs",
	"context.label.decisions":                                 "Decisions",
	"context.label.policies":                                  "Policies",
	"context.location.inside":                                 "inside the workspace",
	"context.location.outside":                                "outside the workspace",
	"context.status.healthy":                                  "healthy",
	"context.status.warning":                                  "warning",
	"context.status.invalid":                                  "invalid",
	"context.workflow.pending":                                "pending",
	"context.workflow.ready":                                  "ready",
	"context.workflow.invalid":                                "invalid",
	"context.workflow.unknown-provider":                       "unknown provider",
	"context.component.workspace":                             "Workspace",
	"context.component.manifest":                              "Manifest",
	"context.component.knowledge":                             "Knowledge",
	"context.component.source":                                "Source",
	"context.component.workflow":                              "Workflow",
	"context.problem.workspace-not-found.detail":              "not found",
	"context.problem.workspace-not-found.correction":          "run the command inside a Cerne workspace",
	"context.problem.manifest-invalid.detail":                 "missing or invalid",
	"context.problem.manifest-invalid.correction":             "repair or restore knowledge/cerne.json",
	"context.problem.manifest-version-unsupported.detail":     "unsupported version",
	"context.problem.manifest-version-unsupported.correction": "use a compatible Cerne version",
	"context.problem.knowledge-invalid.detail":                "missing or unsafe",
	"context.problem.knowledge-invalid.correction":            "restore the knowledge directory",
	"context.problem.source-invalid.detail":                   "missing or unsafe",
	"context.problem.source-invalid.correction":               "correct the source path in the manifest",
	"context.problem.required-directory-invalid.detail":       "required directory missing or invalid",
	"context.problem.required-directory-invalid.correction":   "restore the directory identified by the component",
	"context.problem.workflow-pending.detail":                 "pending",
	"context.problem.workflow-pending.correction":             "run cerne workflow setup when the provider is available",
	"context.problem.workflow-invalid.detail":                 "invalid or partial layout",
	"context.problem.workflow-invalid.correction":             "repair the layout before continuing",
	"context.problem.workflow-unknown-provider.detail":        "unsupported provider",
	"context.problem.workflow-unknown-provider.correction":    "use speckit or openspec in the manifest",
	messageGitHelp:                                            gitHelp,
	"git.usage":                                               "error: invalid argument\nusage: cerne git inspect --agent <codex|claude|gemini> --task <task-id> --json\n",
	"git.inspect.usage":                                       "error: invalid argument\nusage: cerne git inspect --agent <codex|claude|gemini> --task <task-id> --json\n",
	"git.failure":                                             "error: could not inspect workspace Git\ncorrection: check the workspace and try again\n",
	"command.missing":                                         "error: provide a command\nusage: cerne <init|restore|doctor|status|link|workflow|context|skill|git|config>\n",
	"command.unknown":                                         "error: unknown command\nusage: cerne <init|restore|doctor|status|link|workflow|context|skill|git|config>\n",
	"common.cwd":                                              "error: could not get the current directory\ncorrection: run the command in an accessible directory\n",
	"common.git":                                              "error: Git is unavailable\ncorrection: install Git and make it available in PATH\n",
	"common.home":                                             "error: could not locate the home directory\ncorrection: configure an accessible home directory\n",
	"restore.usage":                                           "usage: cerne restore <knowledge-origin> (--source <path> | --clone <source-origin>)\n",
	"restore.invalid-argument":                                "error: invalid argument\n",
	"restore.invalid-knowledge-origin":                        "error: invalid knowledge origin\n",
	"restore.invalid-source-origin":                           "error: invalid source clone origin\n",
	"restore.failure.default":                                 "error: could not restore the workspace\ncorrection: inspect the audit and try again\n",
	"restore.result":                                          "Restored workspace %q.\nKnowledge: %s\n",
	"restore.source.cloned":                                   "Cloned source: %s\n",
	"restore.source.linked":                                   "Linked source: %s\n",
	"restore.manifest.changed":                                "Manifest: source reference updated.\n",
	"init.usage":                                              "usage: cerne init <project-name> [--source <path> | --clone <origin>] [--workflow <speckit|openspec> [--agent <codex|claude>]]\n",
	"init.invalid-argument":                                   "error: invalid argument\n",
	"init.invalid-name":                                       "error: invalid project name; use 1 to 255 ASCII characters, start with a letter or number, and avoid reserved names\n",
	"init.invalid-workflow":                                   "error: invalid workflow: use speckit or openspec\n",
	"init.invalid-clone-origin":                               "error: invalid clone origin\n",
	"init.destination-unsafe":                                 "error: unsafe destination\ncorrection: choose an absent or empty destination\n",
	"init.failure.default":                                    "error: could not create the workspace\ncorrection: check permissions and try again\n",
	"init.workflow.failure":                                   "error: could not initialize workflow %s: %s\n",
	"init.workflow.correction":                                "correction: repair or update %s and run %q inside %s\n",
	"init.result":                                             "Created workspace %q.\nKnowledge: %s\nSource: %s\n",
	"init.result.knowledge":                                   "Created workspace %q.\nKnowledge: %s\n",
	"init.source.linked":                                      "Linked source: %s\n",
	"init.source.cloned":                                      "Cloned source: %s\n",
	"init.workflow.result":                                    "Workflow: %s\nSetup: %s\n",
	"init.result.workflow":                                    "Created workspace %q.\nKnowledge: %s\nSource: %s\nWorkflow: %s\nSetup: %s\n",
	"workflow.state.configured":                               "completed",
	"workflow.state.pending":                                  "pending",
	"agent.discovery":                                         "Agent: %s\nDiscovery: ready\n",
	"workflow.pending.warning":                                "warning: executable %q was not found; workflow %s was not initialized\n",
	"workflow.pending.correction":                             "correction: install %s and run %q inside the workspace\n",
	"workflow.usage":                                          "error: invalid argument\nusage: cerne workflow setup [--agent <codex|claude>]\n",
	"workflow.failure.default":                                "error: could not initialize the workflow\ncorrection: inspect the workspace and try again\n",
	"workflow.executor.missing":                               "error: executable %q was not found\ncorrection: install %s and run the command again\n",
	"workflow.result":                                         "Workflow: %s\nKnowledge: %s\n",
	"workflow.unchanged":                                      "No changes required.\n",
	"workflow.completed":                                      "Setup completed.\n",
	"doctor.usage":                                            "error: invalid argument\nusage: cerne doctor\n",
	"diagnosis.line":                                          "%s %s: %s",
	"diagnosis.correction":                                    "; correction: %s",
	"diagnosis.invalid":                                       "Invalid workspace\n",
	"diagnosis.warning":                                       "Workspace has warnings\n",
	"diagnosis.healthy":                                       "Healthy workspace\n",
	"status.usage":                                            "error: invalid argument\nusage: cerne status\n",
	"status.failure.default":                                  "error: could not inspect the workspace\ncorrection: inspect the workspace and try again\n",
	"status.project":                                          "Project: %s\n",
	"status.workspace":                                        "Workspace: %s\n\n",
	"status.path":                                             "  Path: %s\n",
	"status.branch":                                           "  Branch: %s\n",
	"status.commit":                                           "  Commit: %s\n",
	"status.state":                                            "  State: %s\n",
	"status.modified":                                         "  Modified: %d\n",
	"status.staged":                                           "  Staged: %d\n",
	"status.untracked":                                        "  Untracked: %d\n",
	"status.repository.knowledge":                             "Knowledge",
	"status.repository.source":                                "Source",
	"status.state.clean":                                      "clean",
	"status.state.pending":                                    "pending changes",
	"status.branch.detached-head":                             "detached HEAD",
	"status.commit.no-commits":                                "no commits",
	"link.usage":                                              "error: invalid argument\nusage: cerne link <path> [--replace]\n",
	"link.failure.default":                                    "error: could not link source\ncorrection: inspect the workspace and try again\n",
	"link.project":                                            "Project: %s\n",
	"link.current":                                            "Current source: %s\n",
	"link.unchanged":                                          "No changes required.\n",
	"link.previous":                                           "Previous source: %s\n",
	"link.new":                                                "New source: %s\n",
	"link.updated":                                            "Manifest updated.\n",
	"failure.cause":                                           "error: %s\n",
	"failure.cause.path":                                      "error: %s: %s\n",
	"failure.correction":                                      "correction: %s\n",
	"failure.operational":                                     "operational failure",
	"failure.check-and-retry":                                 "inspect the workspace and try again",
}
