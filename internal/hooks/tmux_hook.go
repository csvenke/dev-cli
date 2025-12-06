package hooks

import (
	"os"
	"os/exec"
)

type TmuxHook struct{}

func (h *TmuxHook) ShouldRun() bool {
	return os.Getenv("TMUX") != ""
}

func (h *TmuxHook) Run(projectName string) error {
	return exec.Command("tmux", "rename-window", projectName).Run()
}
