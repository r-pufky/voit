/*
Regex patterns for matching datetimes in filenames.

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

package models

import (
	"regexp"
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
	"voit":          regexp.MustCompile(`^\s*(\d{4})-(\d{2})-(\d{2})(?:T(\d{2})(?:\.(\d{2})(?:\.(\d{2})(?:\.(\d{3}))?)?)?)?\s*$`), // Voit 8601
	"webkit-chrome": regexp.MustCompile(`(\d{17})`),                                                                               // SSSSSSSSSSSSSSSSS (https://www.epochconverter.com/webkit)
	"created":       regexp.MustCompile(`[^\s\S]`),                                                                                // Never match
	"modified":      regexp.MustCompile(`[^\s\S]`),                                                                                // Never match
}
