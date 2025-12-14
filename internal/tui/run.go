package tui

import (
	"fmt"
	"os"

	"dev/internal/projects"

	"github.com/samber/lo"
	"github.com/samber/mo"

	tea "github.com/charmbracelet/bubbletea"
)

func Run(m tea.Model) mo.Result[projects.Project] {
	program := tea.NewProgram(m,
		tea.WithAltScreen(),
		tea.WithInputTTY(),
		tea.WithOutput(os.Stderr),
	)

	finalModel, err := program.Run()
	if err != nil {
		return mo.Err[projects.Project](fmt.Errorf("tui: %w", err))
	}

	model := finalModel.(Model)

	project, ok := lo.Find(model.projects, func(p projects.Project) bool {
		return p.Path == model.Selected
	})
	if !ok {
		return mo.Err[projects.Project](fmt.Errorf("no project found"))
	}

	return mo.Ok(project)
}
