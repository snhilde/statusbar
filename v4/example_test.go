package statusbar_test

import (
	"github.com/snhilde/statusbar4/sbtime"
)

func ExampleStatusbar_Append() {
	// Add the sbtime routine to our statusbar.

	// sbtime.New() takes two arguments: the format to use for the time string and a triplet of color codes.
	timeFmt := "Jan 2 - 03:04"

	colorNormal := "#FFFFFF"
	colorWarning := "#BB4F2E"
	colorError := "#A1273E"
	colors := [3]string{colorNormal, colorWarning, colorError}

	// Create a new routine.
	timeRoutine := sbtime.New(timeFmt, colors)

	// Append the routine to the bar.
	bar.Append(timeRoutine, 1)

	// Or, as a one-liner:
	bar.Append(sbtime.New("Jan 2 - 03:04", [3]string{"#FFFFFF", "#BB4F2E", "#A1273E"}), 1)
}
