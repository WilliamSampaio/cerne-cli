package filecheck

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestAccessAllowedAndReadOnly(t *testing.T) {
	dir := t.TempDir()
	sentinel := filepath.Join(dir, "sentinel")
	if err := os.WriteFile(sentinel, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}

	result := Access(dir)
	if result.Read != Allowed || result.Write != Allowed {
		t.Fatalf("Access() = %#v, esperado leitura e escrita permitidas", result)
	}
	if data, err := os.ReadFile(sentinel); err != nil || string(data) != "keep" {
		t.Fatalf("Access() alterou conteúdo: %q %v", data, err)
	}
}

func TestAccessDenied(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("chmod não modela DACL do Windows de forma portátil")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "locked")
	if err := os.WriteFile(path, []byte("keep"), 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0o600) })

	result := Access(path)
	if result.Read != Denied || result.Write != Denied {
		t.Fatalf("Access() = %#v, esperado acesso negado", result)
	}
}

func TestAccessUnknownForUncheckablePath(t *testing.T) {
	result := Access(filepath.Join(t.TempDir(), "missing"))
	if result.Read != Unknown || result.Write != Unknown {
		t.Fatalf("Access() = %#v, esperado inconclusivo", result)
	}
}
