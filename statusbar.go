// Package statusbar displays various resources on the DWM statusbar.
package statusbar

import (
	_ "fmt"
)

// Routine interface allows resource monitors to be linked in.
type Routine interface {
	Update() error
	String() string
	Sleep()
}

// Bar type is the main object for the package.
type Bar []Routine

// Creates a new Bar.
func New() Bar {
	var b Bar
	return b
}

// Appends a routine to the statusbar's list.
func (b *Bar) Append(r Routine) {
	*b = append(*b, r)
}

// Spins up every routine and displays them on the statusbar.
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

	// Wait for all routines to finish (shouldn't happen though).
	<-finished
}

// Run the routine in a non-terminating loop.
// TODO: handle errors
func runRoutine(r Routine, i int, ch chan []string) {
	for {
		// Update the contents of the routine.
		r.Update()

		// Get the routine's output and set it in the master output slice.
		output  := r.String()
		outputs := <-ch
		outputs[i] = output
		ch <- outputs

		// Put the routine to sleep.
		r.Sleep()
	}
}
