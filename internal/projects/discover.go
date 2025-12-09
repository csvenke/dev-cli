package projects

import (
	"errors"
	"path/filepath"
	"sync"

	"dev/internal/filesystem"
)

// Project represents a discovered git repository.
type Project struct {
	Name string
	Path string
}

// Discover finds all git repositories within the given search paths.
// It walks directories in parallel and deduplicates results.
func Discover(fs filesystem.FileSystem, searchPaths []string) ([]Project, error) {
	if len(searchPaths) == 0 {
		return []Project{}, nil
	}

	var wg sync.WaitGroup
	resultCh := make(chan Project, 64)
	errCh := make(chan error, 64)

	for _, sp := range searchPaths {
		wg.Add(1)
		go func(searchPath string) {
			defer wg.Done()
			walkRecursive(fs, searchPath, 0, resultCh, errCh)
		}(sp)
	}

	go func() {
		wg.Wait()
		close(resultCh)
		close(errCh)
	}()

	// Collect and deduplicate
	seen := make(map[string]struct{})
	result := make([]Project, 0, 64)
	for p := range resultCh {
		if _, exists := seen[p.Path]; !exists {
			seen[p.Path] = struct{}{}
			result = append(result, p)
		}
	}
	var errs []error
	for err := range errCh {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return result, errors.Join(errs...)
	}

	return result, nil
}

func walkRecursive(fs filesystem.FileSystem, dir string, depth int, out chan<- Project, errCh chan<- error) {
	if depth > 2 {
		return
	}

	entries, err := fs.ReadDir(dir)
	if err != nil {
		errCh <- err
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()

		// Found a git repo
		if name == ".git" {
			out <- Project{
				Name: filepath.Base(dir),
				Path: dir,
			}
			return // Don't descend further
		}

		// Skip hidden directories
		if len(name) > 0 && name[0] == '.' {
			continue
		}

		// Recurse
		walkRecursive(fs, filepath.Join(dir, name), depth+1, out, errCh)
	}
}
