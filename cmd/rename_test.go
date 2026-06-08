package cmd

import (
	"reflect"
	"testing"
	"time"

	. "github.com/r-pufky/voit/config"
	"github.com/r-pufky/voit/voit"
)

var (
	fixedCTime = time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	fixedMTime = time.Date(2026, time.February, 1, 14, 0, 0, 0, time.UTC)
	fixedSTime = time.Date(2026, time.March, 1, 20, 0, 0, 0, time.UTC)
	parsedTime = time.Date(2026, time.May, 17, 10, 45, 36, 300000000, time.UTC)
	voitTime   = time.Date(2026, time.February, 2, 12, 5, 20, 700000000, time.UTC)

	baseTags = []string{"summer", "vacation", "beach"}
	baseDesc = "beach vacation"
)

func TestStageRename(t *testing.T) {
	tests := []struct {
		name     string
		opts     *Opts
		f        []*voit.Voit
		wantVoit []voit.Voit
	}{
		{
			name: "sanity: no match [ZeroTime, matched false]",
			opts: &Opts{
				Rename: RenameOpts{},
			},
			f: []*voit.Voit{
				{
					File: voit.File{
						Source: "/tmp/invalid_format.txt",
						Name:   "invalid_format",
						Ext:    ".txt",
					},
				},
			},
			wantVoit: []voit.Voit{
				{
					File: voit.File{
						Source: "/tmp/invalid_format.txt",
						Name:   "invalid_format",
						Ext:    ".txt",
					},
					Matched: false,
					Target:  "/tmp/0001-01-01T00.00.00.000.txt",
				},
			},
		},
		{
			name: "sanity: default [vtime match]",
			opts: &Opts{
				Rename: RenameOpts{},
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
			wantVoit: []voit.Voit{
				{
					File: voit.File{
						Source: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
						Name:   "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
						Ext:    ".jpg",
					},
					Orig: voit.Meta{
						VTime: voit.VTime{Time: voitTime},
						Tags:  voit.Tag{Items: baseTags},
						Desc:  voit.Desc{Text: baseDesc},
					},
					Mark: voit.Meta{
						VTime: voit.VTime{Time: voitTime},
						Tags:  voit.Tag{Items: baseTags},
						Desc:  voit.Desc{Text: baseDesc},
					},
					Matched: true,
					Target:  "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
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
			f: []*voit.Voit{
				{
					File: voit.File{
						Source: "/tmp/2026-02-02T12.05.20.700 20260517_104536300 beach vacation -- summer vacation beach.jpg",
						Name:   "2026-02-02T12.05.20.700 20260517_104536300 beach vacation -- summer vacation beach",
						Ext:    ".jpg",
					},
				},
			},
			wantVoit: []voit.Voit{
				{
					File: voit.File{
						Source: "/tmp/2026-02-02T12.05.20.700 20260517_104536300 beach vacation -- summer vacation beach.jpg",
						Name:   "2026-02-02T12.05.20.700 20260517_104536300 beach vacation -- summer vacation beach",
						Ext:    ".jpg",
					},
					Orig: voit.Meta{
						VTime: voit.VTime{Time: voitTime},
						PTime: voit.VTime{Time: parsedTime},
						Tags:  voit.Tag{Items: baseTags},
						Desc:  voit.Desc{Text: "20260517_104536300 beach vacation"},
					},
					Mark: voit.Meta{
						VTime: voit.VTime{Time: parsedTime},
						Tags:  voit.Tag{Items: baseTags},
						Desc:  voit.Desc{Text: baseDesc},
					},
					Matched: true,
					Target:  "/tmp/2026-05-17T10.45.36.300 beach vacation -- summer vacation beach.jpg",
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
			f: []*voit.Voit{
				{
					File: voit.File{
						Source: "/tmp/2026-02-02T12.05.20.700 20260517_104536300 beach vacation -- summer vacation beach.jpg",
						Name:   "2026-02-02T12.05.20.700 20260517_104536300 beach vacation -- summer vacation beach",
						Ext:    ".jpg",
					},
				},
			},
			wantVoit: []voit.Voit{
				{
					File: voit.File{
						Source: "/tmp/2026-02-02T12.05.20.700 20260517_104536300 beach vacation -- summer vacation beach.jpg",
						Name:   "2026-02-02T12.05.20.700 20260517_104536300 beach vacation -- summer vacation beach",
						Ext:    ".jpg",
					},
					Orig: voit.Meta{
						VTime: voit.VTime{Time: voitTime},
						PTime: voit.VTime{Time: fixedSTime},
						Tags:  voit.Tag{Items: baseTags},
						Desc:  voit.Desc{Text: "20260517_104536300 beach vacation"},
					},
					Mark: voit.Meta{
						VTime: voit.VTime{Time: fixedSTime},
						Tags:  voit.Tag{Items: baseTags},
						Desc:  voit.Desc{Text: "20260517_104536300 beach vacation"},
					},
					Matched: true,
					Target:  "/tmp/2026-03-01T20.00.00.000 20260517_104536300 beach vacation -- summer vacation beach.jpg",
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
			f: []*voit.Voit{
				{
					File: voit.File{
						Source: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
						Name:   "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
						Ext:    ".jpg",
					},
				},
			},
			wantVoit: []voit.Voit{
				{
					File: voit.File{
						Source: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
						Name:   "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
						Ext:    ".jpg",
					},
					Orig: voit.Meta{
						VTime: voit.VTime{Time: voitTime},
						Tags:  voit.Tag{Items: baseTags},
						Desc:  voit.Desc{Text: baseDesc},
					},
					Mark: voit.Meta{
						VTime: voit.VTime{Time: voitTime},
						Tags:  voit.Tag{Items: []string{}},
						Desc:  voit.Desc{Text: ""},
					},
					Matched: true,
					Target:  "/tmp/2026-02-02T12.05.20.700.jpg",
				},
			},
		},
		{
			name: "collision: default [count added to desc].",
			opts: &Opts{
				Rename: RenameOpts{},
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
						Source: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
						Name:   "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
						Ext:    ".jpg",
					},
				},
				{
					File: voit.File{
						Source: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
						Name:   "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
						Ext:    ".jpg",
					},
				},
			},
			wantVoit: []voit.Voit{
				{
					File: voit.File{
						Source: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
						Name:   "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
						Ext:    ".jpg",
					},
					Orig: voit.Meta{
						VTime: voit.VTime{Time: voitTime},
						Tags:  voit.Tag{Items: baseTags},
						Desc:  voit.Desc{Text: baseDesc},
					},
					Mark: voit.Meta{
						VTime: voit.VTime{Time: voitTime},
						Tags:  voit.Tag{Items: baseTags},
						Desc:  voit.Desc{Text: baseDesc},
					},
					Matched: true,
					Target:  "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
				},
				{
					File: voit.File{
						Source: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
						Name:   "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
						Ext:    ".jpg",
					},
					Orig: voit.Meta{
						VTime: voit.VTime{Time: voitTime},
						Tags:  voit.Tag{Items: baseTags},
						Desc:  voit.Desc{Text: baseDesc},
					},
					Mark: voit.Meta{
						VTime: voit.VTime{Time: voitTime},
						Tags:  voit.Tag{Items: baseTags},
						Desc:  voit.Desc{Text: "beach vacation_1"},
					},
					Matched: true,
					Target:  "/tmp/2026-02-02T12.05.20.700 beach vacation_1 -- summer vacation beach.jpg",
				},
				{
					File: voit.File{
						Source: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
						Name:   "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
						Ext:    ".jpg",
					},
					Orig: voit.Meta{
						VTime: voit.VTime{Time: voitTime},
						Tags:  voit.Tag{Items: baseTags},
						Desc:  voit.Desc{Text: baseDesc},
					},
					Mark: voit.Meta{
						VTime: voit.VTime{Time: voitTime},
						Tags:  voit.Tag{Items: baseTags},
						Desc:  voit.Desc{Text: "beach vacation_2"},
					},
					Matched: true,
					Target:  "/tmp/2026-02-02T12.05.20.700 beach vacation_2 -- summer vacation beach.jpg",
				},
			},
		},
		{
			name: "collision: no desc, no tags [count added to desc].",
			opts: &Opts{
				Rename: RenameOpts{
					NoDesc: true,
					NoTags: true,
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
						Source: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
						Name:   "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
						Ext:    ".jpg",
					},
				},
				{
					File: voit.File{
						Source: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
						Name:   "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
						Ext:    ".jpg",
					},
				},
			},
			wantVoit: []voit.Voit{
				{
					File: voit.File{
						Source: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
						Name:   "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
						Ext:    ".jpg",
					},
					Orig: voit.Meta{
						VTime: voit.VTime{Time: voitTime},
						Tags:  voit.Tag{Items: baseTags},
						Desc:  voit.Desc{Text: baseDesc},
					},
					Mark: voit.Meta{
						VTime: voit.VTime{Time: voitTime},
						Tags:  voit.Tag{Items: []string{}},
						Desc:  voit.Desc{Text: ""},
					},
					Matched: true,
					Target:  "/tmp/2026-02-02T12.05.20.700.jpg",
				},
				{
					File: voit.File{
						Source: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
						Name:   "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
						Ext:    ".jpg",
					},
					Orig: voit.Meta{
						VTime: voit.VTime{Time: voitTime},
						Tags:  voit.Tag{Items: baseTags},
						Desc:  voit.Desc{Text: baseDesc},
					},
					Mark: voit.Meta{
						VTime: voit.VTime{Time: voitTime},
						Tags:  voit.Tag{Items: []string{}},
						Desc:  voit.Desc{Text: "1"},
					},
					Matched: true,
					Target:  "/tmp/2026-02-02T12.05.20.700 1.jpg",
				},
				{
					File: voit.File{
						Source: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
						Name:   "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
						Ext:    ".jpg",
					},
					Orig: voit.Meta{
						VTime: voit.VTime{Time: voitTime},
						Tags:  voit.Tag{Items: baseTags},
						Desc:  voit.Desc{Text: baseDesc},
					},
					Mark: voit.Meta{
						VTime: voit.VTime{Time: voitTime},
						Tags:  voit.Tag{Items: []string{}},
						Desc:  voit.Desc{Text: "2"},
					},
					Matched: true,
					Target:  "/tmp/2026-02-02T12.05.20.700 2.jpg",
				},
			},
		},
		{
			name: "collision: no desc, tags [count added to desc].",
			opts: &Opts{
				Rename: RenameOpts{
					NoDesc: true,
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
						Source: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
						Name:   "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
						Ext:    ".jpg",
					},
				},
				{
					File: voit.File{
						Source: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
						Name:   "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
						Ext:    ".jpg",
					},
				},
			},
			wantVoit: []voit.Voit{
				{
					File: voit.File{
						Source: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
						Name:   "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
						Ext:    ".jpg",
					},
					Orig: voit.Meta{
						VTime: voit.VTime{Time: voitTime},
						Tags:  voit.Tag{Items: baseTags},
						Desc:  voit.Desc{Text: baseDesc},
					},
					Mark: voit.Meta{
						VTime: voit.VTime{Time: voitTime},
						Tags:  voit.Tag{Items: baseTags},
						Desc:  voit.Desc{Text: ""},
					},
					Matched: true,
					Target:  "/tmp/2026-02-02T12.05.20.700 -- summer vacation beach.jpg",
				},
				{
					File: voit.File{
						Source: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
						Name:   "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
						Ext:    ".jpg",
					},
					Orig: voit.Meta{
						VTime: voit.VTime{Time: voitTime},
						Tags:  voit.Tag{Items: baseTags},
						Desc:  voit.Desc{Text: baseDesc},
					},
					Mark: voit.Meta{
						VTime: voit.VTime{Time: voitTime},
						Tags:  voit.Tag{Items: baseTags},
						Desc:  voit.Desc{Text: "1"},
					},
					Matched: true,
					Target:  "/tmp/2026-02-02T12.05.20.700 1 -- summer vacation beach.jpg",
				},
				{
					File: voit.File{
						Source: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
						Name:   "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
						Ext:    ".jpg",
					},
					Orig: voit.Meta{
						VTime: voit.VTime{Time: voitTime},
						Tags:  voit.Tag{Items: baseTags},
						Desc:  voit.Desc{Text: baseDesc},
					},
					Mark: voit.Meta{
						VTime: voit.VTime{Time: voitTime},
						Tags:  voit.Tag{Items: baseTags},
						Desc:  voit.Desc{Text: "2"},
					},
					Matched: true,
					Target:  "/tmp/2026-02-02T12.05.20.700 2 -- summer vacation beach.jpg",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stageRename(tt.f, tt.opts)

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
