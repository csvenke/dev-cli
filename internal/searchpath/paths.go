package searchpath

import (
	"io/fs"
	"path/filepath"
	"strings"
)

func Resolve(readDir func(string) ([]fs.DirEntry, error), args []string, devPaths string, homeDir string) []string {
	if len(args) > 0 {
		return args
	}

	if devPaths != "" {
		return strings.Fields(devPaths)
	}

	if homeDir == "" {
		return nil
	}

	entries, err := readDir(homeDir)
	if err != nil {
		return nil
	}

	var paths []string
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			paths = append(paths, filepath.Join(homeDir, entry.Name()))
		}
	}
	return paths
}
