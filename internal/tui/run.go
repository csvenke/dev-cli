package tui

import (
	"fmt"

	"dev/internal/config"
	"dev/internal/projects"

	tea "github.com/charmbracelet/bubbletea"
)

// Run starts the TUI and returns the selected project.
// Returns zero-value Project if user cancelled, error if TUI failed.
func Run(allProjects []projects.Project, keyMap KeyMap, icons config.Icons) (projects.Project, error) {
	program := tea.NewProgram(
		NewModel(allProjects, keyMap, icons),
		tea.WithAltScreen(),
	)

	finalModel, err := program.Run()
	if err != nil {
		return projects.Project{}, fmt.Errorf("tui: %w", err)
	}

	m := finalModel.(Model)

	for _, p := range allProjects {
		if p.Path == m.Selected {
			return p, nil
		}
	}

	return projects.Project{}, nil
}
