package internal

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/r-pufky/voit/voit"
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
	if files[0].File.Source != tmpFile.Name() {
		t.Errorf("Expected file path %s, got %s", tmpFile.Name(), files[0].File.Source)
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

func TestScan_InitialStatError(t *testing.T) {
	files, err := Scan("/invalid/system/path/that/does/not/exist/anywhere")
	if files != nil {
		t.Fatalf("Got:  %+v\nWant: nil", files)
	}
	if err != nil {
		t.Fatalf("Got: %v\nWant: nil", err)
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

func TestScan_Globbing(t *testing.T) {
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
			name:      "sanity: no match",
			pattern:   filepath.Join(tmpDir, "*.png"),
			wantCount: 0,
		},
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
				t.Errorf("Glob: %q\nGot:  %d\nWant %d", tt.pattern, len(got), tt.wantCount)
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

func TestNew_SidecarLogic(t *testing.T) {
	// Create a temporary directory for our test files
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

			v, err := new(filePath)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Ignore absolute path prefix.
			actualBaseName := filepath.Base(v.File.Name)

			if actualBaseName != tt.wantName {
				t.Errorf("Got:  %q\nWant: %q", actualBaseName, tt.wantName)
			}
			if v.File.Ext != tt.wantExt {
				t.Errorf("Got:  %q\nWant %q", v.File.Ext, tt.wantExt)
			}
		})
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

	expectedErr := "skipped (directory):"
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

	if fileModel.File.MTime.IsZero() {
		t.Error("Expected MTime to be parsed from linux syscall stat, got zero time")
	}
}

func TestExecuteRename(t *testing.T) {
	tests := []struct {
		name      string
		overwrite bool
		setupFS   func(t *testing.T, baseDir string) []*voit.Voit
		verifyFS  func(t *testing.T, baseDir string, files []*voit.Voit, output string)
	}{
		{
			name:      "sanity: skip unmatched files [no file changes]",
			overwrite: false,
			setupFS: func(t *testing.T, baseDir string) []*voit.Voit {
				src := filepath.Join(baseDir, "unmatched.jpg")
				if err := os.WriteFile(src, []byte("data"), 0644); err != nil {
					t.Fatal(err)
				}
				return []*voit.Voit{
					{
						File:    voit.File{Source: src, Ext: ".jpg"},
						Matched: false,
						Target:  filepath.Join(baseDir, "target.jpg"),
					},
				}
			},
			verifyFS: func(t *testing.T, baseDir string, files []*voit.Voit, output string) {
				if _, err := os.Stat(files[0].File.Source); os.IsNotExist(err) {
					t.Errorf("Expected source file to remain untouched, but it was deleted or moved")
				}
				if !strings.Contains(output, "Renamed 1 files") {
					t.Errorf("Expected summary metric message, got: %q", output)
				}
			},
		},
		{
			name:      "sanity: default [file renamed]",
			overwrite: false,
			setupFS: func(t *testing.T, baseDir string) []*voit.Voit {
				src := filepath.Join(baseDir, "photo.jpg")
				if err := os.WriteFile(src, []byte("data"), 0644); err != nil {
					t.Fatal(err)
				}
				return []*voit.Voit{
					{
						File:    voit.File{Source: src, Ext: ".jpg"},
						Matched: true,
						Target:  filepath.Join(baseDir, "vacation.jpg"),
					},
				}
			},
			verifyFS: func(t *testing.T, baseDir string, files []*voit.Voit, output string) {
				if _, err := os.Stat(files[0].Target); os.IsNotExist(err) {
					t.Errorf("Expected target file %s to exist, but it was not found", files[0].Target)
				}
				if !strings.Contains(output, "Renamed:") {
					t.Errorf("Expected verbose rename tracking output, got: %q", output)
				}
			},
		},
		{
			name:      "sanity: overwrite [vacation.jpg is overwritten]",
			overwrite: true,
			setupFS: func(t *testing.T, baseDir string) []*voit.Voit {
				src := filepath.Join(baseDir, "photo.jpg")
				tgt := filepath.Join(baseDir, "vacation.jpg")

				if err := os.WriteFile(src, []byte("new-content"), 0644); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(tgt, []byte("old-content"), 0644); err != nil {
					t.Fatal(err)
				}
				return []*voit.Voit{
					{
						File:    voit.File{Source: src, Ext: ".jpg"},
						Matched: true,
						Target:  tgt,
					},
				}
			},
			verifyFS: func(t *testing.T, baseDir string, files []*voit.Voit, output string) {
				content, err := os.ReadFile(files[0].Target)
				if err != nil {
					t.Fatal(err)
				}
				if string(content) != "new-content" {
					t.Errorf("Expected file to be overwritten with 'new-content', got %q", string(content))
				}
				if strings.Contains(output, "Collision:") {
					t.Errorf("Expected no collision notification when overwrite is true")
				}
			},
		},
		{
			name:      "sanity: collision [target is renamed to vacation_2.jpg]",
			overwrite: false,
			setupFS: func(t *testing.T, baseDir string) []*voit.Voit {
				src := filepath.Join(baseDir, "photo.jpg")
				tgt := filepath.Join(baseDir, "vacation.jpg")
				collision1 := filepath.Join(baseDir, "vacation_1.jpg")

				if err := os.WriteFile(src, []byte("moving-file"), 0644); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(tgt, []byte("blocker"), 0644); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(collision1, []byte("blocker-sub"), 0644); err != nil {
					t.Fatal(err)
				}
				return []*voit.Voit{
					{
						File:    voit.File{Source: src, Ext: ".jpg"},
						Matched: true,
						Target:  tgt,
					},
				}
			},
			verifyFS: func(t *testing.T, baseDir string, files []*voit.Voit, output string) {
				expectedFinalTarget := filepath.Join(baseDir, "vacation_2.jpg")
				if _, err := os.Stat(expectedFinalTarget); os.IsNotExist(err) {
					t.Errorf("Expected %s to exist", expectedFinalTarget)
				}
				if !strings.Contains(output, "Collision:") || !strings.Contains(output, "Collision (new target):") {
					t.Errorf("Expected collision notification, got context: %q", output)
				}
			},
		},
		{
			name:      "sanity: missing source [file deleted mid-execution, error logged]",
			overwrite: false,
			setupFS: func(t *testing.T, baseDir string) []*voit.Voit {
				return []*voit.Voit{
					{
						File:    voit.File{Source: filepath.Join(baseDir, "non-existent.jpg"), Ext: ".jpg"},
						Matched: true,
						Target:  filepath.Join(baseDir, "output.jpg"),
					},
				}
			},
			verifyFS: func(t *testing.T, baseDir string, files []*voit.Voit, output string) {
				if !strings.Contains(output, "Error renaming") {
					t.Errorf("Expected Error renaming, got: %q", output)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			files := tt.setupFS(t, tmpDir)

			var buf bytes.Buffer
			Rename(&buf, files, tt.overwrite, true)

			tt.verifyFS(t, tmpDir, files, buf.String())
		})
	}
}

func TestResolveFSCollisions(t *testing.T) {
	tests := []struct {
		name       string
		files      []string
		path       string
		wantSuffix string
		wantOutput string
	}{
		{
			name:       "sanity: no additional collisions [file_1.txt]",
			files:      []string{},
			path:       "file.txt",
			wantSuffix: "_1.txt",
			wantOutput: "Collision (new target): ",
		},
		{
			name: "sanity: additional collisions [file_2.txt]",
			files: []string{
				"file_1.txt",
			},
			path:       "file.txt",
			wantSuffix: "_2.txt",
			wantOutput: "Collision (new target): ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			for _, file := range tt.files {
				fullPath := filepath.Join(tmpDir, file)
				if err := os.WriteFile(fullPath, []byte("dummy"), 0644); err != nil {
					t.Fatalf("failed to set up test file %s: %v", fullPath, err)
				}
			}

			fullpath := filepath.Join(tmpDir, tt.path)
			var buf bytes.Buffer
			result := resolveFSCollisions(&buf, fullpath)

			if !strings.HasSuffix(result, tt.wantSuffix) {
				t.Errorf("expected path to end with %q, got %q", tt.wantSuffix, result)
			}

			if _, err := os.Stat(result); !os.IsNotExist(err) {
				t.Errorf("expected returned path %q to not exist, but it does", result)
			}

			expectedFullOutput := fmt.Sprintf("%s%s\n", tt.wantOutput, result)
			if buf.String() != expectedFullOutput {
				t.Errorf("expected output %q, got %q", expectedFullOutput, buf.String())
			}
		})
	}
}
