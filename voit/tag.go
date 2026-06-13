package voit

import (
	"slices"
	"strings"
)

// Add tag (case insensitive, tags automatically lowercased).
// TODO - add tag format filter for digikam (stripping invalid such as /people/person/tag)
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
