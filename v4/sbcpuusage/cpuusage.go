// Package sbcpuusage displays the percentage of CPU resources currently being used.
package sbcpuusage

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var colorEnd = "^d^"

// Routine is the main object for this package.
type Routine struct {
	// Error encountered along the way, if any.
	err error

	// Number of threads per CPU core.
	threads int

	// CPU stats from last read.
	oldStats stats

	// Percentage of CPU currently being used.
	perc int

	// Trio of user-provided colors for displaying various states.
	colors struct {
		normal  string
		warning string
		error   string
	}
}

// stats holds values of different CPU stats.
type stats struct {
	user int
	nice int
	sys  int
	idle int
}

// New gets current CPU stats and makes a new routine object. colors is an optional triplet of hex color codes for
// colorizing the output based on these rules:
//   1. Normal color, CPU is running at less than 75% of its capacity.
//   2. Warning color, CPU is running at between 75% and 90% of its capacity.
//   3. Error color, CPU is running at more than 90% of its capacity.
func New(colors ...[3]string) *Routine {
	var r Routine

	// Do a minor sanity check on the color codes.
	if len(colors) == 1 {
		for _, color := range colors[0] {
			if !strings.HasPrefix(color, "#") || len(color) != 7 {
				r.err = errors.New("invalid color")
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

	r.threads, r.err = numThreads()
	if r.err != nil {
		return &r
	}

	err := readFile(&(r.oldStats))
	if err != nil {
		r.err = err
	}

	return &r
}

// Update gets the current CPU stats, compares them to the last-read stats, and calculates the percentage of CPU
// currently being used.
func (r *Routine) Update() {
	var newStats stats

	err := readFile(&newStats)
	if err != nil {
		r.err = err
		return
	}

	used := (newStats.user - r.oldStats.user) + (newStats.nice - r.oldStats.nice) + (newStats.sys - r.oldStats.sys)
	total := (newStats.user - r.oldStats.user) + (newStats.nice - r.oldStats.nice) + (newStats.sys - r.oldStats.sys) + (newStats.idle - r.oldStats.idle)
	total *= r.threads

	// Prevent divide-by-zero error
	if total == 0 {
		r.perc = 0
	} else {
		r.perc = (used * 100) / total
		if r.perc < 0 {
			r.perc = 0
		} else if r.perc > 100 {
			r.perc = 100
		}
	}

	r.oldStats.user = newStats.user
	r.oldStats.nice = newStats.nice
	r.oldStats.sys = newStats.sys
	r.oldStats.idle = newStats.idle
}

// String prints the formatted CPU percentage.
func (r *Routine) String() string {
	var c string

	if r.err != nil {
		return r.colors.error + r.err.Error() + colorEnd
	}

	if r.perc < 75 {
		c = r.colors.normal
	} else if r.perc < 90 {
		c = r.colors.warning
	} else {
		c = r.colors.error
	}

	return fmt.Sprintf("%s%2d%% CPU%s", c, r.perc, colorEnd)
}

// readFile opens /proc/stat and reads out the CPU stats from the first line.
func readFile(newStats *stats) error {
	// The first line of /proc/stat will look like this:
	// "cpu userVal niceVal sysVal idleVal ..."
	f, err := os.Open("/proc/stat")
	if err != nil {
		return err
	}
	defer f.Close()

	reader := bufio.NewReader(f)

	line, err := reader.ReadString('\n')
	if err != nil {
		return err
	}

	// Error will be handled in String().
	_, err = fmt.Sscanf(line, "cpu %v %v %v %v", &(newStats.user), &(newStats.nice), &(newStats.sys), &(newStats.idle))
	return err
}

// The shell command 'lscpu' will return a variety of CPU information, including the number of threads
// per CPU core. We don't care about the number of cores, because we're already reading in the
// averaged total. We only want to know if we need to be changing its range. To get this number, we're
// going to loop through each line of the output until we find "Thread(s) per socket".
func numThreads() (int, error) {
	proc := exec.Command("lscpu")
	out, err := proc.Output()
	if err != nil {
		return -1, err
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Thread(s) per core") {
			fields := strings.Fields(line)
			if len(fields) != 4 {
				return -1, errors.New("invalid fields")
			}
			return strconv.Atoi(fields[3])
		}
	}

	// If we made it this far, then we didn't find anything.
	return -1, errors.New("failed to find number of threads")
}
