package paths

import (
	"errors"
	"io/fs"
	"testing"

	"dev/internal/testutil"
)

func TestFind_PrefersArgs(t *testing.T) {
	mockReadDir := func(path string) ([]fs.DirEntry, error) {
		return nil, nil
	}

	result := Find(mockReadDir, []string{"/custom/path", "/another/path"}, "/should/be/ignored", "/home/user")

	if len(result) != 2 {
		t.Errorf("expected 2 paths, got %d", len(result))
	}
	if result[0] != "/custom/path" {
		t.Errorf("expected '/custom/path', got %q", result[0])
	}
}

func TestFind_FallsBackToDevPaths(t *testing.T) {
	mockReadDir := func(path string) ([]fs.DirEntry, error) {
		return nil, nil
	}

	result := Find(mockReadDir, []string{}, "/repos /work /projects", "/home/user")

	if len(result) != 3 {
		t.Errorf("expected 3 paths, got %d", len(result))
	}
	expected := []string{"/repos", "/work", "/projects"}
	for i, exp := range expected {
		if result[i] != exp {
			t.Errorf("expected %q at index %d, got %q", exp, i, result[i])
		}
	}
}

func TestFind_FallsBackToHomeSubdirs(t *testing.T) {
	mockReadDir := func(path string) ([]fs.DirEntry, error) {
		return []fs.DirEntry{
			testutil.NewMockDir("repos"),
			testutil.NewMockDir("work"),
			testutil.NewMockFile("file.txt"),
		}, nil
	}

	result := Find(mockReadDir, []string{}, "", "/home/user")

	if len(result) != 2 {
		t.Errorf("expected 2 paths (dirs only), got %d", len(result))
	}
}

func TestFind_ExcludesHiddenDirs(t *testing.T) {
	mockReadDir := func(path string) ([]fs.DirEntry, error) {
		return []fs.DirEntry{
			testutil.NewMockDir(".config"),
			testutil.NewMockDir(".local"),
			testutil.NewMockDir("repos"),
			testutil.NewMockDir("Documents"),
		}, nil
	}

	result := Find(mockReadDir, []string{}, "", "/home/user")

	if len(result) != 2 {
		t.Errorf("expected 2 paths (non-hidden), got %d", len(result))
	}
	for _, path := range result {
		if path == "/home/user/.config" || path == "/home/user/.local" {
			t.Errorf("hidden directory should not be included: %q", path)
		}
	}
}

func TestFind_ReturnsNilOnEmptyHomeDir(t *testing.T) {
	mockReadDir := func(path string) ([]fs.DirEntry, error) {
		return nil, nil
	}

	result := Find(mockReadDir, []string{}, "", "")

	if result != nil {
		t.Errorf("expected nil result on empty home dir, got %v", result)
	}
}

func TestFind_HandlesReadDirError(t *testing.T) {
	mockReadDir := func(path string) ([]fs.DirEntry, error) {
		return nil, errors.New("permission denied")
	}

	result := Find(mockReadDir, []string{}, "", "/home/user")

	if result != nil {
		t.Errorf("expected nil result on read dir error, got %v", result)
	}
}

func TestFind_EmptyDevPathsReturnsHomeSubdirs(t *testing.T) {
	mockReadDir := func(path string) ([]fs.DirEntry, error) {
		return []fs.DirEntry{
			testutil.NewMockDir("repos"),
		}, nil
	}

	result := Find(mockReadDir, []string{}, "", "/home/user")

	if len(result) != 1 {
		t.Errorf("expected 1 path, got %d", len(result))
	}
	if result[0] != "/home/user/repos" {
		t.Errorf("expected '/home/user/repos', got %q", result[0])
	}
}

func TestFind_BuildsFullPaths(t *testing.T) {
	mockReadDir := func(path string) ([]fs.DirEntry, error) {
		return []fs.DirEntry{
			testutil.NewMockDir("repos"),
			testutil.NewMockDir("work"),
		}, nil
	}

	result := Find(mockReadDir, []string{}, "", "/home/user")

	expected := []string{"/home/user/repos", "/home/user/work"}
	for i, exp := range expected {
		if result[i] != exp {
			t.Errorf("expected %q at index %d, got %q", exp, i, result[i])
		}
	}
}
