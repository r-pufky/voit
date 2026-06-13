package voit

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/k0kubun/pp/v3"
)

// Voit package owns configure; cmd is a consumer that sets options.
type Opts struct {
	Yes       bool       `mapstructure:"yes"`
	Verbose   bool       `mapstructure:"verbose"`
	Build     bool       `mapstructure:"build"`
	TagSep    string     `mapstructure:"tag-sep"`
	DescSep   string     `mapstructure:"desc-sep"`
	SpanSep   string     `mapstructure:"span-sep"`
	AbsSource string     `mapstructure:"abs-source"`
	Format    string     `mapstructure:"format"`
	Lower     bool       `mapstructure:"lower"`
	Overwrite bool       `mapstructure:"overwrite"`
	Rename    RenameOpts `mapstructure:"rename"`
	Tag       TagOpts    `mapstructure:"tag"`
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
	Add     []string `mapstructure:"add"`
	Remove  []string `mapstructure:"remove"`
	Set     []string `mapstructure:"set"`
	Select  []string `mapstructure:"select"`
	SyncXMP bool     `mapstructure:"sync-xmp"`
	Delete  bool     `mapstructure:"delete"`
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
