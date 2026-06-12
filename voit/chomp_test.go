package voit

import (
	"reflect"
	"testing"
	"time"
)

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
			name:     "sanity: no desc invalid separators [remaining file name]",
			fName:    "2026-02-02T12.05.20.700 -- summer vacation beach",
			wantIdx:  -1,
			wantText: "2026-02-02T12.05.20.700",
		},
		{
			name:     "sanity: bare vtime [remaining file name]",
			fName:    "2026-02-02T12.05.20.700",
			wantIdx:  -1,
			wantText: "2026-02-02T12.05.20.700",
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
