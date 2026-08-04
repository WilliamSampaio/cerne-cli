package gitexec

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type LinkRepository struct {
	RequestedPath string
	WorktreeRoot  string
	CommonDir     string
	IsBare        bool
	HasWorktree   bool
}

func FindLinkInspector() (func(string) (LinkRepository, error), error) {
	git, err := exec.LookPath("git")
	if err != nil {
		return nil, fmt.Errorf("Git não encontrado no PATH: %w", err)
	}
	return func(directory string) (LinkRepository, error) {
		return inspectLinkRepository(git, directory)
	}, nil
}

func inspectLinkRepository(git, directory string) (LinkRepository, error) {
	requested, _ := filepath.Abs(directory)
	bare, err := linkOutput(git, directory, "rev-parse", "--is-bare-repository")
	if err != nil {
		return LinkRepository{}, err
	}
	if bare == "true" {
		return LinkRepository{RequestedPath: cleanPath(requested), IsBare: true}, nil
	}
	root, err := linkOutput(git, directory, "rev-parse", "--show-toplevel")
	if err != nil {
		return LinkRepository{}, err
	}
	common, err := linkOutput(git, directory, "rev-parse", "--git-common-dir")
	if err != nil {
		return LinkRepository{}, err
	}
	if !filepath.IsAbs(common) {
		common = filepath.Join(root, common)
	}
	return LinkRepository{
		RequestedPath: cleanPath(requested),
		WorktreeRoot:  cleanPath(root),
		CommonDir:     cleanPath(common),
		HasWorktree:   true,
	}, nil
}

func linkOutput(git, directory string, args ...string) (string, error) {
	command := exec.Command(git, append([]string{"-C", directory}, args...)...)
	command.Env = gitEnvironment(os.Environ())
	output, err := command.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("consulta Git falhou em %q: %w", directory, err)
	}
	return strings.TrimSpace(string(output)), nil
}
