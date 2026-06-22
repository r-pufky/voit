// Assets manage each voit file asset. All voit operations should be done via
// Assets.

package voit

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/k0kubun/pp/v3"
)

type Assets interface {
	// MaxWidth returns the largest file width contained within assets. Returns 0
	// if there are no assets.
	MaxWidth() int

	// LoadDir uses the provided path to load file metadata non-recursively into
	// assets. Globbing is supported. Assets populated with complete Voit
	// structs or return an error.
	LoadDir(f string, cfg ...Config) error

	// HasAssetDestCollision checks if key asset target collides with any other
	// asset targets. The key asset is not checked.
	HasAssetDestCollision(key string, cfg ...Config) bool

	// ResolveCollisions handles naming conflicts. A collision is detected when
	// an asset has the same dest as another asset, and when said dest will
	// clobber a file on disk.
	//
	// Collisions are resolved by appending ' #' to the end of the description,
	// where # is the number of skips, starting at one, until no collision. Both
	// file and sidecar are modified in tandem.
	ResolveCollisions(w io.Writer, cfg ...Config)

	// Display source and target file name changes in column. Returns total
	// matched file count.
	DisplayPending(w io.Writer, cfg ...Config) int

	// Prompt and rename on matched files after resolving collisions. Returns
	// error if a catastrophic FS error occurs.
	PromptRename(w io.Writer, r io.Reader, cfg ...Config) error
}

type AssetImpl struct {
	m     map[string]Voit // Voit assets to track.
	width int             // Max width of source Name.Ext assets.
}

func NewAssets() Assets {
	return &AssetImpl{
		m: make(map[string]Voit),
	}
}

var _ Assets = (*AssetImpl)(nil) // Compile time check to validate interface.

// MaxWidth returns the largest file width contained within assets. Returns 0
// if the asset map is empty.
func (ai *AssetImpl) MaxWidth() int {
	return ai.width
}

// LoadDir uses the provided path to load file metadata non-recursively into
// assets. Globbing is supported. Assets populated with complete Voit
// structs or return an error.
func (ai *AssetImpl) LoadDir(f string, cfg ...Config) error {
	c := NewConfig(cfg...)
	if ai == nil || ai.m == nil {
		return fmt.Errorf("asset is uninitialized: call NewAssets() before Scan")
	}

	matches, err := filepath.Glob(f)
	if err != nil {
		return fmt.Errorf("invalid glob pattern: %w", err)
	}

	// Bare directories require globbing: /tmp ➔ /tmp/*.
	if len(matches) == 1 {
		stat, err := c.FS.Stat(matches[0])
		if err == nil && stat.IsDir() {
			matches, err = filepath.Glob(filepath.Join(matches[0], "*"))
			if err != nil {
				return err
			}
		}
	}

	for _, path := range matches {
		f, err := NewFile(path, c)
		if err != nil {
			if strings.Contains(err.Error(), "directory:") {
				continue // Skip directories.
			}
			return err // All other errors fatal.
		}

		key := f.Key()
		v, ok := ai.m[key]

		if !ok {
			v = NewVoit()
			ai.m[key] = v
		}

		v.Add(f)
		v.ExtractMetadata(c)

		if w := v.Width(); w > ai.width {
			ai.width = w
		}
	}

	if len(ai.m) == 0 {
		return fmt.Errorf("No files matched the known datetime formats (is globbing quoted?)")
	}

	return nil
}

// HasAssetDestCollision checks if key asset target collides with any other
// asset targets. The key asset is not checked.
func (ai *AssetImpl) HasAssetDestCollision(key string, cfg ...Config) bool {
	c := NewConfig(cfg...)

	kVoit, exists := ai.m[key]
	if !exists {
		return false
	}

	tFile, tSidecar := kVoit.Abs(c)

	for searchKey := range ai.m {
		if searchKey == key {
			continue
		}

		sFile, sSidecar := ai.m[searchKey].Abs(c)

		if tFile == sFile || tSidecar == sSidecar {
			return true
		}
	}
	return false
}

// ResolveCollisions handles naming conflicts. A collision is detected when an
// asset has the same dest as another asset, and when said dest will clobber a
// file on disk.
//
// Collisions are resolved by appending ' #' to the end of the description,
// where # is the number of skips, starting at one, until no collision. Both
// file and sidecar are modified in tandem.
func (ai *AssetImpl) ResolveCollisions(w io.Writer, cfg ...Config) {
	c := NewConfig(cfg...)

	processed := make(map[string]bool)

	for key, voit := range ai.m {
		if processed[key] {
			continue
		}

		count := 0

		for {
			FSCollision := voit.Exists(c)
			assetCollision := ai.HasAssetDestCollision(key, c)

			if c.Voit.Verbose {
				pp.Fprintf(w, "\nKey: %s\nFS collision: %t\nAsset collision: %t\nAsset: %s\n", key, FSCollision, assetCollision, voit)
			}

			if !FSCollision && !assetCollision {
				processed[key] = true
				break
			}

			// Collision found.
			count++
			voit.CollisionCount(count)
		}

		if c.Voit.Verbose && voit.IsMatched() {
			file, _ := voit.Abs(c)
			pp.Fprintf(w, "Matched: %v\n", file)
		}
	}
}

// Display source and target file name changes in column. Returns total matched
// file count.
func (ai *AssetImpl) DisplayPending(w io.Writer, cfg ...Config) int {
	c := NewConfig(cfg...)

	var i int

	for _, v := range ai.m {
		if v.IsMatched() {
			i += v.Render(w, ai.width, c)
		}
	}

	return i
}

// Prompt and rename on matched files after resolving collisions. Returns
// error if a catastrophic FS error occurs.
func (ai *AssetImpl) PromptRename(w io.Writer, r io.Reader, cfg ...Config) error {
	c := NewConfig(cfg...)

	ai.ResolveCollisions(w, c)
	count := ai.DisplayPending(w, c)

	if count != 0 {
		if c.Voit.Overwrite {
			fmt.Fprintf(w, "\nProposed changes (OVERWRITE ENABLED): %d file(s).\n", count)
		} else {
			fmt.Fprintf(w, "\nProposed changes: %d file(s).\n", count)
		}

		if !c.Voit.Yes && !Confirm(w, r) {
			fmt.Fprintln(w, "Operation aborted by user.")
			return nil // Exit success.
		}

		for _, v := range ai.m {
			if err := v.Rename(w, c); err != nil {
				return err
			}
		}
	} else {
		fmt.Fprintln(w, "No files matched proposed changes.")
	}

	return nil
}

// Display confirmation dialog and wait for user input.
func Confirm(w io.Writer, r io.Reader) bool {
	fmt.Fprint(w, "Proceed? (y/n): ")
	var input string
	fmt.Fscanln(r, &input)

	input = strings.ToLower(strings.TrimSpace(input))
	return input == "y" || input == "yes"
}

// Time file actions.
func timeAction(w io.Writer, start time.Time) {
	fmt.Fprintf(w, "Renamed in %s.\n", time.Since(start))
}
