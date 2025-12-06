package hooks

import (
	"os"
	"os/exec"
)

type ZellijHook struct{}

func (h *ZellijHook) ShouldRun() bool {
	return os.Getenv("ZELLIJ") != ""
}

func (h *ZellijHook) Run(projectName string) error {
	return exec.Command("zellij", "action", "rename-tab", projectName).Run()
}
