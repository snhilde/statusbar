// Package sbcpuusage displays the percentage of CPU resources currently being used.
package sbcpuusage

import (
	"errors"
	"strings"
	"os/exec"
	"strconv"
	"fmt"
	"os"
	"bufio"
)

var COLOR_END = "^d^"

// routine is the main object for this package.
// err:       error encountered along the way, if any
// old_stats: CPU stats from last read
// perc:      percentage of CPU currently being used
// colors:    trio of user-provided colors for displaying various states
type routine struct {
	err       error
	threads   int
	old_stats stats
	perc      int
	colors    struct {
		normal  string
		warning string
		error   string
	}
}

// Type to hold values of different CPU stats
type stats struct {
	user int
	nice int
	sys  int
	idle int
}

// Get current CPU stats and return routine object.
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
		r.colors.normal  = "^c" + colors[0][0] + "^"
		r.colors.warning = "^c" + colors[0][1] + "^"
		r.colors.error   = "^c" + colors[0][2] + "^"
	} else {
		// If a color array wasn't passed in, then we don't want to print this.
		COLOR_END = ""
	}

	r.threads, r.err = numThreads()
	if r.err != nil {
		return &r
	}

	err := readFile(&(r.old_stats))
	if err != nil {
		r.err = err
	}

	return &r
}

// Get current CPU stats, compare to last-read stats, and calculate percentage of CPU being used.
func (r *routine) Update() {
	var new_stats stats

	err := readFile(&new_stats)
	if err != nil {
		r.err = err
		return
	}

	used  := (new_stats.user-r.old_stats.user) + (new_stats.nice-r.old_stats.nice) + (new_stats.sys-r.old_stats.sys)
	total := (new_stats.user-r.old_stats.user) + (new_stats.nice-r.old_stats.nice) + (new_stats.sys-r.old_stats.sys) + (new_stats.idle-r.old_stats.idle)
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

	r.old_stats.user = new_stats.user
	r.old_stats.nice = new_stats.nice
	r.old_stats.sys  = new_stats.sys
	r.old_stats.idle = new_stats.idle
}

// Print formatted CPU percentage.
func (r *routine) String() string {
	var c string

	if r.err != nil {
		return r.colors.error + r.err.Error() + COLOR_END
	}

	if r.perc < 75 {
		c = r.colors.normal
	} else if r.perc < 90 {
		c = r.colors.warning
	} else {
		c = r.colors.error
	}

	return fmt.Sprintf("%s%2d%% CPU%s", c, r.perc, COLOR_END)
}

// Open /proc/stat and read out the CPU stats from the first line.
func readFile(new_stats *stats) error {
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
	_, err = fmt.Sscanf(line, "cpu %v %v %v %v", &(new_stats.user), &(new_stats.nice), &(new_stats.sys), &(new_stats.idle))
	return err
}

// The shell command 'lscpu' will return a variety of CPU information, including the number of threads
// per CPU core. We don't care about the number of cores, because we're already reading in the
// averaged total. We only want to know if we need to be changing its range. To get this number, we're
// going to loop through each line of the output until we find "Thread(s) per socket".
func numThreads() (int, error) {
	proc     := exec.Command("lscpu")
	out, err := proc.Output()
	if err != nil {
		return -1, err
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Thread(s) per core") {
			fields := strings.Fields(line)
			if len(fields) != 4 {
				return -1, errors.New("Invalid fields")
			}
			return strconv.Atoi(fields[3])
		}
	}

	// If we made it this far, then we didn't find anything.
	return -1, errors.New("Failed to find number of threads")
}
