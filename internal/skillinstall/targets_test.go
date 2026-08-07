package skillinstall

import (
	"path/filepath"
	"testing"
)

func TestTargetPathSupportsOnlyCodexAndClaude(t *testing.T) {
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
}
