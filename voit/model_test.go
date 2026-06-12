package voit

import (
	"testing"
)

func TestNewConfig(t *testing.T) {
	cfg := NewConfig()

	if cfg.Format != DefaultVFormat {
		t.Errorf("wantBool Format to be %q, got %q", DefaultVFormat, cfg.Format)
	}

	if cfg.Pattern != DefaultPattern {
		t.Errorf("wantBool Pattern to be %q, got %q", DefaultPattern, cfg.Pattern)
	}

	if cfg.SSep != DefaultSpanSep {
		t.Errorf("wantBool SSep to be %q, got %q", DefaultSpanSep, cfg.SSep)
	}

	if cfg.DSep != DefaultDescSep {
		t.Errorf("wantBool DSep to be %q, got %q", DefaultDescSep, cfg.DSep)
	}

	if cfg.TSep != DefaultTagsSep {
		t.Errorf("wantBool TSep to be %q, got %q", DefaultTagsSep, cfg.TSep)
	}

	if cfg.Lower != false {
		t.Errorf("wantBool Lower to be false, got %t", cfg.Lower)
	}
}
