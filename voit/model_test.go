package voit

import (
	"reflect"
	"testing"
	"time"
)

var (
	fixedCTime = time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	fixedMTime = time.Date(2026, time.February, 1, 14, 0, 0, 0, time.UTC)
	parsedTime = time.Date(2026, time.May, 17, 10, 45, 36, 300000000, time.UTC)
	voitTime   = time.Date(2026, time.February, 2, 12, 5, 20, 700000000, time.UTC)

	baseTags = []string{"summer", "vacation", "beach"}
	baseDesc = "beach vacation"
)

// ----------------------------------------------------------------------------
// Voit

func TestVoitIngest(t *testing.T) {
	tests := []struct {
		name     string
		f        *Voit
		c        *Config
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
			c: &Config{
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
			c: &Config{
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
			c: &Config{
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
			c: &Config{
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
			c: &Config{
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
			c: &Config{
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
			c: &Config{},
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
			c: &Config{
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
		{ // Regex patterns tested in regex_test.go.
			name: "patterns: regex",
			f: &Voit{
				File: File{
					Source: "/tmp/20260517_104536300.jpg",
					Name:   "20260517_104536300",
					Ext:    ".jpg",
				},
			},
			c: &Config{
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
			c: &Config{
				Pattern: "photo-ms",
			},
			wantVoit: &Voit{
				File: File{
					Source: "/tmp/no_match.jpg",
					Name:   "no_match",
					Ext:    ".jpg",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.f.Ingest(tt.c)

			if !reflect.DeepEqual(tt.f, tt.wantVoit) {
				t.Errorf("\nGot Voit:  %+v\nWant Voit: %+v", tt.f, tt.wantVoit)
			}
		})
	}
}

func TestVoitFormat(t *testing.T) {
	voitTime = time.Date(2026, time.February, 2, 12, 5, 20, 700000000, time.UTC)
	tests := []struct {
		name       string
		f          *Voit
		c          *Config
		wantFormat string
	}{
		{
			name: "sanity: sanitized format",
			f: &Voit{
				File: File{
					Source: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
					Ext:    ".jpg",
				},
				Mark: Meta{
					VTime: VTime{Time: voitTime},
					Tags:  Tag{Items: []string{"summer", "vacation", "beach"}},
					Desc:  Desc{Text: "beach vacation"},
				},
			},
			c:          &Config{},
			wantFormat: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
		},
		{
			name: "sanity: lowercase description",
			f: &Voit{
				File: File{
					Source: "/tmp/2026-02-02T12.05.20.700 Beach VACATION -- summer vacation beach.jpg",
					Ext:    ".jpg",
				},
				Mark: Meta{
					VTime: VTime{Time: voitTime},
					Tags:  Tag{Items: []string{"summer", "vacation", "beach"}},
					Desc:  Desc{Text: "Beach VACATION"},
				},
			},
			c: &Config{
				Lower: true,
			},
			wantFormat: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.f.Format(tt.c)

			if tt.f.Target != tt.wantFormat {
				t.Errorf("\nGot:  %q\nWant: %q", tt.f.Target, tt.wantFormat)
			}
		})
	}
}

// ----------------------------------------------------------------------------
// Description

func TestDescChomp(t *testing.T) {
	tests := []struct {
		name     string
		fName    string
		dSep     string
		tSep     string
		wantIdx  int
		wantText string
	}{
		{
			name:     "sanity: default format",
			fName:    "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
			wantIdx:  23,
			wantText: "beach vacation",
		},
		{
			name:     "sanity: alternative separators",
			fName:    "2026-02-02T12.05.20.700|beach vacation - summer vacation beach",
			dSep:     "|",
			tSep:     " - ",
			wantIdx:  23,
			wantText: "beach vacation",
		},
		{
			name:     "sanity: same separators",
			fName:    "2026-02-02T12.05.20.700|beach vacation|summer vacation beach",
			dSep:     "|",
			tSep:     "|",
			wantIdx:  23,
			wantText: "beach vacation",
		},
		{
			name:     "sanity: no tags",
			fName:    "2026-02-02T12.05.20.700 beach vacation",
			wantIdx:  23,
			wantText: "beach vacation",
		},
		{
			name:     "sanity: invalid tags [parsed to desc]",
			fName:    "2026-02-02T12.05.20.700 beach vacation --",
			wantIdx:  23,
			wantText: "beach vacation --",
		},
		{
			name:     "sanity: no desc [empty desc]",
			fName:    "2026-02-02T12.05.20.700  -- summer vacation beach",
			wantIdx:  23,
			wantText: "",
		},
		{
			name:     "sanity: no desc no tags [empty desc]",
			fName:    "2026-02-02T12.05.20.700 ",
			wantIdx:  23,
			wantText: "",
		},
		{
			name:     "sanity: no desc invalid separators [empty desc]",
			fName:    "2026-02-02T12.05.20.700 -- summer vacation beach",
			wantIdx:  -1,
			wantText: "",
		},
		{
			name:     "sanity: bare vtime [empty desc]",
			fName:    "2026-02-02T12.05.20.700",
			wantIdx:  -1,
			wantText: "",
		},
		{
			name:     "sanity: invalid vtime [parsed to desc]",
			fName:    "2026-02-02 12.05.20.700 beach vacation -- summer vacation beach",
			wantIdx:  10,
			wantText: "12.05.20.700 beach vacation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d Desc

			i := d.Chomp(tt.fName, tt.dSep, tt.tSep)

			if d.Text != tt.wantText {
				t.Errorf("\nText: %q\nwantText: %q", d.Text, tt.wantText)
			}

			if i != tt.wantIdx {
				t.Errorf("\nIdx:     %d\nwantIdx: %d", i, tt.wantIdx)
			}
		})
	}
}

func TestDescFormat(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		lower    bool
		sep      string
		wantText string
	}{
		{
			name:     "sanity: default format",
			text:     "beach vacation",
			wantText: " beach vacation",
		},
		{
			name:     "sanity: lowercase",
			text:     "Beach VACATION",
			lower:    true,
			wantText: " beach vacation",
		},
		{
			name:     "sanity: alternative output separator [correct format]",
			text:     "beach vacation",
			sep:      "|",
			wantText: "|beach vacation",
		},
		{
			name:     "sanity: empty desc [no string returned]",
			text:     "",
			wantText: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Desc{
				Text: tt.text,
			}

			f := d.Format(tt.lower, tt.sep)

			if f != tt.wantText {
				t.Errorf("\nFormat: %q\nwantFormat: %q", f, tt.wantText)
			}
		})
	}
}

// ----------------------------------------------------------------------------
// VTime

func TestVTimeChomp(t *testing.T) {
	voitTime := time.Date(2026, time.February, 2, 12, 5, 20, 700000000, time.UTC)
	voitTimeSpan := time.Date(2027, time.June, 18, 11, 46, 37, 400000000, time.UTC)
	tests := []struct {
		name     string
		fName    string
		sep      string
		wantTime time.Time
		wantSpan time.Time
	}{
		{
			name:     "sanity: default format",
			fName:    "2026-02-02T12.05.20.700",
			wantTime: voitTime,
			wantSpan: time.Time{},
		},
		{
			name:     "sanity: trailing spaces [parse correct]",
			fName:    "2026-02-02T12.05.20.700 ",
			wantTime: voitTime,
			wantSpan: time.Time{},
		},
		{
			name:     "sanity: time span",
			fName:    "2026-02-02T12.05.20.700--2027-06-18T11.46.37.400",
			wantTime: voitTime,
			wantSpan: voitTimeSpan,
		},
		{
			name:     "sanity: alternative separator [time span]",
			fName:    "2026-02-02T12.05.20.700|2027-06-18T11.46.37.400",
			wantTime: voitTime,
			wantSpan: voitTimeSpan,
			sep:      "|",
		},
		{
			name:     "sanity: invalid time span [no times parsed]",
			fName:    "2026-02-02T12.05.20.700--20270618T114637400",
			wantTime: time.Time{},
			wantSpan: time.Time{},
		},
		{
			name:     "sanity: invalid vtime [no times parsed]",
			fName:    "20260202T120520700--20270618T114637400",
			wantTime: time.Time{},
			wantSpan: time.Time{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var v VTime

			v.Chomp(tt.fName, tt.sep)

			if v.Time != tt.wantTime {
				t.Errorf("\nTime:     %q\nwantTime: %q", v.Time, tt.wantTime)
			}

			if v.Span != tt.wantSpan {
				t.Errorf("\nTime:     %q\nwantTime: %q", v.Span, tt.wantSpan)
			}
		})
	}
}

func TestVTimeFormat(t *testing.T) {
	voitTime := time.Date(2026, time.February, 2, 12, 5, 20, 700000000, time.UTC)
	voitTimeSpan := time.Date(2027, time.June, 18, 11, 46, 37, 400000000, time.UTC)
	tests := []struct {
		name       string
		time       VTime
		format     string
		sep        string
		wantFormat string
	}{
		{
			name: "sanity: default time",
			time: VTime{
				Time: voitTime,
			},
			wantFormat: "2026-02-02T12.05.20.700",
		},
		{
			name: "sanity: default time span",
			time: VTime{
				Time: voitTime,
				Span: voitTimeSpan,
			},
			wantFormat: "2026-02-02T12.05.20.700--2027-06-18T11.46.37.400",
		},
		{
			name: "sanity: alternative time format",
			time: VTime{
				Time: voitTime,
			},
			format:     "2006-01-02",
			wantFormat: "2026-02-02",
		},
		{
			name: "sanity: alternative separator",
			time: VTime{
				Time: voitTime,
				Span: voitTimeSpan,
			},
			sep:        "|",
			wantFormat: "2026-02-02T12.05.20.700|2027-06-18T11.46.37.400",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := tt.time.Format(tt.format, tt.sep)

			if f != tt.wantFormat {
				t.Errorf("\nFormat:     %q\nwantFormat: %q", f, tt.wantFormat)
			}
		})
	}
}

// ----------------------------------------------------------------------------
// Tag

func TestTagChomp(t *testing.T) {
	tests := []struct {
		name     string
		fName    string
		tSep     string
		wantTags []string
		wantName string
		wantIdx  int
	}{
		// Sanity checks.
		{
			name:     "sanity: sanitized format",
			fName:    "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
			wantTags: []string{"summer", "vacation", "beach"},
			wantName: "2026-02-02T12.05.20.700 - beach vacation",
			wantIdx:  38,
		},
		{
			name:     "sanity: alternative separator",
			fName:    "2026-02-02T12.05.20.700 beach vacation - summer vacation beach",
			tSep:     " - ",
			wantTags: []string{"summer", "vacation", "beach"},
			wantName: "2026-02-02T12.05.20.700 beach vacation",
			wantIdx:  38,
		},
		{
			name:     "tags: lowercased",
			fName:    "2026-02-02T12.05.20.700 beach vacation -- SUMMER VACATION BEACH",
			wantName: "2026-02-02T12.05.20.700 beach vacation",
			wantTags: []string{"summer", "vacation", "beach"},
			wantIdx:  38,
		},
		{
			name:     "tags: digikam tag spacers",
			fName:    "2026-02-02T12.05.20.700 beach vacation -- nested_tag_summer vacation beach",
			wantName: "2026-02-02T12.05.20.700 beach vacation",
			wantTags: []string{"nested_tag_summer", "vacation", "beach"},
			wantIdx:  38,
		},
		{
			name:     "tags: empty [trailing space]",
			fName:    "2026-02-02T12.05.20.700 beach vacation -- ",
			wantName: "2026-02-02T12.05.20.700 beach vacation",
			wantTags: []string{},
			wantIdx:  38,
		},
		{
			name:     "tags: empty [invalid separator]",
			fName:    "2026-02-02T12.05.20.700 beach vacation --",
			wantName: "2026-02-02T12.05.20.700 beach vacation --",
			wantTags: []string{},
			wantIdx:  -1,
		},
		{
			name:     "tags: no tags",
			fName:    "2026-02-02T12.05.20.700 beach vacation",
			wantName: "2026-02-02T12.05.20.700 beach vacation",
			wantTags: []string{},
			wantIdx:  -1,
		},
		{
			name:     "tags: de-duplicate tags",
			fName:    "2026-02-02T12.05.20.700 beach vacation -- summer summer beach",
			wantName: "2026-02-02T12.05.20.700 beach vacation",
			wantTags: []string{"summer", "beach"},
			wantIdx:  38,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tag Tag

			tIdx := tag.Chomp(tt.fName, tt.tSep)

			if tIdx != tt.wantIdx {
				t.Errorf("\nIdx:     %d\nwantIdx: %d", tIdx, tt.wantIdx)
			}

			if len(tag.Items) == 0 && len(tt.wantTags) == 0 {
				return
			}
			if !reflect.DeepEqual(tag.Items, tt.wantTags) {
				t.Errorf("\nTags:     %q\nwantTags: %q", tag.Items, tt.wantTags)
			}
		})
	}
}

func TestTagFormat(t *testing.T) {
	tests := []struct {
		name       string
		tags       []string
		sep        string
		wantFormat string
	}{
		{
			name:       "sanity: valid tags format [correct format]",
			tags:       []string{"summer", "beach"},
			wantFormat: " -- summer beach",
		},
		{
			name:       "sanity: invalid tags [correct format]",
			tags:       []string{"SUMMER", "Beach"},
			wantFormat: " -- summer beach",
		},
		{
			name:       "sanity: alternative output separator [correct format]",
			tags:       []string{"SUMMER", "Beach"},
			sep:        "|",
			wantFormat: "|summer beach",
		},
		{
			name:       "sanity: empty tags [no string returned]",
			tags:       []string{},
			wantFormat: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tag := Tag{
				Items: tt.tags,
			}

			f := tag.Format(tt.sep)

			if f != tt.wantFormat {
				t.Errorf("\nFormat:     %q\nwantFormat: %q", f, tt.wantFormat)
			}
		})
	}
}

func TestTagAdd(t *testing.T) {
	tests := []struct {
		name     string
		tags     []string
		tag      string
		wantTags []string
	}{
		{
			name:     "sanity: append a tag [tag added]",
			tags:     []string{"summer", "beach"},
			tag:      "vacation",
			wantTags: []string{"summer", "beach", "vacation"},
		},
		{
			name:     "sanity: tag not added [tags unchanged]",
			tags:     []string{"summer", "beach"},
			tag:      "summer",
			wantTags: []string{"summer", "beach"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tag := Tag{
				Items: tt.tags,
			}

			tag.Add(tt.tag)

			if len(tag.Items) == 0 && len(tt.wantTags) == 0 {
				return
			}
			if !reflect.DeepEqual(tag.Items, tt.wantTags) {
				t.Errorf("\nTags:     %q\nwantTags: %q", tag.Items, tt.wantTags)
			}

		})
	}
}

func TestTagDelete(t *testing.T) {
	tests := []struct {
		name     string
		tags     []string
		tag      string
		wantTags []string
	}{
		{
			name:     "sanity: delete a tag [tag removed]",
			tags:     []string{"summer", "beach", "vacation"},
			tag:      "vacation",
			wantTags: []string{"summer", "beach"},
		},
		{
			name:     "sanity: tag not found [tags unchanged]",
			tags:     []string{"summer", "beach"},
			tag:      "vacation",
			wantTags: []string{"summer", "beach"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tag := Tag{
				Items: tt.tags,
			}

			tag.Delete(tt.tag)

			if len(tag.Items) == 0 && len(tt.wantTags) == 0 {
				return
			}
			if !reflect.DeepEqual(tag.Items, tt.wantTags) {
				t.Errorf("\nTags:     %q\nwantTags: %q", tag.Items, tt.wantTags)
			}

		})
	}
}
