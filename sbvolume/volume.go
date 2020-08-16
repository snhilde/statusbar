// Package sbvolume displays the current volume as a percentage.
package sbvolume

import (
	"errors"
	"strings"
	"os/exec"
	"strconv"
	"fmt"
)

var COLOR_END = "^d^"

// routine is the main object for this package.
// err:     error encountered along the way, if any
// control: control to query, as passed in by called
// vol:     system volume, in multiple of ten, as percentage of max
// muted:   true if volume is muted
// colors:  trio of user-provided colors for displaying various states
type routine struct {
	err     error
	control string
	vol     int
	muted   bool
	colors  struct {
		normal  string
		warning string
		error   string
	}
}

// Store the passed-in control value and return a new routine object.
func New(control string, colors ...[3]string) *routine {
	var r routine

	r.control = control

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

// Run the 'amixer' command and parse the output for mute status and volume percentage.
func (r *routine) Update() {
	r.muted = false
	r.vol   = -1

	out, err := r.runCmd()
	if err != nil {
		r.err = err
		return
	}

	// Find the line that has the percentage volume and mute status in it.
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Playback") && strings.Contains(line, "%]") {
			// We found it. Check the mute status, then pull out the volume.
			fields := strings.Fields(line)
			for _, field := range fields {
				field = strings.Trim(field, "[]")
				if field == "off" {
					r.muted = true
				} else if strings.HasSuffix(field, "%") {
					s        := strings.TrimRight(field, "%")
					vol, err := strconv.Atoi(s)
					if err != nil {
						r.err = err
						return
					}
					r.vol = normalize(vol)
				}
			}
			break;
		}
	}

	if r.vol < 0 {
		r.err = errors.New("No volume found for " + r.control)
	}
}

// Print either an error, the mute status, or the volume percentage.
func (r *routine) String() string {
	if r.err != nil {
		return r.colors.error + r.err.Error() + COLOR_END
	}

	if r.muted {
		return r.colors.warning + "Vol mute" + COLOR_END
	}

	return fmt.Sprintf("%sVol %v%%%s", r.colors.normal, r.vol, COLOR_END)
}

// Run the actual 'amixer' command, with the given control.
func (r *routine) runCmd() (string, error) {
	cmd      := exec.Command("amixer", "get", r.control)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(out), nil
}

// Ensure that the volume is a multiple of 10 (so it looks nicer).
func normalize(vol int) int {
	return (vol+5) / 10 * 10
}
