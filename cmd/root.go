// Naked command options.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/k0kubun/pp/v3"
	"github.com/r-pufky/voit/voit"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const rootLong = `
Manage filenames using Karl Voit's Managing Digital Files scheme.

  {VTIME} {DESC} -- {TAGS}.{EXT}

Read about structure and PhD thesis:

  https://karl-voit.at/folder-hierarchy
  https://karl-voit.at/tagstore/en/papers.shtml

Source: https://github.com/r-pufky/voit

Default options: ~/.config/voit.toml (see README.md)

Use quotes when globbing source files. Certain shells will expand globs
before passing arguments to binaries resulting in unexpected behavior.
Wrapping the path in quotes prevents the user shell from expanding these.
`

var (
	Version = "development"
)

func init() {
	rootCmd.PersistentFlags().SortFlags = false

	rootCmd.PersistentFlags().StringVarP(&voit.Cfg.AbsSource, "source", "s", "", "Source path for renaming files [globbing supported, use quotes to prevent shell expansion] (default: current directory)")
	rootCmd.PersistentFlags().StringVarP(&voit.Cfg.TagSep, "tag-sep", "", voit.DefaultTagsSep, "Tag separator")
	rootCmd.PersistentFlags().StringVarP(&voit.Cfg.DescSep, "desc-sep", "", voit.DefaultDescSep, "Description separator")
	rootCmd.PersistentFlags().StringVarP(&voit.Cfg.SpanSep, "span-sep", "", voit.DefaultSpanSep, "VTIME date span separator")
	rootCmd.PersistentFlags().StringVarP(&voit.Cfg.Format, "format", "", voit.DefaultVFormat, "VTIME format")
	rootCmd.PersistentFlags().BoolVarP(&voit.Cfg.Verbose, "verbose", "v", false, "Show verbose information")
	rootCmd.PersistentFlags().BoolVarP(&voit.Cfg.Yes, "yes", "y", false, "Automatically confirm operations")
	rootCmd.PersistentFlags().BoolVarP(&voit.Cfg.Lower, "lower", "l", false, "Lowercase description and extension")
	rootCmd.PersistentFlags().BoolVarP(&voit.Cfg.Overwrite, "overwrite", "", false, "Overwrite existing target files (DANGEROUS)")

	rootCmd.Flags().BoolVarP(&voit.Cfg.Build, "build", "b", false, "Show build version.")

	bindFlagsToPrefix(tagCmd.Flags(), "")
}

var rootCmd = &cobra.Command{
	Use:   "voit",
	Short: "Voit file naming utility.",
	Long:  rootLong,
	// Run before all subcommands unless PersistentPreRunE is re-defined.
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		loadUserConfig()

		if err := viper.BindPFlags(cmd.PersistentFlags()); err != nil {
			return err
		}

		if err := viper.BindPFlags(cmd.Flags()); err != nil {
			return err
		}

		if err := viper.Unmarshal(&voit.Cfg); err != nil {
			return fmt.Errorf("unable to decode into struct: %w", err)
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		if voit.Cfg.Verbose {
			pp.Printf("Parsed Config: %v\nVoit Config: %v\n", voit.Cfg, voit.Cfg.Voit())
		}
		if voit.Cfg.Build {
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

// Load user config file via viper but do not unmarshal until flags are loaded.
func loadUserConfig() {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	viper.SetConfigFile(filepath.Join(home, ".config", "voit.toml"))

	_ = viper.ReadInConfig()
}

// Bind local config flags to correct prefix for subcommands. Enables correct
// mapping for flags which may be duplicated in subcommands resulting in flag
// being stored in incorrect location.
func bindFlagsToPrefix(flags *pflag.FlagSet, prefix string) {
	flags.VisitAll(func(f *pflag.Flag) {
		configKey := f.Name
		if prefix != "" {
			configKey = prefix + "." + f.Name
		}

		viper.BindPFlag(configKey, f)
	})
}
