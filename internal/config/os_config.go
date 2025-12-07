package config

import (
	"os"

	"dev/internal/hooks"
)

func New() Config {
	editor := getEditorFromEnv()
	homeDir, _ := os.UserHomeDir()

	return Config{
		Args:     os.Args[1:],
		DevPaths: os.Getenv("DEV_PATHS"),
		HomeDir:  homeDir,
		Editor:   editor,
		Icons: Icons{
			Dir: "",
		},
		Hooks: []hooks.Hook{
			&hooks.TmuxHook{},
			&hooks.ZellijHook{},
		},
	}
}

func getEditorFromEnv() string {
	editor := os.Getenv("VISUAL")

	if editor == "" {
		editor = os.Getenv("EDITOR")
	}

	return editor
}
