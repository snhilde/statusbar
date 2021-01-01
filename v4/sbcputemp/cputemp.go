// Package sbcputemp displays the temperature of the CPU in degrees Celsius.
// Currently only supported on Linux.
package sbcputemp

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

	// Path to the device directory, as discovered in findDir().
	path string

	// Slice of files that contain temperature readings.
	files []os.FileInfo

	// Average temperature across all sensors, in degrees Celsius.
	temp int

	// Trio of user-provided colors for displaying various states.
	colors struct {
		normal  string
		warning string
		error   string
	}
}

// New finds the device directory, builds a list of all the temperature sensors in it, and makes a new object. colors is
// an optional triplet of hex color codes for colorizing the output based on these rules:
//   1. Normal color, CPU temperature is cooler than 75 °C.
//   2. Warning color, CPU temperature is between 75 °C and 100 °C.
//   3. Error color, CPU temperature is hotter than 100 °C.
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

	// Error will be handled in Update() and String().
	r.path, r.err = findDir()
	if r.err != nil {
		return &r
	}

	// Error will be handled in Update() and String().
	r.files, r.err = findFiles(r.path)

	return &r
}

// Update reads out the value of each sensor, gets an average of all temperatures, and converts it from milliCelsius to
// Celsius. If we have trouble reading a particular sensor, then we'll skip it on this pass.
func (r *Routine) Update() {
	var n int

	if r.path == "" || len(r.files) == 0 {
		return
	}

	r.temp = 0
	for _, file := range r.files {
		b, err := ioutil.ReadFile(r.path + file.Name())
		if err != nil {
			continue
		}

		n, err = strconv.Atoi(strings.TrimSpace(string(b)))
		if err != nil {
			continue
		}

		r.temp += n
	}

	// Get the average temp across all readings.
	r.temp /= len(r.files)

	// Convert to degrees Celsius.
	r.temp /= 1000
}

// String prints a formatted temperature average in degrees Celsius.
func (r *Routine) String() string {
	var c string

	if r.err != nil {
		return r.colors.error + r.err.Error() + colorEnd
	}

	if r.temp < 75 {
		c = r.colors.normal
	} else if r.temp < 100 {
		c = r.colors.warning
	} else {
		c = r.colors.error
	}

	return fmt.Sprintf("%s%v °C%s", c, r.temp, colorEnd)
}

// findDir finds the directory that has the temperature readings. It will be the one with the fan speeds,
// somewhere in /sys/class/hwmon.
func findDir() (string, error) {
	// Get all the device directories in the main directory.
	dirs, err := ioutil.ReadDir(baseDir)
	if err != nil {
		return "", err
	}

	// Search in each device directory to find the fan.
	for _, dir := range dirs {
		path := baseDir + dir.Name() + "/device/"
		files, err := ioutil.ReadDir(path)
		if err != nil {
			return "", err
		}

		// If we encounter a file that matches "fan.*output", then we have the right directory.
		for _, file := range files {
			if strings.HasPrefix(file.Name(), "fan") && strings.HasSuffix(file.Name(), "output") {
				// We found our directory. Return the path.
				return path, nil
			}
		}
	}

	// If we made it here, then we didn't find anything.
	return "", errors.New("No fan file")
}

// findFiles goes through the given path and builds a list of files that contain a temperature reading. These files will
// begin with "temp" and end with "input".
func findFiles(path string) ([]os.FileInfo, error) {
	var b []os.FileInfo

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if strings.HasPrefix(file.Name(), "temp") && strings.HasSuffix(file.Name(), "input") {
			// We found a temperature reading. Add it to the list.
			b = append(b, file)
		}
	}

	// Make sure we found something.
	if len(b) == 0 {
		return nil, errors.New("No temperature files")
	}

	return b, nil
}
