package skillinstall

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func productDiscoverySkillText(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("bundle", "skills", "cerne-product-discovery", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	return strings.Join(strings.Fields(string(data)), " ")
}

func TestProductDiscoverySkillDefinesSafeSharedContract(t *testing.T) {
	text := productDiscoverySkillText(t)
	for _, want := range []string{
		"name: cerne-product-discovery",
		"Evaluate product and feature ideas critically before specification",
		"Respond in the language established by the active conversation",
		"use the effective Cerne language",
		"Do not translate commands, flags, paths",
		"`cerne context --json`",
		"schema_version",
		"status is `invalid`",
		"Do not create or modify files, issues, branches, commits, pushes, or Pull Requests by default",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("missing shared contract %q", want)
		}
	}
}

func TestProductDiscoverySkillEvaluatesIdeasWithoutFixedQuestionnaire(t *testing.T) {
	text := productDiscoverySkillText(t)
	for _, want := range []string{
		"product or feature idea",
		"bugs, refactors, technical debt, and purely architectural decisions",
		"Use information already available",
		"Do not ask the user to repeat",
		"Ask exactly one question at a time",
		"materially change the recommendation",
		"Explain why the answer matters",
		"facts, assumptions, unknowns, risks, and recommendations",
		"`explorar`",
		"`reformular`",
		"`estacionar`",
		"`abandonar`",
		"`especificar`",
		"specific user",
		"current problem",
		"desired outcome",
		"scope boundaries and anti-goals",
		"observable success signals",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("missing discovery contract %q", want)
		}
	}
}

func TestProductDiscoverySkillRequiresFreshAuthorizedHandoff(t *testing.T) {
	text := productDiscoverySkillText(t)
	for _, want := range []string{
		"Run `cerne context --json` again",
		"provider is `speckit`",
		"state is `ready`",
		"fresh explicit authorization",
		"Do not reuse earlier or generic consent",
		"activate the available `speckit-specify` skill",
		"user, problem, desired outcome, boundaries, relevant assumptions, and success signals",
		"secrets, credentials, or unrelated private context",
		"absent, pending, invalid, changed, or unsupported",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("missing handoff contract %q", want)
		}
	}
}
