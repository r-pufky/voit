package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Voit struct {
	dir           string // Absolute path to file directory.
	source        string // source file basename.
	sourceAbsPath string // source file absolute path.
	target        string // target file basename.
	targetAbsPath string // target file absolute path.
	width         int    // width of source file basename.
}

func main() {
	home, _ := os.UserHomeDir()

	opts := Config(filepath.Join(home, ".config", "voit.toml"))

	jobs, width, _ := CreateJobs(opts.File, opts.Directory, opts.Pattern, opts.Lower, opts.Strip)

	if len(jobs) == 0 {
		fmt.Println("No files matched the known datetime formats.")
		return
	}

	DisplayPending(os.Stdout, jobs, width)
	fmt.Printf("\nProposed changes: %d file(s).\n", len(jobs))

	if !opts.Yes && !Confirm(os.Stdin, os.Stdout) {
		fmt.Println("Operation aborted by user.")
		return
	}

	ExecuteRename(jobs)
	fmt.Println("Success.")
}

func DisplayPending(out io.Writer, jobs []Voit, width int) {
	for _, job := range jobs {
		fmt.Fprintf(out, "%-*s ➔ %s\n", width, job.source, job.target)
	}
}

func Confirm(in io.Reader, out io.Writer) bool {
	fmt.Fprint(out, "Proceed? (y/n): ")
	var input string
	fmt.Fscanln(in, &input)

	input = strings.ToLower(strings.TrimSpace(input))
	return input == "y" || input == "yes"
}
