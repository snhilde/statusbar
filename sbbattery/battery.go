// Package sbbattery displays the percentage of battery capacity left with a charging status indicator.
package sbbattery

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

var colorEnd = "^d^"

const (
	UNKNOWN  = -1
	CHARGING = iota
	DISCHARGING
	FULL
)

// Routine is the main type for this package.
// err:    error encountered along the way, if any
// max:    maximum capacity of battery
// perc:   percentage of battery capacity left
// colors: trio of user-provided colors for displaying various states
type Routine struct {
	err    error
	max    int
	perc   int
	status int
	colors struct {
		normal  string
		warning string
		error   string
	}
}

// New reads the maximum capacity of the battery and returns a routinestruct.
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

	// Error will be handled in both Update() and String().
	r.max, r.err = readCharge("/sys/class/power_supply/BAT0/charge_full")

	return &r
}

// Update reads the current battery capacity left and calculates a percentage based on it.
func (r *Routine) Update() {
	// Handle error reading max capacity.
	if r.max < 0 {
		return
	}

	// Get current charge and calculate a percentage.
	now, err := readCharge("/sys/class/power_supply/BAT0/charge_now")
	if err != nil {
		r.err = err
		return
	}

	r.perc = (now * 100) / r.max
	if r.perc < 0 {
		r.perc = 0
	} else if r.perc > 100 {
		r.perc = 100
	}

	// Get charging status.
	status, err := ioutil.ReadFile("/sys/class/power_supply/BAT0/status")
	if err != nil {
		r.err = err
		return
	}

	switch strings.TrimSpace(string(status)) {
	case "Charging":
		r.status = CHARGING
	case "Discharging":
		r.status = DISCHARGING
	case "Full":
		r.status = FULL
	default:
		r.status = UNKNOWN
	}

}

// Print formatted percentage of battery left.
func (r *Routine) String() string {
	var c string
	var s string

	if r.err != nil {
		return r.colors.error + r.err.Error() + colorEnd
	}

	if r.perc > 25 {
		c = r.colors.normal
	} else if r.perc > 10 {
		c = r.colors.warning
	} else {
		c = r.colors.error
	}

	if r.status == CHARGING {
		s = fmt.Sprintf("+%v%%", r.perc)
	} else if r.status == DISCHARGING {
		s = fmt.Sprintf("-%v%%", r.perc)
	} else if r.status == FULL {
		s = "Full"
	} else {
		s = fmt.Sprintf("%v%%", r.perc)
	}

	return fmt.Sprintf("%s%s BAT%s", c, s, colorEnd)
}

// Read out value from file.
func readCharge(path string) (int, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return -1, err
	}

	return strconv.Atoi(strings.TrimSpace(string(b)))
}
