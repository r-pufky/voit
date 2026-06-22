package voit

import "testing"

func TestDescChomp(t *testing.T) {
	tests := []struct {
		name     string
		fName    string
		args     Config
		wantIdx  int
		wantText string
	}{
		{
			name:     "default format",
			fName:    "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
			args:     Config{},
			wantIdx:  23,
			wantText: "beach vacation",
		},
		{
			name:     "alternative separators",
			fName:    "2026-02-02T12.05.20.700|beach vacation - summer vacation beach",
			args:     Config{Voit: VoitConfig{DescSep: "|", TagSep: " - "}},
			wantIdx:  23,
			wantText: "beach vacation",
		},
		{
			name:     "same separators",
			fName:    "2026-02-02T12.05.20.700|beach vacation|summer vacation beach",
			args:     Config{Voit: VoitConfig{DescSep: "|", TagSep: "|"}},
			wantIdx:  23,
			wantText: "beach vacation",
		},
		{
			name:     "no tags",
			fName:    "2026-02-02T12.05.20.700 beach vacation",
			args:     Config{},
			wantIdx:  23,
			wantText: "beach vacation",
		},
		{
			name:     "invalid tags [parsed to desc]",
			fName:    "2026-02-02T12.05.20.700 beach vacation --",
			args:     Config{},
			wantIdx:  23,
			wantText: "beach vacation --",
		},
		{
			name:     "no desc [empty desc]",
			fName:    "2026-02-02T12.05.20.700  -- summer vacation beach",
			args:     Config{},
			wantIdx:  23,
			wantText: "",
		},
		{
			name:     "no desc no tags [empty desc]",
			fName:    "2026-02-02T12.05.20.700 ",
			args:     Config{},
			wantIdx:  23,
			wantText: "",
		},
		{
			name:     "no desc invalid separators [remaining file name]",
			fName:    "2026-02-02T12.05.20.700 -- summer vacation beach",
			args:     Config{},
			wantIdx:  -1,
			wantText: "2026-02-02T12.05.20.700",
		},
		{
			name:     "bare vtime [remaining file name]",
			fName:    "2026-02-02T12.05.20.700",
			args:     Config{},
			wantIdx:  -1,
			wantText: "2026-02-02T12.05.20.700",
		},
		{
			name:     "invalid vtime [parsed to desc]",
			fName:    "2026-02-02 12.05.20.700 beach vacation -- summer vacation beach",
			args:     Config{},
			wantIdx:  10,
			wantText: "12.05.20.700 beach vacation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d Desc

			i := d.Chomp(tt.fName, tt.args)

			if d.Text != tt.wantText {
				t.Errorf("\nText\nGot:  %q\nWant: %q\n", d.Text, tt.wantText)
			}

			if i != tt.wantIdx {
				t.Errorf("\nIdx\nGot:  %d\nWant: %d\n", i, tt.wantIdx)
			}
		})
	}
}

func TestDescFormat(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		count    int
		args     Config
		wantText string
	}{
		{
			name:     "default format",
			text:     "beach vacation",
			args:     Config{},
			wantText: " beach vacation",
		},
		{
			name:     "lowercase",
			text:     "Beach VACATION",
			args:     Config{Voit: VoitConfig{Lower: true}},
			wantText: " beach vacation",
		},
		{
			name:     "alternative output separator [correct format]",
			text:     "beach vacation",
			args:     Config{Voit: VoitConfig{DescSep: "|"}},
			wantText: "|beach vacation",
		},
		{
			name:     "empty desc [no string returned]",
			text:     "",
			args:     Config{},
			wantText: "",
		},
		{
			name:     "empty disc count 1 ['1']",
			text:     "",
			count:    1,
			args:     Config{},
			wantText: " 1",
		},
		{
			name:     "count 1 [append ' 1']",
			text:     "beach vacation",
			count:    1,
			args:     Config{},
			wantText: " beach vacation 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Desc{
				Text:  tt.text,
				Count: tt.count,
			}

			f := d.Format(tt.args)

			if f != tt.wantText {
				t.Errorf("\nGot:  %q\nWant: %q\n", f, tt.wantText)
			}
		})
	}
}
