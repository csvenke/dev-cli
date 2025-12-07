package projects

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
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

	type scored struct {
		project Project
		score   int
	}

	var matches []scored
	for _, p := range projects {
		nameScore := fuzzyScore(query, strings.ToLower(p.Name))
		pathScore := fuzzyScore(query, strings.ToLower(p.Path))
		bestScore := max(pathScore, nameScore)

		if bestScore > 0 {
			matches = append(matches, scored{project: p, score: bestScore})
		}
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].score > matches[j].score
	})

	result := make([]Project, len(matches))
	for i, m := range matches {
		result[i] = m.project
	}

	return result
}

const (
	scoreMatch        = 1
	scoreConsecutive  = 2
	scoreWordBoundary = 3
)

func fuzzyScore(query, target string) int {
	if len(query) == 0 {
		return 1
	}
	if len(query) > len(target) {
		return 0
	}

	score := 0
	qi := 0
	prevMatch := -2

	for ti := 0; ti < len(target) && qi < len(query); ti++ {
		if len(target)-ti < len(query)-qi {
			return 0
		}

		if target[ti] == query[qi] {
			score += scoreMatch
			if ti == prevMatch+1 {
				score += scoreConsecutive
			}
			if ti == 0 || !isLetter(target[ti-1]) {
				score += scoreWordBoundary
			}
			prevMatch = ti
			qi++
		}
	}

	if qi == len(query) {
		return score
	}
	return 0
}

func isLetter(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}
