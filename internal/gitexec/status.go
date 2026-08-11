package gitexec

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	DetachedHEAD = "detached-head"
	NoCommits    = "no-commits"
)

type RepositoryStatus struct {
	Path           string
	Branch         string
	Commit         string
	ModifiedCount  int
	StagedCount    int
	UntrackedCount int
}

func FindStatus() (func(string) (RepositoryStatus, error), error) {
	git, err := exec.LookPath("git")
	if err != nil {
		return nil, fmt.Errorf("Git não encontrado no PATH: %w", err)
	}
	return func(directory string) (RepositoryStatus, error) {
		return collectStatus(git, directory)
	}, nil
}

func collectStatus(git, directory string) (RepositoryStatus, error) {
	root, err := statusOutput(git, directory, "rev-parse", "--show-toplevel")
	if err != nil {
		return RepositoryStatus{}, err
	}
	rootPath := cleanPath(root)
	requestedPath := cleanPath(directory)
	sameRoot := rootPath == requestedPath
	rootInfo, rootErr := os.Stat(rootPath)
	requestedInfo, requestedErr := os.Stat(requestedPath)
	if rootErr == nil && requestedErr == nil {
		sameRoot = os.SameFile(rootInfo, requestedInfo)
	}
	if !sameRoot {
		return RepositoryStatus{}, fmt.Errorf("%q não é raiz Git própria", directory)
	}
	if _, err := statusOutput(git, directory, "rev-parse", "--git-common-dir"); err != nil {
		return RepositoryStatus{}, err
	}

	branch, err := statusOutput(git, directory, "symbolic-ref", "--quiet", "--short", "HEAD")
	if err != nil {
		branch = DetachedHEAD
	}
	commit, err := statusOutput(git, directory, "rev-parse", "--verify", "--short=7", "HEAD")
	if err != nil {
		commit = NoCommits
	}
	modified, err := statusOutput(git, directory, "diff", "--name-only")
	if err != nil {
		return RepositoryStatus{}, err
	}
	staged, err := statusOutput(git, directory, "diff", "--cached", "--name-only")
	if err != nil {
		return RepositoryStatus{}, err
	}
	untracked, err := statusOutput(git, directory, "ls-files", "--others", "--exclude-standard")
	if err != nil {
		return RepositoryStatus{}, err
	}

	return RepositoryStatus{
		Path:           cleanPath(root),
		Branch:         branch,
		Commit:         commit,
		ModifiedCount:  countGitLines(modified),
		StagedCount:    countGitLines(staged),
		UntrackedCount: countGitLines(untracked),
	}, nil
}

func statusOutput(git, directory string, args ...string) (string, error) {
	command := exec.Command(git, append(readOnlyGitArgs(directory), args...)...)
	command.Env = gitEnvironment(os.Environ())
	output, err := command.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("consulta Git falhou em %q: %w", directory, err)
	}
	return strings.TrimSpace(string(output)), nil
}

func countGitLines(output string) int {
	if strings.TrimSpace(output) == "" {
		return 0
	}
	return len(strings.Split(strings.TrimRight(output, "\r\n"), "\n"))
}
