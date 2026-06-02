package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/r-pufky/voit/internal"
	"github.com/r-pufky/voit/models"
	"github.com/r-pufky/voit/voit"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	tagCmd = &cobra.Command{
		Use:   "tag",
		Short: "tag files according to filters",
		Long: "Tag files according to filters.\n\n" +
			"  {VTIME} {DESC} -- {TAGS}.{EXT}\n\n" +
			"Target files are automatically differentiated if there are name collisions.\n\n",
		Example: "  voit tag -s ./photos -e party\n  voit tag -s ./photos -l cake -a candles -a bday",

		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			loadUserConfig()
			// Ensure tags are lowercased for file or CLI source.
			if cmd.Flags().Changed("add") || viper.IsSet("add") {
				opts.Tag.Add = normalize(viper.GetStringSlice("add"))
				viper.Set("add", opts.Tag.Add)
			}

			if cmd.Flags().Changed("remove") || viper.IsSet("remove") {
				opts.Tag.Remove = normalize(viper.GetStringSlice("remove"))
				viper.Set("remove", opts.Tag.Remove)
			}

			if cmd.Flags().Changed("set") || viper.IsSet("set") {
				opts.Tag.Set = normalize(viper.GetStringSlice("set"))
				viper.Set("set", opts.Tag.Set)
			}

			if cmd.Flags().Changed("select") || viper.IsSet("select") {
				opts.Tag.Select = normalize(viper.GetStringSlice("select"))
				viper.Set("select", opts.Tag.Select)
			}

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
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

			config := &voit.Config{
				Format:  voit.DefaultVFormat,
				Pattern: opts.Rename.Pattern,
				SSep:    opts.SpanSep,
				DSep:    opts.DescSep,
				TSep:    opts.TagSep,
				Lower:   opts.Rename.Lower,
			}

			files, err := internal.Scan(opts.AbsSource)
			if err != nil {
				log.Fatalf("Unable to complete source file scan: %v", err)
			}

			if len(files) == 0 {
				fmt.Println("No files matched the known datetime formats.")
				os.Exit(0)
			}

			stageTag(files, &opts, config)

			count := internal.DisplayPending(os.Stdout, files, config)
			if count != 0 {
				if opts.Tag.Overwrite {
					fmt.Printf("\nProposed changes (OVERWRITE ENABLED): %d file(s).\n", count)
				} else {
					fmt.Printf("\nProposed changes: %d file(s).\n", count)
				}

				if !opts.Yes && !internal.Confirm(os.Stdin, os.Stdout) {
					fmt.Println("Operation aborted by user.")
					os.Exit(0)
				}

				internal.Rename(os.Stdout, files, opts.Tag.Overwrite, opts.Verbose)
			} else {
				fmt.Println("No files matched proposed changes.")
			}
		},
	}
)

func init() {
	tagCmd.Flags().SortFlags = false

	tagCmd.Flags().StringSliceVarP(&opts.Tag.Add, "add", "a", []string{}, "Add specified tags")
	tagCmd.Flags().StringSliceVarP(&opts.Tag.Remove, "remove", "r", []string{}, "Remove specified tags")
	tagCmd.Flags().StringSliceVarP(&opts.Tag.Set, "set", "e", []string{}, "Set tags to specified tags")
	tagCmd.Flags().StringSliceVarP(&opts.Tag.Select, "select", "l", []string{}, "Perform operations only on files with matching tags (default: all)")
	tagCmd.Flags().BoolVarP(&opts.Tag.Delete, "delete", "d", false, "Remove all tags")
	tagCmd.Flags().BoolVarP(&opts.Tag.Overwrite, "overwrite", "", false, "Overwrite existing target files (DANGEROUS)")
	tagCmd.MarkFlagsMutuallyExclusive("add", "remove", "set", "delete")
	viper.BindPFlags(tagCmd.Flags())
}

// normalize tag options to lowercase.
func normalize(s []string) []string {
	result := make([]string, len(s))
	for i, v := range s {
		result[i] = strings.ToLower(v)
	}
	return result
}

// Process files from given source path and stage rename transformations.
func stageTag(files []*voit.Voit, opts *models.Opts, config *voit.Config) {
	collisions := make(map[string]int)

	for i := range files {
		files[i].Ingest(config)

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
}
