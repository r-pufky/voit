package internal

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/r-pufky/voit/models"
)

// A job is only added if a transformation is required.
func addJob(sourceFileAbsPath string, pattern string, lower bool, strip bool, created bool, modified bool) (*models.Job, error) {
	if sourceFileAbsPath == "" {
		return nil, errors.New("no source file provided.")
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

		return &models.Job{
			Dir:           absDir,
			Source:        baseName,
			SourceAbsPath: sourceFileAbsPath,
			Target:        targetName,
			TargetAbsPath: targetAbsPath,
			Width:         len(baseName),
		}, nil
	}

	return nil, nil
}

// Create rename jobs from provided source file and directory. Enumerate files
// if source file is empty. Jobs are only added if a transformation is
// required.
func CreateJobs(sourceFileAbsPath string, sourceDirAbsPath string, pattern string, lower bool, strip bool, created bool, modified bool) ([]models.Job, int, error) {
	var jobs []models.Job
	maxWidth := 0

	if sourceFileAbsPath != "" {
		job, err := addJob(sourceFileAbsPath, pattern, lower, strip, created, modified)
		if err != nil {
			return nil, 0, err
		}
		if job == nil {
			return nil, 0, nil
		}
		return append(jobs, *job), job.Width, nil
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
			if job.Width > maxWidth {
				maxWidth = job.Width
			}
		}
		return resolveJobCollisions(jobs), maxWidth, nil
	}
	return nil, 0, nil
}

func ExecuteRename(w io.Writer, jobs []models.Job, overwrite bool, verbose bool) {
	defer timeRename(w, time.Now(), len(jobs))
	for _, job := range jobs {
		target := job.TargetAbsPath

		if !overwrite {
			if _, err := os.Stat(target); err == nil {
				if verbose {
					fmt.Fprintf(w, "Collision: %s", target)
				}
				target = resolveFSCollisions(w, target, verbose)
			}
		}

		if err := os.Rename(job.SourceAbsPath, target); err != nil {
			fmt.Fprintf(w, "Error renaming %s to %s: %v\n", job.Source, target, err)
		} else if verbose {
			fmt.Fprintf(w, "Renamed: %s ➔ %s\n", job.Source, target)
		}
	}
}

// For multiple jobs, resolve target collisions to present user intended action
// FS collisions are still handled during rename if changed before renaming.
func resolveJobCollisions(jobs []models.Job) []models.Job {
	seenTargets := make(map[string]bool)

	for i := range jobs {
		original := jobs[i].Target
		unique := original
		counter := 1

		// Find all collisions with resolved jobs and generate new target number.
		for seenTargets[unique] {
			ext := filepath.Ext(original)
			base := strings.TrimSuffix(original, ext)
			unique = fmt.Sprintf("%s_%d%s", base, counter, ext)
			counter++
		}

		if unique != original {
			jobs[i].Target = unique
			dir := filepath.Dir(jobs[i].TargetAbsPath)
			jobs[i].TargetAbsPath = filepath.Join(dir, unique)
		}

		seenTargets[unique] = true
	}

	return jobs
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
