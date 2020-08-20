// Package sbtime displays the current time in the provided format.
package sbtime

import (
	"errors"
	"strings"
	"time"
)

var colorEnd = "^d^"

// Routine is the main object for the sbtime package.
type Routine struct {
	// Error with the color selection, if any.
	err error

	// Current timestamp.
	time time.Time

	// Format for displaying time, when colons are displayed (every other second).
	formatA string

	// Format for displaying time, when colons are blinked out (every other second).
	formatB string

	// Trio of user-provided colors for displaying various states.
	colors struct {
		normal  string
		warning string
		error   string
	}
}

// New creates a new routine object with the current time. format is the format to use when printing the time, as per
// the go standard used in the time package. If the format includes colons, they will blink every other second. colors
// is an optional triplet of hex color codes for colorizing the output based on these rules:
// 1. Normal color, used for normal printing.
// 2. Warning color, currently unused.
// 3. Error color, used for printing error messages.
func New(format string, colors ...[3]string) *Routine {
	var r Routine

	// Replace all colons in the format string with spaces, to get the blinking effect later.
	r.formatA = format
	r.formatB = strings.Replace(format, ":", " ", -1)
	r.time = time.Now()

	// Do a minor sanity check on the color codes.
	if len(colors) == 1 {
		for _, color := range colors[0] {
			if !strings.HasPrefix(color, "#") || len(color) != 7 {
				r.err = errors.New("Invalid color")
				return &r
			}
		}
		r.colors.normal = "^c" + colors[0][0] + "^"
		r.colors.warning = "^c" + colors[0][1] + "^"
		r.colors.error = "^c" + colors[0][2] + "^"
	} else {
		// If a color array wasn't passed in, then we don't want to print this.
		colorEnd = ""
	}

	return &r
}

// Update updates the routine's current time.
func (r *Routine) Update() {
	r.time = time.Now()
}

// String prints the time in the provided format.
func (r *Routine) String() string {
	if r.err != nil {
		return r.colors.error + r.err.Error() + colorEnd
	}

	if r.time.Second()%2 == 0 {
		return r.colors.normal + r.time.Format(r.formatA) + colorEnd
	}
	return r.colors.normal + r.time.Format(r.formatB) + colorEnd
}
