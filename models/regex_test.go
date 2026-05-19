package models

import (
	"reflect"
	"testing"
)

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
				t.Errorf("\nRegex Match Failure on Input: %q\ngot:  %#v\nwant: %#v", tt.input, got, tt.want)
			}
		})
	}
}

type regexTestCase struct {
	name           string
	input          string
	shouldMatch    bool
	expectedGroups []string
}

func TestPatterns(t *testing.T) {
	tests := map[string][]regexTestCase{
		"photo-ms": {
			{
				name:           "base",
				input:          "20260517_104500123",
				shouldMatch:    true,
				expectedGroups: []string{"2026", "05", "17", "10", "45", "00", "123"},
			},
			{
				name:           "alternative separator",
				input:          "20260517-104500123",
				shouldMatch:    true,
				expectedGroups: []string{"2026", "05", "17", "10", "45", "00", "123"},
			},
			{
				name:        "missing separator",
				input:       "20260517104500123",
				shouldMatch: false,
			},
			{
				name:        "short ms",
				input:       "20260517_10450012",
				shouldMatch: false,
			},
		},
		"signal-ms": {
			{

				name:           "base",
				input:          "2026-05-17 10:45:00.123",
				shouldMatch:    true,
				expectedGroups: []string{"2026", "05", "17", "10", "45", "00", "123"},
			},
			{
				name:           "alternative separator",
				input:          "2026/05/17_10.45.00-123",
				shouldMatch:    true,
				expectedGroups: []string{"2026", "05", "17", "10", "45", "00", "123"},
			},
			{
				name:        "missing separators",
				input:       "20260517 10:45:00.123",
				shouldMatch: false,
			},
		},
		"8601-naked-ms": {
			{
				name:           "base",
				input:          "20260517104500123",
				shouldMatch:    true,
				expectedGroups: []string{"2026", "05", "17", "10", "45", "00", "123"},
			},
			{
				name:        "separator",
				input:       "20260517_104500123",
				shouldMatch: false,
			},
		},
		"photo": {
			{
				name:           "base",
				input:          "20260517_104500",
				shouldMatch:    true,
				expectedGroups: []string{"2026", "05", "17", "10", "45", "00"},
			},
			{
				name:           "ignore ms",
				input:          "20260517_104500123",
				shouldMatch:    true,
				expectedGroups: []string{"2026", "05", "17", "10", "45", "00"},
			},
			{
				name:        "no separators",
				input:       "20260517104500",
				shouldMatch: false,
			},
		},
		"signal": {
			{
				name:           "base",
				input:          "2026-05-17 10:45:00",
				shouldMatch:    true,
				expectedGroups: []string{"2026", "05", "17", "10", "45", "00"},
			},
			{
				name:        "no separators",
				input:       "20260517 10:45:00",
				shouldMatch: false,
			},
		},
		"8601-naked": {
			{
				name:           "base",
				input:          "20260517104500",
				shouldMatch:    true,
				expectedGroups: []string{"2026", "05", "17", "10", "45", "00"},
			},
			{
				name:        "missing ms",
				input:       "202605171045",
				shouldMatch: false,
			},
		},
		"8601-short": {
			{
				name:           "base",
				input:          "2026-05-17T1045",
				shouldMatch:    true,
				expectedGroups: []string{"2026", "05", "17", "10", "45"},
			},
			{
				name:           "alternative separators",
				input:          "2026/05/17 1045",
				shouldMatch:    true,
				expectedGroups: []string{"2026", "05", "17", "10", "45"},
			},
			{
				name:        "missing date separators",
				input:       "20260517 1045",
				shouldMatch: false,
			},
		},
		"8601": {
			{
				name:           "base",
				input:          "2026-05-17T104500",
				shouldMatch:    true,
				expectedGroups: []string{"2026", "05", "17", "10", "45", "00"},
			},
			{
				name:        "missing time separator",
				input:       "2026-05-17104500",
				shouldMatch: false,
			},
		},
		"8601-ms": {
			{
				name:           "base",
				input:          "2026-05-17T104500123",
				shouldMatch:    true,
				expectedGroups: []string{"2026", "05", "17", "10", "45", "00", "123"},
			},
			{
				name:        "short ms",
				input:       "2026-05-17T10450012",
				shouldMatch: false,
			},
		},
		"voit": {
			{
				name:           "base",
				input:          "2026-05-17T10.45.00.123",
				shouldMatch:    true,
				expectedGroups: []string{"2026", "05", "17", "10", "45", "00", "123"},
			},
			{
				name:           "partial date only",
				input:          "  2026-05-17  ",
				shouldMatch:    true,
				expectedGroups: []string{"2026", "05", "17", "", "", "", ""},
			},
			{
				name:           "partial date with hour",
				input:          "2026-05-17T10",
				shouldMatch:    true,
				expectedGroups: []string{"2026", "05", "17", "10", "", "", ""},
			},
			{
				name:           "partial date with hour minute",
				input:          "2026-05-17T10.45",
				shouldMatch:    true,
				expectedGroups: []string{"2026", "05", "17", "10", "45", "", ""},
			},
			{
				name:           "partial date with hour minute second",
				input:          "2026-05-17T10.45.00",
				shouldMatch:    true,
				expectedGroups: []string{"2026", "05", "17", "10", "45", "00", ""},
			},
			{
				name:        "missing minutes but has seconds",
				input:       "2026-05-17T10..00",
				shouldMatch: false,
			},
		},
		"webkit-chrome": {
			{
				name:           "base",
				input:          "13253760000000000",
				shouldMatch:    true,
				expectedGroups: []string{"13253760000000000"},
			},
			{
				name:        "too short (16 digits)",
				input:       "1325376000000000",
				shouldMatch: false,
			},
			{
				name:           "match long (first 17 digits)",
				input:          "132537600000000000",
				shouldMatch:    true,
				expectedGroups: []string{"13253760000000000"},
			},
		},
		"created": {
			{
				name:        "never match",
				input:       "",
				shouldMatch: false,
			},
			{
				name:        "never match",
				input:       "anything",
				shouldMatch: false,
			},
		},
		"modified": {
			{
				name:        "never match",
				input:       "",
				shouldMatch: false,
			},
			{
				name:        "never match",
				input:       "anything",
				shouldMatch: false,
			},
		},
	}

	for patternName, cases := range tests {
		re, exists := Patterns[patternName]
		if !exists {
			t.Errorf("Pattern %q defined in tests but missing from Patterns map", patternName)
			continue
		}

		t.Run(patternName, func(t *testing.T) {
			for _, tc := range cases {
				matches := re.FindStringSubmatch(tc.input)
				matched := matches != nil

				if matched != tc.shouldMatch {
					t.Errorf("Input %q: expected match = %t, got %t", tc.input, tc.shouldMatch, matched)
					continue
				}

				// If it matched and we want to verify individual capture groups
				if matched && len(tc.expectedGroups) > 0 {
					// matches[0] is the full match, matches[1:] are the sub-groups
					capturedGroups := matches[1:]
					if len(capturedGroups) != len(tc.expectedGroups) {
						t.Errorf("Input %q: expected %d capture groups, got %d", tc.input, len(tc.expectedGroups), len(capturedGroups))
						continue
					}

					for i, expected := range tc.expectedGroups {
						if capturedGroups[i] != expected {
							t.Errorf("Input %q: group %d expected %q, got %q", tc.input, i+1, expected, capturedGroups[i])
						}
					}
				}
			}
		})
	}
}
