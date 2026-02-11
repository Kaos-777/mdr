package viewer

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	content string
	lines   []string
	offset  int
	height  int
	width   int
}

func NewModel(rendered string) Model {
	lines := strings.Split(rendered, "\n")
	return Model{
		content: rendered,
		lines:   lines,
		offset:  0,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		return m, nil

	case tea.KeyMsg:
		switch {
		case msg.Type == tea.KeyCtrlC || msg.String() == "q":
			return m, tea.Quit

		case msg.Type == tea.KeyUp || msg.String() == "k":
			if m.offset > 0 {
				m.offset--
			}
			return m, nil

		case msg.Type == tea.KeyDown || msg.String() == "j":
			maxOffset := len(m.lines) - m.height
			if maxOffset < 0 {
				maxOffset = 0
			}
			if m.offset < maxOffset {
				m.offset++
			}
			return m, nil

		case msg.Type == tea.KeyPgDown || msg.String() == "d":
			maxOffset := len(m.lines) - m.height
			if maxOffset < 0 {
				maxOffset = 0
			}
			m.offset += m.height / 2
			if m.offset > maxOffset {
				m.offset = maxOffset
			}
			return m, nil

		case msg.Type == tea.KeyPgUp || msg.String() == "u":
			m.offset -= m.height / 2
			if m.offset < 0 {
				m.offset = 0
			}
			return m, nil

		case msg.String() == "g":
			m.offset = 0
			return m, nil

		case msg.String() == "G":
			maxOffset := len(m.lines) - m.height
			if maxOffset < 0 {
				maxOffset = 0
			}
			m.offset = maxOffset
			return m, nil
		}
	}
	return m, nil
}

var helpStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("241"))

func (m Model) View() string {
	if m.height <= 1 {
		return m.content
	}

	end := m.offset + m.height - 1 // reserve 1 line for status
	if end > len(m.lines) {
		end = len(m.lines)
	}
	if m.offset > end {
		m.offset = end
	}

	visible := m.lines[m.offset:end]
	view := strings.Join(visible, "\n")

	status := helpStyle.Render("  ↑/↓/j/k: scroll • d/u: half-page • g/G: top/bottom • q: quit")
	return view + "\n" + status
}
