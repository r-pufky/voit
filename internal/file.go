// Handle file operations.
package internal

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/r-pufky/voit/voit"
)

// Scan provided file or directory path for files non-recursively. File
// metadata is parsed with no changes to loaded data.
func Scan(f string) ([]voit.File, error) {
	var files []voit.File

	stat, err := os.Stat(f)
	if err != nil {
		return nil, err
	}

	if !stat.IsDir() {
		file, err := new(f)
		if err != nil {
			return nil, err
		}
		if file != nil {
			files = append(files, *file)
		}
		return files, nil
	}

	scan, err := os.ReadDir(f)
	if err != nil {
		return nil, err
	}

	for _, scanFile := range scan {
		info, err := scanFile.Info()
		if err != nil {
			return nil, fmt.Errorf("Unable to get file attributes: %s", scanFile)
		}
		if info.IsDir() {
			continue
		}

		file, err := new(filepath.Join(f, scanFile.Name()))
		if err != nil {
			return nil, err
		}
		if file != nil {
			files = append(files, *file)
		}
	}
	return files, nil
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

	for _, ext := range voit.MultiExts {
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

// Return new File struct based on given source path.
func new(f string) (*voit.File, error) {
	source, err := filepath.Abs(f)
	if err != nil {
		log.Fatalf("Failed to set absolute path: %v", err)
	}

	info, err := os.Stat(source)
	if err != nil {
		log.Fatalf("Failed to stat file: %v", err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("Skipped (directory): %s", source)
	}

	name, ext := SplitMultiExt(source)
	cTime := info.ModTime().UTC()
	mTime := info.ModTime().UTC()

	if runtime.GOOS == "linux" {
		if stat, ok := info.Sys().(*syscall.Stat_t); ok {
			mTime = time.Unix(stat.Ctim.Sec, stat.Ctim.Nsec).UTC()
		}
	}

	return &voit.File{
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
	}, nil
}
