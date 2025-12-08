package editor

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func Open(path string) {
	e := getEditorByEnv()

	if e == "" {
		fmt.Println(path)
		os.Exit(0)
	}

	editorPath, err := exec.LookPath(e)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := syscall.Exec(editorPath, []string{e}, os.Environ()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func getEditorByEnv() string {
	e := os.Getenv("VISUAL")

	if e == "" {
		e = os.Getenv("EDITOR")
	}

	return e
}
