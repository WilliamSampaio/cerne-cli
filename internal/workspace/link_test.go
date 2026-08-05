package workspace

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestLinkUpdatesManifestAfterValidations(t *testing.T) {
	root := newDoctorWorkspace(t, "example")
	start := filepath.Join(root, "knowledge", "product")
	source := filepath.Join(filepath.Dir(root), "geo app Ω")
	if err := os.Mkdir(source, 0o755); err != nil {
		t.Fatal(err)
	}
	writeManifest(t, root, `{"name":"example","source":"../source","kept":true}`)
	manifestPath := filepath.Join(root, "knowledge", "cerne.json")
	if runtime.GOOS != "windows" {
		if err := os.Chmod(manifestPath, 0o640); err != nil {
			t.Fatal(err)
		}
	}
	input, err := filepath.Rel(start, source)
	if err != nil {
		t.Fatal(err)
	}

	got, err := Link(start, LinkRequest{SourceInput: input, Replace: true}, fakeLinkInspect(nil, nil))
	if err != nil {
		t.Fatal(err)
	}
	wantSource := filepath.ToSlash(mustRel(t, filepath.Join(root, "knowledge"), source))
	if !got.Changed || got.ProjectName != "example" || got.PreviousSource != "../source" || got.NewSource != wantSource {
		t.Fatalf("resultado = %#v", got)
	}
	raw := readManifestJSON(t, root)
	if string(raw["source"]) != `"`+wantSource+`"` || string(raw["kept"]) != "true" {
		t.Fatalf("manifesto = %#v", raw)
	}
	if runtime.GOOS != "windows" {
		info, err := os.Stat(manifestPath)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != 0o640 {
			t.Fatalf("permissões do manifesto = %o, esperado 640", info.Mode().Perm())
		}
	}
}

func TestLinkPreservesOpaqueWorkflow(t *testing.T) {
	root := newDoctorWorkspace(t, "example")
	source := filepath.Join(filepath.Dir(root), "new-source")
	if err := os.Mkdir(source, 0o755); err != nil {
		t.Fatal(err)
	}
	writeManifest(t, root, `{"name":"example","source":"../source","workflow":{"provider":"future"}}`)
	if _, err := Link(root, LinkRequest{SourceInput: source, Replace: true}, fakeLinkInspect(nil, nil)); err != nil {
		t.Fatal(err)
	}
	raw := readManifestJSON(t, root)
	var workflow struct {
		Provider string `json:"provider"`
	}
	if err := json.Unmarshal(raw["workflow"], &workflow); err != nil || workflow.Provider != "future" {
		t.Fatalf("workflow=%s erro=%v", raw["workflow"], err)
	}
}

func TestLinkHandlesAbsoluteCanonicalAliasAndRejectsSubdirectory(t *testing.T) {
	root := newDoctorWorkspace(t, "example")
	source := filepath.Join(filepath.Dir(root), "source-real")
	child := filepath.Join(source, "pkg")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := Link(root, LinkRequest{SourceInput: source, Replace: true}, fakeLinkInspect(nil, nil))
	if err != nil || !got.Changed {
		t.Fatalf("absoluto = %#v, erro = %#v", got, err)
	}

	writeManifest(t, root, `{"name":"example","source":"../source"}`)
	_, err = Link(root, LinkRequest{SourceInput: child, Replace: true}, fakeLinkInspect(map[string]LinkRepositoryFacts{
		child: {WorktreeRoot: source, CommonDir: filepath.Join(source, ".git"), HasWorktree: true},
	}, nil))
	assertLinkFailure(t, err, "raiz Git própria")

	alias := filepath.Join(filepath.Dir(root), "source-alias")
	if err := os.Symlink(source, alias); err != nil {
		t.Skip(err)
	}
	got, err = Link(root, LinkRequest{SourceInput: alias, Replace: true}, fakeLinkInspect(nil, nil))
	if err != nil || !got.Changed || got.NewSource != filepath.ToSlash(mustRel(t, filepath.Join(root, "knowledge"), source)) {
		t.Fatalf("alias = %#v, erro = %#v", got, err)
	}
}

func TestLinkReplaceAndNoopRules(t *testing.T) {
	root := newDoctorWorkspace(t, "example")
	source := filepath.Join(filepath.Dir(root), "new-source")
	if err := os.Mkdir(source, 0o755); err != nil {
		t.Fatal(err)
	}

	_, err := Link(root, LinkRequest{SourceInput: source}, fakeLinkInspect(nil, nil))
	assertLinkFailure(t, err, "outro source já está configurado")

	got, err := Link(root, LinkRequest{SourceInput: source, Replace: true}, fakeLinkInspect(nil, nil))
	if err != nil || !got.Changed {
		t.Fatalf("replace = %#v, erro = %#v", got, err)
	}

	manifest := filepath.Join(root, "knowledge", "cerne.json")
	before := readFileForWorkspaceTest(t, manifest)
	got, err = Link(root, LinkRequest{SourceInput: source}, fakeLinkInspect(nil, nil))
	after := readFileForWorkspaceTest(t, manifest)
	if err != nil || got.Changed || before != after {
		t.Fatalf("no-op = %#v, erro = %#v, antes=%q depois=%q", got, err, before, after)
	}
}

func TestLinkInvalidCurrentSourceStillRequiresReplace(t *testing.T) {
	root := newDoctorWorkspace(t, "example")
	source := filepath.Join(filepath.Dir(root), "new-source")
	if err := os.Mkdir(source, 0o755); err != nil {
		t.Fatal(err)
	}
	writeManifest(t, root, `{"name":"example","source":"../missing"}`)

	_, err := Link(root, LinkRequest{SourceInput: source}, fakeLinkInspect(nil, nil))
	assertLinkFailure(t, err, "outro source já está configurado")
	if _, err := Link(root, LinkRequest{SourceInput: source, Replace: true}, fakeLinkInspect(nil, nil)); err != nil {
		t.Fatal(err)
	}
}

func TestLinkBlockingFailures(t *testing.T) {
	root := newDoctorWorkspace(t, "example")
	source := filepath.Join(filepath.Dir(root), "source2")
	if err := os.Mkdir(source, 0o755); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(filepath.Dir(root), "file.txt")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(root, "knowledge", "nested")
	if err := os.Mkdir(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	cases := map[string]struct {
		start     string
		input     string
		manifest  string
		overrides map[string]LinkRepositoryFacts
		failures  map[string]error
		cause     string
	}{
		"workspace not found": {start: t.TempDir(), input: source, cause: "workspace Cerne não localizado"},
		"malformed manifest":  {start: root, input: source, manifest: `{`, cause: "manifesto ausente ou inválido"},
		"unsupported version": {start: root, input: source, manifest: `{"name":"example","source":"../source","version":2}`, cause: "versão do manifesto não suportada"},
		"missing path":        {start: root, input: filepath.Join(filepath.Dir(root), "missing"), cause: "caminho source inexistente"},
		"file path":           {start: root, input: file, cause: "caminho source não é diretório"},
		"invalid git":         {start: root, input: source, failures: map[string]error{source: errors.New("not git")}, cause: "repositório Git válido"},
		"bare":                {start: root, input: source, overrides: map[string]LinkRepositoryFacts{source: {IsBare: true}}, cause: "bare"},
		"same repository":     {start: root, input: filepath.Join(root, "knowledge"), cause: "mesmo repositório"},
		"nested": {start: root, input: nested, overrides: map[string]LinkRepositoryFacts{
			nested: {WorktreeRoot: nested, CommonDir: filepath.Join(nested, ".git"), HasWorktree: true},
		}, cause: "sobreposição perigosa"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if tc.manifest != "" {
				writeManifest(t, root, tc.manifest)
				t.Cleanup(func() { writeManifest(t, root, `{"name":"example","source":"../source"}`) })
			}
			_, err := Link(tc.start, LinkRequest{SourceInput: tc.input, Replace: true}, fakeLinkInspect(tc.overrides, tc.failures))
			assertLinkFailure(t, err, tc.cause)
		})
	}
}

func TestLinkWriteFailurePreservesManifest(t *testing.T) {
	root := newDoctorWorkspace(t, "example")
	source := filepath.Join(filepath.Dir(root), "new-source")
	if err := os.Mkdir(source, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := filepath.Join(root, "knowledge", "cerne.json")
	before := readFileForWorkspaceTest(t, manifest)
	original := replaceManifestFile
	replaceManifestFile = func(string, string) error { return errors.New("disk full") }
	t.Cleanup(func() { replaceManifestFile = original })

	_, err := Link(root, LinkRequest{SourceInput: source, Replace: true}, fakeLinkInspect(nil, nil))
	assertLinkFailure(t, err, "manifesto não pode ser atualizado")
	if after := readFileForWorkspaceTest(t, manifest); after != before {
		t.Fatalf("manifesto mudou após falha\nantes=%q\ndepois=%q", before, after)
	}
}

func fakeLinkInspect(overrides map[string]LinkRepositoryFacts, failures map[string]error) LinkGitInspect {
	factsByPath := map[string]LinkRepositoryFacts{}
	for path, facts := range overrides {
		factsByPath[canonical(path)] = facts
	}
	errorsByPath := map[string]error{}
	for path, err := range failures {
		errorsByPath[canonical(path)] = err
	}
	return func(path string) (LinkRepositoryFacts, error) {
		path = canonical(path)
		if err, ok := errorsByPath[path]; ok {
			return LinkRepositoryFacts{}, err
		}
		if facts, ok := factsByPath[path]; ok {
			facts.RequestedPath = path
			return facts, nil
		}
		return LinkRepositoryFacts{
			RequestedPath: path,
			WorktreeRoot:  path,
			CommonDir:     filepath.Join(path, ".git"),
			HasWorktree:   true,
		}, nil
	}
}

func assertLinkFailure(t *testing.T, err error, cause string) {
	t.Helper()
	var failure LinkFailure
	if !errors.As(err, &failure) || !strings.Contains(failure.Cause, cause) || failure.Correction == "" {
		t.Fatalf("erro = %#v", err)
	}
}

func readManifestJSON(t *testing.T, root string) map[string]json.RawMessage {
	t.Helper()
	data := readFileForWorkspaceTest(t, filepath.Join(root, "knowledge", "cerne.json"))
	raw := map[string]json.RawMessage{}
	if err := json.Unmarshal([]byte(data), &raw); err != nil {
		t.Fatal(err)
	}
	return raw
}

func readFileForWorkspaceTest(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func mustRel(t *testing.T, base, target string) string {
	t.Helper()
	rel, err := filepath.Rel(base, target)
	if err != nil {
		t.Fatal(err)
	}
	return rel
}
