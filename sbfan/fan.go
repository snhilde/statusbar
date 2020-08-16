// Package sbfan displays the current fan speed in RPM.
package sbfan

import (
	"errors"
	"strings"
	"os"
	"io/ioutil"
	"strconv"
	"fmt"
)

var COLOR_END = "^d^"

// We need to root around in this directory for the device directory for the fan.
const base_dir = "/sys/class/hwmon/"

// routine is the main object for this package.
// err:      error encountered along the way, if any
// path:     found path to the device directory
// max_file: file that contains the maximum speed of the fan, in RPM
// out_file: file that contains the current speed of the fan, in RPM
// max:      maximum speed of the fan, in RPM
// out:      current speed of the fan, in RPM
// perc:     percentage of maximum fan speed
// colors:   trio of user-provided colors for displaying various states
type routine struct {
	err      error
	path     string
	max_file os.FileInfo
	out_file os.FileInfo
	max      int
	out      int
	perc     int
	colors   struct {
		normal  string
		warning string
		error   string
	}
}

// Search around in the base directory for a pair of max and current files, and return a new routine object.
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

	// Find the max fan speed file and read its value.
	r.findFiles()
	if r.err != nil {
		return &r
	}

	// Error will be handled later in Update() and String().
	r.max, r.err = readSpeed(r.path + r.max_file.Name())

	return &r
}

// Read the current fan speed in RPM and calculate the percentage of the maximum speed.
func (r *routine) Update() {
	if r.err != nil {
		return
	}

	r.out, r.err = readSpeed(r.path + r.out_file.Name())
	if r.err != nil {
		return
	}

	r.perc = (r.out * 100) / r.max
	if r.perc > 100 {
		r.perc = 100
	}
}

// Print the formatted current speed in RPM.
func (r *routine) String() string {
	var c string

	if r.err != nil {
		return r.colors.error + r.err.Error() + COLOR_END
	}

	if r.perc < 75 {
		c = r.colors.normal
	} else if r.perc < 90 {
		c = r.colors.warning
	} else {
		c = r.colors.error
	}

	return fmt.Sprintf("%s%v RPM%s", c, r.out, COLOR_END)
}

// Find the file that we'll monitor for the fan speed.
// It will be in one of the hardware device directories in /sys/class/hwmon.
func (r *routine) findFiles() {
	var dirs  []os.FileInfo
	var files []os.FileInfo

	// Get all the device directories in the main directory.
	dirs, r.err = ioutil.ReadDir(base_dir)
	if r.err != nil {
		return
	}

	// Search in each device directory to find the fan.
	for _, dir := range dirs {
		path := base_dir + dir.Name() + "/device/"
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
						r.max_file = file
						prefix     = strings.TrimSuffix(file.Name(), "max")
					} else {
						r.out_file = file
						prefix = strings.TrimSuffix(prefix, "output")
					}
				}

				// If we've found both files, we can stop looking.
				if r.max_file != nil && r.out_file != nil {
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

// Read the value of the passed-in file, which will be a speed in RPM.
func readSpeed(path string) (int, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return -1, err
	}

	return strconv.Atoi(strings.TrimSpace(string(b)))
}
