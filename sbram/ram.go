// Package sbram displays the currently used and total system memory.
package sbram

import (
	"errors"
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

// New makes a new routine object.
func New(colors ...[3]string) *Routine {
	var r Routine

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

// Update gets the memory resources. Unfortunately, we can't use syscall.Sysinfo() or another syscall function, because
// it doesn't return the necessary information to calculate the actual amount of RAM in use at the moment (namely, it is
// missing the amount of cached RAM). Instead, we're going to read out /proc/meminfo and grab the values we need from
// there. All lines of that file have three fields: field name, value, and unit
func (r *Routine) Update() {
	file, err := ioutil.ReadFile("/proc/meminfo")
	if err != nil {
		r.err = err
		return
	}

	total, avail, err := parseFile(string(file))
	if err != nil {
		r.err = err
		return
	}

	if total == 0 || avail == 0 {
		r.err = errors.New("Failed to parse out memory fields")
		return
	}

	r.perc = (total - avail) * 100 / total
	r.total, r.totalUnit = shrink(total)
	r.used, r.usedUnit = shrink(total - avail)
}

// String formats and prints the used and total system memory.
func (r *Routine) String() string {
	var c string

	if r.err != nil {
		return r.colors.error + r.err.Error() + colorEnd
	}

	if r.perc < 75 {
		c = r.colors.normal
	} else if r.perc < 90 {
		c = r.colors.warning
	} else {
		c = r.colors.error
	}

	return fmt.Sprintf("%s%.1f%c/%.1f%c%s", c, r.used, r.usedUnit, r.total, r.totalUnit, colorEnd)
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
				return 0, 0, errors.New("Invalid MemTotal fields")
			}
			total, err = strconv.Atoi(fields[1])
			if err != nil {
				return 0, 0, err
			}

		} else if strings.HasPrefix(line, "MemAvailable") {
			fields := strings.Fields(line)
			if len(fields) != 3 {
				return 0, 0, errors.New("Invalid MemAvailable fields")
			}
			avail, err = strconv.Atoi(fields[1])
			if err != nil {
				return 0, 0, err
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
