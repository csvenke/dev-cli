package hooks

import (
	"os"
	"os/exec"

	"github.com/samber/lo"
)

type Hook func(projectName string) error

var Default = []Hook{
	tmuxHook,
	zellijHook,
}

func RunAll(hooks []Hook, projectName string) {
	lo.ForEach(hooks, func(hook Hook, _ int) { _ = hook(projectName) })
}

func tmuxHook(name string) error {
	if os.Getenv("TMUX") == "" {
		return nil
	}
	return exec.Command("tmux", "rename-window", name).Run()
}

func zellijHook(name string) error {
	if os.Getenv("ZELLIJ") == "" {
		return nil
	}
	return exec.Command("zellij", "action", "rename-tab", name).Run()
}
