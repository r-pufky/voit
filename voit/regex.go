/*
Regex patterns for matching date times in filenames.

Voit 8601 Regex Breakdown (YYYY-MM-DDTHH.MM.SS.SSS):
^\s*          - Ignore leading whitespace
(\d{4})       - Group 1: Year (4 digits)
-             - hyphen
(\d{2})       - Group 2: Month (2 digits)
-             - hyphen
(\d{2})       - Group 3: Day (2 digits)
(?:           - Start optional outer group (allows partial matches)

	T(\d{2})   - T Group 4: Hour (2 digits)
	(?:
	  \.(\d{2}) - . Group 5: Minute (2 digits)
	  (?:
	    \.(\d{2})(?:\.(\d{3}))? - . Group 6 (Second), . Group 7 (Millisecond)
	  )?
	)?

)?
\s*$           - Ignore trailing whitespace
*/

package voit

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	WebkitEpochOffset = 11644473600 // Secs between Jan 1, 1601 and Jan 1, 1970.
	MsPerSec          = 1000000     // MS per second.
)

var Patterns = map[string]*regexp.Regexp{
	"photo-ms":      regexp.MustCompile(`(\d{4})(\d{2})(\d{2})\D(\d{2})(\d{2})(\d{2})(\d{3})`),                                    // YYYYMMDD░HHMMSSSSS
	"signal-ms":     regexp.MustCompile(`(\d{4})\D(\d{2})\D(\d{2})\D(\d{2})\D(\d{2})\D(\d{2})\D(\d{3})`),                          // YYYY░MM░DD░HH░MM░SS░SSS
	"8601-naked-ms": regexp.MustCompile(`(\d{4})(\d{2})(\d{2})(\d{2})(\d{2})(\d{2})(\d{3})`),                                      // YYYYMMDDHHMMSSSSS
	"photo":         regexp.MustCompile(`(\d{4})(\d{2})(\d{2})\D(\d{2})(\d{2})(\d{2})`),                                           // YYYYMMDD░HHMMSS
	"signal":        regexp.MustCompile(`(\d{4})\D(\d{2})\D(\d{2})\D(\d{2})\D(\d{2})\D(\d{2})`),                                   // YYYY░MM░DD░HH░MM░SS
	"8601-naked":    regexp.MustCompile(`(\d{4})(\d{2})(\d{2})(\d{2})(\d{2})(\d{2})`),                                             // YYYYMMDDHHMMSS
	"8601-short":    regexp.MustCompile(`(\d{4})\D(\d{2})\D(\d{2})\D(\d{2})(\d{2})`),                                              // YYYY░MM░DD░HHMM
	"8601":          regexp.MustCompile(`(\d{4})\D(\d{2})\D(\d{2})\D(\d{2})(\d{2})(\d{2})`),                                       // YYYY░MM░DD░HHMMSS
	"8601-ms":       regexp.MustCompile(`(\d{4})\D(\d{2})\D(\d{2})\D(\d{2})(\d{2})(\d{2})(\d{3})`),                                // YYYY░MM░DD░HHMMSSSSS
	"unix":          regexp.MustCompile(`(\d{1,13})`),                                                                             // SSSSSSSSSSSSS Unix Epoch (with MS).
	"voit":          regexp.MustCompile(`^\s*(\d{4})-(\d{2})-(\d{2})(?:T(\d{2})(?:\.(\d{2})(?:\.(\d{2})(?:\.(\d{3}))?)?)?)?\s*$`), // Voit 8601
	"voit-span":     regexp.MustCompile(`^\s*(\d{4})-(\d{2})-(\d{2})(?:T(\d{2})(?:\.(\d{2})(?:\.(\d{2})(?:\.(\d{3}))?)?)?)?\s*$`), // Voit 8601 date span (span captured in extract)
	"webkit-chrome": regexp.MustCompile(`(\d{17})`),                                                                               // SSSSSSSSSSSSSSSSS (https://www.epochconverter.com/webkit)
	"created":       regexp.MustCompile(`[^\s\S]`),                                                                                // Never match
	"modified":      regexp.MustCompile(`[^\s\S]`),                                                                                // Never match
	"set":           regexp.MustCompile(`^\s*(\d{4})-(\d{2})-(\d{2})(?:T(\d{2})(?:\.(\d{2})(?:\.(\d{2})(?:\.(\d{3}))?)?)?)?\s*$`), // Voit 8601
}

// Extract time object from given string and filter.
func Extract(name string, pattern string) (time.Time, error) {
	match := Patterns[pattern].FindStringSubmatch(name)
	if match == nil {
		return time.Time{}, fmt.Errorf("no date pattern matched: %s", name)
	}

	// 0 - full match, 1..N - Regex match groups. Unmatched groups ("") and out
	// of bounds matches are set to 0.
	group := func(i int) int {
		if match == nil || i >= len(match) {
			return 0
		}
		val, _ := strconv.Atoi(match[i])
		return val
	}

	if pattern == "webkit-chrome" {
		ms, err := strconv.ParseInt(match[1], 10, 64)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid webkit format: %s", name)
		}
		sec := (ms / MsPerSec) - WebkitEpochOffset
		return time.Unix(sec, (ms%MsPerSec)*1000).UTC(), nil
	}

	if pattern == "unix" {
		epoch, err := strconv.ParseInt(match[1], 10, 64)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid unix format: %s", name)
		}
		if len(match[1]) > 10 {
			return time.UnixMilli(epoch).UTC(), nil
		}

		return time.Unix(epoch, 0).UTC(), nil
	}

	return time.Date(
		group(1),
		time.Month(group(2)),
		group(3),
		group(4),
		group(5),
		group(6),
		group(7)*int(time.Millisecond),
		time.UTC,
	), nil
}

// Strip matched regex from string. Invalid patterns and non-matched regex
// returns original string.
func Strip(s string, pattern string) string {
	if regex, ok := Patterns[pattern]; ok {
		replaced := regex.ReplaceAllString(s, "")
		return strings.TrimSpace(strings.Join(strings.Fields(replaced), " "))
	}
	return s
}
