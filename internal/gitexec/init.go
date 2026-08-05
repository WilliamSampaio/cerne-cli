package gitexec

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
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

var ErrInvalidCloneOrigin = errors.New("origem de clone inválida")

type CloneOrigin struct {
	Location    string
	Transport   string
	Fingerprint string
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

func ClassifyCloneOrigin(start, input string) (CloneOrigin, error) {
	if input == "" || strings.HasPrefix(input, "-") {
		return CloneOrigin{}, ErrInvalidCloneOrigin
	}
	candidate := input
	if !filepath.IsAbs(candidate) {
		candidate = filepath.Join(start, candidate)
	}
	if info, err := os.Stat(candidate); err == nil {
		if !info.IsDir() {
			return CloneOrigin{}, ErrInvalidCloneOrigin
		}
		absolute, err := filepath.Abs(candidate)
		if err != nil {
			return CloneOrigin{}, ErrInvalidCloneOrigin
		}
		return cloneOrigin(filepath.Clean(absolute), "local", input), nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return CloneOrigin{}, ErrInvalidCloneOrigin
	}

	parsed, err := url.Parse(input)
	if err == nil {
		switch strings.ToLower(parsed.Scheme) {
		case "file":
			if parsed.User == nil && parsed.RawQuery == "" && parsed.Fragment == "" && parsed.Path != "" {
				return cloneOrigin(input, "file", input), nil
			}
		case "https":
			if parsed.User == nil && parsed.Host != "" && parsed.RawQuery == "" && parsed.Fragment == "" {
				return cloneOrigin(input, "https", input), nil
			}
		case "ssh":
			password := false
			if parsed.User != nil {
				_, password = parsed.User.Password()
			}
			if !password && parsed.Host != "" && parsed.Path != "" && parsed.RawQuery == "" && parsed.Fragment == "" {
				return cloneOrigin(input, "ssh", input), nil
			}
		}
	}
	if scpLike(input) {
		return cloneOrigin(input, "ssh", input), nil
	}
	return CloneOrigin{}, ErrInvalidCloneOrigin
}

func cloneOrigin(location, transport, fingerprintInput string) CloneOrigin {
	fingerprint := sha256.Sum256([]byte(fingerprintInput))
	return CloneOrigin{Location: location, Transport: transport, Fingerprint: fmt.Sprintf("%x", fingerprint)}
}

func scpLike(input string) bool {
	if strings.ContainsAny(input, "\r\n\t ?#") || strings.Contains(input, "://") || strings.Contains(input, "::") {
		return false
	}
	left, path, ok := strings.Cut(input, ":")
	if !ok || left == "" || path == "" || strings.ContainsAny(left, `/\`) || strings.Contains(path, "@") {
		return false
	}
	if len(left) == 1 && ((left[0] >= 'A' && left[0] <= 'Z') || (left[0] >= 'a' && left[0] <= 'z')) {
		return false
	}
	userHost := strings.Split(left, "@")
	return len(userHost) <= 2 && userHost[len(userHost)-1] != ""
}

var runCloneCommand = func(git string, arguments, environment []string) error {
	command := exec.Command(git, arguments...)
	command.Env = environment
	return command.Run()
}

func FindClone() (func(string, string) error, error) {
	git, err := exec.LookPath("git")
	if err != nil {
		return nil, fmt.Errorf("Git não encontrado no PATH: %w", err)
	}
	return cloneWithGit(git), nil
}

func cloneWithGit(git string) func(string, string) error {
	return func(origin, destination string) error {
		configRoot, err := os.MkdirTemp("", "cerne-git-clone-")
		if err != nil {
			return errors.New("não foi possível preparar o clone Git")
		}
		defer os.RemoveAll(configRoot)
		if err := os.Chmod(configRoot, 0o700); err != nil {
			return errors.New("não foi possível preparar o clone Git")
		}
		hooks := filepath.Join(configRoot, "hooks")
		templates := filepath.Join(configRoot, "templates")
		if err := os.Mkdir(hooks, 0o700); err != nil {
			return errors.New("não foi possível preparar o clone Git")
		}
		if err := os.Mkdir(templates, 0o700); err != nil {
			return errors.New("não foi possível preparar o clone Git")
		}
		arguments := []string{
			"-c", "credential.interactive=false",
			"-c", "protocol.allow=never",
			"-c", "protocol.file.allow=always",
			"-c", "protocol.https.allow=always",
			"-c", "protocol.ssh.allow=always",
			"-c", "core.hooksPath=" + hooks,
			"clone", "--quiet", "--origin=origin", "--no-local", "--template=" + templates,
			"--", origin, destination,
		}
		environment := cleanEnvironment(os.Environ())
		environment = append(environment,
			"GIT_TERMINAL_PROMPT=0", "GCM_INTERACTIVE=Never", "GIT_ASKPASS=",
			"SSH_ASKPASS=", "SSH_ASKPASS_REQUIRE=never")
		if err := runCloneCommand(git, arguments, environment); err != nil {
			return errors.New("clone Git falhou")
		}
		return nil
	}
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
