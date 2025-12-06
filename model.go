package main

import (
	"fmt"
	"strings"

	"dev/internal/projects"
	"dev/internal/selection"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

func initialModel(p []projects.Project) Model {
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
			m.Selected = selection.SelectProject(m.filtered, m.cursor)
			m.quitting = true
			return m, tea.Quit

		case "up", "ctrl+p":
			m.cursor = selection.MoveCursorUp(m.cursor)
			return m, nil

		case "down", "ctrl+n":
			m.cursor = selection.MoveCursorDown(m.cursor, len(m.filtered)-1)
			return m, nil

		case "backspace":
			if len(m.query) > 0 {
				m.query = selection.DeleteLastChar(m.query)
				m.filtered = projects.Filter(m.projects, m.query)
				m.cursor = 0
			}
			return m, nil

		default:
			if len(msg.String()) == 1 {
				m.query = selection.AppendChar(m.query, msg.String())
				m.filtered = projects.Filter(m.projects, m.query)
				m.cursor = 0
			}
			return m, nil
		}
	}

	return m, nil
}

func RenderKeyMap(label string, key string) string {
	return keymapLabelStyle.Render(label) + " " + keymapKeyStyle.Render(key)
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	isSmall := m.width < 150 || m.height < 20

	maxLineWidth := 0
	for _, p := range m.projects {
		lineLen := 5 + len(p.Name) + 1 + 1 + len(p.Path) + 1
		if lineLen > maxLineWidth {
			maxLineWidth = lineLen
		}
	}
	// Add some padding for aesthetics
	maxLineWidth += 4

	// Ensure minimum width for title row ("Projects" + "Esc" + spacing)
	minWidth := 40

	var contentWidth int
	var innerWidth int
	if isSmall {
		contentWidth = m.width - 2
		innerWidth = min(maxLineWidth, contentWidth)
	} else {
		contentWidth = max(maxLineWidth, minWidth)
		contentWidth = min(contentWidth, m.width-4)
		innerWidth = contentWidth - 6
	}

	maxListHeight := max(m.height-10, 5)

	var b strings.Builder

	title := titleStyle.Render("Projects")
	escHint := keymapKeyStyle.Render("esc")
	titlePadding := max(innerWidth-lipgloss.Width(title)-lipgloss.Width(escHint), 1)
	titleLine := title + strings.Repeat(" ", titlePadding) + escHint
	b.WriteString(titleLine)
	b.WriteString("\n\n")

	prompt := "> "
	cursor := "_"
	b.WriteString(inputStyle.Render(prompt + m.query + cursor))
	b.WriteString("\n\n")

	if len(m.filtered) == 0 {
		b.WriteString(pathStyle.Render("No matches"))
		b.WriteString("\n")
	} else {
		visibleCount := min(len(m.filtered), maxListHeight)

		start := 0
		if m.cursor >= visibleCount {
			start = m.cursor - visibleCount + 1
		}
		end := start + visibleCount
		if end > len(m.filtered) {
			end = len(m.filtered)
			start = max(end-visibleCount, 0)
		}

		// Calculate max name length for alignment
		maxNameLen := 0
		for _, p := range m.filtered {
			if len(p.Name) > maxNameLen {
				maxNameLen = len(p.Name)
			}
		}

		for i := start; i < end; i++ {
			p := m.filtered[i]
			folderIcon := "" + " "
			paddedName := fmt.Sprintf("%-*s", maxNameLen, p.Name)
			pathText := fmt.Sprintf("(%s)", p.Path)

			if i == m.cursor {
				lineContent := fmt.Sprintf(" %s%s %s", folderIcon, paddedName, pathText)
				paddedLine := fmt.Sprintf("%-*s", innerWidth, lineContent)
				b.WriteString(selectedStyle.Render(paddedLine))
			} else {
				line := "  " + normalStyle.Render(folderIcon+paddedName) + " " + pathStyle.Render(pathText)
				b.WriteString(line)
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	hints := fmt.Sprintf("%s %s", RenderKeyMap("next", "ctrl-n"), RenderKeyMap("prev", "ctrl-p"))
	hintsLine := lipgloss.PlaceHorizontal(innerWidth, lipgloss.Right, hints)
	b.WriteString(hintsLine)

	content := b.String()

	if isSmall {
		return "\n " + strings.ReplaceAll(content, "\n", "\n ")
	}

	box := borderStyle.Width(contentWidth).Render(content)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		box,
	)
}
