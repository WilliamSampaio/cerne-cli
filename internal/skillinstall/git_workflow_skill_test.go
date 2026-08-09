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
		"Do you confirm that you reviewed every change included in this commit?",
		"use explicit changed paths for commit",
		"use explicit remote and branch for push",
		"open Pull Requests only on GitHub.com when a token is externally available",
		"stop after the first refusal, blocked state, failed result, or partial result",
		"Prefer the repository's documented commit convention",
		"Agent skill activation or consent to load this skill is not authorization for Git",
		"Refuse destructive, ambiguous, arbitrary, or out-of-scope Git operations",
		"Do not pass arbitrary Git flags, URLs, credentials, free refspecs, absolute paths, path traversal, or",
		"partial results, name what",
		"reinspect before deciding",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing commit review confirmation %q", want)
		}
	}
}
