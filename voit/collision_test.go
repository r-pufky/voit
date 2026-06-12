package voit

import (
	"testing"
	"time"
)

func TestResolveCollisions(t *testing.T) {
	vCfg := NewConfig()
	now := time.Time{}

	tests := []struct {
		name      string
		files     VoitFiles
		wantDesc1 string
		wantDesc2 string
	}{
		{
			name: "sanity: no collisions [no modifications]",
			files: VoitFiles{
				{
					Target:  "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
					Matched: true,
					Mark: Meta{
						VTime: VTime{Time: now},
						Desc:  Desc{Text: "beach vacation"},
					},
				},
				{
					Target:  "/tmp/2026-02-02T12.05.20.700 forest vacation -- summer vacation beach.jpg",
					Matched: true,
					Mark: Meta{
						VTime: VTime{Time: now},
						Desc:  Desc{Text: "forest vacation"},
					},
				},
			},
			wantDesc1: "forest vacation",
		},
		{
			name: "sanity: standard collision [count added to desc]",
			files: VoitFiles{
				{
					Target:  "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
					Matched: true,
					Mark: Meta{
						VTime: VTime{Time: now},
						Desc:  Desc{Text: "beach vacation"},
					},
				},
				{
					Target:  "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg",
					Matched: true,
					Mark: Meta{
						VTime: VTime{Time: now},
						Desc:  Desc{Text: "beach vacation"},
					},
				},
			},
			wantDesc1: "beach vacation_1",
		},
		{
			name: "sanity: multi-collision [multiple collisions numerically incremented]",
			files: VoitFiles{
				{
					Target:  "/tmp/2026-02-02T12.05.20.700 beach vacation-- summer vacation beach.jpg",
					Matched: true,
					Mark: Meta{
						VTime: VTime{Time: now},
						Desc:  Desc{Text: "beach vacation"},
					},
				},
				{
					Target:  "/tmp/2026-02-02T12.05.20.700 beach vacation-- summer vacation beach.jpg",
					Matched: true,
					Mark: Meta{
						VTime: VTime{Time: now},
						Desc:  Desc{Text: "beach vacation_1"},
					},
				},
				{
					Target:  "/tmp/2026-02-02T12.05.20.700 beach vacation-- summer vacation beach.jpg",
					Matched: true,
					Mark: Meta{
						VTime: VTime{Time: now},
						Desc:  Desc{Text: "beach vacation_2"},
					},
				},
			},
			wantDesc1: "beach vacation_1",
			wantDesc2: "beach vacation_2",
		},
		{
			name: "no desc: no tags [count added to desc for 1,2]",
			files: VoitFiles{
				{
					Target:  "/tmp/2026-02-02T12.05.20.700.jpg",
					Matched: true,
					Mark: Meta{
						VTime: VTime{Time: now},
						Desc:  Desc{Text: ""},
					},
				},
				{
					Target:  "/tmp/2026-02-02T12.05.20.700.jpg",
					Matched: true,
					Mark: Meta{
						VTime: VTime{Time: now},
						Desc:  Desc{Text: "1"},
					},
				},
				{
					Target:  "/tmp/2026-02-02T12.05.20.700.jpg",
					Matched: true,
					Mark: Meta{
						VTime: VTime{Time: now},
						Desc:  Desc{Text: "2"},
					},
				},
			},
			wantDesc1: "1",
			wantDesc2: "2",
		},
		{
			name: "no desc: tags [count added to desc for 1,2]",
			files: VoitFiles{
				{
					Target:  "/tmp/2026-02-02T12.05.20.700 -- summer vacation beach.jpg",
					Matched: true,
					Mark: Meta{
						VTime: VTime{Time: now},
						Desc:  Desc{Text: ""},
					},
				},
				{
					Target:  "/tmp/2026-02-02T12.05.20.700 1 -- summer vacation beach.jpg",
					Matched: true,
					Mark: Meta{
						VTime: VTime{Time: now},
						Desc:  Desc{Text: "1"},
					},
				},
				{
					Target:  "/tmp/2026-02-02T12.05.20.700 2 -- summer vacation beach.jpg",
					Matched: true,
					Mark: Meta{
						VTime: VTime{Time: now},
						Desc:  Desc{Text: "2"},
					},
				},
			},
			wantDesc1: "1",
			wantDesc2: "2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.files.ResolveCollisions(vCfg, false)

			if tt.name == "Path 1: No Collision" {
				if tt.files[0].Mark.Desc.Text != "Unique File" {
					t.Errorf("wantBool first file description to remain 'Unique File', got %q", tt.files[0].Mark.Desc.Text)
				}
			}

			// Assert for the second file
			if len(tt.files) > 1 {
				gotDesc := tt.files[1].Mark.Desc.Text
				if gotDesc != tt.wantDesc1 {
					t.Errorf("wantBool index 1 description to be %q, got %q", tt.wantDesc1, gotDesc)
				}
			}

			// Assert for the third file (specifically for Path 5)
			if tt.wantDesc2 != "" && len(tt.files) > 2 {
				gotDesc2 := tt.files[2].Mark.Desc.Text
				if gotDesc2 != tt.wantDesc2 {
					t.Errorf("wantBool index 2 description to be %q, got %q", tt.wantDesc2, gotDesc2)
				}
			}
		})
	}
}
