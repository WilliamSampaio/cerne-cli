package workspace

import "testing"

func TestPortableRestoreSourceRejectsUnsafeCrossPlatformPaths(t *testing.T) {
	valid := []string{"../source", "../repositories/source"}
	for _, value := range valid {
		if _, err := portableRestoreSource(value); err != nil {
			t.Fatalf("%q deveria ser aceito: %v", value, err)
		}
	}
	invalid := []string{"", "source", "/source", `C:\\source`, `..\\source`, "../../source", "../source/..", "../.git", "../source//nested", "../source?query"}
	for _, value := range invalid {
		if _, err := portableRestoreSource(value); err == nil {
			t.Fatalf("%q deveria ser recusado", value)
		}
	}
}
