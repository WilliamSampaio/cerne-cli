package main

import (
	"io"
	"os"

	"github.com/WilliamSampaio/cerne-cli/internal/workspace"
)

const (
	ansiReset  = "\x1b[0m"
	ansiGreen  = "\x1b[32m"
	ansiYellow = "\x1b[33m"
	ansiRed    = "\x1b[31m"
)

// isTerminal reports whether f is an interactive terminal. Overridable in tests.
var isTerminal = func(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

// layoutEnabled reports whether w is an interactive terminal, regardless of NO_COLOR. It governs
// layout-only presentation (section rules, column alignment, banner) that is not itself an ANSI
// color/attribute.
func layoutEnabled(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return isTerminal(f)
}

// colorEnabled reports whether ANSI color/attributes should be written to w: only when w is a
// terminal and NO_COLOR (https://no-color.org) is not set.
func colorEnabled(w io.Writer) bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	return layoutEnabled(w)
}

func severityColor(severity workspace.Severity) string {
	switch severity {
	case workspace.Warning:
		return ansiYellow
	case workspace.Error:
		return ansiRed
	default:
		return ansiGreen
	}
}

func repositorySeverity(state string) workspace.Severity {
	if state == workspace.RepositoryPending {
		return workspace.Warning
	}
	return workspace.Pass
}

// colorizeIcon wraps icon with color's ANSI code when w is a colorable terminal; otherwise it
// returns icon unchanged.
func colorizeIcon(w io.Writer, icon, color string) string {
	if !colorEnabled(w) {
		return icon
	}
	return color + icon + ansiReset
}
