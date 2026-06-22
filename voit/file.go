// File contains underlying filesystem metadata.

package voit

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"golang.org/x/sys/unix"
)

// Common compound extensions.
var CompExts = []string{
	".tar.gz",
	".tar.xz",
	".tar.bz2",
	".min.js",
	".min.css",
	".css.map",
	".js.map",
	".sql.gz",
}

// Filesystem metadata for a single file.
type File struct {
	CTime time.Time // Create time (UTC).
	MTime time.Time // Modified time (UTC).
	Path  string    // Absolute base path: /abs/path.
	Name  string    // Name: {FILE}.
	Ext   string    // Extension: {.EXT}.
}

// Return populated File struct based on path. Directories are not supported.
//
// Path is converted to an absolute path. Stat file for creation and mod times.
// Name and Ext split on compound extensions. Width is calculated from
// Name.Ext.
func NewFile(path string, cfg ...Config) (*File, error) {
	c := NewConfig(cfg...)
	source, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("Failed to resolve absolute path: %w", err)
	}

	info, err := c.FS.Stat(source)
	if err != nil {
		return nil, fmt.Errorf("Failed to stat file: %w", err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("directory: %s", source)
	}

	dir, name, ext := CompExt(source)
	cTime := info.ModTime().UTC() // Default to mod for non-unix systems.
	mTime := info.ModTime().UTC()

	if runtime.GOOS == "linux" {
		var stat unix.Statx_t
		err := c.Unix.Statx(unix.AT_FDCWD, path, unix.AT_STATX_SYNC_AS_STAT, unix.STATX_BTIME, &stat)

		if err != nil {
			return nil, fmt.Errorf("Failed to request birth (creation) time: %w", err)
		}

		// Birth time creation call success and filesystem supports birth time.
		if stat.Mask&unix.STATX_BTIME == 0 {
			return nil, fmt.Errorf("Birth (creation) time not supported by filesystem: %w", err)
		}
		cTime = time.Unix(stat.Btime.Sec, int64(stat.Btime.Nsec)).UTC()
	}

	return &File{
		CTime: time.Date(
			cTime.Year(),
			cTime.Month(),
			cTime.Day(),
			cTime.Hour(),
			cTime.Minute(),
			cTime.Second(),
			cTime.Nanosecond(),
			time.UTC,
		),
		MTime: time.Date(
			mTime.Year(),
			mTime.Month(),
			mTime.Day(),
			mTime.Hour(),
			mTime.Minute(),
			mTime.Second(),
			mTime.Nanosecond(),
			time.UTC,
		),
		Path: dir,
		Name: name,
		Ext:  ext,
	}, nil
}

// Key returns file key consisting of /abs/path/file. Extension is dropped to
// support placing both files and sidecars within the same map.
func (f *File) Key() string {
	return filepath.Join(f.Path, f.Name)
}

// AbsPath returns absolute path /abs/path/file.ext. File metadata assumed to
// be valid.
func (f *File) AbsPath() string {
	return filepath.Join(f.Path, f.Name+f.Ext)
}

// CompExt parses compound extensions from an absolute path into (path, name,
// ext) components. Use in lieu of filepath.Ext().
//
// Returns path, base name, and extension:
//   - directory: dir, "", ""
//   - dotfile: dir, name, ""
//   - standard extension: dir, name, {.Ext}
//   - compound extension: dir, name, {.CompExts}
//   - sidecar: dir, name, {.{EXT}.xmp}
func CompExt(file string) (string, string, string) {
	if file == "." || file == "/" {
		return file, "", "" // Directory.
	}
	dir := filepath.Dir(file)
	base := filepath.Base(file)

	if dot := strings.LastIndex(base, "."); dot == -1 || dot == 0 {
		return dir, base, "" // File is dotfile or has no extension.
	}

	f := strings.ToLower(base)

	for _, ext := range CompExts {
		if strings.HasSuffix(f, ext) {
			// Standard muli-extension. Split on calculated length to handle unknown
			// source casing and multi-byte unicode characters instead of TrimSuffix.
			return dir, base[:len(base)-len(ext)], base[len(base)-len(ext):]
		}
	}

	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)

	// If .{EXT}.xmp reparse to compound sidecar extension.
	if strings.ToLower(ext) == ".xmp" && strings.Count(base, ".") >= 2 {
		sidecar := filepath.Ext(name)
		return dir, strings.TrimSuffix(name, sidecar), sidecar + ext
	}

	return dir, name, ext
}

// IsSidecar returns true if file is a sidecar file. A file is considered a
// sidecar when the extension matches the format .{EXT}.xmp (case-insensitive).
func (f *File) IsSidecar() bool {
	base := strings.ToLower(f.Ext)
	return strings.HasSuffix(base, ".xmp") && strings.Count(base, ".") >= 2
}
