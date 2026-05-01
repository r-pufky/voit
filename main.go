package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type renameJob struct {
	oldPath string
	newPath string
	oldName string
	newName string
	offset  int // Original name datetime start offset.
}

// Create job for a single file.
func JobsFile(opts Options) ([]renameJob, int, error) {
	absPath, err := filepath.Abs(opts.File)
	oldName := filepath.Base(absPath)
	newName, offset, matched := ParseNewName(oldName, opts)

	if err != nil {
		return nil, 0, err
	}

	if !matched || oldName == newName {
		return nil, 0, nil
	}

	return []renameJob{{
		oldName: oldName,
		newName: newName,
		oldPath: absPath,
		newPath: filepath.Join(filepath.Dir(absPath), newName),
		offset:  offset,
	}}, len(oldName), nil
}

// Create jobs for a directory.
func JobsDirectory(opts Options) ([]renameJob, int, error) {
	files, err := os.ReadDir(opts.Directory)
	if err != nil {
		return nil, 0, err
	}

	var jobs []renameJob
	maxOldLen := 0
	counts := make(map[string]int)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		oldName := file.Name()
		newName, offset, matched := ParseNewName(oldName, opts)

		if matched {
			finalNewName, finalPath := ResolveCollisions(opts.Directory, newName, counts)

			if oldName != finalNewName {
				jobs = append(jobs, renameJob{
					oldName: oldName,
					newName: finalNewName,
					oldPath: filepath.Join(opts.Directory, oldName),
					newPath: finalPath,
					offset:  offset,
				})
				if len(oldName) > maxOldLen {
					maxOldLen = len(oldName)
				}
			}
		}
	}
	return jobs, maxOldLen, nil
}

func DisplayPending(jobs []renameJob, maxLen int, vertical bool) {
	for _, job := range jobs {
		if vertical {
			fmt.Printf("%s\n%*s%s\n", job.oldName, job.offset, "", job.newName)
		} else {
			fmt.Printf("%-*s ➔ %s\n", maxLen, job.oldName, job.newName)
		}
	}
}

func Confirm() bool {
	fmt.Print("Proceed? (y/n): ")
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		input := strings.ToLower(strings.TrimSpace(scanner.Text()))
		return input == "y" || input == "yes"
	}
	return false
}

func RenameJobs(jobs []renameJob) {
	for _, job := range jobs {
		if err := os.Rename(job.oldPath, job.newPath); err != nil {
			fmt.Printf("Error renaming %s: %v\n", job.oldName, err)
		}
	}
}

func main() {
	opts := ParseFlags()
	var jobs []renameJob
	var maxOldLen int
	var err error

	if opts.File != "" {
		jobs, maxOldLen, err = JobsFile(opts)
	} else {
		jobs, maxOldLen, err = JobsDirectory(opts)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "[error]  %v\n", err)
		os.Exit(3)
	}

	if len(jobs) == 0 {
		fmt.Println("No files matched the known datetime formats.")
		return
	}

	DisplayPending(jobs, maxOldLen, opts.Vertical)
	fmt.Printf("\nProposed changes: %d file(s).\n", len(jobs))

	if !opts.Yes && !Confirm() {
		fmt.Println("Operation aborted by user.")
		return
	}

	RenameJobs(jobs)
	fmt.Println("Success.")
}
