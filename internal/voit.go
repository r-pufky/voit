/*
Parse string file names and extract Voit naming scheme.

{DATE} [DESSEP] {DESCRIPTION} [TAGSEP] {TAGS}.{EXT}

dessep: Description separator. May be Empty. Default: ' - '.
tagsep: Tag separator. Default: ' -- '.

date: File creation or retrieval date (any date most relevant context). Full

	8601 format preferred:

	HHHH-MM-DDTHH.MM.SS.SSS

	Any sub-section of this format is acceptable.

	. is used instead of : for OSX and mounted FS support.

context: context for file (description); no case restrictions.
tags: Tags; lowercase, space separated.

2024-05-17T14.31.23.342 - artichoke production -- research paper.pdf
2026-01-03 - some funny picture I found.jpg
2026-03-04T13.20 - some installer.tar.gz

https://karl-voit.at/folder-hierarchy
*/
package internal

import (
	"fmt"
	"log"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/r-pufky/voit/models"
)

const (
	DefaultDescSep    = " "         // Filename separator for description
	DefaultTagsSep    = " -- "      // Filename separator for tags
	DefaultSpanSep    = "--"        // Span separator for vtime
	WebkitEpochOffset = 11644473600 // Secs between Jan 1, 1601 and Jan 1, 1970
	MsPerSec          = 1000000
)

// Parse voit struct from name.
//  1. File.Tags update with tags or []string{} if not found.
//  2. File.Desc update with desc or all unmatched text from TAGS and VTIME.
//  3. file.VTime update with parsed pattern/VTIME or time.Time{} if not found.
//     If prefer enabled an existing VTIME will be overwritten with found
//     pattern time when different.
func Parse(file *models.File, pattern string, prefer bool, dSep string, tSep string) {
	if len(dSep) == 0 {
		dSep = DefaultDescSep
	}
	if len(tSep) == 0 {
		tSep = DefaultTagsSep
	}
	var name string
	var pTime, vTime time.Time
	var err error

	// Chomp Tags if exist.
	tIdx := strings.LastIndex(file.Name, tSep)
	if tIdx == -1 {
		name = file.Name
		file.Tags = []string{}
	} else {
		name = file.Name[:tIdx]
		file.Tags = strings.Fields(strings.ToLower(file.Name[tIdx+len(tSep):]))
	}

	// Parse pattern date.
	switch pattern {
	case "created":
		pTime = file.CTime
	case "modified":
		pTime = file.MTime
	default:
		pTime, err = extract(name, pattern)
		if err != nil {
			pTime = time.Time{}
		}
	}

	// Chomp Desc if exist.
	dIdx := strings.Index(name, dSep)
	if dIdx == -1 {
		file.Desc = ""
	} else {
		name = file.Name[:dIdx]
		if tIdx == -1 {
			file.Desc = strings.TrimSpace(file.Name[dIdx+len(dSep):])
		} else {
			file.Desc = strings.TrimSpace(file.Name[dIdx+len(dSep) : tIdx])
		}
	}

	// Remaining is either a pure vtime, or invalid desc / vtime + desc.
	if date, err := extract(name, "voit"); err == nil {
		vTime = date
	} else {
		vTime = time.Time{}
		file.Desc = strings.TrimSpace(name)
	}

	if prefer && !pTime.IsZero() {
		file.VTime = pTime
	} else if !vTime.IsZero() {
		file.VTime = vTime
	} else {
		file.VTime = pTime
	}
}

// Extract time object from given string and filter.
func extract(name string, pattern string) (time.Time, error) {
	match := models.Patterns[pattern].FindStringSubmatch(name)
	if match == nil {
		return time.Time{}, fmt.Errorf("No date pattern matched file: %s", name)
	}

	// 0 - full match, 1..N - Regex match groups. Unmatched groups ("") and out
	// of bounds matches are set to 0.
	group := func(i int) int {
		if match == nil || i >= len(match) {
			return 0
		}
		val, _ := strconv.Atoi(match[i])
		return val
	}

	if pattern == "webkit-chrome" {
		ms, err := strconv.ParseInt(match[1], 10, 64)
		if err != nil {
			return time.Time{}, fmt.Errorf("Invalid Webkit format: %s", name)
		}
		sec := (ms / MsPerSec) - WebkitEpochOffset
		return time.Unix(sec, (ms%MsPerSec)*1000).UTC(), nil
	}

	return time.Date(
		group(1),
		time.Month(group(2)),
		group(3),
		group(4),
		group(5),
		group(6),
		group(7)*int(time.Millisecond),
		time.UTC,
	), nil
}

// Generate File.Target modifying no other attributes. Lower lowercases
// extension and description. Strip removes matched pattern from description.
func GenTargetName(file *models.File, pattern string, lower bool, strip bool, NoDesc bool, NoTags bool, dSep string, tSep string) {
	source, err := filepath.Abs(filepath.Dir(file.Source))
	if err != nil {
		log.Fatalf("Failed to set absolute path: %v", err)
	}
	date := fmt.Sprintf("%s", file.VTime.Format("2006-01-02T15.04.05.000"))
	desc := file.Desc
	tags := strings.Join(file.Tags, " ")
	ext := file.Ext

	if lower {
		desc = strings.ToLower(desc)
		ext = strings.ToLower(ext)
	}

	file.Matched = models.Patterns[pattern].MatchString(desc)

	if strip && file.Matched {
		// Return all non-regex matched strings, removing empty strings, and join them.
		pieces := models.Patterns[pattern].Split(desc, -1)
		desc = strings.Join(slices.DeleteFunc(pieces, func(s string) bool { return s == "" }), "")
	}

	if !NoDesc && desc != "" {
		desc = filepath.Join(dSep + desc)
	} else {
		desc = ""
	}

	if !NoTags && len(file.Tags) != 0 {
		tags = filepath.Join(tSep + tags)
	} else {
		tags = ""
	}

	file.Target = filepath.Join(source, date+desc+tags+ext)
}
