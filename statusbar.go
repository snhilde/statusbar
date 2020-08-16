package statusbar

// #cgo pkg-config: x11
// #cgo LDFLAGS: -lX11
// #include <X11/Xlib.h>
import "C"

import (
	"os"
	"os/signal"
	"strings"
	"time"
)

// RoutineHandler interface allows information monitors to be linked in.
type RoutineHandler interface {
	Update()        // Update the routine's information. Will be run according to the provided interval time.
	String() string // Format and return the routine's output.
}

// routine holds the data for an individual unit on the statusbar.
type routine struct {
	rh       RoutineHandler
	interval time.Duration
}

// Statusbar is the main type for this package.
type Statusbar struct {
	routines []routine
	left     string
	right    string
	split    int
}

// New creates a new statusbar. The default delimiters around each routine are square brackets ('[' and ']').
func New() Statusbar {
	s := Statusbar{left: "[", right: "]", split: -1}
	return s
}

// Append adds a routine to the statusbar's list. Routines will be displayed in order of addition to the bar object.
func (sb *Statusbar) Append(rh RoutineHandler, s int) {
	// Convert the given number into proper seconds.
	seconds := time.Duration(s) * time.Second

	r := routine{rh, seconds}
	sb.routines = append(sb.routines, r)
}

// Run spins up every routine and displays them on the statusbar.
func (sb *Statusbar) Run() {
	// Add a signal handler so we can clear the statusbar if the program goes down.
	go sb.handleSignal()

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

// runRoutine runs the routine in a non-terminating loop.
// TODO: handle errors
func runRoutine(r routine, i int, ch chan []string) {
	for {
		// Start the clock.
		start := time.Now()

		// Update the contents of the routine.
		r.rh.Update()

		// Get the routine's output and store it in the master output slice.
		output := r.rh.String()
		outputs := <-ch
		outputs[i] = output
		ch <- outputs

		// If interval was set for infinite sleep, then we'll close routine here.
		if r.interval == 0 {
			break
		}

		// Put the routine to sleep for the given time.
		time.Sleep(r.interval - time.Since(start))
	}
}

// setBar builds the master output and prints it to the statusbar. This runs a loop every second.
func setBar(ch chan []string, sb Statusbar) {
	var b strings.Builder

	dpy := C.XOpenDisplay(nil)
	root := C.XDefaultRootWindow(dpy)

	// This loop will run twice a second to catch any changes that run every second.
	for {
		// Start the clock.
		start := time.Now()
		b.Reset()

		// Receive the outputs slice and build the individual outputs into a master output.
		// TODO: handle empty strings (if b is empty, b.String() will fail too)
		// TODO: handle error strings
		outputs := <-ch
		for i, s := range outputs {
			if len(s) > 0 {
				b.WriteString(sb.left)
				b.WriteString(s)
				b.WriteString(sb.right)
				b.WriteByte(' ')
			}

			if i == sb.split {
				// Insert the breaking delimiter here.
				b.WriteByte(';')
			}
		}
		ch <- outputs

		s := b.String()
		s = s[:b.Len()-1] // remove last space

		// Send the master output to the statusbar.
		C.XStoreName(dpy, root, C.CString(s))
		C.XSync(dpy, 1)

		// Stop the clock and put the routine to sleep for the rest of the second.
		end := time.Now()
		time.Sleep((time.Second / 2) - end.Sub(start))
	}
}

// SetMarkers sets the left and right delimiters around each routine. If not set, these will default to '[' and ']'.
func (sb *Statusbar) SetMarkers(left string, right string) {
	sb.left = left
	sb.right = right
}

// Split splits the statusbar at this point, for dualstatus patch. A semicolon (';') will be inserted at this point in
// the routine list, which will signal to dualstatus to split the statusbar at this point.
func (sb *Statusbar) Split() {
	sb.split = len(sb.routines) - 1
}

// handleSignal clears the statusbar if the program receives an interrupt signal.
func (sb *Statusbar) handleSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	pid := os.Getpid()
	p, err := os.FindProcess(pid)
	if err != nil {
		return
	}

	// Wait until we receive an interrupt signal.
	s := <-c

	dpy := C.XOpenDisplay(nil)
	root := C.XDefaultRootWindow(dpy)
	C.XStoreName(dpy, root, C.CString(s.String()))
	C.XSync(dpy, 1)

	// Stop the program.
	p.Kill()
}
