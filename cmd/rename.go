package cmd

import (
	"log"
	"maps"
	"os"
	"slices"

	"github.com/r-pufky/voit/voit"
	"github.com/spf13/cobra"
)

const renameLong = `
Rename files according to file name & attribute dates.

  {VTIME} {DESC} -- {TAGS}.{EXT}

Target files are automatically differentiated if there are name collisions.
`

var (
	renameCmd = &cobra.Command{
		Use:     "rename",
		Short:   "Rename files according to file name & attribute dates",
		Long:    renameLong,
		Example: "  voit rename -s ./photos --photo-ms\n  voit rename -s image.jpg -l",

		PreRun: func(cmd *cobra.Command, args []string) {
			// Resolve all pattern flags to pattern variable using option name.
			for _, opt := range slices.Collect(maps.Keys(voit.Patterns)) {
				if cmd.Flags().Changed(opt) {
					voit.Cfg.Rename.Pattern = opt
					break
				}
			}
			if voit.Cfg.Rename.Set != "" {
				voit.Cfg.Rename.Pattern = "set"
			}
		},

		Run: func(cmd *cobra.Command, args []string) {
			c := voit.Config{}.UpdateFromOpts(&voit.Cfg)

			if err := voit.Cfg.Validate(os.Stdout); err != nil {
				log.Fatalf("Validate: %v", err)
			}

			if voit.Cfg.Rename.Pattern == "" {
				log.Fatal("a rename pattern must be specified via flags or config file (e.g., --photo-ms) or explicitly set VTIME (--set).")
			}

			assets := voit.NewAssets()

			if err := assets.LoadDir(voit.Cfg.AbsSource); err != nil {
				log.Fatalf("Unable to complete source file scan: %v", err)
			} else {
				voit.StageRename(os.Stdout, assets, &voit.Cfg, c)
				assets.PromptRename(os.Stdout, os.Stdin, c)
			}
		},
	}
)

func init() {
	renameCmd.Flags().SortFlags = false
	rootCmd.AddCommand(renameCmd)

	renameCmd.Flags().BoolVarP(&voit.Cfg.Rename.Strip, "strip", "", false, "Strip matched pattern from description")
	renameCmd.Flags().BoolVarP(&voit.Cfg.Rename.NoDesc, "no-desc", "", false, "Remove description")
	renameCmd.Flags().BoolVarP(&voit.Cfg.Rename.NoTags, "no-tags", "", false, "Remove tags")
	renameCmd.Flags().BoolVarP(&voit.Cfg.Rename.PreferPattern, "prefer-pattern", "p", false, "Use PATTERN date over VTIME if both are non-zero (default: use VTIME if both exist)")
	renameCmd.Flags().StringVarP(&voit.Cfg.Rename.Set, "set", "e", "", "Explicitly set VTIME (see --format)")

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
	renameCmd.Flags().Bool("unix", false, "SSSSSSSSSSSSS           │ Unix Epoch")
	renameCmd.Flags().Bool("voit", false, "YYYY-MM-DDTHH.MM.SS.SSS │ Voit Scheme")
	renameCmd.Flags().Bool("voit-span", false, "{VTIME}--{VTIME}        │ Voit Scheme date span")
	renameCmd.Flags().Bool("created", false, "[ctime]                 │ Use file creation date")
	renameCmd.Flags().Bool("modified", false, "[modtime]               │ Use file modification date")
	renameCmd.MarkFlagsMutuallyExclusive(slices.Collect(maps.Keys(voit.Patterns))...)

	bindFlagsToPrefix(renameCmd.Flags(), "rename")
}
