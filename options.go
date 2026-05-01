package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jessevdk/go-flags"
)

type Options struct {
	Directory string `short:"d" long:"dir" description:"The directory containing files to rename." required:"false"`
	File      string `short:"f" long:"file" description:"File to rename." required:"false"`
	Vertical  bool   `short:"v" long:"vertical" description:"Vertically align renamed files on year instead of ➔." required:"false"`
	Lower     bool   `short:"l" long:"lower" description:"lowercase existing filename and extension." required:"false"`
	Upper     bool   `short:"u" long:"upper" description:"uppercase existing filename and extension." required:"false"`
	Yes       bool   `short:"y" long:"yes" description:"Automatically confirm operations." required:"false"`
}

func ParseFlags() Options {
	var opts Options
	parser := flags.NewParser(&opts, flags.Default)
	parser.LongDescription = "Parse filename dates into 'YYYY-MM-DDTHH.MM.SS.SSS - {file}'"

	if _, err := parser.Parse(); err != nil {
		os.Exit(1)
	}

	if opts.File == "" && opts.Directory == "" {
		fmt.Println("\n[error] Either -f or -d must be provided.")
		os.Exit(2)
	}

	if opts.File != "" && opts.Directory != "" {
		fmt.Println("\n[error] The flags -f and -d are mutually exclusive.")
		os.Exit(2)
	}

	if opts.Lower && opts.Upper {
		fmt.Println("\n[error] The flags -l and -u are mutually exclusive.")
		os.Exit(2)
	}

	// Use provided file's directory as working dir if no directory used.
	if opts.File != "" {
		opts.Directory = filepath.Dir(opts.File)
	}

	return opts
}
