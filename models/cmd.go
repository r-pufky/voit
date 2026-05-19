package models

type Opts struct {
	Yes       bool       // Show verbose information.
	Verbose   bool       // Automatically confirm operations.
	Build     bool       // Show build version.
	TagSep    string     // Tag separator (automatically wrapped with spaces).
	DescSep   string     // Description separator (automatically wrapped with spaces).
	AbsSource string     // AbsPath to Directory containing files or File to rename.
	Rename    RenameOpts // Rename options.
}

type RenameOpts struct {
	Pattern   string // Regex pattern to use for matching.
	Lower     bool   // Lowercase existing filename and extension.
	Strip     bool   // Strip original filename.
	NoDesc    bool   // Remove description from filename.
	Overwrite bool   // Overwrite existing target files.
	Force     bool   // Force pattern date use over voit date if both exist.
}
