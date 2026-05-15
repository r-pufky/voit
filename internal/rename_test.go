package internal

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/r-pufky/voit/models"
)

func TestAddJob(t *testing.T) {
	tempDir := t.TempDir()
	existingFile := filepath.Join(tempDir, "2027-10-27T10.30.05.123 - 20271027_103005123.jpg")
	os.WriteFile(existingFile, []byte("content"), 0644)

	transformFile := filepath.Join(tempDir, "20231027_103005123.jpg")
	os.WriteFile(transformFile, []byte("content"), 0644)

	targetExistsFile := filepath.Join(tempDir, "exists.txt")
	os.WriteFile(targetExistsFile, []byte("content"), 0644)

	pattern := "ms"

	tests := []struct {
		test     string
		path     string
		lower    bool
		strip    bool
		created  bool
		modified bool
		wantErr  bool
		wantNil  bool
	}{
		{
			test:    "Empty path",
			path:    "",
			lower:   false,
			strip:   false,
			wantErr: true,
			wantNil: false,
		},
		{
			test:    "File missing",
			path:    filepath.Join(tempDir, "missing.txt"),
			lower:   true,
			strip:   false,
			wantErr: true,
			wantNil: false,
		},
		{
			test:    "No transformation needed",
			path:    existingFile,
			lower:   false,
			strip:   false,
			wantErr: false,
			wantNil: true,
		},
		{
			test:    "Transform needed",
			path:    transformFile,
			lower:   true,
			strip:   false,
			wantErr: false,
			wantNil: false,
		},
		{
			test:    "Target already exists",
			path:    existingFile,
			lower:   true,
			strip:   false,
			wantErr: false,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			job, err := addJob(tt.path, pattern, tt.lower, tt.strip, tt.created, tt.modified)
			if (err != nil) != tt.wantErr {
				t.Errorf("\naddJob()\nerror: %v,\nwant:  %v", err, tt.wantErr)
			}
			if job == nil && !tt.wantNil && !tt.wantErr {
				t.Errorf("addJob() returned nil unexpectedly")
			}
		})
	}
}

func TestCreateJobs(t *testing.T) {
	tempDir := t.TempDir()
	renameFile := filepath.Join(tempDir, "20231027_103005123.jpg")
	os.WriteFile(renameFile, []byte("1"), 0644)
	renameFile2 := filepath.Join(tempDir, "20231027_103005123-1.jpg")
	os.WriteFile(renameFile2, []byte("1"), 0644)
	skipFile := filepath.Join(tempDir, "file2.txt")
	os.WriteFile(skipFile, []byte("2"), 0644)

	pattern := "ms"

	t.Run("Single file success", func(t *testing.T) {
		jobs, width, err := CreateJobs(renameFile, "", pattern, true, false, false, false)
		if err != nil || len(jobs) != 1 {
			t.Errorf("Expected 1 job, got %d", len(jobs))
		}
		if width == 0 {
			t.Error("Expected width to be set")
		}
	})

	t.Run("Directory walk success", func(t *testing.T) {

		jobs, _, err := CreateJobs("", tempDir, pattern, true, false, false, false)
		if err != nil {
			t.Errorf("Dir walk failed: %v", err)
		}
		if err == nil && len(jobs) != 2 {
			t.Errorf("Expected 2 jobs, got %d", len(jobs))
		}
	})
}

func TestExecuteRename(t *testing.T) {
	buf := &bytes.Buffer{}
	tempDir := t.TempDir()
	src := filepath.Join(tempDir, "source.txt")
	dst := filepath.Join(tempDir, "dest.txt")
	os.WriteFile(src, []byte("hello"), 0644)

	jobs := []models.Job{
		{
			SourceAbsPath: src,
			TargetAbsPath: dst,
		},
	}

	ExecuteRename(buf, jobs, false, false)

	if _, err := os.Stat(dst); os.IsNotExist(err) {
		t.Errorf("File was not renamed to %s", dst)
	}
	if _, err := os.Stat(src); err == nil {
		t.Errorf("Source file %s still exists", src)
	}
}

func TestExecuteRenameFSCollision(t *testing.T) {
	buf := &bytes.Buffer{}
	tempDir := t.TempDir()
	src := filepath.Join(tempDir, "source.txt")
	os.WriteFile(src, []byte("new"), 0644)
	dst := filepath.Join(tempDir, "dest.txt")
	os.WriteFile(dst, []byte("original"), 0644)

	expectedDst := filepath.Join(tempDir, "dest_1.txt")

	jobs := []models.Job{
		{
			SourceAbsPath: src,
			TargetAbsPath: dst,
		},
	}

	ExecuteRename(buf, jobs, false, false)

	if data, _ := os.ReadFile(dst); string(data) != "original" {
		t.Errorf("Original file incorrectly overwritten.")
	}

	if _, err := os.Stat(expectedDst); os.IsNotExist(err) {
		t.Errorf("File not renamed: %s.", expectedDst)
	}

	if data, _ := os.ReadFile(expectedDst); string(data) != "new" {
		t.Errorf("Renamed file content is incorrect.")
	}

	if _, err := os.Stat(src); err == nil {
		t.Errorf("Source file %s still exists.", src)
	}
}

func TestResolveJobCollisions(t *testing.T) {
	tests := []struct {
		name     string
		input    []models.Job
		expected []models.Job
	}{
		{
			name: "No collisions",
			input: []models.Job{
				{Target: "file1.txt", TargetAbsPath: "/tmp/file1.txt"},
				{Target: "file2.txt", TargetAbsPath: "/tmp/file2.txt"},
			},
			expected: []models.Job{
				{Target: "file1.txt", TargetAbsPath: "/tmp/file1.txt"},
				{Target: "file2.txt", TargetAbsPath: "/tmp/file2.txt"},
			},
		},
		{
			name: "Simple collision",
			input: []models.Job{
				{Target: "notes.txt", TargetAbsPath: "/home/dest/notes.txt"},
				{Target: "notes.txt", TargetAbsPath: "/home/dest/notes.txt"},
			},
			expected: []models.Job{
				{Target: "notes.txt", TargetAbsPath: "/home/dest/notes.txt"},
				{Target: "notes_1.txt", TargetAbsPath: "/home/dest/notes_1.txt"},
			},
		},
		{
			name: "Multiple collisions and path preservation",
			input: []models.Job{
				{Target: "data.json", TargetAbsPath: "/mnt/data.json"},
				{Target: "data.json", TargetAbsPath: "/mnt/data.json"},
				{Target: "data.json", TargetAbsPath: "/mnt/data.json"},
			},
			expected: []models.Job{
				{Target: "data.json", TargetAbsPath: "/mnt/data.json"},
				{Target: "data_1.json", TargetAbsPath: "/mnt/data_1.json"},
				{Target: "data_2.json", TargetAbsPath: "/mnt/data_2.json"},
			},
		},
		{
			name: "Collision with existing incremented name",
			input: []models.Job{
				{Target: "doc.pdf", TargetAbsPath: "/usr/doc.pdf"},
				{Target: "doc_1.pdf", TargetAbsPath: "/usr/doc_1.pdf"},
				{Target: "doc.pdf", TargetAbsPath: "/usr/doc.pdf"},
			},
			expected: []models.Job{
				{Target: "doc.pdf", TargetAbsPath: "/usr/doc.pdf"},
				{Target: "doc_1.pdf", TargetAbsPath: "/usr/doc_1.pdf"},
				{Target: "doc_2.pdf", TargetAbsPath: "/usr/doc_2.pdf"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := resolveJobCollisions(tt.input)
			if !reflect.DeepEqual(actual, tt.expected) {
				t.Errorf("\nTest: %s\nExpected: %+v\nActual:   %+v", tt.name, tt.expected, actual)
			}
		})
	}
}
