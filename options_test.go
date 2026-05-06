package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// Wrap exit in sub process to capture exit codes.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("WANT_HELPER_PROCESS") != "1" {
		return
	}

	args := os.Args
	for i := range args {
		if args[i] == "--" {
			os.Args = args[i:]
			break
		}
	}

	ParseFlags()
	os.Exit(0)
}

func HelperAbsPath(path string) string {
	if path == "" {
		return ""
	}
	abs, _ := filepath.Abs(path)
	return abs
}

func TestParseFlagsExits(t *testing.T) {
	tests := []struct {
		test     string
		args     []string
		expected int
	}{
		{
			test:     "Invalid pattern name",
			args:     []string{"-p", "invalid_type"},
			expected: 2,
		},
		{
			test:     "Missing both file and directory",
			args:     []string{"-p", "mns"},
			expected: 2,
		},
		{
			test:     "Mutually exclusive file and directory",
			args:     []string{"-f", "test.txt", "-d", "./data"},
			expected: 2,
		},
		{
			test:     "Build version",
			args:     []string{"-b"},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=TestHelperProcess", "--")
			cmd.Args = append(cmd.Args, tt.args...)
			cmd.Env = append(os.Environ(), "WANT_HELPER_PROCESS=1")

			err := cmd.Run()
			e, ok := err.(*exec.ExitError)

			if ok && e.ExitCode() != tt.expected {
				t.Errorf("expected exit code %d, got %d", tt.expected, e.ExitCode())
			} else if !ok && tt.expected != 0 {
				t.Errorf("expected exit code %d, but process succeeded", tt.expected)
			}
		})
	}
}

func TestParseFlagsSuccess(t *testing.T) {
	tests := []struct {
		test     string
		args     []string
		wantDir  string
		wantFile string
	}{
		{
			test:     "Valid File Input",
			args:     []string{"-f", "logs/test.log", "-l"},
			wantDir:  HelperAbsPath("logs"),
			wantFile: "test.log",
		},
		{
			test:     "Valid Directory Input",
			args:     []string{"--dir", "/tmp/data", "--yes"},
			wantDir:  HelperAbsPath("/tmp/data"),
			wantFile: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			// Save original args and restore after.
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()
			os.Args = append([]string{"cmd"}, tt.args...)

			opts := ParseFlags()

			if opts.Directory != tt.wantDir {
				t.Errorf("Directory: got %s, want %s", opts.Directory, tt.wantDir)
			}
			if opts.File != tt.wantFile {
				t.Errorf("File: got %s, want %s", opts.File, tt.wantFile)
			}
		})
	}
}
