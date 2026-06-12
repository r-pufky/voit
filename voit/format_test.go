package voit

import (
	"testing"
	"time"
)

func TestVoitFormat(t *testing.T) {
	voitTime := time.Date(2026, time.February, 2, 12, 5, 20, 700000000, time.UTC)
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
