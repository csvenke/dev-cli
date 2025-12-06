package projects

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

func TestFilter_EmptyQueryReturnsAll(t *testing.T) {
	projects := []Project{
		{Name: "project-a", Path: "/repos/project-a"},
		{Name: "project-b", Path: "/repos/project-b"},
	}

	result := Filter(projects, "")

	if len(result) != len(projects) {
		t.Errorf("expected %d projects, got %d", len(projects), len(result))
	}
}

func TestFilter_MatchesByName(t *testing.T) {
	projects := []Project{
		{Name: "frontend", Path: "/repos/frontend"},
		{Name: "backend", Path: "/repos/backend"},
		{Name: "api", Path: "/repos/api"},
	}

	result := Filter(projects, "front")

	if len(result) != 1 {
		t.Errorf("expected 1 match, got %d", len(result))
	}
	if result[0].Name != "frontend" {
		t.Errorf("expected 'frontend', got %q", result[0].Name)
	}
}

func TestFilter_MatchesByPath(t *testing.T) {
	projects := []Project{
		{Name: "app", Path: "/home/user/repos/app"},
		{Name: "lib", Path: "/home/user/libs/lib"},
	}

	result := Filter(projects, "libs")

	if len(result) != 1 {
		t.Errorf("expected 1 match, got %d", len(result))
	}
	if result[0].Name != "lib" {
		t.Errorf("expected 'lib', got %q", result[0].Name)
	}
}

func TestFilter_CaseInsensitive(t *testing.T) {
	projects := []Project{
		{Name: "MyProject", Path: "/repos/MyProject"},
	}

	tests := []struct {
		query string
	}{
		{"myproject"},
		{"MYPROJECT"},
		{"MyProject"},
		{"myPROJECT"},
	}

	for _, tt := range tests {
		result := Filter(projects, tt.query)
		if len(result) != 1 {
			t.Errorf("query %q: expected 1 match, got %d", tt.query, len(result))
		}
	}
}

func TestFilter_NoMatchesReturnsEmpty(t *testing.T) {
	projects := []Project{
		{Name: "frontend", Path: "/repos/frontend"},
		{Name: "backend", Path: "/repos/backend"},
	}

	result := Filter(projects, "nonexistent")

	if len(result) != 0 {
		t.Errorf("expected 0 matches, got %d", len(result))
	}
}

func TestFilter_PartialMatch(t *testing.T) {
	projects := []Project{
		{Name: "my-awesome-project", Path: "/repos/my-awesome-project"},
	}

	partials := []string{"my", "awesome", "project", "my-awesome", "awesome-project"}

	for _, query := range partials {
		result := Filter(projects, query)
		if len(result) != 1 {
			t.Errorf("query %q: expected 1 match, got %d", query, len(result))
		}
	}
}

func TestFilter_EmptyProjectList(t *testing.T) {
	var projects []Project

	result := Filter(projects, "anything")

	if len(result) != 0 {
		t.Errorf("expected empty result, got %d items", len(result))
	}
}

// Property-based tests

func TestFilter_ResultIsSubsetOfInput(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("filtered result is always a subset of input", prop.ForAll(
		func(names []string, query string) bool {
			var projects []Project
			for _, name := range names {
				projects = append(projects, Project{Name: name, Path: "/" + name})
			}

			result := Filter(projects, query)

			// Every result must exist in original
			for _, r := range result {
				found := false
				for _, p := range projects {
					if r.Name == p.Name && r.Path == p.Path {
						found = true
						break
					}
				}
				if !found {
					return false
				}
			}
			return true
		},
		gen.SliceOf(gen.AlphaString()),
		gen.AlphaString(),
	))

	properties.TestingRun(t)
}

func TestFilter_ResultNeverLargerThanInput(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("result length never exceeds input length", prop.ForAll(
		func(names []string, query string) bool {
			var projects []Project
			for _, name := range names {
				projects = append(projects, Project{Name: name, Path: "/" + name})
			}

			result := Filter(projects, query)

			return len(result) <= len(projects)
		},
		gen.SliceOf(gen.AlphaString()),
		gen.AlphaString(),
	))

	properties.TestingRun(t)
}

func TestFilter_EmptyQueryReturnsAllProperty(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("empty query returns all projects", prop.ForAll(
		func(names []string) bool {
			var projects []Project
			for _, name := range names {
				projects = append(projects, Project{Name: name, Path: "/" + name})
			}

			result := Filter(projects, "")

			return len(result) == len(projects)
		},
		gen.SliceOf(gen.AlphaString()),
	))

	properties.TestingRun(t)
}

func TestFilter_FilteringIsIdempotent(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("filtering twice with same query gives same result", prop.ForAll(
		func(names []string, query string) bool {
			var projects []Project
			for _, name := range names {
				projects = append(projects, Project{Name: name, Path: "/" + name})
			}

			first := Filter(projects, query)
			second := Filter(first, query)

			if len(first) != len(second) {
				return false
			}
			for i := range first {
				if first[i] != second[i] {
					return false
				}
			}
			return true
		},
		gen.SliceOf(gen.AlphaString()),
		gen.AlphaString(),
	))

	properties.TestingRun(t)
}
