package finder

import (
	"errors"
	"io/fs"
	"strings"
	"testing"

	"dev/internal/testutil"
)

type WalkEntry struct {
	path     string
	dirEntry fs.DirEntry
	err      error
}

func MockWalkDir(entries []WalkEntry) func(string, fs.WalkDirFunc) error {
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

func TestFind_DiscoversGitRepos(t *testing.T) {
	walkDir := MockWalkDir([]WalkEntry{
		{path: "/repos", dirEntry: testutil.NewMockDir("repos")},
		{path: "/repos/project-a", dirEntry: testutil.NewMockDir("project-a")},
		{path: "/repos/project-a/.git", dirEntry: testutil.NewMockDir(".git")},
		{path: "/repos/project-b", dirEntry: testutil.NewMockDir("project-b")},
		{path: "/repos/project-b/.git", dirEntry: testutil.NewMockDir(".git")},
	})

	projects := Find(walkDir, []string{"/repos"})

	if len(projects) != 2 {
		t.Errorf("expected 2 projects, got %d", len(projects))
	}
}

func TestFind_IgnoresNonGitDirs(t *testing.T) {
	walkDir := MockWalkDir([]WalkEntry{
		{path: "/repos", dirEntry: testutil.NewMockDir("repos")},
		{path: "/repos/not-a-project", dirEntry: testutil.NewMockDir("not-a-project")},
		{path: "/repos/not-a-project/src", dirEntry: testutil.NewMockDir("src")},
		{path: "/repos/real-project", dirEntry: testutil.NewMockDir("real-project")},
		{path: "/repos/real-project/.git", dirEntry: testutil.NewMockDir(".git")},
	})

	projects := Find(walkDir, []string{"/repos"})

	if len(projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(projects))
	}
	if projects[0].Name != "real-project" {
		t.Errorf("expected 'real-project', got %q", projects[0].Name)
	}
}

func TestFind_RespectsDepthLimit(t *testing.T) {
	walkDir := MockWalkDir([]WalkEntry{
		{path: "/repos", dirEntry: testutil.NewMockDir("repos")},
		{path: "/repos/org", dirEntry: testutil.NewMockDir("org")},
		{path: "/repos/org/project", dirEntry: testutil.NewMockDir("project")},
		{path: "/repos/org/project/.git", dirEntry: testutil.NewMockDir(".git")},
		{path: "/repos/org/deep/nested", dirEntry: testutil.NewMockDir("nested")},
		{path: "/repos/org/deep/nested/project", dirEntry: testutil.NewMockDir("project")},
		{path: "/repos/org/deep/nested/project/.git", dirEntry: testutil.NewMockDir(".git")},
	})

	projects := Find(walkDir, []string{"/repos"})

	if len(projects) != 1 {
		t.Errorf("expected 1 project (depth limit), got %d", len(projects))
	}
}

func TestFind_DeduplicatesProjects(t *testing.T) {
	walkDir := MockWalkDir([]WalkEntry{
		{path: "/repos/project", dirEntry: testutil.NewMockDir("project")},
		{path: "/repos/project/.git", dirEntry: testutil.NewMockDir(".git")},
	})

	projects := Find(walkDir, []string{"/repos", "/repos"})

	if len(projects) != 1 {
		t.Errorf("expected 1 project (deduplicated), got %d", len(projects))
	}
}

func TestFind_ReturnsEmptyForEmptyPaths(t *testing.T) {
	walkDir := MockWalkDir([]WalkEntry{})

	projects := Find(walkDir, []string{})

	if len(projects) != 0 {
		t.Errorf("expected 0 projects, got %d", len(projects))
	}
}

func TestFind_ReturnsEmptyForNoMatches(t *testing.T) {
	walkDir := MockWalkDir([]WalkEntry{
		{path: "/repos", dirEntry: testutil.NewMockDir("repos")},
		{path: "/repos/not-a-project", dirEntry: testutil.NewMockDir("not-a-project")},
	})

	projects := Find(walkDir, []string{"/repos"})

	if len(projects) != 0 {
		t.Errorf("expected 0 projects, got %d", len(projects))
	}
}

func TestFind_ExtractsProjectNameFromPath(t *testing.T) {
	walkDir := MockWalkDir([]WalkEntry{
		{path: "/home/user/repos", dirEntry: testutil.NewMockDir("repos")},
		{path: "/home/user/repos/my-project", dirEntry: testutil.NewMockDir("my-project")},
		{path: "/home/user/repos/my-project/.git", dirEntry: testutil.NewMockDir(".git")},
	})

	projects := Find(walkDir, []string{"/home/user/repos"})

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

func TestFind_HandlesMultipleSearchPaths(t *testing.T) {
	walkDir := MockWalkDir([]WalkEntry{
		{path: "/repos", dirEntry: testutil.NewMockDir("repos")},
		{path: "/repos/project-a", dirEntry: testutil.NewMockDir("project-a")},
		{path: "/repos/project-a/.git", dirEntry: testutil.NewMockDir(".git")},
		{path: "/work", dirEntry: testutil.NewMockDir("work")},
		{path: "/work/project-b", dirEntry: testutil.NewMockDir("project-b")},
		{path: "/work/project-b/.git", dirEntry: testutil.NewMockDir(".git")},
	})

	projects := Find(walkDir, []string{"/repos", "/work"})

	if len(projects) != 2 {
		t.Errorf("expected 2 projects, got %d", len(projects))
	}
}

func TestFind_ContinuesOnWalkError(t *testing.T) {
	walkDir := MockWalkDir([]WalkEntry{
		{path: "/repos", dirEntry: testutil.NewMockDir("repos")},
		{path: "/repos/unreadable", dirEntry: testutil.NewMockDir("unreadable"), err: errors.New("permission denied")},
		{path: "/repos/project", dirEntry: testutil.NewMockDir("project")},
		{path: "/repos/project/.git", dirEntry: testutil.NewMockDir(".git")},
	})

	projects := Find(walkDir, []string{"/repos"})

	if len(projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(projects))
	}
}
