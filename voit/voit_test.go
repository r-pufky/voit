package voit

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

// Voit mock for testing.
type MockVoit struct {
	*VoitImpl // Embed interface to use real func when not stubbed out.

	RenameStub func(w io.Writer, cfg ...Config) error
}

func (m *MockVoit) Rename(w io.Writer, cfg ...Config) error {
	if m.RenameStub != nil {
		return m.RenameStub(w, cfg...)
	}
	return nil
}

var _ Voit = (*MockVoit)(nil) // Compile time check to validate interface.

func TestVoitHasFuncs(t *testing.T) {
	validFile := File{
		Path:  "/abs/path",
		Name:  "image",
		Ext:   ".jpg",
		CTime: time.Now(),
		MTime: time.Now(),
	}

	tests := []struct {
		name        string
		voit        Voit
		wantFile    bool
		wantSidecar bool
		wantBoth    bool
	}{
		{
			name: "both files empty [all false]",
			voit: &VoitImpl{
				File:    File{},
				Sidecar: File{},
			},
			wantFile:    false,
			wantSidecar: false,
			wantBoth:    false,
		},
		{
			name: "file exists sidecar missing",
			voit: &VoitImpl{
				File:    validFile,
				Sidecar: File{},
			},
			wantFile:    true,
			wantSidecar: false,
			wantBoth:    false,
		},
		{
			name: "file missing sidecar exists",
			voit: &VoitImpl{
				File:    File{},
				Sidecar: validFile,
			},
			wantFile:    false,
			wantSidecar: true,
			wantBoth:    false,
		},
		{
			name: "both files exist [all true]",
			voit: &VoitImpl{
				File:    validFile,
				Sidecar: validFile,
			},
			wantFile:    true,
			wantSidecar: true,
			wantBoth:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.voit.HasFile(); got != tt.wantFile {
				t.Errorf("\nHasFile()\nGot:  %v\nWant: %v\n", got, tt.wantFile)
			}

			if got := tt.voit.HasSidecar(); got != tt.wantSidecar {
				t.Errorf("\nHasSidecar()\nGot:  %v\nWant: %v\n", got, tt.wantSidecar)
			}

			if got := tt.voit.HasBothFiles(); got != tt.wantBoth {
				t.Errorf("\nHasBothFiles()\nGot:  %v\nWant: %v\n", got, tt.wantBoth)
			}
		})
	}
}

func TestVoitIsMatched(t *testing.T) {}

func TestVoitWidth(t *testing.T) {
	tests := []struct {
		name    string
		v       Voit
		wantMax int
	}{
		{
			name:    "empty voit returns 0",
			v:       NewVoit(),
			wantMax: 0,
		},
		{
			name: "file only width",
			v: &VoitImpl{
				File:    File{Name: "small", Ext: ".jpg"},
				Sidecar: File{},
			},
			wantMax: 9,
		},
		{
			name: "largest width found",
			v: &VoitImpl{
				File:    File{Name: "small", Ext: ".jpg"},
				Sidecar: File{Name: "small", Ext: ".jpg.xmp"},
			},
			wantMax: 13,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v.Width(); got != tt.wantMax {
				t.Errorf("\nGot:  %d\nWant: %d\n", got, tt.wantMax)
			}
		})
	}
}

func TestVoitCollisionCountNOOP(t *testing.T) {}

func TestVoitAdd(t *testing.T) {
	tests := []struct {
		name        string
		injectVoit  func(*VoitImpl)
		file        *File
		wantSidecar bool
		wantFile    bool
		wantErr     bool
	}{
		{
			name:        "file pointer is nil",
			file:        nil,
			wantSidecar: false,
			wantFile:    false,
			wantErr:     true,
		},
		{
			name: "non-sidecar file is placed in file",
			file: &File{
				Name: "holiday_photo",
				Ext:  ".jpg",
			},
			wantSidecar: false,
			wantFile:    true,
			wantErr:     false,
		},
		{
			name: "sidecar file is placed in sidecar",
			file: &File{
				Name: "holiday_photo",
				Ext:  ".jpg.xmp",
			},
			wantSidecar: true,
			wantFile:    false,
			wantErr:     false,
		},
		{
			name: "collision - primary file already exists",
			injectVoit: func(v *VoitImpl) {
				v.File = File{Name: "existing_photo", Ext: ".jpg"}
			},
			file: &File{
				Name: "holiday_photo",
				Ext:  ".jpg",
			},
			wantSidecar: false,
			wantFile:    false,
			wantErr:     true,
		},
		{
			name: "collision - sidecar file already exists",
			injectVoit: func(v *VoitImpl) {
				v.Sidecar = File{Name: "existing_photo", Ext: ".jpg.xmp"}
			},
			file: &File{
				Name: "holiday_photo",
				Ext:  ".jpg.xmp",
			},
			wantSidecar: false,
			wantFile:    false,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &VoitImpl{}
			if tt.injectVoit != nil {
				tt.injectVoit(v)
			}

			err := v.Add(tt.file)

			if (err != nil) != tt.wantErr {
				t.Fatalf("\nError\nGot:  %v\nWant: %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			if tt.wantFile {
				if v.File.Name != tt.file.Name || v.File.Ext != tt.file.Ext {
					t.Errorf("\nFile\nGot:  (%q, %q)\nWant: (%q, %q)\n", v.File.Name, v.File.Ext, tt.file.Name, tt.file.Ext)
				}
				if v.Sidecar.Name != "" || v.Sidecar.Ext != "" {
					t.Errorf("\nSidecar\nGot:  %q\nWant: ''\n", v.Sidecar.Name)
				}
			}

			if tt.wantSidecar {
				if v.Sidecar.Name != tt.file.Name || v.Sidecar.Ext != tt.file.Ext {
					t.Errorf("\nSidecar\nGot:  (%q, %q)\nWant: (%q, %q)\n", v.Sidecar.Name, v.Sidecar.Ext, tt.file.Name, tt.file.Ext)
				}
				if v.File.Name != "" || v.File.Ext != "" {
					t.Errorf("\nFile\nGot:  %q\nWant: ''\n", v.File.Name)
				}
			}
		})
	}
}

func TestVoitExtractMetadata(t *testing.T) {
	fixedCTime := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	fixedMTime := time.Date(2026, time.February, 1, 14, 0, 0, 0, time.UTC)
	fixedSTime := time.Date(2026, time.March, 1, 20, 0, 0, 0, time.UTC)
	parsedTime := time.Date(2026, time.May, 17, 10, 45, 36, 300000000, time.UTC)
	voitTime := time.Date(2026, time.February, 2, 12, 5, 20, 700000000, time.UTC)
	baseTags := []string{"summer", "vacation", "beach"}
	baseDesc := "beach vacation"

	tests := []struct {
		name     string
		f        Voit
		args     Config
		wantVoit Voit
	}{
		{
			name: "sanitized format",
			f: &VoitImpl{
				File: File{
					Path: "/tmp",
					Name: "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
					Ext:  ".jpg",
				},
			},
			args: Config{Voit: VoitConfig{Pattern: "photo"}},
			wantVoit: &VoitImpl{
				File: File{
					Path: "/tmp",
					Name: "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
					Ext:  ".jpg",
				},
				Orig: Meta{
					VTime: VTime{Time: voitTime},
					Tags:  Tag{Items: baseTags},
					Desc:  Desc{Text: baseDesc},
				},
			},
		},
		{
			name: "alternative separators",
			f: &VoitImpl{
				File: File{
					Path: "/tmp",
					Name: "2026-02-02T12.05.20.700|beach vacation - summer vacation beach",
					Ext:  ".jpg",
				},
			},
			args: Config{Voit: VoitConfig{Pattern: "photo", DescSep: "|", TagSep: " - "}},
			wantVoit: &VoitImpl{
				File: File{
					Path: "/tmp",
					Name: "2026-02-02T12.05.20.700|beach vacation - summer vacation beach",
					Ext:  ".jpg",
				},
				Orig: Meta{
					VTime: VTime{Time: voitTime},
					Tags:  Tag{Items: baseTags},
					Desc:  Desc{Text: baseDesc},
				},
			},
		},
		{
			name: "duplicate separators",
			f: &VoitImpl{
				File: File{
					Path: "/tmp",
					Name: "2026-02-02T12.05.20.700|beach vacation|summer vacation beach",
					Ext:  ".jpg",
				},
			},
			args: Config{Voit: VoitConfig{Pattern: "photo", DescSep: "|", TagSep: "|"}},
			wantVoit: &VoitImpl{
				File: File{
					Path: "/tmp",
					Name: "2026-02-02T12.05.20.700|beach vacation|summer vacation beach",
					Ext:  ".jpg",
				},
				Orig: Meta{
					VTime: VTime{Time: voitTime},
					Tags:  Tag{Items: baseTags},
					Desc:  Desc{Text: baseDesc},
				},
			},
		},
		{
			name: "fields: missing desc",
			f: &VoitImpl{
				File: File{
					Path: "/tmp",
					Name: "2026-02-02T12.05.20.700  -- summer vacation beach",
					Ext:  ".jpg",
				},
			},
			args: Config{Voit: VoitConfig{Pattern: "photo"}},
			wantVoit: &VoitImpl{
				File: File{
					Path: "/tmp",
					Name: "2026-02-02T12.05.20.700  -- summer vacation beach",
					Ext:  ".jpg",
				},
				Orig: Meta{
					VTime: VTime{Time: voitTime},
					Tags:  Tag{Items: baseTags},
				},
			},
		},
		{
			name: "fields: missing tags",
			f: &VoitImpl{
				File: File{
					Path: "/tmp",
					Name: "2026-02-02T12.05.20.700 beach vacation",
					Ext:  ".jpg",
				},
			},
			args: Config{Voit: VoitConfig{Pattern: "photo"}},
			wantVoit: &VoitImpl{
				File: File{
					Path: "/tmp",
					Name: "2026-02-02T12.05.20.700 beach vacation",
					Ext:  ".jpg",
				},
				Orig: Meta{
					VTime: VTime{Time: voitTime},
					Desc:  Desc{Text: baseDesc},
				},
			},
		},
		{
			name: "fields: missing desc tags",
			f: &VoitImpl{
				File: File{
					Path: "/tmp",
					Name: "2026-02-02T12.05.20.700",
					Ext:  ".jpg",
				},
			},
			args: Config{Voit: VoitConfig{Pattern: "photo"}},
			wantVoit: &VoitImpl{
				File: File{
					Path: "/tmp",
					Name: "2026-02-02T12.05.20.700",
					Ext:  ".jpg",
				},
				Orig: Meta{
					VTime: VTime{Time: voitTime},
					Desc:  Desc{Text: "2026-02-02T12.05.20.700"},
				},
			},
		},
		{
			name: "patterns: ctime",
			f: &VoitImpl{
				File: File{
					Path:  "/tmp",
					Name:  "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
					Ext:   ".jpg",
					CTime: fixedCTime,
				},
			},
			args: Config{},
			wantVoit: &VoitImpl{
				File: File{
					Path:  "/tmp",
					Name:  "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
					Ext:   ".jpg",
					CTime: fixedCTime,
				},
				Orig: Meta{
					VTime: VTime{Time: voitTime},
					PTime: VTime{Time: fixedCTime},
					Tags:  Tag{Items: baseTags},
					Desc:  Desc{Text: baseDesc},
				},
			},
		},
		{
			name: "patterns: mtime",
			f: &VoitImpl{
				File: File{
					Path:  "/tmp",
					Name:  "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
					Ext:   ".jpg",
					MTime: fixedMTime,
				},
			},
			args: Config{Voit: VoitConfig{Pattern: "modified"}},
			wantVoit: &VoitImpl{
				File: File{
					Path:  "/tmp",
					Name:  "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
					Ext:   ".jpg",
					MTime: fixedMTime,
				},
				Orig: Meta{
					VTime: VTime{Time: voitTime},
					PTime: VTime{Time: fixedMTime},
					Tags:  Tag{Items: baseTags},
					Desc:  Desc{Text: baseDesc},
				},
			},
		},
		{
			name: "patterns: set",
			f: &VoitImpl{
				File: File{
					Path: "/tmp",
					Name: "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
					Ext:  ".jpg",
				},
			},
			args: Config{Voit: VoitConfig{Pattern: "set", Set: "2026-03-01T20.00.00.000"}},
			wantVoit: &VoitImpl{
				File: File{
					Path: "/tmp",
					Name: "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
					Ext:  ".jpg",
				},
				Orig: Meta{
					VTime: VTime{Time: voitTime},
					PTime: VTime{Time: fixedSTime},
					Tags:  Tag{Items: baseTags},
					Desc:  Desc{Text: baseDesc},
				},
			},
		},
		{ // Regex patterns tested in regex_test.go.
			name: "patterns: regex",
			f: &VoitImpl{
				File: File{
					Path: "/tmp",
					Name: "20260517_104536300",
					Ext:  ".jpg",
				},
			},
			args: Config{Voit: VoitConfig{Pattern: "photo-ms"}},
			wantVoit: &VoitImpl{
				File: File{
					Path: "/tmp",
					Name: "20260517_104536300",
					Ext:  ".jpg",
				},
				Orig: Meta{
					PTime: VTime{Time: parsedTime},
					Desc:  Desc{Text: "20260517_104536300"},
				},
			},
		},
		{
			name: "patterns: no match",
			f: &VoitImpl{
				File: File{
					Path: "/tmp",
					Name: "no_match",
					Ext:  ".jpg",
				},
			},
			args: Config{Voit: VoitConfig{Pattern: "photo-ms"}},
			wantVoit: &VoitImpl{
				File: File{
					Path: "/tmp",
					Name: "no_match",
					Ext:  ".jpg",
				},
				Orig: Meta{
					Desc: Desc{Text: "no_match"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.f.ExtractMetadata(tt.args)

			if !reflect.DeepEqual(tt.f, tt.wantVoit) {
				t.Errorf("\nGot:  %+v\nWant: %+v\n", tt.f, tt.wantVoit)
			}
		})
	}
}

func TestVoitExtractXMP(t *testing.T) {
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

	type fileSetup struct {
		name string
		ext  string
	}

	tests := []struct {
		name       string
		opts       *Opts
		filesSetup []fileSetup
		sidecar    File
		XMPData    string
		hasSidecar bool
		wantDoc    bool
		wantErr    bool
	}{
		{
			name: "sanity: valid digikam XMP sidecar",
			filesSetup: []fileSetup{
				{name: "2026-02-02T12.05.20.700 beach vacation", ext: ".jpg.xmp"},
			},
			sidecar: File{
				Path: "/mock/path",
				Name: "2026-02-02T12.05.20.700 beach vacation",
				Ext:  ".jpg.xmp",
			},
			XMPData: validDigikamXMP,
			wantDoc: true,
			wantErr: false,
		},
		{
			name: "sanity: empty digikam XMP sidecar",
			filesSetup: []fileSetup{
				{name: "2026-02-02T12.05.20.700 beach vacation", ext: ".jpg.xmp"},
			},
			sidecar: File{
				Path: "/mock/path",
				Name: "2026-02-02T12.05.20.700 beach vacation",
				Ext:  ".jpg.xmp",
			},
			XMPData: emptyDigikamXMP,
			wantDoc: true,
			wantErr: false,
		},
		{
			name:       "sanity: non-existent file returns error",
			filesSetup: []fileSetup{},
			sidecar: File{
				Path: "/mock/path",
				Name: "does_not_exist",
				Ext:  ".xmp",
			},
			XMPData: "",
			wantDoc: false,
			wantErr: true,
		},
		{
			name:       "sanity: no sidecar returns nil",
			filesSetup: []fileSetup{},
			sidecar:    File{},
			XMPData:    "",
			wantDoc:    true,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFS := MockFS{
				MockOpen: func(name string) (io.ReadCloser, error) {
					for _, fs := range tt.filesSetup {
						expectedName := fs.name + fs.ext
						if filepath.Base(name) == expectedName {
							return io.NopCloser(bytes.NewBufferString(tt.XMPData)), nil
						}
					}
					return nil, os.ErrNotExist
				},
			}

			voitImpl := &VoitImpl{
				Sidecar: tt.sidecar,
			}

			cfg := Config{FS: mockFS}
			doc, err := voitImpl.ExtractXMP(cfg)

			if tt.wantErr && err == nil {
				t.Error("\nGot:  nil\nWant: err\n")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("\nGot:  %v\nWant: nil\n", err)
			}

			if tt.wantDoc && doc == nil && voitImpl.HasSidecar() && tt.XMPData != "" {
				t.Error("\nGot:  nil\nWant: parsed document\n")
			}
			if !tt.wantDoc && doc != nil {
				t.Error("\nGot:  parsed document\nWant: nil\n")
			}
		})
	}
}

func TestVoitNameExtAbs(t *testing.T) {
	t.Parallel()

	voitTime := time.Date(2026, time.February, 2, 12, 5, 20, 700000000, time.UTC)

	tests := []struct {
		name      string
		c         Config
		v         Voit
		wantFName string
		wantFExt  string
		wantFAbs  string
		wantSName string
		wantSExt  string
		wantSAbs  string
	}{
		{
			name: "sanity target built correctly",
			c:    Config{},
			v: &VoitImpl{
				File: File{
					Path: "/path/to/files",
					Name: "photo",
					Ext:  ".jpg",
				},
				Sidecar: File{
					Path: "/path/to/files",
					Name: "photo",
					Ext:  ".jpg.xmp",
				},
				Dest: Meta{
					VTime: VTime{Time: voitTime},
				},
			},
			wantFName: "photo",
			wantFExt:  ".jpg",
			wantFAbs:  "/path/to/files/2026-02-02T12.05.20.700.jpg",
			wantSName: "photo",
			wantSExt:  ".jpg.xmp",
			wantSAbs:  "/path/to/files/2026-02-02T12.05.20.700.jpg.xmp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			nameFile, nameSidecar := tt.v.SourceName()
			extFile, extSidecar := tt.v.SourceExt()
			absFile, absSidecar := tt.v.Abs(tt.c)

			if nameFile != tt.wantFName {
				t.Errorf("\nFile name\nGot:  %q\nWant: %q\n", nameFile, tt.wantFName)
			}
			if extFile != tt.wantFExt {
				t.Errorf("\nFile extension\nGot:  %q\nWant: %q\n", extFile, tt.wantFExt)
			}
			if absFile != tt.wantFAbs {
				t.Errorf("\nFile absolute\nGot:  %q\nWant: %q\n", absFile, tt.wantFAbs)
			}

			if nameSidecar != tt.wantSName {
				t.Errorf("\nSidecar name\nGot:  %q\nWant: %q\n", nameSidecar, tt.wantSName)
			}
			if extSidecar != tt.wantSExt {
				t.Errorf("\nSidecar extension\nGot:  %q\nWant: %q\n", extSidecar, tt.wantSExt)
			}
			if absSidecar != tt.wantSAbs {
				t.Errorf("\nSidecar absolute\nGot:  %q\nWant: %q\n", absSidecar, tt.wantSAbs)
			}
		})
	}
}

func TestVoitExists(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		v        Voit
		mockStat func(string) (os.FileInfo, error)
		want     bool
	}{
		{
			name: "no file, no sidecar [false]",
			v:    NewVoit(),
			mockStat: func(name string) (os.FileInfo, error) {
				return nil, nil
			},
			want: false,
		},
		{
			name: "file (missing), no sidecar [false]",
			v: &VoitImpl{
				File: File{Path: "dir", Ext: ".jpg"},
			},
			mockStat: func(name string) (os.FileInfo, error) {
				return nil, os.ErrNotExist
			},
			want: false,
		},
		{
			name: "file (exists), no sidecar [true]",
			v: &VoitImpl{
				File: File{Path: "dir", Ext: ".jpg"},
			},
			mockStat: func(name string) (os.FileInfo, error) {
				return nil, nil
			},
			want: true,
		},
		{
			name: "no file, sidecar (missing) [false]",
			v: &VoitImpl{
				Sidecar: File{Path: "dir", Ext: ".xmp"},
			},
			mockStat: func(name string) (os.FileInfo, error) {
				return nil, os.ErrNotExist
			},
			want: false,
		},
		{
			name: "no file, sidecar (exists) [true]",
			v: &VoitImpl{
				Sidecar: File{Path: "dir", Ext: ".xmp"},
			},
			mockStat: func(name string) (os.FileInfo, error) {
				return nil, nil
			},
			want: true,
		},
		{
			name: "file (missing), sidecar (exists) [false]",
			v: &VoitImpl{
				File:    File{Path: "dir", Ext: ".jpg"},
				Sidecar: File{Path: "dir", Ext: ".xmp"},
			},
			mockStat: func(name string) (os.FileInfo, error) {
				if strings.HasSuffix(name, ".jpg") {
					return nil, os.ErrNotExist
				}
				return nil, nil
			},
			want: false,
		},
		{
			name: "file (exists), sidecar (missing) [false]",
			v: &VoitImpl{
				File:    File{Path: "dir", Ext: ".jpg"},
				Sidecar: File{Path: "dir", Ext: ".xmp"},
			},
			mockStat: func(name string) (os.FileInfo, error) {
				if strings.HasSuffix(name, ".xmp") {
					return nil, os.ErrNotExist
				}
				return nil, nil
			},
			want: false,
		},
		{
			name: "file (exists), sidecar (exists) [true]",
			v: &VoitImpl{
				File:    File{Path: "dir", Ext: ".jpg"},
				Sidecar: File{Path: "dir", Ext: ".xmp"},
			},
			mockStat: func(name string) (os.FileInfo, error) {
				return nil, nil
			},
			want: true,
		},
		{
			name: "permission denied is treated as existing [true]",
			v: &VoitImpl{
				File: File{Path: "dir", Ext: ".jpg"},
			},
			mockStat: func(name string) (os.FileInfo, error) {
				return nil, errors.New("permission denied")
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := NewConfig(Config{
				Voit: VoitConfig{VFormat: "2006-01-02"},
				FS:   MockFS{MockStat: tt.mockStat},
			})

			got := tt.v.Exists(cfg)
			if got != tt.want {
				t.Errorf("\nGot:  %v\nWant: %v\n", got, tt.want)
			}
		})
	}
}

func TestVoitRename(t *testing.T) {
	t.Parallel()
	voitTime := time.Date(2026, time.February, 2, 12, 5, 20, 700000000, time.UTC)
	voitTarget := Meta{VTime: VTime{Time: voitTime}}

	tests := []struct {
		name       string
		args       Config
		voit       Voit
		mockStat   func(name string) (os.FileInfo, error)
		mockRename func(oldpath, newpath string) error
		wantErr    bool
	}{
		{
			name: "no match",
			args: Config{Voit: VoitConfig{Overwrite: false}},
			voit: &VoitImpl{
				Matched: false,
				File:    File{Path: "/src", Name: "photo", Ext: ".jpg"},
				Dest:    voitTarget,
			},
			mockStat: func(name string) (os.FileInfo, error) {
				t.Error("\nosStat should not be called when Matched is false.\n")
				return nil, nil
			},
			mockRename: func(oldpath, newpath string) error {
				t.Error("\nosRename should not be called when Matched is false.\n")
				return nil
			},
			wantErr: false,
		},
		{
			name: "file no sidecar",
			args: Config{Voit: VoitConfig{Overwrite: true}},
			voit: &VoitImpl{
				Matched: true,
				File:    File{Path: "/src", Name: "photo", Ext: ".jpg"},
				Dest:    voitTarget,
			},
			mockStat: func(name string) (os.FileInfo, error) {
				return nil, os.ErrNotExist
			},
			mockRename: func(oldpath, newpath string) error {
				wantOld := "/src/photo.jpg"
				wantNew := "/src/2026-02-02T12.05.20.700.jpg"

				if oldpath != wantOld {
					t.Errorf("\nUnexpected oldpath\nGot:  %q\nWant: %q\n", oldpath, wantOld)
				}
				if newpath != wantNew {
					t.Errorf("\nUnexpected newpath\nGot:  %q\nWant: %q\n", newpath, wantNew)
				}
				return nil
			},
			wantErr: false,
		},
		{
			name: "file and sidecar",
			args: Config{Voit: VoitConfig{Overwrite: true}},
			voit: &VoitImpl{
				Matched: true,
				File:    File{Path: "/src", Name: "photo", Ext: ".jpg"},
				Sidecar: File{Path: "/src", Name: "photo", Ext: ".xmp"},
				Dest:    voitTarget,
			},
			mockStat:   func(name string) (os.FileInfo, error) { return nil, os.ErrNotExist },
			mockRename: func(oldpath, newpath string) error { return nil },
			wantErr:    false,
		},
		{
			name: "sidecar no file",
			args: Config{Voit: VoitConfig{Overwrite: false}},
			voit: &VoitImpl{
				Matched: true,
				Sidecar: File{Path: "/src", Name: "photo", Ext: ".xmp"},
				Dest:    voitTarget,
			},
			mockStat:   func(name string) (os.FileInfo, error) { return nil, os.ErrNotExist },
			mockRename: func(oldpath, newpath string) error { return nil },
			wantErr:    false,
		},
		{
			name: "file collision [error]",
			args: Config{Voit: VoitConfig{Overwrite: false}},
			voit: &VoitImpl{
				Matched: true,
				File:    File{Path: "/src", Name: "photo", Ext: ".jpg"},
				Dest:    voitTarget,
			},
			mockStat: func(name string) (os.FileInfo, error) { return nil, nil },
			mockRename: func(oldpath, newpath string) error {
				t.Error("\nRename should not be called on a collision error.\n")
				return nil
			},
			wantErr: true,
		},
		{ // transactionMove is tested separately, just ensure it is called.
			name: "transactionMove [success]",
			args: Config{Voit: VoitConfig{Overwrite: false}},
			voit: &VoitImpl{
				Matched: true,
				File:    File{Path: "/src", Name: "photo", Ext: ".jpg"},
				Sidecar: File{Path: "/src", Name: "photo", Ext: ".xmp"},
				Dest:    voitTarget,
			},
			mockStat: func(name string) (os.FileInfo, error) {
				return nil, os.ErrNotExist
			},
			mockRename: func(oldpath, newpath string) error {
				return nil
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.args.FS = MockFS{
				MockStat:   tt.mockStat,
				MockRename: tt.mockRename,
			}

			var buf bytes.Buffer
			err := tt.voit.Rename(&buf, tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("\nGot:  %v\nWant: %v\n", err, tt.wantErr)
			}
		})
	}
}

func TestVoitTransactionMove(t *testing.T) {
	t.Parallel()
	voitTime := time.Date(2026, time.February, 2, 12, 5, 20, 0, time.UTC)

	testAsset := func() *VoitImpl {
		return &VoitImpl{
			Matched: true,
			File:    File{Path: "/src", Name: "photo", Ext: ".jpg"},
			Sidecar: File{Path: "/src", Name: "photo", Ext: ".xmp"},
			Dest:    Meta{VTime: VTime{Time: voitTime}},
		}
	}

	tests := []struct {
		name       string
		args       Config
		mockStat   func(name string) (os.FileInfo, error)
		mockRename func(oldpath, newpath string) error
		wantErr    bool
		wantLogStr string
	}{
		{
			name: "standard asset move [success]",
			args: Config{},
			mockStat: func(name string) (os.FileInfo, error) {
				return nil, os.ErrNotExist
			},
			mockRename: func(oldpath, newpath string) error {
				return nil
			},
			wantErr: false,
		},
		{
			name: "file collision [error]",
			args: Config{},
			mockStat: func(name string) (os.FileInfo, error) {
				if strings.HasSuffix(name, ".jpg") {
					return nil, nil
				}
				return nil, os.ErrNotExist
			},
			mockRename: func(oldpath, newpath string) error { return nil },
			wantErr:    true,
		},
		{
			name: "sidecar collision [error]",
			args: Config{},
			mockStat: func(name string) (os.FileInfo, error) {
				if strings.HasSuffix(name, ".xmp") {
					return nil, nil
				}
				return nil, os.ErrNotExist
			},
			mockRename: func(oldpath, newpath string) error { return nil },
			wantErr:    true,
		},
		{
			name: "overwrite rename failure [error]",
			args: Config{Voit: VoitConfig{Overwrite: true}},
			mockStat: func(name string) (os.FileInfo, error) {
				return nil, os.ErrNotExist
			},
			mockRename: func(oldpath, newpath string) error {
				if strings.HasSuffix(oldpath, ".jpg") {
					return errors.New("permission denied")
				}
				return nil
			},
			wantErr: true,
		},
		{
			name: "sidecar fail rollback success [error, rollback success]",
			args: Config{Voit: VoitConfig{Overwrite: true}},
			mockStat: func(name string) (os.FileInfo, error) {
				return nil, os.ErrNotExist
			},
			mockRename: func(oldpath, newpath string) error {
				// Allow primary file to move, but fail on sidecar.
				if strings.HasSuffix(oldpath, ".xmp") {
					return errors.New("disk full")
				}
				return nil
			},
			wantErr:    true,
			wantLogStr: "Failed to move sidecar file, rolling back File:",
		},
		{
			name: "sidecar fail rollback fail [error, rollback error]",
			args: Config{Voit: VoitConfig{Overwrite: true}},
			mockStat: func(name string) (os.FileInfo, error) {
				return nil, os.ErrNotExist
			},
			mockRename: func(oldpath, newpath string) error {
				// Allow primary file to move, but fail on sidecar.
				if strings.HasSuffix(oldpath, ".xmp") {
					return errors.New("disk full")
				}
				// Rollback fails.
				if strings.Contains(oldpath, "2026") && strings.Contains(newpath, "/src") {
					return errors.New("device disconnected")
				}
				return nil
			},
			wantErr:    true,
			wantLogStr: "CRITICAL: Failed to roll back primary file:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.args.FS = MockFS{
				MockStat:   tt.mockStat,
				MockRename: tt.mockRename,
			}

			v := testAsset()

			var buf bytes.Buffer
			err := v.transactionMove(&buf, tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("\nGot:  %v\nWant: %v\n", err, tt.wantErr)
			}

			if tt.wantLogStr != "" && !strings.Contains(buf.String(), tt.wantLogStr) {
				t.Errorf("\nLog mismatch\nGot:  %q\nWant substring: %q\n", buf.String(), tt.wantLogStr)
			}
		})
	}
}

func TestVoitImplRender(t *testing.T) {
	voitTime := time.Date(2026, time.February, 2, 12, 5, 20, 0, time.UTC)

	tests := []struct {
		name       string
		voit       *VoitImpl
		minWidth   int
		wantOutput string
		wantLines  int
	}{
		{
			name: "file sidecar",
			voit: &VoitImpl{
				File:    File{Name: "file", Ext: ".jpg"},
				Sidecar: File{Name: "file", Ext: ".jpg.xmp"},
				Dest:    Meta{VTime: VTime{Time: voitTime}},
			},
			minWidth: 13,
			wantOutput: "╭file.jpg       ➔ 2026-02-02T12.05.20.000.jpg\n" +
				"╰file.jpg.xmp   ➔ 2026-02-02T12.05.20.000.jpg.xmp\n",
			wantLines: 2,
		},
		{
			name: "file",
			voit: &VoitImpl{
				File: File{Name: "file", Ext: ".jpg"},
				Dest: Meta{VTime: VTime{Time: voitTime}},
			},
			minWidth:   10,
			wantOutput: " file.jpg    ➔ 2026-02-02T12.05.20.000.jpg\n",
			wantLines:  1,
		},
		{
			name: "sidecar",
			voit: &VoitImpl{
				Sidecar: File{Name: "file", Ext: ".jpg.xmp"},
				Dest:    Meta{VTime: VTime{Time: voitTime}},
			},
			minWidth:   15,
			wantOutput: " file.jpg.xmp     ➔ 2026-02-02T12.05.20.000.jpg.xmp\n",
			wantLines:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			count := tt.voit.Render(&buf, tt.minWidth)

			if count != tt.wantLines {
				t.Errorf("\nGot:  %d\nWant: %d\n", count, tt.wantLines)
			}

			got := buf.String()
			if got != tt.wantOutput {
				t.Errorf("\nGot:  %q\nWant: %q\n", got, tt.wantOutput)
			}
		})
	}
}
