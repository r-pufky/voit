package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	. "github.com/r-pufky/voit/config"
	"github.com/r-pufky/voit/internal"
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
		Example: "  voit tag -s ./photos -e party\n  voit tag -s ./photos -l cake -a candles -a bday",

		PreRun: func(cmd *cobra.Command, args []string) {
			loadUserConfig()
			// Ensure tags are lowercased for file or CLI source.
			if cmd.Flags().Changed("add") || viper.IsSet("add") {
				Cfg.Tag.Add = normalize(viper.GetStringSlice("add"))
				viper.Set("add", Cfg.Tag.Add)
			}

			if cmd.Flags().Changed("remove") || viper.IsSet("remove") {
				Cfg.Tag.Remove = normalize(viper.GetStringSlice("remove"))
				viper.Set("remove", Cfg.Tag.Remove)
			}

			if cmd.Flags().Changed("set") || viper.IsSet("set") {
				Cfg.Tag.Set = normalize(viper.GetStringSlice("set"))
				viper.Set("set", Cfg.Tag.Set)
			}

			if cmd.Flags().Changed("select") || viper.IsSet("select") {
				Cfg.Tag.Select = normalize(viper.GetStringSlice("select"))
				viper.Set("select", Cfg.Tag.Select)
			}
		},

		Run: func(cmd *cobra.Command, args []string) {
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
				fmt.Printf("Parsed Config: %+v [voit: %+v]\n", Cfg, Cfg.Voit())
			}

			files, err := internal.Scan(Cfg.AbsSource)
			if err != nil {
				log.Fatalf("Unable to complete source file scan: %v", err)
			}

			if len(files) == 0 {
				fmt.Println("No files matched the known datetime formats.")
				os.Exit(0)
			}

			stageTag(files, &Cfg)

			count := internal.DisplayPending(os.Stdout, files)
			if count != 0 {
				if Cfg.Overwrite {
					fmt.Printf("\nProposed changes (OVERWRITE ENABLED): %d file(s).\n", count)
				} else {
					fmt.Printf("\nProposed changes: %d file(s).\n", count)
				}

				if !Cfg.Yes && !internal.Confirm(os.Stdin, os.Stdout) {
					fmt.Println("Operation aborted by user.")
					os.Exit(0)
				}

				internal.Rename(os.Stdout, files, Cfg.Overwrite, Cfg.Verbose)
			} else {
				fmt.Println("No files matched proposed changes.")
			}
		},
	}
)

func init() {
	tagCmd.Flags().SortFlags = false
	rootCmd.AddCommand(tagCmd)

	tagCmd.Flags().StringSliceVarP(&Cfg.Tag.Add, "add", "a", []string{}, "Add specified tags")
	tagCmd.Flags().StringSliceVarP(&Cfg.Tag.Remove, "remove", "r", []string{}, "Remove specified tags")
	tagCmd.Flags().StringSliceVarP(&Cfg.Tag.Set, "set", "e", []string{}, "Set tags to specified tags")
	tagCmd.Flags().StringSliceVarP(&Cfg.Tag.Select, "select", "c", []string{}, "Perform operations only on files with matching tags (default: all)")
	tagCmd.Flags().BoolVarP(&Cfg.Tag.Delete, "delete", "d", false, "Remove all tags")
	tagCmd.MarkFlagsMutuallyExclusive("add", "remove", "set", "delete")
	viper.BindPFlags(tagCmd.Flags())
}

// Normalize tag options to lowercase.
func normalize(s []string) []string {
	result := make([]string, len(s))
	for i, v := range s {
		result[i] = strings.ToLower(v)
	}
	return result
}

// Process files from given source path and stage rename transformations.
func stageTag(files []*voit.Voit, opts *Opts) {
	collisions := make(map[string]int)
	vCfg := opts.Voit()

	for i := range files {
		files[i].Ingest(&vCfg)

		// Copy the original struct and use new reference for tags.
		files[i].Mark = files[i].Orig
		files[i].Mark.Tags.Items = slices.Clone(files[i].Orig.Tags.Items)

		if len(opts.Tag.Select) == 0 {
			files[i].Matched = true // No match filter, match all files.
		} else if files[i].Orig.Tags.Match(opts.Tag.Select) {
			files[i].Matched = true
		}

		if files[i].Matched {
			if opts.Tag.Delete {
				files[i].Mark.Tags.Items = []string{}
			}

			if len(opts.Tag.Add) != 0 {
				for _, tag := range opts.Tag.Add {
					files[i].Mark.Tags.Add(tag)
				}
			}

			if len(opts.Tag.Remove) != 0 {
				for _, tag := range opts.Tag.Remove {
					files[i].Mark.Tags.Delete(tag)
				}
			}

			if len(opts.Tag.Set) != 0 {
				files[i].Mark.Tags.Items = files[i].Mark.Tags.Items[:0]
				for _, tag := range opts.Tag.Set {
					files[i].Mark.Tags.Add(tag)
				}
			}

			desc := files[i].Mark.Desc.Text
			files[i].Format(&vCfg)

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
				files[i].Format(&vCfg)
			}
		}
	}
}
