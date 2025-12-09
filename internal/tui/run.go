package tui

import (
	"fmt"

	"dev/internal/projects"

	tea "github.com/charmbracelet/bubbletea"
)

// Run starts the TUI and returns the selected project.
// Returns zero-value Project if user cancelled, error if TUI failed.
func Run(m tea.Model) (projects.Project, error) {
	program := tea.NewProgram(m, tea.WithAltScreen())

	finalModel, err := program.Run()
	if err != nil {
		return projects.Project{}, fmt.Errorf("tui: %w", err)
	}

	model := finalModel.(Model)

	for _, p := range model.projects {
		if p.Path == model.Selected {
			return p, nil
		}
	}

	return projects.Project{}, nil
}
