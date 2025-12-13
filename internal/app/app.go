package app

import (
	"fmt"

	"github.com/samber/mo"

	"dev/internal/filesystem"
	"dev/internal/projects"
	"dev/internal/terminal"
	"dev/internal/tui"
)

type Flags struct {
	PrintPath bool
}

type Config struct {
	Args  []string
	Flags Flags
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

	if cfg.Flags.PrintPath {
		return mo.Ok(tuiResult.Path)
	}

	title := fmt.Sprintf("%s %s", cfg.Icons.Term, tuiResult.Name)
	_ = cfg.Term.RenameTab(title)

	_, err = cfg.Term.OpenEditor(tuiResult.Path).Get()
	if err != nil {
		return mo.Err[string](err)
	}

	return mo.Ok("")
}
