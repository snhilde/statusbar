// Package sbload displays the average system load over the last one, five, and fifteen minutes.
package sbload

import (
	"fmt"
	"syscall"
)

var colorEnd = "^d^"

// Routine is the main object in the package.
type Routine struct {
	// Error encountered along the way, if any.
	err error

	// Load average over the last second.
	load1 float64

	// Load average over the last 5 seconds.
	load5 float64

	// Load average over the last   15 seconds.
	load15 float64

	// Trio of user-provided colors for displaying various states.
	colors struct {
		normal  string
		warning string
		error   string
	}
}

// New makes a new rountine object. colors is an optional triplet of hex color codes for colorizing the output based on
// these rules:
//   1. Normal color, all load averages are below 1.
//   2. Warning color, one or more load averages is greater than 1, but all are less than 2.
//   3. Error color, one or more load averages is greater than 2.
func New(colors ...[3]string) *Routine {
	var r Routine

	// Store the color codes. Don't do any validation.
	if len(colors) > 0 {
		r.colors.normal = "^c" + colors[0][0] + "^"
		r.colors.warning = "^c" + colors[0][1] + "^"
		r.colors.error = "^c" + colors[0][2] + "^"
	} else {
		// If a color array wasn't passed in, then we don't want to print this.
		colorEnd = ""
	}

	return &r
}

// Update calls Sysinfo() and calculates load averages.
func (r *Routine) Update() (bool, error) {
	if r == nil {
		return false, fmt.Errorf("bad routine")
	}

	// Handle any error encountered in New.
	if r.err != nil {
		return true, r.err
	}

	var info syscall.Sysinfo_t
	err := syscall.Sysinfo(&info)
	if err != nil {
		r.err = fmt.Errorf("error getting stats")
		return true, err
	}

	// Each load average must be divided by 2^16 to get the same format as /proc/loadavg.
	r.load1 = float64(info.Loads[0]) / float64(1<<16)
	r.load5 = float64(info.Loads[1]) / float64(1<<16)
	r.load15 = float64(info.Loads[2]) / float64(1<<16)

	return true, nil
}

// String prints the 3 load averages with 2 decimal places of precision.
func (r *Routine) String() string {
	if r == nil {
		return "bad routine"
	}

	var c string

	if r.load1 >= 2 || r.load5 >= 2 || r.load15 >= 2 {
		c = r.colors.error
	} else if r.load1 >= 1 || r.load5 >= 1 || r.load15 >= 1 {
		c = r.colors.warning
	} else {
		c = r.colors.normal
	}

	return fmt.Sprintf("%s%.2f %.2f %.2f%s", c, r.load1, r.load5, r.load15, colorEnd)
}

// Error formats and returns an error message.
func (r *Routine) Error() string {
	if r == nil {
		return "bad routine"
	}

	if r.err == nil {
		r.err = fmt.Errorf("unknown error")
	}

	s := r.colors.error + r.err.Error() + colorEnd
	r.err = nil

	return s
}

// Name returns the display name of this module.
func (r *Routine) Name() string {
	return "Load"
}
