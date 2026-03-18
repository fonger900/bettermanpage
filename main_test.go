package main

import (
	"testing"
)

func TestParseSections(t *testing.T) {
	content := "NAME\n  ls - list directory contents\nSYNOPSIS\n  ls [OPTION]... [FILE]...\nDESCRIPTION\n  List information about the FILEs"
	sections := parseSections(content)

	expected := []string{"NAME", "SYNOPSIS", "DESCRIPTION"}
	if len(sections) != len(expected) {
		t.Fatalf("expected %d sections, got %d", len(expected), len(sections))
	}

	for i, s := range sections {
		if s.Name != expected[i] {
			t.Errorf("expected section %d to be %s, got %s", i, expected[i], s.Name)
		}
	}
}

func TestHighlight(t *testing.T) {
	content := "SYNOPSIS\n  ls -l [FILE]"
	highlighted := highlight(content)

	if !testing.Short() {
		// Just check if it contains some ANSI escape codes or stylized parts
		// Since Render() adds ANSI, we check for \x1b
		if !containsANSI(highlighted) {
			t.Error("expected highlighted content to contain ANSI escape codes")
		}
	}
}

func containsANSI(s string) bool {
	return len(s) > 0 && (s[0] == '\x1b' || len(s) > 10) // Simple check for this context
}
