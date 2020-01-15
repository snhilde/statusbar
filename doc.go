/*
Package statusbar displays various information on the dwm statusbar.

The design of this statusbar manager is modular by nature; you create the main process with this package and then populate it with only the information you want using separate modules. For example, if you want to show only the time and weather, then you would only import those two modules and add their objects to the statusbar, resulting in only the time and weather to appear on the statusbar for dwm. This modular design allows flexibility in customizing each individual statusbar and ease in not having to worry about supported libraries/dependencies.

This package is only the framework that controls the modular routines. For the modules currently integrated with this framework, see https://godoc.org/github.com/snhilde/sb4routines. There are currently modules that display a range of information, including various system resources, personal TODO lists, the current weather, VPN status, and the current time.

To integrate a custom module into this statusbar framework, the routine's object needs to implement these methods:
	Update()        // Update the routine's information. Will be run according to the provided interval time.
	String() string // Format and return the routine's output.
It is suggested that this object be created by New(), which will also initialize any members of the object (if needed).

This sample code will create a new statusbar, add some routines to it, and begin displaying the formatted output:
	import (
		"github.com/snhilde/statusbar4"
		"github.com/snhilde/sb4routines/sbtime"
		"github.com/snhilde/sb4routines/sbdisk"
		"github.com/snhilde/sb4routines/sbcpuusage"
		"github.com/snhilde/sb4routines/sbcputemp"
	)

	func main() {
		// Create the initial framework for our statusbar.
		bar := statusbar.New()

		// We will use the dualstatus patch for dwm to display output on a top and bottom bar.
		// The top bar will only display the time, according to the go format provided.
		// sbtime.New() takes two arguments: time format and a triplet of color codes for normal, warning,
		// and error outputs. It returns a new routine that implements the RoutineHandler interface.
		// bar.Append() takes two arguments: the new routine object and the update interval (how often in
		// seconds the routine should run its Update() method).
		bar.Append(sbtime.New("Jan 2 - 03:04", [3]string{"#FFFFFF", "#BB4F2E", "#A1273E"}), 1)

		// This will insert the splitting character for the dualstatus patch. After this, everything else
		// will be displayed on the second bar.
		bar.Split()

		// The second bar will start with the output from this disk routine. It will display the space used
		// and total space of the given filesystem. The routine will update every 5 seconds.
		bar.Append(sbdisk.New([]string{"/"}, [3]string{"#FFFFFF", "#BB4F2E", "#A1273E"}), 5)

		// The next two routines will display (separately) the current percentage of CPU used and the
		// temperature of the CPU, each updated every second.
		bar.Append(sbcpuusage.New([3]string{"#FFFFFF", "#BB4F2E", "#A1273E"}), 1)
		bar.Append(sbcputemp.New([3]string{"#8FFFFF", "#BB4F2E", "#A1273E"}), 1)

		// The statusbar will now run indefinitely, updating every routine at the provided interval.
		// All routines will run in their own thread and are independent of each other.
		bar.Run()
	}
*/
package statusbar
