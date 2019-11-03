// Package statusbar displays various resources on the DWM statusbar.
package statusbar

import (
	"fmt"
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
	// A slice of strings to hold the output from each routine
	outputs := make([]string, len(*b)

	// Shared channel used to pass the slice of outputs
	ch := make(chan []string)

	// Channel used to indicate everything is done
	// TODO: currently unused
	finished := make(chan error)

	for i, r := range b {
		go runRoutine(r, i, ch)
	}

	// Wait for all routines to finish (shouldn't happen though).
	<- finished
}
