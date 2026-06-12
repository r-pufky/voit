package voit

import (
	"reflect"
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
				Format:  DefaultVFormat,
				Pattern: "voit",
				SSep:    DefaultSpanSep,
				DSep:    DefaultDescSep,
				TSep:    DefaultTagsSep,
				Lower:   false,
			},
		},
		{
			name: "sanity: non-default values",
			opts: &Opts{
				Format:  "15:04",
				SpanSep: "->",
				DescSep: "|",
				TagSep:  "#",
				Lower:   true,
				Rename: RenameOpts{
					Pattern: "photo-ms",
				},
			},
			want: Config{
				Format:  "15:04",
				Pattern: "photo-ms",
				SSep:    "->",
				DSep:    "|",
				TSep:    "#",
				Lower:   true,
			},
		},
		{
			name: "sanity: partial set [default values used elsewhere]",
			opts: &Opts{
				Format: "2006",
			},
			want: Config{
				Format:  "2006",
				Pattern: "voit",
				SSep:    DefaultSpanSep,
				DSep:    DefaultDescSep,
				TSep:    DefaultTagsSep,
				Lower:   false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.opts.Voit()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Opts.Voit() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Simple logic requires no testing.
func TestValidateNOOP(t *testing.T) {}
