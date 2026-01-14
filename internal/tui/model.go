package tui

import (
	"fmt"
	"strings"

	"dev/internal/projects"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Icons struct {
	Dir  string
	Term string
}

// Layout constants
const (
	smallWidthThreshold  = 120
	smallHeightThreshold = 25
	minContentWidth      = 40
	listPaddingLines     = 10
	minListHeight        = 5
	minFixedListHeight   = 10
	maxBoxedListHeight   = 20
	boxPadding           = 4
	innerPadding         = 6
	linePadding          = 4
	lineWidthBase        = 8 // icon (2) + spaces (3) + parens (2) + buffer (1)
)

type Model struct {
	keys     KeyMap
	projects []projects.Project
	filtered []projects.Project
	query    string
	cursor   int
	Selected string
	width    int
	height   int
	quitting bool
	icons    Icons
}

type layout struct {
	isSmall       bool
	contentWidth  int
	innerWidth    int
	maxListHeight int
}

func NewModel(p []projects.Project, keys KeyMap, icons Icons) Model {
	return Model{
		keys:     keys,
		projects: p,
		filtered: p,
		icons:    icons,
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
		switch {
		case key.Matches(msg, m.keys.Cancel):
			m.quitting = true
			return m, tea.Quit

		case key.Matches(msg, m.keys.Select):
			if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
				m.Selected = m.filtered[m.cursor].Path
			}
			m.quitting = true
			return m, tea.Quit

		case key.Matches(msg, m.keys.PrevItem):
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil

		case key.Matches(msg, m.keys.NextItem):
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			}
			return m, nil

		case key.Matches(msg, m.keys.Backspace):
			if len(m.query) > 0 {
				runes := []rune(m.query)
				m.query = string(runes[:len(runes)-1])
				m.filtered = projects.Filter(m.projects, m.query)
				m.cursor = 0
			}
			return m, nil

		case key.Matches(msg, m.keys.ClearQuery):
			if len(m.query) > 0 {
				m.query = ""
				m.filtered = m.projects
				m.cursor = 0
			}
			return m, nil

		default:
			if msg.Type == tea.KeyRunes {
				m.query += string(msg.Runes)
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
	content := renderHeader(l.innerWidth, m.keys, len(m.filtered), len(m.projects)) +
		renderInput(m.query) +
		renderList(m, l, m.filtered, m.cursor, 0) +
		renderFooter(l.innerWidth, m.keys)

	return "\n " + strings.ReplaceAll(content, "\n", "\n ")
}

func viewBoxed(m Model, l layout) string {
	fixedHeight := max(len(m.projects), minFixedListHeight)
	fixedHeight = min(fixedHeight, maxBoxedListHeight)
	fixedHeight = min(fixedHeight, l.maxListHeight)
	content := renderHeader(l.innerWidth, m.keys, len(m.filtered), len(m.projects)) +
		renderInput(m.query) +
		renderList(m, l, m.filtered, m.cursor, fixedHeight) +
		renderFooter(l.innerWidth, m.keys)

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

func renderHeader(innerWidth int, keys KeyMap, filteredCount, totalCount int) string {
	title := titleStyle.Render("Projects")
	counter := pathStyle.Render(fmt.Sprintf(" (%d/%d)", filteredCount, totalCount))
	escHint := keymapKeyStyle.Render(keys.Cancel.Help().Key)
	padding := max(innerWidth-lipgloss.Width(title)-lipgloss.Width(counter)-lipgloss.Width(escHint), 1)

	return title + counter + strings.Repeat(" ", padding) + escHint + "\n\n"
}

func renderInput(query string) string {
	return inputStyle.Render("> "+query+"_") + "\n\n"
}

func calculateVisibleRange(itemCount, visibleCount, cursor int) (start, end int) {
	if itemCount == 0 {
		return 0, 0
	}

	start = 0
	if cursor >= visibleCount {
		start = cursor - visibleCount + 1
	}

	end = start + visibleCount
	if end > itemCount {
		end = itemCount
		start = max(end-visibleCount, 0)
	}

	return start, end
}

func renderItem(p projects.Project, isSelected bool, maxName, innerWidth int, icon string) string {
	name := fmt.Sprintf("%-*s", maxName, p.Name)
	path := fmt.Sprintf("(%s)", p.Path)

	if isSelected {
		line := fmt.Sprintf("%s  %s %s", icon, name, path)
		return selectedStyle.Render(lipgloss.NewStyle().Width(innerWidth).Render(line))
	}

	return fmt.Sprintf("%s  %s %s",
		normalStyle.Render(icon),
		normalStyle.Render(name),
		pathStyle.Render(path),
	)
}

func renderList(m Model, l layout, filtered []projects.Project, cursor int, fixedHeight int) string {
	var content string
	var renderedLines int

	if len(filtered) == 0 {
		content = pathStyle.Render("No matches") + "\n"
		renderedLines = 1
	} else {
		var b strings.Builder
		listHeight := l.maxListHeight

		if fixedHeight > 0 {
			listHeight = fixedHeight
		}

		visibleCount := min(len(filtered), listHeight)
		start, end := calculateVisibleRange(len(filtered), visibleCount, cursor)
		maxName := maxNameLen(filtered)

		for i := start; i < end; i++ {
			p := filtered[i]
			b.WriteString(renderItem(p, i == cursor, maxName, l.innerWidth, m.icons.Dir))
			b.WriteString("\n")
		}
		content = b.String()
		renderedLines = end - start
	}

	if fixedHeight > 0 {
		padding := max(0, fixedHeight-renderedLines)
		return content + strings.Repeat("\n", padding)
	}

	return content
}

func renderFooter(innerWidth int, keys KeyMap) string {
	hints := fmt.Sprintf("%s %s %s",
		renderKeyHelp(keys.NextItem),
		renderKeyHelp(keys.PrevItem),
		renderKeyHelp(keys.Select),
	)

	return "\n" + lipgloss.PlaceHorizontal(innerWidth, lipgloss.Right, hints)
}

func renderKeyHelp(binding key.Binding) string {
	return fmt.Sprintf("%s %s",
		keymapLabelStyle.Render(binding.Help().Desc),
		keymapKeyStyle.Render(binding.Help().Key),
	)
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
