// Meta contains processed Voit metadata. This is consumed by the Voit
// interface when dealing with specific files or states.
//
// File naming structure: {VTIME}{DESC}{TAGS}.{EXT}
// * VTIME: See vtime.go.
// * DESC: See desc.go.
// * TAGS: See tag.go.
// * EXT: File extension.
//
// Valid examples:
//   2024-05-17T14.31.23.342 artichoke production -- research paper.pdf
//   2026-01-03 some funny picture I found.jpg
//   2026-03-04T13.20 some installer.tar.gz
//
// Reference:
// * https://karl-voit.at/folder-hierarchy
// * https://karl-voit.at/2022/01/29/How-to-Use-Tags/

package voit

import (
	"fmt"
)

type Meta struct {
	VTime VTime // {VTIME} from VTIME field.
	PTime VTime // {VTIME} from regex pattern.
	Tags  Tag   // {TAGS} from TAG field.
	Desc  Desc  // {DESC} from DESC field.
}

// Format Voit name (no path or extension) using metadata and provided Config.
func (m *Meta) Format(cfg ...Config) string {
	c := NewConfig(cfg...)
	return fmt.Sprintf("%s%s%s", m.VTime.Format(c), m.Desc.Format(c), m.Tags.Format(c))
}
