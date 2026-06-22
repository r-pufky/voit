package voit

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

type mockFileInfo struct {
	isDir   bool
	modTime time.Time
}

func (m mockFileInfo) Name() string       { return "mock" }
func (m mockFileInfo) Size() int64        { return 100 }
func (m mockFileInfo) Mode() os.FileMode  { return 0644 }
func (m mockFileInfo) ModTime() time.Time { return m.modTime }
func (m mockFileInfo) IsDir() bool        { return m.isDir }
func (m mockFileInfo) Sys() any           { return nil }

func TestNewFile(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()

	tests := []struct {
		name      string
		path      string
		mockStat  func(name string) (os.FileInfo, error)
		mockStatx func(dirfd int, path string, flags int, mask int, stat *unix.Statx_t) error
		wantErr   bool
		checkFile func(t *testing.T, f *File)
	}{
		{
			name:     "file found [file struct returned]",
			path:     "photo.jpg",
			mockStat: func(name string) (os.FileInfo, error) { return mockFileInfo{isDir: false, modTime: now}, nil },
			mockStatx: func(dirfd int, path string, flags int, mask int, stat *unix.Statx_t) error {
				stat.Mask = unix.STATX_BTIME
				stat.Btime.Sec = now.Unix()
				stat.Btime.Nsec = uint32(now.Nanosecond())
				return nil
			},
			wantErr: false,
			checkFile: func(t *testing.T, f *File) {
				if f.MTime.Location() != time.UTC || f.CTime.Location() != time.UTC {
					t.Errorf("\nExpect UTC timestamps\nMTime: %v\nCTime: %v\n", f.MTime.Location(), f.CTime.Location())
				}
			},
		},
		{
			name: "file is missing [error]",
			path: "missing.txt",
			mockStat: func(name string) (os.FileInfo, error) {
				return nil, errors.New("file not found")
			},
			wantErr: true,
		},
		{
			name: "path is directory [error]",
			path: "/path/is/directory",
			mockStat: func(name string) (os.FileInfo, error) {
				return mockFileInfo{isDir: true, modTime: now}, nil
			},
			wantErr: true,
		},
		{
			name: "statx birth time failed [error]",
			path: "photo.jpg",
			mockStat: func(name string) (os.FileInfo, error) {
				return mockFileInfo{isDir: false, modTime: now}, nil
			},
			mockStatx: func(dirfd int, path string, flags int, mask int, stat *unix.Statx_t) error {
				return errors.New("operation not permitted")
			},
			wantErr: runtime.GOOS == "linux",
		},
		{
			name: "birth time not supported [error]",
			path: "photo.jpg",
			mockStat: func(name string) (os.FileInfo, error) {
				return mockFileInfo{isDir: false, modTime: now}, nil
			},
			mockStatx: func(dirfd int, path string, flags int, mask int, stat *unix.Statx_t) error {
				stat.Mask = 0
				return nil
			},
			wantErr: runtime.GOOS == "linux",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := Config{}
			cfg.FS = MockFS{
				MockStat: tt.mockStat,
			}

			if tt.mockStatx != nil {
				cfg.Unix = MockUnix{MockStatx: tt.mockStatx}
			} else {
				// Default for cases where mockStatx isn't explicitly set.
				cfg.Unix = MockUnix{
					MockStatx: func(dirfd int, path string, flags int, mask int, stat *unix.Statx_t) error {
						stat.Mask = unix.STATX_BTIME
						return nil
					},
				}
			}

			file, err := NewFile(tt.path, cfg)

			if (err != nil) != tt.wantErr {
				t.Fatalf("\nGot:     %v\nWantErr: %v\n", err, tt.wantErr)
			}

			if !tt.wantErr && tt.checkFile != nil {
				tt.checkFile(t, file)
			}
		})
	}
}

func TestFileKey(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		fName   string
		wantKey string
	}{
		{
			name:    "standard file",
			path:    "/abs/path",
			fName:   "file",
			wantKey: "/abs/path/file",
		},
		{
			name:    "trailing slash correctly parsed",
			path:    "/abs/path/",
			fName:   "file",
			wantKey: "/abs/path/file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &File{
				Path: tt.path,
				Name: tt.fName,
			}

			if got := f.Key(); got != tt.wantKey {
				t.Errorf("\nGot:  %q\nWant: %q\n", got, tt.wantKey)
			}
		})
	}
}

func TestFileAbsPath(t *testing.T) {
	tests := []struct {
		name        string
		file        *File
		wantAbsPath string
	}{
		{
			name: "standard file source generated",
			file: &File{
				Path: "uploads/images",
				Name: "avatar",
				Ext:  ".png",
			},
			wantAbsPath: filepath.Join("uploads/images", "avatar.png"),
		},
		{
			name: "files with no extensions are generated",
			file: &File{
				Path: "logs",
				Name: "app",
				Ext:  "",
			},
			wantAbsPath: filepath.Join("logs", "app"),
		},
		{
			name: "missing path [. prepended]",
			file: &File{
				Path: "",
				Name: "avatar",
				Ext:  ".png",
			},
			wantAbsPath: filepath.Join("", "avatar.png"),
		},
		{
			name: "missing name [no filename]",
			file: &File{
				Path: "uploads/images",
				Name: "",
				Ext:  ".png",
			},
			wantAbsPath: filepath.Join("uploads/images", ".png"),
		},
		{
			name: "invalid metadata [no validity checks]",
			file: &File{
				Path: "",
				Name: "",
				Ext:  ".png",
			},
			wantAbsPath: ".png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.file.AbsPath()

			if got != tt.wantAbsPath {
				t.Errorf("\nGot:  %q\nWant: %q\n", got, tt.wantAbsPath)
			}
		})
	}
}

func TestCompExt(t *testing.T) {
	tests := []struct {
		name     string
		file     string
		wantDir  string
		wantBase string
		wantExt  string
	}{
		{
			name:     "current directory [dir returned]",
			file:     ".",
			wantDir:  ".",
			wantBase: "",
			wantExt:  "",
		},
		{
			name:     "root directory [dir returned]",
			file:     "/",
			wantDir:  "/",
			wantBase: "",
			wantExt:  "",
		},
		{
			name:     "file with no extensions [file returned]",
			file:     "README",
			wantDir:  ".",
			wantBase: "README",
			wantExt:  "",
		},
		{
			name:     "dotfile [file returned]",
			file:     ".gitignore",
			wantDir:  ".",
			wantBase: ".gitignore",
			wantExt:  "",
		},
		{
			name:     "compound extension [.tar.gz]",
			file:     "archive.tar.gz",
			wantDir:  ".",
			wantBase: "archive",
			wantExt:  ".tar.gz",
		},
		{
			name:     "compound extension case-insensitivity [.TAR.GZ]",
			file:     "SOURCE.TAR.GZ",
			wantDir:  ".",
			wantBase: "SOURCE",
			wantExt:  ".TAR.GZ",
		},
		{
			name:     "compound extension with long path [full parse]",
			file:     "/home/user/app.sql.gz",
			wantDir:  "/home/user",
			wantBase: "app",
			wantExt:  ".sql.gz",
		},
		{
			name:     "standard extension [.jpg]",
			file:     "photo.jpg",
			wantDir:  ".",
			wantBase: "photo",
			wantExt:  ".jpg",
		},
		{
			name:     "multi-dot filename with standard extension [.png]",
			file:     "my.awesome.photo.png",
			wantDir:  ".",
			wantBase: "my.awesome.photo",
			wantExt:  ".png",
		},
		{
			name:     "sidecar compound extension [.jpg.xmp]",
			file:     "DSC_0001.jpg.xmp",
			wantDir:  ".",
			wantBase: "DSC_0001",
			wantExt:  ".jpg.xmp",
		},
		{
			name:     "sidecar compound extension case-insensitivity [.CR2.XMP]",
			file:     "image.CR2.XMP",
			wantDir:  ".",
			wantBase: "image",
			wantExt:  ".CR2.XMP",
		},
		{
			name:     "naked .xmp file [not detected as sidecar]",
			file:     "doc.xmp",
			wantDir:  ".",
			wantBase: "doc",
			wantExt:  ".xmp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, base, ext := CompExt(tt.file)
			if dir != tt.wantDir || base != tt.wantBase || ext != tt.wantExt {
				t.Errorf("\nGot:  (%q, %q, %q)\nWant: (%q, %q, %q)\n", dir, base, ext, tt.wantDir, tt.wantBase, tt.wantExt)
			}
		})
	}
}

func TestIsSidecar(t *testing.T) {
	tests := []struct {
		name string
		ext  string
		want bool
	}{
		{
			name: "standard sidecar file",
			ext:  ".jpg.xmp",
			want: true,
		},
		{
			name: "alternative sidecar file",
			ext:  ".png.xmp",
			want: true,
		},
		{
			name: "case insensitivity",
			ext:  ".jpg.XMP",
			want: true,
		},
		{
			name: "standard .xmp file",
			ext:  ".xmp",
			want: false,
		},
		{
			name: "wrong extension correct number of dots",
			ext:  ".jpg.txt",
			want: false,
		},
		{
			name: "no dots",
			ext:  "photo_jpg_xmp",
			want: false,
		},
		{
			name: "incorrect delimiters",
			ext:  "photo.jpg_xmp",
			want: false,
		},
		{
			name: "empty string input",
			ext:  "",
			want: false,
		},
		{
			name: "just a dot directory path reference",
			ext:  ".",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &File{Ext: tt.ext}

			if got := f.IsSidecar(); got != tt.want {
				t.Errorf("\nIsSidecar(%q)\nGot:  %v\nWant: %v\n", tt.ext, got, tt.want)
			}
		})
	}
}
