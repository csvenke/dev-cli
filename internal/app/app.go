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
	Args  []string
	Icons tui.Icons
	Term  terminal.Terminal
	Fs    filesystem.FileSystem
}

func Run(cfg Config) mo.Result[string] {
	projectsResult, err := projects.Discover(cfg.Fs, cfg.Args).Get()
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

	_, err = cfg.Fs.Chdir(tuiResult.Path).Get()
	if err != nil {
		return mo.Err[string](err)
	}

	_ = cfg.Term.RenameTab(tuiResult.Name)

	return cfg.Term.OpenEditor(tuiResult.Path)
}
