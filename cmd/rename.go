package cmd

import (
	"fmt"
	"log"
	"maps"
	"os"
	"path/filepath"
	"slices"

	. "github.com/r-pufky/voit/config"
	"github.com/r-pufky/voit/internal"
	"github.com/r-pufky/voit/voit"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	pattern string

	renameCmd = &cobra.Command{
		Use:   "rename",
		Short: "Rename files according to file name & attribute dates",
		Long: "Rename files according to file name & attribute dates.\n\n" +
			"  {VTIME} {DESC} -- {TAGS}.{EXT}\n\n" +
			"Target files are automatically differentiated if there are name collisions.\n\n" +
			"NOTE: Pattern flag is overridden if specified in config.",
		Example: "  voit rename -s ./photos --photo-ms\n  voit rename -s image.jpg -l",

		PreRun: func(cmd *cobra.Command, args []string) {
			// Resolve all patterns to pattern variable using option name.
			for _, opt := range slices.Collect(maps.Keys(voit.Patterns)) {
				if cmd.Flags().Changed(opt) {
					Cfg.Rename.Pattern = opt
					break
				}
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			// Override default Pattern if set in config.
			if viper.IsSet("pattern") {
				Cfg.Rename.Pattern = viper.GetString("rename.pattern")
			}

			if Cfg.Rename.Pattern == "" {
				log.Fatal("a rename pattern must be specified via flags (e.g., --photo-ms) or within your config file.")
			}

			if Cfg.AbsSource == "" {
				var err error
				Cfg.AbsSource, err = os.Getwd()
				if err != nil {
					log.Fatalf("unable to get current working directory (%v).", err)
				}
			}
			absPath, err := filepath.Abs(Cfg.AbsSource)
			if err != nil {
				log.Fatalf("source does not exist (%v).", Cfg.AbsSource)
			}
			Cfg.AbsSource = absPath

			if Cfg.Verbose {
				fmt.Printf("Loaded Options: %+v\n", Cfg)
			}

			config := &voit.Config{
				Format:  voit.DefaultVFormat,
				Pattern: Cfg.Rename.Pattern,
				SSep:    Cfg.SpanSep,
				DSep:    Cfg.DescSep,
				TSep:    Cfg.TagSep,
				Lower:   Cfg.Rename.Lower,
			}

			files, err := internal.Scan(Cfg.AbsSource)
			if err != nil {
				log.Fatalf("Unable to complete source file scan: %v", err)
			}

			if len(files) == 0 {
				fmt.Println("No files matched the known datetime formats.")
				os.Exit(0)
			}

			stageRename(files, &Cfg, config)

			count := internal.DisplayPending(os.Stdout, files, config)
			if count != 0 {
				if Cfg.Rename.Overwrite {
					fmt.Printf("\nProposed changes (OVERWRITE ENABLED): %d file(s).\n", count)
				} else {
					fmt.Printf("\nProposed changes: %d file(s).\n", count)
				}

				if !Cfg.Yes && !internal.Confirm(os.Stdin, os.Stdout) {
					fmt.Println("Operation aborted by user.")
					os.Exit(0)
				}

				internal.Rename(os.Stdout, files, Cfg.Rename.Overwrite, Cfg.Verbose)
			} else {
				fmt.Println("No files matched proposed changes.")
			}
		},
	}
)

func init() {
	renameCmd.Flags().SortFlags = false
	rootCmd.AddCommand(renameCmd)

	renameCmd.Flags().Bool("photo-ms", false, "YYYYMMDD░HHMMSSSSS      │ Photos, Screenshots (ms)")
	renameCmd.Flags().Bool("photo", false, "YYYYMMDD░HHMMSS         │ Photos, Screenshots")
	renameCmd.Flags().Bool("signal-ms", false, "YYYY░MM░DD░HH░MM░SS░SSS │ Signal (ms)")
	renameCmd.Flags().Bool("signal", false, "YYYY░MM░DD░HH░MM░SS     │ Signal")
	renameCmd.Flags().Bool("8601-ms", false, "YYYY░MM░DD░HHMMSSSSS    │ Half 8601 (ms)")
	renameCmd.Flags().Bool("8601", false, "YYYY░MM░DD░HHMMSS       │ Half 8601")
	renameCmd.Flags().Bool("8601-short", false, "YYYY░MM░DD░HHMM         │ Half short 8601")
	renameCmd.Flags().Bool("8601-naked-ms", false, "YYYYMMDDHHMMSSSSS       │ Naked 8601 (ms)")
	renameCmd.Flags().Bool("8601-naked", false, "YYYYMMDDHHMMSS          │ Naked 8601")
	renameCmd.Flags().Bool("webkit-chrome", false, "SSSSSSSSSSSSSSSSS       │ Chrome Webkit Epoch")
	renameCmd.Flags().Bool("voit", false, "YYYY-MM-DDTHH.MM.SS.SSS │ Voit Scheme")
	renameCmd.Flags().Bool("voit-span", false, "{VOIT}--{VOIT}          │ Voit Scheme date span")
	renameCmd.Flags().Bool("created", false, "[ctime]                 │ Use file creation date")
	renameCmd.Flags().Bool("modified", false, "[modtime]               │ Use file modification date")
	renameCmd.MarkFlagsMutuallyExclusive(slices.Collect(maps.Keys(voit.Patterns))...)

	renameCmd.Flags().BoolVarP(&Cfg.Rename.Lower, "lower", "l", false, "Lowercase description and extension")
	renameCmd.Flags().BoolVarP(&Cfg.Rename.Strip, "strip", "", false, "Strip matched pattern from description")
	renameCmd.Flags().BoolVarP(&Cfg.Rename.NoDesc, "no-desc", "", false, "Remove description")
	renameCmd.Flags().BoolVarP(&Cfg.Rename.NoTags, "no-tags", "", false, "Remove tags")
	renameCmd.Flags().BoolVarP(&Cfg.Rename.Overwrite, "overwrite", "", false, "Overwrite existing target files (DANGEROUS)")
	renameCmd.Flags().BoolVarP(&Cfg.Rename.PreferPattern, "prefer-pattern", "p", false, "Use PATTERN date over VTIME if both exist (default: use VTIME if both exist)")

	viper.BindPFlags(rootCmd.Flags())
}

// Process files from given source path and stage rename transformations.
func stageRename(files []*voit.Voit, opts *Opts, config *voit.Config) {
	collisions := make(map[string]int)

	for i := range files {
		files[i].Ingest(config)

		if !files[i].Orig.PTime.Time.IsZero() {
			files[i].Matched = true

			if opts.Rename.PreferPattern && !files[i].Orig.VTime.Time.IsZero() {
				files[i].Mark.VTime = files[i].Orig.PTime
			}

			if opts.Rename.Strip {
				files[i].Mark.Desc.Text = voit.Strip(files[i].Orig.Desc.Text, config.Pattern)
				fmt.Printf("%q", files[i].Mark.Desc.Text)
			}
		} else if !files[i].Orig.VTime.Time.IsZero() {
			files[i].Matched = true
			files[i].Mark.VTime = files[i].Orig.VTime
		}

		if opts.Rename.NoTags {
			files[i].Mark.Tags.Items = []string{}
		} else {
			files[i].Mark.Tags = files[i].Orig.Tags
		}

		if opts.Rename.NoDesc {
			files[i].Mark.Desc.Text = ""
		} else if !opts.Rename.Strip { // Only update if not stripped.
			files[i].Mark.Desc = files[i].Orig.Desc
		}

		desc := files[i].Mark.Desc.Text
		files[i].Format(config)

		for {
			if _, exists := collisions[files[i].Target]; !exists {
				collisions[files[i].Target] = 1
				break // No Collision.
			}

			count := collisions[files[i].Target]
			collisions[files[i].Target]++

			if desc != "" {
				// Standard collision.
				files[i].Mark.Desc.Text = fmt.Sprintf("%s_%d", desc, count)
			} else if len(files[i].Mark.Tags.Items) > 0 {
				// No description, tags.
				files[i].Mark.Desc.Text = fmt.Sprintf("%d", count)
			} else {
				// No description, no tags.
				files[i].Mark.Desc.Text = fmt.Sprintf("%d", count)
			}
			files[i].Format(config)
		}
	}
}
