package config

type Opts struct {
	Yes       bool       `mapstructure:"yes"`
	Verbose   bool       `mapstructure:"verbose"`
	Build     bool       `mapstructure:"build"`
	TagSep    string     `mapstructure:"tag-sep"`
	DescSep   string     `mapstructure:"desc-sep"`
	SpanSep   string     `mapstructure:"span-sep"`
	AbsSource string     `mapstructure:"abs-source"`
	Rename    RenameOpts `mapstructure:"rename"`
	Tag       TagOpts    `mapstructure:"tag"`
}

type RenameOpts struct {
	Pattern       string `mapstructure:"pattern"`
	Lower         bool   `mapstructure:"lower"`
	Strip         bool   `mapstructure:"strip"`
	NoDesc        bool   `mapstructure:"no-desc"`
	NoTags        bool   `mapstructure:"no-tags"`
	Overwrite     bool   `mapstructure:"overwrite"`
	PreferPattern bool   `mapstructure:"prefer-pattern"`
}

type TagOpts struct {
	Add       []string `mapstructure:"add"`
	Remove    []string `mapstructure:"remove"`
	Set       []string `mapstructure:"set"`
	Select    []string `mapstructure:"select"`
	Delete    bool     `mapstructure:"delete"`
	Overwrite bool     `mapstructure:"overwrite"`
}

var Cfg Opts
