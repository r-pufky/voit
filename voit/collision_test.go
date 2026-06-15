package voit

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveCollisions(t *testing.T) {
	vCfg := NewConfig()
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		setupFS  func(dir string)
		input    VoitFiles
		expected []string // Expected Target names
	}{
		{
			name: "sanity: no collisions [no modifications]",
			input: VoitFiles{
				{
					File: File{Source: filepath.Join(tmpDir, "file no collisions a.jpg"), Name: "file", Ext: ".jpg"},
					Mark: Meta{Desc: Desc{Text: "no collisions a"}},
				},
				{
					File: File{Source: filepath.Join(tmpDir, "file no collisions a.jpg"), Name: "file", Ext: ".jpg"},
					Mark: Meta{Desc: Desc{Text: "no collisions b"}},
				},
			},
			expected: []string{
				"0001-01-01T00.00.00.000 no collisions a.jpg",
				"0001-01-01T00.00.00.000 no collisions b.jpg",
			},
		},
		{
			name: "sanity: standard collision [count added to desc]",
			input: VoitFiles{
				{
					File: File{Source: filepath.Join(tmpDir, "file standard collision.jpg"), Name: "file", Ext: ".jpg"},
					Mark: Meta{Desc: Desc{Text: "standard collision"}},
				},
				{
					File: File{Source: filepath.Join(tmpDir, "file standard collision.jpg"), Name: "file", Ext: ".jpg"},
					Mark: Meta{Desc: Desc{Text: "standard collision"}},
				},
			},
			expected: []string{
				"0001-01-01T00.00.00.000 standard collision.jpg",
				"0001-01-01T00.00.00.000 standard collision 1.jpg",
			},
		},
		{
			name: "sanity: multi-collision [multiple collisions numerically incremented]",
			input: VoitFiles{
				{
					File: File{Source: filepath.Join(tmpDir, "file multi-collision.jpg"), Name: "file", Ext: ".jpg"},
					Mark: Meta{Desc: Desc{Text: "multi-collision"}},
				},
				{
					File: File{Source: filepath.Join(tmpDir, "file multi-collision.jpg"), Name: "file", Ext: ".jpg"},
					Mark: Meta{Desc: Desc{Text: "multi-collision"}},
				},
				{
					File: File{Source: filepath.Join(tmpDir, "file multi-collision.jpg"), Name: "file", Ext: ".jpg"},
					Mark: Meta{Desc: Desc{Text: "multi-collision"}},
				},
			},
			expected: []string{
				"0001-01-01T00.00.00.000 multi-collision.jpg",
				"0001-01-01T00.00.00.000 multi-collision 1.jpg",
				"0001-01-01T00.00.00.000 multi-collision 2.jpg",
			},
		},
		{
			name: "no desc: no tags [count added to desc for 1,2]",
			input: VoitFiles{
				{File: File{Source: filepath.Join(tmpDir, "file.jpg"), Name: "file", Ext: ".jpg"}, Mark: Meta{}},
				{File: File{Source: filepath.Join(tmpDir, "file.jpg"), Name: "file", Ext: ".jpg"}, Mark: Meta{}},
			},
			expected: []string{
				"0001-01-01T00.00.00.000.jpg",
				"0001-01-01T00.00.00.000 1.jpg",
			},
		},
		{
			name: "no desc: tags [count added to desc for 1,2]",
			input: VoitFiles{
				{
					File: File{Source: filepath.Join(tmpDir, "file.jpg"), Name: "file", Ext: ".jpg"},
					Mark: Meta{Tags: Tag{Items: []string{"nodesctags"}}, Desc: Desc{Text: ""}},
				},
				{
					File: File{Source: filepath.Join(tmpDir, "file.jpg"), Name: "file", Ext: ".jpg"},
					Mark: Meta{Tags: Tag{Items: []string{"nodesctags"}}, Desc: Desc{Text: ""}},
				},
			},
			expected: []string{
				"0001-01-01T00.00.00.000 -- nodesctags.jpg",
				"0001-01-01T00.00.00.000 1 -- nodesctags.jpg",
			},
		},
		{
			name: "fs-collision: single collision [count 1 added to desc]",
			setupFS: func(dir string) {
				os.WriteFile(filepath.Join(dir, "0001-01-01T00.00.00.000.jpg"), []byte(""), 0644)
			},
			input: VoitFiles{
				{File: File{Source: filepath.Join(tmpDir, "file.jpg"), Name: "file", Ext: ".jpg"}, Mark: Meta{}},
			},
			expected: []string{"0001-01-01T00.00.00.000 1.jpg"},
		},
		{
			name: "fs-collision: multi-collision [count 2 added to desc]",
			setupFS: func(dir string) {
				os.WriteFile(filepath.Join(dir, "0001-01-01T00.00.00.000.jpg"), []byte(""), 0644)
				os.WriteFile(filepath.Join(dir, "0001-01-01T00.00.00.000 1.jpg"), []byte(""), 0644)
			},
			input: VoitFiles{
				{File: File{Source: filepath.Join(tmpDir, "file.jpg"), Name: "file", Ext: ".jpg"}, Mark: Meta{}},
			},
			expected: []string{"0001-01-01T00.00.00.000 2.jpg"},
		},
		{
			name: "multi-collision: list and fs collision [count 2,3 added to desc]",
			setupFS: func(dir string) {
				os.WriteFile(filepath.Join(dir, "0001-01-01T00.00.00.000.jpg"), []byte(""), 0644)
				os.WriteFile(filepath.Join(dir, "0001-01-01T00.00.00.000 1.jpg"), []byte(""), 0644)
			},
			input: VoitFiles{
				{File: File{Source: filepath.Join(tmpDir, "file.jpg"), Name: "file", Ext: ".jpg"}, Mark: Meta{}},
				{File: File{Source: filepath.Join(tmpDir, "file.jpg"), Name: "file", Ext: ".jpg"}, Mark: Meta{}},
			},
			expected: []string{"0001-01-01T00.00.00.000 2.jpg", "0001-01-01T00.00.00.000 3.jpg"},
		},
		{
			name: "sanity: sidecar collision [file sidecar renamed in tandem]",
			setupFS: func(dir string) {
				os.WriteFile(filepath.Join(dir, "0001-01-01T00.00.00.000 sidecar.jpg"), []byte(""), 0644)
				os.WriteFile(filepath.Join(dir, "0001-01-01T00.00.00.000 sidecar.jpg.xmp"), []byte(""), 0644)
			},
			input: VoitFiles{
				{
					File: File{Source: filepath.Join(tmpDir, "file.jpg"), Name: "file", Ext: ".jpg"},
					Mark: Meta{Desc: Desc{Text: "sidecar"}},
				},
				{
					File: File{Source: filepath.Join(tmpDir, "file.jpg.xmp"), Name: "file", Ext: ".jpg.xmp"},
					Mark: Meta{Desc: Desc{Text: "sidecar"}},
				},
			},
			expected: []string{"0001-01-01T00.00.00.000 sidecar 1.jpg", "0001-01-01T00.00.00.000 sidecar 1.jpg.xmp"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFS != nil {
				tt.setupFS(tmpDir)
			}

			tt.input.ResolveCollisions(vCfg, false)

			for i, f := range tt.input {
				got := filepath.Base(f.Target)
				if got != tt.expected[i] {
					t.Errorf("\nIndex %d\nGot:  %q\nWant: %q\n", i, got, tt.expected[i])
				}
			}
		})
	}
}
