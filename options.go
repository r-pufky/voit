// Load configuration and flag options. Non-existing config locations are
// skipped.
//
// Priority:
// 1. flags
// 2. ~/.config/voit.cf
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/spf13/viper"
)

type Options struct {
	Directory string `short:"d" long:"dir" description:"The directory containing files to rename (mutually exclusive -f)." required:"false"`
	File      string `short:"f" long:"file" description:"File to rename (mutually exclusive -d)." required:"false"`
	Pattern   string `short:"p" long:"pattern" description:"Regex pattern to use." required:"false" default:"ms"`
	Lower     bool   `short:"l" long:"lower" description:"Lowercase existing filename and extension (only if transformed)." required:"false"`
	Strip     bool   `short:"s" long:"strip" description:"Strip original filename, leaving only datetime (not recommended)."`
	Yes       bool   `short:"y" long:"yes" description:"Automatically confirm operations (not recommended)." required:"false"`
	Created   bool   `short:"c" long:"created" description:"Use file creation date (fallback to modified date if not found) (not recommended)." required:"false"`
	Modified  bool   `short:"m" long:"modified" description:"Use file modification (not recommended)." required:"false"`
	Overwrite bool   `short:"o" long:"overwrite" description:"Overwrite existing target files if they exist (DANGEROUS)." required:"false"`
	Verbose   bool   `short:"v" long:"verbose" description:"Show verbose information on actions." required:"false"`
	Build     bool   `short:"b" long:"build" description:"Show build version." required:"false"`
}

func validateOptions(opts *Options) {
	patterns := []string{"ms", "mns", "mfs", "s", "ns", "fs", "w", "v"}

	if opts.Pattern != "" {
		if !slices.Contains(patterns, strings.ToLower(opts.Pattern)) {
			fmt.Printf("\n[error] invalid pattern provided (%v).", opts.Pattern)
			os.Exit(2)
		}
	}

	if opts.Build {
		fmt.Printf("Version: %s\n", Version)
		os.Exit(0)
	}

	if opts.File == "" && opts.Directory == "" {
		fmt.Println("\n[error] Either -f or -d must be provided.")
		os.Exit(2)
	}

	if opts.File != "" && opts.Directory != "" {
		fmt.Println("\n[error] -f and -d are mutually exclusive.")
		os.Exit(2)
	}

	if opts.Created && opts.Modified {
		fmt.Println("\n[error] -c and -m are mutually exclusive.")
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
}

// Parse config options. Directory and File options standardized to absolute path.
//
// Exit:
// 1: Invalid options.
// 2: Logical option error.
func Config(w io.Writer, path string) Options {
	var opts Options
	viper.SetConfigFile(path)

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintf(w, "Using: %s\n\n", viper.ConfigFileUsed())
	}

	if err := viper.Unmarshal(&opts); err != nil {
		fmt.Fprintf(w, "Invalid config: %v\n", err)
	}

	parser := flags.NewParser(&opts, flags.Default)
	parser.Usage = "[OPTIONS] \n\nRename files according to filename dates: 'YYYY-MM-DDTHH.MM.SS.SSS - {file}'"
	parser.LongDescription =
		"Target files are automatically differentiated if there are name collisions.\n\n" +
			"Set default options in: ~/.config/voit.toml\n" +
			"NOTE: Pattern flag is overridden if specified in config."

	opt := parser.FindOptionByLongName("pattern")
	opt.Description =
		"  Opt  │ Regex                   │ Common Use\n" +
			"  ms   │ YYYYMMDD░HHMMSSSSS      │ Photos (ms)\n" +
			"  s    │ YYYYMMDD░HHMMSS         │ Photos\n" +
			"  mns  │ YYYY░MM░DD░HH░MM░SS░SSS │ Signal (ms)\n" +
			"  ns   │ YYYY░MM░DD░HH░MM░SS     │ Signal\n" +
			"  hmfs │ YYYY░MM░DD░HHMMSSSSS    │ Partial 8601 (ms)\n" +
			"  hfs  │ YYYY░MM░DD░HHMMSS       │ Partial 8601\n" +
			"  hsfs │ YYYY░MM░DD░HHMM         │ Partial 8601 (short)\n" +
			"  mfs  │ YYYYMMDDHHMMSSSSS       │ Naked 8601 (ms)\n" +
			"  fs   │ YYYYMMDDHHMMSS          │ Naked 8601\n" +
			"  w    │ SSSSSSSSSSSSSSSSS       │ Chrome Webkit Epoch\n" +
			"  v    │ YYYY-MM-DDTHH.MM.SS.SSS │ Voit Scheme\n"

	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		}
		os.Exit(1)
	}

	// Override default Pattern if set in config.
	if viper.IsSet("pattern") {
		opts.Pattern = viper.GetString("pattern")
	}

	validateOptions(&opts)

	if opts.Verbose {
		fmt.Fprintf(w, "Loaded Options: %+v\n", opts)
	}
	return opts
}
