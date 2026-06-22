package voit

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"
	"time"
)

func TestMetaFormat(t *testing.T) {
	voitTime := time.Date(2026, time.February, 2, 12, 5, 20, 700000000, time.UTC)

	tests := []struct {
		name       string
		m          Meta
		args       Config
		wantFormat string
	}{
		{
			name: "sanitized format",
			m: Meta{
				VTime: VTime{Time: voitTime},
				Tags:  Tag{Items: []string{"summer", "vacation", "beach"}},
				Desc:  Desc{Text: "beach vacation"},
			},
			args:       Config{},
			wantFormat: "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
		},
		{
			name: "lowercase description",
			m: Meta{
				VTime: VTime{Time: voitTime},
				Tags:  Tag{Items: []string{"summer", "vacation", "beach"}},
				Desc:  Desc{Text: "Beach VACATION"},
			},
			args:       Config{Voit: VoitConfig{Lower: true}},
			wantFormat: "2026-02-02T12.05.20.700 beach vacation -- summer vacation beach",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.m.Format(tt.args)

			if got != tt.wantFormat {
				t.Errorf("\nGot:  %q\nWant: %q\n", got, tt.wantFormat)
			}
		})
	}
}

func TestVoitMove(t *testing.T) {
	t.Parallel() // TODO check this for all other tests too.

	tests := []struct {
		name       string
		args       Config
		mockStat   func(name string) (os.FileInfo, error)
		mockRename func(oldpath, newpath string) error
		wantErr    bool
		wantErrStr string
		wantLog    string
	}{
		{
			name:       "no overwrite collision [error]",
			args:       Config{Voit: VoitConfig{Verbose: true}},
			mockStat:   func(name string) (os.FileInfo, error) { return nil, nil },
			mockRename: func(oldpath string, newpath string) error { return nil },
			wantErr:    true,
			wantErrStr: "Collision:",
		},
		{
			name:       "overwrite collision [file overwritten]",
			args:       Config{Voit: VoitConfig{Overwrite: true}},
			mockStat:   func(name string) (os.FileInfo, error) { return nil, nil },
			mockRename: func(oldpath string, newpath string) error { return nil },
			wantErr:    false,
			wantLog:    "",
		},
		{
			name:       "standard rename [success]",
			args:       Config{},
			mockStat:   func(name string) (os.FileInfo, error) { return nil, os.ErrNotExist },
			mockRename: func(oldpath string, newpath string) error { return nil },
			wantErr:    false,
			wantLog:    "",
		},
		{
			name:       "verbose [success]",
			args:       Config{Voit: VoitConfig{Verbose: true}},
			mockStat:   func(name string) (os.FileInfo, error) { return nil, os.ErrNotExist },
			mockRename: func(oldpath string, newpath string) error { return nil },
			wantErr:    false,
			wantLog:    "Renamed:",
		},
		{
			name:       "rename failure [logged, error]",
			args:       Config{Voit: VoitConfig{Verbose: true}},
			mockStat:   func(name string) (os.FileInfo, error) { return nil, os.ErrNotExist },
			mockRename: func(oldpath string, newpath string) error { return errors.New("permission denied") },
			wantErr:    true,
			wantLog:    "Error renaming:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.args.FS = MockFS{
				MockStat:   tt.mockStat,
				MockRename: tt.mockRename,
			}

			v := &VoitImpl{}
			var buf bytes.Buffer
			err := v.move(&buf, "src.jpg", "tgt.jpg", tt.args)

			if (err != nil) != tt.wantErr {
				t.Fatalf("\nError\nGot:  %v\nWant: %v\n", err, tt.wantErr)
			}

			if tt.wantErr && !strings.Contains(err.Error(), tt.wantErrStr) {
				t.Errorf("\nError string\nGot:            %q\nWant substring: %q\n", err.Error(), tt.wantErrStr)
			}

			outputLog := buf.String()
			if tt.wantLog == "" && outputLog != "" {
				t.Errorf("\nLog\nGot:  %q\nWant: ''\n", outputLog)
			} else if tt.wantLog != "" && !strings.Contains(outputLog, tt.wantLog) {
				t.Errorf("\nLog string\nGot:        %q\nWant substring: %q\n", tt.wantLog, outputLog)
			}
		})
	}
}
