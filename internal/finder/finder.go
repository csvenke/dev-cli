package finder

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"dev/internal/projects"
)

func Find(walkDir func(string, fs.WalkDirFunc) error, searchPaths []string) []projects.Project {
	seen := make(map[string]bool)
	var result []projects.Project

	for _, searchPath := range searchPaths {
		_ = walkDir(searchPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}

			relPath, _ := filepath.Rel(searchPath, path)
			depth := strings.Count(relPath, string(os.PathSeparator))
			if depth > 2 {
				return filepath.SkipDir
			}

			if d.IsDir() && d.Name() == ".git" {
				projectPath := filepath.Dir(path)
				if !seen[projectPath] {
					seen[projectPath] = true
					result = append(result, projects.Project{
						Name: filepath.Base(projectPath),
						Path: projectPath,
					})
				}
				return filepath.SkipDir
			}

			return nil
		})
	}

	return result
}
