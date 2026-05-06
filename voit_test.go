package main

import (
	"testing"
)

type FormatTestCase struct {
	test     string
	filename string
	pattern  string
	lower    bool
	want     string
}

func runFormatTests(t *testing.T, tests []FormatTestCase) {
	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			got := FormatName(tt.filename, tt.pattern, tt.lower)
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
			want:     "2023-10-27T10.30.05.123 - 20231027_103005123.jpg",
		},
		{
			test:     "ms bare alternative separator",
			filename: "20231027-103005123.jpg",
			pattern:  "ms",
			lower:    false,
			want:     "2023-10-27T10.30.05.123 - 20231027-103005123.jpg",
		},
		{
			test:     "ms leading",
			filename: "IMG_20231027_103005123.jpg",
			pattern:  "ms",
			lower:    false,
			want:     "2023-10-27T10.30.05.123 - IMG_20231027_103005123.jpg",
		},
		{
			test:     "ms leading lowercase",
			filename: "IMG_20231027_103005123.jpg",
			pattern:  "ms",
			lower:    true,
			want:     "2023-10-27T10.30.05.123 - img_20231027_103005123.jpg",
		},
		{
			test:     "ms leading and trailing",
			filename: "PXL_20231027_103005123-1.jpg",
			pattern:  "ms",
			lower:    false,
			want:     "2023-10-27T10.30.05.123 - PXL_20231027_103005123-1.jpg",
		},
		{
			test:     "ms leading and trailing lowercase",
			filename: "PXL_20231027_103005123-1.jpg",
			pattern:  "ms",
			lower:    true,
			want:     "2023-10-27T10.30.05.123 - pxl_20231027_103005123-1.jpg",
		},
		{
			test:     "ms additional numerics",
			filename: "2343_20231027_103005123_34-2342-1.jpg",
			pattern:  "ms",
			lower:    true,
			want:     "2023-10-27T10.30.05.123 - 2343_20231027_103005123_34-2342-1.jpg",
		},
		{
			test:     "ms no match",
			filename: "2023-10-27-10-30-05-456.mp4",
			pattern:  "ms",
			lower:    false,
			want:     "2023-10-27-10-30-05-456.mp4",
		},
		{
			test:     "ms idempotency",
			filename: "2023-10-27T10.30.05.123 - IMG_20231027_103005123.jpg",
			pattern:  "ms",
			lower:    false,
			want:     "2023-10-27T10.30.05.123 - IMG_20231027_103005123.jpg",
		},
		{
			test:     "ms idempotency lowercase",
			filename: "2023-10-27T10.30.05.123 - IMG_20231027_103005123.jpg",
			pattern:  "ms",
			lower:    true,
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
			want:     "2023-10-27T10.30.05.456 - 2023-10-27-10-30-05-456.mp4",
		},
		{
			test:     "mns bare alternative separator",
			filename: "2023_10_27_10_30_05_456.mp4",
			pattern:  "mns",
			lower:    false,
			want:     "2023-10-27T10.30.05.456 - 2023_10_27_10_30_05_456.mp4",
		},
		{
			test:     "mns leading",
			filename: "test vid 2023_10_27_10_30_05_456.mp4",
			pattern:  "mns",
			lower:    false,
			want:     "2023-10-27T10.30.05.456 - test vid 2023_10_27_10_30_05_456.mp4",
		},
		{
			test:     "mns leading lowercase",
			filename: "TEST VID 2023_10_27_10_30_05_456.mp4",
			pattern:  "mns",
			lower:    true,
			want:     "2023-10-27T10.30.05.456 - test vid 2023_10_27_10_30_05_456.mp4",
		},
		{
			test:     "mns leading and trailing",
			filename: "test vid 2023_10_27_10_30_05_456-1.mp4",
			pattern:  "mns",
			lower:    false,
			want:     "2023-10-27T10.30.05.456 - test vid 2023_10_27_10_30_05_456-1.mp4",
		},
		{
			test:     "mns leading and trailing lowercase",
			filename: "TEST VID 2023_10_27_10_30_05_456-TEST.mp4",
			pattern:  "mns",
			lower:    true,
			want:     "2023-10-27T10.30.05.456 - test vid 2023_10_27_10_30_05_456-test.mp4",
		},
		{
			test:     "mns additional numerics",
			filename: "2343_2023_10_27_10_30_05_456_34-2342-1.mp4",
			pattern:  "mns",
			lower:    true,
			want:     "2023-10-27T10.30.05.456 - 2343_2023_10_27_10_30_05_456_34-2342-1.mp4",
		},
		{
			test:     "mns no match",
			filename: "20231027_103005123.jpg",
			pattern:  "mns",
			lower:    false,
			want:     "20231027_103005123.jpg",
		},
		{
			test:     "mns idempotency",
			filename: "2023-10-27T10.30.05.456 - TEST VID 2023_10_27_10_30_05_456-TEST.mp4",
			pattern:  "mns",
			lower:    false,
			want:     "2023-10-27T10.30.05.456 - TEST VID 2023_10_27_10_30_05_456-TEST.mp4",
		},
		{
			test:     "mns idempotency lowercase",
			filename: "2023-10-27T10.30.05.456 - TEST VID 2023_10_27_10_30_05_456-TEST.mp4",
			pattern:  "mns",
			lower:    true,
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
			want:     "2023-10-27T10.30.05.123 - 20231027103005123.jpg",
		},
		{
			test:     "mfs leading",
			filename: "IMG_20231027103005123.jpg",
			pattern:  "mfs",
			lower:    false,
			want:     "2023-10-27T10.30.05.123 - IMG_20231027103005123.jpg",
		},
		{
			test:     "mfs leading lowercase",
			filename: "IMG_20231027103005123.jpg",
			pattern:  "mfs",
			lower:    true,
			want:     "2023-10-27T10.30.05.123 - img_20231027103005123.jpg",
		},
		{
			test:     "mfs leading and trailing",
			filename: "PXL_20231027103005123-1.jpg",
			pattern:  "mfs",
			lower:    false,
			want:     "2023-10-27T10.30.05.123 - PXL_20231027103005123-1.jpg",
		},
		{
			test:     "mfs leading and trailing lowercase",
			filename: "PXL_20231027103005123-1.jpg",
			pattern:  "mfs",
			lower:    true,
			want:     "2023-10-27T10.30.05.123 - pxl_20231027103005123-1.jpg",
		},
		{
			test:     "mfs additional numerics",
			filename: "2343_20231027103005123_34-2342-1.jpg",
			pattern:  "mfs",
			lower:    true,
			want:     "2023-10-27T10.30.05.123 - 2343_20231027103005123_34-2342-1.jpg",
		},
		{
			test:     "mfs no match",
			filename: "2023-10-27-10-30-05-456.mp4",
			pattern:  "mfs",
			lower:    false,
			want:     "2023-10-27-10-30-05-456.mp4",
		},
		{
			test:     "mfs idempotency",
			filename: "2023-10-27T10.30.05.123 - PXL_20231027103005123-1.jpg",
			pattern:  "mfs",
			lower:    false,
			want:     "2023-10-27T10.30.05.123 - PXL_20231027103005123-1.jpg",
		},
		{
			test:     "mfs idempotency lowercase",
			filename: "2023-10-27T10.30.05.123 - PXL_20231027103005123-1.jpg",
			pattern:  "mfs",
			lower:    true,
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
			want:     "2023-10-27T10.30.05.000 - 20231027_103005.jpg",
		},
		{
			test:     "s bare alternative separator",
			filename: "20231027-103005.jpg",
			pattern:  "s",
			lower:    false,
			want:     "2023-10-27T10.30.05.000 - 20231027-103005.jpg",
		},
		{
			test:     "s leading",
			filename: "IMG_20231027_103005.jpg",
			pattern:  "s",
			lower:    false,
			want:     "2023-10-27T10.30.05.000 - IMG_20231027_103005.jpg",
		},
		{
			test:     "s leading lowercase",
			filename: "IMG_20231027_103005.jpg",
			pattern:  "s",
			lower:    true,
			want:     "2023-10-27T10.30.05.000 - img_20231027_103005.jpg",
		},
		{
			test:     "s leading and trailing",
			filename: "PXL_20231027_103005-1.jpg",
			pattern:  "s",
			lower:    false,
			want:     "2023-10-27T10.30.05.000 - PXL_20231027_103005-1.jpg",
		},
		{
			test:     "s leading and trailing lowercase",
			filename: "PXL_20231027_103005-1.jpg",
			pattern:  "s",
			lower:    true,
			want:     "2023-10-27T10.30.05.000 - pxl_20231027_103005-1.jpg",
		},
		{
			test:     "s additional numerics",
			filename: "2343_20231027_103005_34-2342-1.jpg",
			pattern:  "s",
			lower:    true,
			want:     "2023-10-27T10.30.05.000 - 2343_20231027_103005_34-2342-1.jpg",
		},
		{
			test:     "s no match",
			filename: "2023-10-27-10-30-05-456.mp4",
			pattern:  "s",
			lower:    false,
			want:     "2023-10-27-10-30-05-456.mp4",
		},
		{
			test:     "s idempotency",
			filename: "2023-10-27T10.30.05.000 - PXL_20231027_103005-1.jpg",
			pattern:  "s",
			lower:    false,
			want:     "2023-10-27T10.30.05.000 - PXL_20231027_103005-1.jpg",
		},
		{
			test:     "s idempotency lowercase",
			filename: "2023-10-27T10.30.05.000 - PXL_20231027_103005-1.jpg",
			pattern:  "s",
			lower:    true,
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
			want:     "2023-10-27T10.30.05.000 - 2023-10-27-10-30-05.mp4",
		},
		{
			test:     "ns bare alternative separator",
			filename: "2023_10_27_10_30_05.mp4",
			pattern:  "ns",
			lower:    false,
			want:     "2023-10-27T10.30.05.000 - 2023_10_27_10_30_05.mp4",
		},
		{
			test:     "ns leading",
			filename: "test vid 2023_10_27_10_30_05.mp4",
			pattern:  "ns",
			lower:    false,
			want:     "2023-10-27T10.30.05.000 - test vid 2023_10_27_10_30_05.mp4",
		},
		{
			test:     "ns leading lowercase",
			filename: "TEST VID 2023_10_27_10_30_05.mp4",
			pattern:  "ns",
			lower:    true,
			want:     "2023-10-27T10.30.05.000 - test vid 2023_10_27_10_30_05.mp4",
		},
		{
			test:     "ns leading and trailing",
			filename: "test vid 2023_10_27_10_30_05-1.mp4",
			pattern:  "ns",
			lower:    false,
			want:     "2023-10-27T10.30.05.000 - test vid 2023_10_27_10_30_05-1.mp4",
		},
		{
			test:     "ns leading and trailing lowercase",
			filename: "TEST VID 2023_10_27_10_30_05-1.mp4",
			pattern:  "ns",
			lower:    true,
			want:     "2023-10-27T10.30.05.000 - test vid 2023_10_27_10_30_05-1.mp4",
		},
		{
			test:     "ns additional numerics",
			filename: "2343_2023_10_27_10_30_05_34-2342-1.mp4",
			pattern:  "ns",
			lower:    true,
			want:     "2023-10-27T10.30.05.000 - 2343_2023_10_27_10_30_05_34-2342-1.mp4",
		},
		{
			test:     "ns no match",
			filename: "20231027_103005123.jpg",
			pattern:  "ns",
			lower:    false,
			want:     "20231027_103005123.jpg",
		},
		{
			test:     "ns idempotency",
			filename: "2023-10-27T10.30.05.000 - TEST VID 2023_10_27_10_30_05-1.mp4",
			pattern:  "ns",
			lower:    false,
			want:     "2023-10-27T10.30.05.000 - TEST VID 2023_10_27_10_30_05-1.mp4",
		},
		{
			test:     "ns idempotency lowercase",
			filename: "2023-10-27T10.30.05.000 - TEST VID 2023_10_27_10_30_05-1.mp4",
			pattern:  "ns",
			lower:    true,
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
			want:     "2023-10-27T10.30.05.000 - 20231027103005.jpg",
		},
		{
			test:     "fs leading",
			filename: "IMG_20231027103005.jpg",
			pattern:  "fs",
			lower:    false,
			want:     "2023-10-27T10.30.05.000 - IMG_20231027103005.jpg",
		},
		{
			test:     "fs leading lowercase",
			filename: "IMG_20231027103005.jpg",
			pattern:  "fs",
			lower:    true,
			want:     "2023-10-27T10.30.05.000 - img_20231027103005.jpg",
		},
		{
			test:     "fs leading and trailing",
			filename: "PXL_20231027103005123-1.jpg",
			pattern:  "fs",
			lower:    false,
			want:     "2023-10-27T10.30.05.000 - PXL_20231027103005123-1.jpg",
		},
		{
			test:     "fs leading and trailing lowercase",
			filename: "PXL_20231027103005123-1.jpg",
			pattern:  "fs",
			lower:    true,
			want:     "2023-10-27T10.30.05.000 - pxl_20231027103005123-1.jpg",
		},
		{
			test:     "fs additional numerics",
			filename: "2343_20231027103005123_34-2342-1.jpg",
			pattern:  "fs",
			lower:    true,
			want:     "2023-10-27T10.30.05.000 - 2343_20231027103005123_34-2342-1.jpg",
		},
		{
			test:     "fs no match",
			filename: "2023-10-27-10-30-05-456.mp4",
			pattern:  "fs",
			lower:    false,
			want:     "2023-10-27-10-30-05-456.mp4",
		},
		{
			test:     "fs idempotency",
			filename: "2023-10-27T10.30.05.000 - PXL_20231027103005123-1.jpg",
			pattern:  "fs",
			lower:    false,
			want:     "2023-10-27T10.30.05.000 - PXL_20231027103005123-1.jpg",
		},
		{
			test:     "fs idempotency lowercase",
			filename: "2023-10-27T10.30.05.000 - PXL_20231027103005123-1.jpg",
			pattern:  "fs",
			lower:    true,
			want:     "2023-10-27T10.30.05.000 - PXL_20231027103005123-1.jpg",
		},
	}
	runFormatTests(t, tests)
}
