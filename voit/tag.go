// Tags contain user metadata that provides context to the related file.
//
// Rules:
// 1. Use as few tags as possible.
// 2. Limit yourself to a self-defined set of tags.
// 3. Tags within your set must not overlap.
// 4. By convention, tags are in plural.
// 5. Tags are lower-case.
// 6. Tags are single words.
// 7. Keep tags on a general level.
// 8. Omit tags that are obvious.
// 9. Use one tag language.
// 10. Explain your tags.
//
// Formatting:
// All tags are lowercase and space separated, prepended with TagSep, removing
// the tag separator if empty:
//
//   TagSep{tag} {tag} {tag}
//
// A tag may contain a visibly hidden character (SyncSpace or similar)
// to create a single tag that appears with a space. This is explicitly to
// support cases in other tools where hierarchial tags are used.
//
// Reference:
// * https://karl-voit.at/2022/01/29/How-to-Use-Tags

package voit

import (
	"fmt"
	"slices"
	"strings"
)

type Tag struct {
	Items []string // Tags (always stored lowercase).
}

// Add lowercases provided tag and adds to Items if it does not exist.
func (t *Tag) Add(tag string) {
	item := strings.ToLower(tag)

	if slices.Contains(t.Items, item) {
		return
	}
	t.Items = append(t.Items, item)
}

// Add tag synchronized from XMP metadata. These tags may be hierarchial with
// additional separators and spaces. Use configured options to parse each tag
// to a voit-acceptable tag token and call the normal Add() method.
//
// Tags are lowercased and optionally stripped of leading folder/spaces as well
// as converting folder and space separators to valid options. Tags are only
// checked for inclusion after the translation occurs.
func (t *Tag) SyncAdd(tag string, cfg ...Config) {
	c := NewConfig(cfg...)

	if c.Sync.KeepFolder {
		tag = strings.ReplaceAll(tag, c.Sync.MetaFolder, c.Sync.Folder)
	} else if i := strings.LastIndex(tag, c.Sync.MetaFolder); i != -1 {
		tag = tag[i+len(c.Sync.MetaFolder):]
	}

	if c.Sync.KeepSpace {
		tag = strings.ReplaceAll(tag, " ", c.Sync.Space)
	} else {
		tag = strings.ReplaceAll(tag, " ", "")
	}

	t.Add(tag)
}

// Delete removes provided tag if it exists.
func (t *Tag) Delete(tag string) {
	i := slices.Index(t.Items, tag)
	if i != -1 {
		t.Items = slices.Delete(t.Items, i, i+1)
	}
}

// Match checks if provided tags are present (case insensitive).
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

// Chomp parses and adds de-duped tags from a base file name (no path or
// extension) using TagSep. Returns index where tags were sliced from separator
// (-1 for no tags found).
func (t *Tag) Chomp(name string, cfg ...Config) int {
	c := NewConfig(cfg...)
	unique := make(map[string]struct{})

	idx := strings.LastIndex(name, c.Voit.TagSep)
	if idx == -1 {
		return -1
	}

	words := strings.Fields(strings.ToLower(name[idx+len(c.Voit.TagSep):]))
	for _, word := range words {
		if _, exists := unique[word]; !exists {
			unique[word] = struct{}{}
			t.Items = append(t.Items, word)
		}
	}
	return idx
}

// Format uses TagSep to return trimmed tag string.
func (t *Tag) Format(cfg ...Config) string {
	c := NewConfig(cfg...)
	if len(t.Items) == 0 {
		return ""
	}

	return fmt.Sprintf("%s%s", c.Voit.TagSep, strings.ToLower(strings.Join(t.Items, " ")))
}
