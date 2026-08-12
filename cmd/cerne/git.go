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
  cerne git inspect --agent <codex|claude|gemini> --task <task-id> --json
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
	agent, task, ok := parseGitInspectArgs(args)
	if !ok {
		fmt.Fprint(stderr, messages.text("git.inspect.usage"))
		return 2
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
	snapshot, err := workspace.InspectGit(current, workspace.GitInspectRequest{Agent: agent, TaskID: task, Home: home}, inspect)
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

func parseGitInspectArgs(args []string) (string, string, bool) {
	var agent, task string
	jsonOutput := false
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--agent":
			if agent != "" || i+1 >= len(args) || args[i+1] == "" {
				return "", "", false
			}
			agent = args[i+1]
			i++
		case "--task":
			if task != "" || i+1 >= len(args) || args[i+1] == "" {
				return "", "", false
			}
			task = args[i+1]
			i++
		case "--json":
			if jsonOutput {
				return "", "", false
			}
			jsonOutput = true
		default:
			return "", "", false
		}
	}
	return agent, task, jsonOutput && supportedGitAgent(agent) && validGitTaskID(task)
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
