// Package sbnordvpn displays the current status of the NordVPN connection, including the city and any connection errors.
package sbnordvpn

import (
	"errors"
	"os/exec"
	"strings"
)

var colorEnd = "^d^"

// routine is the main object in the package.
// err:    error encountered along the way, if any
// b:      buffer to hold connnection string
// color:  current color of the 3 provided
// colors: trio of user-provided colors for displaying various states
type routine struct {
	s      string
	err    error
	blink  bool
	color  string
	colors struct {
		normal  string
		warning string
		error   string
	}
}

// Return a new routine object.
func New(colors ...[3]string) *routine {
	var r routine

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

	return &r
}

// Run the command and capture the output.
func (r *routine) Update() {
	cmd := exec.Command("nordvpn", "status")
	out, err := cmd.Output()
	if err != nil {
		r.err = err
		return
	}

	r.s, r.err = r.parseCommand(string(out))
}

// Format and print the current connection status.
func (r *routine) String() string {
	if r.err != nil {
		return r.colors.error + "NordVPN: " + r.err.Error() + colorEnd
	}

	return r.color + r.s + colorEnd
}

// Parse the command's output.
func (r *routine) parseCommand(s string) (string, error) {
	// If there is a connection to the VPN, the command will return this format:
	//     Status: Connected
	//     Current server: <server.url>
	//     Country: <country>
	//     City: <city>
	//     Your new IP: <the.new.IP.address>
	//     Current technology: <tech>
	//     Current protocol: <protocol>
	//     Transfer: <bytes> <unit> received, <bytes> <unit> sent
	//     Uptime: <human-readable time>
	//
	// If there is no connection, the command will return this:
	//     Status: Disconnected
	//
	// If there is no Internet connection, the command will return this:
	//     Please check your internet connection and try again.

	// Split up all the lines of the output for parsing.
	lines := strings.Split(s, "\n")

	// Break out each word in the first line. It's possible that there is some garbage (mostly unprintable characters)
	// before the message, so we're going to scan the line until we find the word "Status" and then try to determine the
	// status by the word following that.
	fields := strings.Fields(lines[0])
	field := -1
	for i, v := range fields {
		if strings.HasPrefix(v, "Status") {
			field = i
			break
		}
	}
	if field == -1 {
		return "", errors.New(lines[0])
	}

	switch fields[field+1] {
	case "Connected":
		for _, line := range lines {
			if strings.HasPrefix(line, "City") {
				city := strings.Split(line, ":")
				if len(city) != 2 {
					return "", errors.New("Error parsing City")
				}

				s := "Connected"
				if r.blink {
					r.blink = false
					s += ": "
				} else {
					r.blink = true
					s += "  "
				}
				s += strings.TrimSpace(city[1])
				r.color = r.colors.normal
				return s, nil
			}
		}
	case "Connecting":
		r.color = r.colors.warning
		return "Connecting...", nil
	case "Disconnected":
		r.color = r.colors.warning
		return "Disconnected", nil
	case "Please check your internet connection and try again.":
		return "", errors.New("Internet Down")
	}

	// If we're here, then we have an unknown error.
	r.color = r.colors.error
	return "", errors.New(lines[0])
}
