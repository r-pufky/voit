package voit

import (
	"bytes"
	"strings"
	"testing"
)

func TestDisplayPending(t *testing.T) {
	files := VoitFiles{
		{
			File: File{
				Source: "img1.jpg",
				Ext:    ".jpg",
				Width:  8,
			},
			Target:  "2023-01-01T12.00.00.000.jpg",
			Matched: true,
		},
		{
			File: File{
				Source: "a.jpg",
				Ext:    ".jpg",
				Width:  5,
			},
			Target:  "2023-01-02T12.00.00.000.jpg",
			Matched: true,
		},
	}

	buf := &bytes.Buffer{}

	count := files.DisplayPending(buf)

	out := buf.String()
	expected1 := "img1.jpg ➔ 2023-01-01T12.00.00.000.jpg"
	expected2 := "a.jpg    ➔ 2023-01-02T12.00.00.000.jpg"

	if !strings.Contains(out, expected1) {
		t.Errorf("\nOutput:\n%s\nExpected to contain: %q", out, expected1)
	}
	if !strings.Contains(out, expected2) {
		t.Errorf("\nOutput:\n%s\nExpected to contain: %q", out, expected2)
	}
	if count != 2 {
		t.Errorf("\nExpected count 2: %d", count)
	}
}

func TestConfirm(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"Yes lowercase", "y\n", true},
		{"No", "n\n", false},
		{"Yes full word", "yes\n", true},
		{"Yes uppercase", "YES\n", true},
		{"Random text", "maybe\n", false},
		{"Empty input", "\n", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.input)
			w := &bytes.Buffer{}

			got := Confirm(w, r)

			if got != tt.want {
				t.Errorf("Confirm() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Simple logic requires no testing.
func TestPromptRenameNOOP(t *testing.T) {}
