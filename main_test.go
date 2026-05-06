package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestDisplayPending(t *testing.T) {
	jobs := []Voit{
		{source: "img1.jpg", target: "2023-01-01.jpg"},
		{source: "a.jpg", target: "2023-01-02.jpg"},
	}
	buf := &bytes.Buffer{}

	DisplayPending(buf, jobs, 8)

	out := buf.String()
	expected1 := "img1.jpg ➔ 2023-01-01.jpg"
	expected2 := "a.jpg    ➔ 2023-01-02.jpg"

	if !strings.Contains(out, expected1) {
		t.Errorf("Expected output to contain %q", expected1)
	}
	if !strings.Contains(out, expected2) {
		t.Errorf("Expected output to contain %q (check padding)", expected2)
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
