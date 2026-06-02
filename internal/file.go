// Handle file operations.
package internal

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/r-pufky/voit/voit"
)

// Scan provided file or directory path for files non-recursively.
func Scan(f string) ([]*voit.Voit, error) {
	var files []*voit.Voit

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
			files = append(files, file)
		}
		return files, nil
	}

	scan, err := os.ReadDir(f)
	if err != nil {
		return nil, err
	}

	for _, dFile := range scan {
		info, err := dFile.Info()
		if err != nil {
			return nil, fmt.Errorf("unable to get file attributes: %s", dFile)
		}
		if info.IsDir() {
			continue
		}

		file, err := new(filepath.Join(f, dFile.Name()))
		if err != nil {
			return nil, err
		}
		if file != nil {
			files = append(files, file)
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

// Return new voit.Voit struct based on given source path.
func new(f string) (*voit.Voit, error) {
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
	cTime := info.ModTime().UTC()
	mTime := info.ModTime().UTC()

	if runtime.GOOS == "linux" {
		if stat, ok := info.Sys().(*syscall.Stat_t); ok {
			mTime = time.Unix(stat.Ctim.Sec, stat.Ctim.Nsec).UTC()
		}
	}

	return &voit.Voit{
		File: voit.File{
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

// Rename files marked as Matched using File.Source and Mark.Target resolving
// collisions unless overwrite is enabled.
func Rename(w io.Writer, files []*voit.Voit, overwrite bool, verbose bool) {
	defer timeAction(w, time.Now(), len(files))
	for _, f := range files {
		if f.Matched {
			target := f.Target

			if !overwrite {
				if _, err := os.Stat(target); err == nil {
					if verbose {
						fmt.Fprintf(w, "Collision: %s", target)
					}
					target = resolveFSCollisions(w, target)
				}
			}

			if err := os.Rename(f.File.Source, target); err != nil {
				fmt.Fprintf(w, "Error renaming %s to %s: %v\n", f.File.Source, target, err)
			} else if verbose {
				fmt.Fprintf(w, "Renamed: %s%s ➔ %s%s\n", f.File.Source, f.File.Ext, target, f.File.Ext)
			}
		}
	}
}

// Resolve FS collision during file disk operation.
func resolveFSCollisions(w io.Writer, path string) string {
	ext := filepath.Ext(path)
	name := strings.TrimSuffix(path, ext)
	counter := 1
	uniquePath := path

	for {
		uniquePath = fmt.Sprintf("%s_%d%s", name, counter, ext)
		if _, err := os.Stat(uniquePath); os.IsNotExist(err) {
			fmt.Fprintf(w, "Collision (new target): %s", uniquePath)
			break
		}
		counter++
	}
	return uniquePath
}

// Time file actions. Defer timeAction(w, time.Now(), len(files))
func timeAction(w io.Writer, start time.Time, count int) {
	elapsed := time.Since(start)
	fmt.Fprintf(w, "Renamed %d files in %s.\n", count, elapsed)
}
