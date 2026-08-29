package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/WilliamSampaio/cerne-cli/internal/workspace"
)

func withTerminal(t *testing.T, terminal bool) *os.File {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "render-severity")
	if err != nil {
		t.Fatalf("CreateTemp() error = %v", err)
	}
	t.Cleanup(func() { f.Close() })

	original := isTerminal
	isTerminal = func(*os.File) bool { return terminal }
	t.Cleanup(func() { isTerminal = original })
	return f
}

func TestColorEnabled(t *testing.T) {
	t.Run("TTY without NO_COLOR enables color", func(t *testing.T) {
		t.Setenv("NO_COLOR", "")
		f := withTerminal(t, true)
		if !colorEnabled(f) {
			t.Errorf("colorEnabled() = false, want true")
		}
	})

	t.Run("non-TTY disables color", func(t *testing.T) {
		t.Setenv("NO_COLOR", "")
		f := withTerminal(t, false)
		if colorEnabled(f) {
			t.Errorf("colorEnabled() = true, want false")
		}
	})

	t.Run("NO_COLOR disables color even on TTY", func(t *testing.T) {
		t.Setenv("NO_COLOR", "1")
		f := withTerminal(t, true)
		if colorEnabled(f) {
			t.Errorf("colorEnabled() = true, want false")
		}
	})

	t.Run("non *os.File writer disables color", func(t *testing.T) {
		t.Setenv("NO_COLOR", "")
		var buf bytes.Buffer
		if colorEnabled(&buf) {
			t.Errorf("colorEnabled() = true, want false")
		}
	})
}

func TestSeverityColor(t *testing.T) {
	cases := []struct {
		severity workspace.Severity
		want     string
	}{
		{workspace.Pass, ansiGreen},
		{workspace.Warning, ansiYellow},
		{workspace.Error, ansiRed},
	}
	for _, c := range cases {
		if got := severityColor(c.severity); got != c.want {
			t.Errorf("severityColor(%v) = %q, want %q", c.severity, got, c.want)
		}
	}
}

func TestRepositorySeverity(t *testing.T) {
	if got := repositorySeverity(workspace.RepositoryClean); got != workspace.Pass {
		t.Errorf("repositorySeverity(clean) = %v, want Pass", got)
	}
	if got := repositorySeverity(workspace.RepositoryPending); got != workspace.Warning {
		t.Errorf("repositorySeverity(pending) = %v, want Warning", got)
	}
}

func TestColorizeIcon(t *testing.T) {
	t.Run("color enabled wraps icon", func(t *testing.T) {
		t.Setenv("NO_COLOR", "")
		f := withTerminal(t, true)
		got := colorizeIcon(f, "✓", ansiGreen)
		want := ansiGreen + "✓" + ansiReset
		if got != want {
			t.Errorf("colorizeIcon() = %q, want %q", got, want)
		}
	})

	t.Run("color disabled returns icon unchanged", func(t *testing.T) {
		t.Setenv("NO_COLOR", "1")
		f := withTerminal(t, true)
		got := colorizeIcon(f, "✓", ansiGreen)
		if got != "✓" {
			t.Errorf("colorizeIcon() = %q, want %q", got, "✓")
		}
	})
}
