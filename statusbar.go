package statusbar

// #cgo pkg-config: x11
// #cgo LDFLAGS: -lX11
// #include <X11/Xlib.h>
import "C"

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"
)

// RoutineHandler allows information monitors (commonly called routines) to be linked in.
type RoutineHandler interface {
	// Update updates the routine's information. This is run on a periodic interval according to the time provided. It
	// returns two arguments: a bool and an error. The bool indicates whether or not the engine should continue to run
	// the routine. You can think of it as representing the "ok" status. The error is any error encountered during the
	// process. For example, on a normal run with no error, Update would return (true, nil). On a run with a
	// non-critical error, Update would return (true, errors.New("Warning message")). On a run with a critical error
	// where the routine should be stopped, Update would return (false, errors.New("Critical error message"). The
	// returned error will be logged to stderr. Generally, the error returned from Update should be detailed and
	// specific for debugging the routine, while the error returned from Error should be shorter, more concise, and more
	// general.
	Update() (bool, error)

	// String formats and returns the routine's output.
	String() string

	// Error formats and returns an error message. Generally, the error returned from Error should be shorter, more
	// concise, and more general, while the error returned from Update should be detailed and specific for debugging the
	// routine.
	Error() string

	// Name returns the display name of the module.
	Name() string
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
	leftDelim string

	// Delimiter to use for the right side of each routine's output, as set with SetMarkers.
	rightDelim string

	// Index of the interval after which the routines are split, as set with Split.
	split int
}

// New creates a new statusbar. The default delimiters around each routine are square brackets ('[' and ']').
func New() Statusbar {
	return Statusbar{leftDelim: "[", rightDelim: "]", split: -1}
}

// Append adds a routine to the statusbar's list. Routines are displayed in the order they are added. handler is the
// RoutineHandler module. seconds is the amount of time between each run of the routine.
func (sb *Statusbar) Append(handler RoutineHandler, seconds int) {
	// Convert the given number into proper seconds.
	s := time.Duration(seconds) * time.Second
	r := routine{handler, s}

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
	finished := make(chan routine)

	for i, r := range sb.routines {
		go runRoutine(r, i, outputsChan, finished)
	}

	// Launch a goroutine to build and print the master string.
	go setBar(outputsChan, *sb)

	// Keep running until every routine stops.
	for i := 0; i < len(sb.routines); i++ {
		r := <-finished
		logError("%v: Routine stopped", r.rh.Name())
	}

	logError("All routines have stopped")
}

// runRoutine runs a routine in a non-terminating loop.
func runRoutine(r routine, i int, outputsChan chan []string, finished chan<- routine) {
	handler := r.rh
	for {
		// Start the clock.
		start := time.Now()

		// Update the routine's data.
		ok, err := handler.Update()

		// Get the routine's output and store it in the master output slice.
		var output string
		if err == nil {
			output = handler.String()
		} else {
			output = handler.Error()
			logError("%v: %v", handler.Name(), err.Error())
		}
		outputs := <-outputsChan
		outputs[i] = output
		outputsChan <- outputs

		// If the routine reported a critical error, then we'll break out of the loop now.
		if !ok {
			break
		}

		// If the interval was set to only run once, then we can close the routine now.
		if r.interval == 0 {
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
			case s < 60*15:
				// For routines with intervals up to 15 minutes, sleep for 1 minute.
				interval = 60 * time.Second
			default:
				// For routines with intervals longer than 15 minutes, sleep for 5 minutes.
				interval = 60 * 5 * time.Second
			}
		}

		// Put the routine to sleep for the given time.
		time.Sleep(interval - time.Since(start))
	}

	// Send on the finished channel to signify that we're stopping this routine.
	finished <- r
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
				b.WriteString(sb.leftDelim)

				// Shorten outputs that are longer than 50 characters.
				if len(s) > 50 {
					s = s[:46] + "..."
				}
				b.WriteString(s)

				b.WriteString(sb.rightDelim)
				b.WriteByte(' ')
			}

			if i == sb.split {
				// Insert the breaking delimiter here.
				b.WriteByte(';')
			}
		}
		outputsChan <- outputs

		s := "No output" // Default if nothing else is available
		if b.Len() > 0 {
			s = b.String()
			s = s[:b.Len()-1] // Remove last space.
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
	sb.leftDelim = left
	sb.rightDelim = right
}

// Split splits the statusbar at this point, when using dualstatus patch for dwm. A semicolon (';') is inserted at this
// point in the routine list, which signals to dualstatus to split the statusbar at this point. Before this is called,
// the routines already added are displayed on the top bar. After this is called, all subsequently added routines are
// displayed on the bottom bar.
func (sb *Statusbar) Split() {
	sb.split = len(sb.routines) - 1
}

// logError prints the formatted message to stderr.
func logError(format string, a ...interface{}) {
	// Add a timestamp and a newline to our message.
	format = "(%v) " + format + "\n"
	items := append([]interface{}{time.Now()}, a...)

	fmt.Fprintf(os.Stderr, format, items...)
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
	<-c
	logError("Received interrupt")

	dpy := C.XOpenDisplay(nil)
	root := C.XDefaultRootWindow(dpy)
	C.XStoreName(dpy, root, C.CString("Statusbar stopped"))
	C.XSync(dpy, 1)

	// Stop the program.
	p.Kill()
}
