package voit

import (
	"io"
	"os"

	"golang.org/x/sys/unix"
)

// FileSystem mocks for testing.
type MockFS struct {
	MockStat   func(string) (os.FileInfo, error)
	MockRename func(string, string) error
	MockOpen   func(string) (io.ReadCloser, error)
}

func (m MockFS) Stat(name string) (os.FileInfo, error) { return m.MockStat(name) }
func (m MockFS) Rename(oldpath, newpath string) error  { return m.MockRename(oldpath, newpath) }
func (m MockFS) Open(name string) (io.ReadCloser, error) {
	if m.MockOpen != nil {
		return m.MockOpen(name)
	}
	return nil, os.ErrNotExist
}

// MockUnix mocks testing.
type MockUnix struct {
	MockStatx func(dirfd int, path string, flags int, mask int, stat *unix.Statx_t) error
}

func (m MockUnix) Statx(dirfd int, path string, flags int, mask int, stat *unix.Statx_t) error {
	return m.MockStatx(dirfd, path, flags, mask, stat)
}
