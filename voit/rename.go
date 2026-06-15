package voit

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/k0kubun/pp/v3"
)

// Process files from given source path and stage rename transformations.
func (files VoitFiles) StageRename(w io.Writer, opts *Opts) {
	vCfg := opts.Voit()

	for _, f := range files {
		f.Ingest(&vCfg)
		hasPTime := !f.Orig.PTime.Time.IsZero()
		hasVTime := !f.Orig.VTime.Time.IsZero()
		isSet := opts.Rename.Set != ""

		if isSet {
			f.Mark.VTime = f.Orig.PTime
			if opts.Verbose {
				pp.Fprintf(w, "Date source: Set (PTime), V: %v, P:%v\n", hasVTime, hasPTime)
			}
		} else if hasPTime || hasVTime {
			if opts.Rename.PreferPattern && hasPTime && hasVTime {
				f.Mark.VTime = f.Orig.PTime
				if opts.Verbose {
					pp.Fprintf(w, "Date source: PTime (preferred), V: %v, P:%v\n", hasVTime, hasPTime)
					fmt.Println("prefer pattern")
				}
			} else if hasVTime {
				f.Mark.VTime = f.Orig.VTime
				if opts.Verbose {
					pp.Fprintf(w, "Date source: VTime, V: %v, P:%v\n", hasVTime, hasPTime)
				}
			} else {
				f.Mark.VTime = f.Orig.PTime
				if opts.Verbose {
					pp.Fprintf(w, "Date source: PTime, V: %v, P:%v\n", hasVTime, hasPTime)
				}
			}
		}

		if isSet || hasPTime || hasVTime {
			f.Matched = true

			if opts.Rename.Strip {
				f.Mark.Desc.Text = Strip(f.Orig.Desc.Text, vCfg.Pattern)
			}
		}

		if opts.Rename.NoTags {
			f.Mark.Tags.Items = []string{}
		} else {
			f.Mark.Tags = f.Orig.Tags
		}

		if opts.Rename.NoDesc {
			f.Mark.Desc.Text = ""
		} else if !opts.Rename.Strip { // Only update if not stripped.
			f.Mark.Desc = f.Orig.Desc
		}
	}
}

// Rename files marked as Matched using File.Source and Mark.Target. Collisions
// are fatal unless overwrite is enabled.
func (files VoitFiles) Rename(w io.Writer, overwrite bool, verbose bool) error {
	defer timeAction(w, time.Now())
	for _, f := range files {
		if f.Matched {
			target := f.Target

			if !overwrite {
				if _, err := os.Stat(target); err == nil {
					return fmt.Errorf("Collision: %s%s\n", f.File.Source, f.File.Ext)
				}
			}

			if err := os.Rename(f.File.Source, target); err != nil {
				fmt.Fprintf(w, "Error renaming: %s%s ➔ %s%s: %v\n", f.File.Source, f.File.Ext, target, f.File.Ext, err)
			} else if verbose {
				fmt.Fprintf(w, "Renamed: %s%s ➔ %s%s\n", f.File.Source, f.File.Ext, target, f.File.Ext)
			}
		}
	}
	return nil
}

// Time file actions.
func timeAction(w io.Writer, start time.Time) {
	elapsed := time.Since(start)
	fmt.Fprintf(w, "Renamed in %s.\n", elapsed)
}
