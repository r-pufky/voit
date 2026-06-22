// Interfaces for underlying packages to enable mock unit testing.

package voit

import (
	"io"
	"os"

	"golang.org/x/sys/unix"
)

// FileSystem interface to call os.* calls or use mocks.
type FileSystem interface {
	Rename(oldpath, newpath string) error
	Stat(name string) (os.FileInfo, error)
	Open(name string) (io.ReadCloser, error)
}

type RealFS struct{}

func (RealFS) Rename(oldpath, newpath string) error    { return os.Rename(oldpath, newpath) }
func (RealFS) Stat(name string) (os.FileInfo, error)   { return os.Stat(name) }
func (RealFS) Open(name string) (io.ReadCloser, error) { return os.Open(name) }

// Unix interface to call unix.* calls or use mocks.
type Unix interface {
	Statx(dirfd int, path string, flags int, mask int, stat *unix.Statx_t) error
}

type RealUnix struct{}

func (RealUnix) Statx(dirfd int, path string, flags int, mask int, stat *unix.Statx_t) error {
	return unix.Statx(dirfd, path, flags, mask, stat)
}
