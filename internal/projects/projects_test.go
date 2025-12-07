package projects

import (
	"errors"
	"io/fs"
	"strings"
	"testing"

	"dev/internal/testutil"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

type walkEntry struct {
	path     string
	dirEntry fs.DirEntry
	err      error
}

func mockWalkDir(entries []walkEntry) func(string, fs.WalkDirFunc) error {
	return func(root string, fn fs.WalkDirFunc) error {
		for _, entry := range entries {
			if strings.HasPrefix(entry.path, root) || entry.path == root {
				err := fn(entry.path, entry.dirEntry, entry.err)
				if err == fs.SkipDir {
					continue
				}
				if err != nil {
					return err
				}
			}
		}
		return nil
	}
}

func TestDiscover_DiscoversGitRepos(t *testing.T) {
	walkDir := mockWalkDir([]walkEntry{
		{path: "/repos", dirEntry: testutil.NewMockDir("repos")},
		{path: "/repos/project-a", dirEntry: testutil.NewMockDir("project-a")},
		{path: "/repos/project-a/.git", dirEntry: testutil.NewMockDir(".git")},
		{path: "/repos/project-b", dirEntry: testutil.NewMockDir("project-b")},
		{path: "/repos/project-b/.git", dirEntry: testutil.NewMockDir(".git")},
	})

	projects := Discover(walkDir, []string{"/repos"})

	if len(projects) != 2 {
		t.Errorf("expected 2 projects, got %d", len(projects))
	}
}

func TestDiscover_IgnoresNonGitDirs(t *testing.T) {
	walkDir := mockWalkDir([]walkEntry{
		{path: "/repos", dirEntry: testutil.NewMockDir("repos")},
		{path: "/repos/not-a-project", dirEntry: testutil.NewMockDir("not-a-project")},
		{path: "/repos/not-a-project/src", dirEntry: testutil.NewMockDir("src")},
		{path: "/repos/real-project", dirEntry: testutil.NewMockDir("real-project")},
		{path: "/repos/real-project/.git", dirEntry: testutil.NewMockDir(".git")},
	})

	projects := Discover(walkDir, []string{"/repos"})

	if len(projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(projects))
	}
	if projects[0].Name != "real-project" {
		t.Errorf("expected 'real-project', got %q", projects[0].Name)
	}
}

func TestDiscover_RespectsDepthLimit(t *testing.T) {
	walkDir := mockWalkDir([]walkEntry{
		{path: "/repos", dirEntry: testutil.NewMockDir("repos")},
		{path: "/repos/org", dirEntry: testutil.NewMockDir("org")},
		{path: "/repos/org/project", dirEntry: testutil.NewMockDir("project")},
		{path: "/repos/org/project/.git", dirEntry: testutil.NewMockDir(".git")},
		{path: "/repos/org/deep/nested", dirEntry: testutil.NewMockDir("nested")},
		{path: "/repos/org/deep/nested/project", dirEntry: testutil.NewMockDir("project")},
		{path: "/repos/org/deep/nested/project/.git", dirEntry: testutil.NewMockDir(".git")},
	})

	projects := Discover(walkDir, []string{"/repos"})

	if len(projects) != 1 {
		t.Errorf("expected 1 project (depth limit), got %d", len(projects))
	}
}

func TestDiscover_DeduplicatesProjects(t *testing.T) {
	walkDir := mockWalkDir([]walkEntry{
		{path: "/repos/project", dirEntry: testutil.NewMockDir("project")},
		{path: "/repos/project/.git", dirEntry: testutil.NewMockDir(".git")},
	})

	projects := Discover(walkDir, []string{"/repos", "/repos"})

	if len(projects) != 1 {
		t.Errorf("expected 1 project (deduplicated), got %d", len(projects))
	}
}

func TestDiscover_ReturnsEmptyForEmptyPaths(t *testing.T) {
	walkDir := mockWalkDir([]walkEntry{})

	projects := Discover(walkDir, []string{})

	if len(projects) != 0 {
		t.Errorf("expected 0 projects, got %d", len(projects))
	}
}

func TestDiscover_ReturnsEmptyForNoMatches(t *testing.T) {
	walkDir := mockWalkDir([]walkEntry{
		{path: "/repos", dirEntry: testutil.NewMockDir("repos")},
		{path: "/repos/not-a-project", dirEntry: testutil.NewMockDir("not-a-project")},
	})

	projects := Discover(walkDir, []string{"/repos"})

	if len(projects) != 0 {
		t.Errorf("expected 0 projects, got %d", len(projects))
	}
}

func TestDiscover_ExtractsProjectNameFromPath(t *testing.T) {
	walkDir := mockWalkDir([]walkEntry{
		{path: "/home/user/repos", dirEntry: testutil.NewMockDir("repos")},
		{path: "/home/user/repos/my-project", dirEntry: testutil.NewMockDir("my-project")},
		{path: "/home/user/repos/my-project/.git", dirEntry: testutil.NewMockDir(".git")},
	})

	projects := Discover(walkDir, []string{"/home/user/repos"})

	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
	if projects[0].Name != "my-project" {
		t.Errorf("expected name 'my-project', got %q", projects[0].Name)
	}
	if projects[0].Path != "/home/user/repos/my-project" {
		t.Errorf("expected path '/home/user/repos/my-project', got %q", projects[0].Path)
	}
}

func TestDiscover_HandlesMultipleSearchPaths(t *testing.T) {
	walkDir := mockWalkDir([]walkEntry{
		{path: "/repos", dirEntry: testutil.NewMockDir("repos")},
		{path: "/repos/project-a", dirEntry: testutil.NewMockDir("project-a")},
		{path: "/repos/project-a/.git", dirEntry: testutil.NewMockDir(".git")},
		{path: "/work", dirEntry: testutil.NewMockDir("work")},
		{path: "/work/project-b", dirEntry: testutil.NewMockDir("project-b")},
		{path: "/work/project-b/.git", dirEntry: testutil.NewMockDir(".git")},
	})

	projects := Discover(walkDir, []string{"/repos", "/work"})

	if len(projects) != 2 {
		t.Errorf("expected 2 projects, got %d", len(projects))
	}
}

func TestDiscover_ContinuesOnWalkError(t *testing.T) {
	walkDir := mockWalkDir([]walkEntry{
		{path: "/repos", dirEntry: testutil.NewMockDir("repos")},
		{path: "/repos/unreadable", dirEntry: testutil.NewMockDir("unreadable"), err: errors.New("permission denied")},
		{path: "/repos/project", dirEntry: testutil.NewMockDir("project")},
		{path: "/repos/project/.git", dirEntry: testutil.NewMockDir(".git")},
	})

	projects := Discover(walkDir, []string{"/repos"})

	if len(projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(projects))
	}
}

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
