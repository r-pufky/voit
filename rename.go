package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// A job is only added if a transformation is required.
func addJob(sourceFileAbsPath string, pattern string, lower bool, strip bool, created bool, modified bool) (*Voit, error) {
	if sourceFileAbsPath == "" {
		return nil, errors.New("no source file provided")
	}
	_, err := os.Stat(sourceFileAbsPath)
	if err != nil {
		return nil, fmt.Errorf("%s source file missing.\n", sourceFileAbsPath)
	}

	absDir, baseName := filepath.Split(sourceFileAbsPath)
	targetName := FormatName(baseName, pattern, lower, strip, created, modified)
	targetAbsPath := filepath.Join(absDir, targetName)

	if baseName != targetName {
		_, err = os.Stat(targetAbsPath)
		if err == nil {
			fmt.Printf("%s already exists, skipping.\n", targetAbsPath)
			return nil, nil
		}

		return &Voit{
			dir:           absDir,
			source:        baseName,
			sourceAbsPath: sourceFileAbsPath,
			target:        targetName,
			targetAbsPath: targetAbsPath,
			width:         len(baseName),
		}, nil
	}

	return nil, nil
}

// Create rename jobs from provided source file and directory. Enumerate files
// if source file is empty. Jobs are only added if a transformation is
// required.
func CreateJobs(sourceFileAbsPath string, sourceDirAbsPath string, pattern string, lower bool, strip bool, created bool, modified bool) ([]Voit, int, error) {
	// Defensive recheck options.
	if sourceFileAbsPath == "" && sourceDirAbsPath == "" {
		return nil, 0, errors.New("No source file or directory specified.")
	}
	if sourceFileAbsPath != "" && sourceDirAbsPath != "" {
		return nil, 0, errors.New("Only specify file or directory.")
	}

	var jobs []Voit
	maxWidth := 0

	if sourceFileAbsPath != "" {
		job, err := addJob(sourceFileAbsPath, pattern, lower, strip, created, modified)
		if err != nil {
			return nil, 0, err
		}
		if job == nil {
			return nil, 0, nil
		}
		return append(jobs, *job), job.width, nil
	}

	if sourceDirAbsPath != "" {
		files, err := os.ReadDir(sourceDirAbsPath)
		if err != nil {
			return nil, 0, err
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}
			fileAbsPath := filepath.Join(sourceDirAbsPath, file.Name())

			job, err := addJob(fileAbsPath, pattern, lower, strip, created, modified)
			if err != nil {
				return nil, 0, err
			}
			if job == nil {
				continue
			}

			jobs = append(jobs, *job)
			if job.width > maxWidth {
				maxWidth = job.width
			}
		}
		return jobs, maxWidth, nil
	}
	return nil, 0, nil
}

func ExecuteRename(jobs []Voit, overwrite bool, verbose bool) {
	for _, job := range jobs {
		target := job.targetAbsPath

		if !overwrite {
			if _, err := os.Stat(target); err == nil {
				if verbose {
					fmt.Printf("Collision: %s", target)
				}
				target = preventCollision(target, verbose)
			}
		}

		if err := os.Rename(job.sourceAbsPath, target); err != nil {
			fmt.Printf("Error renaming %s to %s: %v\n", job.source, target, err)
		} else if verbose {
			fmt.Printf("Renamed: %s -> %s\n", job.source, target)
		}
	}
}

func preventCollision(path string, verbose bool) string {
	ext := filepath.Ext(path)
	name := strings.TrimSuffix(path, ext)
	counter := 1
	uniquePath := path

	for {
		uniquePath = fmt.Sprintf("%s_%d%s", name, counter, ext)
		if _, err := os.Stat(uniquePath); os.IsNotExist(err) {
			if verbose {
				fmt.Printf("Collision (new target): %s", uniquePath)
			}
			break
		}
		counter++
	}
	return uniquePath
}
