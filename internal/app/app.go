package app

import (
	"fmt"

	"github.com/samber/mo"

	"dev/internal/filesystem"
	"dev/internal/projects"
	"dev/internal/terminal"
	"dev/internal/tui"
)

type Config struct {
	Icons    tui.Icons
	Terminal terminal.Terminal
}

func Run(args []string, cfg Config, fs filesystem.FileSystem) mo.Result[string] {
	projectsResult, err := projects.Discover(fs, args).Get()
	if err != nil {
		return mo.Err[string](err)
	}
	if len(projectsResult) == 0 {
		return mo.Err[string](fmt.Errorf("no projects found"))
	}

	model := tui.NewModel(projectsResult, tui.DefaultKeyMap(), cfg.Icons)

	tuiResult, err := tui.Run(model).Get()
	if err != nil {
		return mo.Err[string](err)
	}

	_, err = fs.Chdir(tuiResult.Path).Get()
	if err != nil {
		return mo.Err[string](err)
	}

	cfg.Terminal.RenameTab(tuiResult.Name)

	return cfg.Terminal.OpenEditor(tuiResult.Path)
}
