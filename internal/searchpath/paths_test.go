package searchpath

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
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

func (m *mockFileSystem) ReadDir(path string) ([]os.DirEntry, error) {
	if m.readErr != nil {
		return nil, m.readErr
	}
	return m.dirs[path], nil
}

func TestResolve_PrefersArgs(t *testing.T) {
	result := Resolve([]string{"/custom/path", "/another/path"})

	if len(result) != 2 {
		t.Errorf("expected 2 paths, got %d", len(result))
	}
	if result[0] != "/custom/path" {
		t.Errorf("expected '/custom/path', got %q", result[0])
	}
	if result[1] != "/another/path" {
		t.Errorf("expected '/another/path', got %q", result[1])
	}
}

func TestResolve_FallsBackToDevPaths(t *testing.T) {
	t.Setenv("DEV_PATHS", "/repos /work /projects")

	result := Resolve([]string{})

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

func TestResolve_FallsBackToHome(t *testing.T) {
	t.Setenv("DEV_PATHS", "")

	// Create a temp dir to act as home
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	result := Resolve([]string{})

	if len(result) != 1 {
		t.Errorf("expected 1 path, got %d: %v", len(result), result)
	}
	if result[0] != tempHome {
		t.Errorf("expected path to be %q, got %q", tempHome, result[0])
	}
}

func TestExpand_FindsSubdirs(t *testing.T) {
	fs := &mockFileSystem{
		dirs: map[string][]os.DirEntry{
			"/home/user": {
				&mockDirEntry{name: "repos", isDir: true},
				&mockDirEntry{name: "work", isDir: true},
				&mockDirEntry{name: "file.txt", isDir: false},
			},
		},
	}

	result, err := Expand(fs, []string{"/home/user"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 paths (dirs only), got %d: %v", len(result), result)
	}
}

func TestExpand_ExcludesHiddenDirs(t *testing.T) {
	fs := &mockFileSystem{
		dirs: map[string][]os.DirEntry{
			"/home/user": {
				&mockDirEntry{name: ".config", isDir: true},
				&mockDirEntry{name: ".local", isDir: true},
				&mockDirEntry{name: "repos", isDir: true},
				&mockDirEntry{name: "Documents", isDir: true},
			},
		},
	}

	result, err := Expand(fs, []string{"/home/user"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 paths (non-hidden), got %d: %v", len(result), result)
	}
	for _, path := range result {
		if filepath.Base(path) == ".config" || filepath.Base(path) == ".local" {
			t.Errorf("hidden directory should not be included: %q", path)
		}
	}
}

func TestExpand_ReturnsEmptyOnNoSubdirs(t *testing.T) {
	fs := &mockFileSystem{
		dirs: map[string][]os.DirEntry{
			"/home/user": {},
		},
	}

	result, err := Expand(fs, []string{"/home/user"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("expected empty result, got %v", result)
	}
}

func TestExpand_BuildsFullPaths(t *testing.T) {
	fs := &mockFileSystem{
		dirs: map[string][]os.DirEntry{
			"/home/user": {
				&mockDirEntry{name: "repos", isDir: true},
				&mockDirEntry{name: "work", isDir: true},
			},
		},
	}

	result, err := Expand(fs, []string{"/home/user"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 paths, got %d: %v", len(result), result)
	}

	// Check that paths are full paths under home
	for _, p := range result {
		if !filepath.IsAbs(p) {
			t.Errorf("expected absolute path, got %q", p)
		}
		if filepath.Dir(p) != "/home/user" {
			t.Errorf("expected parent to be %q, got %q", "/home/user", filepath.Dir(p))
		}
	}
}

func TestExpand_ReturnsError(t *testing.T) {
	fs := &mockFileSystem{
		readErr: errors.New("read error"),
	}

	_, err := Expand(fs, []string{"/home/user"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "read error" {
		t.Errorf("expected 'read error', got %q", err.Error())
	}
}
