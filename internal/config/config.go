package config

import (
	"dev/internal/hooks"
)

type Config struct {
	Hooks []hooks.Hook
	Icons Icons
}

type Icons struct {
	Dir string
}
