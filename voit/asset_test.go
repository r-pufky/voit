package voit

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

// Assets mock for testing.
type MockAssets struct {
	Assets // Embed interface to use real func when not stubbed out.

	ResolveCollisionsStub func(w io.Writer, cfg ...Config)
	DisplayPendingStub    func(w io.Writer, cfg ...Config) int
	ConfirmStub           func(w io.Writer, r io.Reader) bool
	RenameErr             error
	CallCount             int
}

func (m *MockAssets) ResolveCollisions(w io.Writer, cfg ...Config) {
	if m.ResolveCollisionsStub != nil {
		m.ResolveCollisionsStub(w, cfg...)
	}
}

func (m *MockAssets) DisplayPending(w io.Writer, cfg ...Config) int {
	if m.DisplayPendingStub != nil {
		return m.DisplayPendingStub(w, cfg...)
	}
	return 0
}

var _ Assets = (*MockAssets)(nil) // Compile time check to validate interface.

func TestAssetsMaxWidth(t *testing.T) {
	tests := []struct {
		name    string
		assets  Assets
		wantMax int
	}{
		{
			name:    "empty map returns 0",
			assets:  NewAssets(),
			wantMax: 0,
		},
		{
			name: "largest width found",
			assets: &AssetImpl{
				m: map[string]Voit{
					"small":   &VoitImpl{File: File{Name: "small", Ext: ".jpg"}},
					"longest": &VoitImpl{File: File{Name: "longest", Ext: ".jpg"}},
					"medium":  &VoitImpl{File: File{Name: "medium", Ext: ".jpg"}},
				},
				width: 11,
			},
			wantMax: 11,
		},
		{
			name: "sidecar width is included",
			assets: &AssetImpl{
				m: map[string]Voit{
					"small": &VoitImpl{File: File{Name: "small", Ext: ".jpg"}},
					"longest": &VoitImpl{
						File:    File{Name: "longest", Ext: ".jpg"},
						Sidecar: File{Name: "sidecar", Ext: ".jpg.xmp"},
					},
				},
				width: 15,
			},
			wantMax: 15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.assets.MaxWidth(); got != tt.wantMax {
				t.Errorf("\nGot:  %d\nWant: %d\n", got, tt.wantMax)
			}
		})
	}
}

func TestAssetsLoadDir(t *testing.T) {
	t.Parallel()

	t.Run("uninitialized assets [error]", func(t *testing.T) {
		t.Parallel()
		var uninitializedAssets *AssetImpl = nil
		err := uninitializedAssets.LoadDir("/tmp/*", Config{})
		if err == nil || !strings.Contains(err.Error(), "asset is uninitialized") {
			t.Errorf("\nGot:  %v\nWant: expected uninitialized error\n", err)
		}
	})

	t.Run("invalid glob pattern [error]", func(t *testing.T) {
		t.Parallel()
		assets := NewAssets()
		err := assets.LoadDir("[-]", Config{})
		if err == nil || !strings.Contains(err.Error(), "invalid glob pattern") {
			t.Errorf("\nGot:  %v\nWant: expected invalid glob error\n", err)
		}
	})
}

// filepath.Glob cannot be mocked without monkey patching. Use tempDir.
func TestAssetsLoadDirDisk(t *testing.T) {
	t.Run("bare directory globbed and sub-directory skipped", func(t *testing.T) {
		tmpDir := t.TempDir()
		mainFile := filepath.Join(tmpDir, "photo1.jpg")
		subDir := filepath.Join(tmpDir, "nested_dir")

		if err := os.WriteFile(mainFile, []byte("image data"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.Mkdir(subDir, 0755); err != nil {
			t.Fatal(err)
		}

		assets := NewAssets().(*AssetImpl)
		now := time.Now().UTC()

		cfg := Config{
			FS: MockFS{
				MockStat: func(name string) (os.FileInfo, error) {
					// Simulate directory state matching paths.
					if name == tmpDir || name == subDir {
						return mockFileInfo{isDir: true, modTime: now}, nil
					}
					return mockFileInfo{isDir: false, modTime: now}, nil
				},
			},
			Unix: MockUnix{
				MockStatx: func(dirfd int, path string, flags int, mask int, stat *unix.Statx_t) error {
					stat.Mask = unix.STATX_BTIME
					stat.Btime.Sec = now.Unix()
					return nil
				},
			},
		}

		err := assets.LoadDir(tmpDir, cfg)
		if err != nil {
			t.Fatalf("unexpected scanning error: %v", err)
		}

		for _, v := range assets.m {
			filePath, sidecarPath := v.Abs(cfg)
			if strings.Contains(filePath, "nested_dir") || strings.Contains(sidecarPath, "nested_dir") {
				t.Error("\nsub-directory not skipped and placed in Voit.\n")
			}
		}

		if l := len(assets.m); l != 1 {
			t.Errorf("\nGot %d\nWant: 1\n", l)
		}
	})

	t.Run("file and sidecar added as a single asset", func(t *testing.T) {
		tmpDir := t.TempDir()
		mainFile := filepath.Join(tmpDir, "photo1.jpg")
		sidecarFile := filepath.Join(tmpDir, "photo1.jpg.xmp")

		if err := os.WriteFile(mainFile, []byte("image data"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(sidecarFile, []byte("xmp data"), 0644); err != nil {
			t.Fatal(err)
		}

		assets := NewAssets().(*AssetImpl)
		now := time.Now().UTC()

		cfg := Config{
			FS: MockFS{
				MockStat: func(name string) (os.FileInfo, error) {
					return mockFileInfo{isDir: false, modTime: now}, nil
				},
			},
			Unix: MockUnix{
				MockStatx: func(dirfd int, path string, flags int, mask int, stat *unix.Statx_t) error {
					stat.Mask = unix.STATX_BTIME
					stat.Btime.Sec = now.Unix()
					return nil
				},
			},
		}

		err := assets.LoadDir(filepath.Join(tmpDir, "*"), cfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		wantKey := filepath.Join(tmpDir, "photo1")
		v, exists := assets.m[wantKey]
		if !exists {
			t.Fatalf("key %q missing from assets map", wantKey)
		}

		extFile, extSidecar := v.SourceExt()

		if extFile != ".jpg" {
			t.Errorf("\nGot: %q\nWant: '.jpg'\n", extFile)
		}
		if extSidecar != ".jpg.xmp" {
			t.Errorf("\nGot: %q\nWant: '.jpg.xmp'\n", extSidecar)
		}
	})

	t.Run("files found but none match datetime formats [error]", func(t *testing.T) {
		tmpDir := t.TempDir()
		invalidFile := filepath.Join(tmpDir, "not-a-valid-date-format.jpg")

		if err := os.WriteFile(invalidFile, []byte("image data"), 0644); err != nil {
			t.Fatal(err)
		}

		assets := NewAssets().(*AssetImpl)
		now := time.Now().UTC()

		cfg := Config{
			FS: MockFS{
				MockStat: func(name string) (os.FileInfo, error) {
					return mockFileInfo{isDir: true, modTime: now}, nil
				},
			},
			Unix: MockUnix{
				MockStatx: func(dirfd int, path string, flags int, mask int, stat *unix.Statx_t) error {
					stat.Mask = unix.STATX_BTIME
					return nil
				},
			},
		}

		err := assets.LoadDir(filepath.Join(tmpDir, "*"), cfg)

		if err == nil {
			t.Fatal("expected error for empty asset map processing, got nil")
		}

		wantErr := "No files matched the known datetime formats (is globbing quoted?)"
		if !strings.Contains(err.Error(), wantErr) {
			t.Errorf("\nGot:  %q\nWant: %q\n", err.Error(), wantErr)
		}
	})

	t.Run("width calculated correctly", func(t *testing.T) {
		tmpDir := t.TempDir()
		mainFile := filepath.Join(tmpDir, "photo1.jpg")

		if err := os.WriteFile(mainFile, []byte("image data"), 0644); err != nil {
			t.Fatal(err)
		}

		assets := NewAssets().(*AssetImpl)
		now := time.Now().UTC()

		cfg := Config{
			FS: MockFS{
				MockStat: func(name string) (os.FileInfo, error) {
					return mockFileInfo{isDir: false, modTime: now}, nil
				},
			},
			Unix: MockUnix{
				MockStatx: func(dirfd int, path string, flags int, mask int, stat *unix.Statx_t) error {
					stat.Mask = unix.STATX_BTIME
					stat.Btime.Sec = now.Unix()
					return nil
				},
			},
		}

		if err := assets.LoadDir(filepath.Join(tmpDir, "*"), cfg); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if assets.MaxWidth() != 10 {
			t.Errorf("\nWidth\nGot:  %d\nWant: 10\n", assets.MaxWidth())
		}
	})
}

func TestAssetsHasAssetDestCollision(t *testing.T) {
	t.Parallel()

	cfg := NewConfig(Config{
		Voit: VoitConfig{VFormat: "2006-01-02"},
		FS: MockFS{
			MockStat: func(string) (os.FileInfo, error) {
				return nil, os.ErrNotExist
			},
		},
	})

	voit := &VoitImpl{File: File{Path: "dir", Ext: ".jpg"}}
	fileCollision := &VoitImpl{File: File{Path: "dir", Ext: ".jpg"}}
	sidecarCollision := &VoitImpl{File: File{Path: "dir", Ext: ".jpg.xmp"}}

	tests := []struct {
		name      string
		assets    Assets
		targetKey string
		want      bool
	}{
		{
			name: "missing key [false]",
			assets: &AssetImpl{
				m: map[string]Voit{
					"item1": voit,
				},
			},
			targetKey: "non_existent_key",
			want:      false,
		},
		{
			name: "key does not check self [false]",
			assets: &AssetImpl{
				m: map[string]Voit{
					"primary": voit,
				},
			},
			targetKey: "primary",
			want:      false,
		},
		{
			name: "file collision [true]",
			assets: &AssetImpl{
				m: map[string]Voit{
					"primary": voit,
					"sibling": fileCollision,
				},
			},
			targetKey: "primary",
			want:      true,
		},
		{
			name: "sidecar collision [true]",
			assets: &AssetImpl{
				m: map[string]Voit{
					"primary": voit,
					"sibling": sidecarCollision,
				},
			},
			targetKey: "primary",
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.assets.HasAssetDestCollision(tt.targetKey, cfg)
			if got != tt.want {
				t.Errorf("\nKey: %q\nGot:  %t\nWant: %t\n", tt.targetKey, got, tt.want)
			}
		})
	}
}

func TestAssetResolveCollisions(t *testing.T) {
	voitTime := time.Date(2026, time.February, 2, 12, 5, 20, 0, time.UTC)

	tests := []struct {
		name     string
		assets   *AssetImpl
		mockStat func() func(string) (os.FileInfo, error)
		verify   func(t *testing.T, assets *AssetImpl)
	}{
		{
			name: "no collisions",
			assets: &AssetImpl{
				m: map[string]Voit{
					"summer": &VoitImpl{
						File: File{Path: "/dir", Ext: ".jpg", Name: "file1"},
						Dest: Meta{VTime: VTime{Time: voitTime}, Desc: Desc{Text: "summer"}},
					},
				},
			},
			mockStat: func() func(string) (os.FileInfo, error) {
				return func(name string) (os.FileInfo, error) {
					return nil, os.ErrNotExist
				}
			},
			verify: func(t *testing.T, assets *AssetImpl) {
				name, _ := assets.m["summer"].Abs()
				if name != "/dir/2026-02-02T12.05.20.000 summer.jpg" {
					t.Errorf("\nGot:  %q\nWant: /dir/2026-02-02T12.05.20.000 summer.jpg\n", name)
				}
			},
		},
		{
			name: "fs collision",
			assets: &AssetImpl{
				m: map[string]Voit{
					"summer": &VoitImpl{
						File: File{Path: "/dir", Ext: ".jpg", Name: "file1"},
						Dest: Meta{VTime: VTime{Time: voitTime}, Desc: Desc{Text: "summer"}},
					},
				},
			},
			mockStat: func() func(string) (os.FileInfo, error) {
				calls := 0
				return func(name string) (os.FileInfo, error) {
					calls++
					if calls == 1 {
						return mockFileInfo{}, nil // File exists.
					}
					return nil, os.ErrNotExist
				}
			},
			verify: func(t *testing.T, assets *AssetImpl) {
				name, _ := assets.m["summer"].Abs()
				if name != "/dir/2026-02-02T12.05.20.000 summer 1.jpg" {
					t.Errorf("\nGot:  %q\nWant: /dir/2026-02-02T12.05.20.000 summer 1.jpg\n", name)
				}
			},
		},
		{
			name: "fs collision no desc",
			assets: &AssetImpl{
				m: map[string]Voit{
					"summer": &VoitImpl{
						File: File{Path: "/dir", Ext: ".jpg", Name: "file1"},
						Dest: Meta{VTime: VTime{Time: voitTime}, Desc: Desc{Text: ""}},
					},
				},
			},
			mockStat: func() func(string) (os.FileInfo, error) {
				calls := 0
				return func(name string) (os.FileInfo, error) {
					calls++
					if calls == 1 {
						return mockFileInfo{}, nil // File exists.
					}
					return nil, os.ErrNotExist
				}
			},
			verify: func(t *testing.T, assets *AssetImpl) {
				name, _ := assets.m["summer"].Abs()
				if name != "/dir/2026-02-02T12.05.20.000 1.jpg" {
					t.Errorf("\nGot:  %q\nWant: /dir/2026-02-02T12.05.20.000 1.jpg\n", name)
				}
			},
		},
		{
			name: "asset collision",
			assets: &AssetImpl{
				m: map[string]Voit{
					"asset1": &VoitImpl{
						File: File{Path: "/dir", Ext: ".jpg", Name: "file1"},
						Dest: Meta{VTime: VTime{Time: voitTime}, Desc: Desc{Text: "summer"}},
					},
					"asset2": &VoitImpl{
						File: File{Path: "/dir", Ext: ".jpg", Name: "file2"},
						Dest: Meta{VTime: VTime{Time: voitTime}, Desc: Desc{Text: "summer"}},
					},
				},
			},
			mockStat: func() func(string) (os.FileInfo, error) {
				return func(name string) (os.FileInfo, error) {
					return nil, os.ErrNotExist
				}
			},
			verify: func(t *testing.T, assets *AssetImpl) {
				wantFirst := "/dir/2026-02-02T12.05.20.000 summer.jpg"
				wantSecond := "/dir/2026-02-02T12.05.20.000 summer 1.jpg"

				// Map order random, check both cases.
				d1, _ := assets.m["asset1"].Abs()
				d2, _ := assets.m["asset2"].Abs()
				if (d1 == wantFirst && d2 == wantSecond) || (d1 == wantSecond && d2 == wantFirst) {
					return
				}

				t.Errorf("\nAssets should not collide\nasset1: %q\nasset2: %q\n", d1, d2)
			},
		},
		{
			name: "asset and fs collision [summer 1, summer 2]",
			assets: &AssetImpl{
				m: map[string]Voit{
					"asset1": &VoitImpl{
						File: File{Path: "/dir", Ext: ".jpg", Name: "file1"},
						Dest: Meta{VTime: VTime{Time: voitTime}, Desc: Desc{Text: "summer"}},
					},
					"asset2": &VoitImpl{
						File: File{Path: "/dir", Ext: ".jpg", Name: "file2"},
						Dest: Meta{VTime: VTime{Time: voitTime}, Desc: Desc{Text: "summer"}},
					},
				},
			},
			mockStat: func() func(string) (os.FileInfo, error) {
				pathChecks := make(map[string]int)
				var baseFile string

				return func(name string) (os.FileInfo, error) {
					pathChecks[name]++

					if baseFile == "" {
						baseFile = name // First path is always FS file.
					}

					if name == baseFile {
						return mockFileInfo{}, nil // First file always exists.
					}

					return nil, os.ErrNotExist // Other FS files do not exist.
				}
			},
			verify: func(t *testing.T, assets *AssetImpl) {
				wantFirst := "/dir/2026-02-02T12.05.20.000 summer 1.jpg"
				wantSecond := "/dir/2026-02-02T12.05.20.000 summer 2.jpg"

				// Map order random, check both cases.
				d1, _ := assets.m["asset1"].Abs()
				d2, _ := assets.m["asset2"].Abs()
				if (d1 == wantFirst && d2 == wantSecond) || (d1 == wantSecond && d2 == wantFirst) {
					return
				}

				t.Errorf("\nasset should not collide with FS or each other\nasset1: %q\nasset2: %q\n", d1, d2)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			cfg := NewConfig()
			cfg.FS = MockFS{
				MockStat: tt.mockStat(),
			}
			cfg.Voit.Verbose = true

			tt.assets.ResolveCollisions(&buf, cfg)
			tt.verify(t, tt.assets)
		})
	}
}

func TestDisplayPending(t *testing.T) {
	t.Parallel()

	voitTime := time.Date(2026, time.February, 2, 12, 5, 20, 700000000, time.UTC)
	dest := Meta{VTime: VTime{Time: voitTime}}

	files := &AssetImpl{
		m: map[string]Voit{
			"img1.jpg": &VoitImpl{
				File: File{
					Name: "img1",
					Ext:  ".jpg",
				},
				Dest:    dest,
				Matched: true,
			},
			"a.jpg": &VoitImpl{
				File: File{
					Name: "a",
					Ext:  ".jpg",
				},
				Dest:    dest,
				Matched: true,
			},
			"sidecar.jpg": &VoitImpl{
				File: File{
					Name: "sidecar",
					Ext:  ".jpg",
				},
				Sidecar: File{
					Name: "sidecar",
					Ext:  ".jpg.xmp",
				},
				Dest:    dest,
				Matched: true,
			},
		},
		width: 15,
	}

	buf := &bytes.Buffer{}

	count := files.DisplayPending(buf, Config{})

	if count != 4 { // Voit assets with File and Sidecar count as 2.
		t.Errorf("\nCount\nGot:  %d\nWant: 4\n", count)
	}
}

func TestPromptRename(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		assets     *AssetImpl
		config     Config
		r          *strings.Reader
		wantOutput string
		wantErrStr string
	}{
		{
			name:       "no matches",
			assets:     NewAssets().(*AssetImpl),
			config:     Config{},
			r:          strings.NewReader(""),
			wantOutput: "No files matched proposed changes.",
		},
		{
			name: "displaypending, overwrite, rename",
			assets: &AssetImpl{
				m: map[string]Voit{
					"file1.jpg": &VoitImpl{File: File{Name: "file1", Ext: ".jpg"}, Matched: true},
				},
			},
			config:     Config{Voit: VoitConfig{Overwrite: true}},
			r:          strings.NewReader(""),
			wantOutput: "Proposed changes (OVERWRITE ENABLED): 1 file(s).",
		},
		{
			name: "displaypending, yes, no prompt",
			assets: &AssetImpl{
				m: map[string]Voit{
					"file1.jpg": &MockVoit{
						VoitImpl: &VoitImpl{
							File:    File{Name: "file1", Ext: ".jpg"},
							Matched: true,
						},
						RenameStub: func(w io.Writer, cfg ...Config) error { return nil },
					},
				},
			},
			config:     Config{Voit: VoitConfig{Yes: true}},
			r:          strings.NewReader(""),
			wantOutput: "Proposed changes: 1 file(s).",
		},
		{
			name: "displaypending, prompt, user aborts",
			assets: &AssetImpl{
				m: map[string]Voit{
					"file1.jpg": &VoitImpl{File: File{Name: "file1", Ext: ".jpg"}, Matched: true},
				},
			},
			config:     Config{},
			r:          strings.NewReader("n\n"),
			wantOutput: "Operation aborted by user.",
		},
		{
			name: "rename operation returns execution error",
			assets: &AssetImpl{
				m: map[string]Voit{
					"err_file.jpg": &MockVoit{
						VoitImpl: &VoitImpl{
							File:    File{Name: "err_file", Ext: ".jpg"},
							Matched: true,
						},
						RenameStub: func(w io.Writer, cfg ...Config) error {
							return errors.New("filesystem I/O failure")
						},
					},
				},
			},
			config:     Config{Voit: VoitConfig{Yes: true}},
			r:          strings.NewReader(""),
			wantOutput: "Proposed changes: 1 file(s).",
			wantErrStr: "filesystem I/O failure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var outBuf bytes.Buffer
			err := tt.assets.PromptRename(&outBuf, tt.r, tt.config)

			if tt.wantErrStr != "" {
				if err == nil || !strings.HasPrefix(err.Error(), tt.wantErrStr) {
					t.Errorf("\nGot:         %v\nWant prefix: %q\n", err, tt.wantErrStr)
				}
			} else if err != nil {
				t.Errorf("\nunexpected error: %v\n", err)
			}

			outputStr := outBuf.String()
			if tt.wantOutput != "" && !strings.Contains(outputStr, tt.wantOutput) {
				t.Errorf("\nmissing substring in output:\nGot:\n%s\nWant included: %q\n", outputStr, tt.wantOutput)
			}
		})
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
				t.Errorf("\nGot:  %v\nWant: %v\n", got, tt.want)
			}
		})
	}
}

func TestTimeActionNOOP(t *testing.T) {}
