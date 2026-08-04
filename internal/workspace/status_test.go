package workspace

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCurrentStatusLocatesNearestWorkspaceAndReportsCleanRepositories(t *testing.T) {
	root := newDoctorWorkspace(t, "example")
	start := filepath.Join(root, "knowledge", "product")

	got, err := CurrentStatus(start, fakeGitStatus(nil, nil))
	if err != nil {
		t.Fatal(err)
	}
	if got.ProjectName != "example" || !samePath(got.Root, root) || len(got.Repositories) != 2 {
		t.Fatalf("relatório = %#v", got)
	}
	if got.Repositories[0].Name != "knowledge" || got.Repositories[1].Name != "source" {
		t.Fatalf("ordem dos repositórios = %#v", got.Repositories)
	}
	for _, repository := range got.Repositories {
		if repository.State != RepositoryClean || repository.Branch != "main" ||
			repository.Commit != "sem commits" || repository.ModifiedCount != 0 ||
			repository.StagedCount != 0 || repository.UntrackedCount != 0 {
			t.Fatalf("repositório limpo = %#v", repository)
		}
	}
}

func TestCurrentStatusClassifiesPendingDetachedHeadAndNoCommits(t *testing.T) {
	root := newDoctorWorkspace(t, "example")
	source := filepath.Join(root, "source")
	overrides := map[string]GitRepositoryStatus{
		source: {
			Branch:         "detached HEAD",
			Commit:         "sem commits",
			ModifiedCount:  2,
			StagedCount:    1,
			UntrackedCount: 3,
		},
	}

	got, err := CurrentStatus(root, fakeGitStatus(overrides, nil))
	if err != nil {
		t.Fatal(err)
	}
	sourceReport := got.Repositories[1]
	if sourceReport.State != RepositoryPending || sourceReport.Branch != "detached HEAD" ||
		sourceReport.Commit != "sem commits" || sourceReport.ModifiedCount != 2 ||
		sourceReport.StagedCount != 1 || sourceReport.UntrackedCount != 3 {
		t.Fatalf("source pendente = %#v", sourceReport)
	}
}

func TestCurrentStatusAcceptsExternalSource(t *testing.T) {
	root := newDoctorWorkspace(t, "example")
	external := filepath.Join(filepath.Dir(root), "external-source")
	if err := os.Mkdir(external, 0o755); err != nil {
		t.Fatal(err)
	}

	for _, source := range []string{
		filepath.ToSlash(mustRel(t, filepath.Join(root, "knowledge"), external)),
		filepath.ToSlash(canonical(external)),
	} {
		writeManifest(t, root, fmt.Sprintf(`{"name":"example","source":%q}`, source))
		got, err := CurrentStatus(root, fakeGitStatus(nil, nil))
		if err != nil {
			t.Fatalf("source %q: %v", source, err)
		}
		if !samePath(got.Repositories[1].Path, external) {
			t.Fatalf("source = %q, esperado %q", got.Repositories[1].Path, external)
		}
	}
}

func TestCurrentStatusFailures(t *testing.T) {
	cases := map[string]func(*testing.T) (string, GitStatus, string){
		"workspace not found": func(t *testing.T) (string, GitStatus, string) {
			return t.TempDir(), fakeGitStatus(nil, nil), "workspace Cerne não localizado"
		},
		"missing manifest in candidate workspace": func(t *testing.T) (string, GitStatus, string) {
			root := newDoctorWorkspace(t, "example")
			if err := os.Remove(filepath.Join(root, "knowledge", "cerne.json")); err != nil {
				t.Fatal(err)
			}
			return root, fakeGitStatus(nil, nil), "manifesto Cerne ausente"
		},
		"malformed manifest": func(t *testing.T) (string, GitStatus, string) {
			root := newDoctorWorkspace(t, "example")
			writeManifest(t, root, `{`)
			return root, fakeGitStatus(nil, nil), "manifesto ausente ou inválido"
		},
		"missing source": func(t *testing.T) (string, GitStatus, string) {
			root := newDoctorWorkspace(t, "example")
			writeManifest(t, root, `{"name":"example","source":"../missing"}`)
			return root, fakeGitStatus(nil, nil), "caminho source inválido"
		},
		"invalid git repository": func(t *testing.T) (string, GitStatus, string) {
			root := newDoctorWorkspace(t, "example")
			failures := map[string]error{filepath.Join(root, "source"): errors.New("not git")}
			return root, fakeGitStatus(nil, failures), "não foi possível consultar o repositório Git"
		},
	}
	for name, setup := range cases {
		t.Run(name, func(t *testing.T) {
			start, collect, cause := setup(t)
			_, err := CurrentStatus(start, collect)
			var failure StatusFailure
			if !errors.As(err, &failure) || !strings.Contains(failure.Cause, cause) || failure.Correction == "" {
				t.Fatalf("erro = %#v", err)
			}
			if name != "workspace not found" && failure.Path == "" {
				t.Fatalf("erro sem caminho afetado: %#v", failure)
			}
		})
	}
}

func TestCurrentStatusRejectsSourceSymlink(t *testing.T) {
	root := newDoctorWorkspace(t, "example")
	if err := os.RemoveAll(filepath.Join(root, "source")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(t.TempDir(), filepath.Join(root, "source")); err != nil {
		t.Skip(err)
	}
	_, err := CurrentStatus(root, fakeGitStatus(nil, nil))
	var failure StatusFailure
	if !errors.As(err, &failure) || !strings.Contains(failure.Cause, "caminho source inválido") {
		t.Fatalf("erro = %#v", err)
	}
}

func fakeGitStatus(overrides map[string]GitRepositoryStatus, failures map[string]error) GitStatus {
	statuses := map[string]GitRepositoryStatus{}
	for path, status := range overrides {
		statuses[canonical(path)] = status
	}
	errorsByPath := map[string]error{}
	for path, err := range failures {
		errorsByPath[canonical(path)] = err
	}
	return func(path string) (GitRepositoryStatus, error) {
		path = canonical(path)
		if err, ok := errorsByPath[path]; ok {
			return GitRepositoryStatus{}, err
		}
		if status, ok := statuses[path]; ok {
			status.Path = path
			return status, nil
		}
		return GitRepositoryStatus{Path: path, Branch: "main", Commit: "sem commits"}, nil
	}
}
