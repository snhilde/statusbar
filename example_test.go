package statusbar_test

import (
	"github.com/snhilde/sb4routines/sbtime"
)

func ExampleStatusbar_Append() {
	bar := statusbar.New()

	time_fmt := "Jan 2 - 03:04"

	normal_color := "#FFFFFF"
	warning_color := "#BB4F2E"
	error_color := "#A1273E"
	colors := [3]string{normal_color, warning_color, error_color}

	time_routine := sbtime.New(time_fmt, colors)

	bar.Append(time_routine, 1)
}
