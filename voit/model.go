/*
Model Voit file name structure.

{VTIME} - {CONTEXT} -- {TAGS}.{EXT}

where
* VTIME: File creation or retrieval date (any date most relevant context). Full
  8601 format preferred:

  HHHH-MM-DDTHH.MM.SS.SSS

  But any sub-section of this format is acceptable.

  . is used instead of : for OSX and mounted FS support.
	{VTIME}--{VTIME}: date span.
* ' - ': Description separator. Can be empty.
* DESCRIPTION: Context for file with no case restrictions.
* ' -- ': Tag separator. Reequired if tags are present.
* TAGS: Tags; lowercase, space separated.
* EXT: File extension.

2024-05-17T14.31.23.342 - artichoke production -- research paper.pdf
2026-01-03 - some funny picture I found.jpg
2026-03-04T13.20 - some installer.tar.gz

https://karl-voit.at/folder-hierarchy
*/

package voit

import "time"

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

type File struct {
	CTime     time.Time // Source create time (UTC)
	MTime     time.Time // Source modified time (UTC)
	VTime     time.Time // {VTIME} datetime
	VTimeSpan time.Time // {VTIME} datetime span end
	Tags      []string  // {TAGS} already lowercased
	Desc      string    // {DESCRIPTION}.
	Source    string    // Source absolute path to file: /path/file.ext
	Name      string    // Source original file name: file
	Ext       string    // Source file extension: .ext
	Target    string    // Target absolute path to rename: /path/file.ext
	Width     uint8     // Source file width (Linux max 255 characters)
	Matched   bool      // File matched for potentia rename operation
}
