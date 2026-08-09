package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/WilliamSampaio/cerne-cli/internal/gitexec"
	"github.com/WilliamSampaio/cerne-cli/internal/githubapi"
	"github.com/WilliamSampaio/cerne-cli/internal/workspace"
)

const gitHelp = `Coordena operações Git seguras em um workspace Cerne.

Uso:
  cerne git inspect --agent <codex|claude|gemini> --task <task-id> --json
  cerne git branch create --name <branch> --base knowledge=<base> --base source=<base> --state <state-id> --confirm --agent <codex|claude|gemini> --task <task-id> --json
  cerne git commit <repository> --message <subject> --include <path> --state <state-id> --confirm --agent <codex|claude|gemini> --task <task-id> --json
  cerne git push <repository> --remote <name> --branch <branch> --state <state-id> --confirm --agent <codex|claude|gemini> --task <task-id> --json
  cerne git pr create <repository> --remote <name> --base <branch> --head <branch> --title <title> --body-file <path> --state <state-id> --confirm --agent <codex|claude|gemini> --task <task-id> --json
  cerne git --help

Autorização:
  inspect é somente leitura, mas exige agente e tarefa para auditoria. Mutações
  exigem estado obtido por inspect e confirmação explícita.
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
	case "branch":
		return runGitBranch(args[1:], stdout, stderr, home, messages)
	case "commit":
		return runGitCommit(args[1:], stdout, stderr, home, messages)
	case "push":
		return runGitPush(args[1:], stdout, stderr, home, messages)
	case "pr":
		return runGitPR(args[1:], stdout, stderr, home, messages)
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

func runGitBranch(args []string, stdout, stderr io.Writer, home string, messages localizer) int {
	request, ok := parseGitBranchArgs(args)
	if !ok {
		fmt.Fprint(stderr, messages.text("git.branch.usage"))
		return 2
	}
	request.Home = home
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
	branch, err := gitexec.FindWorkflowBrancher()
	if err != nil {
		fmt.Fprint(stderr, messages.text("common.git"))
		return 1
	}
	report, err := workspace.CreateGitBranch(current, request, inspect, branch)
	if err != nil {
		var failure workspace.GitFailure
		if errors.As(err, &failure) {
			fmt.Fprintln(stderr, failure.Code)
			if failure.Code == "validation_failed" {
				return 2
			}
			return 1
		}
		fmt.Fprint(stderr, messages.text("git.failure"))
		return 1
	}
	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		return 1
	}
	if report.Status != "succeeded" {
		return 1
	}
	return 0
}

func runGitCommit(args []string, stdout, stderr io.Writer, home string, messages localizer) int {
	request, ok := parseGitCommitArgs(args)
	if !ok {
		fmt.Fprint(stderr, messages.text("git.commit.usage"))
		return 2
	}
	request.Home = home
	inspect, committer, current, ok := gitCLIAdapters(stderr, messages)
	if !ok {
		return 1
	}
	report, err := workspace.CommitGit(current, request, inspect, committer)
	return renderGitMutation(report, err, stdout, stderr, messages)
}

func runGitPush(args []string, stdout, stderr io.Writer, home string, messages localizer) int {
	request, ok := parseGitPushArgs(args)
	if !ok {
		fmt.Fprint(stderr, messages.text("git.push.usage"))
		return 2
	}
	request.Home = home
	inspect, _, current, ok := gitCLIAdapters(stderr, messages)
	if !ok {
		return 1
	}
	pusher, err := gitexec.FindWorkflowPusher()
	if err != nil {
		fmt.Fprint(stderr, messages.text("common.git"))
		return 1
	}
	report, err := workspace.PushGit(current, request, inspect, pusher)
	return renderGitMutation(report, err, stdout, stderr, messages)
}

func runGitPR(args []string, stdout, stderr io.Writer, home string, messages localizer) int {
	request, bodyFile, ok := parseGitPRArgs(args)
	if !ok {
		fmt.Fprint(stderr, messages.text("git.pr.usage"))
		return 2
	}
	body, err := os.ReadFile(bodyFile)
	if err != nil {
		fmt.Fprintln(stderr, "validation_failed")
		return 2
	}
	request.Body = string(body)
	request.Home = home
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
	remoteURL, err := gitexec.FindWorkflowRemoteURLReader()
	if err != nil {
		fmt.Fprint(stderr, messages.text("common.git"))
		return 1
	}
	open := func(ctx context.Context, request workspace.GitPullRequestRequest, remote string) (workspace.GitPullRequestResult, error) {
		result, err := githubapi.OpenPullRequest(ctx, githubapi.PullRequestRequest{
			RemoteURL:  remote,
			Base:       request.Base,
			Head:       request.Head,
			Title:      request.Title,
			Body:       request.Body,
			Env:        os.Environ(),
			APIBaseURL: os.Getenv("CERNE_GITHUB_API_BASE"),
			UserAgent:  "cerne/dev",
		})
		return workspace.GitPullRequestResult{Number: result.Number, URL: result.URL, Outcome: result.Outcome}, err
	}
	report, err := workspace.CreateGitPullRequest(current, request, inspect, remoteURL, open)
	return renderGitMutation(report, err, stdout, stderr, messages)
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

func gitCLIAdapters(stderr io.Writer, messages localizer) (gitexec.WorkflowInspector, gitexec.WorkflowCommitter, string, bool) {
	current, err := currentDirectory()
	if err != nil {
		fmt.Fprint(stderr, messages.text("common.cwd"))
		return nil, nil, "", false
	}
	inspect, err := gitexec.FindWorkflowInspector()
	if err != nil {
		fmt.Fprint(stderr, messages.text("common.git"))
		return nil, nil, "", false
	}
	committer, err := gitexec.FindWorkflowCommitter()
	if err != nil {
		fmt.Fprint(stderr, messages.text("common.git"))
		return nil, nil, "", false
	}
	return inspect, committer, current, true
}

func renderGitMutation(report workspace.WorkspaceGitMutationReport, err error, stdout, stderr io.Writer, messages localizer) int {
	if err != nil {
		var failure workspace.GitFailure
		if errors.As(err, &failure) {
			fmt.Fprintln(stderr, failure.Code)
			if failure.Code == "validation_failed" {
				return 2
			}
			return 1
		}
		fmt.Fprint(stderr, messages.text("git.failure"))
		return 1
	}
	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		return 1
	}
	if report.Status != "succeeded" {
		return 1
	}
	return 0
}

func parseGitBranchArgs(args []string) (workspace.GitBranchRequest, bool) {
	if len(args) == 0 || args[0] != "create" {
		return workspace.GitBranchRequest{}, false
	}
	var request workspace.GitBranchRequest
	jsonOutput := false
	request.Bases = map[string]string{}
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--agent":
			if request.Agent != "" || i+1 >= len(args) || args[i+1] == "" {
				return workspace.GitBranchRequest{}, false
			}
			request.Agent = args[i+1]
			i++
		case "--task":
			if request.TaskID != "" || i+1 >= len(args) || args[i+1] == "" {
				return workspace.GitBranchRequest{}, false
			}
			request.TaskID = args[i+1]
			i++
		case "--state":
			if request.StateID != "" || i+1 >= len(args) || args[i+1] == "" {
				return workspace.GitBranchRequest{}, false
			}
			request.StateID = args[i+1]
			i++
		case "--name":
			if request.Name != "" || i+1 >= len(args) || args[i+1] == "" {
				return workspace.GitBranchRequest{}, false
			}
			request.Name = args[i+1]
			i++
		case "--base":
			if i+1 >= len(args) {
				return workspace.GitBranchRequest{}, false
			}
			repo, base, ok := strings.Cut(args[i+1], "=")
			if !ok || repo == "" || base == "" || request.Bases[repo] != "" {
				return workspace.GitBranchRequest{}, false
			}
			request.Bases[repo] = base
			i++
		case "--confirm":
			if request.Confirm {
				return workspace.GitBranchRequest{}, false
			}
			request.Confirm = true
		case "--json":
			if jsonOutput {
				return workspace.GitBranchRequest{}, false
			}
			jsonOutput = true
		default:
			return workspace.GitBranchRequest{}, false
		}
	}
	if !jsonOutput || !request.Confirm || !supportedGitAgent(request.Agent) || !validGitTaskID(request.TaskID) ||
		request.Bases["knowledge"] == "" || request.Bases["source"] == "" || len(request.Bases) != 2 {
		return workspace.GitBranchRequest{}, false
	}
	return request, true
}

func parseGitCommitArgs(args []string) (workspace.GitCommitRequest, bool) {
	if len(args) < 1 || strings.HasPrefix(args[0], "-") {
		return workspace.GitCommitRequest{}, false
	}
	request := workspace.GitCommitRequest{Repository: args[0]}
	jsonOutput := false
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--agent":
			if request.Agent != "" || i+1 >= len(args) || args[i+1] == "" {
				return workspace.GitCommitRequest{}, false
			}
			request.Agent = args[i+1]
			i++
		case "--task":
			if request.TaskID != "" || i+1 >= len(args) || args[i+1] == "" {
				return workspace.GitCommitRequest{}, false
			}
			request.TaskID = args[i+1]
			i++
		case "--state":
			if request.StateID != "" || i+1 >= len(args) || args[i+1] == "" {
				return workspace.GitCommitRequest{}, false
			}
			request.StateID = args[i+1]
			i++
		case "--message":
			if request.Message != "" || i+1 >= len(args) || args[i+1] == "" {
				return workspace.GitCommitRequest{}, false
			}
			request.Message = args[i+1]
			i++
		case "--include":
			if i+1 >= len(args) || args[i+1] == "" {
				return workspace.GitCommitRequest{}, false
			}
			request.Paths = append(request.Paths, args[i+1])
			i++
		case "--confirm":
			if request.Confirm {
				return workspace.GitCommitRequest{}, false
			}
			request.Confirm = true
		case "--json":
			if jsonOutput {
				return workspace.GitCommitRequest{}, false
			}
			jsonOutput = true
		default:
			return workspace.GitCommitRequest{}, false
		}
	}
	return request, jsonOutput && request.Confirm && supportedGitAgent(request.Agent) && validGitTaskID(request.TaskID) &&
		request.StateID != "" && request.Message != "" && len(request.Paths) > 0
}

func parseGitPushArgs(args []string) (workspace.GitPushRequest, bool) {
	if len(args) < 1 || strings.HasPrefix(args[0], "-") {
		return workspace.GitPushRequest{}, false
	}
	request := workspace.GitPushRequest{Repository: args[0]}
	jsonOutput := false
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--agent":
			if request.Agent != "" || i+1 >= len(args) || args[i+1] == "" {
				return workspace.GitPushRequest{}, false
			}
			request.Agent = args[i+1]
			i++
		case "--task":
			if request.TaskID != "" || i+1 >= len(args) || args[i+1] == "" {
				return workspace.GitPushRequest{}, false
			}
			request.TaskID = args[i+1]
			i++
		case "--state":
			if request.StateID != "" || i+1 >= len(args) || args[i+1] == "" {
				return workspace.GitPushRequest{}, false
			}
			request.StateID = args[i+1]
			i++
		case "--remote":
			if request.Remote != "" || i+1 >= len(args) || args[i+1] == "" {
				return workspace.GitPushRequest{}, false
			}
			request.Remote = args[i+1]
			i++
		case "--branch":
			if request.Branch != "" || i+1 >= len(args) || args[i+1] == "" {
				return workspace.GitPushRequest{}, false
			}
			request.Branch = args[i+1]
			i++
		case "--confirm":
			if request.Confirm {
				return workspace.GitPushRequest{}, false
			}
			request.Confirm = true
		case "--json":
			if jsonOutput {
				return workspace.GitPushRequest{}, false
			}
			jsonOutput = true
		default:
			return workspace.GitPushRequest{}, false
		}
	}
	return request, jsonOutput && request.Confirm && supportedGitAgent(request.Agent) && validGitTaskID(request.TaskID) &&
		request.StateID != "" && request.Remote != "" && request.Branch != ""
}

func parseGitPRArgs(args []string) (workspace.GitPullRequestRequest, string, bool) {
	if len(args) < 2 || args[0] != "create" || strings.HasPrefix(args[1], "-") {
		return workspace.GitPullRequestRequest{}, "", false
	}
	request := workspace.GitPullRequestRequest{Repository: args[1]}
	bodyFile := ""
	jsonOutput := false
	for i := 2; i < len(args); i++ {
		switch args[i] {
		case "--agent":
			if request.Agent != "" || i+1 >= len(args) || args[i+1] == "" {
				return workspace.GitPullRequestRequest{}, "", false
			}
			request.Agent = args[i+1]
			i++
		case "--task":
			if request.TaskID != "" || i+1 >= len(args) || args[i+1] == "" {
				return workspace.GitPullRequestRequest{}, "", false
			}
			request.TaskID = args[i+1]
			i++
		case "--state":
			if request.StateID != "" || i+1 >= len(args) || args[i+1] == "" {
				return workspace.GitPullRequestRequest{}, "", false
			}
			request.StateID = args[i+1]
			i++
		case "--remote":
			if request.Remote != "" || i+1 >= len(args) || args[i+1] == "" {
				return workspace.GitPullRequestRequest{}, "", false
			}
			request.Remote = args[i+1]
			i++
		case "--base":
			if request.Base != "" || i+1 >= len(args) || args[i+1] == "" {
				return workspace.GitPullRequestRequest{}, "", false
			}
			request.Base = args[i+1]
			i++
		case "--head":
			if request.Head != "" || i+1 >= len(args) || args[i+1] == "" {
				return workspace.GitPullRequestRequest{}, "", false
			}
			request.Head = args[i+1]
			i++
		case "--title":
			if request.Title != "" || i+1 >= len(args) || args[i+1] == "" {
				return workspace.GitPullRequestRequest{}, "", false
			}
			request.Title = args[i+1]
			i++
		case "--body-file":
			if bodyFile != "" || i+1 >= len(args) || args[i+1] == "" {
				return workspace.GitPullRequestRequest{}, "", false
			}
			bodyFile = args[i+1]
			i++
		case "--confirm":
			if request.Confirm {
				return workspace.GitPullRequestRequest{}, "", false
			}
			request.Confirm = true
		case "--json":
			if jsonOutput {
				return workspace.GitPullRequestRequest{}, "", false
			}
			jsonOutput = true
		default:
			return workspace.GitPullRequestRequest{}, "", false
		}
	}
	return request, bodyFile, jsonOutput && request.Confirm && supportedGitAgent(request.Agent) && validGitTaskID(request.TaskID) &&
		request.StateID != "" && request.Remote != "" && request.Base != "" && request.Head != "" && request.Title != "" && bodyFile != ""
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
