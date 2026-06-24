// Map file, CLI, ENV options to Voit config struct. Voit package owns config
// to simplify future additions.

package voit

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"

	"github.com/k0kubun/pp/v3"
	"trimmer.io/go-xmp/models/digikam"
)

var Cfg Opts

// CLI root options.
type Opts struct {
	DescSep   string     `mapstructure:"desc-sep"`
	TagSep    string     `mapstructure:"tag-sep"`
	SpanSep   string     `mapstructure:"span-sep"`
	VFormat   string     `mapstructure:"v-format"`
	AbsSource string     `mapstructure:"abs-source"`
	Rename    RenameOpts `mapstructure:"rename"`
	Tag       TagOpts    `mapstructure:"tag"`
	Yes       bool       `mapstructure:"yes"`
	Verbose   bool       `mapstructure:"verbose"`
	Build     bool       `mapstructure:"build"`
	Overwrite bool       `mapstructure:"overwrite"`
	Minimize  bool       `mapstructure:"minimize"`
	Lower     bool       `mapstructure:"lower"`
}

// CLI rename options.
type RenameOpts struct {
	Pattern       string `mapstructure:"pattern"`
	Set           string `mapstructure:"set"`
	Strip         bool   `mapstructure:"strip"`
	NoDesc        bool   `mapstructure:"no-desc"`
	NoTags        bool   `mapstructure:"no-tags"`
	PreferPattern bool   `mapstructure:"prefer-pattern"`
}

// CLI tag options.
type TagOpts struct {
	Add            []string `mapstructure:"add"`
	Remove         []string `mapstructure:"remove"`
	Set            []string `mapstructure:"set"`
	Select         []string `mapstructure:"select"`
	SyncXMP        bool     `mapstructure:"sync-xmp"`
	SyncMetaFolder string   `mapstructure:"sync-meta-folder"`
	SyncFolder     string   `mapstructure:"sync-folder"`
	SyncSpace      string   `mapstructure:"sync-space"`
	SyncKeepFolder bool     `mapstructure:"sync-keep-folder"`
	SyncKeepSpace  bool     `mapstructure:"sync-keep-space"`
	Delete         bool     `mapstructure:"delete"`
}

// Validate received CLI options.
func (o *Opts) Validate(w io.Writer) error {
	switch o.Tag.SyncFolder {
	case o.SpanSep, o.DescSep, o.TagSep, o.Tag.SyncSpace, " ", "\t", "\n", "\v", "\f", "\r":
		return fmt.Errorf("tag-folder must be unique non-whitespace character (%v)", o.Tag.SyncFolder)
	}
	switch o.Tag.SyncSpace {
	case o.SpanSep, o.DescSep, o.TagSep, o.Tag.SyncFolder, " ", "\t", "\n", "\v", "\f", "\r":
		return fmt.Errorf("tag-space must be unique non-whitespace character (%v)", o.Tag.SyncSpace)
	}

	if o.AbsSource == "" {
		var err error
		o.AbsSource, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("unable to get current working directory: %w", err)
		}
	}

	absPath, err := filepath.Abs(o.AbsSource)
	if err != nil {
		return fmt.Errorf("source does not exist: %v", o.AbsSource)
	}
	o.AbsSource = absPath

	if o.Verbose {
		pp.Fprintf(w, "Parsed Options: %v\n", o)
	}

	return nil
}

// UpdateFromOpts returns a Config struct updated with CLI options and optional
// existing Config. Optional Config used to set non-default values and mocks.
func (c Config) UpdateFromOpts(opts *Opts, cfg ...Config) Config {
	c = NewConfig(cfg...)

	if opts.DescSep != "" {
		c.Voit.DescSep = opts.DescSep
	}
	if opts.TagSep != "" {
		c.Voit.TagSep = opts.TagSep
	}
	if opts.SpanSep != "" {
		c.Voit.SpanSep = opts.SpanSep
	}
	if opts.VFormat != "" {
		c.Voit.VFormat = opts.VFormat
	}
	if opts.Rename.Set != "" {
		c.Voit.Set = opts.Rename.Set
	}
	if opts.Rename.Pattern != "" {
		c.Voit.Pattern = opts.Rename.Pattern
	} else {
		c.Voit.Pattern = "voit"
	}
	c.Voit.Verbose = opts.Verbose
	c.Voit.Overwrite = opts.Overwrite
	c.Voit.Minimize = opts.Minimize
	c.Voit.Lower = opts.Lower
	c.Voit.Yes = opts.Yes

	if opts.Tag.SyncMetaFolder != "" {
		c.Sync.MetaFolder = opts.Tag.SyncMetaFolder
	}
	if opts.Tag.SyncFolder != "" {
		c.Sync.Folder = opts.Tag.SyncFolder
	}
	if opts.Tag.SyncSpace != "" {
		c.Sync.Space = opts.Tag.SyncSpace
	}
	c.Sync.KeepFolder = opts.Tag.SyncKeepFolder
	c.Sync.KeepSpace = opts.Tag.SyncKeepSpace

	return c
}

// StageRename processes rename command and stages rename transformations.
func StageRename(w io.Writer, assets Assets, opts *Opts, cfg ...Config) {
	c := Config{}.UpdateFromOpts(opts, cfg...)

	a, ok := assets.(*AssetImpl)
	if !ok {
		return
	}

	for _, voit := range a.m {
		v, ok := voit.(*VoitImpl)
		if !ok {
			continue
		}

		hasPTime := !v.Orig.PTime.Time.IsZero()
		hasVTime := !v.Orig.VTime.Time.IsZero()
		isSet := opts.Rename.Set != ""

		if isSet {
			v.Dest.VTime = v.Orig.PTime
			if opts.Verbose {
				pp.Fprintf(w, "Date source: Set (PTime), V: %v, P:%v\n", hasVTime, hasPTime)
			}
		} else if hasPTime || hasVTime {
			if opts.Rename.PreferPattern && hasPTime && hasVTime {
				v.Dest.VTime = v.Orig.PTime
				if opts.Verbose {
					pp.Fprintf(w, "Date source: PTime (preferred), V: %v, P:%v\n", hasVTime, hasPTime)
					fmt.Println("prefer pattern")
				}
			} else if hasVTime {
				v.Dest.VTime = v.Orig.VTime
				if opts.Verbose {
					pp.Fprintf(w, "Date source: VTime, V: %v, P:%v\n", hasVTime, hasPTime)
				}
			} else {
				v.Dest.VTime = v.Orig.PTime
				if opts.Verbose {
					pp.Fprintf(w, "Date source: PTime, V: %v, P:%v\n", hasVTime, hasPTime)
				}
			}
		}

		if isSet || hasPTime || hasVTime {
			v.Matched = true

			if opts.Rename.Strip {
				v.Dest.Desc.Text = Strip(v.Orig.Desc.Text, c)
			}
		}

		if opts.Rename.NoTags {
			v.Dest.Tags.Items = []string{}
		} else {
			v.Dest.Tags = v.Orig.Tags
		}

		if opts.Rename.NoDesc {
			v.Dest.Desc.Text = ""
		} else if !opts.Rename.Strip { // Only update if not stripped.
			v.Dest.Desc = v.Orig.Desc
		}
	}
}

// StageTags processes tag command and stages rename transformations.
func StageTags(assets Assets, opts *Opts) {
	a, ok := assets.(*AssetImpl)
	if !ok {
		return
	}

	for _, voit := range a.m {
		v, ok := voit.(*VoitImpl)
		if !ok {
			continue
		}

		// Copy the original struct and use new reference for tags.
		v.Dest = v.Orig
		v.Dest.Tags.Items = slices.Clone(v.Orig.Tags.Items)

		if len(opts.Tag.Select) == 0 {
			v.Matched = true // No match filter, match all files.
		} else if v.Orig.Tags.Match(opts.Tag.Select) {
			v.Matched = true
		}

		if v.Matched {
			if opts.Tag.Delete {
				v.Dest.Tags.Items = []string{}
			}

			if len(opts.Tag.Add) != 0 {
				for _, tag := range opts.Tag.Add {
					v.Dest.Tags.Add(tag)
				}
			}

			if len(opts.Tag.Remove) != 0 {
				for _, tag := range opts.Tag.Remove {
					v.Dest.Tags.Delete(tag)
				}
			}

			if len(opts.Tag.Set) != 0 {
				v.Dest.Tags.Items = v.Dest.Tags.Items[:0]
				for _, tag := range opts.Tag.Set {
					v.Dest.Tags.Add(tag)
				}
			}
		}
	}
}

// StageXMP processes XMP data and stages rename transformations.
func StageXMP(w io.Writer, assets Assets, opts *Opts, cfg ...Config) {
	c := Config{}.UpdateFromOpts(opts, cfg...)

	a, ok := assets.(*AssetImpl)
	if !ok {
		return
	}

	for _, voit := range a.m {
		v, ok := voit.(*VoitImpl)
		if !ok {
			continue
		}

		if v.HasSidecar() {
			var tags Tag
			doc, err := v.ExtractXMP(c)

			if err != nil {
				fmt.Fprintf(w, "Warning - failed to parse %s: %v\n", v.Sidecar.AbsPath(), err)
				continue
			}

			if model := digikam.FindModel(doc); model != nil && len(model.TagsList) > 0 {
				for _, tag := range model.TagsList {
					tags.SyncAdd(tag, c)
				}
				v.Matched = true
			}
		}
	}
}
