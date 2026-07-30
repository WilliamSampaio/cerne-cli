package gitexec

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var redirectedVariables = map[string]bool{
	"GIT_DIR":                          true,
	"GIT_WORK_TREE":                    true,
	"GIT_COMMON_DIR":                   true,
	"GIT_INDEX_FILE":                   true,
	"GIT_OBJECT_DIRECTORY":             true,
	"GIT_ALTERNATE_OBJECT_DIRECTORIES": true,
	"GIT_NAMESPACE":                    true,
}

func Find() (func(string) error, error) {
	git, err := exec.LookPath("git")
	if err != nil {
		return nil, fmt.Errorf("Git não encontrado no PATH: %w", err)
	}

	return func(directory string) error {
		command := exec.Command(git, "init", "--quiet")
		command.Dir = directory
		command.Env = cleanEnvironment(os.Environ())
		if output, err := command.CombinedOutput(); err != nil {
			return fmt.Errorf("não foi possível inicializar Git em %q: %w: %s",
				directory, err, strings.TrimSpace(string(output)))
		}
		return nil
	}, nil
}

func cleanEnvironment(environment []string) []string {
	clean := make([]string, 0, len(environment))
	for _, entry := range environment {
		name, _, _ := strings.Cut(entry, "=")
		remove := strings.HasPrefix(strings.ToUpper(name), "GIT_")
		for redirected := range redirectedVariables {
			remove = remove || strings.EqualFold(name, redirected)
		}
		if !remove {
			clean = append(clean, entry)
		}
	}
	return clean
}
