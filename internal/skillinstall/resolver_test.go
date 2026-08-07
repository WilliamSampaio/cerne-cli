package skillinstall

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultPackageDirFindsOnlyCompanionLayout(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(exe), PackageName)); err == nil {
		t.Skip("test binary already has companion package")
	}
	if _, err := DefaultPackageDir(); err == nil {
		t.Fatal("resolver should not scan sibling checkouts")
	}
}
