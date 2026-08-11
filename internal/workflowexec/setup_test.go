package workflowexec

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/WilliamSampaio/cerne-cli/internal/workspace"
)

func TestDescribeProvidersWithoutLookingAtPATH(t *testing.T) {
	original := lookPath
	lookPath = func(string) (string, error) { panic("Describe consultou PATH") }
	defer func() { lookPath = original }()

	tests := []struct {
		provider string
		expected workspace.WorkflowDefinition
	}{
		{"speckit", workspace.WorkflowDefinition{Provider: "speckit", Executor: "specify", CanonicalSpecs: "specs", OwnedRoot: ".specify", Marker: ".specify/init-options.json"}},
		{"openspec", workspace.WorkflowDefinition{Provider: "openspec", Executor: "openspec", CanonicalSpecs: "openspec/specs", OwnedRoot: "openspec", Marker: "openspec/config.yaml"}},
	}
	for _, test := range tests {
		t.Run(test.provider, func(t *testing.T) {
			actual, err := Describe(test.provider)
			if err != nil || !reflect.DeepEqual(actual, test.expected) {
				t.Fatalf("Describe() = %#v, %v; esperado %#v", actual, err, test.expected)
			}
		})
	}
	if _, err := Describe("unknown"); !errors.Is(err, ErrUnknownProvider) {
		t.Fatalf("erro = %v, esperado ErrUnknownProvider", err)
	}
}

func TestResolveDefinitionsAndExactInvocations(t *testing.T) {
	originalLookPath, originalRun, originalTempDir, originalRemoveAll := lookPath, runProvider, tempDir, removeAll
	t.Cleanup(func() {
		lookPath, runProvider, tempDir, removeAll = originalLookPath, originalRun, originalTempDir, originalRemoveAll
	})
	lookPath = func(name string) (string, error) { return filepath.Join(t.TempDir(), name), nil }
	tempDir = func(string, string) (string, error) { return filepath.Join(t.TempDir(), "provider-home"), nil }
	removeAll = func(string) error { return nil }

	for _, test := range []struct {
		provider, executor, specs, root, marker string
		arguments                               []string
	}{
		{"speckit", "specify", "specs", ".specify", ".specify/init-options.json", nil},
		{"openspec", "openspec", "openspec/specs", "openspec", "openspec/config.yaml", nil},
	} {
		t.Run(test.provider, func(t *testing.T) {
			knowledge := filepath.Join(t.TempDir(), "knowledge with space")
			if err := os.Mkdir(knowledge, 0o755); err != nil {
				t.Fatal(err)
			}
			var gotExecutable, gotDirectory string
			var gotArguments, gotEnvironment []string
			runProvider = func(executable string, arguments []string, directory string, environment []string) error {
				gotExecutable, gotDirectory = executable, directory
				gotArguments, gotEnvironment = arguments, environment
				return nil
			}
			definition, err := Resolve(test.provider)
			if err != nil || !definition.Available || definition.Executor != test.executor || definition.CanonicalSpecs != test.specs || definition.OwnedRoot != test.root || definition.Marker != test.marker {
				t.Fatalf("definition=%#v erro=%v", definition, err)
			}
			if err := definition.Setup(knowledge); err != nil {
				t.Fatal(err)
			}
			if !filepath.IsAbs(gotExecutable) || gotDirectory != knowledge {
				t.Fatalf("exec=%q dir=%q", gotExecutable, gotDirectory)
			}
			if test.provider == "openspec" {
				test.arguments = []string{"init", knowledge, "--tools", "none", "--profile", "core", "--no-animation"}
				for _, value := range []string{"OPENSPEC_TELEMETRY=0", "DO_NOT_TRACK=1", "NO_COLOR=1"} {
					if !hasEnvironment(gotEnvironment, value) {
						t.Fatalf("ambiente sem %s: %v", value, gotEnvironment)
					}
				}
			} else {
				script := "sh"
				if runtime.GOOS == "windows" {
					script = "ps"
				}
				test.arguments = []string{"init", "--here", "--force", "--integration", "generic", "--integration-options=--commands-dir .specify/commands", "--ignore-agent-tools", "--script", script}
			}
			if strings.Join(gotArguments, "\x00") != strings.Join(test.arguments, "\x00") {
				t.Fatalf("args=%q esperado=%q", gotArguments, test.arguments)
			}
		})
	}
}

func TestResolveSpecKitAgentTargetsAndExactInvocations(t *testing.T) {
	originalLookPath, originalRun, originalTempDir, originalRemoveAll := lookPath, runProvider, tempDir, removeAll
	t.Cleanup(func() {
		lookPath, runProvider, tempDir, removeAll = originalLookPath, originalRun, originalTempDir, originalRemoveAll
	})
	lookPath = func(name string) (string, error) { return filepath.Join(t.TempDir(), name), nil }
	tempDir = func(string, string) (string, error) { return filepath.Join(t.TempDir(), "provider-home"), nil }
	removeAll = func(string) error { return nil }

	var gotArguments []string
	runProvider = func(_ string, arguments []string, _ string, _ []string) error {
		gotArguments = arguments
		return nil
	}
	definition, err := Resolve("speckit")
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		agent, root      string
		integrationRoots []string
		arguments        []string
	}{
		{"codex", ".agents/skills", []string{".agents/skills", "skills"}, []string{"integration", "install", "codex", "--force", "--integration-options=--skills"}},
		{"claude", ".claude/skills", nil, []string{"integration", "install", "claude", "--force"}},
	} {
		t.Run(test.agent, func(t *testing.T) {
			target, ok := definition.Agents[test.agent]
			if !ok || target.Name != test.agent || target.DiscoveryRoot != test.root || !reflect.DeepEqual(target.IntegrationRoots, test.integrationRoots) || target.Setup == nil {
				t.Fatalf("target=%#v ok=%v", target, ok)
			}
			if err := target.Setup(filepath.Join(t.TempDir(), "knowledge")); err != nil {
				t.Fatal(err)
			}
			if strings.Join(gotArguments, "\x00") != strings.Join(test.arguments, "\x00") {
				t.Fatalf("args=%q esperado=%q", gotArguments, test.arguments)
			}
		})
	}
}

func TestResolveMissingUnknownAndSanitizesEnvironmentAndFailure(t *testing.T) {
	originalLookPath, originalRun, originalTempDir, originalRemoveAll := lookPath, runProvider, tempDir, removeAll
	t.Cleanup(func() {
		lookPath, runProvider, tempDir, removeAll = originalLookPath, originalRun, originalTempDir, originalRemoveAll
	})
	lookPath = func(name string) (string, error) {
		if name == "specify" {
			return "", errors.New("missing")
		}
		return filepath.Join(t.TempDir(), name), nil
	}
	missing, err := Resolve("speckit")
	if err != nil || missing.Available || missing.Setup != nil {
		t.Fatalf("missing=%#v erro=%v", missing, err)
	}
	if _, err := Resolve("other"); !errors.Is(err, ErrUnknownProvider) {
		t.Fatalf("erro=%v", err)
	}

	t.Setenv("CERNE_SECRET_TOKEN", "never-pass-this")
	t.Setenv("HOME", "do-not-pass-this-home")
	t.Setenv("PATH", os.Getenv("PATH"))
	profile := filepath.Join(t.TempDir(), "provider-home")
	tempDir = func(string, string) (string, error) { return profile, nil }
	removeAll = func(path string) error {
		if path != profile {
			t.Fatalf("cleanup path=%q want %q", path, profile)
		}
		return nil
	}
	var environment []string
	runProvider = func(_ string, _ []string, _ string, env []string) error {
		environment = env
		return errors.New("raw-token-never-return")
	}
	definition, err := Resolve("openspec")
	if err != nil {
		t.Fatal(err)
	}
	err = definition.Setup(t.TempDir())
	if err == nil || strings.Contains(err.Error(), "raw-token") {
		t.Fatalf("erro=%v", err)
	}
	if strings.Contains(strings.Join(environment, "\n"), "never-pass-this") {
		t.Fatalf("segredo no ambiente: %v", environment)
	}
	if strings.Contains(strings.Join(environment, "\n"), "do-not-pass-this-home") {
		t.Fatalf("provider HOME inseguro: %v", environment)
	}
	if runtime.GOOS == "windows" {
		if !hasEnvironment(environment, "USERPROFILE="+profile) {
			t.Fatalf("provider USERPROFILE ausente: %v", environment)
		}
	} else if !hasEnvironment(environment, "HOME="+profile) {
		t.Fatalf("provider HOME ausente: %v", environment)
	}
}

func hasEnvironment(environment []string, expected string) bool {
	for _, entry := range environment {
		if entry == expected {
			return true
		}
	}
	return false
}
