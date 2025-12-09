package main

import (
	"flag"
	"fmt"
	"os"

	"dev/internal/app"
	"dev/internal/filesystem"
	"dev/internal/hooks"
	"dev/internal/tui"
)

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

	cfg := app.Config{
		Icons: tui.Icons{
			Dir: "",
		},
		Hooks: hooks.Default,
	}
	fs := &filesystem.RealFileSystem{}

	_, err := app.Run(flag.Args(), cfg, fs).Get()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}
