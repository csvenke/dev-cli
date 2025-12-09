package projects

import (
	"fmt"
	"testing"
)

// generateProjects creates a slice of 'count' mock projects for benchmarking.
func generateProjects(count int) []Project {
	projects := make([]Project, count)
	for i := range count {
		name := fmt.Sprintf("project-%04d", i)
		path := fmt.Sprintf("/repos/%s", name)
		projects[i] = Project{Name: name, Path: path}
	}
	return projects
}

func BenchmarkFilter1000Projects(b *testing.B) {
	allProjects := generateProjects(1000)
	query := "proj-9" // A query that will match some projects, testing the fuzzy logic

	for b.Loop() {
		Filter(allProjects, query)
	}
}

func BenchmarkFilter1000Projects_NoMatch(b *testing.B) {
	allProjects := generateProjects(1000)
	query := "nonexistentquery" // A query that will not match any projects

	for b.Loop() {
		Filter(allProjects, query)
	}
}

func BenchmarkFilter1000Projects_ExactMatch(b *testing.B) {
	allProjects := generateProjects(1000)
	query := "project-0500" // A query that will exactly match one project

	for b.Loop() {
		Filter(allProjects, query)
	}
}

func BenchmarkFilter1000Projects_EmptyQuery(b *testing.B) {
	allProjects := generateProjects(1000)
	query := "" // An empty query should return all projects without fuzzy matching

	for b.Loop() {
		Filter(allProjects, query)
	}
}
