package viewer

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewModel(t *testing.T) {
	m := NewModel("# Hello\n\nWorld")
	if m.content == "" {
		t.Fatal("expected content to be set")
	}
}

func TestModel_ScrollDown(t *testing.T) {
	m := NewModel("# Hello\n\nLine 1\nLine 2\nLine 3\nLine 4\nLine 5\nLine 6\nLine 7\nLine 8\nLine 9\nLine 10")
	m.height = 5

	msg := tea.KeyMsg{Type: tea.KeyDown}
	updated, _ := m.Update(msg)
	model := updated.(Model)
	if model.offset <= 0 {
		t.Fatal("expected offset to increase after scroll down")
	}
}

func TestModel_ScrollUp(t *testing.T) {
	m := NewModel("# Hello\n\nContent")
	m.height = 5
	m.offset = 3

	msg := tea.KeyMsg{Type: tea.KeyUp}
	updated, _ := m.Update(msg)
	model := updated.(Model)
	if model.offset >= 3 {
		t.Fatal("expected offset to decrease after scroll up")
	}
}

func TestModel_Quit(t *testing.T) {
	m := NewModel("# Hello")

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, cmd := m.Update(msg)
	if cmd == nil {
		t.Fatal("expected quit command")
	}
}

func TestModel_ScrollToBottom(t *testing.T) {
	// Create content with 20 lines
	lines := make([]string, 20)
	for i := range lines {
		lines[i] = "line"
	}
	content := strings.Join(lines, "\n")
	m := NewModel(content)
	m.height = 10

	// Press G to go to bottom
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})
	model := updated.(Model)

	// The maxOffset should account for the status bar
	if model.offset != model.maxOffset() {
		t.Fatalf("expected offset %d to equal maxOffset %d", model.offset, model.maxOffset())
	}

	// Verify the view includes the last line of content
	view := model.View()
	viewLines := strings.Split(view, "\n")
	// Last visible content line (before the status line) should be "line"
	if len(viewLines) < 2 {
		t.Fatal("expected at least 2 view lines")
	}
	lastContentLine := viewLines[len(viewLines)-2]
	if lastContentLine != "line" {
		t.Fatalf("expected last content line to be 'line', got %q", lastContentLine)
	}
}
