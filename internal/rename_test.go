package internal

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/r-pufky/voit/models"
)

func TestExecuteRename(t *testing.T) {
	buf := &bytes.Buffer{}
	tempDir := t.TempDir()
	src := filepath.Join(tempDir, "source.txt")
	dst := filepath.Join(tempDir, "dest.txt")
	os.WriteFile(src, []byte("hello"), 0644)

	files := []models.File{
		{
			Source:  src,
			Matched: true,
			Target:  dst,
		},
	}

	ExecuteRename(buf, files, false, false)

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

	files := []models.File{
		{
			Source:  src,
			Matched: true,
			Target:  dst,
		},
	}

	ExecuteRename(buf, files, false, false)

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
func TestSelectTargets(t *testing.T) {
	voitTime := time.Date(2026, time.February, 2, 12, 5, 20, 700000000, time.UTC)
	noCollisionTime := time.Date(2027, time.February, 2, 12, 5, 20, 700000000, time.UTC)

	tests := []struct {
		name  string
		input []models.File
		want  []string
	}{
		{
			name: "sanity: no collisions",
			input: []models.File{
				{
					CTime:  voitTime,
					MTime:  voitTime,
					VTime:  voitTime,
					Source: "/tmp/2026-02-02T12.05.20.700 - beach vacation -- summer vacation beach.jpg",
					Name:   "2026-02-02T12.05.20.700 - beach vacation -- summer vacation beach",
					Ext:    ".jpg",
				},
				{
					CTime:  noCollisionTime,
					MTime:  noCollisionTime,
					VTime:  noCollisionTime,
					Source: "/tmp/2027-02-02T12.05.20.700 - beach vacation -- summer vacation beach.jpg",
					Name:   "2027-02-02T12.05.20.700 - beach vacation -- summer vacation beach",
					Ext:    ".jpg",
				},
			},
			want: []string{
				"/tmp/2026-02-02T12.05.20.700.jpg",
				"/tmp/2027-02-02T12.05.20.700.jpg",
			},
		},
		{
			name: "sanity: single file collision",
			input: []models.File{
				{
					CTime:  voitTime,
					MTime:  voitTime,
					VTime:  voitTime,
					Source: "/tmp/2026-02-02T12.05.20.700 - beach vacation -- summer vacation beach.jpg",
					Name:   "2026-02-02T12.05.20.700 - beach vacation -- summer vacation beach",
					Ext:    ".jpg",
				},
				{
					CTime:  voitTime,
					MTime:  voitTime,
					VTime:  voitTime,
					Source: "/tmp/2026-02-02T12.05.20.700 - beach vacation -- summer vacation beach.jpg",
					Name:   "2026-02-02T12.05.20.700 - beach vacation -- summer vacation beach",
					Ext:    ".jpg",
				},
			},
			want: []string{
				"/tmp/2026-02-02T12.05.20.700.jpg",
				"/tmp/2026-02-02T12.05.20.700_1.jpg",
			},
		},
		{
			name: "sanity: cascading collision",
			input: []models.File{
				{
					CTime:  voitTime,
					MTime:  voitTime,
					VTime:  voitTime,
					Source: "/tmp/2026-02-02T12.05.20.700 - beach vacation -- summer vacation beach.jpg",
					Name:   "2026-02-02T12.05.20.700 - beach vacation -- summer vacation beach",
					Ext:    ".jpg",
				},
				{
					CTime:  voitTime,
					MTime:  voitTime,
					VTime:  voitTime,
					Source: "/tmp/2026-02-02T12.05.20.700 - beach vacation -- summer vacation beach.jpg",
					Name:   "2026-02-02T12.05.20.700 - beach vacation -- summer vacation beach",
					Ext:    ".jpg",
				},
				{
					CTime:  voitTime,
					MTime:  voitTime,
					VTime:  voitTime,
					Source: "/tmp/2026-02-02T12.05.20.700 - beach vacation -- summer vacation beach.jpg",
					Name:   "2026-02-02T12.05.20.700 - beach vacation -- summer vacation beach",
					Ext:    ".jpg",
				},
			},
			want: []string{
				"/tmp/2026-02-02T12.05.20.700.jpg",
				"/tmp/2026-02-02T12.05.20.700_1.jpg",
				"/tmp/2026-02-02T12.05.20.700_2.jpg",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := models.Opts{
				Yes:       true,
				TagSep:    " -- ",
				DescSep:   " - ",
				SpanSep:   "--",
				AbsSource: "/tmp",
				Rename: models.RenameOpts{
					Pattern: "voit",
					NoDesc:  true,
					NoTags:  true,
				},
			}

			selectTargets(tt.input, opts)

			for i, file := range tt.input {
				if file.Target != tt.want[i] {
					t.Errorf("\nIndex: %d\nWant: %q\nGot:  %q", i, tt.want[i], file.Target)
				}
			}
		})
	}
}
