package voit

import (
	"reflect"
	"testing"
	"time"
)

func containsString(str, substr string) bool {
	return len(str) >= len(substr) && func() bool {
		for i := 0; i <= len(str)-len(substr); i++ {
			if str[i:i+len(substr)] == substr {
				return true
			}
		}
		return false
	}()
}

func TestVoitRegex(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "Date only (Year, Month, Day)",
			input: "2026-05-16",
			want:  []string{"2026-05-16", "2026", "05", "16", "", "", "", ""},
		},
		{
			name:  "Date and Hour",
			input: "2026-05-16T16",
			want:  []string{"2026-05-16T16", "2026", "05", "16", "16", "", "", ""},
		},
		{
			name:  "Date, Hour, and Minute",
			input: "2026-05-16T16.52",
			want:  []string{"2026-05-16T16.52", "2026", "05", "16", "16", "52", "", ""},
		},
		{
			name:  "Date, Hour, Minute, and Second",
			input: "2026-05-16T16.52.30",
			want:  []string{"2026-05-16T16.52.30", "2026", "05", "16", "16", "52", "30", ""},
		},
		{
			name:  "Full string down to Milliseconds",
			input: "2026-05-16T16.52.30.123",
			want:  []string{"2026-05-16T16.52.30.123", "2026", "05", "16", "16", "52", "30", "123"},
		},
		{
			name:  "Ignore leading and trailing whitespace",
			input: "   \t 2026-05-16T16.52.30.123 \n\r ",
			want:  []string{"   \t 2026-05-16T16.52.30.123 \n\r ", "2026", "05", "16", "16", "52", "30", "123"},
		},
		{
			name:  "Weird non-digit delimiters",
			input: "2026_05_16#16a52b30c123",
			want:  nil,
		},

		{
			name:  "Completely non-matching text",
			input: "Not a date string",
			want:  nil,
		},
		{
			name:  "Missing mandatory Day field",
			input: "2026-05",
			want:  nil,
		},
		{
			name:  "Skipped middle chain link (Hour missing, but Minute present)",
			input: "2026-05-16..52",
			want:  nil,
		},
		{
			name:  "Invalid Year length (3 digits)",
			input: "999-05-16",
			want:  nil,
		},
		{
			name:  "Invalid Month length (1 digit)",
			input: "2026-5-16",
			want:  nil,
		},
		{
			name:  "Invalid Millisecond length (4 digits)",
			input: "2026-05-16T16.52.30.1234",
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Patterns["voit"].FindStringSubmatch(tt.input)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("\nVoit Regex\nGot:   %#v\nWant:  %#v\nInput: %q\n", got, tt.want, tt.input)
			}
		})
	}
}

func TestPatterns(t *testing.T) {
	tests := map[string][]struct {
		name       string
		input      string
		wantMatch  bool
		wantGroups []string
	}{
		"photo-ms": {
			{
				name:       "base",
				input:      "20260517_104500123",
				wantMatch:  true,
				wantGroups: []string{"2026", "05", "17", "10", "45", "00", "123"},
			},
			{
				name:       "alternative separator",
				input:      "20260517-104500123",
				wantMatch:  true,
				wantGroups: []string{"2026", "05", "17", "10", "45", "00", "123"},
			},
			{
				name:      "missing separator",
				input:     "20260517104500123",
				wantMatch: false,
			},
			{
				name:      "short ms",
				input:     "20260517_10450012",
				wantMatch: false,
			},
		},
		"signal-ms": {
			{

				name:       "base",
				input:      "2026-05-17 10:45:00.123",
				wantMatch:  true,
				wantGroups: []string{"2026", "05", "17", "10", "45", "00", "123"},
			},
			{
				name:       "alternative separator",
				input:      "2026/05/17_10.45.00-123",
				wantMatch:  true,
				wantGroups: []string{"2026", "05", "17", "10", "45", "00", "123"},
			},
			{
				name:      "missing separators",
				input:     "20260517 10:45:00.123",
				wantMatch: false,
			},
		},
		"8601-naked-ms": {
			{
				name:       "base",
				input:      "20260517104500123",
				wantMatch:  true,
				wantGroups: []string{"2026", "05", "17", "10", "45", "00", "123"},
			},
			{
				name:      "separator",
				input:     "20260517_104500123",
				wantMatch: false,
			},
		},
		"photo": {
			{
				name:       "base",
				input:      "20260517_104500",
				wantMatch:  true,
				wantGroups: []string{"2026", "05", "17", "10", "45", "00"},
			},
			{
				name:       "ignore ms",
				input:      "20260517_104500123",
				wantMatch:  true,
				wantGroups: []string{"2026", "05", "17", "10", "45", "00"},
			},
			{
				name:      "no separators",
				input:     "20260517104500",
				wantMatch: false,
			},
		},
		"signal": {
			{
				name:       "base",
				input:      "2026-05-17 10:45:00",
				wantMatch:  true,
				wantGroups: []string{"2026", "05", "17", "10", "45", "00"},
			},
			{
				name:      "no separators",
				input:     "20260517 10:45:00",
				wantMatch: false,
			},
		},
		"8601-naked": {
			{
				name:       "base",
				input:      "20260517104500",
				wantMatch:  true,
				wantGroups: []string{"2026", "05", "17", "10", "45", "00"},
			},
			{
				name:      "missing ms",
				input:     "202605171045",
				wantMatch: false,
			},
		},
		"8601-short": {
			{
				name:       "base",
				input:      "2026-05-17T1045",
				wantMatch:  true,
				wantGroups: []string{"2026", "05", "17", "10", "45"},
			},
			{
				name:       "alternative separators",
				input:      "2026/05/17 1045",
				wantMatch:  true,
				wantGroups: []string{"2026", "05", "17", "10", "45"},
			},
			{
				name:      "missing date separators",
				input:     "20260517 1045",
				wantMatch: false,
			},
		},
		"8601": {
			{
				name:       "base",
				input:      "2026-05-17T104500",
				wantMatch:  true,
				wantGroups: []string{"2026", "05", "17", "10", "45", "00"},
			},
			{
				name:      "missing time separator",
				input:     "2026-05-17104500",
				wantMatch: false,
			},
		},
		"8601-ms": {
			{
				name:       "base",
				input:      "2026-05-17T104500123",
				wantMatch:  true,
				wantGroups: []string{"2026", "05", "17", "10", "45", "00", "123"},
			},
			{
				name:      "short ms",
				input:     "2026-05-17T10450012",
				wantMatch: false,
			},
		},
		"unix": {
			{
				name:       "base",
				input:      "1325376000",
				wantMatch:  true,
				wantGroups: []string{"1325376000"},
			},
			{
				name:       "match long (first 13 digits)",
				input:      "132537600000000000",
				wantMatch:  true,
				wantGroups: []string{"1325376000000"},
			},
		},
		"voit": {
			{
				name:       "base",
				input:      "2026-05-17T10.45.00.123",
				wantMatch:  true,
				wantGroups: []string{"2026", "05", "17", "10", "45", "00", "123"},
			},
			{
				name:       "partial date only",
				input:      "  2026-05-17  ",
				wantMatch:  true,
				wantGroups: []string{"2026", "05", "17", "", "", "", ""},
			},
			{
				name:       "partial date with hour",
				input:      "2026-05-17T10",
				wantMatch:  true,
				wantGroups: []string{"2026", "05", "17", "10", "", "", ""},
			},
			{
				name:       "partial date with hour minute",
				input:      "2026-05-17T10.45",
				wantMatch:  true,
				wantGroups: []string{"2026", "05", "17", "10", "45", "", ""},
			},
			{
				name:       "partial date with hour minute second",
				input:      "2026-05-17T10.45.00",
				wantMatch:  true,
				wantGroups: []string{"2026", "05", "17", "10", "45", "00", ""},
			},
			{
				name:      "missing minutes but has seconds",
				input:     "2026-05-17T10..00",
				wantMatch: false,
			},
		},
		"webkit-chrome": {
			{
				name:       "base",
				input:      "13253760000000000",
				wantMatch:  true,
				wantGroups: []string{"13253760000000000"},
			},
			{
				name:      "too short (16 digits)",
				input:     "1325376000000000",
				wantMatch: false,
			},
			{
				name:       "match long (first 17 digits)",
				input:      "132537600000000000",
				wantMatch:  true,
				wantGroups: []string{"13253760000000000"},
			},
		},
		"created": {
			{
				name:      "never match",
				input:     "",
				wantMatch: false,
			},
			{
				name:      "never match",
				input:     "anything",
				wantMatch: false,
			},
		},
		"modified": {
			{
				name:      "never match",
				input:     "",
				wantMatch: false,
			},
			{
				name:      "never match",
				input:     "anything",
				wantMatch: false,
			},
		},
	}

	for patternName, cases := range tests {
		re, exists := Patterns[patternName]
		if !exists {
			t.Errorf("\nMissing pattern: %q\n", patternName)
			continue
		}

		t.Run(patternName, func(t *testing.T) {
			for _, tc := range cases {
				matches := re.FindStringSubmatch(tc.input)
				matched := matches != nil

				if matched != tc.wantMatch {
					t.Errorf("\nMatch\nGot:   %t\nWant:  %t\nInput: %q\n", matched, tc.wantMatch, tc.input)
					continue
				}

				if matched && len(tc.wantGroups) > 0 {
					// matches[0] is the full match, matches[1:] are the sub-groups
					capturedGroups := matches[1:]
					if len(capturedGroups) != len(tc.wantGroups) {
						t.Errorf("\nGroups\nGot:   %d\nWant:  %d\nInput: %q\n", len(capturedGroups), len(tc.wantGroups), tc.input)
						continue
					}

					for i, expected := range tc.wantGroups {
						if capturedGroups[i] != expected {
							t.Errorf("\nGroup: %d\nGot:   %q\nWant:  %q\nInput: %q\n", i+1, capturedGroups[i], expected, tc.input)
						}
					}
				}
			}
		})
	}
}

func TestExtract(t *testing.T) {
	tests := []struct {
		name       string
		file       string
		args       []Config
		wantTime   time.Time
		wantErr    bool
		wantErrStr string
	}{
		{
			name:       "no match [pattern]",
			file:       "not-a-date-file.txt",
			args:       []Config{{Voit: VoitConfig{Pattern: "photo"}}},
			wantTime:   time.Time{},
			wantErr:    true,
			wantErrStr: "no date pattern matched",
		},
		{
			name:     "webkit-chrome match",
			file:     "13253932800000000.dat",
			args:     []Config{{Voit: VoitConfig{Pattern: "webkit-chrome"}}},
			wantTime: time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "unix match",
			file:     "1262304000.dat",
			args:     []Config{{Voit: VoitConfig{Pattern: "unix"}}},
			wantTime: time.Date(2010, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "unix match (ms)",
			file:     "1262304000123.dat",
			args:     []Config{{Voit: VoitConfig{Pattern: "unix"}}},
			wantTime: time.Date(2010, time.January, 1, 0, 0, 0, 123*int(time.Millisecond), time.UTC),
		},
		{
			name:     "full pattern parse extraction",
			file:     "20260517_112356123.jpg",
			args:     []Config{{Voit: VoitConfig{Pattern: "photo-ms"}}},
			wantTime: time.Date(2026, time.May, 17, 11, 23, 56, 123*int(time.Millisecond), time.UTC),
		},
		{
			name:     "partial pattern extraction",
			file:     "2026-05-17",
			args:     []Config{{Voit: VoitConfig{Pattern: "voit"}}},
			wantTime: time.Date(2026, time.May, 17, 0, 0, 0, 0, time.UTC),
		},
		{
			name:       "no match [ctime]",
			file:       "anything.txt",
			args:       []Config{{Voit: VoitConfig{Pattern: "created"}}},
			wantTime:   time.Time{},
			wantErr:    true,
			wantErrStr: "no date pattern matched",
		},
		{
			name:       "no match [mtime]",
			file:       "anything.txt",
			args:       []Config{{Voit: VoitConfig{Pattern: "modified"}}},
			wantTime:   time.Time{},
			wantErr:    true,
			wantErrStr: "no date pattern matched",
		},
		{
			name:       "no match [set]",
			file:       "anything.txt",
			args:       []Config{{Voit: VoitConfig{Pattern: "modified"}}},
			wantTime:   time.Time{},
			wantErr:    true,
			wantErrStr: "no date pattern matched",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTime, err := Extract(tt.file, tt.args...)

			if (err != nil) != tt.wantErr {
				t.Fatalf("\nError\nGot:  %v\nWant: %v\n", err, tt.wantErr)
			}

			if err != nil && tt.wantErrStr != "" {
				if !containsString(err.Error(), tt.wantErrStr) {
					t.Errorf("\nError string\nGot:            %q\nWant substring: %q\n", err.Error(), tt.wantErrStr)
				}
				return
			}

			if !gotTime.Equal(tt.wantTime) {
				t.Errorf("\nTime\nGot:  %v\nWant: %v\n", gotTime, tt.wantTime)
			}
		})
	}
}

func TestStrip(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		args     []Config
		wantName string
	}{
		{
			name:     "sanity: invalid pattern [original string]",
			s:        "20260517_104536300 beach vacation",
			args:     []Config{{Voit: VoitConfig{Pattern: "invalid"}}},
			wantName: "20260517_104536300 beach vacation",
		},
		{
			name:     "sanity: no match [original string]",
			s:        "20260517_104536300 beach vacation",
			args:     []Config{{Voit: VoitConfig{Pattern: "signal"}}},
			wantName: "20260517_104536300 beach vacation",
		},
		{
			name:     "sanity: match [pattern removed]",
			s:        "20260517_104536300 beach vacation",
			args:     []Config{{Voit: VoitConfig{Pattern: "photo-ms"}}},
			wantName: "beach vacation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name := Strip(tt.s, tt.args...)

			if name != tt.wantName {
				t.Fatalf("\nName\nGot:  %q\nWant: %q", name, tt.wantName)
			}
		})
	}
}
