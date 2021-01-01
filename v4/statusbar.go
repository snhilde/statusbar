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

// RoutineHandler allows information monitors (commonly called routines) to be linked in.
type RoutineHandler interface {
	Update()        // Update the routine's information. This is run according to the provided interval time.
	String() string // Format and return the routine's output.
}

// routine holds the data for an individual unit on the statusbar.
type routine struct {
	rh       RoutineHandler
	interval time.Duration
}

// Statusbar is the main type for this package. It holds information about the bar as a whole.
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

// Append adds a routine to the statusbar's list. Routines are displayed in the order they are added.
func (sb *Statusbar) Append(rh RoutineHandler, s int) {
	// Convert the given number into proper seconds.
	seconds := time.Duration(s) * time.Second

	r := routine{rh, seconds}
	sb.routines = append(sb.routines, r)
}

// Run spins up all the routines and displays them on the statusbar.
func (sb *Statusbar) Run() {
	// Add a signal handler so we can clear the statusbar if the program goes down.
	go sb.handleSignal()

	// A slice of strings to hold the output from each routine
	outputs := make([]string, len(sb.routines))

	// Shared channel used to pass the slice of outputs
	ch := make(chan []string, 1)
	ch <- outputs

	// Channel used to indicate everything is done
	finished := make(chan error)

	for i, r := range sb.routines {
		go runRoutine(r, i, ch)
	}

	// Launch a goroutine to build and print the master string.
	go setBar(ch, *sb)

	// Keep running forever (TODO: this is currently unused).
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

		// If the interval was set for infinite sleep, then we can close the routine now.
		if r.interval == 0 {
			break
		}

		// Put the routine to sleep for the given time.
		time.Sleep(r.interval - time.Since(start))
	}
}

// setBar builds the master output and prints it to the statusbar. This runs a loop twice a second to catch any changes
// that run every second.
func setBar(ch chan []string, sb Statusbar) {
	dpy := C.XOpenDisplay(nil)
	root := C.XDefaultRootWindow(dpy)

	for {
		// Start the clock.
		start := time.Now()
		b := new(strings.Builder)

		// Receive the outputs slice and build the individual outputs into a master output.
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

		s := ""
		if b.Len() > 0 {
			s = b.String()
			s = s[:b.Len()-1] // remove last space
		}

		// Send the master output to the statusbar.
		C.XStoreName(dpy, root, C.CString(s))
		C.XSync(dpy, 1)

		// Put the routine to sleep for the rest of the half second.
		time.Sleep((time.Second / 2) - time.Since(start))
	}
}

// SetMarkers sets the left and right delimiters around each routine. If not set, these will default to '[' and ']'.
func (sb *Statusbar) SetMarkers(left string, right string) {
	sb.left = left
	sb.right = right
}

// Split splits the statusbar at this point, when using dualstatus patch for dwm. A semicolon (';') is inserted at this
// point in the routine list, which signals to dualstatus to split the statusbar at this point. Before this is called,
// the routines already added are displayed on the top bar. After this is called, all subsequently added routines are
// displayed on the bottom bar.
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
