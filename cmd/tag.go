package cmd

import (
	"log"
	"os"

	"github.com/r-pufky/voit/voit"
	"github.com/spf13/cobra"
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
			c := voit.Config{}.UpdateFromOpts(&voit.Cfg)

			if err := voit.Cfg.Validate(os.Stdout); err != nil {
				log.Fatalf("Validate: %v", err)
			}

			assets := voit.NewAssets()

			if err := assets.LoadDir(voit.Cfg.AbsSource, c); err != nil {
				log.Fatalf("Unable to complete source file scan: %v", err)
			} else {
				if voit.Cfg.Tag.SyncXMP {
					voit.StageXMP(os.Stdout, assets, &voit.Cfg, c)
				} else {
					voit.StageTags(assets, &voit.Cfg)
				}

				assets.PromptRename(os.Stdout, os.Stdin, c)
			}
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
	tagCmd.Flags().BoolVarP(&voit.Cfg.Tag.SyncXMP, "sync-xmp", "x", false, "Sync tags from XMP sidecar (ignores and overwrites file/sidecar filename tags)")
	tagCmd.Flags().StringVarP(&voit.Cfg.Tag.SyncMetaFolder, "sync-meta-folder", "", voit.SyncMetaFolder, "sync-xmp metadata tag folder separator (from metadata)")
	tagCmd.Flags().StringVarP(&voit.Cfg.Tag.SyncFolder, "sync-folder", "", voit.SyncFolder, "sync-xmp tag folder separator (to filename)")
	tagCmd.Flags().StringVarP(&voit.Cfg.Tag.SyncSpace, "sync-space", "", voit.SyncSpace, "sync-xmp tag space separator (to filename)")
	tagCmd.Flags().BoolVarP(&voit.Cfg.Tag.SyncKeepFolder, "sync-keep-folder", "", false, "Keep --sync-in-folder markers using runes from --sync-out-folder (otherwise folder is stripped)")
	tagCmd.Flags().BoolVarP(&voit.Cfg.Tag.SyncKeepSpace, "sync-keep-space", "", false, "Keep tag space spaces using runes from --sync-out-space (otherwise spaces are stripped)")

	tagCmd.Flags().BoolVarP(&voit.Cfg.Tag.Delete, "delete", "d", false, "Remove all tags")
	tagCmd.MarkFlagsMutuallyExclusive("add", "remove", "set", "sync-xmp", "delete")

	bindFlagsToPrefix(tagCmd.Flags(), "tag")
}
