package voit

import "time"

// Ingest Voit.Orig fields from Voit.File. Undefined options use
// DefaultDescSep, DefaultTagsSep, DefaultSpanSep, DefaultVFormat,
// DefaultPattern.
func (f *Voit) Ingest(c *Config) {
	var err error

	tIdx := f.Orig.Tags.Chomp(f.File.Name, c.TSep)
	dIdx := f.Orig.Desc.Chomp(f.File.Name, c.DSep, c.TSep)
	if dIdx > -1 {
		f.Orig.VTime.Chomp(f.File.Name[:dIdx], c.SSep)
	} else if tIdx > -1 {
		f.Orig.VTime.Chomp(f.File.Name[:tIdx], c.SSep)
	} else {
		f.Orig.VTime.Chomp(f.File.Name, c.SSep)
	}

	switch c.Pattern {
	case "created":
		f.Orig.PTime.Time = f.File.CTime
	case "modified":
		f.Orig.PTime.Time = f.File.MTime
	case "set":
		f.Orig.PTime.Time, err = Extract(c.Set, "voit")
		if err != nil {
			f.Orig.PTime.Time = time.Time{}
		}
	default:
		f.Orig.PTime.Time, err = Extract(f.File.Name, c.Pattern)
		if err != nil {
			f.Orig.PTime.Time = time.Time{}
		}
	}
}
