package viewer

import (
	"fmt"
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ansiRe matches ANSI escape sequences for stripping before search comparison.
var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)

type Model struct {
	content string
	lines   []string
	offset  int
	height  int
	width   int

	// Search state
	searching   bool
	searchInput string
	searchQuery string
	matchLines  []int
	matchIndex  int
}

func NewModel(rendered string) Model {
	lines := strings.Split(rendered, "\n")
	return Model{
		content:    rendered,
		lines:      lines,
		offset:     0,
		matchIndex: -1,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

// visibleHeight returns the number of content lines visible (excluding the status bar).
func (m Model) visibleHeight() int {
	if m.height <= 1 {
		return m.height
	}
	return m.height - 1
}

// maxOffset returns the maximum scroll offset so the last line is still visible.
func (m Model) maxOffset() int {
	max := len(m.lines) - m.visibleHeight()
	if max < 0 {
		return 0
	}
	return max
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		return m, nil

	case tea.KeyMsg:
		// Handle search input mode
		if m.searching {
			switch msg.Type {
			case tea.KeyEscape:
				m.searching = false
				m.searchInput = ""
				return m, nil
			case tea.KeyEnter:
				m.searching = false
				m.searchQuery = m.searchInput
				m.searchInput = ""
				m.findMatches()
				if len(m.matchLines) > 0 {
					m.matchIndex = 0
					m.scrollToMatch()
				}
				return m, nil
			case tea.KeyBackspace:
				if len(m.searchInput) > 0 {
					m.searchInput = m.searchInput[:len(m.searchInput)-1]
				}
				return m, nil
			case tea.KeyRunes:
				m.searchInput += string(msg.Runes)
				return m, nil
			default:
				return m, nil
			}
		}

		// Normal mode
		switch {
		case msg.Type == tea.KeyCtrlC || msg.String() == "q":
			return m, tea.Quit

		case msg.Type == tea.KeyUp || msg.String() == "k":
			if m.offset > 0 {
				m.offset--
			}
			return m, nil

		case msg.Type == tea.KeyDown || msg.String() == "j":
			if m.offset < m.maxOffset() {
				m.offset++
			}
			return m, nil

		case msg.Type == tea.KeyPgDown || msg.String() == "d":
			m.offset += m.visibleHeight() / 2
			if m.offset > m.maxOffset() {
				m.offset = m.maxOffset()
			}
			return m, nil

		case msg.Type == tea.KeyPgUp || msg.String() == "u":
			m.offset -= m.visibleHeight() / 2
			if m.offset < 0 {
				m.offset = 0
			}
			return m, nil

		case msg.String() == "g":
			m.offset = 0
			return m, nil

		case msg.String() == "G":
			m.offset = m.maxOffset()
			return m, nil

		case msg.String() == "/":
			m.searching = true
			m.searchInput = ""
			return m, nil

		case msg.String() == "n":
			if len(m.matchLines) > 0 {
				m.matchIndex = (m.matchIndex + 1) % len(m.matchLines)
				m.scrollToMatch()
			}
			return m, nil

		case msg.String() == "N":
			if len(m.matchLines) > 0 {
				m.matchIndex--
				if m.matchIndex < 0 {
					m.matchIndex = len(m.matchLines) - 1
				}
				m.scrollToMatch()
			}
			return m, nil

		case msg.Type == tea.KeyEscape:
			m.searchQuery = ""
			m.matchLines = nil
			m.matchIndex = -1
			return m, nil
		}
	}
	return m, nil
}

func (m *Model) findMatches() {
	m.matchLines = nil
	m.matchIndex = -1
	if m.searchQuery == "" {
		return
	}
	query := strings.ToLower(m.searchQuery)
	for i, line := range m.lines {
		// Strip ANSI codes before matching
		plain := ansiRe.ReplaceAllString(line, "")
		if strings.Contains(strings.ToLower(plain), query) {
			m.matchLines = append(m.matchLines, i)
		}
	}
}

func (m *Model) scrollToMatch() {
	if m.matchIndex < 0 || m.matchIndex >= len(m.matchLines) {
		return
	}
	targetLine := m.matchLines[m.matchIndex]
	// Center the match on screen
	m.offset = targetLine - m.visibleHeight()/2
	if m.offset < 0 {
		m.offset = 0
	}
	if m.offset > m.maxOffset() {
		m.offset = m.maxOffset()
	}
}

var helpStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("241"))

func (m Model) View() string {
	if m.height <= 1 {
		return m.content
	}

	end := m.offset + m.visibleHeight()
	if end > len(m.lines) {
		end = len(m.lines)
	}
	if m.offset > end {
		m.offset = end
	}

	visible := m.lines[m.offset:end]
	view := strings.Join(visible, "\n")

	var status string
	if m.searching {
		status = helpStyle.Render(fmt.Sprintf("  /%s█", m.searchInput))
	} else if m.searchQuery != "" {
		matchInfo := "[no matches]"
		if len(m.matchLines) > 0 {
			matchInfo = fmt.Sprintf("[%d/%d]", m.matchIndex+1, len(m.matchLines))
		}
		status = helpStyle.Render(fmt.Sprintf("  /%s %s  n/N: next/prev • Esc: clear • q: quit", m.searchQuery, matchInfo))
	} else {
		status = helpStyle.Render("  ↑/↓/j/k: scroll • d/u: half-page • g/G: top/bottom • /: search • q: quit")
	}
	return view + "\n" + status
}
