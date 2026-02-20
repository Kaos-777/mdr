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

func TestModel_Quit(t *testing.T) {
	m := NewModel("# Hello")

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, cmd := m.Update(msg)
	if cmd == nil {
		t.Fatal("expected quit command")
	}
}

func TestModel_SearchMode(t *testing.T) {
	m := NewModel("Line one\nLine two\nLine three")
	m.height = 10
	m.width = 80

	// Enter search mode
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	model := updated.(Model)
	if !model.searching {
		t.Fatal("expected searching to be true")
	}

	// Type query
	for _, r := range "two" {
		updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		model = updated.(Model)
	}
	if model.searchInput != "two" {
		t.Fatalf("expected search input 'two', got %q", model.searchInput)
	}

	// Press enter to execute search
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)
	if model.searching {
		t.Fatal("searching should be false after enter")
	}
	if model.searchQuery != "two" {
		t.Fatalf("expected search query 'two', got %q", model.searchQuery)
	}
	if len(model.matchLines) != 1 {
		t.Fatalf("expected 1 match, got %d", len(model.matchLines))
	}
}

func TestModel_SearchNavigate(t *testing.T) {
	// Create content with "match" appearing on lines 0, 10, 20
	lines := make([]string, 30)
	for i := range lines {
		if i%10 == 0 {
			lines[i] = "match line"
		} else {
			lines[i] = "other line"
		}
	}
	content := strings.Join(lines, "\n")
	m := NewModel(content)
	m.height = 10
	m.width = 80

	// Enter search, type, and execute
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	model := Model{
		content:    m.content,
		lines:      m.lines,
		height:     m.height,
		width:      m.width,
		searching:  true,
		matchIndex: -1,
	}
	for _, r := range "match" {
		updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		model = updated.(Model)
	}
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	if len(model.matchLines) != 3 {
		t.Fatalf("expected 3 matches, got %d", len(model.matchLines))
	}
	if model.matchIndex != 0 {
		t.Fatalf("expected match index 0, got %d", model.matchIndex)
	}

	// Press 'n' to go to next match
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	model = updated.(Model)
	if model.matchIndex != 1 {
		t.Fatalf("expected match index 1 after 'n', got %d", model.matchIndex)
	}

	// Press 'N' to go to previous match
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'N'}})
	model = updated.(Model)
	if model.matchIndex != 0 {
		t.Fatalf("expected match index 0 after 'N', got %d", model.matchIndex)
	}

	// Press 'N' again — should wrap to last match
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'N'}})
	model = updated.(Model)
	if model.matchIndex != 2 {
		t.Fatalf("expected match index 2 after wrap, got %d", model.matchIndex)
	}
}

func TestModel_SearchEscapeClear(t *testing.T) {
	m := NewModel("Line one\nLine two\nLine three")
	m.height = 10
	m.width = 80

	// Set up a search result
	m.searchQuery = "two"
	m.matchLines = []int{1}
	m.matchIndex = 0

	// Press Escape to clear search
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	model := updated.(Model)
	if model.searchQuery != "" {
		t.Fatal("expected search query to be cleared")
	}
	if model.matchLines != nil {
		t.Fatal("expected match lines to be nil")
	}
}

func TestModel_SearchCancelInput(t *testing.T) {
	m := NewModel("Line one\nLine two")
	m.height = 10
	m.width = 80

	// Enter search mode
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	model := updated.(Model)

	// Type something
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	model = updated.(Model)

	// Cancel with Escape
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEscape})
	model = updated.(Model)
	if model.searching {
		t.Fatal("expected searching to be false after escape")
	}
	if model.searchInput != "" {
		t.Fatal("expected search input to be cleared")
	}
}

func TestModel_SearchViewStatus(t *testing.T) {
	m := NewModel("Line one\nLine two\nLine three")
	m.height = 10
	m.width = 80

	// During search, View should show search prompt
	m.searching = true
	m.searchInput = "test"
	view := m.View()
	if !strings.Contains(view, "/test") {
		t.Fatal("expected search prompt in view during search")
	}

	// With active results, View should show match count
	m.searching = false
	m.searchQuery = "Line"
	m.matchLines = []int{0, 1, 2}
	m.matchIndex = 1
	view = m.View()
	if !strings.Contains(view, "[2/3]") {
		t.Fatal("expected match count in view")
	}
}
