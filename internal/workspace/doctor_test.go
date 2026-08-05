package workspace

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestDoctorHealthyLegacyAndExplicitVersion(t *testing.T) {
	for _, version := range []string{"", `,"version":1`} {
		t.Run(version, func(t *testing.T) {
			root := newDoctorWorkspace(t, "example")
			if version != "" {
				writeManifest(t, root, `{"name":"example","source":"../source"`+version+`}`)
			}

			got := Doctor(root, fakeInspect(nil), allowAccess)
			if got.Status != Healthy || len(got.Checks) != 10 {
				t.Fatalf("diagnóstico = %#v", got)
			}
			assertCheck(t, got, "manifest", Pass)
			assertCheck(t, got, "manifest-version", Pass)
			if version == "" && byID(got, "manifest-version").Detail != "versão 1 implícita e suportada" {
				t.Fatalf("versão implícita = %#v", byID(got, "manifest-version"))
			}
		})
	}

	t.Run("external source", func(t *testing.T) {
		root := newDoctorWorkspace(t, "example")
		external := filepath.Join(filepath.Dir(root), "external-source")
		if err := os.Mkdir(external, 0o755); err != nil {
			t.Fatal(err)
		}
		source := filepath.ToSlash(mustRel(t, filepath.Join(root, "knowledge"), external))
		writeManifest(t, root, fmt.Sprintf(`{"name":"example","source":%q}`, source))
		got := Doctor(root, fakeInspect(nil), allowAccess)
		if got.Status != Healthy {
			t.Fatalf("diagnóstico = %#v", got)
		}
	})
}

func TestDoctorBlockingFailures(t *testing.T) {
	cases := map[string]func(*testing.T, string) (GitInspect, AccessCheck){
		"missing manifest": func(t *testing.T, root string) (GitInspect, AccessCheck) {
			if err := os.Remove(filepath.Join(root, "knowledge", "cerne.json")); err != nil {
				t.Fatal(err)
			}
			return fakeInspect(nil), allowAccess
		},
		"malformed manifest": func(t *testing.T, root string) (GitInspect, AccessCheck) {
			writeManifest(t, root, `{`)
			return fakeInspect(nil), allowAccess
		},
		"extra manifest content": func(t *testing.T, root string) (GitInspect, AccessCheck) {
			writeManifest(t, root, `{"name":"example","source":"../source"} {"extra":true}`)
			return fakeInspect(nil), allowAccess
		},
		"invalid name": func(t *testing.T, root string) (GitInspect, AccessCheck) {
			writeManifest(t, root, `{"name":"bad/name","source":"../source"}`)
			return fakeInspect(nil), allowAccess
		},
		"source symlink": func(t *testing.T, root string) (GitInspect, AccessCheck) {
			if err := os.RemoveAll(filepath.Join(root, "source")); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(t.TempDir(), filepath.Join(root, "source")); err != nil {
				t.Skip(err)
			}
			return fakeInspect(nil), allowAccess
		},
		"missing required directory": func(t *testing.T, root string) (GitInspect, AccessCheck) {
			if err := os.RemoveAll(filepath.Join(root, "knowledge", "product")); err != nil {
				t.Fatal(err)
			}
			return fakeInspect(nil), allowAccess
		},
		"ancestor git": func(t *testing.T, root string) (GitInspect, AccessCheck) {
			return fakeInspect(map[string]RepositoryFacts{
				filepath.Join(root, "knowledge"): {WorktreeRoot: root, CommonDir: filepath.Join(root, ".git")},
			}), allowAccess
		},
		"shared common dir": func(t *testing.T, root string) (GitInspect, AccessCheck) {
			common := filepath.Join(root, "common.git")
			return fakeInspect(map[string]RepositoryFacts{
				filepath.Join(root, "knowledge"): {WorktreeRoot: filepath.Join(root, "knowledge"), CommonDir: common},
				filepath.Join(root, "source"):    {WorktreeRoot: filepath.Join(root, "source"), CommonDir: common},
			}), allowAccess
		},
		"access denied": func(t *testing.T, root string) (GitInspect, AccessCheck) {
			return fakeInspect(nil), func(string) AccessResult {
				return AccessResult{Read: AccessAllowed, Write: AccessDenied}
			}
		},
	}

	for name, setup := range cases {
		t.Run(name, func(t *testing.T) {
			root := newDoctorWorkspace(t, "example")
			inspect, access := setup(t, root)
			got := Doctor(root, inspect, access)
			if got.Status != Invalid || len(got.Checks) != 10 {
				t.Fatalf("diagnóstico = %#v", got)
			}
			if !hasSeverity(got, Error) {
				t.Fatalf("nenhum erro em %#v", got)
			}
		})
	}
}

func TestSamePathUsesFilesystemIdentity(t *testing.T) {
	parent := t.TempDir()
	upper := filepath.Join(parent, "Source")
	lower := filepath.Join(parent, "source")
	if err := os.Mkdir(upper, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(lower, 0o755); err != nil {
		t.Skip("sistema de arquivos não diferencia maiúsculas de minúsculas")
	}
	if samePath(upper, lower) {
		t.Fatal("diretórios distintos foram considerados iguais")
	}
}

func TestDoctorRejectsInvalidManifestVersions(t *testing.T) {
	for _, version := range []string{`"1"`, `1.0`, `null`, `2`} {
		t.Run(version, func(t *testing.T) {
			root := newDoctorWorkspace(t, "example")
			writeManifest(t, root, `{"name":"example","source":"../source","version":`+version+`}`)
			got := Doctor(root, fakeInspect(nil), allowAccess)
			assertCheck(t, got, "manifest-version", Error)
			if got.Status != Invalid {
				t.Fatalf("status = %s", got.Status)
			}
		})
	}
}

func TestDoctorWarningsAndPrecedence(t *testing.T) {
	t.Run("name differs", func(t *testing.T) {
		root := newDoctorWorkspace(t, "example")
		writeManifest(t, root, `{"name":"other","source":"../source"}`)
		got := Doctor(root, fakeInspect(nil), allowAccess)
		assertCheck(t, got, "manifest", Warning)
		if got.Status != Warnings {
			t.Fatalf("status = %s", got.Status)
		}
	})

	t.Run("access unknown", func(t *testing.T) {
		root := newDoctorWorkspace(t, "example")
		got := Doctor(root, fakeInspect(nil), func(string) AccessResult {
			return AccessResult{Read: AccessAllowed, Write: AccessUnknown}
		})
		assertCheck(t, got, "permissions", Warning)
		if got.Status != Warnings {
			t.Fatalf("status = %s", got.Status)
		}
	})

	t.Run("error beats warning", func(t *testing.T) {
		root := newDoctorWorkspace(t, "example")
		writeManifest(t, root, `{"name":"other","source":"../missing"}`)
		got := Doctor(root, fakeInspect(nil), allowAccess)
		assertCheck(t, got, "manifest", Warning)
		assertCheck(t, got, "manifest-paths", Error)
		if got.Status != Invalid || len(got.Checks) != 10 {
			t.Fatalf("diagnóstico = %#v", got)
		}
	})
}

func TestDoctorWorkflowStates(t *testing.T) {
	root := newDoctorWorkspace(t, "example")
	writeManifest(t, root, `{"name":"example","source":"../source","workflow":{"provider":"alpha"}}`)
	definition := readyDefinition("alpha", "specs", ".alpha", ".alpha/options.json")
	resolve := func(string) (WorkflowDefinition, error) { return definition, nil }

	definition.Available = false
	got := DoctorWithWorkflow(root, fakeInspect(nil), allowAccess, resolve)
	assertCheck(t, got, "workflow", Warning)
	if got.Status != Warnings || len(got.Checks) != 11 {
		t.Fatalf("diagnóstico=%#v", got)
	}

	definition.Available = true
	if err := definition.Setup(filepath.Join(root, "knowledge")); err != nil {
		t.Fatal(err)
	}
	got = DoctorWithWorkflow(root, fakeInspect(nil), allowAccess, resolve)
	assertCheck(t, got, "workflow", Pass)
	if got.Status != Healthy {
		t.Fatalf("status=%s", got.Status)
	}

	definition.Available = false
	got = DoctorWithWorkflow(root, fakeInspect(nil), allowAccess, resolve)
	assertCheck(t, got, "workflow", Warning)
}

func TestDoctorAcceptsWorkflowWithoutTopLevelSpecs(t *testing.T) {
	definition := readyDefinition("beta", "beta/specs", "beta", "beta/config.yaml")
	base, _, err := InitWithWorkflow(t.TempDir(), "example", definition, fakeInitRepository)
	if err != nil {
		t.Fatal(err)
	}
	accessExisting := func(path string) AccessResult {
		if _, err := os.Stat(path); err != nil {
			return AccessResult{Read: AccessDenied, Write: AccessDenied}
		}
		return allowAccess(path)
	}
	got := DoctorWithWorkflow(filepath.Dir(base.KnowledgePath), fakeInspect(nil), accessExisting, func(string) (WorkflowDefinition, error) { return definition, nil })
	if got.Status != Healthy {
		t.Fatalf("diagnóstico=%#v", got)
	}
}

func TestDoctorRejectsMalformedUnknownPartialAndNestedGitWorkflow(t *testing.T) {
	for _, test := range []struct {
		name, workflow string
		prepare        func(*testing.T, string)
		resolve        WorkflowResolver
	}{
		{"malformed", `{"provider":""}`, nil, func(string) (WorkflowDefinition, error) { return WorkflowDefinition{}, errors.New("unknown") }},
		{"unknown", `{"provider":"other"}`, nil, func(string) (WorkflowDefinition, error) { return WorkflowDefinition{}, errors.New("unknown") }},
		{"partial", `{"provider":"alpha"}`, func(t *testing.T, knowledge string) {
			if err := os.Mkdir(filepath.Join(knowledge, ".alpha"), 0o755); err != nil {
				t.Fatal(err)
			}
		}, func(string) (WorkflowDefinition, error) {
			return readyDefinition("alpha", "specs", ".alpha", ".alpha/options.json"), nil
		}},
		{"nested git", `{"provider":"alpha"}`, func(t *testing.T, knowledge string) {
			definition := readyDefinition("alpha", "specs", ".alpha", ".alpha/options.json")
			if err := definition.Setup(knowledge); err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(filepath.Join(knowledge, ".alpha", "nested", ".git"), 0o755); err != nil {
				t.Fatal(err)
			}
		}, func(string) (WorkflowDefinition, error) {
			return readyDefinition("alpha", "specs", ".alpha", ".alpha/options.json"), nil
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := newDoctorWorkspace(t, "example")
			writeManifest(t, root, `{"name":"example","source":"../source","workflow":`+test.workflow+`}`)
			if test.prepare != nil {
				test.prepare(t, filepath.Join(root, "knowledge"))
			}
			got := DoctorWithWorkflow(root, fakeInspect(nil), allowAccess, test.resolve)
			assertCheck(t, got, "workflow", Error)
			if got.Status != Invalid {
				t.Fatalf("status=%s", got.Status)
			}
		})
	}
}

func newDoctorWorkspace(t *testing.T, name string) string {
	t.Helper()
	parent := t.TempDir()
	result, err := Init(parent, name, func(path string) error {
		return os.Mkdir(filepath.Join(path, ".git"), 0o755)
	})
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Dir(result.KnowledgePath)
}

func writeManifest(t *testing.T, root, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, "knowledge", "cerne.json"), []byte(content+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func fakeInspect(overrides map[string]RepositoryFacts) GitInspect {
	canonicalOverrides := map[string]RepositoryFacts{}
	for path, facts := range overrides {
		canonicalOverrides[canonical(path)] = facts
	}
	return func(path string) (RepositoryFacts, error) {
		path = canonical(path)
		if facts, ok := canonicalOverrides[path]; ok {
			facts.RequestedRoot = path
			return facts, nil
		}
		return RepositoryFacts{
			RequestedRoot: path,
			WorktreeRoot:  path,
			CommonDir:     filepath.Join(path, ".git"),
		}, nil
	}
}

func allowAccess(string) AccessResult {
	return AccessResult{Read: AccessAllowed, Write: AccessAllowed}
}

func byID(diagnosis Diagnosis, id string) CheckResult {
	for _, check := range diagnosis.Checks {
		if check.ID == id {
			return check
		}
	}
	return CheckResult{}
}

func assertCheck(t *testing.T, diagnosis Diagnosis, id string, severity Severity) {
	t.Helper()
	got := byID(diagnosis, id)
	if got.ID == "" || got.Severity != severity {
		t.Fatalf("%s = %#v, severidade esperada %v", id, got, severity)
	}
}

func hasSeverity(diagnosis Diagnosis, severity Severity) bool {
	for _, check := range diagnosis.Checks {
		if check.Severity == severity {
			return true
		}
	}
	return false
}
