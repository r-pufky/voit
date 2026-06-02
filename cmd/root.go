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

var (
	Version = "development"
)

func init() {
	rootCmd.PersistentFlags().SortFlags = false

	rootCmd.PersistentFlags().StringVarP(&Cfg.AbsSource, "source", "s", "", "Directory containing files or File to rename (default: current directory)")
	rootCmd.PersistentFlags().StringVarP(&Cfg.TagSep, "tag-sep", "", voit.DefaultTagsSep, "Tag separator")
	rootCmd.PersistentFlags().StringVarP(&Cfg.DescSep, "desc-sep", "", voit.DefaultDescSep, "Description separator")
	rootCmd.PersistentFlags().StringVarP(&Cfg.SpanSep, "span-sep", "", voit.DefaultSpanSep, "VTIME date span separator")
	rootCmd.PersistentFlags().BoolVarP(&Cfg.Verbose, "verbose", "v", false, "Show verbose information")
	rootCmd.PersistentFlags().BoolVarP(&Cfg.Yes, "yes", "y", false, "Automatically confirm operations")

	rootCmd.Flags().BoolVarP(&Cfg.Build, "build", "b", false, "Show build version.")
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
			fmt.Printf("Parsed Config: %+v\n", Cfg)
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

func loadUserConfig() {
	home, _ := os.UserHomeDir()
	viper.SetConfigFile(filepath.Join(home, ".config", "voit.toml"))

	if err := viper.ReadInConfig(); err == nil {
		if Cfg.Verbose {
			fmt.Printf("Using config file: %s\n", viper.ConfigFileUsed())
		}
	}

	// Unmarshal config into opts struct.
	if err := viper.Unmarshal(&Cfg); err != nil {
		fmt.Printf("Invalid config: %v\n", err)
	}
}
