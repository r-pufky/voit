package voit

import (
	"testing"
	"time"
)

func TestVTimeChomp(t *testing.T) {
	voitTime := time.Date(2026, time.February, 2, 12, 5, 20, 700000000, time.UTC)
	voitTimeSpan := time.Date(2027, time.June, 18, 11, 46, 37, 400000000, time.UTC)

	tests := []struct {
		name     string
		fName    string
		args     Config
		wantTime time.Time
		wantSpan time.Time
	}{
		{
			name:     "default format",
			fName:    "2026-02-02T12.05.20.700",
			args:     Config{},
			wantTime: voitTime,
			wantSpan: time.Time{},
		},
		{
			name:     "trailing spaces [parse correct]",
			fName:    "2026-02-02T12.05.20.700 ",
			args:     Config{},
			wantTime: voitTime,
			wantSpan: time.Time{},
		},
		{
			name:     "time span",
			fName:    "2026-02-02T12.05.20.700--2027-06-18T11.46.37.400",
			args:     Config{},
			wantTime: voitTime,
			wantSpan: voitTimeSpan,
		},
		{
			name:     "alternative separator [time span]",
			fName:    "2026-02-02T12.05.20.700|2027-06-18T11.46.37.400",
			args:     Config{Voit: VoitConfig{SpanSep: "|"}},
			wantTime: voitTime,
			wantSpan: voitTimeSpan,
		},
		{
			name:     "invalid time span [no times parsed]",
			fName:    "2026-02-02T12.05.20.700--20270618T114637400",
			args:     Config{},
			wantTime: time.Time{},
			wantSpan: time.Time{},
		},
		{
			name:     "invalid vtime [no times parsed]",
			fName:    "20260202T120520700--20270618T114637400",
			args:     Config{},
			wantTime: time.Time{},
			wantSpan: time.Time{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var v VTime

			v.Chomp(tt.fName, tt.args)

			if v.Time != tt.wantTime {
				t.Errorf("\nv.Time\nGot:  %q\nWant: %q\n", v.Time, tt.wantTime)
			}

			if v.Span != tt.wantSpan {
				t.Errorf("\nv.Span\nGot:  %q\nWant: %q\n", v.Span, tt.wantSpan)
			}
		})
	}
}

func TestVTimeFormat(t *testing.T) {
	voitTime := time.Date(2026, time.February, 2, 12, 5, 20, 700000000, time.UTC)
	voitTimeSpan := time.Date(2027, time.June, 18, 11, 46, 37, 400000000, time.UTC)
	minVoitTime := time.Date(2026, time.February, 2, 12, 5, 0, 0, time.UTC)
	minVoitTimeSpan := time.Date(2027, time.June, 18, 11, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		time       VTime
		args       Config
		wantFormat string
	}{
		{
			name: "default time",
			time: VTime{
				Time: voitTime,
			},
			args:       Config{},
			wantFormat: "2026-02-02T12.05.20.700",
		},
		{
			name: "default time span",
			time: VTime{
				Time: voitTime,
				Span: voitTimeSpan,
			},
			args:       Config{},
			wantFormat: "2026-02-02T12.05.20.700--2027-06-18T11.46.37.400",
		},
		{
			name: "alternative time format",
			time: VTime{
				Time: voitTime,
			},
			args:       Config{Voit: VoitConfig{VFormat: "2006-01-02"}},
			wantFormat: "2026-02-02",
		},
		{
			name: "alternative separator",
			time: VTime{
				Time: voitTime,
				Span: voitTimeSpan,
			},
			args:       Config{Voit: VoitConfig{SpanSep: "|"}},
			wantFormat: "2026-02-02T12.05.20.700|2027-06-18T11.46.37.400",
		},
		{
			name: "minimized default time",
			time: VTime{
				Time: minVoitTime,
			},
			args:       Config{Voit: VoitConfig{Minimize: true}},
			wantFormat: "2026-02-02T12.05",
		},
		{
			name: "default time span",
			time: VTime{
				Time: minVoitTime,
				Span: minVoitTimeSpan,
			},
			args:       Config{Voit: VoitConfig{Minimize: true}},
			wantFormat: "2026-02-02T12.05--2027-06-18T11",
		},
		{
			name: "minimized alternative separator",
			time: VTime{
				Time: minVoitTime,
				Span: minVoitTimeSpan,
			},
			args:       Config{Voit: VoitConfig{SpanSep: "|", Minimize: true}},
			wantFormat: "2026-02-02T12.05|2027-06-18T11",
		},
		{
			name: "minimized alternative time format [standard format used]",
			time: VTime{
				Time: minVoitTime,
			},
			args:       Config{Voit: VoitConfig{VFormat: "2006-01-02", Minimize: true}},
			wantFormat: "2026-02-02",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := tt.time.Format(tt.args)

			if f != tt.wantFormat {
				t.Errorf("\nFormat()\nGot:  %q\nWant: %q\n", f, tt.wantFormat)
			}
		})
	}
}

func TestMinimize(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			name:     "full format [2006-01-02T15.04.05.000]",
			input:    time.Date(2026, 6, 23, 14, 5, 30, 150*int(time.Millisecond), time.UTC),
			expected: "2026-06-23T14.05.30.150",
		},
		{
			name:     "no ms [2006-01-02T15.04.05]",
			input:    time.Date(2026, 6, 23, 14, 5, 30, 0, time.UTC),
			expected: "2026-06-23T14.05.30",
		},
		{
			name:     "no s, ms [2006-01-02T15.04]",
			input:    time.Date(2026, 6, 23, 14, 5, 0, 0, time.UTC),
			expected: "2026-06-23T14.05",
		},
		{
			name:     "no m, s, ms [2006-01-02T15]",
			input:    time.Date(2026, 6, 23, 14, 0, 0, 0, time.UTC),
			expected: "2026-06-23T14",
		},
		{
			name:     "no time [2006-01-02T]",
			input:    time.Date(2026, 6, 23, 0, 0, 0, 0, time.UTC),
			expected: "2026-06-23",
		},
		{
			name:     "zero-time [2006-01-02]",
			input:    time.Time{},
			expected: "0001-01-01",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := minimize(tt.input)
			if result != tt.expected {
				t.Errorf("Minimize:  %v\nGot:  %q\nWant: %q\n", tt.input, result, tt.expected)
			}
		})
	}
}
