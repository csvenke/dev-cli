package app

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/samber/mo"

	"dev/internal/filesystem"
	"dev/internal/hooks"
	"dev/internal/projects"
	"dev/internal/tui"
)

type Config struct {
	Icons tui.Icons
	Hooks []hooks.Hook
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

	hooks.RunAll(cfg.Hooks, tuiResult.Name)

	return openEditor(tuiResult.Path)
}

func openEditor(path string) mo.Result[string] {
	editor, err := getEditorFromEnv().Get()
	if err != nil {
		return mo.Err[string](err)
	}

	editorPath, err := exec.LookPath(editor)
	if err != nil {
		return mo.Err[string](err)
	}
	if err := syscall.Exec(editorPath, []string{editor, path}, os.Environ()); err != nil {
		return mo.Err[string](err)
	}

	return mo.Ok(editorPath)
}

func getEditorFromEnv() mo.Result[string] {
	if editor := os.Getenv("VISUAL"); editor != "" {
		return mo.Ok(editor)
	}
	if editor := os.Getenv("EDITOR"); editor != "" {
		return mo.Ok(editor)
	}
	return mo.Err[string](fmt.Errorf("$VISUAL or $EDITOR is not set"))
}
