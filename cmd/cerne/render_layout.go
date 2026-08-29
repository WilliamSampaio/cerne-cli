package main

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

const (
	ansiBold = "\x1b[1m"
	ansiDim  = "\x1b[2m"

	ruleChar       = "─"
	ruleWidth      = 72
	shortRuleWidth = 24
	columnGap      = 2
)

// styledText wraps s with attr (ansiBold or ansiDim) when w is a colorable terminal; otherwise it
// returns s unchanged.
func styledText(w io.Writer, s, attr string) string {
	if !colorEnabled(w) {
		return s
	}
	return attr + s + ansiReset
}

// sectionRule renders a titled section rule padded to ruleWidth columns, e.g. "── title ────…".
// The title is bold when w is a colorable terminal.
func sectionRule(w io.Writer, title string) string {
	prefix := ruleChar + ruleChar + " "
	suffix := " "
	filler := ruleWidth - utf8.RuneCountInString(prefix) - utf8.RuneCountInString(title) - utf8.RuneCountInString(suffix)
	if filler < 0 {
		filler = 0
	}
	return prefix + styledText(w, title, ansiBold) + suffix + strings.Repeat(ruleChar, filler)
}

// shortRule renders an untitled rule of shortRuleWidth columns.
func shortRule() string {
	return strings.Repeat(ruleChar, shortRuleWidth)
}

// columnWidth returns the display width (rune count) of the longest label, for aligning a single
// column across an entire command execution.
func columnWidth(labels []string) int {
	max := 0
	for _, label := range labels {
		if n := utf8.RuneCountInString(label); n > max {
			max = n
		}
	}
	return max
}

// alignedField renders "label" left-padded to width, columnGap spaces, then value. The label is
// dim when w is a colorable terminal.
func alignedField(w io.Writer, label string, width int, value string) string {
	pad := width - utf8.RuneCountInString(label)
	if pad < 0 {
		pad = 0
	}
	return styledText(w, label, ansiDim) + strings.Repeat(" ", pad+columnGap) + value
}

// renderVersion prints "cerne <version>", preceded by a "CERNE" banner when stdout is an
// interactive terminal. Outside a terminal, it is exactly the version line, unchanged.
func renderVersion(stdout io.Writer) {
	if layoutEnabled(stdout) {
		fmt.Fprintln(stdout, sectionRule(stdout, "CERNE"))
	}
	fmt.Fprintf(stdout, "cerne %s\n", version)
}
