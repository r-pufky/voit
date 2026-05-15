package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/r-pufky/voit/internal"
)

var Version = "development"

func main() {
	home, _ := os.UserHomeDir()
	opts := Config(os.Stdout, filepath.Join(home, ".config", "voit.toml"))

	jobs, width, _ := internal.CreateJobs(opts.File, opts.Directory, opts.Pattern, opts.Lower, opts.Strip, opts.Created, opts.Modified)

	if len(jobs) == 0 {
		fmt.Println("No files matched the known datetime formats.")
		return
	}

	internal.DisplayPending(os.Stdout, jobs, width)
	if opts.Overwrite {
		fmt.Printf("\nProposed changes (OVERWRITE ENABLED): %d file(s).\n", len(jobs))
	} else {
		fmt.Printf("\nProposed changes: %d file(s).\n", len(jobs))
	}

	if !opts.Yes && !internal.Confirm(os.Stdin, os.Stdout) {
		fmt.Println("Operation aborted by user.")
		return
	}

	internal.ExecuteRename(os.Stdout, jobs, opts.Overwrite, opts.Verbose)
}
