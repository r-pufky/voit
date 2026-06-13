package voit

import (
	"reflect"
	"testing"
	"time"
)

func TestTagSyncAdd(t *testing.T) {
	tests := []struct {
		name     string
		opts     *Opts
		tags     []string
		tag      string
		wantTags []string
	}{
		{
			name: "sanity: append a tag [tag added]",
			opts: &Opts{
				Tag: TagOpts{
					SyncInFolder:   DefaultSyncInFolder,
					SyncOutFolder:  DefaultSyncOutFolder,
					SyncOutSpace:   DefaultSyncOutSpace,
					SyncKeepFolder: false,
					SyncKeepSpace:  false,
				},
			},
			tags:     []string{"summer", "beach"},
			tag:      "vacation",
			wantTags: []string{"summer", "beach", "vacation"},
		},
		{
			name: "sanity: keep-folder [folder retained]",
			opts: &Opts{
				Tag: TagOpts{
					SyncInFolder:   DefaultSyncInFolder,
					SyncOutFolder:  DefaultSyncOutFolder,
					SyncOutSpace:   DefaultSyncOutSpace,
					SyncKeepFolder: true,
					SyncKeepSpace:  false,
				},
			},
			tags:     []string{"summer", "beach"},
			tag:      "places/vacation",
			wantTags: []string{"summer", "beach", "places➔vacation"},
		},
		{
			name: "sanity: keep-space [tag space retained]",
			opts: &Opts{
				Tag: TagOpts{
					SyncInFolder:   DefaultSyncInFolder,
					SyncOutFolder:  DefaultSyncOutFolder,
					SyncOutSpace:   DefaultSyncOutSpace,
					SyncKeepFolder: false,
					SyncKeepSpace:  true,
				},
			},
			tags: []string{"summer", "beach"},
			tag:  "person name",
			// Non-printable brail space counts as non-whitespace.
			wantTags: []string{"summer", "beach", "person⠀name"},
		},
		{
			name: "sanity: keep-all [folder, tag space retained]",
			opts: &Opts{
				Tag: TagOpts{
					SyncInFolder:   DefaultSyncInFolder,
					SyncOutFolder:  DefaultSyncOutFolder,
					SyncOutSpace:   DefaultSyncOutSpace,
					SyncKeepFolder: true,
					SyncKeepSpace:  true,
				},
			},
			tags: []string{"summer", "beach"},
			tag:  "places/cabo/person name",
			// Non-printable brail space counts as non-whitespace.
			wantTags: []string{"summer", "beach", "places➔cabo➔person⠀name"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tag := Tag{
				Items: tt.tags,
			}

			tag.SyncAdd(tt.tag, tt.opts)

			if len(tag.Items) == 0 && len(tt.wantTags) == 0 {
				return
			}
			if !reflect.DeepEqual(tag.Items, tt.wantTags) {
				t.Errorf("\nTags:     %q\nwantTags: %q", tag.Items, tt.wantTags)
			}

		})
	}
}

func TestTagAdd(t *testing.T) {
	tests := []struct {
		name     string
		tags     []string
		tag      string
		wantTags []string
	}{
		{
			name:     "sanity: append a tag [tag added]",
			tags:     []string{"summer", "beach"},
			tag:      "vacation",
			wantTags: []string{"summer", "beach", "vacation"},
		},
		{
			name:     "sanity: append a mixed-case tag [tag added lowercase]",
			tags:     []string{"summer", "beach"},
			tag:      "Vacation",
			wantTags: []string{"summer", "beach", "vacation"},
		},
		{
			name:     "sanity: mixed-case tag not added [tags unchanged]",
			tags:     []string{"summer", "beach"},
			tag:      "Beach",
			wantTags: []string{"summer", "beach"},
		},
		{
			name:     "sanity: tag not added [tags unchanged]",
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
				t.Errorf("\nTags:     %q\nwantTags: %q", tag.Items, tt.wantTags)
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
			name:     "sanity: delete a tag [tag removed]",
			tags:     []string{"summer", "beach", "vacation"},
			tag:      "vacation",
			wantTags: []string{"summer", "beach"},
		},
		{
			name:     "sanity: tag not found [tags unchanged]",
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
				t.Errorf("\nTags:     %q\nwantTags: %q", tag.Items, tt.wantTags)
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
			name:     "sanity: all matched",
			tags:     []string{"summer", "vacation", "beach"},
			match:    []string{"summer", "beach"},
			wantBool: true,
		},
		{
			name:     "sanity: mixed-case matched [case insensitive match]",
			tags:     []string{"summer", "vacation", "beach"},
			match:    []string{"SUmmer", "BEACH"},
			wantBool: true,
		},
		{
			name:     "sanity: no match [rust not matched]",
			tags:     []string{"summer", "vacation", "beach"},
			match:    []string{"summer", "rust"},
			wantBool: false,
		},
		{
			name:     "sanity: empty tags [no match]",
			tags:     []string{},
			match:    []string{"summer"},
			wantBool: false,
		},
		{
			name:     "sanity: empty match with tags [no match]",
			tags:     []string{"golang", "programming"},
			match:    []string{},
			wantBool: false,
		},
		{
			name:     "sanity: Nil match with tags [no match]",
			tags:     []string{"golang"},
			match:    nil,
			wantBool: false,
		},
		{
			name:     "sanity: both empty [no match]",
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
				t.Errorf("Tag.Match() = %v, want %v for match %v against items %v",
					got, tt.wantBool, tt.match, tt.tags)
			}
		})
	}
}

func TestStageTag(t *testing.T) {
	vTime := time.Date(2026, time.February, 2, 12, 5, 20, 700000000, time.UTC)
	bDesc := "beach vacation"

	tests := []struct {
		name     string
		opts     *Opts
		f        VoitFiles
		wantVoit []Voit
	}{
		{
			name: "sanity: add tag [tag added]",
			opts: &Opts{
				Tag: TagOpts{
					Add: []string{"additional"},
				},
			},
			f: VoitFiles{
				{
					File: File{
						Source: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
						Name:   "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
						Ext:    ".jpg",
					},
				},
			},
			wantVoit: []Voit{
				{
					File: File{
						Source: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
						Name:   "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
						Ext:    ".jpg",
					},
					Orig: Meta{
						VTime: VTime{Time: vTime},
						Tags:  Tag{Items: []string{"summer", "vacation", "beach"}},
						Desc:  Desc{Text: bDesc},
					},
					Mark: Meta{
						VTime: VTime{Time: vTime},
						Tags:  Tag{Items: []string{"summer", "vacation", "beach", "additional"}},
						Desc:  Desc{Text: bDesc},
					},
					Matched: true,
					Target:  "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach additional.jpg",
				},
			},
		},
		{
			name: "sanity: remove tag [tag removed]",
			opts: &Opts{
				Tag: TagOpts{
					Remove: []string{"summer"},
				},
			},
			f: VoitFiles{
				{
					File: File{
						Source: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
						Name:   "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
						Ext:    ".jpg",
					},
				},
			},
			wantVoit: []Voit{
				{
					File: File{
						Source: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
						Name:   "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
						Ext:    ".jpg",
					},
					Orig: Meta{
						VTime: VTime{Time: vTime},
						Tags:  Tag{Items: []string{"summer", "vacation", "beach"}},
						Desc:  Desc{Text: bDesc},
					},
					Mark: Meta{
						VTime: VTime{Time: vTime},
						Tags:  Tag{Items: []string{"vacation", "beach"}},
						Desc:  Desc{Text: bDesc},
					},
					Matched: true,
					Target:  "/tmp/2026-02-02T12.05.20.700 beach vacation -- vacation beach.jpg",
				},
			},
		},
		{
			name: "sanity: set tags [tag overwritten]",
			opts: &Opts{
				Tag: TagOpts{
					Set: []string{"family", "europe"},
				},
			},
			f: VoitFiles{
				{
					File: File{
						Source: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
						Name:   "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
						Ext:    ".jpg",
					},
				},
			},
			wantVoit: []Voit{
				{
					File: File{
						Source: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
						Name:   "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
						Ext:    ".jpg",
					},
					Orig: Meta{
						VTime: VTime{Time: vTime},
						Tags:  Tag{Items: []string{"summer", "vacation", "beach"}},
						Desc:  Desc{Text: bDesc},
					},
					Mark: Meta{
						VTime: VTime{Time: vTime},
						Tags:  Tag{Items: []string{"family", "europe"}},
						Desc:  Desc{Text: bDesc},
					},
					Matched: true,
					Target:  "/tmp/2026-02-02T12.05.20.700 beach vacation -- family europe.jpg",
				},
			},
		},
		{
			name: "sanity: select add tag [subset of tags have tags added]",
			opts: &Opts{
				Tag: TagOpts{
					Select: []string{"summer", "vacation", "park"},
					Add:    []string{"europe"},
				},
			},
			f: VoitFiles{
				{
					File: File{
						Source: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
						Name:   "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
						Ext:    ".jpg",
					},
				},
				{
					File: File{
						Source: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation park.jpg",
						Name:   "2026-02-02T12.05.20.700 beach vacation -- summer vacation park",
						Ext:    ".jpg",
					},
				},
			},
			wantVoit: []Voit{
				{
					File: File{
						Source: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
						Name:   "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
						Ext:    ".jpg",
					},
					Orig: Meta{
						VTime: VTime{Time: vTime},
						Tags:  Tag{Items: []string{"summer", "vacation", "beach"}},
						Desc:  Desc{Text: bDesc},
					},
					Mark: Meta{
						VTime: VTime{Time: vTime},
						Tags:  Tag{Items: []string{"summer", "vacation", "beach"}},
						Desc:  Desc{Text: bDesc},
					},
					Matched: false,
					Target:  "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
				},
				{
					File: File{
						Source: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation park.jpg",
						Name:   "2026-02-02T12.05.20.700 beach vacation -- summer vacation park",
						Ext:    ".jpg",
					},
					Orig: Meta{
						VTime: VTime{Time: vTime},
						Tags:  Tag{Items: []string{"summer", "vacation", "park"}},
						Desc:  Desc{Text: bDesc},
					},
					Mark: Meta{
						VTime: VTime{Time: vTime},
						Tags:  Tag{Items: []string{"summer", "vacation", "park", "europe"}},
						Desc:  Desc{Text: bDesc},
					},
					Matched: true,
					Target:  "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation park europe.jpg",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.f.StageTag(tt.opts)

			// DeepEqual will compare memory addresses if pointers, not values.
			// Convert to value. This is required as pointers are needed to update
			// the struct in place during stageTag.
			got := make([]Voit, len(tt.f))
			for i, ptr := range tt.f {
				if ptr != nil {
					got[i] = *ptr
				}
			}

			if !reflect.DeepEqual(got, tt.wantVoit) {
				t.Errorf("\nGot Voit:  %+v\nWant Voit: %+v", got, tt.wantVoit)
			}
		})
	}
}
