// Package sbnetwork displays the number of bytes sent and received per given time period for either the provided
// network interfaces or the ones currently marked as active.
package sbnetwork

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"strconv"
	"strings"
)

var colorEnd = "^d^"

// Routine is the main object for this package.
type Routine struct {
	// Error encountered along the way, if any.
	err error

	// List of user-supplied interfaces to monitor. If nothing was supplied, we'll grab the interfaces currently up.
	givenNames []string

	// List of interfaces names that we want to display on the statusbar.
	printNames []string

	// Cache of data for every interface monitored.
	cache map[string]sbiface

	// Trio of user-provided colors for displaying various states.
	colors struct {
		normal  string
		warning string
		error   string
	}
}

// sbiface groups different pieces of information for a single interface.
type sbiface struct {
	// Whether the interface is currently up or down.
	enabled bool

	// Last reading of rx_bytes file.
	oldDown int

	// Last reading of tx_bytes file.
	oldUp int

	// Current reading of rx_bytes file.
	newDown int

	// Current reading of tx_bytes file.
	newUp int
}

// New returns a new routine object populated with either the given interfaces or the active ones if no interfaces are
// specified. colors is an optional triplet of hex color codes for colorizing the output based on these rules:
//   1. Normal color, all interfaces are running at Kpbs speeds or less.
//   2. Warning color, one of more interface is running at Mbps speeds.
//   3. Error color, one of more interface is running at greater than Mbps speeds.
func New(inames []string, colors ...[3]string) *Routine {
	var r Routine

	// Do a minor sanity check on the color codes.
	if len(colors) == 1 {
		for _, color := range colors[0] {
			if !strings.HasPrefix(color, "#") || len(color) != 7 {
				r.err = errors.New("Invalid color")
				return &r
			}
		}
		r.colors.normal = "^c" + colors[0][0] + "^"
		r.colors.warning = "^c" + colors[0][1] + "^"
		r.colors.error = "^c" + colors[0][2] + "^"
	} else {
		// If a color array wasn't passed in, then we don't want to print this.
		colorEnd = ""
	}

	r.givenNames = inames
	r.cache = make(map[string]sbiface)

	return &r
}

// Update gets the current readings of the rx/tx files for each interface.
func (r *Routine) Update() (bool, error) {
	if r == nil {
		return false, errors.New("Bad routine")
	}

	// Get the interfaces that we want to monitor on this loop.
	r.printNames = r.givenNames
	if len(r.printNames) == 0 {
		// If no interfaces were specified, then we'll grab all the ones currently up. We want to run this process each
		// loop to catch any changes in interface statuses as they happen.
		is, err := findInterfaces()
		if err != nil {
			r.err = errors.New("Error finding interfaces")
			return true, err
		}
		r.printNames = is
	}
	if len(r.printNames) == 0 {
		r.err = errors.New("No interfaces up")
		return true, r.err
	}

	// Get the new data for each monitored interface.
	for _, iname := range r.printNames {
		iface := r.cache[iname]

		iface.oldDown = iface.newDown
		iface.oldUp = iface.newUp

		downPath := "/sys/class/net/" + iname + "/statistics/rx_bytes"
		down, err := readFile(downPath)
		if err != nil {
			iface.enabled = false
			r.cache[iname] = iface
			continue
		}
		iface.newDown = down

		upPath := "/sys/class/net/" + iname + "/statistics/tx_bytes"
		up, err := readFile(upPath)
		if err != nil {
			iface.enabled = false
			r.cache[iname] = iface
			continue
		}
		iface.newUp = up

		iface.enabled = true
		r.cache[iname] = iface
	}

	return true, nil
}

// String calculates the byte difference for each interface, and formats and prints it.
func (r *Routine) String() string {
	if r == nil {
		return "Bad routine"
	}

	var c string
	var b strings.Builder
	for _, iname := range r.printNames {
		iface, ok := r.cache[iname]
		if !ok {
			continue
		}

		if b.Len() > 0 {
			b.WriteString(", ")
		}

		if iface.enabled {
			down, downUnit := shrink(iface.newDown - iface.oldDown)
			up, upUnit := shrink(iface.newUp - iface.oldUp)

			if downUnit == 'B' || upUnit == 'B' || downUnit == 'K' || upUnit == 'K' {
				c = r.colors.normal
			} else if downUnit == 'M' || upUnit == 'M' {
				c = r.colors.warning
			} else {
				c = r.colors.error
			}

			b.WriteString(c)
			fmt.Fprintf(&b, "%v: %4v%c↓|%4v%c↑", iname, down, downUnit, up, upUnit)
			b.WriteString(colorEnd)
		} else {
			b.WriteString(r.colors.error)
			fmt.Fprintf(&b, "%v: Down", iname)
			b.WriteString(colorEnd)
		}
	}

	return b.String()
}

// Error formats and returns an error message.
func (r *Routine) Error() string {
	if r == nil {
		return "Bad routine"
	}

	if r.err == nil {
		r.err = errors.New("Unknown error")
	}

	return r.colors.error + r.err.Error() + colorEnd
}

// Name returns the display name of this module.
func (r *Routine) Name() string {
	return "Network"
}

// findInterfaces finds all network interfaces that are currently active.
func findInterfaces() ([]string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var inames []string
	for _, iface := range ifaces {
		if iface.Name == "lo" {
			// Skip loopback.
			continue
		} else if !strings.Contains(iface.Flags.String(), "up") {
			// If the network is not up, then we don't need to monitor it.
			continue
		}
		inames = append(inames, iface.Name)
	}

	return inames, nil
}

// readFile reads out the contents of the given file.
func readFile(path string) (int, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return -1, err
	}

	return strconv.Atoi(strings.TrimSpace(string(b)))
}

// shrink iteratively decreases the amount of bytes by a step of 2^10 until human-readable.
func shrink(bytes int) (int, rune) {
	var units = [...]rune{'B', 'K', 'M', 'G', 'T', 'P', 'E'}
	var i int

	for bytes > 1024 {
		bytes >>= 10
		i++
	}

	return bytes, units[i]
}
