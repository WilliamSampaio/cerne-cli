package localization

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestConfigSetLoadAndUnset(t *testing.T) {
	home := t.TempDir()
	if _, present, err := Load(home); err != nil || present {
		t.Fatalf("Load ausente = present %v, err %v", present, err)
	}
	if err := Set(home, English); err != nil {
		t.Fatal(err)
	}
	config, present, err := Load(home)
	if err != nil || !present || config.Language != English {
		t.Fatalf("Load = %#v, %v, %v", config, present, err)
	}
	data, err := os.ReadFile(Path(home))
	if err != nil || string(data) != "{\n  \"language\": \"en\"\n}\n" {
		t.Fatalf("config = %q, %v", data, err)
	}
	if err := Unset(home); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(Path(home)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("config permaneceu: %v", err)
	}
	if err := Unset(home); err != nil {
		t.Fatalf("unset ausente = %v", err)
	}
}

func TestConfigRejectsInvalidContentAndSetRepairsIt(t *testing.T) {
	home := t.TempDir()
	dir := filepath.Join(home, ".cerne")
	if err := os.Mkdir(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(Path(home), []byte("not-json\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Load(home); err == nil {
		t.Fatal("config inválida foi aceita")
	}
	if err := Set(home, PortugueseBrazil); err != nil {
		t.Fatalf("set não reparou config regular: %v", err)
	}
	config, present, err := Load(home)
	if err != nil || !present || config.Language != PortugueseBrazil {
		t.Fatalf("config reparada = %#v, %v, %v", config, present, err)
	}
}

func TestConfigRejectsUnsupportedSavedLanguage(t *testing.T) {
	home := t.TempDir()
	if err := os.Mkdir(filepath.Join(home, ".cerne"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(Path(home), []byte("{\"language\":\"es\"}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Load(home); err == nil {
		t.Fatal("idioma salvo inválido foi aceito")
	}
}

func TestConfigRefusesSymlinks(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink sem privilégio é coberto no CI Unix")
	}
	t.Run("root", func(t *testing.T) {
		home := t.TempDir()
		target := t.TempDir()
		if err := os.Symlink(target, filepath.Join(home, ".cerne")); err != nil {
			t.Fatal(err)
		}
		if err := Set(home, English); err == nil {
			t.Fatal("raiz symlink foi aceita")
		}
		if _, err := os.Stat(filepath.Join(target, "config.json")); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("alvo alterado: %v", err)
		}
	})

	t.Run("file", func(t *testing.T) {
		home := t.TempDir()
		if err := os.Mkdir(filepath.Join(home, ".cerne"), 0o700); err != nil {
			t.Fatal(err)
		}
		target := filepath.Join(t.TempDir(), "target")
		if err := os.WriteFile(target, []byte("keep"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(target, Path(home)); err != nil {
			t.Fatal(err)
		}
		if err := Set(home, English); err == nil {
			t.Fatal("arquivo symlink foi aceito")
		}
		data, _ := os.ReadFile(target)
		if string(data) != "keep" {
			t.Fatalf("alvo alterado: %q", data)
		}
	})
}

func TestFailedConfigWritePreservesExistingPreference(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permissões POSIX")
	}
	home := t.TempDir()
	if err := Set(home, English); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Dir(Path(home))
	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o700) })
	err := Set(home, PortugueseBrazil)
	if err == nil {
		t.Skip("filesystem permite escrita apesar do modo somente leitura")
	}
	data, readErr := os.ReadFile(Path(home))
	if readErr != nil || !strings.Contains(string(data), `"en"`) {
		t.Fatalf("preferência anterior perdida: %q, %v, set=%v", data, readErr, err)
	}
}
