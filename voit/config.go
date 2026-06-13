package voit

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/k0kubun/pp/v3"
)

// Voit package owns configure; cmd is a consumer that sets options.
type Opts struct {
	TagSep    string     `mapstructure:"tag-sep"`
	DescSep   string     `mapstructure:"desc-sep"`
	SpanSep   string     `mapstructure:"span-sep"`
	AbsSource string     `mapstructure:"abs-source"`
	Format    string     `mapstructure:"format"`
	Rename    RenameOpts `mapstructure:"rename"`
	Tag       TagOpts    `mapstructure:"tag"`
	Yes       bool       `mapstructure:"yes"`
	Verbose   bool       `mapstructure:"verbose"`
	Build     bool       `mapstructure:"build"`
	Overwrite bool       `mapstructure:"overwrite"`
	Lower     bool       `mapstructure:"lower"`
}

type RenameOpts struct {
	Pattern       string `mapstructure:"pattern"`
	Set           string `mapstructure:"set"`
	Strip         bool   `mapstructure:"strip"`
	NoDesc        bool   `mapstructure:"no-desc"`
	NoTags        bool   `mapstructure:"no-tags"`
	PreferPattern bool   `mapstructure:"prefer-pattern"`
}

type TagOpts struct {
	Add            []string `mapstructure:"add"`
	Remove         []string `mapstructure:"remove"`
	Set            []string `mapstructure:"set"`
	Select         []string `mapstructure:"select"`
	SyncXMP        bool     `mapstructure:"sync-xmp"`
	SyncInFolder   string   `mapstructure:"sync-in-folder"`
	SyncOutFolder  string   `mapstructure:"sync-out-folder"`
	SyncOutSpace   string   `mapstructure:"sync-out-space"`
	SyncKeepFolder bool     `mapstructure:"sync-keep-folder"`
	SyncKeepSpace  bool     `mapstructure:"sync-keep-space"`
	Delete         bool     `mapstructure:"delete"`
}

var Cfg Opts

// Return voit config using parsed options with model default values if unset.
func (o *Opts) Voit() Config {
	c := NewConfig()

	if o.Format != "" {
		c.Format = o.Format
	}
	if o.SpanSep != "" {
		c.SSep = o.SpanSep
	}
	if o.DescSep != "" {
		c.DSep = o.DescSep
	}
	if o.TagSep != "" {
		c.TSep = o.TagSep
	}
	if o.Tag.SyncInFolder != "" {
		c.SyncInFolder = o.Tag.SyncInFolder
	}
	if o.Tag.SyncOutFolder != "" {
		c.SyncOutFolder = o.Tag.SyncOutFolder
	}
	if o.Tag.SyncOutSpace != "" {
		c.SyncOutSpace = o.Tag.SyncOutSpace
	}
	c.SyncKeepFolder = o.Tag.SyncKeepFolder
	c.SyncKeepSpace = o.Tag.SyncKeepSpace
	c.Lower = o.Lower

	if o.Rename.Set != "" {
		c.Set = o.Rename.Set
	}
	if o.Rename.Pattern != "" {
		c.Pattern = o.Rename.Pattern
	} else {
		c.Pattern = "voit"
	}
	return c
}

// Validate received options.
func (o *Opts) Validate() error {
	switch o.Tag.SyncOutFolder {
	case o.SpanSep, o.DescSep, o.TagSep, o.Tag.SyncOutSpace, " ", "\t", "\n", "\v", "\f", "\r":
		return fmt.Errorf("tag-folder must be unique non-whitespace character (%v).", o.Tag.SyncOutFolder)
	}
	switch o.Tag.SyncOutSpace {
	case o.SpanSep, o.DescSep, o.TagSep, o.Tag.SyncOutFolder, " ", "\t", "\n", "\v", "\f", "\r":
		return fmt.Errorf("tag-space must be unique non-whitespace character (%v).", o.Tag.SyncOutSpace)
	}

	if o.AbsSource == "" {
		var err error
		o.AbsSource, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("unable to get current working directory (%v).", err)
		}
	}

	absPath, err := filepath.Abs(o.AbsSource)
	if err != nil {
		return fmt.Errorf("source does not exist (%v).", o.AbsSource)
	}
	o.AbsSource = absPath

	if o.Verbose {
		pp.Printf("Parsed Config: %v\nVoit Config: %v\n", o, o.Voit())
	}

	return nil
}
