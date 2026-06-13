package cmd

import (
	"log"
	"os"

	"github.com/r-pufky/voit/voit"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const tagLong = `
Tag files according to filters.

  {VTIME} {DESC} -- {TAGS}.{EXT}

Target files are automatically differentiated if there are name collisions.
`

var (
	tagCmd = &cobra.Command{
		Use:     "tag",
		Short:   "tag files according to filters",
		Long:    tagLong,
		Example: "  voit tag -e party\n  voit tag -s ./photos -l cake -a candles -a bday\n  voit tag --sync-xmp",

		Run: func(cmd *cobra.Command, args []string) {
			if err := voit.Cfg.Validate(); err != nil {
				log.Fatalf("Validate: %v", err)
			}

			files, err := voit.Scan(voit.Cfg.AbsSource)
			if err != nil {
				log.Fatalf("Unable to complete source file scan: %v", err)
			}

			if voit.Cfg.Tag.SyncXMP {
				files.SyncXMP(&voit.Cfg)
			} else {
				files.StageTag(&voit.Cfg)
			}

			files.PromptRename(os.Stdout, os.Stdin, &voit.Cfg)
		},
	}
)

func init() {
	tagCmd.Flags().SortFlags = false
	rootCmd.AddCommand(tagCmd)

	tagCmd.Flags().StringSliceVarP(&voit.Cfg.Tag.Add, "add", "a", []string{}, "Add specified tags")
	tagCmd.Flags().StringSliceVarP(&voit.Cfg.Tag.Remove, "remove", "r", []string{}, "Remove specified tags")
	tagCmd.Flags().StringSliceVarP(&voit.Cfg.Tag.Set, "set", "e", []string{}, "Set tags to specified tags")
	tagCmd.Flags().StringSliceVarP(&voit.Cfg.Tag.Select, "select", "c", []string{}, "Perform operations only on files with matching tags (default: all)")
	tagCmd.Flags().BoolVarP(&voit.Cfg.Tag.SyncXMP, "sync-xmp", "x", false, "Sync tags from XMP sidecar (overwrites existing tags)")
	tagCmd.Flags().BoolVarP(&voit.Cfg.Tag.Delete, "delete", "d", false, "Remove all tags")
	tagCmd.MarkFlagsMutuallyExclusive("add", "remove", "set", "sync-xmp", "delete")
	viper.BindPFlags(tagCmd.Flags())
}
