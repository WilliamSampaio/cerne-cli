package skillinstall

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestInstallUsesEmbeddedPackageByDefault(t *testing.T) {
	home := t.TempDir()
	result, err := Install("codex", Options{HomeDir: home})
	if err != nil {
		t.Fatal(err)
	}
	if result.Outcome != "installed" || result.Version != "0.5.0" || result.Skill != SkillName {
		t.Fatalf("result = %#v", result)
	}
	skill := readText(t, filepath.Join(result.Destination, "SKILL.md"))
	if !strings.Contains(skill, "name: cerne-context") {
		t.Fatal("embedded skill not installed")
	}
}

func TestInstallNamedProductDiscoverySkill(t *testing.T) {
	for _, agent := range []string{"codex", "claude"} {
		t.Run(agent, func(t *testing.T) {
			home := t.TempDir()
			result, err := Install(agent, Options{HomeDir: home, PackageDir: packageFixture(t, "1.0.0"), Skill: ProductDiscoverySkill})
			if err != nil {
				t.Fatal(err)
			}
			if result.Skill != ProductDiscoverySkill || result.Outcome != "installed" {
				t.Fatalf("result = %#v", result)
			}
			if readText(t, filepath.Join(result.Destination, "SKILL.md")) != "# Product\n" {
				t.Fatal("product discovery skill not copied")
			}
		})
	}
	home := t.TempDir()
	if _, err := Install("gemini", Options{HomeDir: home, PackageDir: packageFixture(t, "1.0.0"), Skill: ProductDiscoverySkill}); err == nil {
		t.Fatal("gemini should reject product discovery skill")
	}
	if _, err := os.Stat(filepath.Join(home, ".cerne", "audit")); !os.IsNotExist(err) {
		t.Fatalf("rejected product skill created audit: %v", err)
	}
}

func TestInstallNamedGitWorkflowSkill(t *testing.T) {
	for _, agent := range []string{"codex", "claude", "gemini"} {
		t.Run(agent, func(t *testing.T) {
			home := t.TempDir()
			result, err := Install(agent, Options{HomeDir: home, PackageDir: packageFixture(t, "1.0.0"), Skill: GitWorkflowSkill})
			if err != nil {
				t.Fatal(err)
			}
			if result.Skill != GitWorkflowSkill || result.Outcome != "installed" {
				t.Fatalf("result = %#v", result)
			}
			if readText(t, filepath.Join(result.Destination, "SKILL.md")) != "# Git\n" {
				t.Fatal("git workflow skill not copied")
			}
			audit := readText(t, result.AuditPath)
			if !strings.Contains(audit, `"skill": "cerne-git-workflow"`) {
				t.Fatalf("audit did not record dynamic skill: %s", audit)
			}
		})
	}
}

func TestInstallCreatesDestinationAndPrivateAudit(t *testing.T) {
	home := t.TempDir()
	result, err := Install("codex", Options{HomeDir: home, PackageDir: packageFixture(t, "1.0.0")})
	if err != nil {
		t.Fatal(err)
	}
	if result.Outcome != "installed" || result.Version != "1.0.0" {
		t.Fatalf("result = %#v", result)
	}
	if readText(t, filepath.Join(home, ".codex", "skills", SkillName, "SKILL.md")) != "# Cerne\n" {
		t.Fatal("skill not copied")
	}
	audit := readText(t, result.AuditPath)
	if !strings.Contains(audit, `"status": "succeeded"`) || strings.Contains(audit, "# Cerne") || strings.Contains(audit, "HOME=") || strings.Contains(audit, "github.com") {
		t.Fatalf("unsafe audit: %s", audit)
	}
	if runtime.GOOS != "windows" {
		for path, want := range map[string]os.FileMode{
			filepath.Join(home, ".cerne"):          0o700,
			filepath.Join(home, ".cerne", "audit"): 0o700,
			result.AuditPath:                       0o600,
		} {
			info, err := os.Stat(path)
			if err != nil {
				t.Fatal(err)
			}
			if got := info.Mode().Perm(); got != want {
				t.Fatalf("%s mode=%o want=%o", path, got, want)
			}
		}
	}
}

func TestInstallSameVersionIsIdempotent(t *testing.T) {
	home := t.TempDir()
	pkg := packageFixture(t, "1.0.0")
	first, err := Install("claude", Options{HomeDir: home, PackageDir: pkg})
	if err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(first.Destination, "SKILL.md")
	before, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Install("claude", Options{HomeDir: home, PackageDir: pkg})
	if err != nil {
		t.Fatal(err)
	}
	after, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	if second.Outcome != "already" || !after.ModTime().Equal(before.ModTime()) {
		t.Fatalf("not idempotent: %#v before=%s after=%s", second, before.ModTime(), after.ModTime())
	}
}

func TestInstallRecognizesMarkerWrittenByPreviousBinary(t *testing.T) {
	home := t.TempDir()
	pkg := packageFixture(t, "1.0.0")
	first, err := Install("claude", Options{HomeDir: home, PackageDir: pkg})
	if err != nil {
		t.Fatal(err)
	}
	markerPath := filepath.Join(first.Destination, ".cerne-install.json")
	data, err := os.ReadFile(markerPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"agent": "claude"`) {
		t.Fatalf("marker does not preserve the on-disk \"agent\" field: %s", data)
	}
	target := filepath.Join(first.Destination, "SKILL.md")
	before, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Install("claude", Options{HomeDir: home, PackageDir: pkg})
	if err != nil {
		t.Fatal(err)
	}
	after, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	if second.Outcome != "already" || !after.ModTime().Equal(before.ModTime()) {
		t.Fatalf("marker written before the runtime rename was not recognized as managed: %#v", second)
	}
}

func TestInstallAuditUsesRuntimeFieldAndPreservesOlderRecords(t *testing.T) {
	home := t.TempDir()
	auditDir := filepath.Join(home, ".cerne", "audit")
	if err := os.MkdirAll(auditDir, 0o700); err != nil {
		t.Fatal(err)
	}
	older := `{"schema_version": 1, "operation": "skill.install", "agent": "codex", "skill": "cerne-context", "package": "cerne-skills", "destination": "/old", "status": "succeeded", "error_code": "", "started_at": "2026-01-01T00:00:00Z"}` + "\n"
	olderPath := filepath.Join(auditDir, "skill-install-older.json")
	if err := os.WriteFile(olderPath, []byte(older), 0o600); err != nil {
		t.Fatal(err)
	}

	result, err := Install("codex", Options{HomeDir: home, PackageDir: packageFixture(t, "1.0.0")})
	if err != nil {
		t.Fatal(err)
	}
	audit := readText(t, result.AuditPath)
	if !strings.Contains(audit, `"schema_version": 2`) || !strings.Contains(audit, `"runtime": "codex"`) || strings.Contains(audit, `"agent"`) {
		t.Fatalf("new audit record does not use the runtime field: %s", audit)
	}
	if got := readText(t, olderPath); got != older {
		t.Fatalf("older audit record was rewritten: %s", got)
	}
}

func TestInstallSameVersionRefreshesChangedManagedFiles(t *testing.T) {
	home := t.TempDir()
	pkg := packageFixture(t, "1.0.0")
	first, err := Install("claude", Options{HomeDir: home, PackageDir: pkg})
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(pkg, "skills", SkillName, "SKILL.md"), "# Cerne changed\n")
	second, err := Install("claude", Options{HomeDir: home, PackageDir: pkg})
	if err != nil {
		t.Fatal(err)
	}
	if second.Outcome != "upgraded" || first.Destination != second.Destination ||
		readText(t, filepath.Join(second.Destination, "SKILL.md")) != "# Cerne changed\n" {
		t.Fatalf("same-version refresh failed: %#v", second)
	}
}

func TestInstallUpgradesManagedDestination(t *testing.T) {
	home := t.TempDir()
	if _, err := Install("codex", Options{HomeDir: home, PackageDir: packageFixture(t, "1.0.0")}); err != nil {
		t.Fatal(err)
	}
	target, _ := TargetPath(home, "codex")
	writeFile(t, filepath.Join(target, "notes.txt"), "keep")
	pkg := packageFixture(t, "2.0.0")
	writeFile(t, filepath.Join(pkg, "skills", SkillName, "SKILL.md"), "# Cerne 2\n")
	result, err := Install("codex", Options{HomeDir: home, PackageDir: pkg})
	if err != nil {
		t.Fatal(err)
	}
	if result.Outcome != "upgraded" || readText(t, filepath.Join(result.Destination, "SKILL.md")) != "# Cerne 2\n" ||
		readText(t, filepath.Join(result.Destination, "notes.txt")) != "keep" {
		t.Fatalf("upgrade failed: %#v", result)
	}
	var m marker
	if err := json.Unmarshal([]byte(readText(t, filepath.Join(result.Destination, ".cerne-install.json"))), &m); err != nil {
		t.Fatal(err)
	}
	if m.Version != "2.0.0" {
		t.Fatalf("marker = %#v", m)
	}
}

func TestInstallFailsSafelyForMissingPackageAndUnknownDestination(t *testing.T) {
	t.Run("missing package", func(t *testing.T) {
		home := t.TempDir()
		result, err := Install("codex", Options{HomeDir: home, PackageDir: filepath.Join(t.TempDir(), "missing")})
		var failure Failure
		if !errors.As(err, &failure) || failure.Code != "package-unavailable" {
			t.Fatalf("err=%v result=%#v", err, result)
		}
		if _, statErr := os.Stat(result.Destination); !errors.Is(statErr, os.ErrNotExist) {
			t.Fatalf("destination mutated: %v", statErr)
		}
	})
	t.Run("unknown destination", func(t *testing.T) {
		home := t.TempDir()
		target, _ := TargetPath(home, "codex")
		writeFile(t, filepath.Join(target, "mine.txt"), "keep")
		_, err := Install("codex", Options{HomeDir: home, PackageDir: packageFixture(t, "1.0.0")})
		var failure Failure
		if !errors.As(err, &failure) || failure.Code != "unknown-destination" || readText(t, filepath.Join(target, "mine.txt")) != "keep" {
			t.Fatalf("err=%v", err)
		}
	})
}

func TestInstallAuditFailuresAndUpgradeRollbackStaySafe(t *testing.T) {
	t.Run("audit start failure", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("symlink audit check is covered on Unix")
		}
		home := t.TempDir()
		if err := os.Symlink(t.TempDir(), filepath.Join(home, ".cerne")); err != nil {
			t.Skipf("symlink indisponível: %v", err)
		}
		result, err := Install("codex", Options{HomeDir: home, PackageDir: packageFixture(t, "1.0.0")})
		var failure Failure
		if !errors.As(err, &failure) || failure.Code != "audit-start-failed" {
			t.Fatalf("err=%v", err)
		}
		if _, statErr := os.Stat(result.Destination); !errors.Is(statErr, os.ErrNotExist) {
			t.Fatalf("destination mutated: %v", statErr)
		}
	})

	t.Run("audit finalization failure", func(t *testing.T) {
		home := t.TempDir()
		calls := 0
		original := replaceAuditFile
		replaceAuditFile = func(old, new string) error {
			calls++
			if calls == 2 {
				return errors.New("disk full")
			}
			return os.Rename(old, new)
		}
		t.Cleanup(func() { replaceAuditFile = original })
		result, err := Install("codex", Options{HomeDir: home, PackageDir: packageFixture(t, "1.0.0")})
		var failure Failure
		if !errors.As(err, &failure) || failure.Code != "audit-finalization-failed" {
			t.Fatalf("err=%v", err)
		}
		if readText(t, filepath.Join(result.Destination, "SKILL.md")) != "# Cerne\n" {
			t.Fatal("safe promoted state not preserved")
		}
	})

	t.Run("upgrade rollback", func(t *testing.T) {
		home := t.TempDir()
		first, err := Install("codex", Options{HomeDir: home, PackageDir: packageFixture(t, "1.0.0")})
		if err != nil {
			t.Fatal(err)
		}
		original := renameInstallDir
		renameInstallDir = func(old, new string) error {
			if strings.Contains(filepath.Base(old), ".cerne-context-") {
				return errors.New("rename failed")
			}
			return os.Rename(old, new)
		}
		t.Cleanup(func() { renameInstallDir = original })
		_, err = Install("codex", Options{HomeDir: home, PackageDir: packageFixture(t, "2.0.0")})
		var failure Failure
		if !errors.As(err, &failure) || failure.Code != "promotion-failed" {
			t.Fatalf("err=%v", err)
		}
		if readText(t, filepath.Join(first.Destination, "SKILL.md")) != "# Cerne\n" {
			t.Fatal("previous install not restored")
		}
		if !failedAuditHasVersion(t, home, "2.0.0") {
			t.Fatal("failed audit did not preserve validated package version")
		}
	})
}

func TestInstallRejectsSymlinkProfileAncestor(t *testing.T) {
	home := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(home, ".codex")); err != nil {
		t.Skipf("symlink indisponível: %v", err)
	}
	_, err := Install("codex", Options{HomeDir: home, PackageDir: packageFixture(t, "1.0.0")})
	var failure Failure
	if !errors.As(err, &failure) || failure.Code != "destination-inaccessible" {
		t.Fatalf("err=%v", err)
	}
	if entries, readErr := os.ReadDir(outside); readErr != nil || len(entries) != 0 {
		t.Fatalf("destino externo foi alterado: entries=%v err=%v", entries, readErr)
	}
}

func TestInstallRejectsUnsafeOwnershipMarker(t *testing.T) {
	tests := map[string]string{
		"wrong agent":   `{"package":"cerne-skills","version":"1.0.0","agent":"claude","skill":"cerne-context","files":["SKILL.md"]}`,
		"unknown skill": `{"package":"cerne-skills","version":"1.0.0","agent":"codex","skill":"unknown","files":["SKILL.md"]}`,
		"bad version":   `{"package":"cerne-skills","version":"v1","agent":"codex","skill":"cerne-context","files":["SKILL.md"]}`,
		"empty files":   `{"package":"cerne-skills","version":"1.0.0","agent":"codex","skill":"cerne-context","files":[]}`,
		"unsafe files":  `{"package":"cerne-skills","version":"1.0.0","agent":"codex","skill":"cerne-context","files":["../x"]}`,
	}
	for name, content := range tests {
		t.Run(name, func(t *testing.T) {
			home := t.TempDir()
			target, _ := TargetPath(home, "codex")
			writeFile(t, filepath.Join(target, ".cerne-install.json"), content)
			writeFile(t, filepath.Join(target, "SKILL.md"), "# old\n")
			_, err := Install("codex", Options{HomeDir: home, PackageDir: packageFixture(t, "2.0.0")})
			var failure Failure
			if !errors.As(err, &failure) || failure.Code != "unknown-destination" || readText(t, filepath.Join(target, "SKILL.md")) != "# old\n" {
				t.Fatalf("err=%v", err)
			}
		})
	}
}

func failedAuditHasVersion(t *testing.T, home, version string) bool {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(home, ".cerne", "audit"))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		data := readText(t, filepath.Join(home, ".cerne", "audit", entry.Name()))
		if strings.Contains(data, `"status": "failed"`) && strings.Contains(data, `"package_version": "`+version+`"`) {
			return true
		}
	}
	return false
}

func readText(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
