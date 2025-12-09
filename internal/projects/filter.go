package projects

import (
	"sort"
	"strings"
)

func Filter(projects []Project, query string) []Project {
	if query == "" {
		return projects
	}

	query = strings.ToLower(query)

	type scored struct {
		idx   int
		score int
	}

	matches := make([]scored, 0, len(projects)/4+1)
	for i, p := range projects {
		nameScore := FuzzyScore(query, strings.ToLower(p.Name))
		pathScore := FuzzyScore(query, strings.ToLower(p.Path))
		if bestScore := max(pathScore, nameScore); bestScore > 0 {
			matches = append(matches, scored{idx: i, score: bestScore})
		}
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].score > matches[j].score
	})

	result := make([]Project, len(matches))
	for i, m := range matches {
		result[i] = projects[m.idx]
	}

	return result
}

const (
	scoreMatch        = 1
	scoreConsecutive  = 2
	scoreWordBoundary = 3
)

// FuzzyScore calculates a fuzzy match score between query and target.
// Returns 0 if no match, higher scores indicate better matches.
// Exported for testing.
func FuzzyScore(query string, target string) int {
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
