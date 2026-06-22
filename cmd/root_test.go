package cmd

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func TestBindFlagsToPrefix(t *testing.T) {
	viper.Reset()

	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.StringSlice("add", []string{}, "Add specified tags")
	fs.Bool("sync-xmp", false, "Sync tags from XMP")

	args := []string{"--add", "choco", "--sync-xmp"}
	if err := fs.Parse(args); err != nil {
		t.Fatalf("Failed to parse dummy arguments: %v", err)
	}

	bindFlagsToPrefix(fs, "tag")

	gotTags := viper.GetStringSlice("tag.add")
	if len(gotTags) != 1 || gotTags[0] != "choco" {
		t.Errorf("\nGot:  %v\nWant: choco\n", gotTags)
	}

	gotSync := viper.GetBool("tag.sync-xmp")
	if !gotSync {
		t.Errorf("\nGot:  %v\nWant: true\n", gotSync)
	}
}

func TestBindFlagsWithoutPrefix(t *testing.T) {
	viper.Reset()

	fs := pflag.NewFlagSet("test-root", pflag.ContinueOnError)
	fs.Bool("verbose", false, "Verbose output")

	args := []string{"--verbose"}
	if err := fs.Parse(args); err != nil {
		t.Fatalf("Failed to parse dummy arguments: %v", err)
	}

	bindFlagsToPrefix(fs, "")

	if !viper.GetBool("verbose") {
		t.Errorf("\nExpected root level 'verbose' to be true.\n")
	}
}
