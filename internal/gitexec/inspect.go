package gitexec

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Repository struct {
	RequestedRoot string
	WorktreeRoot  string
	CommonDir     string
}

func FindInspector() (func(string) (Repository, error), error) {
	git, err := exec.LookPath("git")
	if err != nil {
		return nil, fmt.Errorf("Git não encontrado no PATH: %w", err)
	}
	return func(directory string) (Repository, error) {
		root, err := revParse(git, directory, "--show-toplevel")
		if err != nil {
			return Repository{}, err
		}
		common, err := revParse(git, directory, "--git-common-dir")
		if err != nil {
			return Repository{}, err
		}
		if !filepath.IsAbs(common) {
			common = filepath.Join(directory, common)
		}
		requested, _ := filepath.Abs(directory)
		return Repository{
			RequestedRoot: cleanPath(requested),
			WorktreeRoot:  cleanPath(root),
			CommonDir:     cleanPath(common),
		}, nil
	}, nil
}

func revParse(git, directory, flag string) (string, error) {
	command := exec.Command(git, "-C", directory, "rev-parse", flag)
	command.Env = gitEnvironment(os.Environ())
	output, err := command.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("Git não reconheceu %q como repositório local: %w", directory, err)
	}
	return strings.TrimSpace(string(output)), nil
}

func gitEnvironment(environment []string) []string {
	clean := cleanEnvironment(environment)
	clean = append(clean, "GIT_OPTIONAL_LOCKS=0", "GIT_TERMINAL_PROMPT=0")
	return clean
}

func cleanPath(path string) string {
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		path = resolved
	}
	return filepath.Clean(path)
}
