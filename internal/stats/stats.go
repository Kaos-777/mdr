package stats

import (
	"fmt"
	"io"
	"math"
	"os"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// Stats holds document statistics.
type Stats struct {
	Words      int
	Lines      int
	Headings   int
	Links      int
	Images     int
	CodeBlocks int
	ReadTime   string
}

// FromBytes computes stats from raw markdown source.
func FromBytes(source []byte) Stats {
	s := Stats{}

	// Line count
	s.Lines = strings.Count(string(source), "\n")
	if len(source) > 0 && source[len(source)-1] != '\n' {
		s.Lines++
	}

	// Word count from raw source
	s.Words = len(strings.Fields(string(source)))

	// AST-based counts
	md := goldmark.New()
	reader := text.NewReader(source)
	doc := md.Parser().Parse(reader)

	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch n.(type) {
		case *ast.Heading:
			s.Headings++
		case *ast.Link:
			s.Links++
		case *ast.Image:
			s.Images++
		case *ast.FencedCodeBlock, *ast.CodeBlock:
			s.CodeBlocks++
		}
		return ast.WalkContinue, nil
	})

	// Reading time estimate at ~200 words per minute
	minutes := float64(s.Words) / 200.0
	if minutes < 1 {
		s.ReadTime = "< 1 min"
	} else {
		s.ReadTime = fmt.Sprintf("%d min", int(math.Ceil(minutes)))
	}

	return s
}

// FromFile reads a file and computes stats.
func FromFile(path string) (Stats, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Stats{}, err
	}
	return FromBytes(data), nil
}

// FromReader reads from an io.Reader and computes stats.
func FromReader(r io.Reader) (Stats, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return Stats{}, err
	}
	return FromBytes(data), nil
}

// String returns a formatted summary.
func (s Stats) String() string {
	return fmt.Sprintf(
		"  Words:        %d\n"+
			"  Lines:        %d\n"+
			"  Headings:     %d\n"+
			"  Links:        %d\n"+
			"  Images:       %d\n"+
			"  Code blocks:  %d\n"+
			"  Reading time: %s",
		s.Words, s.Lines, s.Headings, s.Links, s.Images, s.CodeBlocks, s.ReadTime,
	)
}
