// Package sbbattery displays the percentage of battery capacity left as well as
// a charging status indicator.
// TODO: run lint
package sbbattery

import (
	"errors"
	"strings"
	"io/ioutil"
	"strconv"
	"fmt"
)

var COLOR_END = "^d^"

const (
	UNKNOWN  = -1
	CHARGING = iota
	DISCHARGING
	FULL
)

// Main type for package.
// err:    error encountered along the way, if any
// max:    maximum capacity of battery
// perc:   percentage of battery capacity left
// colors: trio of user-provided colors for displaying various states
type routine struct {
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

// Read the maximum capacity of the battery and return struct.
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

	// Error will be handled in both Update() and String().
	r.max, r.err = readCharge("/sys/class/power_supply/BAT0/charge_full")

	return &r
}

// Read current capacity left and calculate percentage from that.
func (r *routine) Update() {
	// Handle error reading max capacity.
	if r.max == -1 {
		return
	}

	// Get current charge and calculate percentage.
	now, err := readCharge("/sys/class/power_supply/BAT0/charge_now")
	if err != nil {
		r.err = err
		return
	}

	r.perc  = (now * 100) / r.max
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
func (r *routine) String() string {
	var c string
	var s string

	if r.err != nil {
		return r.colors.error + r.err.Error() + COLOR_END
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

	return fmt.Sprintf("%s%s BAT%s", c, s, COLOR_END)
}

// Read out value from file.
func readCharge(path string) (int, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return -1, err
	}

	return strconv.Atoi(strings.TrimSpace(string(b)))
}
