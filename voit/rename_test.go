package voit

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestStageRename(t *testing.T) {
	fixedSTime := time.Date(2026, time.March, 1, 20, 0, 0, 0, time.UTC)
	parsedTime := time.Date(2026, time.May, 17, 10, 45, 36, 300000000, time.UTC)
	voitTime := time.Date(2026, time.February, 2, 12, 5, 20, 700000000, time.UTC)
	baseTags := []string{"summer", "vacation", "beach"}
	baseDesc := "beach vacation"

	tests := []struct {
		name     string
		opts     *Opts
		f        VoitFiles
		wantVoit []Voit
	}{
		{
			name: "sanity: no match [ZeroTime, matched false]",
			opts: &Opts{
				Rename: RenameOpts{},
			},
			f: VoitFiles{
				{
					File: File{
						Source: "/tmp/invalid_format.txt",
						Name:   "invalid_format",
						Ext:    ".txt",
					},
				},
			},
			wantVoit: []Voit{
				{
					File: File{
						Source: "/tmp/invalid_format.txt",
						Name:   "invalid_format",
						Ext:    ".txt",
					},
					Orig: Meta{
						Desc: Desc{Text: "invalid_format"},
					},
					Mark: Meta{
						Desc: Desc{Text: "invalid_format"},
					},
					Matched: false,
				},
			},
		},
		{
			name: "sanity: default [vtime match]",
			opts: &Opts{
				Rename: RenameOpts{},
			},
			f: VoitFiles{
				{
					File: File{
						Source: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
						Name:   "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
						Ext:    ".jpg",
					},
				},
			},
			wantVoit: []Voit{
				{
					File: File{
						Source: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
						Name:   "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
						Ext:    ".jpg",
					},
					Orig: Meta{
						VTime: VTime{Time: voitTime},
						Tags:  Tag{Items: baseTags},
						Desc:  Desc{Text: baseDesc},
					},
					Mark: Meta{
						VTime: VTime{Time: voitTime},
						Tags:  Tag{Items: baseTags},
						Desc:  Desc{Text: baseDesc},
					},
					Matched: true,
				},
			},
		},
		{
			name: "pattern: match prefer pattern strip [ptime match, stripped]",
			opts: &Opts{
				Rename: RenameOpts{
					Pattern:       "photo-ms",
					PreferPattern: true,
					Strip:         true,
				},
			},
			f: VoitFiles{
				{
					File: File{
						Source: "/tmp/2026-02-02T12.05.20.700 20260517_104536300 beach vacation -- summer vacation beach.jpg",
						Name:   "2026-02-02T12.05.20.700 20260517_104536300 beach vacation -- summer vacation beach",
						Ext:    ".jpg",
					},
				},
			},
			wantVoit: []Voit{
				{
					File: File{
						Source: "/tmp/2026-02-02T12.05.20.700 20260517_104536300 beach vacation -- summer vacation beach.jpg",
						Name:   "2026-02-02T12.05.20.700 20260517_104536300 beach vacation -- summer vacation beach",
						Ext:    ".jpg",
					},
					Orig: Meta{
						VTime: VTime{Time: voitTime},
						PTime: VTime{Time: parsedTime},
						Tags:  Tag{Items: baseTags},
						Desc:  Desc{Text: "20260517_104536300 beach vacation"},
					},
					Mark: Meta{
						VTime: VTime{Time: parsedTime},
						Tags:  Tag{Items: baseTags},
						Desc:  Desc{Text: baseDesc},
					},
					Matched: true,
				},
			},
		},
		{
			name: "set: [ptime match with explicit date set]",
			opts: &Opts{
				Rename: RenameOpts{
					Set:     "2026-03-01T20.00.00.000",
					Pattern: "set",
				},
			},
			f: VoitFiles{
				{
					File: File{
						Source: "/tmp/2026-02-02T12.05.20.700 20260517_104536300 beach vacation -- summer vacation beach.jpg",
						Name:   "2026-02-02T12.05.20.700 20260517_104536300 beach vacation -- summer vacation beach",
						Ext:    ".jpg",
					},
				},
			},
			wantVoit: []Voit{
				{
					File: File{
						Source: "/tmp/2026-02-02T12.05.20.700 20260517_104536300 beach vacation -- summer vacation beach.jpg",
						Name:   "2026-02-02T12.05.20.700 20260517_104536300 beach vacation -- summer vacation beach",
						Ext:    ".jpg",
					},
					Orig: Meta{
						VTime: VTime{Time: voitTime},
						PTime: VTime{Time: fixedSTime},
						Tags:  Tag{Items: baseTags},
						Desc:  Desc{Text: "20260517_104536300 beach vacation"},
					},
					Mark: Meta{
						VTime: VTime{Time: fixedSTime},
						Tags:  Tag{Items: baseTags},
						Desc:  Desc{Text: "20260517_104536300 beach vacation"},
					},
					Matched: true,
				},
			},
		},
		{
			name: "no desc no tags [only date and extension]",
			opts: &Opts{
				Rename: RenameOpts{
					NoDesc: true,
					NoTags: true,
				},
			},
			f: VoitFiles{
				{
					File: File{
						Source: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
						Name:   "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
						Ext:    ".jpg",
					},
				},
			},
			wantVoit: []Voit{
				{
					File: File{
						Source: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
						Name:   "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
						Ext:    ".jpg",
					},
					Orig: Meta{
						VTime: VTime{Time: voitTime},
						Tags:  Tag{Items: baseTags},
						Desc:  Desc{Text: baseDesc},
					},
					Mark: Meta{
						VTime: VTime{Time: voitTime},
						Tags:  Tag{Items: []string{}},
						Desc:  Desc{Text: ""},
					},
					Matched: true,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			tt.f.StageRename(&buf, tt.opts)

			// DeepEqual will compare memory addresses if pointers, not values.
			// Convert to value. This is required as pointers are needed to update
			// the struct in place during stageRename.
			got := make([]Voit, len(tt.f))
			for i, ptr := range tt.f {
				if ptr != nil {
					got[i] = *ptr
				}
			}

			if !reflect.DeepEqual(got, tt.wantVoit) {
				t.Errorf("\nGot Voit:  %+v\nWant Voit: %+v", got, tt.wantVoit)
			}
		})
	}
}

func TestRename(t *testing.T) {
	tests := []struct {
		name      string
		overwrite bool
		wantErr   bool
		setupFS   func(t *testing.T, baseDir string) VoitFiles
		verifyFS  func(t *testing.T, baseDir string, files VoitFiles, output string)
	}{
		{
			name:      "sanity: skip unmatched files [no file changes]",
			overwrite: false,
			setupFS: func(t *testing.T, baseDir string) VoitFiles {
				src := filepath.Join(baseDir, "unmatched.jpg")
				if err := os.WriteFile(src, []byte("data"), 0644); err != nil {
					t.Fatal(err)
				}
				return VoitFiles{
					&Voit{
						File:    File{Source: src, Ext: ".jpg"},
						Matched: false,
						Target:  filepath.Join(baseDir, "target.jpg"),
					},
				}
			},
			verifyFS: func(t *testing.T, baseDir string, files VoitFiles, output string) {
				if _, err := os.Stat(files[0].File.Source); os.IsNotExist(err) {
					t.Errorf("Expected source file to remain untouched, but it was deleted or moved")
				}
				if !strings.Contains(output, "Renamed in") {
					t.Errorf("Expected summary metric message, got: %q", output)
				}
			},
		},
		{
			name:      "sanity: default [file renamed]",
			overwrite: false,
			setupFS: func(t *testing.T, baseDir string) VoitFiles {
				src := filepath.Join(baseDir, "photo.jpg")
				if err := os.WriteFile(src, []byte("data"), 0644); err != nil {
					t.Fatal(err)
				}
				return VoitFiles{
					&Voit{
						File:    File{Source: src, Ext: ".jpg"},
						Matched: true,
						Target:  filepath.Join(baseDir, "vacation.jpg"),
					},
				}
			},
			verifyFS: func(t *testing.T, baseDir string, files VoitFiles, output string) {
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
			setupFS: func(t *testing.T, baseDir string) VoitFiles {
				src := filepath.Join(baseDir, "photo.jpg")
				tgt := filepath.Join(baseDir, "vacation.jpg")

				if err := os.WriteFile(src, []byte("new-content"), 0644); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(tgt, []byte("old-content"), 0644); err != nil {
					t.Fatal(err)
				}
				return VoitFiles{
					&Voit{
						File:    File{Source: src, Ext: ".jpg"},
						Matched: true,
						Target:  tgt,
					},
				}
			},
			verifyFS: func(t *testing.T, baseDir string, files VoitFiles, output string) {
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
			name:      "sanity: collision [error raised]",
			overwrite: false,
			wantErr:   true,
			setupFS: func(t *testing.T, baseDir string) VoitFiles {
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
				return VoitFiles{
					&Voit{
						File:    File{Source: src, Ext: ".jpg"},
						Matched: true,
						Target:  tgt,
					},
				}
			},
		},
		{
			name:      "sanity: missing source [file deleted mid-execution, error logged]",
			overwrite: false,
			setupFS: func(t *testing.T, baseDir string) VoitFiles {
				return VoitFiles{
					&Voit{
						File:    File{Source: filepath.Join(baseDir, "non-existent.jpg"), Ext: ".jpg"},
						Matched: true,
						Target:  filepath.Join(baseDir, "output.jpg"),
					},
				}
			},
			verifyFS: func(t *testing.T, baseDir string, files VoitFiles, output string) {
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
			err := files.Rename(&buf, tt.overwrite, true)

			if tt.wantErr && err == nil {
				t.Errorf("\nGot:  %q\nWant: nil\n", err)
			}

			if tt.verifyFS != nil {
				tt.verifyFS(t, tmpDir, files, buf.String())
			}
		})
	}
}

// Simple logic requires no testing.
func TestTimeActionNOOP(t *testing.T) {}
