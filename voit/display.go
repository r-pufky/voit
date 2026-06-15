package voit

import (
	"cmp"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// Display source and target file name changes in column. Returns total matched
// file count.
func (files VoitFiles) DisplayPending(w io.Writer) int {
	if len(files) == 0 {
		return 0
	}
	i := 0

	maxJob := slices.MaxFunc(files, func(a, b *Voit) int {
		return cmp.Compare(a.File.Width, b.File.Width)
	})

	for _, file := range files {
		if file.Matched {
			fmt.Fprintf(w, "%-*s ➔ %s\n", int(maxJob.File.Width), filepath.Base(file.File.Source), filepath.Base(file.Target))
			i++
		}
	}
	return i
}

// Display confirmation dialog and wait for user input.
func Confirm(w io.Writer, r io.Reader) bool {
	fmt.Fprint(w, "Proceed? (y/n): ")
	var input string
	fmt.Fscanln(r, &input)

	input = strings.ToLower(strings.TrimSpace(input))
	return input == "y" || input == "yes"
}

// Prompt and rename on matched files after resolving collisions.
func (files VoitFiles) PromptRename(w io.Writer, r io.Reader, opts *Opts) {
	files.ResolveCollisions(opts.Voit(), opts.Verbose)
	count := files.DisplayPending(w)
	if count != 0 {
		if opts.Overwrite {
			fmt.Fprintf(w, "\nProposed changes (OVERWRITE ENABLED): %d file(s).\n", count)
		} else {
			fmt.Fprintf(w, "\nProposed changes: %d file(s).\n", count)
		}

		if !opts.Yes && !Confirm(w, r) {
			fmt.Fprintln(w, "Operation aborted by user.")
			os.Exit(0)
		}

		files.Rename(w, opts.Overwrite, opts.Verbose)
	} else {
		fmt.Fprintln(w, "No files matched proposed changes.")
	}
}
