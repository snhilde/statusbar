// Package sbram displays the currently used and total system memory.
package sbram

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

var colorEnd = "^d^"

// Routine is the main object for this package.
type Routine struct {
	// Error encountered along the way, if any.
	err error

	// Percentage of memory in use.
	perc int

	// Total amount of memory.
	total float32

	// Unit of total memory.
	totalUnit rune

	// Amount of memory in current use.
	used float32

	// Unit of used memory.
	usedUnit rune

	// Trio of user-provided colors for displaying various states.
	colors struct {
		normal  string
		warning string
		error   string
	}
}

// New makes a new routine object. colors is an optional triplet of hex color codes for colorizing the output based on
// these rules:
//   1. Normal color, less than 75% of available RAM is being used.
//   2. Warning color, between 75% and 90% of available RAM is being used.
//   3. Error color, more than 90% of available RAM is being used.
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

// Update gets the memory resources. Unfortunately, we can't use syscall.Sysinfo() or another syscall function, because
// it doesn't return the necessary information to calculate the actual amount of RAM in use at the moment (namely, it is
// missing the amount of cached RAM). Instead, we're going to read out /proc/meminfo and grab the values we need from
// there. All lines of that file have three fields: field name, value, and unit
func (r *Routine) Update() (bool, error) {
	if r == nil {
		return false, fmt.Errorf("bad routine")
	}

	file, err := ioutil.ReadFile("/proc/meminfo")
	if err != nil {
		r.err = fmt.Errorf("error reading file")
		return true, err
	}

	total, avail, err := parseFile(string(file))
	if err != nil {
		r.err = err
		return true, err
	}

	if total == 0 || avail == 0 {
		r.err = fmt.Errorf("failed to parse memory fields")
		return true, r.err
	}

	r.perc = (total - avail) * 100 / total
	r.total, r.totalUnit = shrink(total)
	r.used, r.usedUnit = shrink(total - avail)

	return true, nil
}

// String formats and prints the used and total system memory.
func (r *Routine) String() string {
	if r == nil {
		return "bad routine"
	}

	var color string

	if r.perc < 75 {
		color = r.colors.normal
	} else if r.perc < 90 {
		color = r.colors.warning
	} else {
		color = r.colors.error
	}

	return fmt.Sprintf("%s%.1f%c/%.1f%c%s", color, r.used, r.usedUnit, r.total, r.totalUnit, colorEnd)
}

// Error formats and returns an error message.
func (r *Routine) Error() string {
	if r == nil {
		return "bad routine"
	}

	if r.err == nil {
		r.err = fmt.Errorf("unknown error")
	}

	return r.colors.error + r.err.Error() + colorEnd
}

// Name returns the display name of this module.
func (r *Routine) Name() string {
	return "RAM"
}

// parseFile parses the meminfo file.
func parseFile(output string) (int, int, error) {
	var total int
	var avail int
	var err error

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "MemTotal") {
			fields := strings.Fields(line)
			if len(fields) != 3 {
				return 0, 0, fmt.Errorf("invalid MemTotal fields")
			}
			total, err = strconv.Atoi(fields[1])
			if err != nil {
				return 0, 0, fmt.Errorf("error parsing MemTotal fields")
			}

		} else if strings.HasPrefix(line, "MemAvailable") {
			fields := strings.Fields(line)
			if len(fields) != 3 {
				return 0, 0, fmt.Errorf("invalid MemAvailable fields")
			}
			avail, err = strconv.Atoi(fields[1])
			if err != nil {
				return 0, 0, fmt.Errorf("error parsing MemAvailable fields")
			}
		}
	}

	return total, avail, nil
}

// shrink iteratively decreases the amount of bytes by a step of 2^10 until human-readable.
func shrink(memory int) (float32, rune) {
	var units = [...]rune{'K', 'M', 'G', 'T', 'P', 'E'}
	var i int

	f := float32(memory)
	for f > 1024 {
		f /= 1024
		i++
	}

	return f, units[i]
}
