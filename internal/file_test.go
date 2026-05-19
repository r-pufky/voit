package internal

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestScan_SingleFileSuccess(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "scan_target_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	files, err := Scan(tmpFile.Name())
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(files) != 1 {
		t.Errorf("Expected 1 file in slice, got %d", len(files))
	}
	if files[0].Source != tmpFile.Name() {
		t.Errorf("Expected file path %s, got %s", tmpFile.Name(), files[0].Source)
	}
}

func TestScan_DirectoryWithMixedContents(t *testing.T) {
	baseDir, err := os.MkdirTemp("", "scan_dir_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(baseDir)

	file1, err := os.Create(filepath.Join(baseDir, "file1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	file1.Close()

	file2, err := os.Create(filepath.Join(baseDir, "file2.log"))
	if err != nil {
		t.Fatal(err)
	}
	file2.Close()

	hidden, err := os.Create(filepath.Join(baseDir, ".gitignore"))
	if err != nil {
		t.Fatal(err)
	}
	hidden.Close()

	subDir := filepath.Join(baseDir, "ignore_sub_dir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	files, err := Scan(baseDir)
	if err != nil {
		t.Fatalf("Expected no errors, got error: %v", err)
	}

	if len(files) != 3 {
		t.Errorf("Expected 2 files processed from directory, found %d", len(files))
	}

	foundFile1 := false
	foundFile2 := false
	foundFile3 := false
	for _, f := range files {
		switch filepath.Base(f.Source) {
		case "file1.txt":
			foundFile1 = true
		case "file2.log":
			foundFile2 = true
		case ".gitignore":
			foundFile3 = true
		}
	}

	if !foundFile1 || !foundFile2 || !foundFile3 {
		t.Error("Scan failed to properly capture all immediate target files")
	}
}

func TestScan_InitialStatError(t *testing.T) {
	_, err := Scan("/invalid/system/path/that/does/not/exist/anywhere")
	if err == nil {
		t.Fatal("Expected an error from os.Stat tracking a fake path, got nil")
	}
}

func TestScan_EmptyDirectory(t *testing.T) {
	emptyDir, err := os.MkdirTemp("", "empty_dir_*")
	if err != nil {
		t.Fatalf("Failed to create empty temp dir: %v", err)
	}
	defer os.RemoveAll(emptyDir)

	files, err := Scan(emptyDir)
	if err != nil {
		t.Fatalf("Expected no error on empty folder, got: %v", err)
	}

	if len(files) != 0 {
		t.Errorf("Expected 0 elements returned, got %d", len(files))
	}
}

func TestMultiExt(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedName string
		expectedExt  string
	}{
		{
			name:         "Multi-ext standard match",
			input:        "photo.tar.gz",
			expectedName: "photo",
			expectedExt:  ".tar.gz",
		},
		{
			name:         "Multi-ext match with directory path",
			input:        "/var/log/app.min.js",
			expectedName: "app",
			expectedExt:  ".min.js",
		},
		{
			name:         "Single ext standard match",
			input:        "document.pdf",
			expectedName: "document",
			expectedExt:  ".pdf",
		},
		{
			name:         "Single ext with directory path",
			input:        "images/vacation/sunset.jpg",
			expectedName: "sunset",
			expectedExt:  ".jpg",
		},
		{
			name:         "File with intermediate dots but single extension",
			input:        "backup.v1.0.zip",
			expectedName: "backup.v1.0",
			expectedExt:  ".zip",
		},
		{
			name:         "No extension standard file",
			input:        "README",
			expectedName: "README",
			expectedExt:  "",
		},
		{
			name:         "No extension file with path",
			input:        "/usr/local/bin/scratch",
			expectedName: "scratch",
			expectedExt:  "",
		},
		{
			name:         "Hidden Unix file (starts with dot)",
			input:        ".gitignore",
			expectedName: ".gitignore",
			expectedExt:  "",
		},
		{
			name:         "Empty string input",
			input:        "",
			expectedName: ".",
			expectedExt:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, gotExt := SplitMultiExt(tt.input)

			if gotName != tt.expectedName || gotExt != tt.expectedExt {
				t.Errorf("\nMultiExt(%q)\nGot:  (%q, %q)\nWant: (%q, %q)",
					tt.input, gotName, gotExt, tt.expectedName, tt.expectedExt)
			}
		})
	}
}

func TestNew_Success(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_file_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	fileModel, err := new(tmpFile.Name())
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if fileModel.Source == "" {
		t.Error("Expected Source path, got empty string")
	}
	if !strings.HasSuffix(fileModel.Ext, ".txt") {
		t.Errorf("Expected extension .txt, got %s", fileModel.Ext)
	}
	if fileModel.CTime.Location() != time.UTC || fileModel.MTime.Location() != time.UTC {
		t.Error("Expected CTime and MTime to be in UTC")
	}
}

func TestNew_DirectoryError(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_dir_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fileModel, err := new(tmpDir)
	if err == nil {
		t.Fatal("Expected an error when passing a directory, got nil")
	}

	expectedErr := "Skipped (directory):"
	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("Expected error to contain %q, got %q", expectedErr, err.Error())
	}
	if fileModel != nil {
		t.Errorf("Expected returned file model to be nil, got %v", fileModel)
	}
}

func TestNew_Stat_FatalError(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		_, _ = new("/nonexistent/path/to/file/that/does/not/exist.txt")
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestNew_Stat_FatalError")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()

	e, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("Expected exit error from log.Fatalf, got: %v", err)
	}
	if e.Success() {
		t.Error("Expected subprocess to crash with non-zero exit code, but it succeeded")
	}
}

func TestNew_LinuxOSPath(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping Linux-specific stat check path test")
	}

	tmpFile, err := os.CreateTemp("", "test_linux_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	fileModel, err := new(tmpFile.Name())
	if err != nil {
		t.Fatalf("Expected no error on Linux, got: %v", err)
	}

	if fileModel.MTime.IsZero() {
		t.Error("Expected MTime to be parsed from linux syscall stat, got zero time")
	}
}
