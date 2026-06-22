package voit

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestValidate(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory for test: %v", err)
	}

	tests := []struct {
		name        string
		input       Opts
		wantErr     bool
		errContains string
		checkSource func(t *testing.T, result string)
	}{
		{
			name: "sanity: relative abs source [abs source resolved]",
			input: Opts{
				SpanSep:   "--",
				DescSep:   " ",
				TagSep:    " -- ",
				AbsSource: ".",
				Tag: TagOpts{
					SyncFolder: "➔",
					SyncSpace:  "⠀",
				},
			},
			wantErr: false,
			checkSource: func(t *testing.T, result string) {
				if result != wd {
					t.Errorf("\nGot %s\nWant: %s\n", result, wd)
				}
			},
		},
		{
			name: "sanity: empty abs source [abs source resolved to cwd]",
			input: Opts{
				SpanSep:   "--",
				DescSep:   " ",
				TagSep:    " -- ",
				AbsSource: "",
				Tag: TagOpts{
					SyncFolder: "➔",
					SyncSpace:  "⠀",
				},
			},
			wantErr: false,
			checkSource: func(t *testing.T, result string) {
				if result != wd {
					t.Errorf("\nGot: %s\nWant: %s\n", result, wd)
				}
			},
		},
		{
			name: "sanity: SyncFolder collision [error raised]",
			input: Opts{
				SpanSep:   "-",
				DescSep:   " ",
				TagSep:    " -- ",
				AbsSource: "",
				Tag: TagOpts{
					SyncFolder: "-",
					SyncSpace:  "⠀",
				},
			},
			wantErr:     true,
			errContains: "tag-folder must be unique non-whitespace character",
		},
		{
			name: "sanity: SyncFolder whitespace [error raised]",
			input: Opts{
				SpanSep:   "-",
				DescSep:   " ",
				TagSep:    " -- ",
				AbsSource: "",
				Tag: TagOpts{
					SyncFolder: " ",
					SyncSpace:  "⠀",
				},
			},
			wantErr:     true,
			errContains: "tag-folder must be unique non-whitespace character",
		},
		{
			name: "sanity: SyncSpace collision [error raised]",
			input: Opts{
				SpanSep:   "-",
				DescSep:   " ",
				TagSep:    " -- ",
				AbsSource: "",
				Tag: TagOpts{
					SyncFolder: "➔",
					SyncSpace:  "-",
				},
			},
			wantErr:     true,
			errContains: "tag-space must be unique non-whitespace character",
		},
		{
			name: "sanity: SyncSpace whitespace [error raised]",
			input: Opts{
				SpanSep:   "--",
				DescSep:   " ",
				TagSep:    " -- ",
				AbsSource: "",
				Tag: TagOpts{
					SyncFolder: "➔",
					SyncSpace:  " ",
				},
			},
			wantErr:     true,
			errContains: "tag-space must be unique non-whitespace character",
		},
		// Verbose flag simple, not tested.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := tt.input

			err := opts.Validate(os.Stdout)

			if (err != nil) != tt.wantErr {
				t.Fatalf("\nGot:  %v\nWant: %v\n", err, tt.wantErr)
			}

			if tt.wantErr && err != nil {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("\nGot:             %q\nWant to contain: %q\n", tt.errContains, err.Error())
				}
			}

			if !tt.wantErr && tt.checkSource != nil {
				tt.checkSource(t, opts.AbsSource)
			}
		})
	}
}

func TestUpdateFromOpts(t *testing.T) {
	tests := []struct {
		name string
		opts *Opts
		want Config
	}{
		{
			name: "sanity: empty defaults [pattern: voit]",
			opts: &Opts{},
			want: NewConfig(Config{Voit: VoitConfig{Pattern: "voit"}}),
		},
		{
			name: "sanity: non-default values",
			opts: &Opts{
				TagSep: "#",
				Tag: TagOpts{
					SyncMetaFolder: "_",
					SyncFolder:     ">",
					SyncSpace:      "_",
					SyncKeepFolder: true,
					SyncKeepSpace:  true,
				},
				DescSep: "|",
				SpanSep: "->",
				VFormat: "15:04",
				Rename: RenameOpts{
					Pattern: "photo-ms",
				},
				Lower: true,
			},
			want: Config{
				Voit: VoitConfig{
					DescSep: "|",
					TagSep:  "#",
					SpanSep: "->",
					VFormat: "15:04",
					Pattern: "photo-ms",
					Set:     "",
					Verbose: false,
					Lower:   true,
				},
				Sync: SyncConfig{
					MetaFolder: "_",
					Folder:     ">",
					Space:      "_",
					KeepFolder: true,
					KeepSpace:  true,
				},
				FS:   RealFS{},
				Unix: RealUnix{},
			},
		},
		{
			name: "sanity: partial set [default values used elsewhere]",
			opts: &Opts{
				VFormat: "2006",
			},
			want: NewConfig(Config{
				Voit: VoitConfig{
					VFormat: "2006",
					Pattern: "voit",
				},
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Config{}.UpdateFromOpts(tt.opts)

			// Normalize unused mock options so DeepEqual tests correctly.
			got.FS = nil
			got.Unix = nil
			tt.want.FS = nil
			tt.want.Unix = nil

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("\nGot:  %+v\nWant: %+v\n", got, tt.want)
			}
		})
	}
}

func TestStageRename(t *testing.T) {
	type testCase struct {
		name     string
		opts     *Opts
		input    *VoitImpl
		wantVoit VoitImpl
	}

	fixedSTime := time.Date(2026, time.March, 1, 20, 0, 0, 0, time.UTC)
	parsedTime := time.Date(2026, time.May, 17, 10, 45, 36, 300000000, time.UTC)
	voitTime := time.Date(2026, time.February, 2, 12, 5, 20, 700000000, time.UTC)
	baseTags := []string{"summer", "vacation", "beach"}
	baseDesc := "beach vacation"

	tests := []testCase{
		{
			name: "sanity: no match [ZeroTime, matched false]",
			opts: &Opts{
				Rename: RenameOpts{},
			},
			input: &VoitImpl{
				File: File{
					Path: "/tmp/invalid_format.txt",
					Name: "invalid_format",
					Ext:  ".txt",
				},
				Orig: Meta{
					Desc: Desc{Text: "invalid_format"},
				},
			},
			wantVoit: VoitImpl{
				File: File{
					Path: "/tmp/invalid_format.txt",
					Name: "invalid_format",
					Ext:  ".txt",
				},
				Orig: Meta{
					Desc: Desc{Text: "invalid_format"},
				},
				Dest: Meta{
					Desc: Desc{Text: "invalid_format"},
				},
				Matched: false,
			},
		},
		{
			name: "sanity: default [vtime match]",
			opts: &Opts{
				Rename: RenameOpts{},
			},
			input: &VoitImpl{
				File: File{
					Path: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
					Name: "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
					Ext:  ".jpg",
				},
				Orig: Meta{
					VTime: VTime{Time: voitTime},
					Tags:  Tag{Items: baseTags},
					Desc:  Desc{Text: baseDesc},
				},
			},
			wantVoit: VoitImpl{
				File: File{
					Path: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
					Name: "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
					Ext:  ".jpg",
				},
				Orig: Meta{
					VTime: VTime{Time: voitTime},
					Tags:  Tag{Items: baseTags},
					Desc:  Desc{Text: baseDesc},
				},
				Dest: Meta{
					VTime: VTime{Time: voitTime},
					Tags:  Tag{Items: baseTags},
					Desc:  Desc{Text: baseDesc},
				},
				Matched: true,
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
			input: &VoitImpl{
				File: File{
					Path: "/tmp/2026-02-02T12.05.20.700 20260517_104536300 beach vacation -- summer vacation beach.jpg",
					Name: "2026-02-02T12.05.20.700 20260517_104536300 beach vacation -- summer vacation beach",
					Ext:  ".jpg",
				},
				Orig: Meta{
					VTime: VTime{Time: voitTime},
					PTime: VTime{Time: parsedTime},
					Tags:  Tag{Items: baseTags},
					Desc:  Desc{Text: "20260517_104536300 beach vacation"},
				},
			},
			wantVoit: VoitImpl{
				File: File{
					Path: "/tmp/2026-02-02T12.05.20.700 20260517_104536300 beach vacation -- summer vacation beach.jpg",
					Name: "2026-02-02T12.05.20.700 20260517_104536300 beach vacation -- summer vacation beach",
					Ext:  ".jpg",
				},
				Orig: Meta{
					VTime: VTime{Time: voitTime},
					PTime: VTime{Time: parsedTime},
					Tags:  Tag{Items: baseTags},
					Desc:  Desc{Text: "20260517_104536300 beach vacation"},
				},
				Dest: Meta{
					VTime: VTime{Time: parsedTime},
					PTime: VTime{Time: time.Time{}},
					Tags:  Tag{Items: baseTags},
					Desc:  Desc{Text: baseDesc},
				},
				Matched: true,
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
			input: &VoitImpl{
				File: File{
					Path: "/tmp/2026-02-02T12.05.20.700 20260517_104536300 beach vacation -- summer vacation beach.jpg",
					Name: "2026-02-02T12.05.20.700 20260517_104536300 beach vacation -- summer vacation beach",
					Ext:  ".jpg",
				},
				Orig: Meta{
					VTime: VTime{Time: voitTime},
					PTime: VTime{Time: fixedSTime},
					Tags:  Tag{Items: baseTags},
					Desc:  Desc{Text: "20260517_104536300 beach vacation"},
				},
			},
			wantVoit: VoitImpl{
				File: File{
					Path: "/tmp/2026-02-02T12.05.20.700 20260517_104536300 beach vacation -- summer vacation beach.jpg",
					Name: "2026-02-02T12.05.20.700 20260517_104536300 beach vacation -- summer vacation beach",
					Ext:  ".jpg",
				},
				Orig: Meta{
					VTime: VTime{Time: voitTime},
					PTime: VTime{Time: fixedSTime},
					Tags:  Tag{Items: baseTags},
					Desc:  Desc{Text: "20260517_104536300 beach vacation"},
				},
				Dest: Meta{
					VTime: VTime{Time: fixedSTime},
					PTime: VTime{Time: time.Time{}},
					Tags:  Tag{Items: baseTags},
					Desc:  Desc{Text: "20260517_104536300 beach vacation"},
				},
				Matched: true,
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
			input: &VoitImpl{
				File: File{
					Path: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
					Name: "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
					Ext:  ".jpg",
				},
				Orig: Meta{
					VTime: VTime{Time: voitTime},
					Tags:  Tag{Items: baseTags},
					Desc:  Desc{Text: baseDesc},
				},
			},
			wantVoit: VoitImpl{
				File: File{
					Path: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
					Name: "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
					Ext:  ".jpg",
				},
				Orig: Meta{
					VTime: VTime{Time: voitTime},
					Tags:  Tag{Items: baseTags},
					Desc:  Desc{Text: baseDesc},
				},
				Dest: Meta{
					VTime: VTime{Time: voitTime},
					Tags:  Tag{Items: []string{}},
					Desc:  Desc{Text: ""},
				},
				Matched: true,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer

			mockAssets := &AssetImpl{
				m: map[string]Voit{
					"test_asset": tc.input,
				},
			}

			StageRename(&buf, mockAssets, tc.opts)

			if !reflect.DeepEqual(*tc.input, tc.wantVoit) {
				t.Errorf("\nGot Voit:  %+v\nWant Voit: %+v\n", *tc.input, tc.wantVoit)
			}
		})
	}
}

func TestStageTags(t *testing.T) {
	type testCase struct {
		name         string
		tags         []string
		opts         Opts
		wantDestTags []string
		wantMatched  bool
	}

	origTags := []string{"summer", "vacation", "beach"}

	tests := []testCase{
		{
			name: "delete [all tags removed]",
			tags: slices.Clone(origTags),
			opts: Opts{
				Tag: TagOpts{
					Select: []string{},
					Delete: true,
				},
			},
			wantDestTags: []string{},
			wantMatched:  true,
		},
		{
			name: "add [adds additional]",
			tags: slices.Clone(origTags),
			opts: Opts{
				Tag: TagOpts{
					Select: []string{},
					Add:    []string{"additional"},
				},
			},
			wantDestTags: []string{"summer", "vacation", "beach", "additional"},
			wantMatched:  true,
		},
		{
			name: "remove [remove vacation]",
			tags: slices.Clone(origTags),
			opts: Opts{
				Tag: TagOpts{
					Select: []string{},
					Remove: []string{"vacation"},
				},
			},
			wantDestTags: []string{"summer", "beach"},
			wantMatched:  true,
		},
		{
			name: "set [explicitly set family, europe]",
			tags: slices.Clone(origTags),
			opts: Opts{
				Tag: TagOpts{
					Select: []string{},
					Set:    []string{"family", "europe"},
				},
			},
			wantDestTags: []string{"family", "europe"},
			wantMatched:  true,
		},
		{
			name: "match [match summer tags, add europe]",
			tags: slices.Clone(origTags),
			opts: Opts{
				Tag: TagOpts{
					Select: []string{"summer"},
					Add:    []string{"europe"},
				},
			},
			wantDestTags: []string{"summer", "vacation", "beach", "europe"},
			wantMatched:  true,
		},
		{
			name: "no match [match nonexistent_tag, no changes]",
			tags: slices.Clone(origTags),
			opts: Opts{
				Tag: TagOpts{
					Select: []string{"nonexistent_tag"},
					Add:    []string{"should_not_be_added"},
				},
			},
			wantDestTags: slices.Clone(origTags),
			wantMatched:  false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			voitImpl := &VoitImpl{
				Orig: Meta{
					Tags: Tag{
						Items: tc.tags,
					},
				},
			}

			mockAssets := &AssetImpl{
				m: map[string]Voit{
					"test_asset": voitImpl,
				},
			}

			StageTags(mockAssets, &tc.opts)

			if !slices.Equal(voitImpl.Orig.Tags.Items, origTags) {
				t.Errorf("\nOrig tags mismatch\nGot:  %v\nWant: %v\n", voitImpl.Orig.Tags.Items, origTags)
			}

			if voitImpl.Matched != tc.wantMatched {
				t.Errorf("\nMatched flag\nGot:  %v\nWant: %v\n", voitImpl.Matched, tc.wantMatched)
			}

			if !slices.Equal(voitImpl.Dest.Tags.Items, tc.wantDestTags) {
				t.Errorf("\nDest tags mismatch\nGot:  %v\nWant: %v\n", voitImpl.Dest.Tags.Items, tc.wantDestTags)
			}
		})
	}
}

func TestStageXMP(t *testing.T) {
	vTime := time.Date(2026, time.February, 2, 12, 5, 20, 700000000, time.UTC)
	bDesc := "beach vacation"

	validDigikamXMP := `<?xml version="1.0" encoding="UTF-8"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about=""
    xmlns:digiKam="http://www.digikam.org/ns/1.0/">
   <digiKam:TagsList>
    <rdf:Seq>
     <rdf:li>summer</rdf:li>
     <rdf:li>vacation</rdf:li>
     <rdf:li>beach</rdf:li>
     <rdf:li>place/location/sidecar name</rdf:li>
    </rdf:Seq>
   </digiKam:TagsList>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`

	emptyDigikamXMP := `<?xml version="1.0" encoding="UTF-8"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about=""
    xmlns:digiKam="http://www.digikam.org/ns/1.0/">
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`

	type fileSetup struct {
		name    string
		ext     string
		XMPData string
	}

	tests := []struct {
		name       string
		opts       *Opts
		filesSetup []fileSetup
		sidecar    File
		wantVoit   VoitImpl
	}{
		{
			name: "sanity: tagged sidecar match [matching tags synchronized]",
			opts: &Opts{
				Tag: TagOpts{
					SyncXMP:        true,
					SyncMetaFolder: "place/",
					SyncFolder:     "➔",
					SyncSpace:      "⠀",
					SyncKeepFolder: true,
					SyncKeepSpace:  true,
				},
			},
			filesSetup: []fileSetup{
				{name: "2026-02-02T12.05.20.700 beach vacation", ext: ".jpg.xmp", XMPData: validDigikamXMP},
			},
			sidecar: File{
				Path: "/mock/path",
				Name: "2026-02-02T12.05.20.700 beach vacation",
				Ext:  ".jpg.xmp",
			},
			wantVoit: VoitImpl{
				Sidecar: File{
					Path: "/mock/path",
					Name: "2026-02-02T12.05.20.700 beach vacation",
					Ext:  ".jpg.xmp",
				},
				Orig: Meta{
					VTime: VTime{Time: vTime},
					Desc:  Desc{Text: bDesc},
				},
				Matched: true,
			},
		},
		{
			name: "sanity: untagged sidecar no match [no tags found]",
			opts: &Opts{
				Tag: TagOpts{
					SyncXMP: true,
				},
			},
			filesSetup: []fileSetup{
				{name: "2026-02-02T12.05.20.700 beach vacation", ext: ".jpg.xmp", XMPData: emptyDigikamXMP},
			},
			sidecar: File{
				Path: "/mock/path",
				Name: "2026-02-02T12.05.20.700 beach vacation",
				Ext:  ".jpg.xmp",
			},
			wantVoit: VoitImpl{
				Sidecar: File{
					Path: "/mock/path",
					Name: "2026-02-02T12.05.20.700 beach vacation",
					Ext:  ".jpg.xmp",
				},
				Orig: Meta{
					VTime: VTime{Time: vTime},
					Desc:  Desc{Text: bDesc},
				},
				Matched: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFS := MockFS{
				MockOpen: func(name string) (io.ReadCloser, error) {
					for _, fs := range tt.filesSetup {
						expectedName := fs.name + fs.ext
						if filepath.Base(name) == expectedName {
							return io.NopCloser(bytes.NewBufferString(fs.XMPData)), nil
						}
					}
					return nil, os.ErrNotExist
				},
			}

			assetMap := make(map[string]Voit)
			v := &VoitImpl{
				Sidecar: tt.sidecar,
				Orig: Meta{
					VTime: VTime{Time: vTime},
					Desc:  Desc{Text: bDesc},
				},
			}
			assetMap["asset_key"] = v

			mockAssets := &AssetImpl{
				m: assetMap,
			}

			var buf bytes.Buffer
			cfg := Config{FS: mockFS}

			StageXMP(&buf, mockAssets, tt.opts, cfg)

			if v.Matched != tt.wantVoit.Matched {
				t.Errorf("\nGot:  %v\nWant: %v\n", tt.wantVoit.Matched, v.Matched)
			}
		})
	}
}
