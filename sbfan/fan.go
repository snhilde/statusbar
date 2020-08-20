// Package sbfan displays the current fan speed in RPM.
package sbfan

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

var colorEnd = "^d^"

// We need to root around in this directory for the device directory for the fan.
const baseDir = "/sys/class/hwmon/"

// Routine is the main object for this package.
type Routine struct {
	// Error encountered along the way, if any.
	err error

	// Found path to the device directory.
	path string

	// File that contains the maximum speed of the fan, in RPM.
	maxFile os.FileInfo

	// File that contains the current speed of the fan, in RPM.
	outFile os.FileInfo

	// Maximum speed of the fan, in RPM.
	max int

	// Current speed of the fan, in RPM.
	out int

	// Percentage of maximum fan speed.
	perc int

	// Trio of user-provided colors for displaying various states.
	colors struct {
		normal  string
		warning string
		error   string
	}
}

// New searches around in the base directory for a pair of max and current files and makes a new routine object. colors
// is an optional triplet of hex color codes for colorizing the output based on these rules:
// Color 1: Normal color, fan is running at less than 75% of the maximum RPM.
// Color 2: Warning color, fan is running at between 75% and 90% of the maximum RPM.
// Color 3: Error color, fan is running at more than 90% of the maximum RPM.
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

	// Find the max fan speed file and read its value.
	r.findFiles()
	if r.err != nil {
		return &r
	}

	// Error will be handled later in Update() and String().
	r.max, r.err = readSpeed(r.path + r.maxFile.Name())

	return &r
}

// Update reads the current fan speed in RPM and calculates the percentage of the maximum speed.
func (r *Routine) Update() {
	if r.err != nil {
		return
	}

	r.out, r.err = readSpeed(r.path + r.outFile.Name())
	if r.err != nil {
		return
	}

	r.perc = (r.out * 100) / r.max
	if r.perc > 100 {
		r.perc = 100
	}
}

// String prints the current speed in RPM.
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

	return fmt.Sprintf("%s%v RPM%s", c, r.out, colorEnd)
}

// findFiles finds the files that we'll monitor for the fan speed. It will be in one of the hardware device directories
// in /sys/class/hwmon.
func (r *Routine) findFiles() {
	var dirs []os.FileInfo
	var files []os.FileInfo

	// Get all the device directories in the main directory.
	dirs, r.err = ioutil.ReadDir(baseDir)
	if r.err != nil {
		return
	}

	// Search in each device directory to find the fan.
	for _, dir := range dirs {
		path := baseDir + dir.Name() + "/device/"
		files, r.err = ioutil.ReadDir(path)
		if r.err != nil {
			return
		}

		// Find the first file that has a name match. The files we want will start with "fan" and end
		// with "max" or "output".
		prefix := "fan"
		for _, file := range files {
			if strings.HasPrefix(file.Name(), prefix) {
				if strings.HasSuffix(file.Name(), "max") || strings.HasSuffix(file.Name(), "output") {
					// We found one of the two.
					if strings.HasSuffix(file.Name(), "max") {
						r.maxFile = file
						prefix = strings.TrimSuffix(file.Name(), "max")
					} else {
						r.outFile = file
						prefix = strings.TrimSuffix(prefix, "output")
					}
				}

				// If we've found both files, we can stop looking.
				if r.maxFile != nil && r.outFile != nil {
					r.path = path
					return
				}
			}
		}
	}

	// If we made it here, then we didn't find anything.
	r.err = errors.New("No fan file")
	return
}

// readSpeed reads the value of the provided file. The value will be a speed in RPM.
func readSpeed(path string) (int, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return -1, err
	}

	return strconv.Atoi(strings.TrimSpace(string(b)))
}
