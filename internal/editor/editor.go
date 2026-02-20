package editor

import (
	"fmt"
	"io/fs"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tirthpatell/mdr/internal/markdown"
)

type Model struct {
	buffer    *Buffer
	filePath  string
	fileMode  fs.FileMode
	cursorRow int
	cursorCol int
	offsetRow int
	width     int
	height    int
	editWidth int
	showHelp  bool
	err       error
	saveMsg   string
}

func NewModel(content string, filePath string) Model {
	return Model{
		buffer:   NewBuffer(content),
		filePath: filePath,
		fileMode: 0644,
	}
}

func NewModelFromFile(path string) (Model, error) {
	info, err := os.Stat(path)
	if err != nil {
		return Model{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Model{}, err
	}
	m := NewModel(string(data), path)
	m.fileMode = info.Mode().Perm()
	return m, nil
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.editWidth = msg.Width / 2
		return m, nil

	case tea.KeyMsg:
		// Clear transient messages on any non-save keypress
		if msg.Type != tea.KeyCtrlS {
			m.err = nil
			m.saveMsg = ""
		}

		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyCtrlS:
			m.err = m.save()
			if m.err == nil {
				m.buffer.ResetModified()
				m.saveMsg = "Saved!"
			} else {
				m.saveMsg = ""
			}
			return m, nil

		case tea.KeyCtrlH:
			m.showHelp = !m.showHelp
			return m, nil

		case tea.KeyUp:
			if m.cursorRow > 0 {
				m.cursorRow--
				m.clampCol()
				m.scrollToCursor()
			}
			return m, nil

		case tea.KeyDown:
			if m.cursorRow < m.buffer.LineCount()-1 {
				m.cursorRow++
				m.clampCol()
				m.scrollToCursor()
			}
			return m, nil

		case tea.KeyLeft:
			if m.cursorCol > 0 {
				m.cursorCol--
			} else if m.cursorRow > 0 {
				m.cursorRow--
				m.cursorCol = m.buffer.LineLen(m.cursorRow)
			}
			return m, nil

		case tea.KeyRight:
			lineLen := m.buffer.LineLen(m.cursorRow)
			if m.cursorCol < lineLen {
				m.cursorCol++
			} else if m.cursorRow < m.buffer.LineCount()-1 {
				m.cursorRow++
				m.cursorCol = 0
			}
			return m, nil

		case tea.KeyHome:
			m.cursorCol = 0
			return m, nil

		case tea.KeyEnd:
			m.cursorCol = m.buffer.LineLen(m.cursorRow)
			return m, nil

		case tea.KeyEnter:
			m.buffer.InsertNewline(m.cursorRow, m.cursorCol)
			m.cursorRow++
			m.cursorCol = 0
			m.scrollToCursor()
			return m, nil

		case tea.KeyBackspace:
			if m.cursorCol > 0 {
				m.buffer.Backspace(m.cursorRow, m.cursorCol)
				m.cursorCol--
			} else if m.cursorRow > 0 {
				m.cursorCol = m.buffer.JoinLines(m.cursorRow)
				m.cursorRow--
				m.scrollToCursor()
			}
			return m, nil

		case tea.KeyDelete:
			lineLen := m.buffer.LineLen(m.cursorRow)
			if m.cursorCol < lineLen {
				m.buffer.DeleteChar(m.cursorRow, m.cursorCol)
			} else if m.cursorRow < m.buffer.LineCount()-1 {
				m.buffer.JoinLines(m.cursorRow + 1)
			}
			return m, nil

		case tea.KeyTab:
			m.buffer.InsertChar(m.cursorRow, m.cursorCol, '\t')
			m.cursorCol++
			return m, nil

		case tea.KeyRunes:
			for _, r := range msg.Runes {
				m.buffer.InsertChar(m.cursorRow, m.cursorCol, r)
				m.cursorCol++
			}
			return m, nil
		}
	}
	return m, nil
}

func (m *Model) clampCol() {
	lineLen := m.buffer.LineLen(m.cursorRow)
	if m.cursorCol > lineLen {
		m.cursorCol = lineLen
	}
}

func (m *Model) scrollToCursor() {
	editHeight := m.editableHeight()
	if editHeight <= 0 {
		return
	}
	if m.cursorRow < m.offsetRow {
		m.offsetRow = m.cursorRow
	}
	if m.cursorRow >= m.offsetRow+editHeight {
		m.offsetRow = m.cursorRow - editHeight + 1
	}
}

func (m Model) editableHeight() int {
	h := m.height - 2
	if h < 0 {
		h = 0
	}
	return h
}

func (m Model) save() error {
	return os.WriteFile(m.filePath, []byte(m.buffer.String()), m.fileMode)
}

var (
	statusStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("230")).
			Padding(0, 1)
	lineNumStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Width(4).
			Align(lipgloss.Right)
	cursorLineStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("235"))
	previewBorder = lipgloss.NewStyle().
			BorderLeft(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("241"))
	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
	errStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Bold(true)
	savedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10"))
)

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	editHeight := m.editableHeight()
	editW := m.editWidth
	if editW == 0 {
		editW = m.width / 2
	}

	if m.showHelp {
		help := helpStyle.Render(strings.Join([]string{
			"Editor Help",
			"",
			"  Arrow keys   Move cursor",
			"  Home/End     Start/end of line",
			"  Ctrl+S       Save file",
			"  Ctrl+H       Toggle this help",
			"  Ctrl+C       Quit",
			"",
			"  Press Ctrl+H to close",
		}, "\n"))
		return help
	}

	var editorLines []string
	for i := m.offsetRow; i < m.offsetRow+editHeight && i < m.buffer.LineCount(); i++ {
		lineNum := lineNumStyle.Render(fmt.Sprintf("%d", i+1))
		line := m.buffer.GetLine(i)

		if i == m.cursorRow {
			runes := []rune(line)
			col := m.cursorCol
			if col > len(runes) {
				col = len(runes)
			}
			displayLine := string(runes[:col]) + "\u2588" + string(runes[col:])
			editorLines = append(editorLines, lineNum+" "+cursorLineStyle.Render(truncate(displayLine, editW-6)))
		} else {
			editorLines = append(editorLines, lineNum+" "+truncate(line, editW-6))
		}
	}
	for len(editorLines) < editHeight {
		editorLines = append(editorLines, lineNumStyle.Render("~"))
	}

	editorPane := strings.Join(editorLines, "\n")

	previewW := m.width - editW - 1
	rendered, renderErr := markdown.Render(m.buffer.String())
	if renderErr != nil {
		rendered = errStyle.Render("Preview error: " + renderErr.Error())
	}
	previewLines := strings.Split(rendered, "\n")
	if len(previewLines) > editHeight {
		previewLines = previewLines[:editHeight]
	}
	for i := range previewLines {
		previewLines[i] = truncate(previewLines[i], previewW-2)
	}
	for len(previewLines) < editHeight {
		previewLines = append(previewLines, "")
	}
	previewPane := previewBorder.Render(strings.Join(previewLines, "\n"))

	body := lipgloss.JoinHorizontal(lipgloss.Top, editorPane, previewPane)

	modIndicator := ""
	if m.buffer.Modified() {
		modIndicator = " [+]"
	}
	statusText := fmt.Sprintf(" %s%s  Ln %d, Col %d  Ctrl+S: save  Ctrl+H: help  Ctrl+C: quit",
		m.filePath, modIndicator, m.cursorRow+1, m.cursorCol+1)
	if m.err != nil {
		statusText += "  " + errStyle.Render("Error: "+m.err.Error())
	}
	if m.saveMsg != "" {
		statusText += "  " + savedStyle.Render(m.saveMsg)
	}
	status := statusStyle.Render(statusText)

	return body + "\n" + status
}

func truncate(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) > maxWidth {
		return string(runes[:maxWidth])
	}
	return s
}
