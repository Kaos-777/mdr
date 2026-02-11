package editor

import (
	"strings"
	"unicode/utf8"
)

type Buffer struct {
	lines    []string
	modified bool
}

func NewBuffer(content string) *Buffer {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		lines = []string{""}
	}
	return &Buffer{lines: lines}
}

func (b *Buffer) LineCount() int {
	return len(b.lines)
}

func (b *Buffer) GetLine(row int) string {
	if row < 0 || row >= len(b.lines) {
		return ""
	}
	return b.lines[row]
}

// LineLen returns the rune count of the given line.
func (b *Buffer) LineLen(row int) int {
	if row < 0 || row >= len(b.lines) {
		return 0
	}
	return utf8.RuneCountInString(b.lines[row])
}

func (b *Buffer) InsertChar(row, col int, ch rune) {
	if row < 0 || row >= len(b.lines) {
		return
	}
	runes := []rune(b.lines[row])
	if col > len(runes) {
		col = len(runes)
	}
	runes = append(runes[:col], append([]rune{ch}, runes[col:]...)...)
	b.lines[row] = string(runes)
	b.modified = true
}

func (b *Buffer) DeleteChar(row, col int) {
	if row < 0 || row >= len(b.lines) {
		return
	}
	runes := []rune(b.lines[row])
	if col < 0 || col >= len(runes) {
		return
	}
	runes = append(runes[:col], runes[col+1:]...)
	b.lines[row] = string(runes)
	b.modified = true
}

func (b *Buffer) Backspace(row, col int) {
	if col > 0 {
		b.DeleteChar(row, col-1)
	}
}

func (b *Buffer) InsertNewline(row, col int) {
	if row < 0 || row >= len(b.lines) {
		return
	}
	runes := []rune(b.lines[row])
	if col > len(runes) {
		col = len(runes)
	}
	before := string(runes[:col])
	after := string(runes[col:])

	newLines := make([]string, 0, len(b.lines)+1)
	newLines = append(newLines, b.lines[:row]...)
	newLines = append(newLines, before, after)
	newLines = append(newLines, b.lines[row+1:]...)
	b.lines = newLines
	b.modified = true
}

// JoinLines joins line `row` with the line above it. Returns the rune-based column where the cursor should be.
func (b *Buffer) JoinLines(row int) int {
	if row <= 0 || row >= len(b.lines) {
		return 0
	}
	col := utf8.RuneCountInString(b.lines[row-1])
	b.lines[row-1] += b.lines[row]

	newLines := make([]string, 0, len(b.lines)-1)
	newLines = append(newLines, b.lines[:row]...)
	newLines = append(newLines, b.lines[row+1:]...)
	b.lines = newLines
	b.modified = true
	return col
}

func (b *Buffer) String() string {
	return strings.Join(b.lines, "\n")
}

func (b *Buffer) Modified() bool {
	return b.modified
}

func (b *Buffer) ResetModified() {
	b.modified = false
}
