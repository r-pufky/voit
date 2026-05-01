package main

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// Use dynamic separators. Always parse full year, and assume time format is
// correctly zero padded when no separators are used.
// Match longest common patterns first.
var datePatterns = []*regexp.Regexp{
	// 20230110_081705324 - PXL_20230110_081705324, IMG_20230110_081705324, etc.
	regexp.MustCompile(`(\d{4})(\d{2})(\d{2})\D?(\d{2})(\d{2})(\d{2})(\d{3})?`),

	// 20230122-214021 - Screenshot_20230122-214021.
	regexp.MustCompile(`(\d{4})(\d{1,2})(\d{1,2})\D?(\d{2})(\d{2})(\d{2})`),

	// 2022-11-22-09-13-12-123, 2022-11-22-09-13-12 - signal-2022-11-22-09-13-12-123.
	regexp.MustCompile(`(\d{4})\D?(\d{2})\D?(\d{2})\D?(\d{2})\D?(\d{2})\D?(\d{2})\D?(\d{3})?`),

	// 2022-11-22-151312 - signal-2022-11-22-151312
	regexp.MustCompile(`(\d{4})\D?(\d{2})\D?(\d{2})\D?(\d{2})(\d{2})(\d{2})`),
}

func ParseNewName(filename string, opts Options) (string, int, bool) {
	for _, re := range datePatterns {
		loc := re.FindStringIndex(filename)
		if loc == nil {
			continue
		}

		matches := re.FindStringSubmatch(filename)
		if len(matches) < 7 {
			continue
		}

		y, m, d := matches[1], Pad(matches[2], 2), Pad(matches[3], 2)
		hh, mm, ss := Pad(matches[4], 2), Pad(matches[5], 2), Pad(matches[6], 2)

		ms := "000"
		if len(matches) > 7 && matches[7] != "" {
			rawMs := matches[7]
			ms = Pad(rawMs, 3)
			if len(rawMs) > 3 {
				ms = rawMs[:3]
			}
		}

		timestamp := fmt.Sprintf("%s-%s-%sT%s.%s.%s.%s", y, m, d, hh, mm, ss, ms)

		processedFilename := filename
		if opts.Lower {
			processedFilename = strings.ToLower(filename)
		} else if opts.Upper {
			processedFilename = strings.ToUpper(filename)
		}

		if strings.HasPrefix(processedFilename, timestamp) {
			if processedFilename == filename {
				return filename, 0, false
			}
			return processedFilename, 0, true
		}

		return fmt.Sprintf("%s - %s", timestamp, processedFilename), loc[0], true
	}
	return filename, 0, false
}

func ResolveCollisions(dir, name string, counts map[string]int) (string, string) {
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)

	finalName := name
	if count, exists := counts[name]; exists && count > 0 {
		finalName = fmt.Sprintf("%s (%d)%s", base, count, ext)
	}

	counts[name]++
	return finalName, filepath.Join(dir, finalName)
}

func Pad(s string, length int) string {
	if len(s) >= length {
		return s[:length]
	}
	return strings.Repeat("0", length-len(s)) + s
}
