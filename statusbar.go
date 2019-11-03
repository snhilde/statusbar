// Package statusbar displays various resources on the DWM statusbar.
package statusbar

import (
	"fmt"
	"time"
	"string"
)

// Routine interface allows resource monitors to be linked in.
type Routine interface {
	Update() error
	String() string
	Sleep(time.Duration)
}

// A Bar holds all the routines, in the order specified.
type Bar []Routine

// Create a new Bar.
func New() Bar {
	var b Bar
	return b
}

// Append a routine to the statusbar's list.
func (b *Bar) Append(r Routine) {
	*b = append(*b, r)
}

// Spin up every routine and display them on the statusbar.
func (b *Bar) Run() {
	// Shared channel used to pass the slice of outputs
	ch := make(chan []string)

	// A slice of strings to hold the output from each routine
	outputs := make([]string, len(*b))
	ch <- outputs

	// Channel used to indicate everything is done
	// TODO: currently unused
	finished := make(chan error)

	for i, r := range *b {
		go runRoutine(r, i, ch)
	}

	// Launch a goroutine to build and print the master string.
	go buildBar(ch)

	// Wait for all routines to finish (shouldn't happen though).
	<-finished
}

// Run the routine in a non-terminating loop.
// TODO: handle errors
func runRoutine(r Routine, i int, ch chan []string) {
	for {
		// Start the clock.
		start := time.Now()

		// Update the contents of the routine.
		r.Update()

		// Get the routine's output and set it in the master output slice.
		output  := r.String()
		outputs := <-ch
		outputs[i] = output
		ch <- outputs

		// Stop the clock and put the routine to sleep.
		end := time.Now()
		r.Sleep(end.Sub(start))
	}
}

// Build the master output and print in to the statusbar. Runs a loop every second.
func buildBar(ch chan []string) {
	var b strings.Builder

	for {
		// Start the clock.
		start := time.Now()

		// Receive the outputs slice and build the individual outputs into a master output.
		outputs := <-ch
		for _, s := range outputs {
			fmt.Fprintf(&b, "[%s] ", s)
		}
		ch <- outputs

		// Print the master output to the statusbar.
		printBar(b.String())
		b.Reset()

		// Stop the clock and put the routine to sleep for the rest of the second.
		end := time.Now()
		time.Sleep(time.Second - end.Sub(start))
	}
}
