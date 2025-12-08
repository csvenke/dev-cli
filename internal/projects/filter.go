package projects

import (
	"dev/internal/fuzzy"
	"sort"
	"strings"
)

// Filter returns projects matching the query, sorted by relevance.
func Filter(projects []Project, query string) []Project {
	if query == "" {
		return projects
	}

	query = strings.ToLower(query)

	type scored struct {
		idx   int
		score int
	}

	// Pre-allocate with capacity hint assuming ~25% match rate
	matches := make([]scored, 0, len(projects)/4+1)
	for i, p := range projects {
		nameScore := fuzzy.Score(query, strings.ToLower(p.Name))
		pathScore := fuzzy.Score(query, strings.ToLower(p.Path))
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
