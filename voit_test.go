package main

import (
	"os"
	"testing"
	"time"
)

type FormatTestCase struct {
	test     string
	filename string
	pattern  string
	lower    bool
	strip    bool
	created  bool
	modified bool
	want     string
}

func runFormatTests(t *testing.T, tests []FormatTestCase) {
	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			got := FormatName(tt.filename, tt.pattern, tt.lower, tt.strip, tt.created, tt.modified)
			if got != tt.want {
				t.Errorf("\nInput: %s\nGot:   %s\nWant:  %s", tt.filename, got, tt.want)
			}
		})
	}
}

func TestMS(t *testing.T) {
	tests := []FormatTestCase{
		{
			test:     "ms bare",
			filename: "20231027_103005123.jpg",
			pattern:  "ms",
			lower:    false,
			strip:    false,
			want:     "2023-10-27T10.30.05.123 - 20231027_103005123.jpg",
		},
		{
			test:     "ms bare alternative separator",
			filename: "20231027-103005123.jpg",
			pattern:  "ms",
			lower:    false,
			strip:    false,
			want:     "2023-10-27T10.30.05.123 - 20231027-103005123.jpg",
		},
		{
			test:     "ms bare strip",
			filename: "20231027_103005123.jpg",
			pattern:  "ms",
			lower:    false,
			strip:    true,
			want:     "2023-10-27T10.30.05.123.jpg",
		},
		{
			test:     "ms leading",
			filename: "IMG_20231027_103005123.jpg",
			pattern:  "ms",
			lower:    false,
			strip:    false,
			want:     "2023-10-27T10.30.05.123 - IMG_20231027_103005123.jpg",
		},
		{
			test:     "ms leading lowercase",
			filename: "IMG_20231027_103005123.JPG",
			pattern:  "ms",
			lower:    true,
			strip:    false,
			want:     "2023-10-27T10.30.05.123 - img_20231027_103005123.jpg",
		},
		{
			test:     "ms leading lowercase strip",
			filename: "IMG_20231027_103005123.JPG",
			pattern:  "ms",
			lower:    true,
			strip:    true,
			want:     "2023-10-27T10.30.05.123.jpg",
		},
		{
			test:     "ms leading and trailing",
			filename: "PXL_20231027_103005123-1.jpg",
			pattern:  "ms",
			lower:    false,
			strip:    false,
			want:     "2023-10-27T10.30.05.123 - PXL_20231027_103005123-1.jpg",
		},
		{
			test:     "ms leading and trailing lowercase",
			filename: "PXL_20231027_103005123-1.jpg",
			pattern:  "ms",
			lower:    true,
			strip:    false,
			want:     "2023-10-27T10.30.05.123 - pxl_20231027_103005123-1.jpg",
		},
		{
			test:     "ms additional numerics",
			filename: "2343_20231027_103005123_34-2342-1.jpg",
			pattern:  "ms",
			lower:    true,
			strip:    false,
			want:     "2023-10-27T10.30.05.123 - 2343_20231027_103005123_34-2342-1.jpg",
		},
		{
			test:     "ms no match",
			filename: "2023-10-27-10-30-05-456.mp4",
			pattern:  "ms",
			lower:    false,
			strip:    false,
			want:     "2023-10-27-10-30-05-456.mp4",
		},
		{
			test:     "ms idempotency",
			filename: "2023-10-27T10.30.05.123 - IMG_20231027_103005123.jpg",
			pattern:  "ms",
			lower:    false,
			strip:    false,
			want:     "2023-10-27T10.30.05.123 - IMG_20231027_103005123.jpg",
		},
		{
			test:     "ms idempotency lowercase",
			filename: "2023-10-27T10.30.05.123 - IMG_20231027_103005123.jpg",
			pattern:  "ms",
			lower:    true,
			strip:    false,
			want:     "2023-10-27T10.30.05.123 - IMG_20231027_103005123.jpg",
		},
	}
	runFormatTests(t, tests)
}

func TestMNS(t *testing.T) {
	tests := []FormatTestCase{
		{
			test:     "mns bare",
			filename: "2023-10-27-10-30-05-456.mp4",
			pattern:  "mns",
			lower:    false,
			strip:    false,
			want:     "2023-10-27T10.30.05.456 - 2023-10-27-10-30-05-456.mp4",
		},
		{
			test:     "mns bare alternative separator",
			filename: "2023_10_27_10_30_05_456.mp4",
			pattern:  "mns",
			lower:    false,
			strip:    false,
			want:     "2023-10-27T10.30.05.456 - 2023_10_27_10_30_05_456.mp4",
		},
		{
			test:     "mns bare strip",
			filename: "2023-10-27-10-30-05-456.mp4",
			pattern:  "mns",
			lower:    false,
			strip:    true,
			want:     "2023-10-27T10.30.05.456.mp4",
		},
		{
			test:     "mns leading",
			filename: "test vid 2023_10_27_10_30_05_456.mp4",
			pattern:  "mns",
			lower:    false,
			strip:    false,
			want:     "2023-10-27T10.30.05.456 - test vid 2023_10_27_10_30_05_456.mp4",
		},
		{
			test:     "mns leading lowercase",
			filename: "TEST VID 2023_10_27_10_30_05_456.MP4",
			pattern:  "mns",
			lower:    true,
			strip:    false,
			want:     "2023-10-27T10.30.05.456 - test vid 2023_10_27_10_30_05_456.mp4",
		},
		{
			test:     "mns leading lowercase strip",
			filename: "TEST VID 2023_10_27_10_30_05_456.MP4",
			pattern:  "mns",
			lower:    true,
			strip:    true,
			want:     "2023-10-27T10.30.05.456.mp4",
		},
		{
			test:     "mns leading and trailing",
			filename: "test vid 2023_10_27_10_30_05_456-1.mp4",
			pattern:  "mns",
			lower:    false,
			strip:    false,
			want:     "2023-10-27T10.30.05.456 - test vid 2023_10_27_10_30_05_456-1.mp4",
		},
		{
			test:     "mns leading and trailing lowercase",
			filename: "TEST VID 2023_10_27_10_30_05_456-TEST.mp4",
			pattern:  "mns",
			lower:    true,
			strip:    false,
			want:     "2023-10-27T10.30.05.456 - test vid 2023_10_27_10_30_05_456-test.mp4",
		},
		{
			test:     "mns additional numerics",
			filename: "2343_2023_10_27_10_30_05_456_34-2342-1.mp4",
			pattern:  "mns",
			lower:    true,
			strip:    false,
			want:     "2023-10-27T10.30.05.456 - 2343_2023_10_27_10_30_05_456_34-2342-1.mp4",
		},
		{
			test:     "mns no match",
			filename: "20231027_103005123.jpg",
			pattern:  "mns",
			lower:    false,
			strip:    false,
			want:     "20231027_103005123.jpg",
		},
		{
			test:     "mns idempotency",
			filename: "2023-10-27T10.30.05.456 - TEST VID 2023_10_27_10_30_05_456-TEST.mp4",
			pattern:  "mns",
			lower:    false,
			strip:    false,
			want:     "2023-10-27T10.30.05.456 - TEST VID 2023_10_27_10_30_05_456-TEST.mp4",
		},
		{
			test:     "mns idempotency lowercase",
			filename: "2023-10-27T10.30.05.456 - TEST VID 2023_10_27_10_30_05_456-TEST.mp4",
			pattern:  "mns",
			lower:    true,
			strip:    false,
			want:     "2023-10-27T10.30.05.456 - TEST VID 2023_10_27_10_30_05_456-TEST.mp4",
		},
	}
	runFormatTests(t, tests)
}

func TestMFS(t *testing.T) {
	tests := []FormatTestCase{
		{
			test:     "mfs bare",
			filename: "20231027103005123.jpg",
			pattern:  "mfs",
			lower:    false,
			strip:    false,
			want:     "2023-10-27T10.30.05.123 - 20231027103005123.jpg",
		},
		{
			test:     "mfs bare strip",
			filename: "20231027103005123.jpg",
			pattern:  "mfs",
			lower:    false,
			strip:    true,
			want:     "2023-10-27T10.30.05.123.jpg",
		},
		{
			test:     "mfs leading",
			filename: "IMG_20231027103005123.jpg",
			pattern:  "mfs",
			lower:    false,
			strip:    false,
			want:     "2023-10-27T10.30.05.123 - IMG_20231027103005123.jpg",
		},
		{
			test:     "mfs leading lowercase",
			filename: "IMG_20231027103005123.JPG",
			pattern:  "mfs",
			lower:    true,
			strip:    false,
			want:     "2023-10-27T10.30.05.123 - img_20231027103005123.jpg",
		},
		{
			test:     "mfs leading lowercase strip",
			filename: "IMG_20231027103005123.JPG",
			pattern:  "mfs",
			lower:    true,
			strip:    true,
			want:     "2023-10-27T10.30.05.123.jpg",
		},
		{
			test:     "mfs leading and trailing",
			filename: "PXL_20231027103005123-1.jpg",
			pattern:  "mfs",
			lower:    false,
			strip:    false,
			want:     "2023-10-27T10.30.05.123 - PXL_20231027103005123-1.jpg",
		},
		{
			test:     "mfs leading and trailing lowercase",
			filename: "PXL_20231027103005123-1.jpg",
			pattern:  "mfs",
			lower:    true,
			strip:    false,
			want:     "2023-10-27T10.30.05.123 - pxl_20231027103005123-1.jpg",
		},
		{
			test:     "mfs additional numerics",
			filename: "2343_20231027103005123_34-2342-1.jpg",
			pattern:  "mfs",
			lower:    true,
			strip:    false,
			want:     "2023-10-27T10.30.05.123 - 2343_20231027103005123_34-2342-1.jpg",
		},
		{
			test:     "mfs no match",
			filename: "2023-10-27-10-30-05-456.mp4",
			pattern:  "mfs",
			lower:    false,
			strip:    false,
			want:     "2023-10-27-10-30-05-456.mp4",
		},
		{
			test:     "mfs idempotency",
			filename: "2023-10-27T10.30.05.123 - PXL_20231027103005123-1.jpg",
			pattern:  "mfs",
			lower:    false,
			strip:    false,
			want:     "2023-10-27T10.30.05.123 - PXL_20231027103005123-1.jpg",
		},
		{
			test:     "mfs idempotency lowercase",
			filename: "2023-10-27T10.30.05.123 - PXL_20231027103005123-1.jpg",
			pattern:  "mfs",
			lower:    true,
			strip:    false,
			want:     "2023-10-27T10.30.05.123 - PXL_20231027103005123-1.jpg",
		},
	}
	runFormatTests(t, tests)
}

func TestS(t *testing.T) {
	tests := []FormatTestCase{
		{
			test:     "s bare",
			filename: "20231027_103005.jpg",
			pattern:  "s",
			lower:    false,
			strip:    false,
			want:     "2023-10-27T10.30.05.000 - 20231027_103005.jpg",
		},
		{
			test:     "s bare alternative separator",
			filename: "20231027-103005.jpg",
			pattern:  "s",
			lower:    false,
			strip:    false,
			want:     "2023-10-27T10.30.05.000 - 20231027-103005.jpg",
		},
		{
			test:     "s bare strip",
			filename: "20231027_103005.jpg",
			pattern:  "s",
			lower:    false,
			strip:    true,
			want:     "2023-10-27T10.30.05.000.jpg",
		},
		{
			test:     "s leading",
			filename: "IMG_20231027_103005.jpg",
			pattern:  "s",
			lower:    false,
			strip:    false,
			want:     "2023-10-27T10.30.05.000 - IMG_20231027_103005.jpg",
		},
		{
			test:     "s leading lowercase",
			filename: "IMG_20231027_103005.JPG",
			pattern:  "s",
			lower:    true,
			strip:    false,
			want:     "2023-10-27T10.30.05.000 - img_20231027_103005.jpg",
		},
		{
			test:     "s leading lowercase strip",
			filename: "IMG_20231027_103005.JPG",
			pattern:  "s",
			lower:    true,
			strip:    true,
			want:     "2023-10-27T10.30.05.000.jpg",
		},
		{
			test:     "s leading and trailing",
			filename: "PXL_20231027_103005-1.jpg",
			pattern:  "s",
			lower:    false,
			strip:    false,
			want:     "2023-10-27T10.30.05.000 - PXL_20231027_103005-1.jpg",
		},
		{
			test:     "s leading and trailing lowercase",
			filename: "PXL_20231027_103005-1.jpg",
			pattern:  "s",
			lower:    true,
			strip:    false,
			want:     "2023-10-27T10.30.05.000 - pxl_20231027_103005-1.jpg",
		},
		{
			test:     "s additional numerics",
			filename: "2343_20231027_103005_34-2342-1.jpg",
			pattern:  "s",
			lower:    true,
			strip:    false,
			want:     "2023-10-27T10.30.05.000 - 2343_20231027_103005_34-2342-1.jpg",
		},
		{
			test:     "s no match",
			filename: "2023-10-27-10-30-05-456.mp4",
			pattern:  "s",
			lower:    false,
			strip:    false,
			want:     "2023-10-27-10-30-05-456.mp4",
		},
		{
			test:     "s idempotency",
			filename: "2023-10-27T10.30.05.000 - PXL_20231027_103005-1.jpg",
			pattern:  "s",
			lower:    false,
			strip:    false,
			want:     "2023-10-27T10.30.05.000 - PXL_20231027_103005-1.jpg",
		},
		{
			test:     "s idempotency lowercase",
			filename: "2023-10-27T10.30.05.000 - PXL_20231027_103005-1.jpg",
			pattern:  "s",
			lower:    true,
			strip:    false,
			want:     "2023-10-27T10.30.05.000 - PXL_20231027_103005-1.jpg",
		},
	}
	runFormatTests(t, tests)
}

func TestNS(t *testing.T) {
	tests := []FormatTestCase{
		{
			test:     "ns bare",
			filename: "2023-10-27-10-30-05.mp4",
			pattern:  "ns",
			lower:    false,
			strip:    false,
			want:     "2023-10-27T10.30.05.000 - 2023-10-27-10-30-05.mp4",
		},
		{
			test:     "ns bare strip",
			filename: "2023-10-27-10-30-05.mp4",
			pattern:  "ns",
			lower:    false,
			strip:    true,
			want:     "2023-10-27T10.30.05.000.mp4",
		},
		{
			test:     "ns bare alternative separator",
			filename: "2023_10_27_10_30_05.mp4",
			pattern:  "ns",
			lower:    false,
			strip:    false,
			want:     "2023-10-27T10.30.05.000 - 2023_10_27_10_30_05.mp4",
		},
		{
			test:     "ns leading",
			filename: "test vid 2023_10_27_10_30_05.mp4",
			pattern:  "ns",
			lower:    false,
			strip:    false,
			want:     "2023-10-27T10.30.05.000 - test vid 2023_10_27_10_30_05.mp4",
		},
		{
			test:     "ns leading lowercase",
			filename: "TEST VID 2023_10_27_10_30_05.MP4",
			pattern:  "ns",
			lower:    true,
			strip:    false,
			want:     "2023-10-27T10.30.05.000 - test vid 2023_10_27_10_30_05.mp4",
		},
		{
			test:     "ns leading lowercase strip",
			filename: "TEST VID 2023_10_27_10_30_05.MP4",
			pattern:  "ns",
			lower:    true,
			strip:    true,
			want:     "2023-10-27T10.30.05.000.mp4",
		},
		{
			test:     "ns leading and trailing",
			filename: "test vid 2023_10_27_10_30_05-1.mp4",
			pattern:  "ns",
			lower:    false,
			strip:    false,
			want:     "2023-10-27T10.30.05.000 - test vid 2023_10_27_10_30_05-1.mp4",
		},
		{
			test:     "ns leading and trailing lowercase",
			filename: "TEST VID 2023_10_27_10_30_05-1.mp4",
			pattern:  "ns",
			lower:    true,
			strip:    false,
			want:     "2023-10-27T10.30.05.000 - test vid 2023_10_27_10_30_05-1.mp4",
		},
		{
			test:     "ns additional numerics",
			filename: "2343_2023_10_27_10_30_05_34-2342-1.mp4",
			pattern:  "ns",
			lower:    true,
			strip:    false,
			want:     "2023-10-27T10.30.05.000 - 2343_2023_10_27_10_30_05_34-2342-1.mp4",
		},
		{
			test:     "ns no match",
			filename: "20231027_103005123.jpg",
			pattern:  "ns",
			lower:    false,
			strip:    false,
			want:     "20231027_103005123.jpg",
		},
		{
			test:     "ns idempotency",
			filename: "2023-10-27T10.30.05.000 - TEST VID 2023_10_27_10_30_05-1.mp4",
			pattern:  "ns",
			lower:    false,
			strip:    false,
			want:     "2023-10-27T10.30.05.000 - TEST VID 2023_10_27_10_30_05-1.mp4",
		},
		{
			test:     "ns idempotency lowercase",
			filename: "2023-10-27T10.30.05.000 - TEST VID 2023_10_27_10_30_05-1.mp4",
			pattern:  "ns",
			lower:    true,
			strip:    false,
			want:     "2023-10-27T10.30.05.000 - TEST VID 2023_10_27_10_30_05-1.mp4",
		},
	}
	runFormatTests(t, tests)
}

func TestFS(t *testing.T) {
	tests := []FormatTestCase{
		{
			test:     "fs bare",
			filename: "20231027103005.jpg",
			pattern:  "fs",
			lower:    false,
			strip:    false,
			want:     "2023-10-27T10.30.05.000 - 20231027103005.jpg",
		},
		{
			test:     "fs bare strip",
			filename: "20231027103005.jpg",
			pattern:  "fs",
			lower:    false,
			strip:    true,
			want:     "2023-10-27T10.30.05.000.jpg",
		},
		{
			test:     "fs leading",
			filename: "IMG_20231027103005.jpg",
			pattern:  "fs",
			lower:    false,
			strip:    false,
			want:     "2023-10-27T10.30.05.000 - IMG_20231027103005.jpg",
		},
		{
			test:     "fs leading lowercase",
			filename: "IMG_20231027103005.JPG",
			pattern:  "fs",
			lower:    true,
			strip:    false,
			want:     "2023-10-27T10.30.05.000 - img_20231027103005.jpg",
		},
		{
			test:     "fs leading lowercase strip",
			filename: "IMG_20231027103005.JPG",
			pattern:  "fs",
			lower:    true,
			strip:    true,
			want:     "2023-10-27T10.30.05.000.jpg",
		},
		{
			test:     "fs leading and trailing",
			filename: "PXL_20231027103005123-1.jpg",
			pattern:  "fs",
			lower:    false,
			strip:    false,
			want:     "2023-10-27T10.30.05.000 - PXL_20231027103005123-1.jpg",
		},
		{
			test:     "fs leading and trailing lowercase",
			filename: "PXL_20231027103005123-1.jpg",
			pattern:  "fs",
			lower:    true,
			strip:    false,
			want:     "2023-10-27T10.30.05.000 - pxl_20231027103005123-1.jpg",
		},
		{
			test:     "fs additional numerics",
			filename: "2343_20231027103005123_34-2342-1.jpg",
			pattern:  "fs",
			lower:    true,
			strip:    false,
			want:     "2023-10-27T10.30.05.000 - 2343_20231027103005123_34-2342-1.jpg",
		},
		{
			test:     "fs no match",
			filename: "2023-10-27-10-30-05-456.mp4",
			pattern:  "fs",
			lower:    false,
			strip:    false,
			want:     "2023-10-27-10-30-05-456.mp4",
		},
		{
			test:     "fs idempotency",
			filename: "2023-10-27T10.30.05.000 - PXL_20231027103005123-1.jpg",
			pattern:  "fs",
			lower:    false,
			strip:    false,
			want:     "2023-10-27T10.30.05.000 - PXL_20231027103005123-1.jpg",
		},
		{
			test:     "fs idempotency lowercase",
			filename: "2023-10-27T10.30.05.000 - PXL_20231027103005123-1.jpg",
			pattern:  "fs",
			lower:    true,
			strip:    false,
			want:     "2023-10-27T10.30.05.000 - PXL_20231027103005123-1.jpg",
		},
	}
	runFormatTests(t, tests)
}

func TestW(t *testing.T) {
	tests := []FormatTestCase{
		{
			test:     "w bare",
			filename: "13423083387000000.jpg",
			pattern:  "w",
			lower:    false,
			strip:    false,
			want:     "2026-05-12T18.16.27.000 - 13423083387000000.jpg",
		},
		{
			test:     "w bare strip",
			filename: "13423083387000000.jpg",
			pattern:  "w",
			lower:    false,
			strip:    true,
			want:     "2026-05-12T18.16.27.000.jpg",
		},
		{
			test:     "w leading",
			filename: "IMG_13423083387000000.jpg",
			pattern:  "w",
			lower:    false,
			strip:    false,
			want:     "2026-05-12T18.16.27.000 - IMG_13423083387000000.jpg",
		},
		{
			test:     "w leading lowercase",
			filename: "IMG_13423083387000000.jpg",
			pattern:  "w",
			lower:    true,
			strip:    false,
			want:     "2026-05-12T18.16.27.000 - img_13423083387000000.jpg",
		},
		{
			test:     "w leading lowercase strip",
			filename: "IMG_13423083387000000.jpg",
			pattern:  "w",
			lower:    true,
			strip:    true,
			want:     "2026-05-12T18.16.27.000.jpg",
		},
		{
			test:     "w leading and trailing",
			filename: "PXL_13423083387000000-1.jpg",
			pattern:  "w",
			lower:    false,
			strip:    false,
			want:     "2026-05-12T18.16.27.000 - PXL_13423083387000000-1.jpg",
		},
		{
			test:     "w leading and trailing lowercase",
			filename: "PXL_13423083387000000-1.jpg",
			pattern:  "w",
			lower:    true,
			strip:    false,
			want:     "2026-05-12T18.16.27.000 - pxl_13423083387000000-1.jpg",
		},
		{
			test:     "w additional numerics",
			filename: "2343_13423083387000000-2342-1.jpg",
			pattern:  "w",
			lower:    true,
			strip:    false,
			want:     "2026-05-12T18.16.27.000 - 2343_13423083387000000-2342-1.jpg",
		},
		{
			test:     "w no match",
			filename: "2023-10-27-10-30-05-456.mp4",
			pattern:  "w",
			lower:    false,
			strip:    false,
			want:     "2023-10-27-10-30-05-456.mp4",
		},
		// No idempotency as matching source and target patterns will always expand
		// filename.
	}
	runFormatTests(t, tests)
}

func TestV(t *testing.T) {
	tests := []FormatTestCase{
		{
			test:     "v bare",
			filename: "2023-10-27T10.30.05.000.jpg",
			pattern:  "v",
			lower:    false,
			strip:    false,
			want:     "2023-10-27T10.30.05.000 - 2023-10-27T10.30.05.000.jpg",
		},
		{
			test:     "v bare strip",
			filename: "2023-10-27T10.30.05.000.jpg",
			pattern:  "v",
			lower:    false,
			strip:    true,
			want:     "2023-10-27T10.30.05.000.jpg",
		},
		{
			test:     "v leading",
			filename: "IMG_2023-10-27T10.30.05.000.jpg",
			pattern:  "v",
			lower:    false,
			strip:    false,
			want:     "2023-10-27T10.30.05.000 - IMG_2023-10-27T10.30.05.000.jpg",
		},
		{
			test:     "v leading lowercase",
			filename: "IMG_2023-10-27T10.30.05.000.jpg",
			pattern:  "v",
			lower:    true,
			strip:    false,
			want:     "2023-10-27T10.30.05.000 - img_2023-10-27t10.30.05.000.jpg",
		},
		{
			test:     "v leading lowercase strip",
			filename: "IMG_2023-10-27T10.30.05.000.jpg",
			pattern:  "v",
			lower:    true,
			strip:    true,
			want:     "2023-10-27T10.30.05.000.jpg",
		},
		{
			test:     "v leading and trailing",
			filename: "PXL_2023-10-27T10.30.05.000-1.jpg",
			pattern:  "v",
			lower:    false,
			strip:    false,
			want:     "2023-10-27T10.30.05.000 - PXL_2023-10-27T10.30.05.000-1.jpg",
		},
		{
			test:     "v leading and trailing lowercase",
			filename: "PXL_2023-10-27T10.30.05.000-1.jpg",
			pattern:  "v",
			lower:    true,
			strip:    false,
			want:     "2023-10-27T10.30.05.000 - pxl_2023-10-27t10.30.05.000-1.jpg",
		},
		{
			test:     "v additional numerics",
			filename: "2343_2023-10-27T10.30.05.000-2342-1.jpg",
			pattern:  "v",
			lower:    true,
			strip:    false,
			want:     "2023-10-27T10.30.05.000 - 2343_2023-10-27t10.30.05.000-2342-1.jpg",
		},
		{
			test:     "v no match",
			filename: "2023-10-27-10-30-05-456.mp4",
			pattern:  "v",
			lower:    false,
			strip:    false,
			want:     "2023-10-27-10-30-05-456.mp4",
		},
		// No idempotency as matching source and target patterns will always expand
		// filename.
	}
	runFormatTests(t, tests)
}

func TestParseFileTime(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	info, _ := os.Stat(tmpFile.Name())
	expectedModTime := info.ModTime().UTC()

	tests := []struct {
		name     string
		file     string
		created  bool
		modified bool
		wantErr  bool
	}{
		{
			name:     "File does not exist",
			file:     "non_existent_file.txt",
			created:  true,
			modified: true,
			wantErr:  true,
		},
		{
			name:     "Both flags false returns empty time",
			file:     tmpFile.Name(),
			created:  false,
			modified: false,
			wantErr:  false,
		},
		{
			name:     "Modified time",
			file:     tmpFile.Name(),
			created:  false,
			modified: true,
			wantErr:  false,
		},
		{
			name:     "Created time",
			file:     tmpFile.Name(),
			created:  true,
			modified: false,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFileTime(tt.file, tt.created, tt.modified)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseFileTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if !tt.created && !tt.modified {
					if !got.IsZero() {
						t.Errorf("Expected zero time when both flags are false, got %v", got)
					}
				} else {
					// Ensure UTC.
					if got.Location() != time.UTC {
						t.Errorf("Expected UTC location, got %v", got.Location())
					}

					// cTime matches mTime for a new file.
					if tt.modified && !tt.created {
						if got.Unix() != expectedModTime.Unix() {
							t.Errorf("Time mismatch. Got %v, want %v", got, expectedModTime)
						}
					}
				}
			}
		})
	}
}
