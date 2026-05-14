// Parse filename to Voit standard using provided regex patterns.
//
// Always parse full year, and assume time format is correctly zero padded when
// no separators are used. Valid separator is any non-numeric character.
//
// https://karl-voit.at/folder-hierarchy
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	webkitEpochOffset = 11644473600 // Secs between Jan 1, 1601 and Jan 1, 1970.
	msPerSec          = 1000000
)

var patterns = map[string]*regexp.Regexp{
	// ms - YYYYMMDD_HHMMSSSSS.
	"ms": regexp.MustCompile(`(\d{4})(\d{2})(\d{2})\D(\d{2})(\d{2})(\d{2})(\d{3})`),
	// mns - YYYY-MM-DD-HH-MM-SS-SSS.
	"mns": regexp.MustCompile(`(\d{4})\D(\d{2})\D(\d{2})\D(\d{2})\D(\d{2})\D(\d{2})\D(\d{3})`),
	// mfs - YYYYMMDDHHMMSSSSS.
	"mfs": regexp.MustCompile(`(\d{4})(\d{2})(\d{2})(\d{2})(\d{2})(\d{2})(\d{3})`),
	// s - YYYYMMDD_HHMMSS.
	"s": regexp.MustCompile(`(\d{4})(\d{2})(\d{2})\D(\d{2})(\d{2})(\d{2})`),
	// ns - YYYY-MM-DD-HH-MM-SS.
	"ns": regexp.MustCompile(`(\d{4})\D(\d{2})\D(\d{2})\D(\d{2})\D(\d{2})\D(\d{2})`),
	// fs - YYYYMMDDHHMMSS.
	"fs": regexp.MustCompile(`(\d{4})(\d{2})(\d{2})(\d{2})(\d{2})(\d{2})`),
	// hsfs - YYYY-MM-DD-HHMM
	"hsfs": regexp.MustCompile(`(\d{4})\D(\d{2})\D(\d{2})\D(\d{2})(\d{2})`),
	// hfs - YYYY-MM-DD-HHMMSS
	"hfs": regexp.MustCompile(`(\d{4})\D(\d{2})\D(\d{2})\D(\d{2})(\d{2})(\d{2})`),
	// hmfs - YYYY-MM-DD-HHMMSSSSS
	"hmfs": regexp.MustCompile(`(\d{4})\D(\d{2})\D(\d{2})\D(\d{2})(\d{2})(\d{2})(\d{3})`),
	// v - YYYY-MM-DDTHH.MM.SS.SSS.
	"v": regexp.MustCompile(`(\d{4})-(\d{2})-(\d{2})T(\d{2})\.(\d{2})\.(\d{2})\.(\d{3})`),
	// w - SSSSSSSSSSSSSSSSS (https://www.epochconverter.com/webkit).
	"w": regexp.MustCompile(`(\d{17})`),
}

// Parse time object from given file name and filter.
func parseFile(file string, pattern string) (time.Time, error) {
	match := patterns[pattern].FindStringSubmatch(file)
	if len(match) >= 7 {
		year, _ := strconv.Atoi(match[1])
		month, _ := strconv.Atoi(match[2])
		day, _ := strconv.Atoi(match[3])
		hour, _ := strconv.Atoi(match[4])
		min, _ := strconv.Atoi(match[5])
		sec, _ := strconv.Atoi(match[6])

		ms := 0
		if len(match) > 7 && match[7] != "" {
			ms, _ = strconv.Atoi(match[7])
		}

		return time.Date(year, time.Month(month), day, hour, min, sec, ms*int(time.Millisecond), time.UTC), nil
	}

	if len(match) == 6 {
		year, _ := strconv.Atoi(match[1])
		month, _ := strconv.Atoi(match[2])
		day, _ := strconv.Atoi(match[3])
		hour, _ := strconv.Atoi(match[4])
		min, _ := strconv.Atoi(match[5])

		return time.Date(year, time.Month(month), day, hour, min, 0, 0, time.UTC), nil
	}

	if len(match) == 2 {
		ms, err := strconv.ParseInt(match[1], 10, 64)
		if err != nil {
			return time.Time{}, fmt.Errorf("Invalid Webkit format: %s", file)
		}
		sec := (ms / msPerSec) - webkitEpochOffset
		date := time.Unix(sec, (ms%msPerSec)*1000).UTC()

		return time.Date(
			date.Year(),
			date.Month(),
			date.Day(),
			date.Hour(),
			date.Minute(),
			date.Second(),
			date.Nanosecond(),
			time.UTC,
		), nil
	}

	return time.Time{}, fmt.Errorf("No date pattern matched file: %s", file)
}

// Parse file creation or modification time.
// Fallback to modified time if underlying FS does not support creation time.
func parseFileTime(file string, created bool, modified bool) (time.Time, error) {
	info, err := os.Stat(file)
	if err != nil || (!created && !modified) {
		return time.Time{}, err
	}
	cTime := info.ModTime().UTC()

	if created {
		switch runtime.GOOS {
		case "linux":
			if stat, ok := info.Sys().(*syscall.Stat_t); ok {
				cTime = time.Unix(stat.Ctim.Sec, stat.Ctim.Nsec).UTC()
			}
		}
	}

	return time.Date(
		cTime.Year(),
		cTime.Month(),
		cTime.Day(),
		cTime.Hour(),
		cTime.Minute(),
		cTime.Second(),
		cTime.Nanosecond(),
		time.UTC,
	), nil
}

// Return target file. Ignore files already using Voit method unless explicitly
// renaming Voit file names.
func FormatName(filename string, pattern string, lower bool, strip bool, created bool, modified bool) string {
	var date time.Time
	var err error
	if created || modified {
		date, err = parseFileTime(filename, created, modified)
	} else {
		date, err = parseFile(filename, pattern)
	}

	if err != nil {
		return filename
	}
	if strings.HasPrefix(filename, date.Format("2006-01-02T15.04.05.000")) && pattern != "v" {
		return filename
	}

	file := filename
	if lower {
		file = strings.ToLower(filename)
	}

	if strip {
		extension := filepath.Ext(file)
		return fmt.Sprintf("%s%s", date.Format("2006-01-02T15.04.05.000"), extension)
	}
	return fmt.Sprintf("%s - %s", date.Format("2006-01-02T15.04.05.000"), file)
}
