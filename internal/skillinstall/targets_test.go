package skillinstall

import (
	"path/filepath"
	"testing"
)

func TestTargetPathSupportsAgentsAndSkills(t *testing.T) {
	home := t.TempDir()
	tests := map[string]bool{"codex": true, "claude": true, "generic": false, "Codex": false, "": false}
	for agent, ok := range tests {
		target, err := TargetPath(home, agent)
		if ok && (err != nil || target == "") {
			t.Fatalf("%s: target=%q err=%v", agent, target, err)
		}
		if !ok && err == nil {
			t.Fatalf("%s should be rejected", agent)
		}
	}
	codex, _ := TargetPath(home, "codex")
	if codex != filepath.Join(home, ".codex", "skills", SkillName) {
		t.Fatalf("codex target = %q", codex)
	}
	gitSkill, err := TargetPath(home, "gemini", GitWorkflowSkill)
	if err != nil {
		t.Fatal(err)
	}
	if gitSkill != filepath.Join(home, ".gemini", "skills", GitWorkflowSkill) {
		t.Fatalf("gemini target = %q", gitSkill)
	}
	for _, args := range []struct {
		agent string
		skill string
	}{
		{"gemini", SkillName},
		{"codex", "unknown"},
		{"codex", "Cerne-Git-Workflow"},
		{"codex", "../cerne-git-workflow"},
	} {
		if _, err := TargetPath(home, args.agent, args.skill); err == nil {
			t.Fatalf("TargetPath(%q, %q) should fail", args.agent, args.skill)
		}
	}
}
