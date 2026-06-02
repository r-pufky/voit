package cmd

import (
	"reflect"
	"testing"
	"time"

	. "github.com/r-pufky/voit/config"
	"github.com/r-pufky/voit/voit"
)

var (
	vTime = time.Date(2026, time.February, 2, 12, 5, 20, 700000000, time.UTC)

	bTags = []string{"summer", "vacation", "beach"}
	bDesc = "beach vacation"
)

func TestNormalize(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "All uppercase strings",
			input:    []string{"APPLE", "BANANA", "CHERRY"},
			expected: []string{"apple", "banana", "cherry"},
		},
		{
			name:     "Mixed case strings",
			input:    []string{"GoLang", "NeW YoRk", "123!@#"},
			expected: []string{"golang", "new york", "123!@#"},
		},
		{
			name:     "Already lowercase",
			input:    []string{"cat", "dog"},
			expected: []string{"cat", "dog"},
		},
		{
			name:     "Empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "Nil slice",
			input:    nil,
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var orig []string
			if tt.input != nil {
				orig = make([]string, len(tt.input))
				copy(orig, tt.input)
			}

			actual := normalize(tt.input)

			if !reflect.DeepEqual(actual, tt.expected) {
				t.Errorf("\nGot:  %v\nWant: %v", actual, tt.expected)
			}

			if tt.input != nil && !reflect.DeepEqual(tt.input, orig) {
				t.Errorf("normalize() mutated the input slice\nGot:  %v\nWant: %v", tt.input, orig)
			}

			if len(tt.input) > 0 && len(actual) > 0 && &actual[0] == &tt.input[0] {
				t.Errorf("normalize() returned same pointer instead of a new allocation")
			}
		})
	}
}

func TestStageTag(t *testing.T) {
	tests := []struct {
		name     string
		opts     *Opts
		f        []*voit.Voit
		c        *voit.Config
		wantVoit []voit.Voit
	}{
		{
			name: "sanity: add tag [tag added]",
			opts: &Opts{
				Tag: TagOpts{
					Add: []string{"additional"},
				},
			},
			f: []*voit.Voit{
				{
					File: voit.File{
						Source: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
						Name:   "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
						Ext:    ".jpg",
					},
				},
			},
			c: &voit.Config{},
			wantVoit: []voit.Voit{
				{
					File: voit.File{
						Source: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
						Name:   "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
						Ext:    ".jpg",
					},
					Orig: voit.Meta{
						VTime: voit.VTime{Time: vTime},
						Tags:  voit.Tag{Items: []string{"summer", "vacation", "beach"}},
						Desc:  voit.Desc{Text: bDesc},
					},
					Mark: voit.Meta{
						VTime: voit.VTime{Time: vTime},
						Tags:  voit.Tag{Items: []string{"summer", "vacation", "beach", "additional"}},
						Desc:  voit.Desc{Text: bDesc},
					},
					Matched: true,
					Target:  "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach additional.jpg",
				},
			},
		},
		{
			name: "sanity: remove tag [tag removed]",
			opts: &Opts{
				Tag: TagOpts{
					Remove: []string{"summer"},
				},
			},
			f: []*voit.Voit{
				{
					File: voit.File{
						Source: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
						Name:   "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
						Ext:    ".jpg",
					},
				},
			},
			c: &voit.Config{},
			wantVoit: []voit.Voit{
				{
					File: voit.File{
						Source: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
						Name:   "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
						Ext:    ".jpg",
					},
					Orig: voit.Meta{
						VTime: voit.VTime{Time: vTime},
						Tags:  voit.Tag{Items: []string{"summer", "vacation", "beach"}},
						Desc:  voit.Desc{Text: bDesc},
					},
					Mark: voit.Meta{
						VTime: voit.VTime{Time: vTime},
						Tags:  voit.Tag{Items: []string{"vacation", "beach"}},
						Desc:  voit.Desc{Text: bDesc},
					},
					Matched: true,
					Target:  "/tmp/2026-02-02T12.05.20.700 beach vacation -- vacation beach.jpg",
				},
			},
		},
		{
			name: "sanity: set tags [tag overwritten]",
			opts: &Opts{
				Tag: TagOpts{
					Set: []string{"family", "europe"},
				},
			},
			f: []*voit.Voit{
				{
					File: voit.File{
						Source: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
						Name:   "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
						Ext:    ".jpg",
					},
				},
			},
			c: &voit.Config{},
			wantVoit: []voit.Voit{
				{
					File: voit.File{
						Source: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
						Name:   "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
						Ext:    ".jpg",
					},
					Orig: voit.Meta{
						VTime: voit.VTime{Time: vTime},
						Tags:  voit.Tag{Items: []string{"summer", "vacation", "beach"}},
						Desc:  voit.Desc{Text: bDesc},
					},
					Mark: voit.Meta{
						VTime: voit.VTime{Time: vTime},
						Tags:  voit.Tag{Items: []string{"family", "europe"}},
						Desc:  voit.Desc{Text: bDesc},
					},
					Matched: true,
					Target:  "/tmp/2026-02-02T12.05.20.700 beach vacation -- family europe.jpg",
				},
			},
		},
		{
			name: "sanity: select add tag [subset of tags have tags added]",
			opts: &Opts{
				Tag: TagOpts{
					Select: []string{"summer", "vacation", "park"},
					Add:    []string{"europe"},
				},
			},
			f: []*voit.Voit{
				{
					File: voit.File{
						Source: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
						Name:   "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
						Ext:    ".jpg",
					},
				},
				{
					File: voit.File{
						Source: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation park.jpg",
						Name:   "2026-02-02T12.05.20.700 beach vacation -- summer vacation park",
						Ext:    ".jpg",
					},
				},
			},
			c: &voit.Config{},
			wantVoit: []voit.Voit{
				{
					File: voit.File{
						Source: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
						Name:   "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
						Ext:    ".jpg",
					},
					Orig: voit.Meta{
						VTime: voit.VTime{Time: vTime},
						Tags:  voit.Tag{Items: []string{"summer", "vacation", "beach"}},
						Desc:  voit.Desc{Text: bDesc},
					},
					Mark: voit.Meta{
						VTime: voit.VTime{Time: vTime},
						Tags:  voit.Tag{Items: []string{"summer", "vacation", "beach"}},
						Desc:  voit.Desc{Text: bDesc},
					},
					Matched: false,
					Target:  "",
				},
				{
					File: voit.File{
						Source: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation park.jpg",
						Name:   "2026-02-02T12.05.20.700 beach vacation -- summer vacation park",
						Ext:    ".jpg",
					},
					Orig: voit.Meta{
						VTime: voit.VTime{Time: vTime},
						Tags:  voit.Tag{Items: []string{"summer", "vacation", "park"}},
						Desc:  voit.Desc{Text: bDesc},
					},
					Mark: voit.Meta{
						VTime: voit.VTime{Time: vTime},
						Tags:  voit.Tag{Items: []string{"summer", "vacation", "park", "europe"}},
						Desc:  voit.Desc{Text: bDesc},
					},
					Matched: true,
					Target:  "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation park europe.jpg",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.c.TSep == "" {
				tt.c.TSep = Cfg.TagSep
			}
			if tt.c.DSep == "" {
				tt.c.DSep = Cfg.DescSep
			}
			if tt.c.SSep == "" {
				tt.c.SSep = Cfg.SpanSep
			}
			if tt.c.Format == "" {
				tt.c.Format = voit.DefaultVFormat
			}
			stageTag(tt.f, tt.opts, tt.c)

			// DeepEqual will compare memory addresses if pointers, not values.
			// Convert to value. This is required as pointers are needed to update
			// the struct in place during stageRename.
			got := make([]voit.Voit, len(tt.f))
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
