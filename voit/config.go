package voit

// Voit config options including interfaces to use.
type Config struct {
	Voit VoitConfig // Voit package configuration.
	Sync SyncConfig // Sync XMP metadata.
	FS   FileSystem // FileSystem interface.
	Unix Unix       // Unix interface.
}

// Voit config.
type VoitConfig struct {
	DescSep   string // Desc separator.
	TagSep    string // Tag separator.
	SpanSep   string // VTime span separator.
	VFormat   string // VTime time format.
	Pattern   string // Regex matching pattern.
	Set       string // Set static VTime time format (set option).
	Verbose   bool   // Show extra details during execution.
	Overwrite bool   // Force overwrite existing files.
	Minimize  bool   // Minimize VTIME to use only defined fields.
	Lower     bool   // Lowercase description.
	Yes       bool   // Automatically confirm operations.
}

// Sync XMP metadata config.
type SyncConfig struct {
	MetaFolder string // Sync XMP Tag input folder separator (expected folder separator for tags).
	Folder     string // Sync XMP Tag folder separator.
	Space      string // Sync XMP Tag space separator.
	KeepFolder bool   // Tag keep nested folders when ingesting tags.
	KeepSpace  bool   // Tag keep tag spaces when ingesting tags.
}

// NewConfig returns a Config struct with package default options set.
// Optional Config used to set non-default values and mocks.
func NewConfig(cfg ...Config) Config {
	// No performance impact for using custom or default config in func.
	// func Example(cfg ...Config) {
	//   c := NewConfig(cfg...)

	c := Config{
		Voit: VoitConfig{
			DescSep:   DescSep,
			TagSep:    TagSep,
			SpanSep:   SpanSep,
			VFormat:   VFormat,
			Pattern:   Pattern,
			Set:       "",
			Verbose:   false,
			Overwrite: false,
			Minimize:  false,
			Lower:     false,
			Yes:       false,
		},
		Sync: SyncConfig{
			MetaFolder: SyncMetaFolder,
			Folder:     SyncFolder,
			Space:      SyncSpace,
			KeepFolder: false,
			KeepSpace:  false,
		},
		FS:   RealFS{},
		Unix: RealUnix{},
	}

	// Only update if provided config has non-default values.
	if len(cfg) > 0 {
		user := cfg[0]

		if user.Voit.DescSep != "" {
			c.Voit.DescSep = user.Voit.DescSep
		}
		if user.Voit.TagSep != "" {
			c.Voit.TagSep = user.Voit.TagSep
		}
		if user.Voit.SpanSep != "" {
			c.Voit.SpanSep = user.Voit.SpanSep
		}
		if user.Voit.VFormat != "" {
			c.Voit.VFormat = user.Voit.VFormat
		}
		if user.Voit.Pattern != "" {
			c.Voit.Pattern = user.Voit.Pattern
		}
		if user.Voit.Set != "" {
			c.Voit.Set = user.Voit.Set
		}
		c.Voit.Verbose = cfg[0].Voit.Verbose
		c.Voit.Overwrite = cfg[0].Voit.Overwrite
		c.Voit.Minimize = cfg[0].Voit.Minimize
		c.Voit.Lower = cfg[0].Voit.Lower
		c.Voit.Yes = cfg[0].Voit.Yes

		if user.Sync.MetaFolder != "" {
			c.Sync.MetaFolder = user.Sync.MetaFolder
		}
		if user.Sync.Folder != "" {
			c.Sync.Folder = user.Sync.Folder
		}
		if user.Sync.Space != "" {
			c.Sync.Space = user.Sync.Space
		}
		c.Sync.KeepFolder = cfg[0].Sync.KeepFolder
		c.Sync.KeepSpace = cfg[0].Sync.KeepSpace

		if cfg[0].FS != nil {
			c.FS = cfg[0].FS
		}
		if cfg[0].Unix != nil {
			c.Unix = cfg[0].Unix
		}
	}

	return c
}

// WithPattern constructs Config struct with Pattern set.
func (c Config) WithPattern(pattern string) Config {
	c.Voit.Pattern = pattern // Use by value to return instance only change.
	return c
}
