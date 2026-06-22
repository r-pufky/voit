// VTime contains file creation, retrieval date, or a span of those dates (any
// date most relevant to context). Does not need to align with file attributes
// or metadata within the file itself.
//
// Formatting:
// VTime always appears first using VFormat to format the datetime. VTime can
// be a span of times (VTime[SpanSep]VTime) which indicates the file contents
// within the range of time.
//
//	{VTime}
//	{VTime}SpanSep{VTime}
//
// VTime is a distinct implementation of ISO8601, replacing : with . to support
// MacOS filesystems. Any portion of a datetime up to a full ISO8601 datetime
// with milliseconds is valid:
//
//   2024-05-17T14.31.23.342
//   2026-01-03
//   2026-03-04T13.20
//
// Reference:
// * https://karl-voit.at/folder-hierarchy

package voit

import (
	"strings"
	"time"
)

type VTime struct {
	Time time.Time // VTime datetime start.
	Span time.Time // VTime datetime end (existence dictates time span).
}

// Chomp parses VTime or VTime span from a base file name (vtime portion only)
// using SpanSep. Invalid VTime format sets ZeroTime.
func (v *VTime) Chomp(name string, cfg ...Config) {
	c := NewConfig(cfg...)
	var err error

	start, end, isSpan := strings.Cut(name, c.Voit.SpanSep)

	if v.Time, err = Extract(start, c.WithPattern("voit")); err != nil {
		return // Valid VTime will always have at least a start date.
	}

	if isSpan {
		if v.Span, err = Extract(end, c.WithPattern("voit-span")); err != nil {
			v.Time = time.Time{} // Start date parsed, but TimeSpan requested.
			return
		}
	}
}

// Format uses SpanSep and VFormat to return a valid VTime string.
func (v *VTime) Format(cfg ...Config) string {
	c := NewConfig(cfg...)
	if v.Span.IsZero() {
		return v.Time.Format(c.Voit.VFormat)
	}

	return v.Time.Format(c.Voit.VFormat) + c.Voit.SpanSep + v.Span.Format(c.Voit.VFormat)
}
