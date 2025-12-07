package config

import (
	"dev/internal/hooks"
)

type Config struct {
	Args     []string
	DevPaths string
	HomeDir  string
	Editor   string
	Hooks    []hooks.Hook
	Icons    Icons
}

type Icons struct {
	Dir string
}
