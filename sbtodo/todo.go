// Package sbtodo displays the first two lines of a TODO list.
package sbtodo

import (
	"bufio"
	"errors"
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

// New makes a new routine object. path is the absolute path to the TODO file.
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

	r.readFile()
	if r.err != nil {
		// We'll print the error in String().
		return &r
	}

	return &r
}

// Update reads the TODO file again, if it was modified since the last read.
func (r *Routine) Update() {
	var newInfo os.FileInfo

	newInfo, r.err = os.Stat(r.path)
	if r.err != nil {
		return
	}

	// If mtime is not newer than what we already have, we can skip reading the file.
	newMtime := newInfo.ModTime().UnixNano()
	oldMtime := r.info.ModTime().UnixNano()
	if newMtime > oldMtime {
		// The file was modified. Let's parse it.
		r.readFile()
		if r.err != nil {
			return
		}
	}

	r.info = newInfo
}

// String formats the first two lines of the file according to these rules:
//   1. If the file is empty, print "Finished".
//   2. If the first line has content but the second line is empty, print only the first line.
//   3. If the first line is empty but the second line has content, print only the second line.
//   4. If the first line has content and the second line is indented, print "line1 -> line2".
//   5. If both lines have content and both are flush, print "line1 | line2".
func (r *Routine) String() string {
	var b strings.Builder

	// Handle any error we might have received in another stage.
	if r.err != nil {
		return r.colors.error + r.err.Error() + colorEnd
	}

	r.line1 = strings.TrimSpace(r.line1)
	b.WriteString(r.colors.normal)
	if len(r.line1) > 0 {
		// We have content in the first line. Start by adding that.
		b.WriteString(r.line1)
		if len(r.line2) > 0 {
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

// Grab the first two lines of the TODO file.
func (r *Routine) readFile() {
	var file *os.File

	file, r.err = os.Open(r.path)
	if r.err != nil {
		return
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	r.line1, r.err = reader.ReadString('\n')
	r.line2, r.err = reader.ReadString('\n')
}
