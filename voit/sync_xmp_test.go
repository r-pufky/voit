package voit

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

// Simple logic requires no testing.
func TestSyncXMPNOOP(t *testing.T) {}

func TestBuildFileMap(t *testing.T) {
	tests := []struct {
		name      string
		opts      *Opts
		f         VoitFiles
		wantMapFn func(f VoitFiles) map[string]*LinkedFiles
	}{
		{
			name: "sanity: sidecar matched [file sidecar populated]",
			opts: &Opts{
				Tag: TagOpts{
					SyncXMP: true,
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
						Source: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg.xmp",
						Name:   "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
						Ext:    ".jpg.xmp",
					},
				},
			},
			wantMapFn: func(f VoitFiles) map[string]*LinkedFiles {
				return map[string]*LinkedFiles{
					"2026-02-02T12.05.20.700 beach vacation -- summer vacation beach": {
						File:    f[0],
						Sidecar: f[1],
					},
				}
			},
		},
		{
			name: "sanity: no sidecar [file populated]",
			opts: &Opts{
				Tag: TagOpts{
					SyncXMP: true,
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
			wantMapFn: func(f VoitFiles) map[string]*LinkedFiles {
				return map[string]*LinkedFiles{
					"2026-02-02T12.05.20.700 beach vacation -- summer vacation beach": {
						File: f[0],
					},
				}
			},
		},
		{
			name: "sanity: no file [sidecar populated]",
			opts: &Opts{
				Tag: TagOpts{
					SyncXMP: true,
				},
			},
			f: VoitFiles{
				{
					File: File{
						Source: "/tmp/2026-02-02T12.05.20.700 beach vacation -- summer vacation beach.jpg.xmp",
						Name:   "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
						Ext:    ".jpg.xmp",
					},
				},
			},
			wantMapFn: func(f VoitFiles) map[string]*LinkedFiles {
				return map[string]*LinkedFiles{
					"2026-02-02T12.05.20.700 beach vacation -- summer vacation beach": {
						Sidecar: f[0],
					},
				}
			},
		},
		{
			name: "sanity: no files [no matches]",
			opts: &Opts{
				Tag: TagOpts{
					SyncXMP: true,
				},
			},
			f: VoitFiles{},
			wantMapFn: func(f VoitFiles) map[string]*LinkedFiles {
				return map[string]*LinkedFiles{}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileMap := make(map[string]*LinkedFiles)
			wantMap := tt.wantMapFn(tt.f)

			tt.f.buildFileMap(tt.opts.Voit(), fileMap)

			if !reflect.DeepEqual(fileMap, wantMap) {
				t.Errorf("\nGot FileMap:  %+v\nWant FileMap: %+v", fileMap, wantMap)
			}
		})
	}
}

func TestParseSidecar(t *testing.T) {
	vTime := time.Date(2026, time.February, 2, 12, 5, 20, 700000000, time.UTC)
	bDesc := "beach vacation"
	validDigikamXMP := `<?xml version="1.0" encoding="UTF-8"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about=""
    xmlns:digiKam="http://www.digikam.org/ns/1.0/">
   <digiKam:TagsList>
    <rdf:Seq>
     <rdf:li>summer</rdf:li>
     <rdf:li>vacation</rdf:li>
     <rdf:li>beach</rdf:li>
     <rdf:li>place/location/sidecar name</rdf:li>
    </rdf:Seq>
   </digiKam:TagsList>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`
	emptyDigikamXMP := `<?xml version="1.0" encoding="UTF-8"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about=""
    xmlns:digiKam="http://www.digikam.org/ns/1.0/">
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`

	tests := []struct {
		name       string
		opts       *Opts
		filesSetup []struct {
			name    string
			ext     string
			XMPData string
		}
		wantVoitFunc func(tmpDir string) []Voit
		wantMapFn    func(tmpDir string, inputs VoitFiles) FileMap
	}{
		{
			name: "sanity: tagged sidecar match [file/sidecar updated with xmp tags]",
			opts: &Opts{
				Tag: TagOpts{
					SyncXMP:        true,
					SyncInFolder:   "/",
					SyncOutFolder:  "➔",
					SyncOutSpace:   "⠀",
					SyncKeepFolder: true,
					SyncKeepSpace:  true,
				},
			},
			filesSetup: []struct {
				name    string
				ext     string
				XMPData string
			}{
				{name: "2026-02-02T12.05.20.700 beach vacation", ext: ".jpg", XMPData: ""},
				{name: "2026-02-02T12.05.20.700 beach vacation", ext: ".jpg.xmp", XMPData: validDigikamXMP},
			},
			wantVoitFunc: func(tmpDir string) []Voit {
				return []Voit{
					{
						File: File{
							Source: filepath.Join(tmpDir, "2026-02-02T12.05.20.700 beach vacation.jpg"),
							Name:   "2026-02-02T12.05.20.700 beach vacation",
							Ext:    ".jpg",
						},
						Orig: Meta{
							VTime: VTime{Time: vTime},
							Desc:  Desc{Text: bDesc},
						},
						Mark: Meta{
							VTime: VTime{Time: vTime},
							Tags:  Tag{Items: []string{"summer", "vacation", "beach", "place➔location➔sidecar⠀name"}},
							Desc:  Desc{Text: bDesc},
						},
						Matched: true,
					},
					{
						File: File{
							Source: filepath.Join(tmpDir, "2026-02-02T12.05.20.700 beach vacation.jpg.xmp"),
							Name:   "2026-02-02T12.05.20.700 beach vacation",
							Ext:    ".jpg.xmp",
						},
						Orig: Meta{
							VTime: VTime{Time: vTime},
							Desc:  Desc{Text: bDesc},
						},
						Mark: Meta{
							VTime: VTime{Time: vTime},
							Tags:  Tag{Items: []string{"summer", "vacation", "beach", "place➔location➔sidecar⠀name"}},
							Desc:  Desc{Text: bDesc},
						},
						Matched: true,
					},
				}
			},
			wantMapFn: func(tmpDir string, inputs VoitFiles) FileMap {
				return FileMap{
					"2026-02-02T12.05.20.700 beach vacation": {
						File:    inputs[0],
						Sidecar: inputs[1],
					},
				}
			},
		},
		{
			name: "sanity: untagged sidecar no match [file/sidecar no match]",
			opts: &Opts{
				Tag: TagOpts{
					SyncXMP: true,
				},
			},
			filesSetup: []struct {
				name    string
				ext     string
				XMPData string
			}{
				{name: "2026-02-02T12.05.20.700 beach vacation", ext: ".jpg", XMPData: ""},
				{name: "2026-02-02T12.05.20.700 beach vacation", ext: ".jpg.xmp", XMPData: emptyDigikamXMP},
			},
			wantVoitFunc: func(tmpDir string) []Voit {
				return []Voit{
					{
						File: File{
							Source: filepath.Join(tmpDir, "2026-02-02T12.05.20.700 beach vacation.jpg"),
							Name:   "2026-02-02T12.05.20.700 beach vacation",
							Ext:    ".jpg",
						},
						Orig: Meta{
							VTime: VTime{Time: vTime},
							Desc:  Desc{Text: bDesc},
						},
						Mark: Meta{
							VTime: VTime{Time: vTime},
							Desc:  Desc{Text: bDesc},
						},
						Matched: false,
					},
					{
						File: File{
							Source: filepath.Join(tmpDir, "2026-02-02T12.05.20.700 beach vacation.jpg.xmp"),
							Name:   "2026-02-02T12.05.20.700 beach vacation",
							Ext:    ".jpg.xmp",
						},
						Orig: Meta{
							VTime: VTime{Time: vTime},
							Desc:  Desc{Text: bDesc},
						},
						Mark: Meta{
							VTime: VTime{Time: vTime},
							Desc:  Desc{Text: bDesc},
						},
						Matched: false,
					},
				}
			},
			wantMapFn: func(tmpDir string, inputs VoitFiles) FileMap {
				return FileMap{
					"2026-02-02T12.05.20.700 beach vacation": {
						File:    inputs[0],
						Sidecar: inputs[1],
					},
				}
			},
		},
		{
			name: "sanity: partial match [file/sidecar no match]",
			opts: &Opts{
				Tag: TagOpts{
					SyncXMP: true,
				},
			},
			filesSetup: []struct {
				name    string
				ext     string
				XMPData string
			}{
				{name: "2026-02-02T12.05.20.700 beach vacation", ext: ".jpg", XMPData: ""},
			},
			wantVoitFunc: func(tmpDir string) []Voit {
				return []Voit{
					{
						File: File{
							Source: filepath.Join(tmpDir, "2026-02-02T12.05.20.700 beach vacation.jpg"),
							Name:   "2026-02-02T12.05.20.700 beach vacation",
							Ext:    ".jpg",
						},
						Orig: Meta{
							VTime: VTime{Time: vTime},
							Desc:  Desc{Text: bDesc},
						},
						Mark: Meta{
							VTime: VTime{Time: vTime},
							Desc:  Desc{Text: bDesc},
						},
						Matched: false,
					},
				}
			},
			wantMapFn: func(tmpDir string, inputs VoitFiles) FileMap {
				return FileMap{
					"2026-02-02T12.05.20.700 beach vacation": {
						File:    inputs[0],
						Sidecar: nil,
					},
				}
			},
		},
		{
			name: "sanity: partial match [file/sidecar no match]",
			opts: &Opts{
				Tag: TagOpts{
					SyncXMP: true,
				},
			},
			filesSetup: []struct {
				name    string
				ext     string
				XMPData string
			}{
				{name: "2026-02-02T12.05.20.700 beach vacation", ext: ".jpg.xmp", XMPData: validDigikamXMP},
			},
			wantVoitFunc: func(tmpDir string) []Voit {
				return []Voit{
					{
						File: File{
							Source: filepath.Join(tmpDir, "2026-02-02T12.05.20.700 beach vacation.jpg.xmp"),
							Name:   "2026-02-02T12.05.20.700 beach vacation",
							Ext:    ".jpg.xmp",
						},
						Orig: Meta{
							VTime: VTime{Time: vTime},
							Desc:  Desc{Text: bDesc},
						},
						Mark: Meta{
							VTime: VTime{Time: vTime},
							Desc:  Desc{Text: bDesc},
						},
						Matched: false,
					},
				}
			},
			wantMapFn: func(tmpDir string, inputs VoitFiles) FileMap {
				return FileMap{
					"2026-02-02T12.05.20.700 beach vacation": {
						File:    nil,
						Sidecar: inputs[0],
					},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Generate test files on disk.
			inputFiles := make(VoitFiles, len(tt.filesSetup))
			for i, fs := range tt.filesSetup {
				fullPath := filepath.Join(tmpDir, fs.name+fs.ext)

				if fs.XMPData != "" {
					err := os.WriteFile(fullPath, []byte(fs.XMPData), 0644)
					if err != nil {
						t.Fatalf("failed to write mock XMP file: %v", err)
					}
				} else {
					err := os.WriteFile(fullPath, []byte("fake image data"), 0644)
					if err != nil {
						t.Fatalf("failed to write mock image file: %v", err)
					}
				}

				inputFiles[i] = &Voit{
					File: File{
						Source: fullPath,
						Name:   fs.name,
						Ext:    fs.ext,
					},
				}
			}

			fileMap := make(FileMap)
			inputFiles.buildFileMap(tt.opts.Voit(), fileMap)
			fileMap.ParseSidecar(tt.opts)

			// DeepEqual will compare memory addresses if pointers, not values.
			// Convert to value. This is required as pointers are needed to update
			// the struct in place during SyncXMP.
			got := make([]Voit, len(inputFiles))
			for i, ptr := range inputFiles {
				if ptr != nil {
					got[i] = *ptr
				}
			}

			wantVoit := tt.wantVoitFunc(tmpDir)
			if !reflect.DeepEqual(got, wantVoit) {
				t.Errorf("\nGot Voit:  %+v\nWant Voit: %+v", got, wantVoit)
			}

			wantMap := tt.wantMapFn(tmpDir, inputFiles)
			if !reflect.DeepEqual(fileMap, wantMap) {
				t.Errorf("\nGot FileMap:  %+v\nWant FileMap: %+v", fileMap, wantMap)
			}
		})
	}
}
