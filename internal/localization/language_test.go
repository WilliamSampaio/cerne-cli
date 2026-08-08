package localization

import "testing"

func TestParseLanguage(t *testing.T) {
	for _, value := range []string{"en", "pt-BR"} {
		language, err := Parse(value)
		if err != nil || string(language) != value {
			t.Fatalf("Parse(%q) = %q, %v", value, language, err)
		}
	}
	if _, err := Parse("pt-br"); err == nil {
		t.Fatal("idioma não canônico foi aceito")
	}
}

func TestResolveLanguagePrecedence(t *testing.T) {
	home := t.TempDir()
	if err := Set(home, PortugueseBrazil); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		flag   string
		env    string
		envSet bool
		want   Language
	}{
		{name: "flag", flag: "en", env: "pt-BR", envSet: true, want: English},
		{name: "environment", env: "en", envSet: true, want: English},
		{name: "saved", want: PortugueseBrazil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := Resolve(test.flag, test.env, test.envSet, home)
			if err != nil || got != test.want {
				t.Fatalf("Resolve() = %q, %v; want %q", got, err, test.want)
			}
		})
	}

	if err := Unset(home); err != nil {
		t.Fatal(err)
	}
	got, err := Resolve("", "", false, home)
	if err != nil || got != Default {
		t.Fatalf("default = %q, %v", got, err)
	}
}

func TestResolveRejectsInvalidPresentValue(t *testing.T) {
	home := t.TempDir()
	if _, err := Resolve("es", "", false, home); err == nil {
		t.Fatal("flag inválida foi ignorada")
	}
	if _, err := Resolve("", "", true, home); err == nil {
		t.Fatal("ambiente vazio foi ignorado")
	}
}

func TestResolveUsesDefaultWhenHomeIsUnavailable(t *testing.T) {
	got, err := Resolve("", "", false, "")
	if err != nil || got != Default {
		t.Fatalf("Resolve sem home = %q, %v", got, err)
	}
}
