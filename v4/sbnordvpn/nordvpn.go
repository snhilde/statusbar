// Package sbnordvpn displays the current status of the NordVPN connection, including the city and any connection errors.
package sbnordvpn

import (
	"fmt"
	"os/exec"
	"strings"
)

var colorEnd = "^d^"

// Routine is the main object in the package.
type Routine struct {
	// Error encountered along the way, if any.
	err error

	// Parsed and formatted output string.
	parsed string

	// Buffer to hold connnection string.
	blink bool

	// Current color of the 3 provided.
	color string

	// Trio of user-provided colors for displaying various states.
	colors struct {
		normal  string
		warning string
		error   string
	}
}

// New makes a new routine object. colors is an optional triplet of hex color codes for colorizing the output based on
// these rules:
//   1. Normal color, VPN is connected.
//   2. Warning color, VPN is disconnected or is in the process of connecting.
//   3. Error color, error determining status, or network is down.
func New(colors ...[3]string) *Routine {
	var r Routine

	// Do a minor sanity check on the color codes.
	if len(colors) == 1 {
		for _, color := range colors[0] {
			if !strings.HasPrefix(color, "#") || len(color) != 7 {
				r.err = fmt.Errorf("invalid color")
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

// Update runs the command and captures the output.
func (r *Routine) Update() {
	cmd := exec.Command("nordvpn", "status")
	output, err := cmd.Output()
	if err != nil {
		r.err = err
		return
	}

	r.parsed, r.err = r.parseOutput(string(output))
}

// String formats and prints the current connection status.
func (r *Routine) String() string {
	if r.err != nil {
		return r.colors.error + "NordVPN: " + r.err.Error() + colorEnd
	}

	return r.color + r.parsed + colorEnd
}

// parseOutput parses the command's output.
func (r *Routine) parseOutput(output string) (string, error) {
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
	lines := strings.Split(output, "\n")

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
		return "", fmt.Errorf(lines[0])
	}

	switch fields[field+1] {
	case "Connected":
		for _, line := range lines {
			if strings.HasPrefix(line, "City") {
				city := strings.Split(line, ":")
				if len(city) != 2 {
					return "", fmt.Errorf("error parsing city")
				}

				parsed := "Connected"
				if r.blink {
					r.blink = false
					parsed += ": "
				} else {
					r.blink = true
					parsed += "  "
				}
				parsed += strings.TrimSpace(city[1])
				r.color = r.colors.normal
				return parsed, nil
			}
		}
	case "Connecting":
		r.color = r.colors.warning
		return "Connecting...", nil
	case "Disconnected":
		r.color = r.colors.warning
		return "Disconnected", nil
	case "Please check your internet connection and try again.":
		return "", fmt.Errorf("internet down")
	}

	// If we're here, then we have an unknown error.
	r.color = r.colors.error
	return "", fmt.Errorf(lines[0])
}
