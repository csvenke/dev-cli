package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"dev/internal/config"
	"dev/internal/hooks"
	"dev/internal/projects"
	"dev/internal/searchpath"
	"dev/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	cfg := config.New()

	searchPaths := searchpath.Resolve(os.ReadDir, cfg.Args, cfg.DevPaths, cfg.HomeDir)
	allProjects := projects.Discover(filepath.WalkDir, searchPaths)

	if len(allProjects) == 0 {
		fmt.Fprintln(os.Stderr, "No projects found")
		os.Exit(1)
	}

	program := tea.NewProgram(
		tui.NewModel(allProjects),
		tea.WithAltScreen(),
	)

	finalModel, err := program.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	m := finalModel.(tui.Model)
	if m.Selected == "" {
		os.Exit(0)
	}

	if err := os.Chdir(m.Selected); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	hooks.Run(cfg.Hooks, filepath.Base(m.Selected))

	if cfg.Editor == "" {
		fmt.Println(m.Selected)
		os.Exit(0)
	}

	editorPath, err := exec.LookPath(cfg.Editor)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := syscall.Exec(editorPath, []string{cfg.Editor}, os.Environ()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
