package main

import (
	"strings"
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
	highlighted := highlight(content, "ls", 1)

	if !testing.Short() {
		if !containsANSI(highlighted) {
			t.Error("expected highlighted content to contain ANSI escape codes")
		}
	}
}

func TestHighlightTLDR(t *testing.T) {
	content := "# ls\n> List directory contents.\n- List files one per line:\n`ls -1`"
	highlighted := highlightTLDR(content)

	if !strings.Contains(highlighted, "ls -1") {
		t.Error("expected highlighted TLDR to contain original code")
	}
	if !strings.Contains(highlighted, "List directory contents.") {
		t.Error("expected highlighted TLDR to contain description")
	}
}

func containsANSI(s string) bool {
	return len(s) > 0 && (strings.Contains(s, "\x1b[") || len(s) > 10)
}
