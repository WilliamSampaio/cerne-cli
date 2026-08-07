package skillinstall

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadPackageValidatesManifestAdapterAndSchema(t *testing.T) {
	root := packageFixture(t, "1.2.3")
	pkg, err := LoadPackage(root, "codex")
	if err != nil {
		t.Fatal(err)
	}
	if pkg.Version != "1.2.3" || pkg.Skill != filepath.Join("skills", "cerne-context") || len(pkg.Files) == 0 {
		t.Fatalf("package = %#v", pkg)
	}

	tests := map[string]string{
		"malformed":           `{`,
		"missing skill":       `{"schemaVersion":1,"name":"cerne-skills","version":"1.0.0","skills":[]}`,
		"missing adapter":     manifestJSON("1.0.0", `"claude":{}`, "cerne.context.v1"),
		"incompatible schema": manifestJSON("1.0.0", `"codex":{},"claude":{}`, "cerne.context.v2"),
		"invalid version":     manifestJSON("v1", `"codex":{},"claude":{}`, "cerne.context.v1"),
		"path escape":         manifestWithSource("1.0.0", "../outside"),
	}
	for name, manifest := range tests {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			writeFile(t, filepath.Join(root, manifestFile), manifest)
			writeFile(t, filepath.Join(root, "skills", SkillName, "SKILL.md"), "# skill\n")
			if _, err := LoadPackage(root, "codex"); err == nil {
				t.Fatal("expected invalid package")
			}
		})
	}
}

func TestLoadPackageRejectsSymlink(t *testing.T) {
	root := packageFixture(t, "1.0.0")
	if err := os.Symlink("SKILL.md", filepath.Join(root, "skills", SkillName, "link")); err != nil {
		t.Skipf("symlink indisponível: %v", err)
	}
	if _, err := LoadPackage(root, "codex"); err == nil {
		t.Fatal("expected symlink rejection")
	}
}

func packageFixture(t *testing.T, version string) string {
	t.Helper()
	root := t.TempDir()
	writeFile(t, filepath.Join(root, manifestFile), manifestJSON(version, `"codex":{},"claude":{}`, "cerne.context.v1"))
	writeFile(t, filepath.Join(root, "skills", SkillName, "SKILL.md"), "# Cerne\n")
	writeFile(t, filepath.Join(root, "skills", SkillName, "references", "context-contract.md"), "contract\n")
	return root
}

func manifestJSON(version, adapters, contextSchema string) string {
	return manifestWithSourceAndAdapters(version, "skills/cerne-context", adapters, contextSchema)
}

func manifestWithSource(version, source string) string {
	return manifestWithSourceAndAdapters(version, source, `"codex":{},"claude":{}`, "cerne.context.v1")
}

func manifestWithSourceAndAdapters(version, source, adapters, contextSchema string) string {
	return `{
  "schemaVersion": 1,
  "name": "cerne-skills",
  "version": "` + version + `",
  "skills": [{
    "id": "cerne-context",
    "source": "` + source + `",
    "entrypoint": "SKILL.md",
    "adapters": {` + adapters + `},
    "requires": {"contextSchema": "` + contextSchema + `"}
  }]
}`
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
