package filesystem

import (
	"os"

	"github.com/samber/mo"
)

type FileSystem interface {
	ReadDir(path string) mo.Result[[]os.DirEntry]
	Chdir(path string) mo.Result[string]
}

type RealFileSystem struct{}

func (fs *RealFileSystem) ReadDir(path string) mo.Result[[]os.DirEntry] {
	dirEntry, err := os.ReadDir(path)
	if err != nil {
		return mo.Err[[]os.DirEntry](err)
	}
	return mo.Ok(dirEntry)
}

func (fs *RealFileSystem) Chdir(path string) mo.Result[string] {
	if err := os.Chdir(path); err != nil {
		return mo.Err[string](err)
	}
	return mo.Ok(path)
}
