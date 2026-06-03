// Naked command options.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	. "github.com/r-pufky/voit/config"
	"github.com/r-pufky/voit/voit"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const rootLong = `
Manage filenames using Karl Voit's Managing Digital Files scheme.

  {VTIME} {DESC} -- {TAGS}.{EXT}

Read about structure and PhD thesis:
  https://karl-voit.at/folder-hierarchy
  https://karl-voit.at/tagstore/en/papers.shtml
Source: https://github.com/r-pufky/voit
Set default options in: ~/.config/voit.toml (see README.md)
`

var (
	Version = "development"
)

func init() {
	loadUserConfig()
	rootCmd.PersistentFlags().SortFlags = false

	rootCmd.PersistentFlags().StringVarP(&Cfg.AbsSource, "source", "s", "", "Directory containing files or File to rename (default: current directory)")
	rootCmd.PersistentFlags().StringVarP(&Cfg.TagSep, "tag-sep", "", voit.DefaultTagsSep, "Tag separator")
	rootCmd.PersistentFlags().StringVarP(&Cfg.DescSep, "desc-sep", "", voit.DefaultDescSep, "Description separator")
	rootCmd.PersistentFlags().StringVarP(&Cfg.SpanSep, "span-sep", "", voit.DefaultSpanSep, "VTIME date span separator")
	rootCmd.PersistentFlags().StringVarP(&Cfg.Format, "format", "", voit.DefaultVFormat, "VTIME format")
	rootCmd.PersistentFlags().BoolVarP(&Cfg.Verbose, "verbose", "v", false, "Show verbose information")
	rootCmd.PersistentFlags().BoolVarP(&Cfg.Yes, "yes", "y", false, "Automatically confirm operations")
	rootCmd.PersistentFlags().BoolVarP(&Cfg.Lower, "lower", "l", false, "Lowercase description and extension")
	rootCmd.PersistentFlags().BoolVarP(&Cfg.Overwrite, "overwrite", "", false, "Overwrite existing target files (DANGEROUS)")

	rootCmd.Flags().BoolVarP(&Cfg.Build, "build", "b", false, "Show build version.")
}

var rootCmd = &cobra.Command{
	Use:   "voit",
	Short: "Voit file naming utility.",
	Long:  rootLong,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := viper.BindPFlags(cmd.Flags()); err != nil {
			return err
		}

		if err := viper.Unmarshal(&Cfg); err != nil {
			return fmt.Errorf("unable to decode into struct: %w", err)
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		if Cfg.Verbose {
			fmt.Printf("Parsed Config: %+v [voit: %+v]\n", Cfg, Cfg.Voit())
		}
		if Cfg.Build {
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
