package voit

import (
	"slices"
	"strings"
)

// Add tag synchronized from XMP metadata. These tags may be hierarchial with
// additional separators and spaces. Use configured options to parse each tag
// to a voit-acceptable tag token and call the normal Add() method.
//
// Tags are lowercased and optionally stripped of leading folder/spaces as well
// as converting folder and space separators to valid options. Tags are only
// checked for inclusion after the translation occurs.
func (t *Tag) SyncAdd(s string, o *Opts) {
	if o.Tag.SyncKeepFolder {
		s = strings.ReplaceAll(s, o.Tag.SyncInFolder, o.Tag.SyncOutFolder)
	} else if i := strings.LastIndex(s, o.Tag.SyncInFolder); i != -1 {
		s = s[i+len(o.Tag.SyncInFolder):]
	}

	if o.Tag.SyncKeepSpace {
		s = strings.ReplaceAll(s, " ", o.Tag.SyncOutSpace)
	} else {
		s = strings.ReplaceAll(s, " ", "")
	}

	t.Add(s)
}

// Add tag. Tags lowercased and checked for inclusion after translation occurs.
func (t *Tag) Add(s string) {
	tag := strings.ToLower(s)

	if slices.Contains(t.Items, tag) {
		return
	}
	t.Items = append(t.Items, tag)
}

// Delete tag.
func (t *Tag) Delete(s string) {
	i := slices.Index(t.Items, s)
	if i != -1 {
		t.Items = slices.Delete(t.Items, i, i+1)
	}
}

// Match tags (case insensitive).
func (t *Tag) Match(tags []string) bool {
	if len(tags) == 0 {
		return false
	}

	for _, match := range tags {
		if !slices.Contains(t.Items, strings.ToLower(match)) {
			return false
		}
	}

	return true
}

// Process files from given source path and stage rename transformations.
func (files VoitFiles) StageTag(opts *Opts) {
	vCfg := opts.Voit()

	for _, f := range files {
		f.Ingest(&vCfg)

		// Copy the original struct and use new reference for tags.
		f.Mark = f.Orig
		f.Mark.Tags.Items = slices.Clone(f.Orig.Tags.Items)

		if len(opts.Tag.Select) == 0 {
			f.Matched = true // No match filter, match all files.
		} else if f.Orig.Tags.Match(opts.Tag.Select) {
			f.Matched = true
		}

		if f.Matched {
			if opts.Tag.Delete {
				f.Mark.Tags.Items = []string{}
			}

			if len(opts.Tag.Add) != 0 {
				for _, tag := range opts.Tag.Add {
					f.Mark.Tags.Add(tag)
				}
			}

			if len(opts.Tag.Remove) != 0 {
				for _, tag := range opts.Tag.Remove {
					f.Mark.Tags.Delete(tag)
				}
			}

			if len(opts.Tag.Set) != 0 {
				f.Mark.Tags.Items = f.Mark.Tags.Items[:0]
				for _, tag := range opts.Tag.Set {
					f.Mark.Tags.Add(tag)
				}
			}
		}
	}

	files.ResolveCollisions(vCfg, opts.Verbose)
}
