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

	// List of interfaces.
	ilist []sbiface

	// Trio of user-provided colors for displaying various states.
	colors struct {
		normal  string
		warning string
		error   string
	}
}

// sbiface groups different pieces of information for a single interface.
type sbiface struct {
	// Name of interface.
	name string

	// Path to rx_bytes file.
	downPath string

	// Path to tx_bytes file.
	upPath string

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

	var ilist []string
	var err error
	if len(inames) == 0 {
		// Nothing was passed in. We'll grab the default interfaces.
		ilist, err = getInterfaces()
	} else {
		for _, iname := range inames {
			// Make sure we have a valid interface name.
			_, err = net.InterfaceByName(iname)
			if err != nil {
				err = errors.New(iname + ": " + err.Error())
				break
			}
			ilist = append(ilist, iname)
		}
	}

	// Handle any problems that came up, or build up list of interfaces for later use.
	if err != nil {
		r.err = err
		r.ilist = nil
	} else if len(ilist) == 0 {
		r.err = errors.New("No interfaces found")
	} else {
		for _, iname := range ilist {
			downPath := "/sys/class/net/" + iname + "/statistics/rx_bytes"
			upPath := "/sys/class/net/" + iname + "/statistics/tx_bytes"
			r.ilist = append(r.ilist, sbiface{name: iname, downPath: downPath, upPath: upPath})
		}
	}

	return &r
}

// Update gets the current readings of the rx/tx files for each interface.
func (r *Routine) Update() (bool, error) {
	// Handle any error from New.
	if len(r.ilist) == 0 {
		return false, r.err
	}

	for i, iface := range r.ilist {
		r.ilist[i].oldDown = iface.newDown
		r.ilist[i].oldUp = iface.newUp

		down, err := readFile(iface.downPath)
		if err != nil {
			r.err = errors.New("Error reading " + iface.name + " (Down)")
			return true, err
		}
		r.ilist[i].newDown = down

		up, err := readFile(iface.upPath)
		if err != nil {
			r.err = errors.New("Error reading " + iface.name + " (Up)")
			return true, err
		}
		r.ilist[i].newUp = up
	}

	return true, nil
}

// String calculates the byte difference for each interface, and formats and prints it.
func (r *Routine) String() string {
	var c string
	var b strings.Builder

	for i, iface := range r.ilist {
		down, downUnit := shrink(iface.newDown - iface.oldDown)
		up, upUnit := shrink(iface.newUp - iface.oldUp)

		if downUnit == 'B' || upUnit == 'B' || downUnit == 'K' || upUnit == 'K' {
			c = r.colors.normal
		} else if downUnit == 'M' || upUnit == 'M' {
			c = r.colors.warning
		} else {
			c = r.colors.error
		}

		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(c)
		fmt.Fprintf(&b, "%s: %4v%c↓|%4v%c↑", iface.name, down, downUnit, up, upUnit)
		b.WriteString(colorEnd)
	}

	return b.String()
}

// Error formats and returns an error message.
func (r *Routine) Error() string {
	if r.err == nil {
		r.err = errors.New("Unknown error")
	}

	return r.colors.error + r.err.Error() + colorEnd
}

// Name returns the display name of this module.
func (r *Routine) Name() string {
	return "Network"
}

// getInterfaces finds all network interfaces that are currently active.
func getInterfaces() ([]string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	inames := make([]string, 0)
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
