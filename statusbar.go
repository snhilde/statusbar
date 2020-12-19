// Package statusbar formats and displays information on the dwm statusbar by managing modular data routines.
package statusbar

// #cgo pkg-config: x11
// #cgo LDFLAGS: -lX11
// #include <X11/Xlib.h>
import "C"

import (
	"github.com/snhilde/statusbar/restapi"
	"log"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"time"
)

// RoutineHandler allows information monitors (commonly called routines) to be linked in.
type RoutineHandler interface {
	// Update updates the routine's information. This is run on a periodic interval according to the time provided when
	// adding the routine to the statusbar engine. It returns two arguments: a bool and an error. The bool indicates
	// whether or not the engine should continue to run the routine. You can think of it as representing the "ok"
	// status. The error is any error encountered during the process. For example, on a normal run with no error, Update
	// would return (true, nil). On a run with a non-critical error, Update would return (true, error). On a run with a
	// critical error where the routine should be stopped, Update would return (false, error). The returned error will
	// be logged to stderr. Generally, the error returned from Update should be detailed and specific for debugging the
	// routine, while the error returned from Error should be shorter, more concise, and more general.
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

// Statusbar is the main type for this package. It holds information about the bar as a whole.
type Statusbar struct {
	// List of routines, in the order they were added.
	routines []*routine

	// Delimiter to use for the left side of each routine's output, as set with SetMarkers.
	leftDelim string

	// Delimiter to use for the right side of each routine's output, as set with SetMarkers.
	rightDelim string

	// Index of the routine after which the routines are split, as set with Split.
	split int

	// Timer that is started when the statusbar is started. This is used to measure the statusbar's uptime.
	startTime time.Time

	// Whether or not to enable and run the APIs.
	apiEnabled bool
}

// New creates a new statusbar. The default delimiters around each routine are square brackets ('[' and ']').
func New() Statusbar {
	return Statusbar{leftDelim: "[", rightDelim: "]", split: -1}
}

// Append adds a routine to the statusbar's list. Routines are displayed in the order they are added. handler is the
// RoutineHandler module. seconds is the amount of time between each run of the routine.
func (sb *Statusbar) Append(handler RoutineHandler, seconds int) {
	// Convert the given number into proper seconds.
	r := newRoutine()
	r.setHandler(handler)
	r.setInterval(seconds)

	// Get the package name of the module that is implementing this RoutineHandler. We are going to use this to match
	// the routine's name for the API. TypeOf returns "*{package}.Routine", like "*sbbattery.Routine". We want to
	// capture only the package name.
	refType := reflect.TypeOf(handler).String()
	if fields := strings.Split(refType, "."); len(fields) == 2 {
		module := strings.TrimPrefix(fields[0], "*")
		r.setModuleName(module)
	} else {
		if refType == "" {
			refType = "unknown"
		}
		log.Printf("Failed to determine package name (%s)", refType)
	}

	sb.routines = append(sb.routines, r)
}

// Run spins up all the routines and displays them on the statusbar. If the APIs are enabled, this also runs the API
// engines.
func (sb *Statusbar) Run() {
	// Start the uptime clock.
	sb.startTime = time.Now()

	// Add a signal handler so we can clear the statusbar if the program goes down.
	go sb.handleSignal()

	// A slice of strings to hold the output from each routine
	outputs := make([]string, len(sb.routines))

	// Shared channel used to pass the slice of outputs
	outputsChan := make(chan []string, 1)
	outputsChan <- outputs

	// Set up a channel used to indicate everything is done. This must have a buffer large enough for every channel to
	// send on without blocking.
	finished := make(chan *routine, len(sb.routines))

	// Run each routine.
	for i, v := range sb.routines {
		go v.run(i, outputsChan, finished)
	}

	// Launch a goroutine to build and print the master string.
	go setBar(outputsChan, *sb)

	// If enabled, build and run the APIs in their own goroutine.
	if sb.apiEnabled {
		go sb.runAPIs()
	}

	// Keep running until every routine stops.
	for i := 0; i < len(sb.routines); i++ {
		r := <-finished
		log.Printf("%v: Routine stopped", r.displayName())
	}

	log.Printf("All routines have stopped")
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

// Uptime returns the time in seconds denoting how long the statusbar has been running.
func (sb *Statusbar) Uptime() int {
	t := time.Since(sb.startTime)
	return int(t.Seconds())
}

// EnableAPI enables the engine to run the APIs on port port. These can be used to interact with the statusbar and its routines while
// they are running.
func (sb *Statusbar) EnableAPI(port int) {
	sb.apiEnabled = true
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
					// If the output ends with the color terminator, then we need to make sure to keep that so the color
					// doesn't bleed onto the delimiter and beyond.
					hasColor := strings.HasSuffix(s, "^d^")
					s = s[:46] + "..."
					if hasColor {
						s += "^d^"
					}
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
	log.Printf("Received interrupt")

	dpy := C.XOpenDisplay(nil)
	root := C.XDefaultRootWindow(dpy)
	C.XStoreName(dpy, root, C.CString("Statusbar stopped"))
	C.XSync(dpy, 1)

	// Stop the program.
	p.Kill()
}

// runAPIs runs the various APIs and their versions using the callback methods implemented by handler. New APIs/versions
// should be added here.
func (sb *Statusbar) runAPIs() {
	// Begin with the REST API.
	r := restapi.NewEngine()

	// Spin up REST API v1. Use an apiHandler to wrap the statusbar object for convenience (see type definition).
	if err := r.AddSpecFile("api_specs/restv1.json", apiHandler{sb}); err != nil {
		log.Printf("Error building REST API v1: %s", err.Error())
	} else {
		r.Run(3939)
	}
}
