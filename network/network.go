// Package sbnetwork displays the number of bytes sent and received per given time period for either the provided
// network interfaces or the ones currently marked as active.
package sbnetwork

import (
	"errors"
	"strings"
	"net"
	"io/ioutil"
	"strconv"
	"fmt"
)

var COLOR_END = "^d^"

// routine is the main object for this package.
// err:    error encountered along the way, if any
// ilist:  list of interfaces
// colors: trio of user-provided colors for displaying various states
type routine struct {
	err      error
	ilist  []sbiface
	colors   struct {
		normal  string
		warning string
		error   string
	}
}

// sbiface groups different pieces of information for a single interface.
// name:      name of interface
// down_path: path to rx_bytes file
// up_path:   path to tx_bytes file
// old_down:  last reading of rx_bytes file
// old_up:    last reading of tx_bytes file
// new_down:  current reading of rx_bytes file
// new_up:    current reading of tx_bytes file
type sbiface struct {
	name      string
	down_path string
	up_path   string
	old_down  int
	old_up    int
	new_down  int
	new_up    int
}

// Return a new routine object populated with either the given interfaces or the active ones.
func New(inames []string, colors ...[3]string) *routine {
	var r       routine
	var ilist []string
	var err     error

	if len(inames) == 0 {
		// Nothing was passed in. We'll grab the default interfaces.
		ilist, err = getInterfaces()
	} else {
		for _, iname := range inames {
			// Make sure we have a valid interface name.
			_, err = net.InterfaceByName(iname)
			if err != nil {
				// Error will be handled in Update() and String().
				err = errors.New(iname + ": " + err.Error())
				break
			}
			ilist = append(ilist, iname)
		}
	}

	// Handle any problems that came up, or build up list of interfaces for later use.
	if err != nil {
		r.err = err
	} else if len(ilist) == 0 {
		r.err = errors.New("No interfaces found")
	} else {
		for _, iname := range ilist {
			down_path := "/sys/class/net/" + iname + "/statistics/rx_bytes"
			up_path   := "/sys/class/net/" + iname + "/statistics/tx_bytes"
			r.ilist = append(r.ilist, sbiface{name: iname, down_path: down_path, up_path: up_path})
		}
	}

	// Do a minor sanity check on the color codes.
	if len(colors) == 1 {
		for _, color := range colors[0] {
			if !strings.HasPrefix(color, "#") || len(color) != 7 {
				r.err = errors.New("Invalid color")
				return &r
			}
		}
		r.colors.normal  = "^c" + colors[0][0] + "^"
		r.colors.warning = "^c" + colors[0][1] + "^"
		r.colors.error   = "^c" + colors[0][2] + "^"
	} else {
		// If a color array wasn't passed in, then we don't want to print this.
		COLOR_END = ""
	}

	return &r
}

// Get the current readings of the rx/tx files for each interface.
func (r *routine) Update() {
	for i, iface := range r.ilist {
		r.ilist[i].old_down = iface.new_down
		r.ilist[i].old_up   = iface.new_up

		down, err := readFile(iface.down_path)
		if err != nil {
			// r.err = err
			continue
		}
		r.ilist[i].new_down = down

		up, err := readFile(iface.up_path)
		if err != nil {
			// r.err = err
			continue
		}
		r.ilist[i].new_up = up
	}
}

// Calculate the byte difference for each interface, and format and print it.
func (r *routine) String() string {
	var c string
	var b strings.Builder

	if r.err != nil {
		return r.colors.error + r.err.Error() + COLOR_END
	}

	for i, iface := range r.ilist {
		down, down_u := shrink(iface.new_down - iface.old_down)
		up, up_u     := shrink(iface.new_up   - iface.old_up)

		if down_u == 'B' || up_u == 'B' || down_u == 'K' || up_u == 'K' {
			c = r.colors.normal
		} else if down_u == 'M' || up_u == 'M' {
			c = r.colors.warning
		} else {
			c = r.colors.error
		}

		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(c)
		fmt.Fprintf(&b, "%s: %4v%c↓/%4v%c↑", iface.name, down, down_u, up, up_u)
		b.WriteString(COLOR_END)
	}

	return b.String()
}

// Find all network interfaces that are currently active.
func getInterfaces() ([]string, error) {
	var inames []string

	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

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

// Read out the contents of the given file.
func readFile(path string) (int, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return -1, err
	}

	return strconv.Atoi(strings.TrimSpace(string(b)))
}

// Iteratively decrease the amount of bytes by a step of 2^10 until human-readable.
func shrink(bytes int) (int, rune) {
	var units = [...]rune{'B', 'K', 'M', 'G', 'T', 'P', 'E'}
	var i int

	for bytes > 1024 {
		bytes >>= 10
		i++
	}

	return bytes, units[i]
}
