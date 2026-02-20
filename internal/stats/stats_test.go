package stats

import (
	"strings"
	"testing"
)

func TestFromBytes_Basic(t *testing.T) {
	input := []byte("# Hello World\n\nThis is a paragraph with some words.\n\n## Section\n\n[Link](https://example.com)\n")
	s := FromBytes(input)

	if s.Headings != 2 {
		t.Fatalf("expected 2 headings, got %d", s.Headings)
	}
	if s.Links != 1 {
		t.Fatalf("expected 1 link, got %d", s.Links)
	}
	if s.Lines < 5 {
		t.Fatalf("expected at least 5 lines, got %d", s.Lines)
	}
	if s.Words == 0 {
		t.Fatal("expected non-zero word count")
	}
	if s.ReadTime == "" {
		t.Fatal("expected reading time to be set")
	}
}

func TestFromBytes_Empty(t *testing.T) {
	s := FromBytes([]byte(""))
	if s.Words != 0 {
		t.Fatalf("expected 0 words, got %d", s.Words)
	}
	if s.Lines != 0 {
		t.Fatalf("expected 0 lines, got %d", s.Lines)
	}
}

func TestFromBytes_CodeBlocks(t *testing.T) {
	input := []byte("# Title\n\n```go\nfmt.Println(\"hello\")\n```\n\n```\nplain block\n```\n")
	s := FromBytes(input)
	if s.CodeBlocks != 2 {
		t.Fatalf("expected 2 code blocks, got %d", s.CodeBlocks)
	}
}

func TestFromBytes_Images(t *testing.T) {
	input := []byte("# Gallery\n\n![Alt text](image.png)\n\n![Another](pic.jpg)\n")
	s := FromBytes(input)
	if s.Images != 2 {
		t.Fatalf("expected 2 images, got %d", s.Images)
	}
}

func TestFromFile(t *testing.T) {
	s, err := FromFile("../../testdata/complex.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Headings == 0 {
		t.Fatal("expected headings in complex.md")
	}
}

func TestFromFile_NotFound(t *testing.T) {
	_, err := FromFile("nonexistent.md")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestFromReader(t *testing.T) {
	r := strings.NewReader("# Title\n\nSome words here.\n")
	s, err := FromReader(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Headings != 1 {
		t.Fatalf("expected 1 heading, got %d", s.Headings)
	}
}

func TestStats_String(t *testing.T) {
	s := FromBytes([]byte("# Hello\n\nWorld\n"))
	output := s.String()
	if !strings.Contains(output, "Words:") {
		t.Fatal("expected 'Words:' in string output")
	}
	if !strings.Contains(output, "Reading time:") {
		t.Fatal("expected 'Reading time:' in string output")
	}
}

func TestFromBytes_ReadingTime(t *testing.T) {
	// 200 words should be ~1 min
	words := make([]string, 200)
	for i := range words {
		words[i] = "word"
	}
	input := []byte(strings.Join(words, " "))
	s := FromBytes(input)
	if s.ReadTime != "1 min" {
		t.Fatalf("expected '1 min' for 200 words, got %q", s.ReadTime)
	}
}
