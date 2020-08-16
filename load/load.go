// Package sbload displays the average system load over the last one, five, and fifteen minutes.
package sbload

import (
	"errors"
	"strings"
	"syscall"
	"fmt"
)

var COLOR_END = "^d^"

// routine is the main object in the package.
// err:     error encountered along the way, if any
// load_1:  load average over the last second
// load_5:  load average over the last 5 seconds
// load_15: load average over the last 15 seconds
// colors:  trio of user-provided colors for displaying various states
type routine struct {
	err     error
	load_1  float64
	load_5  float64
	load_15 float64
	colors  struct {
		normal  string
		warning string
		error   string
	}
}

// Return a new rountine object.
func New(colors ...[3]string) *routine {
	var r routine

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

// Call Sysinfo() method and calculate load averages.
func (r *routine) Update() {
	var info syscall.Sysinfo_t

	r.err = syscall.Sysinfo(&info)
	if r.err != nil {
		return
	}

	// Each load average must be divided by 2^16 to get the same format as /proc/loadavg.
	r.load_1  = float64(info.Loads[0]) / float64(1 << 16)
	r.load_5  = float64(info.Loads[1]) / float64(1 << 16)
	r.load_15 = float64(info.Loads[2]) / float64(1 << 16)
}

// Print the 3 load averages with 2 decimal places of precision.
func (r *routine) String() string {
	var c string

	if r.err != nil {
		return r.colors.error + r.err.Error() + COLOR_END
	}

	if r.load_1 >= 2 || r.load_5 >= 2 || r.load_15 >= 2 {
		c = r.colors.error
	} else if r.load_1 >= 1 || r.load_5 >= 1 || r.load_15 >= 1 {
		c = r.colors.warning
	} else {
		c = r.colors.normal
	}

	return fmt.Sprintf("%s%.2f %.2f %.2f%s", c, r.load_1, r.load_5, r.load_15, COLOR_END)
}
