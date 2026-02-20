package linter

import (
	"testing"
)

func TestHeadingLevelSkip(t *testing.T) {
	input := []byte("# Title\n\n### Skipped H2\n")
	issues := checkHeadingHierarchy(input)
	if len(issues) == 0 {
		t.Fatal("expected heading skip issue")
	}
	if issues[0].Rule != "heading-hierarchy" {
		t.Fatalf("expected rule 'heading-hierarchy', got %q", issues[0].Rule)
	}
}

func TestHeadingLevelSkip_NoIssue(t *testing.T) {
	input := []byte("# Title\n\n## Subtitle\n\n### Sub-sub\n")
	issues := checkHeadingHierarchy(input)
	if len(issues) != 0 {
		t.Fatalf("expected no issues, got %d", len(issues))
	}
}

func TestDuplicateHeadings(t *testing.T) {
	input := []byte("# Title\n\n# Title\n")
	issues := checkDuplicateHeadings(input)
	if len(issues) == 0 {
		t.Fatal("expected duplicate heading issue")
	}
}

func TestEmptyLinks(t *testing.T) {
	input := []byte("[Empty link]()\n")
	issues := checkEmptyLinks(input)
	if len(issues) == 0 {
		t.Fatal("expected empty link issue")
	}
}

func TestEmptyLinks_NoIssue(t *testing.T) {
	input := []byte("[Valid link](https://example.com)\n")
	issues := checkEmptyLinks(input)
	if len(issues) != 0 {
		t.Fatalf("expected no issues, got %d", len(issues))
	}
}

func TestDuplicateHeadings_WithFormattedText(t *testing.T) {
	input := []byte("# Title with **bold** text\n\n# Title with **bold** text\n")
	issues := checkDuplicateHeadings(input)
	if len(issues) == 0 {
		t.Fatal("expected duplicate heading issue for headings with bold text")
	}
}

func TestDuplicateHeadings_DifferentText(t *testing.T) {
	input := []byte("# Hello\n\n# World\n")
	issues := checkDuplicateHeadings(input)
	if len(issues) != 0 {
		t.Fatalf("expected no issues, got %d", len(issues))
	}
}
