package voit

import (
	"reflect"
	"testing"
)

func TestTagAdd(t *testing.T) {
	tests := []struct {
		name     string
		tags     []string
		tag      string
		wantTags []string
	}{
		{
			name:     "append a tag [tag added]",
			tags:     []string{"summer", "beach"},
			tag:      "vacation",
			wantTags: []string{"summer", "beach", "vacation"},
		},
		{
			name:     "append a mixed-case tag [tag added lowercase]",
			tags:     []string{"summer", "beach"},
			tag:      "Vacation",
			wantTags: []string{"summer", "beach", "vacation"},
		},
		{
			name:     "mixed-case tag not added [tags unchanged]",
			tags:     []string{"summer", "beach"},
			tag:      "Beach",
			wantTags: []string{"summer", "beach"},
		},
		{
			name:     "tag not added [tags unchanged]",
			tags:     []string{"summer", "beach"},
			tag:      "summer",
			wantTags: []string{"summer", "beach"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tag := Tag{
				Items: tt.tags,
			}

			tag.Add(tt.tag)

			if len(tag.Items) == 0 && len(tt.wantTags) == 0 {
				return
			}
			if !reflect.DeepEqual(tag.Items, tt.wantTags) {
				t.Errorf("\nItems\nGot:  %q\nWant: %q\n", tag.Items, tt.wantTags)
			}
		})
	}
}

func TestTagSyncAdd(t *testing.T) {
	tests := []struct {
		name     string
		cfg      Config
		tags     []string
		tag      string
		wantTags []string
	}{
		{
			name:     "append a tag [tag added]",
			cfg:      NewConfig(),
			tags:     []string{"summer", "beach"},
			tag:      "vacation",
			wantTags: []string{"summer", "beach", "vacation"},
		},
		{
			name:     "keep-folder [folder retained]",
			cfg:      NewConfig(Config{Sync: SyncConfig{KeepFolder: true}}),
			tags:     []string{"summer", "beach"},
			tag:      "places/vacation",
			wantTags: []string{"summer", "beach", "places➔vacation"},
		},
		{
			name:     "keep-space [tag space retained]",
			cfg:      NewConfig(Config{Sync: SyncConfig{KeepSpace: true}}),
			tags:     []string{"summer", "beach"},
			tag:      "person name",
			wantTags: []string{"summer", "beach", "person⠀name"},
		},
		{
			name:     "keep-all [folder, tag space retained]",
			cfg:      NewConfig(Config{Sync: SyncConfig{KeepFolder: true, KeepSpace: true}}),
			tags:     []string{"summer", "beach"},
			tag:      "places/cabo/person name",
			wantTags: []string{"summer", "beach", "places➔cabo➔person⠀name"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tagObj := Tag{
				Items: tt.tags,
			}

			tagObj.SyncAdd(tt.tag, tt.cfg)

			if len(tagObj.Items) == 0 && len(tt.wantTags) == 0 {
				return
			}
			if !reflect.DeepEqual(tagObj.Items, tt.wantTags) {
				t.Errorf("\nTags:     %q\nwantTags: %q\n", tagObj.Items, tt.wantTags)
			}
		})
	}
}

func TestTagDelete(t *testing.T) {
	tests := []struct {
		name     string
		tags     []string
		tag      string
		wantTags []string
	}{
		{
			name:     "delete a tag [tag removed]",
			tags:     []string{"summer", "beach", "vacation"},
			tag:      "vacation",
			wantTags: []string{"summer", "beach"},
		},
		{
			name:     "tag not found [tags unchanged]",
			tags:     []string{"summer", "beach"},
			tag:      "vacation",
			wantTags: []string{"summer", "beach"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tag := Tag{
				Items: tt.tags,
			}

			tag.Delete(tt.tag)

			if len(tag.Items) == 0 && len(tt.wantTags) == 0 {
				return
			}
			if !reflect.DeepEqual(tag.Items, tt.wantTags) {
				t.Errorf("\nItems\nGot:  %q\nWant: %q\n", tag.Items, tt.wantTags)
			}
		})
	}
}

func TestTagMatch(t *testing.T) {
	tests := []struct {
		name     string
		tags     []string
		match    []string
		wantBool bool
	}{
		{
			name:     "all matched",
			tags:     []string{"summer", "vacation", "beach"},
			match:    []string{"summer", "beach"},
			wantBool: true,
		},
		{
			name:     "mixed-case matched [case insensitive match]",
			tags:     []string{"summer", "vacation", "beach"},
			match:    []string{"SUmmer", "BEACH"},
			wantBool: true,
		},
		{
			name:     "no match [rust not matched]",
			tags:     []string{"summer", "vacation", "beach"},
			match:    []string{"summer", "rust"},
			wantBool: false,
		},
		{
			name:     "empty tags [no match]",
			tags:     []string{},
			match:    []string{"summer"},
			wantBool: false,
		},
		{
			name:     "empty match with tags [no match]",
			tags:     []string{"golang", "programming"},
			match:    []string{},
			wantBool: false,
		},
		{
			name:     "nil match with tags [no match]",
			tags:     []string{"golang"},
			match:    nil,
			wantBool: false,
		},
		{
			name:     "both empty [no match]",
			tags:     []string{},
			match:    []string{},
			wantBool: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tag := &Tag{Items: tt.tags}

			got := tag.Match(tt.match)

			if got != tt.wantBool {
				t.Errorf("\nGot:   %v\nWant:  %v\nMatch: %v\nItems: %v\n", got, tt.wantBool, tt.match, tt.tags)
			}
		})
	}
}

func TestTagChomp(t *testing.T) {
	tests := []struct {
		name     string
		fName    string
		args     Config
		wantTags []string
		wantName string
		wantIdx  int
	}{
		{
			name:     "sanitized format",
			fName:    "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
			args:     Config{},
			wantTags: []string{"summer", "vacation", "beach"},
			wantName: "2026-02-02T12.05.20.700 - beach vacation",
			wantIdx:  38,
		},
		{
			name:     "alternative separator",
			fName:    "2026-02-02T12.05.20.700 beach vacation - summer vacation beach",
			args:     Config{Voit: VoitConfig{TagSep: " - "}},
			wantTags: []string{"summer", "vacation", "beach"},
			wantName: "2026-02-02T12.05.20.700 beach vacation",
			wantIdx:  38,
		},
		{
			name:     "tags lowercased when added",
			fName:    "2026-02-02T12.05.20.700 beach vacation -- SUMMER VACATION BEACH",
			args:     Config{},
			wantName: "2026-02-02T12.05.20.700 beach vacation",
			wantTags: []string{"summer", "vacation", "beach"},
			wantIdx:  38,
		},
		{
			name:     "digikam tag spacers are considered a unbroken string",
			fName:    "2026-02-02T12.05.20.700 beach vacation -- nested⠀tag⠀summer vacation beach",
			args:     Config{},
			wantName: "2026-02-02T12.05.20.700 beach vacation",
			wantTags: []string{"nested⠀tag⠀summer", "vacation", "beach"},
			wantIdx:  38,
		},
		{
			name:     "empty [trailing space]",
			fName:    "2026-02-02T12.05.20.700 beach vacation -- ",
			args:     Config{},
			wantName: "2026-02-02T12.05.20.700 beach vacation",
			wantTags: []string{},
			wantIdx:  38,
		},
		{
			name:     "empty [invalid separator]",
			fName:    "2026-02-02T12.05.20.700 beach vacation --",
			args:     Config{},
			wantName: "2026-02-02T12.05.20.700 beach vacation --",
			wantTags: []string{},
			wantIdx:  -1,
		},
		{
			name:     "no tags",
			fName:    "2026-02-02T12.05.20.700 beach vacation",
			args:     Config{},
			wantName: "2026-02-02T12.05.20.700 beach vacation",
			wantTags: []string{},
			wantIdx:  -1,
		},
		{
			name:     "tags de-duplicated",
			fName:    "2026-02-02T12.05.20.700 beach vacation -- summer summer beach",
			args:     Config{},
			wantName: "2026-02-02T12.05.20.700 beach vacation",
			wantTags: []string{"summer", "beach"},
			wantIdx:  38,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tag Tag

			tIdx := tag.Chomp(tt.fName, tt.args)

			if tIdx != tt.wantIdx {
				t.Errorf("\nIdx\nGot:  %d\nWant: %d\n", tIdx, tt.wantIdx)
			}

			if len(tag.Items) == 0 && len(tt.wantTags) == 0 {
				return
			}
			if !reflect.DeepEqual(tag.Items, tt.wantTags) {
				t.Errorf("\nItems\nGot:  %q\nWant: %q\n", tag.Items, tt.wantTags)
			}
		})
	}
}

func TestTagFormat(t *testing.T) {
	tests := []struct {
		name       string
		tags       []string
		args       Config
		wantFormat string
	}{
		{
			name:       "valid tags format [correct format]",
			tags:       []string{"summer", "beach"},
			args:       Config{},
			wantFormat: " -- summer beach",
		},
		{
			name:       "invalid tags [correct format]",
			tags:       []string{"SUMMER", "Beach"},
			args:       Config{},
			wantFormat: " -- summer beach",
		},
		{
			name:       "alternative output separator [correct format]",
			tags:       []string{"SUMMER", "Beach"},
			args:       Config{Voit: VoitConfig{TagSep: "|"}},
			wantFormat: "|summer beach",
		},
		{
			name:       "empty tags [no string returned]",
			tags:       []string{},
			args:       Config{},
			wantFormat: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tag := Tag{
				Items: tt.tags,
			}

			f := tag.Format(tt.args)

			if f != tt.wantFormat {
				t.Errorf("\nFormat()\nGot:  %q\nWant: %q\n", f, tt.wantFormat)
			}
		})
	}
}
