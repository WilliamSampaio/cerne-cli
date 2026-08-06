package workflowexec

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/WilliamSampaio/cerne-cli/internal/workspace"
)

var ErrUnknownProvider = errors.New("workflow desconhecido")

var lookPath = exec.LookPath

var runProvider = func(executable string, arguments []string, directory string, environment []string) error {
	command := exec.Command(executable, arguments...)
	command.Dir = directory
	command.Env = environment
	return command.Run()
}

func Resolve(provider string) (workspace.WorkflowDefinition, error) {
	definition, err := Describe(provider)
	if err != nil {
		return workspace.WorkflowDefinition{}, err
	}
	var arguments func(string) []string
	switch provider {
	case "speckit":
		arguments = func(string) []string {
			script := "sh"
			if runtime.GOOS == "windows" {
				script = "ps"
			}
			return []string{"init", "--here", "--force", "--integration", "generic", "--integration-options=--commands-dir .specify/commands", "--ignore-agent-tools", "--script", script}
		}
	case "openspec":
		arguments = func(knowledge string) []string {
			return []string{"init", knowledge, "--tools", "none", "--profile", "core", "--no-animation"}
		}
	}

	executable, err := lookPath(definition.Executor)
	if err != nil {
		return definition, nil
	}
	executable, err = filepath.Abs(executable)
	if err != nil {
		return definition, nil
	}
	definition.Available = true
	definition.Setup = func(knowledge string) error {
		if err := runProvider(executable, arguments(knowledge), knowledge, workflowEnvironment(os.Environ(), provider)); err != nil {
			return errors.New("provider falhou")
		}
		return nil
	}
	return definition, nil
}

func Describe(provider string) (workspace.WorkflowDefinition, error) {
	switch provider {
	case "speckit":
		return workspace.WorkflowDefinition{Provider: provider, Executor: "specify", CanonicalSpecs: "specs", OwnedRoot: ".specify", Marker: ".specify/init-options.json"}, nil
	case "openspec":
		return workspace.WorkflowDefinition{Provider: provider, Executor: "openspec", CanonicalSpecs: "openspec/specs", OwnedRoot: "openspec", Marker: "openspec/config.yaml"}, nil
	default:
		return workspace.WorkflowDefinition{}, ErrUnknownProvider
	}
}

func workflowEnvironment(environment []string, provider string) []string {
	allowed := map[string]bool{
		"PATH": true, "TMPDIR": true, "TMP": true, "TEMP": true,
		"LANG": true, "LC_ALL": true, "LC_CTYPE": true,
	}
	if runtime.GOOS == "windows" {
		for _, name := range []string{"SystemRoot", "WINDIR", "ComSpec", "PATHEXT", "USERPROFILE", "APPDATA", "LOCALAPPDATA"} {
			allowed[strings.ToUpper(name)] = true
		}
	} else {
		allowed["HOME"] = true
	}
	clean := make([]string, 0, len(environment)+3)
	for _, entry := range environment {
		name, _, ok := strings.Cut(entry, "=")
		if ok && allowed[strings.ToUpper(name)] {
			clean = append(clean, entry)
		}
	}
	if provider == "openspec" {
		clean = append(clean, "OPENSPEC_TELEMETRY=0", "DO_NOT_TRACK=1", "NO_COLOR=1")
	}
	return clean
}
