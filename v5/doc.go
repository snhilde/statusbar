/*
Package statusbar displays various information on the dwm statusbar.

The design of this statusbar manager is modular by nature; you create the main process with this package and then
populate it with only the information you want using individual modules. For example, if you want to show only the time
and weather, then you would import only those two modules and add their objects to the statusbar, resulting in only the
time and weather to appear on the statusbar for dwm. This modular design allows flexibility in customizing each
individual statusbar and ease in management.

This package is the engine that controls the modular routines. For the modules currently integrated with this framework
and included in this repository, see the child packages that begin with "sb". There are currently modules that display a
range of information, including various system resources, personal TODO lists, the current weather, VPN status, and the
current time.

To integrate a custom module into this statusbar framework, the routine's object needs to implement the RoutineHandler
interface, which includes these methods:
	Update() (bool, error)
	String() string
	Error() string
	Name() string

It is suggested that this object be created by New(), which will also initialize any members of the object (if needed).

The sample code below creates a new statusbar, adds some routines to it, and begins displaying the formatted output. In
dwm, we are using the dualstatus patch, which creates a top and bottom bar for extra statusbar real estate. The top bar
will display the time, and the bottom bar will display the disk usage and CPU stats.

	import (
		"github.com/snhilde/statusbar/v5"
		"github.com/snhilde/statusbar/v5/sbtime"
		"github.com/snhilde/statusbar/v5/sbdisk"
		"github.com/snhilde/statusbar/v5/sbcpuusage"
		"github.com/snhilde/statusbar/v5/sbcputemp"
	)

	func main() {
		// Create the initial engine.
		bar := statusbar.New()

		// sbtime.New() takes two arguments: time format and a triplet of color codes for normal, warning, and error
		// outputs. It returns a new routine that implements the RoutineHandler interface.
		// bar.Append() takes two arguments: the new routine object and the update interval (how often in seconds the
		// routine should run its Update() method).
		bar.Append(sbtime.New("Jan 2 - 03:04", [3]string{"#FFFFFF", "#BB4F2E", "#A1273E"}), 1)

		// This inserts the splitting character for the dualstatus patch. Before this is called, the routines already
		// added are displayed on the top bar. After this is called, all subsequently added routines are displayed on
		// the bottom bar.
		bar.Split()

		// The second bar will start with the output from the disk routine. It will display the space used
		// and total space of the given filesystem. The routine will update every 5 seconds.
		bar.Append(sbdisk.New([]string{"/"}, [3]string{"#FFFFFF", "#BB4F2E", "#A1273E"}), 5)

		// The next two routines will display (separately) the current percentage of CPU used and the
		// temperature of the CPU, each updated every second.
		bar.Append(sbcpuusage.New([3]string{"#FFFFFF", "#BB4F2E", "#A1273E"}), 1)
		bar.Append(sbcputemp.New([3]string{"#8FFFFF", "#BB4F2E", "#A1273E"}), 1)

		// The statusbar will now run indefinitely, updating every routine at the provided interval. All routines run
		// concurrently in their own thread and are independent of each other.
		bar.Run()
	}
*/
package statusbar
