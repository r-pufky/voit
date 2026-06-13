package voit

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestVoit(t *testing.T) {
	tests := []struct {
		name string
		opts *Opts
		want Config
	}{
		{
			name: "sanity: empty defaults [pattern: voit]",
			opts: &Opts{},
			want: Config{
				Format:         DefaultVFormat,
				Pattern:        "voit",
				SSep:           DefaultSpanSep,
				DSep:           DefaultDescSep,
				TSep:           DefaultTagsSep,
				SyncInFolder:   DefaultSyncInFolder,
				SyncOutFolder:  DefaultSyncOutFolder,
				SyncOutSpace:   DefaultSyncOutSpace,
				SyncKeepFolder: false,
				SyncKeepSpace:  false,
				Lower:          false,
			},
		},
		{
			name: "sanity: non-default values",
			opts: &Opts{
				TagSep: "#",
				Tag: TagOpts{
					SyncInFolder:   "_",
					SyncOutFolder:  ">",
					SyncOutSpace:   "_",
					SyncKeepFolder: true,
					SyncKeepSpace:  true,
				},
				DescSep: "|",
				SpanSep: "->",
				Format:  "15:04",
				Rename: RenameOpts{
					Pattern: "photo-ms",
				},
				Lower: true,
			},
			want: Config{
				Format:         "15:04",
				Pattern:        "photo-ms",
				SSep:           "->",
				DSep:           "|",
				TSep:           "#",
				SyncInFolder:   "_",
				SyncOutFolder:  ">",
				SyncOutSpace:   "_",
				SyncKeepFolder: true,
				SyncKeepSpace:  true,
				Lower:          true,
			},
		},
		{
			name: "sanity: partial set [default values used elsewhere]",
			opts: &Opts{
				Format: "2006",
			},
			want: Config{
				Format:         "2006",
				Pattern:        "voit",
				SSep:           DefaultSpanSep,
				DSep:           DefaultDescSep,
				TSep:           DefaultTagsSep,
				SyncInFolder:   DefaultSyncInFolder,
				SyncOutFolder:  DefaultSyncOutFolder,
				SyncOutSpace:   DefaultSyncOutSpace,
				SyncKeepFolder: false,
				SyncKeepSpace:  false,
				Lower:          false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.opts.Voit()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("\nGot:  %v\nWant: %v\n", got, tt.want)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	// Cache working directory for AbsSource testing.
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
			name: "santiy: relative abs source [abs source resolved]",
			input: Opts{
				SpanSep:   "--",
				DescSep:   " ",
				TagSep:    " -- ",
				AbsSource: ".",
				Tag: TagOpts{
					SyncOutFolder: "➔",
					SyncOutSpace:  "⠀",
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
					SyncOutFolder: "➔",
					SyncOutSpace:  "⠀",
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
			name: "sanity: SyncOutFolder collision [error raised]",
			input: Opts{
				SpanSep:   "-",
				DescSep:   " ",
				TagSep:    " -- ",
				AbsSource: "",
				Tag: TagOpts{
					SyncOutFolder: "-",
					SyncOutSpace:  "⠀",
				},
			},
			wantErr:     true,
			errContains: "tag-folder must be unique non-whitespace character",
		},
		{
			name: "sanity: SyncOutFolder whitespace [error raised]",
			input: Opts{
				SpanSep:   "-",
				DescSep:   " ",
				TagSep:    " -- ",
				AbsSource: "",
				Tag: TagOpts{
					SyncOutFolder: " ",
					SyncOutSpace:  "⠀",
				},
			},
			wantErr:     true,
			errContains: "tag-folder must be unique non-whitespace character",
		},
		{
			name: "sanity: SyncOutSpace collision [error raised]",
			input: Opts{
				SpanSep:   "-",
				DescSep:   " ",
				TagSep:    " -- ",
				AbsSource: "",
				Tag: TagOpts{
					SyncOutFolder: "➔",
					SyncOutSpace:  "-",
				},
			},
			wantErr:     true,
			errContains: "tag-space must be unique non-whitespace character",
		},
		{
			name: "sanity: SyncOutSpace whitespace [error raised]",
			input: Opts{
				SpanSep:   "--",
				DescSep:   " ",
				TagSep:    " -- ",
				AbsSource: "",
				Tag: TagOpts{
					SyncOutFolder: "➔",
					SyncOutSpace:  " ",
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

			err := opts.Validate()

			if (err != nil) != tt.wantErr {
				t.Fatalf("\nGot:  %v\nWant: %v\n", err, tt.wantErr)
			}

			if tt.wantErr && err != nil {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Expected error to contain %q, but got %q", tt.errContains, err.Error())
				}
			}

			if !tt.wantErr && tt.checkSource != nil {
				tt.checkSource(t, opts.AbsSource)
			}
		})
	}
}
