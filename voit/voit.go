// Voit contains all file metadata pertaining to a given file, it's sidecar,
// and desired filename state. Sidecar metadata is not used for file operations
// other than keeping both files linked. Update Target struct manually using
// sidecar metadata if that data should be used instead (see sync_xmp.go).
//
// File naming structure: {VTIME}{DESC}{TAGS}.{EXT}
// * VTIME: See vtime.go.
// * DESC: See description.go.
// * TAGS: See tag.go.
// * EXT: File extension.
//
// Valid examples:
//   2024-05-17T14.31.23.342 artichoke production -- research paper.pdf
//   2026-01-03 some funny picture I found.jpg
//   2026-03-04T13.20 some installer.tar.gz
//
// Most tools that tag still use hierarchy for tagging. Ideally we don't want
// to support this, but must to keep the tool useful outside of strict CLI use
// (e.g. using it in combination with Digikam XMP syncing, etc). Sync specific
// constants supporting this behavior are separated as they are not part of the
// Voit standard.
//
// Reference:
// * https://karl-voit.at/folder-hierarchy
// * https://karl-voit.at/2022/01/29/How-to-Use-Tags/

package voit

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"trimmer.io/go-xmp/xmp"
)

const (
	DescSep = " "                       // Default Voit description separator.
	TagSep  = " -- "                    // Default Voit tag separator.
	SpanSep = "--"                      // Default Voit span separator for VTime.
	VFormat = "2006-01-02T15.04.05.000" // Default Voit VTime output format (golang reference time format).
	Pattern = "created"                 // Default Voit Regex.Patterns pattern.
)

const (
	SyncMetaFolder = "/" // Default sync metadata folder separator (in XMP sidecar).
	SyncFolder     = "➔" // Default sync folder separator (u+2794).
	SyncSpace      = "⠀" // Default sync space (non-breaking, non-space) separator (u+2800).
)

type VoitImpl struct {
	File    File // Source FS metadata.
	Sidecar File // Source sidecar FS metadata.
	Orig    Meta // Original Voit FS metadata (from File struct).
	Dest    Meta // Destination Voit FS target metadata (desired state).
	Matched bool // File matched for operation.
}

type Voit interface {
	// HasSidecar checks if sidecar metadata exists.
	HasSidecar() bool

	// HasFile checks if file metadata exists.
	HasFile() bool

	// HasBothFiles checks if file and sidecar exist.
	HasBothFiles() bool

	// IsMatched returns whether or not Voit asset flagged as a match.
	IsMatched() bool

	// Width returns the largest width of File or Sidecar determined by Name.Ext.
	Width() int

	// CollisionCount sets collision count for File objects.
	CollisionCount(c int)

	// Add file to voit struct as file or sidecar determined by metadata. Returns
	// error if a file collision occurs.
	Add(file *File) error

	// ExtractMetadata from File struct and update in place.
	ExtractMetadata(cfg ...Config)

	// ExtractXMP reads XMP sidecar data and returns an XMP document or an error
	// if the file cannot be read or parsed. If there is no Sidecar nil is
	// returned.
	ExtractXMP(cfg ...Config) (*xmp.Document, error)

	// Name returns current source names.
	SourceName() (file, sidecar string)

	// Ext returns current source extensions.
	SourceExt() (file, sidecar string)

	// Abs returns current destination absolute paths.
	Abs(cfg ...Config) (file, sidecar string)

	// Exists checks if the destinations present on disk. An asset exists
	// when any component piece that is defined (File, Sidecar, or both) is any
	// status that is not 'not existing'.
	Exists(cfg ...Config) bool

	// Rename Voit asset using Dest metadata when Matched is set. Voit assets
	// with both File and Sidecar are transactional and will rollback asset
	// changes to the original state and fail. Assets with only File or Sidecar
	// will simply fail before committing the change.
	//
	// Returns error on collisions unless overwrite is enabled. Invalid file
	// paths return error.
	Rename(w io.Writer, cfg ...Config) error

	// Render current Voit asset source and current target for File and Sidecar.
	// minWidth sets the left column minimum width. Returns the number of lines
	// Rendered.
	Render(w io.Writer, minWidth int, cfg ...Config) int
}

func NewVoit() Voit { return &VoitImpl{} }

var _ Voit = (*VoitImpl)(nil) // Compile time check to validate interface.

// HasSidecar checks if sidecar metadata exists.
func (v *VoitImpl) HasSidecar() bool {
	return v.Sidecar.Path != "" || v.Sidecar.Name != ""
}

// HasFile checks if file metadata exists.
func (v *VoitImpl) HasFile() bool {
	return v.File.Path != "" || v.File.Name != ""
}

// HasBothFiles checks if file and sidecar exist.
func (v *VoitImpl) HasBothFiles() bool {
	return v.HasFile() && v.HasSidecar()
}

// IsMatched returns whether or not Voit asset flagged as a match.
func (v *VoitImpl) IsMatched() bool {
	return v.Matched
}

// Width returns the largest width of File or Sidecar determined by Name.Ext.
func (v *VoitImpl) Width() int {
	return max((len(v.File.Name) + len(v.File.Ext)), (len(v.Sidecar.Name) + len(v.Sidecar.Ext)))
}

// CollisionCount sets collision count for File objects.
func (v *VoitImpl) CollisionCount(c int) {
	v.Dest.Desc.Count = c
}

// Add file to voit struct as file or sidecar determined by metadata. Returns
// error if a file collision occurs.
func (v *VoitImpl) Add(file *File) error {
	if file == nil {
		return fmt.Errorf("file pointer is nil")
	}

	if file.IsSidecar() {
		if v.Sidecar.Path != "" || v.Sidecar.Name != "" {
			return fmt.Errorf("collision [sidecar already exists]: %s", file.AbsPath())
		}
		v.Sidecar = *file
	} else {
		if v.File.Path != "" || v.File.Name != "" {
			return fmt.Errorf("collision [file already exists]: %s", file.AbsPath())
		}
		v.File = *file
	}

	return nil
}

// ExtractMetadata from File struct and update in place.
func (v *VoitImpl) ExtractMetadata(cfg ...Config) {
	c := NewConfig(cfg...)
	var err error

	tIdx := v.Orig.Tags.Chomp(v.File.Name, c)
	dIdx := v.Orig.Desc.Chomp(v.File.Name, c)
	if dIdx > -1 {
		v.Orig.VTime.Chomp(v.File.Name[:dIdx], c)
	} else if tIdx > -1 {
		v.Orig.VTime.Chomp(v.File.Name[:tIdx], c)
	} else {
		v.Orig.VTime.Chomp(v.File.Name, c)
	}

	switch c.Voit.Pattern {
	case "created":
		v.Orig.PTime.Time = v.File.CTime
	case "modified":
		v.Orig.PTime.Time = v.File.MTime
	case "set":
		v.Orig.PTime.Time, err = Extract(c.Voit.Set, c)
		if err != nil {
			v.Orig.PTime.Time = time.Time{}
		}
	default:
		v.Orig.PTime.Time, err = Extract(v.File.Name, c)
		if err != nil {
			v.Orig.PTime.Time = time.Time{}
		}
	}
}

// ExtractXMP reads XMP sidecar data and returns an XMP document or an error if
// the file cannot be read or parsed. If there is no Sidecar nil is returned.
func (v *VoitImpl) ExtractXMP(cfg ...Config) (*xmp.Document, error) {
	c := NewConfig(cfg...)

	if !v.HasSidecar() {
		return nil, nil
	}

	file, err := c.FS.Open(v.Sidecar.AbsPath())
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	xmpData, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	doc := xmp.NewDocument()
	err = xmp.Unmarshal(xmpData, doc)
	if err != nil {
		return nil, fmt.Errorf("error parsing XMP data: %w", err)
	}

	return doc, nil
}

// Name returns current source names.
func (v *VoitImpl) SourceName() (file, sidecar string) {
	return v.File.Name, v.Sidecar.Name
}

// Ext returns current source extensions.
func (v *VoitImpl) SourceExt() (file, sidecar string) {
	return v.File.Ext, v.Sidecar.Ext
}

// Abs returns current destination absolute paths.
func (v *VoitImpl) Abs(cfg ...Config) (file, sidecar string) {
	c := NewConfig(cfg...)

	// Dest VTime always contains the metadata to use for both File/Sidecar.
	file = filepath.Join(v.File.Path, v.Dest.Format(c)+v.File.Ext)
	sidecar = filepath.Join(v.Sidecar.Path, v.Dest.Format(c)+v.Sidecar.Ext)

	return file, sidecar
}

// Exists checks if the destinations are present on disk. An asset exists
// when any component piece that is defined (File, Sidecar, or both) is not
// explicitly returning IsNotExist.
func (v *VoitImpl) Exists(cfg ...Config) bool {
	c := NewConfig(cfg...)

	if !v.HasFile() && !v.HasSidecar() {
		return false
	}

	absFile, absSidecar := v.Abs(c)

	if v.HasFile() && v.HasSidecar() {
		_, errFile := c.FS.Stat(absFile)
		_, errSidecar := c.FS.Stat(absSidecar)

		// Only IsNotExist is valid; all other cases assume existence.
		fileExists := errFile == nil || !os.IsNotExist(errFile)
		sidecarExists := errSidecar == nil || !os.IsNotExist(errSidecar)

		return fileExists && sidecarExists
	}

	if v.HasFile() {
		if _, err := c.FS.Stat(absFile); err == nil || !os.IsNotExist(err) {
			return true
		}
	}

	if v.HasSidecar() {
		if _, err := c.FS.Stat(absSidecar); err == nil || !os.IsNotExist(err) {
			return true
		}
	}

	return false
}

// Rename Voit asset using Dest metadata when Matched is set. Voit assets
// with both File and Sidecar are transactional and will rollback asset changes
// to the original state and fail. Assets with only File or Sidecar will simply
// fail before committing the change.
//
// Returns error on collisions unless overwrite is enabled. Invalid file paths
// return error.
func (v *VoitImpl) Rename(w io.Writer, cfg ...Config) error {
	c := NewConfig(cfg...)
	if !v.Matched {
		return nil
	}

	if v.HasBothFiles() {
		return v.transactionMove(w, c)
	}

	absFile, absSidecar := v.Abs(c)

	if v.HasFile() {
		if err := v.move(w, v.File.AbsPath(), absFile, c); err != nil {
			return err
		}
	}

	if v.HasSidecar() {
		if err := v.move(w, v.Sidecar.AbsPath(), absSidecar, c); err != nil {
			return err
		}
	}

	return nil
}

// Render current Voit asset source and current target for File and Sidecar.
// minWidth sets the left column minimum width. Returns the number of lines
// Rendered.
func (v *VoitImpl) Render(w io.Writer, minWidth int, cfg ...Config) int {
	c := NewConfig(cfg...)

	bFile, bSidecar := v.SourceName()
	eFile, eSidecar := v.SourceExt()
	absFile, absSidecar := v.Abs(c)

	if v.HasBothFiles() {
		fmt.Fprintf(w, "╭%-*s ➔ %s\n", minWidth+1, bFile+eFile, filepath.Base(absFile))
		fmt.Fprintf(w, "╰%-*s ➔ %s\n", minWidth+1, bSidecar+eSidecar, filepath.Base(absSidecar))
		return 2
	}

	if v.HasFile() {
		fmt.Fprintf(w, " %-*s ➔ %s\n", minWidth+1, bFile+eFile, filepath.Base(absFile))
		return 1
	}

	if v.HasSidecar() {
		fmt.Fprintf(w, " %-*s ➔ %s\n", minWidth+1, bSidecar+eSidecar, filepath.Base(absSidecar))
		return 1
	}

	return 0
}

// Move source to dest. Returns error on collision if Overwrite is not set.
func (v *VoitImpl) move(w io.Writer, source, dest string, cfg ...Config) error {
	c := NewConfig(cfg...)

	if !c.Voit.Overwrite {
		if _, err := c.FS.Stat(dest); err == nil {
			return fmt.Errorf("Collision: %s", source)
		}
	}

	// os.Rename overwrites destination by default.
	if err := c.FS.Rename(source, dest); err != nil {
		fmt.Fprintf(w, "Error renaming: %s ➔ %s: %v\n", source, dest, err)
		return err
	} else if c.Voit.Verbose {
		fmt.Fprintf(w, "Renamed: %s ➔ %s\n", source, dest)
	}

	return nil
}

// transactionMove atomically moves Voit assets (both File and Sidecar). If
// either fails or collides, the state is rolled back.
//
// Returns error on collisions unless overwrite is enabled. Invalid file paths
// return error. Overwriting is extremely dangerous as a failed rollback
// effectively unlinks the sidecar file.
func (v *VoitImpl) transactionMove(w io.Writer, cfg ...Config) error {
	c := NewConfig(cfg...)
	dstFile, dstSidecar := v.Abs(c)
	srcFile := v.File.AbsPath()
	srcSidecar := v.Sidecar.AbsPath()

	// If not overwriting, pre-check destinations and immediately fail.
	if !c.Voit.Overwrite {
		if _, err := c.FS.Stat(dstFile); err == nil {
			return fmt.Errorf("Collision: %s", srcFile)
		}
		if _, err := c.FS.Stat(dstSidecar); err == nil {
			return fmt.Errorf("Collision: %s", srcSidecar)
		}
	}

	if err := v.move(w, srcFile, dstFile, c); err != nil {
		return err
	}

	// If the sidecar move fails, rollback the original move and fail.
	if err := v.move(w, srcSidecar, dstSidecar, c); err != nil {
		fmt.Fprintf(w, "Failed to move sidecar file, rolling back File: %v\n", dstFile)
		if rollbackErr := c.FS.Rename(dstFile, srcFile); rollbackErr != nil {
			fmt.Fprintf(w, "CRITICAL: Failed to roll back primary file: %v\n", rollbackErr)
		}
		return err
	}

	return nil
}
