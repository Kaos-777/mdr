package linter

import (
	"io"
	"os"
	"sort"
)

// Lint runs all lint rules against the given markdown source and returns issues sorted by line number.
func Lint(source []byte) []Issue {
	var issues []Issue
	issues = append(issues, checkHeadingHierarchy(source)...)
	issues = append(issues, checkDuplicateHeadings(source)...)
	issues = append(issues, checkEmptyLinks(source)...)
	issues = append(issues, checkTrailingWhitespace(source)...)
	issues = append(issues, checkEmptySections(source)...)

	sort.Slice(issues, func(i, j int) bool {
		return issues[i].Line < issues[j].Line
	})
	return issues
}

// LintFile reads a file and lints its contents.
func LintFile(path string) ([]Issue, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Lint(data), nil
}

// LintReader reads from an io.Reader and lints the content.
func LintReader(r io.Reader) ([]Issue, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return Lint(data), nil
}
