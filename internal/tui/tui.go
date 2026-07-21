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
	quitting bool
	confirm  bool // true when user pressed enter to apply
	width    int
}

var (
	titleStyle  = lipgloss.NewStyle().Bold(true)
	cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	onStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	offStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	helpStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

// Run opens a multi-select TUI. Returns the enabled name list and whether to apply.
func Run(items []Item) (enabled []string, apply bool, err error) {
	m := model{items: items}
	p := tea.NewProgram(m)
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
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case " ":
			if len(m.items) > 0 {
				m.items[m.cursor].Enabled = !m.items[m.cursor].Enabled
			}
		case "a":
			for i := range m.items {
				m.items[i].Enabled = true
			}
		case "n":
			for i := range m.items {
				m.items[i].Enabled = false
			}
		case "enter":
			m.confirm = true
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.quitting && !m.confirm {
		return ""
	}
	var b strings.Builder
	b.WriteString(titleStyle.Render("skill-manager — toggle skills (space) · enter apply · q quit"))
	b.WriteString("\n\n")
	for i, it := range m.items {
		cursor := "  "
		if i == m.cursor {
			cursor = cursorStyle.Render("> ")
		}
		mark := offStyle.Render("[ ]")
		if it.Enabled {
			mark = onStyle.Render("[x]")
		}
		b.WriteString(fmt.Sprintf("%s%s %s\n", cursor, mark, it.Name))
	}
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("a=all  n=none  j/k=move"))
	b.WriteString("\n")
	return b.String()
}
