package searchpath

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"dev/internal/filesystem"
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

func Expand(fs filesystem.FileSystem, searchPaths []string) ([]string, error) {
	if len(searchPaths) == 0 {
		return []string{}, nil
	}

	var paths []string
	var errs []error
	for _, p := range searchPaths {
		entries, err := fs.ReadDir(p)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
				paths = append(paths, filepath.Join(p, entry.Name()))
			}
		}
	}
	if len(errs) > 0 {
		return paths, errors.Join(errs...)
	}
	return paths, nil
}
