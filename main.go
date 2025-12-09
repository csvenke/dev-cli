package main

import (
	"flag"
	"fmt"
	"os"

	"dev/internal/config"
	"dev/internal/editor"
	"dev/internal/filesystem"
	"dev/internal/hooks"
	"dev/internal/projects"
	"dev/internal/searchpath"
	"dev/internal/tui"
)

// Set by ldflags build time
var version string

func main() {
	var printVersion bool

	flag.BoolVar(&printVersion, "v", false, "print version")
	flag.BoolVar(&printVersion, "version", false, "print version")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: dev [options] [path...]\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if printVersion {
		if version == "" {
			version = "snapshot"
		}
		fmt.Printf("%s\n", version)
		os.Exit(0)
	}

	cfg := config.Config{
		Icons: config.Icons{
			Dir: "",
		},
		Hooks: []hooks.Hook{
			&hooks.TmuxHook{},
			&hooks.ZellijHook{},
		},
	}
	fs := &filesystem.RealFileSystem{}

	resolvedPaths := searchpath.Resolve(flag.Args())

	expandedPaths, err := searchpath.Expand(fs, resolvedPaths)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	discoveredProjects, err := projects.Discover(fs, expandedPaths)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	if len(discoveredProjects) == 0 {
		fmt.Fprintln(os.Stderr, "No projects found")
		os.Exit(1)
	}

	model := tui.NewModel(discoveredProjects, tui.DefaultKeyMap(), cfg.Icons)

	selectedProject, err := tui.Run(model)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if selectedProject.Path == "" {
		os.Exit(0)
	}

	if err := os.Chdir(selectedProject.Path); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	hooks.RunHooks(cfg.Hooks, selectedProject.Name)

	editor.Open(selectedProject.Path)
}
