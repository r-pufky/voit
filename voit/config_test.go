package voit

import (
	"reflect"
	"testing"
)

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name string
		args Config
		want Config
	}{
		{
			name: "new config without existing options",
			args: Config{}, // Simulate no Config passed.
			want: Config{
				Voit: VoitConfig{
					DescSep:   DescSep,
					TagSep:    TagSep,
					SpanSep:   SpanSep,
					VFormat:   VFormat,
					Pattern:   Pattern,
					Set:       "",
					Verbose:   false,
					Overwrite: false,
					Lower:     false,
					Yes:       false,
				},
				Sync: SyncConfig{
					MetaFolder: SyncMetaFolder,
					Folder:     SyncFolder,
					Space:      SyncSpace,
					KeepFolder: false,
					KeepSpace:  false,
				},
				FS:   RealFS{},
				Unix: RealUnix{},
			},
		},
		{
			name: "new config with custom options",
			args: Config{
				Voit: VoitConfig{
					DescSep:   "custom_desc",
					TagSep:    "custom_tag",
					SpanSep:   "custom_span",
					VFormat:   "2006-01-02",
					Pattern:   "custom_pattern",
					Set:       "custom_set",
					Verbose:   true,
					Overwrite: true,
					Lower:     true,
					Yes:       true,
				},
				Sync: SyncConfig{
					MetaFolder: "custom_meta",
					Folder:     "custom_folder",
					Space:      "custom_space",
					KeepFolder: true,
					KeepSpace:  true,
				},
				FS:   MockFS{},   // Use a mock FS.
				Unix: MockUnix{}, // Use a mock Unix.
			},
			want: Config{
				Voit: VoitConfig{
					DescSep:   "custom_desc",
					TagSep:    "custom_tag",
					SpanSep:   "custom_span",
					VFormat:   "2006-01-02",
					Pattern:   "custom_pattern",
					Set:       "custom_set",
					Verbose:   true,
					Overwrite: true,
					Lower:     true,
					Yes:       true,
				},
				Sync: SyncConfig{
					MetaFolder: "custom_meta",
					Folder:     "custom_folder",
					Space:      "custom_space",
					KeepFolder: true,
					KeepSpace:  true,
				},
				FS:   MockFS{},
				Unix: MockUnix{},
			},
		},
		{
			name: "new minimal config with undefined FS, Unix [defaults, override_only, RealFS, RealUnix]",
			args: Config{
				Voit: VoitConfig{
					DescSep: "override_only",
				},
			},
			want: Config{
				Voit: VoitConfig{
					DescSep:   "override_only",
					TagSep:    TagSep,
					SpanSep:   SpanSep,
					VFormat:   VFormat,
					Pattern:   Pattern,
					Set:       "",
					Verbose:   false,
					Overwrite: false,
					Lower:     false,
					Yes:       false,
				},
				Sync: SyncConfig{
					MetaFolder: SyncMetaFolder,
					Folder:     SyncFolder,
					Space:      SyncSpace,
					KeepFolder: false,
					KeepSpace:  false,
				},
				FS:   RealFS{},
				Unix: RealUnix{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewConfig(tt.args)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("\nGot:  %+v\nWant: %+v\n", got, tt.want)
			}
		})
	}
}

func TestConfigWithPattern(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		c           Config
		pattern     string
		wantPattern string
	}{
		{
			name:        "set pattern on empty config",
			c:           Config{},
			pattern:     "voit",
			wantPattern: "voit",
		},
		{
			name: "set pattern preserve other fields",
			c: Config{
				Voit: VoitConfig{
					Pattern: "old-pattern",
					VFormat: "2006-01-02",
				},
			},
			pattern:     "new-pattern",
			wantPattern: "new-pattern",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			orig := tt.c.Voit.VFormat

			got := tt.c.WithPattern(tt.pattern)

			if got.Voit.Pattern != tt.wantPattern {
				t.Errorf("\nPattern\nGot:  %q\nWant: %q\n", got.Voit.Pattern, tt.wantPattern)
			}

			if got.Voit.VFormat != orig {
				t.Errorf("\nVFormat\nGot:  %q\nWant: %q\n", got.Voit.VFormat, orig)
			}

			if tt.c.Voit.Pattern == tt.pattern && tt.pattern != "" {
				t.Error("\nOriginal Config struct mutated.\n")
			}
		})
	}
}
