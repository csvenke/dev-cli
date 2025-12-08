package searchpath

import (
	"os"
	"path/filepath"
	"strings"
)

func Resolve(args []string) []string {
	if len(args) > 0 {
		return args
	}

	devPaths := strings.Fields(os.Getenv("DEV_PATHS"))
	if len(devPaths) > 0 {
		return devPaths
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return []string{}
	}

	return []string{homeDir}
}

func Expand(searchPaths []string) []string {
	if len(searchPaths) == 0 {
		return []string{}
	}

	var paths []string
	for _, p := range searchPaths {
		entries, err := os.ReadDir(p)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
				paths = append(paths, filepath.Join(p, entry.Name()))
			}
		}
	}
	return paths
}
