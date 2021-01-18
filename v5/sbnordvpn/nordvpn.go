// Package sbnordvpn displays the current status of the NordVPN connection, including the city and
// any connection errors.
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

// New makes a new routine object. colors is an optional triplet of hex color codes for colorizing
// the output based on these rules:
//   1. Normal color, VPN is connected.
//   2. Warning color, VPN is disconnected or is in the process of connecting.
//   3. Error color, error determining status, or network is down.
func New(colors ...[3]string) *Routine {
	var r Routine

	// Store the color codes. Don't do any validation.
	if len(colors) > 0 {
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
func (r *Routine) Update() (bool, error) {
	if r == nil {
		return false, fmt.Errorf("bad routine")
	}

	// If the command is successful but there's an error with nordvpn (like if the internet is
	// down), this will return an error code. We still want to capture and parse the error message,
	// so we're going to ignore any returned error.
	cmd := exec.Command("nordvpn", "status")
	output, _ := cmd.Output()

	if err := r.parseOutput(string(output)); err != nil {
		r.err = err
		return true, err
	}

	return true, nil
}

// String formats and prints the current connection status.
func (r *Routine) String() string {
	if r == nil {
		return "bad routine"
	}

	return r.color + r.parsed + colorEnd
}

// Error formats and returns an error message.
func (r *Routine) Error() string {
	if r == nil {
		return "bad routine"
	}

	if r.err == nil {
		r.err = fmt.Errorf("unknown error")
	}

	return r.colors.error + r.err.Error() + colorEnd
}

// Name returns the display name of this module.
func (r *Routine) Name() string {
	return "NordVPN"
}

// parseOutput parses the command's output.
func (r *Routine) parseOutput(output string) error {
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

	// Break out each word in the first line. It's possible that there is some garbage (mostly
	// unprintable characters) before the message, so we're going to scan the line until we find the
	// word "Status" and then try to determine the status by the word following that.
	fields := strings.Fields(lines[0])
	field := -1
	for i, v := range fields {
		if strings.HasPrefix(v, "Status") {
			field = i
			break
		}
	}
	if field == -1 {
		// We didn't receive the usual status output.
		if strings.Contains(lines[0], "Please check your internet connection") {
			return fmt.Errorf("internet down")
		}
		return fmt.Errorf(lines[0])
	} else if len(fields) <= field+1 {
		return fmt.Errorf("bad response")
	}

	if fields[field+1] == "Connected" {
		for _, line := range lines {
			if strings.HasPrefix(line, "City") {
				city := strings.Split(line, ":")
				if len(city) != 2 {
					return fmt.Errorf("error parsing City")
				}

				r.parsed = "Connected"
				r.parsed += r.getBlink()
				r.parsed += strings.TrimSpace(city[1])
				r.color = r.colors.normal
			}
		}
	} else {
		r.parsed = fields[field+1]
		r.color = r.colors.warning
	}

	return nil
}

func (r *Routine) getBlink() string {
	r.blink = !r.blink

	if r.blink {
		return ": "
	}
	return "  "
}
