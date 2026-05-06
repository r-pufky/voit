// Parse filename to Voit standard using provided regex patterns.
//
// Always parse full year, and assume time format is correctly zero padded when
// no separators are used. Valid separator is any non-numeric character.
//
// https://karl-voit.at/folder-hierarchy
package main

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
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
}

// Parse time object from given file name and filter.
func parseFile(file string, pattern string) (time.Time, error) {
	regex := patterns[pattern]
	if match := regex.FindStringSubmatch(file); len(match) >= 7 {
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

	return time.Time{}, fmt.Errorf("no date pattern matched file: %s", file)
}

// Return target file name, ignoring any files already using the Voit method.
func FormatName(filename string, pattern string, lower bool, strip bool) string {
	date, err := parseFile(filename, pattern)
	if err != nil {
		return filename
	}
	if strings.HasPrefix(filename, date.Format("2006-01-02T15.04.05.000")) {
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
