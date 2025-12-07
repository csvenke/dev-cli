package tui

import (
	"fmt"
	"strings"

	"dev/internal/projects"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Layout constants
const (
	smallWidthThreshold  = 120
	smallHeightThreshold = 20
	minContentWidth      = 40
	listPaddingLines     = 10
	minListHeight        = 5
	boxPadding           = 4
	innerPadding         = 6
	linePadding          = 4
	lineWidthBase        = 8 // icon (2) + spaces (3) + parens (2) + buffer (1)
)

type Model struct {
	projects []projects.Project
	filtered []projects.Project
	query    string
	cursor   int
	Selected string
	width    int
	height   int
	quitting bool
}

type layout struct {
	isSmall       bool
	contentWidth  int
	innerWidth    int
	maxListHeight int
}

func NewModel(p []projects.Project) Model {
	return Model{
		projects: p,
		filtered: p,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
				m.Selected = m.filtered[m.cursor].Path
			}
			m.quitting = true
			return m, tea.Quit

		case "up", "ctrl+p":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil

		case "down", "ctrl+n":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			}
			return m, nil

		case "backspace":
			if len(m.query) > 0 {
				m.query = m.query[:len(m.query)-1]
				m.filtered = projects.Filter(m.projects, m.query)
				m.cursor = 0
			}
			return m, nil

		default:
			if len(msg.String()) == 1 {
				m.query += msg.String()
				m.filtered = projects.Filter(m.projects, m.query)
				m.cursor = 0
			}
			return m, nil
		}
	}

	return m, nil
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	l := calculateLayout(m.width, m.height, maxLineWidth(m.projects))

	if l.isSmall {
		return viewSmall(m, l)
	}
	return viewBoxed(m, l)
}

func viewSmall(m Model, l layout) string {
	content := renderHeader(l.innerWidth) +
		renderInput(m.query) +
		renderList(l, m.filtered, m.cursor) +
		renderFooter(l.innerWidth)

	return "\n " + strings.ReplaceAll(content, "\n", "\n ")
}

func viewBoxed(m Model, l layout) string {
	content := renderHeader(l.innerWidth) +
		renderInput(m.query) +
		renderList(l, m.filtered, m.cursor) +
		renderFooter(l.innerWidth)

	box := borderStyle.Width(l.contentWidth).Render(content)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func calculateLayout(width, height, maxLineWidth int) layout {
	isSmall := width < smallWidthThreshold || height < smallHeightThreshold

	var contentWidth, innerWidth int
	if isSmall {
		contentWidth = width - 2
		innerWidth = min(maxLineWidth, contentWidth)
	} else {
		contentWidth = max(maxLineWidth, minContentWidth)
		contentWidth = min(contentWidth, width-boxPadding)
		innerWidth = contentWidth - innerPadding
	}

	return layout{
		isSmall:       isSmall,
		contentWidth:  contentWidth,
		innerWidth:    innerWidth,
		maxListHeight: max(height-listPaddingLines, minListHeight),
	}
}

func renderHeader(innerWidth int) string {
	title := titleStyle.Render("Projects")
	escHint := keymapKeyStyle.Render("esc")
	padding := max(innerWidth-lipgloss.Width(title)-lipgloss.Width(escHint), 1)

	return title + strings.Repeat(" ", padding) + escHint + "\n\n"
}

func renderInput(query string) string {
	return inputStyle.Render("> "+query+"_") + "\n\n"
}

func renderList(l layout, filtered []projects.Project, cursor int) string {
	if len(filtered) == 0 {
		return pathStyle.Render("No matches") + "\n"
	}

	visibleCount := min(len(filtered), l.maxListHeight)

	start := 0
	if cursor >= visibleCount {
		start = cursor - visibleCount + 1
	}
	end := start + visibleCount
	if end > len(filtered) {
		end = len(filtered)
		start = max(end-visibleCount, 0)
	}

	maxName := maxNameLen(filtered)

	var b strings.Builder
	for i := start; i < end; i++ {
		p := filtered[i]
		icon := " "
		name := fmt.Sprintf("%-*s", maxName, p.Name)
		path := fmt.Sprintf("(%s)", p.Path)

		if i == cursor {
			line := fmt.Sprintf(" %s%s %s", icon, name, path)
			padded := fmt.Sprintf("%-*s", l.innerWidth, line)
			b.WriteString(selectedStyle.Render(padded))
		} else {
			b.WriteString("  " + normalStyle.Render(icon+name) + " " + pathStyle.Render(path))
		}
		b.WriteString("\n")
	}

	return b.String()
}

func renderFooter(innerWidth int) string {
	hints := renderKeyMap("next", "ctrl-n") + " " + renderKeyMap("prev", "ctrl-p")
	return "\n" + lipgloss.PlaceHorizontal(innerWidth, lipgloss.Right, hints)
}

func renderKeyMap(label, key string) string {
	return keymapLabelStyle.Render(label) + " " + keymapKeyStyle.Render(key)
}

func maxLineWidth(projs []projects.Project) int {
	maxWidth := 0
	for _, p := range projs {
		lineLen := lineWidthBase + len(p.Name) + len(p.Path)
		if lineLen > maxWidth {
			maxWidth = lineLen
		}
	}
	return maxWidth + linePadding
}

func maxNameLen(projs []projects.Project) int {
	maxLen := 0
	for _, p := range projs {
		if len(p.Name) > maxLen {
			maxLen = len(p.Name)
		}
	}
	return maxLen
}
