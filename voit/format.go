package voit

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Format file name from current Voit.Mark struct. Undefined options use
// DefaultDescSep, DefaultTagsSep, DefaultSpanSep, DefaultVFormat,
// DefaultPattern.
func (m *Voit) Format(c *Config) {
	m.Target = filepath.Join(
		filepath.Dir(m.File.Source),
		fmt.Sprintf("%s%s%s%s",
			m.Mark.VTime.Format(c.Format, c.SSep),
			m.Mark.Desc.Format(c.Lower, c.DSep),
			m.Mark.Tags.Format(c.TSep),
			m.File.Ext))
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
