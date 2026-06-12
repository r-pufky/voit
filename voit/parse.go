package voit

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

// Scan provided file or directory path for files non-recursively.
func Scan(f string) (VoitFiles, error) {
	var files VoitFiles

	matches, err := filepath.Glob(f)
	if err != nil {
		return nil, fmt.Errorf("invalid glob pattern: %w\n", err)
	}

	// Bare directories require globbing: /tmp ➔ /tmp/*.
	if len(matches) == 1 {
		stat, err := os.Stat(matches[0])
		if err == nil && stat.IsDir() {
			matches, err = filepath.Glob(filepath.Join(matches[0], "*"))
			if err != nil {
				return nil, err
			}
		}
	}

	for _, path := range matches {
		stat, err := os.Stat(path)
		if err != nil {
			return nil, err
		}

		if stat.IsDir() {
			continue
		}

		file, err := New(path)
		if err != nil {
			return nil, err
		}
		if file != nil {
			files = append(files, file)
		}
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("No files matched the known datetime formats (is globbing quoted?).")
	}

	return files, nil
}

// Return new Voit struct based on given source path. XMP sidecar files
// (.{EXT}.xmp) are automatically detected if .xmp exists with two dots in the
// name.
func New(f string) (*Voit, error) {
	source, err := filepath.Abs(f)
	if err != nil {
		log.Fatalf("Failed to set absolute path: %v", err)
	}

	info, err := os.Stat(source)
	if err != nil {
		log.Fatalf("Failed to stat file: %v", err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("skipped (directory): %s", source)
	}

	name, ext := SplitMultiExt(source)
	// Reparse extension to .{EXT}.xmp if file is sidecar.
	if strings.ToLower(ext) == ".xmp" && strings.Count(filepath.Base(source), ".") >= 2 {
		sideCar := filepath.Ext(name)
		name = strings.TrimSuffix(name, sideCar)
		ext = sideCar + ext
	}
	cTime := info.ModTime().UTC()
	mTime := info.ModTime().UTC()

	if runtime.GOOS == "linux" {
		if stat, ok := info.Sys().(*syscall.Stat_t); ok {
			mTime = time.Unix(stat.Ctim.Sec, stat.Ctim.Nsec).UTC()
		}
	}

	return &Voit{
		File: File{
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
			Source: source,
			Name:   name,
			Ext:    ext,
			Width:  uint8(len(name) + len(ext)),
		},
	}, nil
}

// Split file to (name, ext) handling cases with common multi-part extensions.
// * Preceding paths are automatically stripped.
// * Files with no extensions are returned with empty extension string.
// * Dotfiles are treated as filenames (with no extensions).
//
// Examples:
//
//	test.tar.gz ➔ test .tar.gz.
//	/path/.gitignore ➔ .gitignore ""
//	/path/test.zip ➔ test .zip
func SplitMultiExt(f string) (string, string) {
	name := filepath.Base(f)

	if name == "." || name == "/" {
		return name, "" // Directory.
	}

	for _, ext := range MultiExts {
		if strings.HasSuffix(name, ext) {
			name := strings.TrimSuffix(name, ext)
			return name, ext
		}
	}

	if dot := strings.LastIndex(name, "."); dot != -1 && dot != 0 {
		ext := filepath.Ext(name)
		return strings.TrimSuffix(name, ext), ext
	}

	return name, ""
}
