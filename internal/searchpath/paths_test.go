package searchpath

import (
	"os"
	"path/filepath"
	"testing"
)

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

func TestDiscover_FindsSubdirs(t *testing.T) {
	tempHome := t.TempDir()

	if err := os.MkdirAll(filepath.Join(tempHome, "repos"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tempHome, "work"), 0755); err != nil {
		t.Fatal(err)
	}
	// Create a file (should be ignored)
	if err := os.WriteFile(filepath.Join(tempHome, "file.txt"), []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	result := Expand([]string{tempHome})

	if len(result) != 2 {
		t.Errorf("expected 2 paths (dirs only), got %d: %v", len(result), result)
	}
}

func TestDiscover_ExcludesHiddenDirs(t *testing.T) {
	tempHome := t.TempDir()

	if err := os.MkdirAll(filepath.Join(tempHome, ".config"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tempHome, ".local"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tempHome, "repos"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tempHome, "Documents"), 0755); err != nil {
		t.Fatal(err)
	}

	result := Expand([]string{tempHome})

	if len(result) != 2 {
		t.Errorf("expected 2 paths (non-hidden), got %d: %v", len(result), result)
	}
	for _, path := range result {
		if filepath.Base(path) == ".config" || filepath.Base(path) == ".local" {
			t.Errorf("hidden directory should not be included: %q", path)
		}
	}
}

func TestDiscover_ReturnsEmptyOnNoSubdirs(t *testing.T) {
	tempHome := t.TempDir()

	result := Expand([]string{tempHome})

	if len(result) != 0 {
		t.Errorf("expected empty result, got %v", result)
	}
}

func TestDiscover_BuildsFullPaths(t *testing.T) {
	tempHome := t.TempDir()

	if err := os.MkdirAll(filepath.Join(tempHome, "repos"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tempHome, "work"), 0755); err != nil {
		t.Fatal(err)
	}

	result := Expand([]string{tempHome})

	if len(result) != 2 {
		t.Fatalf("expected 2 paths, got %d: %v", len(result), result)
	}

	// Check that paths are full paths under home
	for _, p := range result {
		if !filepath.IsAbs(p) {
			t.Errorf("expected absolute path, got %q", p)
		}
		if filepath.Dir(p) != tempHome {
			t.Errorf("expected parent to be %q, got %q", tempHome, filepath.Dir(p))
		}
	}
}
