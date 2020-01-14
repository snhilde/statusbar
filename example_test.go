package statusbar_test

import (
	"github.com/snhilde/sb4routines/sbtime"
)

func ExampleStatusbar_Append() {
	// Add the sbtime routine to our statusbar.

	// sbtime.New() takes two arguments: format to use for the time string and triplet of color codes.
	time_fmt := "Jan 2 - 03:04"

	normal_color := "#FFFFFF"
	warning_color := "#BB4F2E"
	error_color := "#A1273E"
	colors := [3]string{normal_color, warning_color, error_color}

	// Create a new routine.
	time_routine := sbtime.New(time_fmt, colors)

	// Append the routine to the bar. Routines will be displayed in order of addition to the bar object.
	bar.Append(time_routine, 1)

	// Or, as a one-liner:
	bar.Append(sbtime.New("Jan 2 - 03:04", [3]string{"#FFFFFF", "#BB4F2E", "#A1273E"}), 1)
}
