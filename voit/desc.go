// Description is a free-form field that describes the contents of the related
// file. Tag-like information should be moved to tags.
//
// Formatting:
// All descriptions appear between VTime and Tags; optionally lowercased.
// The DescTag separator is removed if empty:
//
//   DescSep{description}
//
// Spaces are allowed. If the DescSep is a space, the separator is the first
// space after a valid VTime. Collision count is only inserted if > 0.
//
// Reference:
// * https://karl-voit.at/folder-hierarchy

package voit

import (
	"fmt"
	"strings"
)

type Desc struct {
	Text  string // Description.
	Count int    // Collision count.
}

// Chomp parses description from a base file name (no path or extension) using
// DescSep, TagSep. Returns index where description sliced from DescSep (-1 for
// not found). TagSep separator used for outer bounds.
func (d *Desc) Chomp(name string, cfg ...Config) int {
	c := NewConfig(cfg...)
	tIdx := strings.LastIndex(name, c.Voit.TagSep)
	if tIdx != -1 {
		name = name[:tIdx]
	}

	dIdx := strings.Index(name, c.Voit.DescSep)
	if dIdx == -1 {
		d.Text = name // No match set entire name.
	} else {
		if tIdx == -1 {
			d.Text = strings.TrimSpace(name[dIdx+len(c.Voit.DescSep):])
		} else {
			d.Text = strings.TrimSpace(name[dIdx+len(c.Voit.DescSep) : tIdx])
		}
	}
	return dIdx
}

// Format uses DescSep to return trimmed description including collision count,
// lowercasing if set.
func (d *Desc) Format(cfg ...Config) string {
	c := NewConfig(cfg...)
	var count string

	if len(d.Text) == 0 {
		if d.Count == 0 {
			return ""
		} else {
			return fmt.Sprintf("%s%d", c.Voit.DescSep, d.Count)
		}
	}

	if d.Count > 0 {
		count = fmt.Sprintf(" %d", d.Count)
	}

	if c.Voit.Lower {
		return fmt.Sprintf("%s%s%s", c.Voit.DescSep, strings.ToLower(d.Text), count)
	}
	return fmt.Sprintf("%s%s%s", c.Voit.DescSep, d.Text, count)
}
