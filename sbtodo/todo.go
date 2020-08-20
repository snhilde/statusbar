// Package sbtodo displays the first two lines of a TODO list.
package sbtodo

import (
	"errors"
	"io/ioutil"
	"os"
	"strings"
)

var colorEnd = "^d^"

// Routine is the main object for this package. It contains the data obtained from the specified TODO file, including
// file info and a copy of the first 2 lines.
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

// New makes a new routine object. path is the absolute path to the TODO file. colors is an optional triplet of hex
//   1. Normal color, used for normal printing.
//   2. Warning color, currently unused.
//   3. Error color, used for printing error messages.
func New(path string, colors ...[3]string) *Routine {
	var r Routine

	r.path = path

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

	// Grab the base details of the TODO file.
	r.info, r.err = os.Stat(path)
	if r.err != nil {
		// We'll print the error in String().
		return &r
	}

	if err := r.readFile(); err != nil {
		// We'll print the error in String().
		r.err = err
		return &r
	}

	return &r
}

// Update reads the TODO file again, if it was modified since the last read.
func (r *Routine) Update() {
	newInfo, err := os.Stat(r.path)
	if err != nil {
		r.err = err
		return
	}

	// If mtime is not newer than what we already have, we can skip reading the file.
	newMtime := newInfo.ModTime().UnixNano()
	oldMtime := r.info.ModTime().UnixNano()
	if newMtime > oldMtime {
		// The file was modified. Let's parse it.
		if err := r.readFile(); err != nil {
			// We'll print the error in String().
			r.err = err
			return
		}
	}

	r.info = newInfo
}

// String formats the first two lines of the file according to these rules:
//   1. If the file is empty, print "Finished".
//   2. If only one line in the file has content, print only that line.
//   3. If one line has content and the next line with content is indented (tabs or spaces), print "line1 -> line2".
//   4. If two lines have content and both are flush, print "line1 | line2".
func (r *Routine) String() string {
	var b strings.Builder

	// Handle any error we might have received in another stage.
	if r.err != nil {
		return r.colors.error + r.err.Error() + colorEnd
	}

	r.line1 = strings.TrimSpace(r.line1)
	b.WriteString(r.colors.normal)
	if r.line1 != "" {
		// We have content in the first line. Start by adding that.
		b.WriteString(r.line1)
		if r.line2 != "" {
			// We have content in the second line as well. First, let's find out which joiner to use.
			if (strings.HasPrefix(r.line2, "\t")) || (strings.HasPrefix(r.line2, " ")) {
				b.WriteString(" -> ")
			} else {
				b.WriteString(" | ")
			}
			// Next, we'll add the second line.
			b.WriteString(strings.TrimSpace(r.line2))
		}
	} else {
		if len(r.line2) > 0 {
			// We only have a second line. Print just that.
			b.WriteString(strings.TrimSpace(r.line2))
		} else {
			// We don't have content in either line.
			b.WriteString("Finished")
		}
	}
	b.WriteString(colorEnd)

	return b.String()
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
