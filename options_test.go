package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

// Wrap exit in sub process to capture exit codes.
func TestHelperProcess(t *testing.T) {
	buf := &bytes.Buffer{}

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

	Config(buf, "/tmp/non-existent-voit.toml")
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
			test:     "Mutually exclusive created and modified",
			args:     []string{"-f", "test.txt", "-c", "-m"},
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
	configData := []byte(`
Pattern = "ns"
Lower = true
Directory = "/tmp/override"
`)
	tests := []struct {
		test        string
		args        []string
		useConfig   bool
		wantDir     string
		wantFile    string
		wantPattern string
		wantLower   bool
	}{
		{
			test:        "Valid File Input",
			args:        []string{"-f", "/tmp/test.jpg", "-l"},
			useConfig:   false,
			wantDir:     HelperAbsPath("/tmp"),
			wantFile:    "test.jpg",
			wantPattern: "ms",
			wantLower:   true,
		},
		{
			test:        "Valid Directory Input",
			args:        []string{"--dir", "/tmp/data", "--yes"},
			useConfig:   false,
			wantDir:     HelperAbsPath("/tmp/data"),
			wantFile:    "",
			wantPattern: "ms",
		},
		{
			test:        "Pattern override",
			args:        []string{},
			useConfig:   true,
			wantDir:     HelperAbsPath("/tmp/override"),
			wantFile:    "",
			wantPattern: "ns",
			wantLower:   true,
		},
		{
			test:        "Directory/file flag overrides config",
			args:        []string{"--dir", "/tmp/override"},
			useConfig:   true,
			wantDir:     HelperAbsPath("/tmp/override"),
			wantFile:    "",
			wantPattern: "ns",
			wantLower:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			buf := &bytes.Buffer{}
			// Setup testing config if used.
			viper.Reset()
			tmpDir := t.TempDir()
			realConfig := filepath.Join(tmpDir, "voit.toml")
			noConfig := filepath.Join(tmpDir, "non-existent.toml")

			var config string
			if tt.useConfig {
				os.WriteFile(realConfig, configData, 0644)
				config = realConfig
			} else {
				config = noConfig
			}

			// Save original args and restore after.
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()
			os.Args = append([]string{"cmd"}, tt.args...)

			opts := Config(buf, config)

			if opts.Directory != tt.wantDir {
				t.Errorf("Directory: got %s, want %s", opts.Directory, tt.wantDir)
			}
			if opts.File != tt.wantFile {
				t.Errorf("File: got %s, want %s", opts.File, tt.wantFile)
			}
			if opts.Lower != tt.wantLower {
				t.Errorf("Lower: got %t, want %t", opts.Lower, tt.wantLower)
			}
		})
	}
}
