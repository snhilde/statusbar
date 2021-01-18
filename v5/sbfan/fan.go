// Package sbfan displays the current fan speed in RPM.
package sbfan

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
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

	// Path to file that contains the current speed of the fan, in RPM.
	fanPath string

	// Maximum speed of the fan, in RPM.
	max int

	// Current speed of the fan, in RPM.
	speed int

	// Trio of user-provided colors for displaying various states.
	colors struct {
		normal  string
		warning string
		error   string
	}
}

// New searches around in the base directory for a pair of max and current files and makes a new
// routine object. colors is an optional triplet of hex color codes for colorizing the output based
// on these rules:
//   1. Normal color, fan is running at less than 75% of the maximum RPM.
//   2. Warning color, fan is running at between 75% and 90% of the maximum RPM.
//   3. Error color, fan is running at more than 90% of the maximum RPM.
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

	// Find the files holding the values for the maximum fan speed and the current fan speed.
	maxFile, outFile, err := findFiles()
	if err != nil {
		r.err = err
		return &r
	}

	// Find the max fan speed file and read its value.
	r.max, r.err = readSpeed(maxFile)

	r.fanPath = outFile
	return &r
}

// Update reads the current fan speed in RPM and calculates the percentage of the maximum speed.
func (r *Routine) Update() (bool, error) {
	if r == nil {
		return false, fmt.Errorf("bad routine")
	}

	// Handle any error encountered in New.
	if r.fanPath == "" || r.max == 0 {
		return false, r.err
	}

	speed, err := readSpeed(r.fanPath)
	if err != nil {
		r.err = fmt.Errorf("error reading speed")
		return true, err
	}

	r.speed = speed
	return true, nil
}

// String prints the current speed in RPM.
func (r *Routine) String() string {
	if r == nil {
		return "bad routine"
	}

	var c string

	perc := (r.speed * 100) / r.max
	if perc > 100 {
		perc = 100
	}

	if perc < 75 {
		c = r.colors.normal
	} else if perc < 90 {
		c = r.colors.warning
	} else {
		c = r.colors.error
	}

	return fmt.Sprintf("%s%v RPM%s", c, r.speed, colorEnd)
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
	return "Fan"
}

// findFiles finds the files that we'll monitor for the fan speed. It will be in one of the hardware
// device directories in /sys/class/hwmon.
func findFiles() (string, string, error) {
	// Get all the device directories in the main directory.
	dirs, err := ioutil.ReadDir(baseDir)
	if err != nil {
		return "", "", err
	}

	// Search in each device directory to find the fan.
	for _, dir := range dirs {
		path := filepath.Join(baseDir, dir.Name(), "/device/")
		files, err := ioutil.ReadDir(path)
		if err != nil {
			return "", "", err
		}

		// Find the first file that has a name match. The files we want will start with "fan" and
		// end with "max" or "output".
		prefix := "fan"
		maxFilename := "" // File that contains the maximum speed of the fan, in RPM.
		outFilename := "" // File that contains the current speed of the fan, in RPM.
		for _, file := range files {
			filename := file.Name()
			if strings.HasPrefix(filename, prefix) {
				if strings.HasSuffix(filename, "max") {
					maxFilename = filepath.Join(path, filename)
					prefix = strings.TrimSuffix(filename, "max")
				} else if strings.HasSuffix(filename, "output") {
					outFilename = filepath.Join(path, filename)
					prefix = strings.TrimSuffix(prefix, "output")
				}

				// If we've found both files, we can stop looking.
				if maxFilename != "" && outFilename != "" {
					return maxFilename, outFilename, nil
				}
			}
		}
	}

	// If we made it here, then we didn't find anything.
	return "", "", fmt.Errorf("no fan file")
}

// readSpeed reads the value of the provided file. The value will be a speed in RPM.
func readSpeed(path string) (int, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return -1, err
	}

	return strconv.Atoi(strings.TrimSpace(string(b)))
}
