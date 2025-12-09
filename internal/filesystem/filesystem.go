package filesystem

import "os"

type FileSystem interface {
	ReadDir(path string) ([]os.DirEntry, error)
}

type RealFileSystem struct{}

func (fs *RealFileSystem) ReadDir(path string) ([]os.DirEntry, error) {
	return os.ReadDir(path)
}
