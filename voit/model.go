/*
Model Voit file name structure.

{VTIME}[DSEP]{DESC}[TSEP]{TAGS}.{EXT}

VTIME: File creation or retrieval date (any date most relevant context). Full
  8601 format preferred:

  HHHH-MM-DDTHH.MM.SS.SSS

  But any sub-section of this format is acceptable.

  . is used instead of : for OSX and mounted FS support.
	{VTIME}--{VTIME}: date span.
DSEP: Description separator ' '.
DESC: Description. Can be empty (remove DSEP if empty). no case restrictions.
TSEP: Tag separator '--'.
TAGS: Tags. Lowercase, space separated (remove TSEP if empty).
EXT: File extension.

2024-05-17T14.31.23.342 - artichoke production -- research paper.pdf
2026-01-03 - some funny picture I found.jpg
2026-03-04T13.20 - some installer.tar.gz

https://karl-voit.at/folder-hierarchy
*/

package voit

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

const (
	DefaultDescSep = " "                       // Desc separator.
	DefaultTagsSep = " -- "                    // Tag separator.
	DefaultSpanSep = "--"                      // Span separator for vtime.
	DefaultVFormat = "2006-01-02T15.04.05.000" // VTime standard format.
	DefaultPattern = "created"                 // Regex.Patterns.
)

var MultiExts = []string{
	".tar.gz",
	".tar.xz",
	".tar.bz2",
	".min.js",
	".min.css",
	".css.map",
	".js.map",
	".sql.gz",
}

type Config struct {
	Format  string // VTime time format.
	Pattern string // Regex matching pattern.
	SSep    string // VTime span separator.
	DSep    string // Desc separator.
	TSep    string // Tag separator.
	Lower   bool   // Lowercase description.
}

type Voit struct {
	File    File   // Filesystem metadata.
	Orig    Meta   // Voit metadata from filesystem file.
	Mark    Meta   // Voit metadata after manipulations.
	Target  string // Target file name generated from Format().
	Matched bool   // File matched for potential rename operation.
}

type File struct {
	CTime  time.Time // Source create time (UTC).
	MTime  time.Time // Source modified time (UTC).
	Source string    // Source absolute path to file (/path/file.ext).
	Name   string    // Source original file name (file).
	Ext    string    // Source file extension: (.ext).
	Width  uint8     // Source file width (Linux max 255 characters).
}

// Voit metadata with all separators removed.
type Meta struct {
	VTime VTime // {VTIME} from VTIME field.
	PTime VTime // {VTIME} from regex filename parsing (only used for ingesting).
	Tags  Tag   // {TAGS} from TAG field.
	Desc  Desc  // {DESC} from DESC field.
}

type VTime struct {
	Time time.Time // {VTIME} datetime start.
	Span time.Time // {VTIME} datetime end (existence dictates time span).
}

type Tag struct {
	Items []string // {TAGS} already lowercased.
}

type Desc struct {
	Text string // {DESCRIPTION}.
}

// ----------------------------------------------------------------------------
// Voit

// Ingest Voit.Orig fields from Voit.File. Undefined options use
// DefaultDescSep, DefaultTagsSep, DefaultSpanSep, DefaultVFormat,
// DefaultPattern.
func (f *Voit) Ingest(c *Config) {
	if c.Format == "" {
		c.Format = DefaultVFormat
	}
	if c.Pattern == "" {
		c.Pattern = DefaultPattern
	}
	if len(c.DSep) == 0 {
		c.DSep = DefaultDescSep
	}
	if len(c.TSep) == 0 {
		c.TSep = DefaultTagsSep
	}
	if len(c.SSep) == 0 {
		c.SSep = DefaultSpanSep
	}
	var err error

	tIdx := f.Orig.Tags.Chomp(f.File.Name, c.TSep)
	dIdx := f.Orig.Desc.Chomp(f.File.Name, c.DSep, c.TSep)
	if dIdx > -1 {
		f.Orig.VTime.Chomp(f.File.Name[:dIdx], c.SSep)
	} else if tIdx > -1 {
		f.Orig.VTime.Chomp(f.File.Name[:tIdx], c.SSep)
	} else {
		f.Orig.VTime.Chomp(f.File.Name, c.SSep)
	}

	switch c.Pattern {
	case "created":
		f.Orig.PTime.Time = f.File.CTime
	case "modified":
		f.Orig.PTime.Time = f.File.MTime
	default:
		f.Orig.PTime.Time, err = Extract(f.File.Name, c.Pattern)
		if err != nil {
			f.Orig.PTime.Time = time.Time{}
		}
	}
}

// Format file name from current Voit.Mark struct. Undefined options use
// DefaultDescSep, DefaultTagsSep, DefaultSpanSep, DefaultVFormat,
// DefaultPattern.
func (m *Voit) Format(c *Config) {
	if c.Format == "" {
		c.Format = DefaultVFormat
	}
	if len(c.DSep) == 0 {
		c.DSep = DefaultDescSep
	}
	if len(c.TSep) == 0 {
		c.TSep = DefaultTagsSep
	}
	if len(c.SSep) == 0 {
		c.SSep = DefaultSpanSep
	}

	m.Target = filepath.Join(filepath.Dir(m.File.Source), fmt.Sprintf("%s%s%s%s", m.Mark.VTime.Format(c.Format, c.SSep), m.Mark.Desc.Format(c.Lower, c.DSep), m.Mark.Tags.Format(c.TSep), m.File.Ext))
}

// ----------------------------------------------------------------------------
// Description

// Chomp Description. Name must be base file name (no path or extension).
// Return index where description sliced from desc separator (-1 for no desc
// found). Tag separator used for outer bounds.
func (d *Desc) Chomp(name string, dSep string, tSep string) int {
	if len(dSep) == 0 {
		dSep = DefaultDescSep
	}
	if len(tSep) == 0 {
		tSep = DefaultTagsSep
	}

	tIdx := strings.LastIndex(name, tSep)
	if tIdx != -1 {
		name = name[:tIdx]
	}

	dIdx := strings.Index(name, dSep)
	if dIdx == -1 {
		d.Text = ""
	} else {
		if tIdx == -1 {
			d.Text = strings.TrimSpace(name[dIdx+len(dSep):])
		} else {
			d.Text = strings.TrimSpace(name[dIdx+len(dSep) : tIdx])
		}
	}
	return dIdx
}

// Format Desc for file naming, using desc separator, forcing lowercase if set.
func (d *Desc) Format(lower bool, s string) string {
	if len(s) == 0 {
		s = DefaultDescSep
	}

	if len(d.Text) == 0 {
		return ""
	}

	if lower {
		return fmt.Sprintf("%s%s", s, strings.ToLower(d.Text))
	}
	return fmt.Sprintf("%s%s", s, d.Text)
}

// ----------------------------------------------------------------------------
// VTime

// Chomp VTime. Name must be base file name (vtime portion only). Invalid VTime
// format will result in ZeroTime being set.
func (v *VTime) Chomp(name string, sep string) {
	if len(sep) == 0 {
		sep = DefaultSpanSep
	}
	var err error

	start, end, isSpan := strings.Cut(name, sep)

	if v.Time, err = Extract(start, "voit"); err != nil {
		return // Valid VTime will always have at least a start date.
	}

	if isSpan {
		if v.Span, err = Extract(end, "voit-span"); err != nil {
			v.Time = time.Time{} // Start date parsed, but TimeSpan requested.
			return
		}
	}
}

// Format VTime for file naming.
// Default "2006-01-02T15.04.05.000" time format and standard Voit separator.
func (v *VTime) Format(f string, s string) string {
	if len(f) == 0 {
		f = DefaultVFormat
	}
	if len(s) == 0 {
		s = DefaultSpanSep
	}

	if v.Span.IsZero() {
		return v.Time.Format(f)
	}

	return v.Time.Format(f) + s + v.Span.Format(f)
}

// ----------------------------------------------------------------------------
// Tag

// Chomp Tags. Name must be base file name (no path or extension). Return index
// where tags were sliced from separator (-1 for no tags found). Tags
// automatically de-duped.
func (t *Tag) Chomp(name string, sep string) int {
	if len(sep) == 0 {
		sep = DefaultTagsSep
	}
	unique := make(map[string]struct{})

	tIdx := strings.LastIndex(name, sep)
	if tIdx == -1 {
		return -1
	}

	words := strings.Fields(strings.ToLower(name[tIdx+len(sep):]))
	for _, word := range words {
		if _, exists := unique[word]; !exists {
			unique[word] = struct{}{}
			t.Items = append(t.Items, word)
		}
	}
	return tIdx
}

// Format Tags for file naming using tag separator.
func (t *Tag) Format(s string) string {
	if len(s) == 0 {
		s = DefaultTagsSep
	}

	if len(t.Items) == 0 {
		return ""
	}

	return fmt.Sprintf("%s%s", s, strings.ToLower(strings.Join(t.Items, " ")))
}

// Add tag.
func (t *Tag) Add(s string) {
	if slices.Contains(t.Items, s) {
		return
	}
	t.Items = append(t.Items, strings.ToLower(s))
}

// Delete tag.
func (t *Tag) Delete(s string) {
	i := slices.Index(t.Items, s)
	if i != -1 {
		t.Items = slices.Delete(t.Items, i, i+1)
	}
}
