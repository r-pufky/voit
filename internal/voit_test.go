package internal

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/r-pufky/voit/models"
)

var (
	fixedCTime      = time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	fixedMTime      = time.Date(2026, time.February, 1, 14, 0, 0, 0, time.UTC)
	parsedTime      = time.Date(2026, time.May, 17, 10, 45, 36, 300000000, time.UTC)
	formattedTime   = "2026-05-17T10.45.36.300"
	parsedTimeNoMS  = time.Date(2026, time.May, 17, 10, 45, 36, 0, time.UTC)
	parsedTimeShort = time.Date(2026, time.May, 17, 10, 45, 0, 0, time.UTC)
	voitTime        = time.Date(2026, time.February, 2, 12, 5, 20, 700000000, time.UTC)
	baseTags        = []string{"summer", "vacation", "beach"}
	baseDesc        = "beach vacation"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name      string
		fName     string
		pattern   string
		force     bool
		descSep   string
		tagsSep   string
		wantVTime time.Time
		wantTags  []string
		wantDesc  string
	}{
		// Sanity checks.
		{
			name:      "sanity: sanitized format",
			fName:     "2026-02-02T12.05.20.700 - beach vacation -- summer vacation beach",
			pattern:   "photo",
			wantVTime: voitTime,
			wantDesc:  baseDesc,
			wantTags:  baseTags,
		},
		{
			name:      "sanity: sanitized format [desc:' ', tag:' - ']",
			fName:     "2026-02-02T12.05.20.700 beach vacation - summer vacation beach",
			pattern:   "photo",
			descSep:   " ",
			tagsSep:   " - ",
			wantVTime: voitTime,
			wantDesc:  baseDesc,
			wantTags:  baseTags,
		},
		{
			name:      "sanity: no matches",
			fName:     "not-a-date-file",
			pattern:   "photo",
			wantVTime: time.Time{},
			wantDesc:  "not-a-date-file",
			wantTags:  []string{},
		},
		// Tags.
		{
			name:      "tags: lowercased",
			fName:     "2026-02-02T12.05.20.700 - beach vacation -- SUMMER VACATION BEACH",
			pattern:   "photo",
			wantVTime: voitTime,
			wantDesc:  baseDesc,
			wantTags:  baseTags,
		},
		{
			name:      "tags: digikam tag spacers",
			fName:     "2026-02-02T12.05.20.700 - beach vacation -- nested_tag_summer vacation beach",
			pattern:   "photo",
			wantVTime: voitTime,
			wantDesc:  baseDesc,
			wantTags:  []string{"nested_tag_summer", "vacation", "beach"},
		},
		{
			name:      "tags: empty [trailing space]",
			fName:     "2026-02-02T12.05.20.700 - beach vacation -- ",
			pattern:   "photo",
			wantVTime: voitTime,
			wantDesc:  baseDesc,
			wantTags:  []string{},
		},
		{
			name:      "tags: empty [invalid separator]",
			fName:     "2026-02-02T12.05.20.700 - beach vacation --",
			pattern:   "photo",
			wantVTime: voitTime,
			wantDesc:  "beach vacation --",
			wantTags:  []string{},
		},
		// VTIME.
		{
			name:      "vtime: non matched date pattern",
			fName:     "2026-02-02T12.05.20.700 - 2026-05-02T17.10.45.000 beach vacation -- summer vacation beach",
			pattern:   "photo",
			wantVTime: voitTime,
			wantDesc:  "2026-05-02T17.10.45.000 beach vacation",
			wantTags:  baseTags,
		},
		{
			name:      "vtime: non matched pattern [desc:' ', tag:' - ']",
			fName:     "2026-02-02T12.05.20.700 2026-05-02T17.10.45.000 beach vacation - summer vacation beach",
			pattern:   "photo",
			descSep:   " ",
			tagsSep:   " - ",
			wantVTime: voitTime,
			wantDesc:  "2026-05-02T17.10.45.000 beach vacation",
			wantTags:  baseTags,
		},
		{
			name:      "vtime: invalid separator causing regex to fail full name to description",
			fName:     "2026-02-02T12.05.20.700 2026-05-02T17.10.45.000 beach vacation -- summer vacation beach",
			pattern:   "photo",
			wantVTime: time.Time{},
			wantDesc:  "2026-02-02T12.05.20.700 2026-05-02T17.10.45.000 beach vacation",
			wantTags:  baseTags,
		},
		// Desc.
		{
			name:      "desc: missing tag separator",
			fName:     "2026-02-02T12.05.20.700 - beach vacation",
			pattern:   "photo",
			wantVTime: voitTime,
			wantDesc:  baseDesc,
			wantTags:  []string{},
		},
		{
			name:      "desc: missing description [desc: '', all separator space]",
			fName:     "2026-02-02T12.05.20.700 -  -- summer vacation beach",
			pattern:   "photo",
			wantVTime: voitTime,
			wantDesc:  "",
			wantTags:  baseTags,
		},
		{
			name:      "desc: missing description [desc: '', separators overlap]",
			fName:     "2026-02-02T12.05.20.700 - -- summer vacation beach",
			pattern:   "photo",
			wantVTime: time.Time{}, // Will not parse as ' - ' is not found.
			wantDesc:  "2026-02-02T12.05.20.700 -",
			wantTags:  baseTags,
		},
		{
			name:      "desc: missing date separator",
			fName:     "2026-02-02T12.05.20.700 beach vacation -- SUMMER VACATION BEACH",
			pattern:   "photo",
			wantVTime: time.Time{}, // Will not parse as ' - ' is not found.
			wantDesc:  "2026-02-02T12.05.20.700 beach vacation",
			wantTags:  baseTags,
		},
		// Patterns.
		{
			name:      "pattern: ctime used",
			fName:     "beach vacation -- summer vacation beach",
			pattern:   "created",
			wantVTime: fixedCTime,
			wantDesc:  baseDesc,
			wantTags:  baseTags,
		},
		{
			name:      "pattern: ctime used with vtime",
			fName:     "2026-02-02T12.05.20.700 - beach vacation -- summer vacation beach",
			pattern:   "created",
			wantVTime: voitTime,
			wantDesc:  baseDesc,
			wantTags:  baseTags,
		},
		{
			name:      "pattern: ctime used with vtime [forced]",
			fName:     "2026-02-02T12.05.20.700 - beach vacation -- summer vacation beach",
			pattern:   "created",
			force:     true,
			wantVTime: fixedCTime,
			wantDesc:  baseDesc,
			wantTags:  baseTags,
		},
		{
			name:      "pattern: mtime used",
			fName:     "beach vacation -- summer vacation beach",
			pattern:   "modified",
			wantVTime: fixedMTime,
			wantDesc:  baseDesc,
			wantTags:  baseTags,
		},
		{
			name:      "pattern: mtime used with vtime",
			fName:     "2026-02-02T12.05.20.700 - beach vacation -- summer vacation beach",
			pattern:   "modified",
			wantVTime: voitTime,
			wantDesc:  baseDesc,
			wantTags:  baseTags,
		},
		{
			name:      "pattern: mtime used with vtime [forced]",
			fName:     "beach vacation -- summer vacation beach",
			pattern:   "modified",
			force:     true,
			wantVTime: fixedMTime,
			wantDesc:  baseDesc,
			wantTags:  baseTags,
		},
		// Actual filename tests.
		{
			name:      "actual: photo-ms",
			fName:     "PXL_20260517_104536300~1",
			pattern:   "photo-ms",
			wantVTime: parsedTime,
			wantDesc:  "PXL_20260517_104536300~1",
			wantTags:  []string{},
		},
		{
			name:      "actual: photo",
			fName:     "PXL_20260517_104536~1",
			pattern:   "photo",
			wantVTime: parsedTimeNoMS,
			wantDesc:  "PXL_20260517_104536~1",
			wantTags:  []string{},
		},
		{
			name:      "actual: signal-ms",
			fName:     "2026-05-17-10-45-36-300~1",
			pattern:   "signal-ms",
			wantVTime: parsedTime,
			wantDesc:  "2026-05-17-10-45-36-300~1",
			wantTags:  []string{},
		},
		{
			name:      "actual: signal",
			fName:     "2026-05-17-10-45-36~1",
			pattern:   "signal",
			wantVTime: parsedTimeNoMS,
			wantDesc:  "2026-05-17-10-45-36~1",
			wantTags:  []string{},
		},
		{
			name:      "actual: 8601-naked-ms",
			fName:     "20260517104536300~1",
			pattern:   "8601-naked-ms",
			wantVTime: parsedTime,
			wantDesc:  "20260517104536300~1",
			wantTags:  []string{},
		},
		{
			name:      "actual: 8601-naked",
			fName:     "IMG_20260517104536~1",
			pattern:   "8601-naked",
			wantVTime: parsedTimeNoMS,
			wantDesc:  "IMG_20260517104536~1",
			wantTags:  []string{},
		},
		{
			name:      "actual: 8601-short",
			fName:     "2026-05-17-1045~1",
			pattern:   "8601-short",
			wantVTime: parsedTimeShort,
			wantDesc:  "2026-05-17-1045~1",
			wantTags:  []string{},
		},
		{
			name:      "actual: 8601",
			fName:     "2026-05-17-104536~1",
			pattern:   "8601",
			wantVTime: parsedTimeNoMS,
			wantDesc:  "2026-05-17-104536~1",
			wantTags:  []string{},
		},
		{
			name:      "actual: 8601-ms",
			fName:     "2026-05-17-104536300~1",
			pattern:   "8601-ms",
			wantVTime: parsedTime,
			wantDesc:  "2026-05-17-104536300~1",
			wantTags:  []string{},
		},
		{
			name:      "actual: webkit-chrome",
			fName:     "13423488336300000",
			pattern:   "webkit-chrome",
			wantVTime: parsedTime,
			wantDesc:  "13423488336300000",
			wantTags:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := &models.File{
				Name:  tt.fName,
				CTime: fixedCTime,
				MTime: fixedMTime,
			}

			Parse(file, tt.pattern, tt.force, tt.descSep, tt.tagsSep)

			if !file.VTime.Equal(tt.wantVTime) {
				t.Errorf("\nVTime:     %v\nwantVTime: %v", file.VTime, tt.wantVTime)
			}

			if file.Desc != tt.wantDesc {
				t.Errorf("\nDesc:     %q\nwantDesc: %q", file.Desc, tt.wantDesc)
			}

			if len(file.Tags) == 0 && len(tt.wantTags) == 0 {
				return
			}
			if !reflect.DeepEqual(file.Tags, tt.wantTags) {
				t.Errorf("\nTags:     %q\nwantTags: %q", file.Tags, tt.wantTags)
			}
		})
	}
}

func TestExtract(t *testing.T) {
	// Regex tested in models.

	tests := []struct {
		name        string
		file        string
		pattern     string
		wantTime    time.Time
		wantErr     bool
		errContains string
	}{
		{
			name:        "no match",
			file:        "not-a-date-file.txt",
			pattern:     "photo",
			wantTime:    time.Time{},
			wantErr:     true,
			errContains: "No date pattern matched file",
		},
		{
			name:     "webkit-chrome match",
			file:     "13253932800000000.dat",
			pattern:  "webkit-chrome",
			wantTime: time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "full pattern parse extraction",
			file:     "20260517_112356123.jpg",
			pattern:  "photo-ms",
			wantTime: time.Date(2026, time.May, 17, 11, 23, 56, 123*int(time.Millisecond), time.UTC),
			wantErr:  false,
		},
		{
			name:     "partial pattern extraction",
			file:     "2026-05-17",
			pattern:  "voit",
			wantTime: time.Date(2026, time.May, 17, 0, 0, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:        "no match",
			file:        "anything.txt",
			pattern:     "created",
			wantTime:    time.Time{},
			wantErr:     true,
			errContains: "No date pattern matched file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTime, err := extract(tt.file, tt.pattern)

			if (err != nil) != tt.wantErr {
				t.Fatalf("\nError:   %v\nwantErr: %v", err, tt.wantErr)
			}

			if err != nil && tt.errContains != "" {
				if !containsString(err.Error(), tt.errContains) {
					t.Errorf("\nString: %q\nDoes not contain expected substring: %q", err.Error(), tt.errContains)
				}
				return
			}

			if !gotTime.Equal(tt.wantTime) {
				t.Errorf("\ngotTime:  %v\nwantTime: %v", gotTime, tt.wantTime)
			}
		})
	}
}

func containsString(str, substr string) bool {
	return len(str) >= len(substr) && func() bool {
		for i := 0; i <= len(str)-len(substr); i++ {
			if str[i:i+len(substr)] == substr {
				return true
			}
		}
		return false
	}()
}

func TestGenTargetName(t *testing.T) {
	tests := []struct {
		name      string
		file      *models.File
		pattern   string
		lower     bool
		strip     bool
		noDesc    bool
		dSep      string
		tSep      string
		wantDesc  string
		wantExt   string
		wantMatch bool
	}{
		// Santiy checks.
		{
			name: "sanity: sanitized format",
			file: &models.File{
				Source: "/tmp/2026-02-02T12.05.20.700 - beach vacation -- summer vacation beach.jpg",
				VTime:  parsedTime,
				Desc:   baseDesc,
				Ext:    ".jpg",
				Tags:   baseTags,
			},
			pattern:   "photo",
			lower:     false,
			strip:     false,
			dSep:      " - ",
			tSep:      " -- ",
			wantDesc:  baseDesc,
			wantExt:   ".jpg",
			wantMatch: true,
		},
		{
			name: "sanity: sanitized format [desc:' ', tag:' - ']",
			file: &models.File{
				Source: "/tmp/2026-02-02T12.05.20.700 beach vacation - summer vacation beach.jpg",
				VTime:  parsedTime,
				Desc:   baseDesc,
				Ext:    ".jpg",
				Tags:   baseTags,
			},
			pattern:   "photo",
			lower:     false,
			strip:     false,
			dSep:      " ",
			tSep:      " - ",
			wantDesc:  baseDesc,
			wantExt:   ".jpg",
			wantMatch: true,
		},
		{
			name: "sanity: no matches",
			file: &models.File{
				Source: "/tmp/not-a-date-file.jpg",
				VTime:  parsedTime,
				Desc:   "not-a-date-file",
				Ext:    ".jpg",
				Tags:   []string{},
			},
			pattern:   "photo",
			lower:     false,
			strip:     false,
			dSep:      " - ",
			tSep:      " -- ",
			wantDesc:  "not-a-date-file",
			wantExt:   ".jpg",
			wantMatch: false,
		},
		{
			name: "sanity: bare photo-ms",
			file: &models.File{
				Source: "/tmp/PXL_20260517_104536300.jpg",
				VTime:  parsedTime,
				Desc:   "PXL_20260517_104536300",
				Ext:    ".jpg",
				Tags:   []string{},
			},
			pattern:   "photo-ms",
			lower:     false,
			strip:     false,
			dSep:      " - ",
			tSep:      " -- ",
			wantDesc:  "PXL_20260517_104536300",
			wantExt:   ".jpg",
			wantMatch: true,
		},
		// Lower.
		{
			name: "lower: desc and ext",
			file: &models.File{
				Source: "/tmp/2026-02-02T12.05.20.700 Beach VACATION - summer vacation beach.JPG",
				VTime:  parsedTime,
				Desc:   baseDesc,
				Ext:    ".jpg",
				Tags:   baseTags,
			},
			pattern:   "photo",
			lower:     true,
			strip:     false,
			dSep:      " ",
			tSep:      " - ",
			wantDesc:  baseDesc,
			wantExt:   ".jpg",
			wantMatch: true,
		},
		// Strip.
		{
			name: "strip: bare photo-ms",
			file: &models.File{
				Source: "/tmp/PXL_20260517_104536300.jpg",
				VTime:  parsedTime,
				Desc:   "PXL_20260517_104536300",
				Ext:    ".jpg",
				Tags:   []string{},
			},
			pattern:   "photo-ms",
			lower:     false,
			strip:     true,
			dSep:      " - ",
			tSep:      " -- ",
			wantDesc:  "PXL_",
			wantExt:   ".jpg",
			wantMatch: true,
		},
		{
			name: "strip: bare photo-ms [prefix and suffix]",
			file: &models.File{
				Source: "/tmp/PXL_20260517_104536300~1.jpg",
				VTime:  parsedTime,
				Desc:   "PXL_20260517_104536300~1",
				Ext:    ".jpg",
				Tags:   []string{},
			},
			pattern:   "photo-ms",
			lower:     false,
			strip:     true,
			dSep:      " - ",
			tSep:      " -- ",
			wantDesc:  "PXL_~1",
			wantExt:   ".jpg",
			wantMatch: true,
		},
		{
			name: "strip: no match",
			file: &models.File{
				Source: "/tmp/PXL_20260517_104536300.jpg",
				VTime:  parsedTime,
				Desc:   "PXL_20260517_104536300",
				Ext:    ".jpg",
				Tags:   []string{},
			},
			pattern:   "signal",
			lower:     false,
			strip:     true,
			dSep:      " - ",
			tSep:      " -- ",
			wantDesc:  "PXL_20260517_104536300",
			wantExt:   ".jpg",
			wantMatch: false,
		},
		{
			name: "strip: bare photo-ms [prefix, suffix, lower]",
			file: &models.File{
				Source: "/tmp/PXL_20260517_104536300~1.jpg",
				VTime:  parsedTime,
				Desc:   "PXL_20260517_104536300~1",
				Ext:    ".jpg",
				Tags:   []string{},
			},
			pattern:   "photo-ms",
			lower:     true,
			strip:     true,
			dSep:      " - ",
			tSep:      " -- ",
			wantDesc:  "pxl_~1",
			wantExt:   ".jpg",
			wantMatch: true,
		},
		{
			name: "strip: bare photo-ms [prefix, suffix, lower, nodesc]",
			file: &models.File{
				Source: "/tmp/PXL_20260517_104536300~1.jpg",
				VTime:  parsedTime,
				Desc:   "",
				Ext:    ".jpg",
				Tags:   []string{},
			},
			pattern:   "photo-ms",
			lower:     true,
			strip:     true,
			noDesc:    true,
			dSep:      " - ",
			tSep:      " -- ",
			wantDesc:  "",
			wantExt:   ".jpg",
			wantMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseDir, err := filepath.Abs("/tmp")
			if err != nil {
				t.Fatalf("Failed to setup test base dir: %v", err)
			}

			tagsStr := strings.Join(tt.file.Tags, " ")
			expectedFileName := formattedTime + tt.dSep + tt.wantDesc + tt.tSep + tagsStr + tt.wantExt
			expectedTarget := filepath.Join(baseDir, expectedFileName)

			GenTargetName(tt.file, tt.pattern, tt.lower, tt.strip, tt.noDesc, tt.dSep, tt.tSep)

			if tt.file.Target != expectedTarget {
				t.Errorf("\nGot:  %s\nWant: %s", tt.file.Target, expectedTarget)
			}
		})
	}
}
