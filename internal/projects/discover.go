package projects

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"dev/internal/filesystem"

	"github.com/samber/lo"
	"github.com/samber/mo"
)

type Project struct {
	Name string
	Path string
}

func Discover(fs filesystem.FileSystem, args []string) mo.Result[[]Project] {
	searchPaths, err := expandPaths(fs, resolvePaths(args)).Get()
	if err != nil {
		return mo.Err[[]Project](err)
	}

	if len(searchPaths) == 0 {
		return mo.Ok([]Project{})
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

	if len(result) == 0 && len(errs) > 0 {
		return mo.Err[[]Project](errs[0])
	}

	return mo.Ok(result)
}

func resolvePaths(args []string) []string {
	if len(args) > 0 {
		return args
	}

	devPaths := strings.Fields(os.Getenv("DEV_PATHS"))
	if len(devPaths) > 0 {
		return devPaths
	}

	homeResult := mo.TupleToResult(os.UserHomeDir())
	if homeResult.IsOk() {
		return []string{homeResult.MustGet()}
	}
	return []string{}
}

func expandPaths(fs filesystem.FileSystem, searchPaths []string) mo.Result[[]string] {
	if len(searchPaths) == 0 {
		return mo.Ok([]string{})
	}

	results := lo.Map(searchPaths, func(p string, _ int) mo.Result[[]string] {
		return expandPath(fs, p)
	})

	paths := lo.FlatMap(results, func(r mo.Result[[]string], _ int) []string {
		return r.OrElse([]string{})
	})

	errors := lo.FilterMap(results, func(r mo.Result[[]string], _ int) (error, bool) {
		if r.IsError() {
			return r.Error(), true
		}
		return nil, false
	})

	if len(paths) == 0 && len(errors) > 0 {
		return mo.Err[[]string](errors[0])
	}

	return mo.Ok(paths)
}

func expandPath(fs filesystem.FileSystem, searchPath string) mo.Result[[]string] {
	entries, err := fs.ReadDir(searchPath).Get()
	if err != nil {
		return mo.Err[[]string](err)
	}

	paths := lo.FilterMap(entries, func(entry os.DirEntry, _ int) (string, bool) {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			return filepath.Join(searchPath, entry.Name()), true
		}
		return "", false
	})

	return mo.Ok(paths)
}

func walkRecursive(fs filesystem.FileSystem, dir string, depth int, out chan<- Project, errCh chan<- error) {
	if depth > 2 {
		return
	}

	entries, err := fs.ReadDir(dir).Get()
	if err != nil {
		errCh <- err
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()

		if name == ".git" {
			out <- Project{
				Name: filepath.Base(dir),
				Path: dir,
			}
			return
		}

		if len(name) > 0 && name[0] == '.' {
			continue
		}

		walkRecursive(fs, filepath.Join(dir, name), depth+1, out, errCh)
	}
}
