// Package sbbattery displays the percentage of battery capacity left with a charging status
// indicator.
package sbbattery

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

var colorEnd = "^d^"

// These are the possible charging states of the battery.
const (
	statusUnknown  = 0
	statusCharging = iota
	statusDischarging
	statusFull
)

// Routine is the main type for this package.
type Routine struct {
	// Error encountered along the way, if any.
	err error

	// Maximum capacity of battery.
	max int

	// Percentage of battery capacity left.
	perc int

	// Status of the battery (unknown, charging, discharging, or full).
	status int

	// The three user-provided colors for displaying the various states.
	colors struct {
		normal  string
		warning string
		error   string
	}
}

// New reads the maximum capacity of the battery and returns a Routine object. colors is an optional
// triplet of hex color codes for colorizing the output based on these rules:
//   1. Normal color, battery has more than 25% left.
//   2. Warning color, battery has between 10% and 25% left.
//   3. Error color, battery has less than 10% left.
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

	// Error will be handled in both Update() and String().
	r.max, r.err = readCharge("/sys/class/power_supply/BAT0/charge_full")

	return &r
}

// Update reads the current battery capacity left and calculates a percentage based on it.
func (r *Routine) Update() (bool, error) {
	if r == nil {
		return false, fmt.Errorf("bad routine")
	}

	// Handle error in New or error reading max capacity.
	if r.max <= 0 {
		return false, r.err
	}

	// Get current charge and calculate a percentage.
	now, err := readCharge("/sys/class/power_supply/BAT0/charge_now")
	if err != nil {
		r.err = fmt.Errorf("error reading charge")
		return true, err
	}

	r.perc = (now * 100) / r.max
	switch {
	case r.perc < 0:
		r.perc = 0
	case r.perc > 100:
		r.perc = 100
	}

	// Get charging status.
	status, err := ioutil.ReadFile("/sys/class/power_supply/BAT0/status")
	if err != nil {
		r.err = fmt.Errorf("error reading status")
		return true, err
	}

	switch strings.TrimSpace(string(status)) {
	case "Charging":
		r.status = statusCharging
	case "Discharging":
		r.status = statusDischarging
	case "Full":
		r.status = statusFull
	default:
		r.status = statusUnknown
	}

	return true, nil
}

// String formats the percentage of battery left.
func (r *Routine) String() string {
	if r == nil {
		return "bad routine"
	}

	var c string
	if r.perc > 25 {
		c = r.colors.normal
	} else if r.perc > 10 {
		c = r.colors.warning
	} else {
		c = r.colors.error
	}

	s := fmt.Sprintf("%v%%", r.perc)
	if r.status == statusCharging {
		s = "+" + s
	} else if r.status == statusDischarging {
		s = "-" + s
	} else if r.status == statusFull {
		s = "Full"
	}

	return fmt.Sprintf("%s%s BAT%s", c, s, colorEnd)
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
	return "Battery"
}

// readCharge reads out the value from the file at the provided path.
func readCharge(path string) (int, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return -1, err
	}

	return strconv.Atoi(strings.TrimSpace(string(b)))
}
