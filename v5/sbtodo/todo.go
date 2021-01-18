// Package sbtodo displays the first two lines of a TODO list.
package sbtodo

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

var colorEnd = "^d^"

// Routine is the main object for this package. It contains the data obtained from the specified
// TODO file, including file info and a copy of the first 2 lines.
type Routine struct {
	// Error encountered along the way, if any.
	err error

	// Path to the TODO file.
	path string

	// TODO file info, as returned by os.Stat().
	info os.FileInfo

	// First line of the TODO file.
	line1 string

	// Second line of the TODO file.
	line2 string

	// Trio of user-provided colors for displaying various states.
	colors struct {
		normal  string
		warning string
		error   string
	}
}

// New makes a new routine object. path is the absolute path to the TODO file. colors is an optional
// triplet of hex color codes for colorizing the output based on these rules:
//   1. Normal color, used for normal printing.
//   2. Warning color, currently unused.
//   3. Error color, used for printing error messages.
func New(path string, colors ...[3]string) *Routine {
	var r Routine

	r.path = path

	// Store the color codes. Don't do any validation.
	if len(colors) > 0 {
		r.colors.normal = "^c" + colors[0][0] + "^"
		r.colors.warning = "^c" + colors[0][1] + "^"
		r.colors.error = "^c" + colors[0][2] + "^"
	} else {
		// If a color array wasn't passed in, then we don't want to print this.
		colorEnd = ""
	}

	// Grab the base details of the TODO file.
	info, err := os.Stat(path)
	if err != nil {
		r.err = err
		return &r
	}

	if err := r.readFile(); err != nil {
		r.err = fmt.Errorf("error reading file")
		return &r
	}

	r.info = info
	return &r
}

// Update reads the TODO file again, if it was modified since the last read.
func (r *Routine) Update() (bool, error) {
	if r == nil {
		return false, fmt.Errorf("bad routine")
	}

	// Catch a possible error raised in New.
	if r.info == nil {
		if r.err == nil {
			r.err = fmt.Errorf("invalid TODO file")
		}
		return false, r.err
	}

	newInfo, err := os.Stat(r.path)
	if err != nil {
		r.err = fmt.Errorf("error getting file stats")
		return true, err
	}

	// If mtime is not newer than what we already have, we can skip reading the file.
	newMtime := newInfo.ModTime().UnixNano()
	oldMtime := r.info.ModTime().UnixNano()
	if newMtime > oldMtime {
		// The file was modified. Let's parse it.
		if err := r.readFile(); err != nil {
			r.err = fmt.Errorf("error reading file")
			return true, err
		}
	}
	r.info = newInfo

	return true, nil
}

// String formats the first two lines of the file according to these rules:
//   1. If the file is empty, print "Finished".
//   2. If only one line in the file has content, print only that line.
//   3. If one line has content and the next line with content is indented (tabs or spaces), print
//      "line1 -> line2".
//   4. If two lines have content and both are flush, print "line1 | line2".
func (r *Routine) String() string {
	if r == nil {
		return "bad routine"
	}

	// First, let's figure out what joiner (if any) we need.
	joiner := ""
	if r.line1 != "" && r.line2 != "" {
		if (strings.HasPrefix(r.line2, "\t")) || (strings.HasPrefix(r.line2, " ")) {
			joiner = " -> "
		} else {
			joiner = " | "
		}
	}

	line1 := strings.TrimSpace(r.line1)
	line2 := strings.TrimSpace(r.line2)

	output := line1 + joiner + line2
	if output == "" {
		output = "Finished"
	}

	return r.colors.normal + output + colorEnd
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
	return "TODO"
}

// readFile grabs the first two lines of the TODO file that are not blank.
func (r *Routine) readFile() error {
	r.line1 = ""
	r.line2 = ""

	contents, err := ioutil.ReadFile(r.path)
	if err != nil {
		return err
	}

	lines := strings.Split(string(contents), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			if r.line1 == "" {
				r.line1 = line
			} else {
				r.line2 = line
				break
			}
		}
	}

	return nil
}
