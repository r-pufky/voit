package voit

import (
	"reflect"
	"testing"
	"time"
)

func TestVoitIngest(t *testing.T) {
	fixedCTime := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	fixedMTime := time.Date(2026, time.February, 1, 14, 0, 0, 0, time.UTC)
	fixedSTime := time.Date(2026, time.March, 1, 20, 0, 0, 0, time.UTC)
	parsedTime := time.Date(2026, time.May, 17, 10, 45, 36, 300000000, time.UTC)
	voitTime := time.Date(2026, time.February, 2, 12, 5, 20, 700000000, time.UTC)
	baseTags := []string{"summer", "vacation", "beach"}
	baseDesc := "beach vacation"

	tests := []struct {
		name     string
		f        *Voit
		c        Config
		wantVoit *Voit
	}{
		// Sanity checks.
		{
			name: "sanity: sanitized format",
			f: &Voit{
				File: File{
					Source: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
					Name:   "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
					Ext:    ".jpg",
				},
			},
			c: Config{
				Pattern: "photo",
			},
			wantVoit: &Voit{
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
			},
		},
		{
			name: "sanity: alternative separators",
			f: &Voit{
				File: File{
					Source: "/tmp/2026-02-02T12.05.20.700|beach vacation - summer vacation beach.jpg",
					Name:   "2026-02-02T12.05.20.700|beach vacation - summer vacation beach",
					Ext:    ".jpg",
				},
			},
			c: Config{
				Pattern: "photo",
				DSep:    "|",
				TSep:    " - ",
			},
			wantVoit: &Voit{
				File: File{
					Source: "/tmp/2026-02-02T12.05.20.700|beach vacation - summer vacation beach.jpg",
					Name:   "2026-02-02T12.05.20.700|beach vacation - summer vacation beach",
					Ext:    ".jpg",
				},
				Orig: Meta{
					VTime: VTime{Time: voitTime},
					Tags:  Tag{Items: baseTags},
					Desc:  Desc{Text: baseDesc},
				},
			},
		},
		{
			name: "sanity: duplicate separators",
			f: &Voit{
				File: File{
					Source: "/tmp/2026-02-02T12.05.20.700|beach vacation|summer vacation beach.jpg",
					Name:   "2026-02-02T12.05.20.700|beach vacation|summer vacation beach",
					Ext:    ".jpg",
				},
			},
			c: Config{
				Pattern: "photo",
				DSep:    "|",
				TSep:    "|",
			},
			wantVoit: &Voit{
				File: File{
					Source: "/tmp/2026-02-02T12.05.20.700|beach vacation|summer vacation beach.jpg",
					Name:   "2026-02-02T12.05.20.700|beach vacation|summer vacation beach",
					Ext:    ".jpg",
				},
				Orig: Meta{
					VTime: VTime{Time: voitTime},
					Tags:  Tag{Items: baseTags},
					Desc:  Desc{Text: baseDesc},
				},
			},
		},
		// Missing fields.
		{
			name: "fields: missing desc",
			f: &Voit{
				File: File{
					Source: "/tmp/2026-02-02T12.05.20.700  -- summer vacation beach.jpg",
					Name:   "2026-02-02T12.05.20.700  -- summer vacation beach",
					Ext:    ".jpg",
				},
			},
			c: Config{
				Pattern: "photo",
			},
			wantVoit: &Voit{
				File: File{
					Source: "/tmp/2026-02-02T12.05.20.700  -- summer vacation beach.jpg",
					Name:   "2026-02-02T12.05.20.700  -- summer vacation beach",
					Ext:    ".jpg",
				},
				Orig: Meta{
					VTime: VTime{Time: voitTime},
					Tags:  Tag{Items: baseTags},
				},
			},
		},
		{
			name: "fields: missing tags",
			f: &Voit{
				File: File{
					Source: "/tmp/2026-02-02T12.05.20.700 beach vacation.jpg",
					Name:   "2026-02-02T12.05.20.700 beach vacation",
					Ext:    ".jpg",
				},
			},
			c: Config{
				Pattern: "photo",
			},
			wantVoit: &Voit{
				File: File{
					Source: "/tmp/2026-02-02T12.05.20.700 beach vacation.jpg",
					Name:   "2026-02-02T12.05.20.700 beach vacation",
					Ext:    ".jpg",
				},
				Orig: Meta{
					VTime: VTime{Time: voitTime},
					Desc:  Desc{Text: baseDesc},
				},
			},
		},
		{
			name: "fields: missing desc tags",
			f: &Voit{
				File: File{
					Source: "/tmp/2026-02-02T12.05.20.700.jpg",
					Name:   "2026-02-02T12.05.20.700",
					Ext:    ".jpg",
				},
			},
			c: Config{
				Pattern: "photo",
			},
			wantVoit: &Voit{
				File: File{
					Source: "/tmp/2026-02-02T12.05.20.700.jpg",
					Name:   "2026-02-02T12.05.20.700",
					Ext:    ".jpg",
				},
				Orig: Meta{
					VTime: VTime{Time: voitTime},
					Desc:  Desc{Text: "2026-02-02T12.05.20.700"},
				},
			},
		},
		// Patterns.
		{
			name: "patterns: ctime",
			f: &Voit{
				File: File{
					Source: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
					Name:   "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
					Ext:    ".jpg",
					CTime:  fixedCTime,
				},
			},
			c: NewConfig(),
			wantVoit: &Voit{
				File: File{
					Source: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
					Name:   "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
					Ext:    ".jpg",
					CTime:  fixedCTime,
				},
				Orig: Meta{
					VTime: VTime{Time: voitTime},
					PTime: VTime{Time: fixedCTime},
					Tags:  Tag{Items: baseTags},
					Desc:  Desc{Text: baseDesc},
				},
			},
		},
		{
			name: "patterns: mtime",
			f: &Voit{
				File: File{
					Source: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
					Name:   "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
					Ext:    ".jpg",
					MTime:  fixedMTime,
				},
			},
			c: Config{
				Pattern: "modified",
			},
			wantVoit: &Voit{
				File: File{
					Source: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
					Name:   "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
					Ext:    ".jpg",
					MTime:  fixedMTime,
				},
				Orig: Meta{
					VTime: VTime{Time: voitTime},
					PTime: VTime{Time: fixedMTime},
					Tags:  Tag{Items: baseTags},
					Desc:  Desc{Text: baseDesc},
				},
			},
		},
		{
			name: "patterns: set",
			f: &Voit{
				File: File{
					Source: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
					Name:   "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
					Ext:    ".jpg",
				},
			},
			c: Config{
				Pattern: "set",
				Set:     "2026-03-01T20.00.00.000",
			},
			wantVoit: &Voit{
				File: File{
					Source: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
					Name:   "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
					Ext:    ".jpg",
				},
				Orig: Meta{
					VTime: VTime{Time: voitTime},
					PTime: VTime{Time: fixedSTime},
					Tags:  Tag{Items: baseTags},
					Desc:  Desc{Text: baseDesc},
				},
			},
		},
		{ // Regex patterns tested in regex_test.go.
			name: "patterns: regex",
			f: &Voit{
				File: File{
					Source: "/tmp/20260517_104536300.jpg",
					Name:   "20260517_104536300",
					Ext:    ".jpg",
				},
			},
			c: Config{
				Pattern: "photo-ms",
			},
			wantVoit: &Voit{
				File: File{
					Source: "/tmp/20260517_104536300.jpg",
					Name:   "20260517_104536300",
					Ext:    ".jpg",
				},
				Orig: Meta{
					PTime: VTime{Time: parsedTime},
					Desc:  Desc{Text: "20260517_104536300"},
				},
			},
		},
		{
			name: "patterns: no match",
			f: &Voit{
				File: File{
					Source: "/tmp/no_match.jpg",
					Name:   "no_match",
					Ext:    ".jpg",
				},
			},
			c: Config{
				Pattern: "photo-ms",
			},
			wantVoit: &Voit{
				File: File{
					Source: "/tmp/no_match.jpg",
					Name:   "no_match",
					Ext:    ".jpg",
				},
				Orig: Meta{
					Desc: Desc{Text: "no_match"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.f.Ingest(&tt.c)

			if !reflect.DeepEqual(tt.f, tt.wantVoit) {
				t.Errorf("\nGot Voit:  %+v\nWant Voit: %+v", tt.f, tt.wantVoit)
			}
		})
	}
}
