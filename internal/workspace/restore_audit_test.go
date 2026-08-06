package workspace

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestRestoreAuditIsPrivateUniqueAndRejectsSymlink(t *testing.T) {
	home := t.TempDir()
	first, attempt, err := startRestoreAudit(home, SourceClone)
	if err != nil {
		t.Fatal(err)
	}
	defer first.root.Close()
	second, _, err := startRestoreAudit(home, SourceClone)
	if err != nil {
		t.Fatal(err)
	}
	defer second.root.Close()
	if first.path == second.path || attempt.Authorization != "restore --clone" {
		t.Fatalf("audits não exclusivos: %q %q", first.path, second.path)
	}
	if runtime.GOOS != "windows" {
		for path, want := range map[string]os.FileMode{
			filepath.Join(home, ".cerne"):          0o700,
			filepath.Join(home, ".cerne", "audit"): 0o700,
			first.path:                             0o600,
		} {
			info, err := os.Stat(path)
			if err != nil || info.Mode().Perm() != want {
				t.Fatalf("permissão de %s = %v, %v", path, info.Mode().Perm(), err)
			}
		}
	}

	if runtime.GOOS == "windows" {
		return
	}
	unsafeHome := t.TempDir()
	if err := os.Symlink(t.TempDir(), filepath.Join(unsafeHome, ".cerne")); err != nil {
		t.Fatal(err)
	}
	if _, _, err := startRestoreAudit(unsafeHome, SourceClone); err == nil {
		t.Fatal("audit deveria recusar .cerne symlink")
	}
}
