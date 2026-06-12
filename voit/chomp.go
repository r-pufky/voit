package voit

import (
	"strings"
	"time"
)

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

// Chomp Description. Name must be base file name (no path or extension).
// Return index where description sliced from desc separator (-1 for no desc
// found, which returns remaining name). Tag separator used for outer bounds.
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
		d.Text = name // Signal no match and set to entire filename.
	} else {
		if tIdx == -1 {
			d.Text = strings.TrimSpace(name[dIdx+len(dSep):])
		} else {
			d.Text = strings.TrimSpace(name[dIdx+len(dSep) : tIdx])
		}
	}
	return dIdx
}

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
