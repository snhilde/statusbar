// Package statusbar displays various resources on the DWM statusbar.
package statusbar

// #cgo pkg-config: x11
// #cgo LDFLAGS: -lX11
// #include <X11/Xlib.h>
import "C"

import (
	"fmt"
	"strings"
	"time"
)

// Updater interface allows resource monitors to be linked in.
type Updater interface {
	Update() error
	String() string
}

// A routine holds the data for an individual unit on the statusbar.
type routine struct {
	u        Updater
	interval time.Duration
}

// A statusbar is the main type for this package. It holds the slice of routines (ordered according
// to the user's instructions) and the left and right delimiters for each routine, as well as the
// position to insert the delimiter (";") for splitting the statusbar.
type statusbar struct {
	routines []routine
	left     string
	right    string
	split    int
}

// Create a new statusbar.
func New() statusbar {
	s := statusbar{left: "[", right: "]", split: -1}
	return s
}

// Append a routine to the statusbar's list.
func (sb *statusbar) Append(u Updater, s int) {
	// Convert the given number into proper seconds.
	seconds := time.Duration(s) * time.Second

	r          := routine{u, seconds}
	sb.routines = append(sb.routines, r)
}

// Spin up every routine and display them on the statusbar.
func (sb *statusbar) Run() {
	// Shared channel used to pass the slice of outputs
	ch := make(chan []string, 1)

	// A slice of strings to hold the output from each routine
	outputs := make([]string, len(sb.routines))
	ch <- outputs

	// Channel used to indicate everything is done
	// TODO: currently unused
	finished := make(chan error)

	for i, r := range sb.routines {
		go runRoutine(r, i, ch)
	}

	// Launch a goroutine to build and print the master string.
	go setBar(ch, *sb)

	// Wait for all routines to finish (shouldn't happen though).
	<-finished
}

// Run the routine in a non-terminating loop.
// TODO: handle errors
func runRoutine(r routine, i int, ch chan []string) {
	for {
		// Start the clock.
		start := time.Now()

		// Update the contents of the routine.
		r.u.Update()

		// Get the routine's output and set it in the master output slice.
		output    := r.u.String()
		outputs   := <-ch
		outputs[i] = output
		ch <- outputs

		// Stop the clock and put the routine to sleep for the given time.
		end := time.Now()
		time.Sleep(r.interval - end.Sub(start))
	}
}

// Build the master output and print in to the statusbar. Runs a loop every second.
func setBar(ch chan []string, sb statusbar) {
	var b strings.Builder

	dpy  := C.XOpenDisplay(nil)
	root := C.XDefaultRootWindow(dpy)

	for {
		// Start the clock.
		start := time.Now()

		// Receive the outputs slice and build the individual outputs into a master output.
		// TODO: handle empty strings (if b is empty, b.String() will fail too)
		// TODO: handle error strings
		outputs := <-ch
		for i, s := range outputs {
			if len(s) > 0 {
				fmt.Fprintf(&b, "%s%s%s ", sb.left, s, sb.right)
			}

			if i == sb.split {
				// Insert the breaking delimiter here.
				fmt.Fprintf(&b, ";")
			}
		}
		ch <- outputs

		s := b.String()
		s  = s[:b.Len()-1] // remove last space

		// Send the master output to the statusbar.
		C.XStoreName(dpy, root, C.CString(s));
		C.XSync(dpy, 1)
		b.Reset()

		// Stop the clock and put the routine to sleep for the rest of the second.
		end := time.Now()
		time.Sleep(time.Second - end.Sub(start))
	}
}

// Set the left and right markers around each routine.
func (sb *statusbar) SetBoundary(left string, right string) {
	sb.left  = left
	sb.right = right
}

// Split the statusbar at this index, for dualstatus patch.
func (sb *statusbar) Split() {
	sb.split = len(sb.routines)
}
