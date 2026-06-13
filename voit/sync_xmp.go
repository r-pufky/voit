package voit

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"trimmer.io/go-xmp/models/digikam"
	"trimmer.io/go-xmp/xmp"
)

// Sync tags from XMP sidecar overwriting existing tags.
func (files VoitFiles) SyncXMP(opts *Opts) {
	vCfg := opts.Voit()

	fileMap := make(FileMap)
	files.buildFileMap(vCfg, fileMap)
	fileMap.ParseSidecar(opts)
	files.ResolveCollisions(vCfg, opts.Verbose)
}

// Ingest and link related files. Valid links contain both file and sidecar.
func (files VoitFiles) buildFileMap(vCfg Config, fileMap map[string]*LinkedFiles) {
	for _, f := range files {
		f.Ingest(&vCfg)

		// Copy the original struct and use new reference for tags.
		f.Mark = f.Orig
		f.Mark.Tags.Items = slices.Clone(f.Orig.Tags.Items)

		// Track each base file name and link corresponding file, sidecar.
		link, exists := fileMap[f.File.Name]
		if !exists {
			link = &LinkedFiles{}
			fileMap[f.File.Name] = link
		}

		if strings.HasSuffix(strings.ToLower(f.File.Ext), ".xmp") {
			link.Sidecar = f
		} else {
			link.File = f
		}
	}
}

// Parse sidecar tags and update both file and sidecar with tags if both exist.
func (fMap FileMap) ParseSidecar(opts *Opts) {
	for _, pair := range fMap {
		if pair == nil || pair.File == nil || pair.Sidecar == nil {
			continue // Skip incomplete links.
		}

		var tags Tag

		doc, err := parseSidecar(pair.Sidecar.File.Source)
		if err != nil {
			fmt.Printf("Warning: Failed to parse %s: %v\n", pair.Sidecar.File.Source, err)
			continue
		}

		if model := digikam.FindModel(doc); model != nil && len(model.TagsList) > 0 {
			for _, tag := range model.TagsList {
				tags.SyncAdd(tag, opts)
			}

			pair.File.Matched = true
			pair.Sidecar.Matched = true
			pair.File.Mark.Tags = tags
			pair.Sidecar.Mark.Tags = tags
		}
	}
}

// Parse XMP sidecar from file.
func parseSidecar(path string) (*xmp.Document, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	xmpData, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	doc := xmp.NewDocument()
	err = xmp.Unmarshal(xmpData, doc)
	if err != nil {
		return nil, fmt.Errorf("error parsing XMP data: %w", err)
	}

	return doc, nil
}
