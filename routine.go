// This file holds the business logic for a routine object.

package statusbar

import (
	"log"
	"time"
)

// routine holds the data for an individual unit on the statusbar.
type routine struct {
	// Routine object that handles running the actual process
	handler RoutineHandler

	// Name of routine
	name string

	// Time in seconds to wait between each run
	interval time.Duration

	// Channel to use for signaling manual update
	updateChan chan struct{}

	// Channel to use for signaling stop
	stopChan chan struct{}
}

// newRoutine returns a new routine object that is handled by handler.
func newRoutine() *routine {
	r := new(routine)

	// Set up the update and stop channels. We'll use a buffer size of 1 so the engine doesn't block sending on them.
	r.updateChan = make(chan struct{}, 1)
	r.stopChan = make(chan struct{}, 1)

	return r
}

// run runs a routine in a non-terminating loop. The routine's output is stored in index in the string slice received
// from outputsChan. If the routine does stop, it sends itself back on finished so the caller is aware.
func (r *routine) run(index int, outputsChan chan []string, finished chan<- *routine) {
	if r == nil {
		return
	}

	for {
		// Start the clock.
		start := time.Now()

		// Update the routine's data.
		ok, err := r.handler.Update()

		// Get the routine's output and store it in the master output slice.
		var output string
		if err == nil {
			output = r.handler.String()
		} else {
			output = r.handler.Error()
			log.Printf("%v: %v", r.handler.Name(), err.Error())
		}
		outputs := <-outputsChan
		outputs[index] = output
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
		// If the routine reported an error, then we'll give the process a little time to cool down before trying again.
		if err != nil {
			seconds := r.interval / time.Second
			switch {
			// For routines with intervals up to 1 minute, sleep for 5 seconds.
			case seconds < 60:
				interval = 5 * time.Second
			// For routines with intervals up to 15 minutes, sleep for 1 minute.
			case seconds < 60*15:
				interval = 60 * time.Second
			// For routines with intervals longer than 15 minutes, sleep for 5 minutes.
			default:
				interval = 60 * 5 * time.Second
			}
		}

		// Wait until either a signal is received from the engine or the time elapses for another update to run.
		select {
		case <-r.updateChan:
			// Update now.
			break
		case <-r.stopChan:
			// Stop the routine.
			finished <- r
			return
		case <-time.After(interval - time.Since(start)):
			// Time elapsed. Run another update loop.
			break
		}
	}

	// Send on the finished channel to signify that we're stopping this routine.
	finished <- r
}

// setHandler sets the routine's handler.
func (r *routine) setHandler(handler RoutineHandler) {
	if r != nil {
		r.handler = handler
	}
}

// setInterval sets the routine's interval in seconds.
func (r *routine) setInterval(interval int) {
	if r != nil {
		r.interval = time.Duration(interval) * time.Second
	}
}

// displayName returns the routine's display name.
func (r *routine) displayName() string {
	if r != nil {
		return r.handler.Name()
	}
	return "Unknown"
}

// moduleName returns the routine's module name.
func (r *routine) moduleName() string {
	if r != nil {
		return r.name
	}
	return ""
}

// setDisplayName sets the routine's display name.
func (r *routine) setModuleName(name string) {
	if r != nil {
		r.name = name
	}
}

// update refreshes the routine by calling Update.
func (r *routine) update() {
	// Update the routine by sending an empty struct on its update channel.
	if r != nil && r.updateChan != nil {
		r.updateChan <- struct{}{}
	}
}

// stop stops the routine.
func (r *routine) stop() {
	// Stop the routine by sending an empty struct on its stop channel.
	if r != nil && r.stopChan != nil {
		r.stopChan <- struct{}{}

	}
}
