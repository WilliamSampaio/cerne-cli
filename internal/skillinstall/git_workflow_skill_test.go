package skillinstall

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGitWorkflowSkillRequiresReviewedChangesConfirmationBeforeCommit(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("bundle", "skills", GitWorkflowSkill, "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, want := range []string{
		"reinspect before each step",
		"fresh confirmation for that step only",
		"before a commit, ask the user to confirm they reviewed every included change",
		"Ask the equivalent of this safety question in the response language",
		"Do not quote it in English unless English is the response language",
		"use explicit changed paths for commit",
		"use explicit remote and branch for push",
		"Cerne provides the inspected repository names, paths, branches, remotes, changed paths, and",
		"create Pull Requests only on GitHub.com using the agent's native capability",
		"do not read, request, store, or pass GitHub tokens through Cerne",
		"stop after the first refusal, blocked state, failed result, or partial result",
		"Prefer the repository's documented commit convention",
		"Agent skill activation or consent to load this skill is not authorization for Git",
		"Refuse destructive, ambiguous, arbitrary, or out-of-scope Git operations",
		"Do not ask Cerne to execute Git effects",
		"partial results, name what",
		"reinspect before deciding",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing commit review confirmation %q", want)
		}
	}
}

func TestBundledSkillsFollowConversationOrCerneLanguage(t *testing.T) {
	for _, skill := range []string{SkillName, GitWorkflowSkill} {
		data, err := os.ReadFile(filepath.Join("bundle", "skills", skill, "SKILL.md"))
		if err != nil {
			t.Fatal(err)
		}
		text := string(data)
		for _, want := range []string{
			"Respond in the language established by the active conversation",
			"use the effective Cerne language",
			"Do not translate commands, flags, paths",
		} {
			if !strings.Contains(text, want) {
				t.Errorf("%s missing language rule %q", skill, want)
			}
		}
	}
}
