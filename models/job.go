package models

// File attributes requiring for processing renaming jobs.
type Job struct {
	Dir           string // Absolute path to file directory.
	Source        string // source file basename.
	SourceAbsPath string // source file absolute path.
	Target        string // target file basename.
	TargetAbsPath string // target file absolute path.
	Width         int    // width of source file basename.
}
