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
