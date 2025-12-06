package testutil

import (
	"io/fs"
	"time"
)

type MockDirEntry struct {
	EntryName  string
	EntryIsDir bool
	EntryType  fs.FileMode
}

func (m *MockDirEntry) Name() string {
	return m.EntryName
}

func (m *MockDirEntry) IsDir() bool {
	return m.EntryIsDir
}

func (m *MockDirEntry) Type() fs.FileMode {
	return m.EntryType
}

func (m *MockDirEntry) Info() (fs.FileInfo, error) {
	return &MockFileInfo{
		FileName:  m.EntryName,
		FileIsDir: m.EntryIsDir,
		FileMode:  m.EntryType,
	}, nil
}

type MockFileInfo struct {
	FileName  string
	FileIsDir bool
	FileMode  fs.FileMode
}

func (m *MockFileInfo) Name() string       { return m.FileName }
func (m *MockFileInfo) Size() int64        { return 0 }
func (m *MockFileInfo) Mode() fs.FileMode  { return m.FileMode }
func (m *MockFileInfo) ModTime() time.Time { return time.Time{} }
func (m *MockFileInfo) IsDir() bool        { return m.FileIsDir }
func (m *MockFileInfo) Sys() any           { return nil }

func NewMockDir(name string) *MockDirEntry {
	return &MockDirEntry{
		EntryName:  name,
		EntryIsDir: true,
		EntryType:  fs.ModeDir,
	}
}

func NewMockFile(name string) *MockDirEntry {
	return &MockDirEntry{
		EntryName:  name,
		EntryIsDir: false,
		EntryType:  0,
	}
}
