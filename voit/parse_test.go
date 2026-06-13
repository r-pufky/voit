package voit

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestScanSingleFileSuccess(t *testing.T) {
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
	if files[0].File.Source != tmpFile.Name() {
		t.Errorf("Expected file path %s, got %s", tmpFile.Name(), files[0].File.Source)
	}
}

func TestScanDirectoryWithMixedContents(t *testing.T) {
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
		t.Errorf("Expected 3 files processed from directory, found %d", len(files))
	}

	foundFile1 := false
	foundFile2 := false
	foundFile3 := false
	for _, f := range files {
		switch filepath.Base(f.File.Source) {
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

func TestScanInitialStatError(t *testing.T) {
	files, err := Scan("/invalid/system/path/that/does/not/exist/anywhere")
	if files != nil {
		t.Fatalf("\nGot:  %+v\nWant: nil", files)
	}
	if err == nil || (err != nil && !strings.Contains(err.Error(), "globbing quoted?")) {
		t.Fatalf("\nGot:  %v\nWant: No files matched the known datetime formats (is globbing quoted?).", err)
	}
}

func TestScanEmptyDirectory(t *testing.T) {
	emptyDir, err := os.MkdirTemp("", "empty_dir_*")
	if err != nil {
		t.Fatalf("Failed to create empty temp dir: %v", err)
	}
	defer os.RemoveAll(emptyDir)

	files, err := Scan(emptyDir)
	if len(files) != 0 {
		t.Errorf("Expected 0 elements returned, got %d", len(files))
	}
	if err == nil || (err != nil && !strings.Contains(err.Error(), "globbing quoted?")) {
		t.Fatalf("\nGot:  %v\nWant: No files matched the known datetime formats (is globbing quoted?).", err)
	}
}

func TestScanGlobbing(t *testing.T) {
	tmpDir := t.TempDir()

	files := []string{
		"2026-06-03_photo1.jpg",
		"2026-06-03_photo2.jpg",
		"notes.txt",
	}

	for _, file := range files {
		err := os.WriteFile(filepath.Join(tmpDir, file), []byte("fake data"), 0644)
		if err != nil {
			t.Fatalf("failed to set up mock file %s: %v", file, err)
		}
	}

	// Directory matching file glob should not be parsed.
	subDir := filepath.Join(tmpDir, "2026-06-03_nested_dir.jpg")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("failed to create mock directory: %v", err)
	}

	tests := []struct {
		name      string
		pattern   string
		IsXMP     bool
		wantCount int
	}{
		{
			name:      "sanity: bare directory [all files matched]",
			pattern:   tmpDir,
			wantCount: 3,
		},
		{
			name:      "sanity: bare file [single file matched]",
			pattern:   filepath.Join(tmpDir, "2026-06-03_photo1.jpg"),
			wantCount: 1,
		},
		{
			name:      "sanity: glob files [photo 1,2 matched]",
			pattern:   filepath.Join(tmpDir, "2026-06-03_*.jpg"),
			wantCount: 2,
		},
		{
			name:      "sanity: glob all [directory skipped]",
			pattern:   filepath.Join(tmpDir, "*"),
			wantCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Scan(tt.pattern)
			if err != nil {
				t.Fatalf("Scan() returned an unexpected error: %v", err)
			}

			if len(got) != tt.wantCount {
				t.Errorf("\nGlob: %q\nGot:  %d\nWant %d", tt.pattern, len(got), tt.wantCount)
			}

			for _, f := range got {
				stat, err := os.Stat(f.File.Source)
				if err != nil {
					t.Errorf("returned file source path cannot be stat-ed: %v", err)
				}
				if stat.IsDir() {
					t.Errorf("directories should not be included: %s", f.File.Source)
				}
			}
		})
	}
}

func TestNewSuccess(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_file_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	fileModel, err := New(tmpFile.Name())
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if fileModel.File.Source == "" {
		t.Error("Expected Source path, got empty string")
	}
	if !strings.HasSuffix(fileModel.File.Ext, ".txt") {
		t.Errorf("Expected extension .txt, got %s", fileModel.File.Ext)
	}
	if fileModel.File.CTime.Location() != time.UTC || fileModel.File.MTime.Location() != time.UTC {
		t.Error("Expected CTime and MTime to be in UTC")
	}
}

func TestNewSidecarLogic(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		filename string
		wantName string
		wantExt  string
	}{
		{
			name:     "sanity: valid sidecar [sidecar parsed]",
			filename: "file.jpg.xmp",
			wantName: "file",
			wantExt:  ".jpg.xmp",
		},
		{
			name:     "sanity: xmp with no sidecar [<2 dots]",
			filename: "file.xmp",
			wantName: "file",
			wantExt:  ".xmp",
		},
		{
			name:     "sanity: standard file [jpg parsed]",
			filename: "file.xmp.jpg",
			wantName: "file.xmp",
			wantExt:  ".jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := filepath.Join(tmpDir, tt.filename)
			err := os.WriteFile(filePath, []byte("dummy data"), 0644)
			if err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			v, err := New(filePath)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Ignore absolute path prefix.
			actualBaseName := filepath.Base(v.File.Name)

			if actualBaseName != tt.wantName {
				t.Errorf("\nGot:  %q\nWant: %q", actualBaseName, tt.wantName)
			}
			if v.File.Ext != tt.wantExt {
				t.Errorf("\nGot:  %q\nWant %q", v.File.Ext, tt.wantExt)
			}
		})
	}
}

func TestNewDirectoryError(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_dir_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fileModel, err := New(tmpDir)
	if err == nil {
		t.Fatal("Expected an error when passing a directory, got nil")
	}

	expectedErr := "skipped (directory):"
	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("Expected error to contain %q, got %q", expectedErr, err.Error())
	}
	if fileModel != nil {
		t.Errorf("Expected returned file model to be nil, got %v", fileModel)
	}
}

func TestNewStatFatalError(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		_, _ = New("/nonexistent/path/to/file/that/does/not/exist.txt")
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestNewStatFatalError")
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

func TestNewLinuxOSPath(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping Linux-specific stat check path test")
	}

	tmpFile, err := os.CreateTemp("", "test_linux_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	fileModel, err := New(tmpFile.Name())
	if err != nil {
		t.Fatalf("Expected no error on Linux, got: %v", err)
	}

	if fileModel.File.MTime.IsZero() {
		t.Error("Expected MTime to be parsed from linux syscall stat, got zero time")
	}
}

func TestMultiExt(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantName string
		wantExt  string
	}{
		{
			name:     "Multi-ext standard match",
			input:    "photo.tar.gz",
			wantName: "photo",
			wantExt:  ".tar.gz",
		},
		{
			name:     "Multi-ext match with directory path",
			input:    "/var/log/app.min.js",
			wantName: "app",
			wantExt:  ".min.js",
		},
		{
			name:     "Single ext standard match",
			input:    "document.pdf",
			wantName: "document",
			wantExt:  ".pdf",
		},
		{
			name:     "Single ext with directory path",
			input:    "images/vacation/sunset.jpg",
			wantName: "sunset",
			wantExt:  ".jpg",
		},
		{
			name:     "File with intermediate dots but single extension",
			input:    "backup.v1.0.zip",
			wantName: "backup.v1.0",
			wantExt:  ".zip",
		},
		{
			name:     "No extension standard file",
			input:    "README",
			wantName: "README",
			wantExt:  "",
		},
		{
			name:     "No extension file with path",
			input:    "/usr/local/bin/scratch",
			wantName: "scratch",
			wantExt:  "",
		},
		{
			name:     "Hidden Unix file (starts with dot)",
			input:    ".gitignore",
			wantName: ".gitignore",
			wantExt:  "",
		},
		{
			name:     "Empty string input",
			input:    "",
			wantName: ".",
			wantExt:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, gotExt := SplitMultiExt(tt.input)

			if gotName != tt.wantName || gotExt != tt.wantExt {
				t.Errorf("\nMultiExt(%q)\nGot:  (%q, %q)\nWant: (%q, %q)",
					tt.input, gotName, gotExt, tt.wantName, tt.wantExt)
			}
		})
	}
}
