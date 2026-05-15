package internal

import (
	"fmt"
	"io"
	"strings"

	"github.com/r-pufky/voit/models"
)

func DisplayPending(out io.Writer, jobs []models.Job, width int) {
	for _, job := range jobs {
		fmt.Fprintf(out, "%-*s ➔ %s\n", width, job.Source, job.Target)
	}
}

func Confirm(in io.Reader, out io.Writer) bool {
	fmt.Fprint(out, "Proceed? (y/n): ")
	var input string
	fmt.Fscanln(in, &input)

	input = strings.ToLower(strings.TrimSpace(input))
	return input == "y" || input == "yes"
}
