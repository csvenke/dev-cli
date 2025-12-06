package config

import (
	"os"

	"dev/internal/hooks"
)

func New() Config {
	editor := os.Getenv("VISUAL")

	if editor == "" {
		editor = os.Getenv("EDITOR")
	}

	homeDir, _ := os.UserHomeDir()

	return Config{
		Args:     os.Args[1:],
		DevPaths: os.Getenv("DEV_PATHS"),
		HomeDir:  homeDir,
		Editor:   editor,
		Hooks: []hooks.Hook{
			&hooks.TmuxHook{},
			&hooks.ZellijHook{},
		},
	}
}
