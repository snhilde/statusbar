// Package sbtime displays the current time in the provided format.
package sbtime

import (
	"errors"
	"strings"
	"time"
)

var COLOR_END = "^d^"

// A routine is the main object for the sbtime package.
// error:    error in colors, if any
// time:     current timestamp
// format_a: format for displaying time, when colons are displayed (every other second)
// format_b: format for displaying time, when colons are blinked out (every other second)
// colors:   trio of user-provided colors for displaying various states
type routine struct {
	err      error
	time     time.Time
	format_a string
	format_b string
	colors   struct {
		normal  string
		warning string
		error   string
	}
}

// Create a new routine object with the current time.
func New(format string, colors ...[3]string) *routine {
	var r routine

	// Replace all colons in the format string with spaces, to get the blinking effect later.
	r.format_a = format
	r.format_b = strings.Replace(format, ":", " ", -1)
	r.time     = time.Now()

	// Do a minor sanity check on the color codes.
	if len(colors) == 1 {
		for _, color := range colors[0] {
			if !strings.HasPrefix(color, "#") || len(color) != 7 {
				r.err = errors.New("Invalid color")
				return &r
			}
		}
		r.colors.normal  = "^c" + colors[0][0] + "^"
		r.colors.warning = "^c" + colors[0][1] + "^"
		r.colors.error   = "^c" + colors[0][2] + "^"
	} else {
		// If a color array wasn't passed in, then we don't want to print this.
		COLOR_END = ""
	}

	return &r
}

// Update the routine's current time.
func (r *routine) Update() {
	r.time = time.Now()
}

// Print the time in provided format.
func (r *routine) String() string {
	if r.err != nil {
		return r.colors.error + r.err.Error() + COLOR_END
	}

	if r.time.Second() % 2 == 0 {
		return r.colors.normal + r.time.Format(r.format_a) + COLOR_END
	} else {
		return r.colors.normal + r.time.Format(r.format_b) + COLOR_END
	}
}
