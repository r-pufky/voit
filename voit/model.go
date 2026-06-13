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
  * Use as few tags as possible.
  * Limit yourself to a self-defined set of tags.
  * Tags within your set must not overlap.
  * By convention, tags are in plural.
  * Tags are lower-case.
  * Tags are single words.
  * Keep tags on a general level.
  * Omit tags that are obvious.
  * Use one tag language.
  * Explain your tags.
EXT: File extension.

2024-05-17T14.31.23.342 - artichoke production -- research paper.pdf
2026-01-03 - some funny picture I found.jpg
2026-03-04T13.20 - some installer.tar.gz

Reference:
* https://karl-voit.at/folder-hierarchy
* https://karl-voit.at/2022/01/29/How-to-Use-Tags/
*/

package voit

import (
	"time"
)

const (
	DefaultDescSep = " "                       // Desc separator.
	DefaultTagsSep = " -- "                    // Tag separator.
	DefaultSpanSep = "--"                      // Span separator for vtime.
	DefaultVFormat = "2006-01-02T15.04.05.000" // VTime standard format.
	DefaultPattern = "created"                 // Regex.Patterns.
)

// Most tools that tag still use hierarchy for tagging. Ideally we don't want
// to support this, but must to keep the tool useful outside of strict CLI use
// (e.g. using it in combination with Digikam XMP syncing, etc).
const (
	DefaultSyncInFolder  = "/" // Tag input folder separator (from XMP sidecar, not file).
	DefaultSyncOutFolder = "➔" // Tag folder separator (u+2794).
	DefaultSyncOutSpace  = "⠀" // Tag space (non-breaking, non-space) separator (u+2800).
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

type VoitFiles []*Voit
type FileMap map[string]*LinkedFiles

// Sidecar file mapping.
type LinkedFiles struct {
	File    *Voit // File ({NAME}.{EXT}).
	Sidecar *Voit // Sidecar ({NAME}.{EXT}.xmp) for File.
}

type Config struct {
	Format         string // VTime time format.
	Pattern        string // Regex matching pattern.
	Set            string // Set static VTime time format (set option).
	SSep           string // VTime span separator.
	DSep           string // Desc separator.
	TSep           string // Tag separator.
	SyncInFolder   string // Sync XMP Tag input folder separator (expected folder separator for tags).
	SyncOutFolder  string // Sync XMP Tag folder separator.
	SyncOutSpace   string // Sync XMP Tag space separator.
	SyncKeepFolder bool   // Tag keep nested folders when ingesting tags.
	SyncKeepSpace  bool   // Tag keep tag spaces when ingesting tags.
	Lower          bool   // Lowercase description.
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

func NewConfig() Config {
	return Config{
		Format:         DefaultVFormat,
		Pattern:        DefaultPattern,
		Set:            "",
		SSep:           DefaultSpanSep,
		DSep:           DefaultDescSep,
		TSep:           DefaultTagsSep,
		SyncInFolder:   DefaultSyncInFolder,
		SyncOutFolder:  DefaultSyncOutFolder,
		SyncOutSpace:   DefaultSyncOutSpace,
		SyncKeepFolder: false,
		SyncKeepSpace:  false,
		Lower:          false,
	}
}
