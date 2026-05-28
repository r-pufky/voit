package internal

import (
	"cmp"
	"fmt"
	"io"
	"path/filepath"
	"slices"
	"strings"

	"github.com/r-pufky/voit/voit"
)

// Display source and target file name changes in column. Returns total matched
// file count.
func DisplayPending(out io.Writer, files []*voit.Voit, c *voit.Config) int {
	if len(files) == 0 {
		return 0
	}
	i := 0

	maxJob := slices.MaxFunc(files, func(a, b *voit.Voit) int {
		return cmp.Compare(a.File.Width, b.File.Width)
	})

	for _, file := range files {
		if file.Matched {
			fmt.Fprintf(out, "%-*s ➔ %s\n", int(maxJob.File.Width), filepath.Base(file.File.Source), filepath.Base(file.Target))
			i++
		}
	}
	return i
}

// Display confirmation dialog and wait for user input.
func Confirm(in io.Reader, out io.Writer) bool {
	fmt.Fprint(out, "Proceed? (y/n): ")
	var input string
	fmt.Fscanln(in, &input)

	input = strings.ToLower(strings.TrimSpace(input))
	return input == "y" || input == "yes"
}
