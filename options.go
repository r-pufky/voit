package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/jessevdk/go-flags"
)

type Options struct {
	Directory string `short:"d" long:"dir" description:"The directory containing files to rename." required:"false"`
	File      string `short:"f" long:"file" description:"File to rename." required:"false"`
	Lower     bool   `short:"l" long:"lower" description:"Lowercase existing filename and extension (only if transformed)." required:"false"`
	Yes       bool   `short:"y" long:"yes" description:"Automatically confirm operations (dangerous)." required:"false"`
	Pattern   string `short:"p" long:"pattern" description:"Regex pattern to use." required:"false" default:"ms"`
}

// Parse CLI options. Directory and File options standardized to absolute path.
//
// Exit:
// 1: Invalid options.
// 2: Logical option error.
func ParseFlags() Options {
	var opts Options
	patterns := []string{"ms", "mns", "mfs", "s", "ns", "fs"}

	parser := flags.NewParser(&opts, flags.Default)

	opt := parser.FindOptionByLongName("pattern")
	opt.Description =
		"  s   - YYYYMMDD?HHMMSS\n" +
			"  ns  - YYYY?MM?DD?HH?MM?SS\n" +
			"  fs  - YYYYMMDDHHMMSS\n" +
			"  mns - YYYY?MM?DD?HH?MM?SS?SSS\n" +
			"  mfs - YYYYMMDDHHMMSSSSS\n" +
			"  ms  - YYYYMMDD?HHMMSSSSS\n"
	parser.LongDescription =
		"Parse filename dates into 'YYYY-MM-DDTHH.MM.SS.SSS - {file}'\n" +
			"missing fields are set to zero-padded 0 or 1 (0000-01-01T00.00.00.000)."

	if _, err := parser.Parse(); err != nil {
		os.Exit(1)
	}

	if opts.Pattern != "" {
		if !slices.Contains(patterns, strings.ToLower(opts.Pattern)) {
			fmt.Printf("\n[error] invalid pattern provided (%v).", opts.Pattern)
			os.Exit(2)
		}
	}

	if opts.File == "" && opts.Directory == "" {
		fmt.Println("\n[error] Either -f or -d must be provided.")
		os.Exit(2)
	}

	if opts.File != "" && opts.Directory != "" {
		fmt.Println("\n[error] -f and -d are mutually exclusive.")
		os.Exit(2)
	}

	if opts.File != "" {
		absPath, err := filepath.Abs(opts.File)
		if err != nil {
			fmt.Printf("\n[error] source file does not exist (%v).", opts.File)
			os.Exit(2)
		}

		opts.Directory, opts.File = filepath.Split(absPath)
	}

	if opts.Directory != "" {
		absPath, err := filepath.Abs(opts.Directory)
		if err != nil {
			fmt.Printf("\n[error] source directory does not exist (%v).", opts.Directory)
			os.Exit(2)
		}

		opts.Directory = absPath
	}

	return opts
}
