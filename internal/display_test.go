package internal

import (
	"bytes"
	"strings"
	"testing"

	"github.com/r-pufky/voit/models"
)

func TestDisplayPending(t *testing.T) {
	files := []models.File{
		{Source: "img1.jpg", Matched: true, Width: 8, Target: "2023-01-01.jpg"},
		{Source: "a.jpg", Matched: true, Width: 5, Target: "2023-01-02.jpg"},
	}

	buf := &bytes.Buffer{}

	count := DisplayPending(buf, files)

	out := buf.String()
	expected1 := "img1.jpg ➔ 2023-01-01.jpg"
	expected2 := "a.jpg    ➔ 2023-01-02.jpg"

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
			input := strings.NewReader(tt.input)
			output := &bytes.Buffer{}

			got := Confirm(input, output)

			if got != tt.want {
				t.Errorf("Confirm() = %v, want %v", got, tt.want)
			}
		})
	}
}
