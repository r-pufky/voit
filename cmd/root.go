// Naked command options.
package cmd

import (
	"fmt"
	"os"

	"github.com/r-pufky/voit/internal"
	"github.com/r-pufky/voit/models"
	"github.com/spf13/cobra"
)

var (
	Version = "development"
	opts    models.Opts
)

func init() {
	rootCmd.PersistentFlags().SortFlags = false

	rootCmd.PersistentFlags().StringVarP(&opts.AbsSource, "source", "s", "", "Directory containing files or File to rename (default: current directory)")
	rootCmd.PersistentFlags().StringVarP(&opts.TagSep, "tag-sep", "", internal.DefaultTagsSep, "Tag separator")
	rootCmd.PersistentFlags().StringVarP(&opts.DescSep, "desc-sep", "", internal.DefaultDescSep, "Description separator")
	rootCmd.PersistentFlags().StringVarP(&opts.SpanSep, "span-sep", "", internal.DefaultSpanSep, "VTIME date span separator")
	rootCmd.PersistentFlags().BoolVarP(&opts.Verbose, "verbose", "v", false, "Show verbose information")
	rootCmd.PersistentFlags().BoolVarP(&opts.Yes, "yes", "y", false, "Automatically confirm operations")

	rootCmd.Flags().BoolVarP(&opts.Build, "build", "b", false, "Show build version.")

	rootCmd.AddCommand(renameCmd)
}

var rootCmd = &cobra.Command{
	Use:   "voit",
	Short: "Voit file naming utility.",
	Long: "Manage filenames using Karl Voit's Managing Digital Files scheme.\n\n" +
		"  {VTIME} {DESC} -- {TAGS}.{EXT}\n\n" +
		"Read about structure and PhD thesis:\n" +
		"  https://karl-voit.at/folder-hierarchy\n" +
		"  https://karl-voit.at/tagstore/en/papers.shtml\n\n" +
		"Source: https://github.com/r-pufky/voit\n\n" +
		"Set default options in: ~/.config/voit.toml (see README.md)\n",
	Run: func(cmd *cobra.Command, args []string) {
		if opts.Build {
			fmt.Printf("Version: %s\n", Version)
			os.Exit(0)
		}
		cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
