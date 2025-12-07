package projects

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type Project struct {
	Name string
	Path string
}

func Discover(walkDir func(string, fs.WalkDirFunc) error, searchPaths []string) []Project {
	seen := make(map[string]bool)
	var result []Project

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
					result = append(result, Project{
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

func Filter(projects []Project, query string) []Project {
	if query == "" {
		return projects
	}

	query = strings.ToLower(query)
	var result []Project
	for _, p := range projects {
		if strings.Contains(strings.ToLower(p.Name), query) ||
			strings.Contains(strings.ToLower(p.Path), query) {
			result = append(result, p)
		}
	}
	return result
}
