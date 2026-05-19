package internal

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/r-pufky/voit/models"
)

func Rename(opts models.Opts) {
	files, err := Scan(opts.AbsSource)
	if err != nil {
		log.Fatalf("Unable to complete source file scan: %v", err)
	}

	if len(files) == 0 {
		fmt.Println("No files matched the known datetime formats.")
		os.Exit(0)
	}

	for i := range files {
		Parse(&files[i], opts.Rename.Pattern, opts.Rename.Force, opts.DescSep, opts.TagSep)
		GenTargetName(&files[i], opts.Rename.Pattern, opts.Rename.Lower, opts.Rename.Strip, opts.Rename.NoDesc, opts.DescSep, opts.TagSep)
	}

	count := DisplayPending(os.Stdout, files)
	if opts.Rename.Overwrite {
		fmt.Printf("\nProposed changes (OVERWRITE ENABLED): %d file(s).\n", count)
	} else {
		fmt.Printf("\nProposed changes: %d file(s).\n", count)
	}

	if !opts.Yes && !Confirm(os.Stdin, os.Stdout) {
		fmt.Println("Operation aborted by user.")
		os.Exit(0)
	}

	ExecuteRename(os.Stdout, files, opts.Rename.Overwrite, opts.Verbose)
}

func ExecuteRename(w io.Writer, files []models.File, overwrite bool, verbose bool) {
	defer timeRename(w, time.Now(), len(files))
	for _, file := range files {
		if file.Matched {
			target := file.Target

			if !overwrite {
				if _, err := os.Stat(target); err == nil {
					if verbose {
						fmt.Fprintf(w, "Collision: %s", target)
					}
					target = resolveFSCollisions(w, target, verbose)
				}
			}

			if err := os.Rename(file.Source, target); err != nil {
				fmt.Fprintf(w, "Error renaming %s to %s: %v\n", file.Source, target, err)
			} else if verbose {
				fmt.Fprintf(w, "Renamed: %s%s ➔ %s%s\n", file.Source, file.Ext, target, file.Ext)
			}
		}
	}
}

func resolveFSCollisions(w io.Writer, path string, verbose bool) string {
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

func timeRename(w io.Writer, start time.Time, count int) {
	elapsed := time.Since(start)
	fmt.Fprintf(w, "Renamed %d files in %s.", count, elapsed)
}
