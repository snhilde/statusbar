// Package sbvolume displays the current volume as a percentage.
package sbvolume

import (
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

var colorEnd = "^d^"

// Routine is the main object for this package.
type Routine struct {
	// Error encountered along the way, if any.
	err error

	// Control to query, as provided by caller.
	control string

	// System volume, in multiples of ten, as percentage of max.
	vol int

	// True if volume is muted.
	muted bool

	// Trio of user-provided colors for displaying various states.
	colors struct {
		normal  string
		warning string
		error   string
	}
}

// New stores the provided control value and makes a new routine object. control is the mixer control to monitor. See
// the man pages for amixer for more information on that. colors is an optional triplet of hex color codes for
// colorizing the output based on these rules:
// Color 1: Normal color, for normal printing.
// Color 2: Warning color, for when the volume is muted.
// Color 3: Error color, for error messages.
func New(control string, colors ...[3]string) *Routine {
	var r Routine

	r.control = control

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

// Update runs the 'amixer' command and parses the output for mute status and volume percentage.
func (r *Routine) Update() {
	r.muted = false
	r.vol = -1

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
					s := strings.TrimRight(field, "%")
					vol, err := strconv.Atoi(s)
					if err != nil {
						r.err = err
						return
					}
					r.vol = normalize(vol)
				}
			}
			break
		}
	}

	if r.vol < 0 {
		r.err = errors.New("No volume found for " + r.control)
	}
}

// String prints either an error, the mute status, or the volume percentage.
func (r *Routine) String() string {
	if r.err != nil {
		return r.colors.error + r.err.Error() + colorEnd
	}

	if r.muted {
		return r.colors.warning + "Vol mute" + colorEnd
	}

	return fmt.Sprintf("%sVol %v%%%s", r.colors.normal, r.vol, colorEnd)
}

// runCmd runs the actual 'amixer' command, with the given control.
func (r *Routine) runCmd() (string, error) {
	cmd := exec.Command("amixer", "get", r.control)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(out), nil
}

// normalize ensures that the volume is a multiple of 10 (so it looks nicer).
func normalize(vol int) int {
	return (vol + 5) / 10 * 10
}
