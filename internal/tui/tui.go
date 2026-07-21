package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Item is one skill row.
type Item struct {
	Name    string
	Enabled bool
}

type model struct {
	items    []Item
	cursor   int
	offset   int // first visible row index
	height   int // terminal height
	width    int
	quitting bool
	confirm  bool
}

var (
	titleStyle    = lipgloss.NewStyle().Bold(true)
	cursorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true)
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Background(lipgloss.Color("238"))
	onStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	offStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

const (
	headerLines = 3 // title + blank + (optional status)
	footerLines = 3 // blank + help + blank
)

// Run opens a multi-select TUI. Returns the enabled name list and whether to apply.
func Run(items []Item) (enabled []string, apply bool, err error) {
	m := model{items: items, height: 24, width: 80}
	p := tea.NewProgram(m, tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return nil, false, err
	}
	out := final.(model)
	if !out.confirm {
		return nil, false, nil
	}
	var names []string
	for _, it := range out.items {
		if it.Enabled {
			names = append(names, it.Name)
		}
	}
	return names, true, nil
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ensureVisible()
	case tea.KeyMsg:
		switch {
		case msg.Type == tea.KeyCtrlC || msg.String() == "q" || msg.Type == tea.KeyEsc:
			m.quitting = true
			return m, tea.Quit
		case msg.Type == tea.KeyUp || msg.String() == "k" || msg.String() == "ctrl+p":
			if len(m.items) == 0 {
				break
			}
			if m.cursor <= 0 {
				m.cursor = len(m.items) - 1
			} else {
				m.cursor--
			}
			m.ensureVisible()
		case msg.Type == tea.KeyDown || msg.String() == "j" || msg.String() == "ctrl+n":
			if len(m.items) == 0 {
				break
			}
			if m.cursor >= len(m.items)-1 {
				m.cursor = 0
			} else {
				m.cursor++
			}
			m.ensureVisible()
		case msg.Type == tea.KeyPgUp:
			m.cursor -= m.visibleRows()
			if m.cursor < 0 {
				m.cursor = 0
			}
			m.ensureVisible()
		case msg.Type == tea.KeyPgDown:
			m.cursor += m.visibleRows()
			if m.cursor > len(m.items)-1 {
				m.cursor = len(m.items) - 1
			}
			if m.cursor < 0 {
				m.cursor = 0
			}
			m.ensureVisible()
		case msg.Type == tea.KeyHome:
			m.cursor = 0
			m.ensureVisible()
		case msg.Type == tea.KeyEnd:
			if len(m.items) > 0 {
				m.cursor = len(m.items) - 1
			}
			m.ensureVisible()
		case msg.Type == tea.KeySpace || msg.String() == " ":
			if len(m.items) > 0 {
				m.items[m.cursor].Enabled = !m.items[m.cursor].Enabled
			}
		case msg.String() == "a":
			for i := range m.items {
				m.items[i].Enabled = true
			}
		case msg.String() == "n":
			for i := range m.items {
				m.items[i].Enabled = false
			}
		case msg.Type == tea.KeyEnter:
			m.confirm = true
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) visibleRows() int {
	rows := m.height - headerLines - footerLines
	if rows < 3 {
		rows = 3
	}
	return rows
}

func (m *model) ensureVisible() {
	vis := m.visibleRows()
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+vis {
		m.offset = m.cursor - vis + 1
	}
	if m.offset < 0 {
		m.offset = 0
	}
	maxOff := len(m.items) - vis
	if maxOff < 0 {
		maxOff = 0
	}
	if m.offset > maxOff {
		m.offset = maxOff
	}
}

func (m model) enabledCount() int {
	n := 0
	for _, it := range m.items {
		if it.Enabled {
			n++
		}
	}
	return n
}

func (m model) View() string {
	if m.quitting && !m.confirm {
		return ""
	}
	vis := m.visibleRows()
	end := m.offset + vis
	if end > len(m.items) {
		end = len(m.items)
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render("skill-manager — 空格切换 · 回车应用 · q 取消"))
	b.WriteString("\n")
	b.WriteString(helpStyle.Render(fmt.Sprintf(
		"已选 %d / %d · 光标 %d · ↑↓/jk 移动 · PgUp/PgDn 翻页",
		m.enabledCount(), len(m.items), m.cursor+1,
	)))
	b.WriteString("\n\n")

	if m.offset > 0 {
		b.WriteString(helpStyle.Render(fmt.Sprintf("  … 上方还有 %d 项", m.offset)))
		b.WriteString("\n")
	}

	for i := m.offset; i < end; i++ {
		it := m.items[i]
		mark := offStyle.Render("[ ]")
		if it.Enabled {
			mark = onStyle.Render("[x]")
		}
		line := fmt.Sprintf("%s %s", mark, it.Name)
		if i == m.cursor {
			line = cursorStyle.Render("> ") + selectedStyle.Render(line)
		} else {
			line = "  " + line
		}
		b.WriteString(line)
		b.WriteString("\n")
	}

	remain := len(m.items) - end
	if remain > 0 {
		b.WriteString(helpStyle.Render(fmt.Sprintf("  … 下方还有 %d 项", remain)))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("a 全选  n 全不选  Home/End 首尾  ↑↓循环"))
	b.WriteString("\n")
	return b.String()
}
