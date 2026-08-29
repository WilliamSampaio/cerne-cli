package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestStyledText(t *testing.T) {
	t.Run("color enabled wraps text", func(t *testing.T) {
		t.Setenv("NO_COLOR", "")
		f := withTerminal(t, true)
		got := styledText(f, "title", ansiBold)
		want := ansiBold + "title" + ansiReset
		if got != want {
			t.Errorf("styledText() = %q, want %q", got, want)
		}
	})

	t.Run("color disabled returns text unchanged", func(t *testing.T) {
		var buf bytes.Buffer
		got := styledText(&buf, "title", ansiBold)
		if got != "title" {
			t.Errorf("styledText() = %q, want %q", got, "title")
		}
	})
}

func TestSectionRule(t *testing.T) {
	var buf bytes.Buffer // non-TTY: no color attribute applied
	got := sectionRule(&buf, "doctor")
	if !strings.HasPrefix(got, "── doctor ") {
		t.Fatalf("sectionRule() = %q, want prefix %q", got, "── doctor ")
	}
	if width := utf8.RuneCountInString(got); width != ruleWidth {
		t.Errorf("sectionRule() width = %d, want %d", width, ruleWidth)
	}
	if !strings.HasSuffix(got, ruleChar) {
		t.Errorf("sectionRule() = %q, want suffix %q", got, ruleChar)
	}
}

func TestSectionRuleLongTitleNeverGoesNegativeWidth(t *testing.T) {
	var buf bytes.Buffer
	longTitle := strings.Repeat("x", ruleWidth*2)
	got := sectionRule(&buf, longTitle)
	if !strings.Contains(got, longTitle) {
		t.Errorf("sectionRule() with long title should still contain the title verbatim")
	}
}

func TestShortRule(t *testing.T) {
	got := shortRule()
	if width := utf8.RuneCountInString(got); width != shortRuleWidth {
		t.Errorf("shortRule() width = %d, want %d", width, shortRuleWidth)
	}
	if strings.Trim(got, ruleChar) != "" {
		t.Errorf("shortRule() = %q, want only %q characters", got, ruleChar)
	}
}

func TestColumnWidth(t *testing.T) {
	labels := []string{"Branch", "Não rastreados", "Commit"}
	if got := columnWidth(labels); got != utf8.RuneCountInString("Não rastreados") {
		t.Errorf("columnWidth() = %d, want %d", got, utf8.RuneCountInString("Não rastreados"))
	}
	if got := columnWidth(nil); got != 0 {
		t.Errorf("columnWidth(nil) = %d, want 0", got)
	}
}

func TestRenderVersion(t *testing.T) {
	t.Run("non-TTY prints only the version line", func(t *testing.T) {
		var buf bytes.Buffer
		renderVersion(&buf)
		want := "cerne " + version + "\n"
		if buf.String() != want {
			t.Errorf("renderVersion() = %q, want %q", buf.String(), want)
		}
	})

	t.Run("TTY shows banner before the version line", func(t *testing.T) {
		t.Setenv("NO_COLOR", "")
		f := withTerminal(t, true)
		renderVersion(f)

		content, err := os.ReadFile(f.Name())
		if err != nil {
			t.Fatalf("ReadFile() error = %v", err)
		}
		output := string(content)
		if !strings.Contains(output, "── "+ansiBold+"CERNE"+ansiReset) {
			t.Errorf("output missing bold CERNE banner:\n%s", output)
		}
		if !strings.HasSuffix(output, "cerne "+version+"\n") {
			t.Errorf("output missing version line at the end:\n%s", output)
		}
	})

	t.Run("TTY with NO_COLOR shows banner without ANSI", func(t *testing.T) {
		t.Setenv("NO_COLOR", "1")
		f := withTerminal(t, true)
		renderVersion(f)

		content, err := os.ReadFile(f.Name())
		if err != nil {
			t.Fatalf("ReadFile() error = %v", err)
		}
		output := string(content)
		if strings.ContainsRune(output, '\x1b') {
			t.Errorf("output contains ANSI sequence with NO_COLOR set:\n%s", output)
		}
		if !strings.Contains(output, "── CERNE ") {
			t.Errorf("output missing CERNE banner with NO_COLOR set:\n%s", output)
		}
	})
}

func TestAlignedField(t *testing.T) {
	var buf bytes.Buffer
	width := columnWidth([]string{"Branch", "Commit"})
	got := alignedField(&buf, "Branch", width, "main")
	want := "Branch" + strings.Repeat(" ", (width-utf8.RuneCountInString("Branch"))+columnGap) + "main"
	if got != want {
		t.Errorf("alignedField() = %q, want %q", got, want)
	}
}
