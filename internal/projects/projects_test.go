package projects

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/samber/mo"
)

type mockDirEntry struct {
	name  string
	isDir bool
}

func (m *mockDirEntry) Name() string {
	return m.name
}

func (m *mockDirEntry) IsDir() bool {
	return m.isDir
}

func (m *mockDirEntry) Type() os.FileMode {
	if m.isDir {
		return os.ModeDir
	}
	return 0
}

func (m *mockDirEntry) Info() (os.FileInfo, error) {
	return nil, nil
}

type mockFileSystem struct {
	dirs    map[string][]os.DirEntry
	readErr error
}

func (m *mockFileSystem) ReadDir(path string) mo.Result[[]os.DirEntry] {
	if m.readErr != nil {
		return mo.Err[[]os.DirEntry](m.readErr)
	}
	return mo.Ok(m.dirs[path])
}

func (m *mockFileSystem) Chdir(path string) mo.Result[string] {
	return mo.Ok(path)
}

func TestDiscover_DiscoversGitRepos(t *testing.T) {
	fs := &mockFileSystem{
		dirs: map[string][]os.DirEntry{
			"/home/user": {
				&mockDirEntry{name: "project-a", isDir: true},
				&mockDirEntry{name: "project-b", isDir: true},
			},
			"/home/user/project-a": {&mockDirEntry{name: ".git", isDir: true}},
			"/home/user/project-b": {&mockDirEntry{name: ".git", isDir: true}},
		},
	}
	result := Discover(fs, []string{"/home/user"})
	if result.IsError() {
		t.Fatalf("unexpected error: %v", result.Error())
	}
	projects := result.MustGet()

	if len(projects) != 2 {
		t.Errorf("expected 2 projects, got %d", len(projects))
	}
}

func TestDiscover_IgnoresNonGitDirs(t *testing.T) {
	fs := &mockFileSystem{
		dirs: map[string][]os.DirEntry{
			"/home/user": {
				&mockDirEntry{name: "not-a-project", isDir: true},
				&mockDirEntry{name: "real-project", isDir: true},
			},
			"/home/user/not-a-project": {},
			"/home/user/real-project":  {&mockDirEntry{name: ".git", isDir: true}},
		},
	}
	result := Discover(fs, []string{"/home/user"})
	if result.IsError() {
		t.Fatalf("unexpected error: %v", result.Error())
	}
	projects := result.MustGet()

	if len(projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(projects))
	}
	if projects[0].Name != "real-project" {
		t.Errorf("expected 'real-project', got %q", projects[0].Name)
	}
}

func TestDiscover_RespectsDepthLimit(t *testing.T) {
	fs := &mockFileSystem{
		dirs: map[string][]os.DirEntry{
			"/": {
				&mockDirEntry{name: "org", isDir: true},
			},
			"/org": {
				&mockDirEntry{name: "project", isDir: true},
				&mockDirEntry{name: "deep", isDir: true},
			},
			"/org/project": {&mockDirEntry{name: ".git", isDir: true}},
			"/org/deep": {
				&mockDirEntry{name: "nested", isDir: true},
			},
			"/org/deep/nested": {
				&mockDirEntry{name: "project", isDir: true},
			},
			"/org/deep/nested/project": {&mockDirEntry{name: ".git", isDir: true}},
		},
	}
	result := Discover(fs, []string{"/"})
	if result.IsError() {
		t.Fatalf("unexpected error: %v", result.Error())
	}
	projects := result.MustGet()

	if len(projects) != 1 {
		t.Errorf("expected 1 project (depth limit), got %d", len(projects))
	}
}

func TestDiscover_DeduplicatesProjects(t *testing.T) {
	fs := &mockFileSystem{
		dirs: map[string][]os.DirEntry{
			"/home/user": {
				&mockDirEntry{name: "project", isDir: true},
			},
			"/home/user/project": {&mockDirEntry{name: ".git", isDir: true}},
		},
	}
	result := Discover(fs, []string{"/home/user", "/home/user"})
	if result.IsError() {
		t.Fatalf("unexpected error: %v", result.Error())
	}
	projects := result.MustGet()

	if len(projects) != 1 {
		t.Errorf("expected 1 project (deduplicated), got %d", len(projects))
	}
}

func TestDiscover_ReturnsEmptyForEmptyPaths(t *testing.T) {
	fs := &mockFileSystem{}
	result := Discover(fs, []string{})
	if result.IsError() {
		t.Fatalf("unexpected error: %v", result.Error())
	}
	projects := result.MustGet()

	if len(projects) != 0 {
		t.Errorf("expected 0 projects, got %d", len(projects))
	}
}

func TestDiscover_ReturnsEmptyForNoMatches(t *testing.T) {
	fs := &mockFileSystem{
		dirs: map[string][]os.DirEntry{
			"/home/user": {
				&mockDirEntry{name: "not-a-project", isDir: true},
			},
		},
	}
	result := Discover(fs, []string{"/home/user"})
	if result.IsError() {
		t.Fatalf("unexpected error: %v", result.Error())
	}
	projects := result.MustGet()

	if len(projects) != 0 {
		t.Errorf("expected 0 projects, got %d", len(projects))
	}
}

func TestDiscover_ExtractsProjectNameFromPath(t *testing.T) {
	fs := &mockFileSystem{
		dirs: map[string][]os.DirEntry{
			"/home/user": {
				&mockDirEntry{name: "my-project", isDir: true},
			},
			"/home/user/my-project": {&mockDirEntry{name: ".git", isDir: true}},
		},
	}
	result := Discover(fs, []string{"/home/user"})
	if result.IsError() {
		t.Fatalf("unexpected error: %v", result.Error())
	}
	projects := result.MustGet()

	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
	if projects[0].Name != "my-project" {
		t.Errorf("expected name 'my-project', got %q", projects[0].Name)
	}
	if projects[0].Path != "/home/user/my-project" {
		t.Errorf("expected path '/home/user/my-project', got %q", projects[0].Path)
	}
}

func TestDiscover_HandlesMultipleSearchPaths(t *testing.T) {
	fs := &mockFileSystem{
		dirs: map[string][]os.DirEntry{
			"/root1": {
				&mockDirEntry{name: "project-a", isDir: true},
			},
			"/root1/project-a": {&mockDirEntry{name: ".git", isDir: true}},
			"/root2": {
				&mockDirEntry{name: "project-b", isDir: true},
			},
			"/root2/project-b": {&mockDirEntry{name: ".git", isDir: true}},
		},
	}
	result := Discover(fs, []string{"/root1", "/root2"})
	if result.IsError() {
		t.Fatalf("unexpected error: %v", result.Error())
	}
	projects := result.MustGet()

	if len(projects) != 2 {
		t.Errorf("expected 2 projects, got %d", len(projects))
	}
}

func TestDiscover_SkipsHiddenDirectories(t *testing.T) {
	fs := &mockFileSystem{
		dirs: map[string][]os.DirEntry{
			"/home/user": {
				&mockDirEntry{name: ".hidden", isDir: true},
				&mockDirEntry{name: "visible-project", isDir: true},
			},
			"/home/user/.hidden": {
				&mockDirEntry{name: "secret-project", isDir: true},
			},
			"/home/user/.hidden/secret-project": {&mockDirEntry{name: ".git", isDir: true}},
			"/home/user/visible-project":        {&mockDirEntry{name: ".git", isDir: true}},
		},
	}
	result := Discover(fs, []string{"/home/user"})
	if result.IsError() {
		t.Fatalf("unexpected error: %v", result.Error())
	}
	projects := result.MustGet()

	if len(projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(projects))
	}
	if projects[0].Name != "visible-project" {
		t.Errorf("expected 'visible-project', got %q", projects[0].Name)
	}
}

func TestDiscover_HandlesNonExistentPath(t *testing.T) {
	fs := &mockFileSystem{}
	result := Discover(fs, []string{"/nonexistent/path/that/does/not/exist"})
	if result.IsError() {
		t.Fatalf("unexpected error: %v", result.Error())
	}
	projects := result.MustGet()

	if len(projects) != 0 {
		t.Errorf("expected 0 projects, got %d", len(projects))
	}
}

func TestDiscover_ReturnsError(t *testing.T) {
	fs := &mockFileSystem{
		readErr: errors.New("read error"),
	}
	result := Discover(fs, []string{"/home/user"})
	if result.IsOk() {
		t.Fatal("expected error, got ok")
	}
	err := result.Error()
	if err.Error() != "read error" {
		t.Errorf("expected 'read error', got %q", err.Error())
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

func TestFilter_FuzzyMatchesOutOfOrder(t *testing.T) {
	projects := []Project{
		{Name: "dev-cli", Path: "/repos/dev-cli"},
		{Name: "frontend", Path: "/repos/frontend"},
	}

	// "dc" should match "dev-cli" (d...c)
	result := Filter(projects, "dc")

	if len(result) != 1 {
		t.Errorf("expected 1 match, got %d", len(result))
	}
	if result[0].Name != "dev-cli" {
		t.Errorf("expected 'dev-cli', got %q", result[0].Name)
	}
}

func TestFilter_FuzzyRanksBetterMatchesFirst(t *testing.T) {
	projects := []Project{
		{Name: "other-project", Path: "/repos/other-project"},
		{Name: "dev-cli", Path: "/repos/dev-cli"},
		{Name: "xdevxcli", Path: "/repos/xdevxcli"},
	}

	result := Filter(projects, "devcli")

	if len(result) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(result))
	}
	// "dev-cli" should rank higher than "xdevxcli" (word boundary matches)
	if result[0].Name != "dev-cli" {
		t.Errorf("expected 'dev-cli' first, got %q", result[0].Name)
	}
}

func TestFilter_FuzzyPrefersWordBoundaryMatches(t *testing.T) {
	projects := []Project{
		{Name: "xdev", Path: "/repos/xdev"},
		{Name: "dev-cli", Path: "/repos/dev-cli"},
	}

	result := Filter(projects, "dev")

	if len(result) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(result))
	}
	// "dev-cli" should rank higher (starts with "dev")
	if result[0].Name != "dev-cli" {
		t.Errorf("expected 'dev-cli' first (word boundary), got %q", result[0].Name)
	}
}

func TestScore_ScoreIsNonNegative(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("score is always non-negative", prop.ForAll(
		func(query, target string) bool {
			return FuzzyScore(query, target) >= 0
		},
		gen.AlphaString(),
		gen.AlphaString(),
	))

	properties.TestingRun(t)
}

func TestScore_EmptyQueryAlwaysMatches(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("empty query matches any target", prop.ForAll(
		func(target string) bool {
			return FuzzyScore("", target) > 0
		},
		gen.AlphaString(),
	))

	properties.TestingRun(t)
}

func TestScore_QueryLongerThanTargetNeverMatches(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("query longer than target returns 0", prop.ForAll(
		func(query, target string) bool {
			if len(query) > len(target) {
				return FuzzyScore(query, target) == 0
			}
			return true
		},
		gen.AlphaString(),
		gen.AlphaString(),
	))

	properties.TestingRun(t)
}

func TestScore_ExactMatchAlwaysSucceeds(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("identical strings always match", prop.ForAll(
		func(s string) bool {
			if len(s) == 0 {
				return true
			}
			return FuzzyScore(s, s) > 0
		},
		gen.AlphaString(),
	))

	properties.TestingRun(t)
}

func TestScore_PrefixAlwaysMatches(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("prefix of target always matches", prop.ForAll(
		func(target string) bool {
			if len(target) == 0 {
				return true
			}
			prefixLen := max(len(target)/2, 1)
			prefix := target[:prefixLen]
			return FuzzyScore(strings.ToLower(prefix), strings.ToLower(target)) > 0
		},
		gen.AlphaString(),
	))

	properties.TestingRun(t)
}

func TestScore_MatchesAllCharactersInOrder(t *testing.T) {
	tests := []struct {
		query  string
		target string
		match  bool
	}{
		{"abc", "abc", true},
		{"abc", "aXbXc", true},
		{"dc", "dev-cli", true},
		{"fnt", "frontend", true},
		{"abc", "cba", false},
		{"abc", "ab", false},
		{"xyz", "abc", false},
	}

	for _, tt := range tests {
		score := FuzzyScore(tt.query, tt.target)
		matched := score > 0
		if matched != tt.match {
			t.Errorf("FuzzyScore(%q, %q): expected match=%v, got score=%d",
				tt.query, tt.target, tt.match, score)
		}
	}
}

func TestScore_ConsecutiveMatchesScoreHigher(t *testing.T) {
	// "dev" in "dev-cli" (consecutive) should score higher than "dev" in "dXeXv"
	consecutive := FuzzyScore("dev", "dev-cli")
	scattered := FuzzyScore("dev", "dXeXv")

	if consecutive <= scattered {
		t.Errorf("consecutive match should score higher: %d vs %d", consecutive, scattered)
	}
}

func TestScore_WordBoundaryMatchesScoreHigher(t *testing.T) {
	// "dev" at start should score higher than "dev" in middle
	atStart := FuzzyScore("dev", "dev-cli")
	inMiddle := FuzzyScore("dev", "mydev")

	if atStart <= inMiddle {
		t.Errorf("word boundary match should score higher: %d vs %d", atStart, inMiddle)
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

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
