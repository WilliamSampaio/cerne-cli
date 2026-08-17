package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/WilliamSampaio/cerne-cli/internal/gitexec"
	"github.com/WilliamSampaio/cerne-cli/internal/workspace"
)

const gitHelp = `Coordena operações Git seguras em um workspace Cerne.

Uso:
  cerne git inspect --runtime <codex|claude|gemini> --task <task-id> --json
  cerne git --help

Autorização:
  inspect é somente leitura e fornece estado sanitizado para o agente executar
  Git diretamente. O Cerne não executa branch, commit, push ou Pull Request.
`

func runGit(args []string, stdout, stderr io.Writer, messages localizer, home string) int {
	if len(args) == 1 && args[0] == "--help" {
		fmt.Fprint(stdout, messages.text(messageGitHelp))
		return 0
	}
	if len(args) == 0 {
		fmt.Fprint(stderr, messages.text("git.usage"))
		return 2
	}
	switch args[0] {
	case "inspect":
		return runGitInspect(args[1:], stdout, stderr, home, messages)
	default:
		fmt.Fprint(stderr, messages.text("git.usage"))
		return 2
	}
}

func runGitInspect(args []string, stdout, stderr io.Writer, home string, messages localizer) int {
	parsed, ok := parseGitInspectArgs(args)
	if !ok {
		fmt.Fprint(stderr, messages.text("git.inspect.usage"))
		return 2
	}
	if parsed.RuntimeDeprecated {
		fmt.Fprint(stderr, messages.text("git.inspect.agent-deprecated", parsed.Runtime))
	}
	current, err := currentDirectory()
	if err != nil {
		fmt.Fprint(stderr, messages.text("common.cwd"))
		return 1
	}
	inspect, err := gitexec.FindWorkflowInspector()
	if err != nil {
		fmt.Fprint(stderr, messages.text("common.git"))
		return 1
	}
	snapshot, err := workspace.InspectGit(current, workspace.GitInspectRequest{Runtime: parsed.Runtime, TaskID: parsed.Task, Home: home}, inspect)
	if err != nil {
		var failure workspace.GitFailure
		if errors.As(err, &failure) {
			if failure.Code == "audit_unavailable" {
				fmt.Fprintln(stderr, "audit_unavailable")
				return 1
			}
			fmt.Fprintln(stderr, failure.Code)
			return 1
		}
		fmt.Fprint(stderr, messages.text("git.failure"))
		return 1
	}
	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(snapshot); err != nil {
		return 1
	}
	if snapshot.Status == "invalid" {
		return 1
	}
	return 0
}

type gitInspectArguments struct {
	Runtime           string
	Task              string
	RuntimeDeprecated bool
}

func parseGitInspectArgs(args []string) (gitInspectArguments, bool) {
	var parsed gitInspectArguments
	jsonOutput := false
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--runtime":
			if parsed.Runtime != "" || i+1 >= len(args) || args[i+1] == "" {
				return gitInspectArguments{}, false
			}
			parsed.Runtime = args[i+1]
			i++
		case "--agent":
			if parsed.Runtime != "" || i+1 >= len(args) || args[i+1] == "" {
				return gitInspectArguments{}, false
			}
			parsed.Runtime = args[i+1]
			parsed.RuntimeDeprecated = true
			i++
		case "--task":
			if parsed.Task != "" || i+1 >= len(args) || args[i+1] == "" {
				return gitInspectArguments{}, false
			}
			parsed.Task = args[i+1]
			i++
		case "--json":
			if jsonOutput {
				return gitInspectArguments{}, false
			}
			jsonOutput = true
		default:
			return gitInspectArguments{}, false
		}
	}
	if !jsonOutput || !supportedGitAgent(parsed.Runtime) || !validGitTaskID(parsed.Task) {
		return gitInspectArguments{}, false
	}
	return parsed, true
}

func supportedGitAgent(agent string) bool {
	return agent == "codex" || agent == "claude" || agent == "gemini"
}

func validGitTaskID(task string) bool {
	if task == "" || len(task) > 80 {
		return false
	}
	for _, r := range task {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' || r == '.' {
			continue
		}
		return false
	}
	return true
}
