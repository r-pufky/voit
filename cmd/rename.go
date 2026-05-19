package cmd

import (
	"fmt"
	"log"
	"maps"
	"os"
	"path/filepath"
	"slices"

	"github.com/r-pufky/voit/internal"
	"github.com/r-pufky/voit/models"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	pattern string

	renameCmd = &cobra.Command{
		Use:   "rename",
		Short: "Rename files according to file name & attribute dates",
		Long: "Rename files according to file name & attribute dates.\n\n" +
			"  {DATE} - {DESC} -- {TAGS}.{EXT}\n\n" +
			"Target files are automatically differentiated if there are name collisions.\n\n" +
			"NOTE: Pattern flag is overridden if specified in config.",
		Example: "  voit rename -s ./photos --photo-ms\n  voit rename -s image.jpg -l",

		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			loadUserConfig()

			// Resolve all patterns to pattern variable using option name.
			for _, opt := range slices.Collect(maps.Keys(models.Patterns)) {
				if cmd.Flags().Changed(opt) {
					opts.Rename.Pattern = opt
					break
				}
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			// Override default Pattern if set in config.
			if viper.IsSet("pattern") {
				opts.Rename.Pattern = viper.GetString("pattern")
			}

			if opts.Rename.Pattern == "" {
				log.Fatal("a rename pattern must be specified via flags (e.g., --photo-ms) or within your config file.")
			}

			if opts.AbsSource == "" {
				var err error
				opts.AbsSource, err = os.Getwd()
				if err != nil {
					log.Fatalf("unable to get current working directory (%v).", err)
				}
			}
			absPath, err := filepath.Abs(opts.AbsSource)
			if err != nil {
				log.Fatalf("source does not exist (%v).", opts.AbsSource)
			}
			opts.AbsSource = absPath

			if opts.Verbose {
				fmt.Printf("Loaded Options: %+v\n", opts)
			}

			internal.Rename(opts)
		},
	}
)

func init() {
	renameCmd.Flags().SortFlags = false

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
	renameCmd.Flags().Bool("created", false, "[ctime]                 │ Use file creation date")
	renameCmd.Flags().Bool("modified", false, "[modtime]               │ Use file modification date")
	renameCmd.MarkFlagsMutuallyExclusive(slices.Collect(maps.Keys(models.Patterns))...)

	renameCmd.Flags().BoolVarP(&opts.Rename.Lower, "lower", "l", false, "Lowercase description and extension")
	renameCmd.Flags().BoolVarP(&opts.Rename.Strip, "strip", "r", false, "Strip matched pattern for target file description")
	renameCmd.Flags().BoolVarP(&opts.Rename.NoDesc, "no-desc", "n", false, "Remove description")
	renameCmd.Flags().BoolVarP(&opts.Rename.Overwrite, "overwrite", "o", false, "Overwrite existing target files (DANGEROUS)")
	renameCmd.Flags().BoolVarP(&opts.Rename.Force, "force", "f", false, "Use pattern date if both voit and pattern dates are found (default: use voit date if both exist)")

	viper.BindPFlags(rootCmd.Flags())
}

func loadUserConfig() {
	home, _ := os.UserHomeDir()
	viper.SetConfigFile(filepath.Join(home, ".config", "voit.toml"))

	if err := viper.ReadInConfig(); err == nil {
		if opts.Verbose {
			fmt.Printf("Using config file: %s\n", viper.ConfigFileUsed())
		}
	}

	// Unmarshal config into opts struct.
	if err := viper.Unmarshal(&opts); err != nil {
		fmt.Printf("Invalid config: %v\n", err)
	}
}
