package main

import (
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/WilliamSampaio/cerne-cli/internal/localization"
)

func TestCLICompletionHelp(t *testing.T) {
	binary := buildCLI(t)
	status, stdout, stderr := executeCLI(t, binary, t.TempDir(), nil, "completion", "--help")
	if status != 0 || stderr != "" ||
		!strings.Contains(stdout, "cerne completion <bash|zsh>") ||
		!strings.Contains(stdout, `eval "$(cerne completion bash)"`) ||
		!strings.Contains(stdout, `eval "$(cerne completion zsh)"`) {
		t.Fatalf("status = %d\nstdout = %q\nstderr = %q", status, stdout, stderr)
	}
}

func TestCLIGlobalHelpListsCompletion(t *testing.T) {
	binary := buildCLI(t)
	status, stdout, stderr := executeCLI(t, binary, t.TempDir(), nil, "--help")
	if status != 0 || stderr != "" || !strings.Contains(stdout, "completion") {
		t.Fatalf("status = %d\nstdout = %q\nstderr = %q", status, stdout, stderr)
	}
}

func runBashCompletion(t *testing.T, compWords string) []string {
	t.Helper()
	return runBashCompletionAt(t, compWords, 1)
}

func runBashCompletionAt(t *testing.T, compWords string, cword int) []string {
	t.Helper()
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash não está disponível")
	}
	script := bashCompletionScriptText() + "\n" +
		"COMP_WORDS=(" + compWords + ")\n" +
		"COMP_CWORD=" + strconv.Itoa(cword) + "\n" +
		"_cerne_completions\n" +
		`printf '%s\n' "${COMPREPLY[@]}"` + "\n"
	out, err := exec.Command("bash", "-c", script).Output()
	if err != nil {
		t.Fatalf("bash -c: %v", err)
	}
	var result []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			result = append(result, line)
		}
	}
	sort.Strings(result)
	return result
}

func TestBashCompletionSuggestsSubcommands(t *testing.T) {
	want := strings.Fields(completionSubcommands)
	sort.Strings(want)

	got := runBashCompletion(t, `cerne ""`)
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("cerne <TAB> = %v, want %v", got, want)
	}
	for _, global := range []string{"--help", "--version", "--lang"} {
		for _, suggestion := range got {
			if suggestion == global {
				t.Fatalf("cerne <TAB> suggested global flag %q, want only subcommands (FR-007)", global)
			}
		}
	}

	gotPrefix := runBashCompletion(t, "cerne d")
	if len(gotPrefix) != 1 || gotPrefix[0] != "doctor" {
		t.Fatalf("cerne d<TAB> = %v, want [doctor]", gotPrefix)
	}
}

func TestBashCompletionSuggestsSecondLevelActions(t *testing.T) {
	cases := map[string][]string{
		"cerne config":   {"get", "set", "unset"},
		"cerne skill":    {"install"},
		"cerne workflow": {"setup"},
		"cerne git":      {"inspect"},
	}
	for compWords, want := range cases {
		sort.Strings(want)
		t.Run(compWords, func(t *testing.T) {
			got := runBashCompletionAt(t, compWords, 2)
			if strings.Join(got, ",") != strings.Join(want, ",") {
				t.Fatalf("%s <TAB> = %v, want %v", compWords, got, want)
			}
		})
	}

	// Commands without a fixed second-level vocabulary (free-form arguments) suggest nothing.
	for _, compWords := range []string{"cerne doctor", "cerne status", "cerne init", "cerne link"} {
		t.Run(compWords+" (no action)", func(t *testing.T) {
			got := runBashCompletionAt(t, compWords, 2)
			if len(got) != 0 {
				t.Fatalf("%s <TAB> = %v, want no suggestions", compWords, got)
			}
		})
	}
}

func TestBashCompletionSuggestsThirdAndFourthLevelValues(t *testing.T) {
	cases := []struct {
		compWords string
		cword     int
		want      []string
	}{
		{"cerne config set", 3, []string{"language"}},
		{"cerne config get", 3, []string{"language"}},
		{"cerne config unset", 3, []string{"language"}},
		{"cerne skill install", 3, []string{"claude", "codex", "gemini"}},
		{"cerne skill install codex", 4, []string{"cerne-context", "cerne-git-workflow", "cerne-product-discovery"}},
		{"cerne skill install claude", 4, []string{"cerne-context", "cerne-git-workflow", "cerne-product-discovery"}},
		{"cerne skill install gemini", 4, []string{"cerne-git-workflow"}},
		{"cerne workflow setup --runtime", 4, []string{"claude", "codex", "gemini"}},
		{"cerne git inspect --runtime", 4, []string{"claude", "codex", "gemini"}},
	}
	for _, c := range cases {
		want := append([]string(nil), c.want...)
		sort.Strings(want)
		t.Run(c.compWords, func(t *testing.T) {
			got := runBashCompletionAt(t, c.compWords, c.cword)
			if strings.Join(got, ",") != strings.Join(want, ",") {
				t.Fatalf("%s <TAB> (cword=%d) = %v, want %v", c.compWords, c.cword, got, want)
			}
		})
	}

	// Commands without a fixed third/fourth-level vocabulary suggest nothing at that position.
	for _, c := range []struct {
		compWords string
		cword     int
	}{
		{"cerne doctor", 2},
		{"cerne skill install unknown-agent", 4},
	} {
		t.Run(c.compWords+" (no value)", func(t *testing.T) {
			got := runBashCompletionAt(t, c.compWords, c.cword)
			if len(got) != 0 {
				t.Fatalf("%s <TAB> = %v, want no suggestions", c.compWords, got)
			}
		})
	}
}

func TestZshCompletionContainsSubcommandsAndIsValidSyntax(t *testing.T) {
	if _, err := exec.LookPath("zsh"); err != nil {
		t.Skip("zsh não está disponível")
	}
	script := zshCompletionScriptText(localizer{language: localization.Default})
	for _, name := range strings.Fields(completionSubcommands) {
		if !strings.Contains(script, name) {
			t.Errorf("zshCompletionScriptText não contém o subcomando %q", name)
		}
	}
	for _, global := range []string{"--help", "--version", "--lang"} {
		if strings.Contains(script, global) {
			t.Errorf("zshCompletionScriptText contém a flag global %q, quer só subcomandos (FR-007)", global)
		}
	}
	if !strings.Contains(script, "CURRENT == 2") {
		t.Error(`zshCompletionScriptText não restringe a sugestão à primeira palavra (esperado "CURRENT == 2"); sem isso, "cerne <subcomando> <TAB>" volta a sugerir os subcomandos principais`)
	}
	for _, action := range []string{"set", "get", "unset", "install", "setup", "inspect"} {
		if !strings.Contains(script, action) {
			t.Errorf("zshCompletionScriptText não contém a ação de segundo nível %q", action)
		}
	}
	for _, value := range []string{"language", "codex", "claude", "gemini",
		"cerne-context", "cerne-product-discovery", "cerne-git-workflow"} {
		if !strings.Contains(script, value) {
			t.Errorf("zshCompletionScriptText não contém o valor de terceiro/quarto nível %q", value)
		}
	}
	if !strings.Contains(script, "CURRENT == 4") || !strings.Contains(script, "CURRENT == 5") {
		t.Error(`zshCompletionScriptText não cobre CURRENT == 4 / CURRENT == 5 (terceiro e quarto nível)`)
	}
	if !strings.Contains(script, "--runtime") {
		t.Error(`zshCompletionScriptText não detecta a flag "--runtime" para completar o agente`)
	}
	for _, desc := range []string{"Creates a Cerne workspace", "Saves the language preference",
		"Codex agent runtime", "Safe Git inspection skill"} {
		if !strings.Contains(script, desc) {
			t.Errorf("zshCompletionScriptText não contém a descrição %q (usada pelo _describe para o item selecionado)", desc)
		}
	}

	file, err := os.CreateTemp(t.TempDir(), "cerne-completion-*.zsh")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.WriteString(script); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("zsh", "-n", file.Name()).CombinedOutput(); err != nil {
		t.Fatalf("zsh -n reportou erro de sintaxe: %v\n%s", err, out)
	}
}

func TestZshCompletionLoadsWithoutPriorCompinit(t *testing.T) {
	if _, err := exec.LookPath("zsh"); err != nil {
		t.Skip("zsh não está disponível")
	}
	// -f evita carregar rc files, simulando uma sessão onde compinit ainda não rodou —
	// sem o guard em zshCompletionScriptText, "compdef" não existe e o source falha.
	script := zshCompletionScriptText(localizer{language: localization.Default}) + "\nwhence -w _cerne\n"
	out, err := exec.Command("zsh", "-f", "-c", script).CombinedOutput()
	if err != nil {
		t.Fatalf("zsh -f -c falhou ao carregar o script sem compinit prévio: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "_cerne: function") {
		t.Fatalf("_cerne não foi definida como função após carregar o script:\n%s", out)
	}
}

func TestCLICompletionBashNoSideEffectsOutsideWorkspace(t *testing.T) {
	binary := buildCLI(t)
	home := t.TempDir()
	before, err := os.ReadDir(home)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	env := replaceEnvironment(portugueseTestEnvironment(), "HOME", home)
	status, stdout, stderr := executeCLI(t, binary, dir, env, "completion", "bash")
	if status != 0 || stdout == "" || stderr != "" {
		t.Fatalf("status = %d\nstdout empty = %v\nstderr = %q", status, stdout == "", stderr)
	}
	after, err := os.ReadDir(home)
	if err != nil {
		t.Fatal(err)
	}
	if len(after) != len(before) {
		t.Fatalf("HOME ganhou arquivo(s) novo(s): antes=%v depois=%v", before, after)
	}
}

func TestCLICompletionInvalidUsage(t *testing.T) {
	binary := buildCLI(t)
	expected := "erro: shell inválido\nuso: cerne completion <bash|zsh>\n"

	cases := map[string][]string{
		"sem argumento":      {"completion"},
		"shell desconhecido": {"completion", "fish"},
		"argumentos extras":  {"completion", "bash", "extra"},
	}
	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			status, stdout, stderr := executeCLI(t, binary, t.TempDir(), nil, args...)
			if status != 2 || stdout != "" || stderr != expected {
				t.Fatalf("status = %d\nstdout = %q\nstderr = %q", status, stdout, stderr)
			}
		})
	}
}
