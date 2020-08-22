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
	// Update updates the routine's information. This is run on a periodic interval according to the time provided.
	// It returns two arguments: a bool and an error. The bool indicates whether or not the engine should continue to
	// run the routine. You can think of it as representing the "ok" status. The error is any error encountered during
	// the process. For example, on a normal run with no error, Update would return (true, nil). On a run with a
	// non-critical error, Update would return (true, errors.New("Warning message")). On a run with a critical error
	// where the routine should not be attempted again, Update would return (false, errors.New("Critical error message").
	Update() (bool, error)

	// String formats and returns the routine's output.
	String() string
}

// routine holds the data for an individual unit on the statusbar.
type routine struct {
	// Routine object that handles running the actual process.
	rh RoutineHandler

	// Time in seconds to wait between each run.
	interval time.Duration
}

// Statusbar is the main type for this package. It holds information about the bar as a whole.
type Statusbar struct {
	// List of routines, in the order they were added.
	routines []routine

	// Delimiter to use for the left side of each routine's output, as set with SetMarkers.
	left string

	// Delimiter to use for the right side of each routine's output, as set with SetMarkers.
	right string

	// Index of the interval after which the routines are split, as set with Split.
	split int
}

// New creates a new statusbar. The default delimiters around each routine are square brackets ('[' and ']').
func New() Statusbar {
	return Statusbar{left: "[", right: "]", split: -1}
}

// Append adds a routine to the statusbar's list. Routines are displayed in the order they are added. rh is the
// RoutineHandler module. seconds is the amount of time between each run of the routine.
func (sb *Statusbar) Append(rh RoutineHandler, seconds int) {
	// Convert the given number into proper seconds.
	s := time.Duration(seconds) * time.Second
	r := routine{rh, s}

	sb.routines = append(sb.routines, r)
}

// Run spins up all the routines and displays them on the statusbar.
func (sb *Statusbar) Run() {
	// Add a signal handler so we can clear the statusbar if the program goes down.
	go sb.handleSignal()

	// A slice of strings to hold the output from each routine
	outputs := make([]string, len(sb.routines))

	// Shared channel used to pass the slice of outputs
	outputsChan := make(chan []string, 1)
	outputsChan <- outputs

	// Channel used to indicate everything is done
	finished := make(chan struct{})

	for i, r := range sb.routines {
		go runRoutine(r, i, outputsChan, finished)
	}

	// Launch a goroutine to build and print the master string.
	go setBar(outputsChan, *sb)

	// Keep running until every routine stops.
	for i := 0; i < len(sb.routines); i++ {
		<-finished
	}
}

// runRoutine runs the routine in a non-terminating loop.
func runRoutine(r routine, i int, outputsChan chan []string, finished chan struct{}) {
	for {
		// Start the clock.
		start := time.Now()

		// Update the contents of the routine.
		ok, err := r.rh.Update()

		// Get the routine's output and store it in the master output slice.
		output := r.rh.String()
		outputs := <-outputsChan
		outputs[i] = output
		outputsChan <- outputs

		if !ok {
			// The routine reported a critical error.
			break
		}

		interval := r.interval
		if err != nil {
			// If the routine reported an error, then we'll give the process a little time to cool down before trying again.
			s := r.interval.Seconds()
			switch {
			case s < 60:
				// For routines with intervals up to 1 minute, sleep for 5 seconds.
				interval = 5 * time.Second
			case s < 60 * 15:
				// For routines with intervals up to 15 minutes, sleep for 1 minute.
				interval = 60 * time.Second
			default:
				// For routines with intervals longer than 15 minutes, sleep for 5 minutes.
				interval = 60 * 5 * time.Second
			}
		}

		// If the interval was set to only run once, then we can close the routine now.
		if interval == 0 {
			break
		}

		// Put the routine to sleep for the given time.
		time.Sleep(interval - time.Since(start))
	}

	finished <- struct{}{}
}

// setBar builds the master output and prints it to the statusbar. This runs a loop twice a second to catch any changes
// that run every second.
func setBar(outputsChan chan []string, sb Statusbar) {
	dpy := C.XOpenDisplay(nil)
	root := C.XDefaultRootWindow(dpy)

	for {
		// Start the clock.
		start := time.Now()
		b := new(strings.Builder)

		// Receive the outputs slice and build the individual outputs into a master output.
		outputs := <-outputsChan
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
		outputsChan <- outputs

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
