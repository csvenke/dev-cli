package terminal

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/samber/mo"
)

type Terminal interface {
	OpenEditor(path string) mo.Result[string]
	RenameTab(name string) error
}

func Detect() Terminal {
	if os.Getenv("ZELLIJ") != "" {
		return &Zellij{}
	}
	if os.Getenv("TMUX") != "" {
		return &Tmux{}
	}
	return &Default{}
}

type Zellij struct{}

func (z *Zellij) OpenEditor(path string) mo.Result[string] {
	return run("zellij", "", "edit", "--cwd", path, "-i", ".")
}

func (z *Zellij) RenameTab(name string) error {
	return exec.Command("zellij", "action", "rename-tab", name).Run()
}

type Tmux struct{}

func (t *Tmux) OpenEditor(path string) mo.Result[string] {
	editor, err := getEditorFromEnv().Get()
	if err != nil {
		return mo.Err[string](err)
	}
	return run("tmux", "", "respawn-pane", "-k", "-c", path, editor, ".")
}

func (t *Tmux) RenameTab(name string) error {
	return exec.Command("tmux", "rename-window", name).Run()
}

type Default struct{}

func (d *Default) OpenEditor(path string) mo.Result[string] {
	editor, err := getEditorFromEnv().Get()
	if err != nil {
		return mo.Err[string](err)
	}
	return run(editor, path, ".")
}

func (d *Default) RenameTab(name string) error {
	return nil
}

func run(name string, dir string, args ...string) mo.Result[string] {
	path, err := exec.LookPath(name)
	if err != nil {
		return mo.Err[string](err)
	}

	cmd := exec.Command(path, args...)
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return mo.Err[string](err)
	}

	return mo.Ok(path)
}

func getEditorFromEnv() mo.Result[string] {
	if editor := os.Getenv("EDITOR"); editor != "" {
		return mo.Ok(editor)
	}
	if editor := os.Getenv("VISUAL"); editor != "" {
		return mo.Ok(editor)
	}
	return mo.Err[string](fmt.Errorf("$VISUAL or $EDITOR is not set"))
}
