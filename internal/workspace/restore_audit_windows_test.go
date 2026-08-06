//go:build windows

package workspace

import (
	"testing"

	"golang.org/x/sys/windows"
)

func TestRestoreAuditWindowsUsesProtectedTwoPrincipalDACL(t *testing.T) {
	audit, _, err := startRestoreAudit(t.TempDir(), SourceClone)
	if err != nil {
		t.Fatal(err)
	}
	defer audit.root.Close()
	descriptor, err := windows.GetNamedSecurityInfo(audit.path, windows.SE_FILE_OBJECT,
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
