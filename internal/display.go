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

func DisplayPending(out io.Writer, files []voit.File) int {
	if len(files) == 0 {
		return 0
	}
	c := 0

	maxJob := slices.MaxFunc(files, func(a, b voit.File) int {
		return cmp.Compare(a.Width, b.Width)
	})

	for _, file := range files {
		if file.Matched {
			fmt.Fprintf(out, "%-*s ➔ %s\n", int(maxJob.Width), filepath.Base(file.Source), filepath.Base(file.Target))
			c++
		}
	}
	return c
}

func Confirm(in io.Reader, out io.Writer) bool {
	fmt.Fprint(out, "Proceed? (y/n): ")
	var input string
	fmt.Fscanln(in, &input)

	input = strings.ToLower(strings.TrimSpace(input))
	return input == "y" || input == "yes"
}
