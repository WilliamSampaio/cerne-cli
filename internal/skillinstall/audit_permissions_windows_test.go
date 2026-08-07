//go:build windows

package skillinstall

import (
	"testing"

	"golang.org/x/sys/windows"
)

func TestSkillInstallAuditWindowsUsesProtectedTwoPrincipalDACL(t *testing.T) {
	result, err := Install("codex", Options{HomeDir: t.TempDir(), PackageDir: packageFixture(t, "1.0.0")})
	if err != nil {
		t.Fatal(err)
	}
	descriptor, err := windows.GetNamedSecurityInfo(result.AuditPath, windows.SE_FILE_OBJECT,
		windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION)
	if err != nil {
		t.Fatal(err)
	}
	control, _, err := descriptor.Control()
	if err != nil || control&windows.SE_DACL_PROTECTED == 0 {
		t.Fatalf("DACL não protegida: control=%x err=%v", control, err)
	}
	acl, _, err := descriptor.DACL()
	if err != nil || acl.AceCount != 2 {
		t.Fatalf("DACL deve conter usuário atual e SYSTEM: count=%d err=%v", acl.AceCount, err)
	}
}
