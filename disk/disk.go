// Package sbdisk displays disk resources for each filesystem provided.
package sbdisk

import (
	"errors"
	"strings"
	"syscall"
	"fmt"
)

var COLOR_END = "^d^"

// routine is the main object for this package.
// err:    error encountered along the way, if any
// disks:  slice of provided filesystems to stat
// colors: trio of user-provided colors for displaying various states
type routine struct {
	err      error
	disks  []fs
	colors   struct {
		normal  string
		warning string
		error   string
	}
}

// fs holds information about a single filesystem.
// path:    given path that will be used to stat the partition
// avail:   used bytes for this filesystem
// avail_u: unit for the used bytes
// total:   total bytes for this filesystem
// total_u: unit for the total bytes
// perc:    percentage of total disk space used
type fs struct {
	path    string
	used    uint64
	used_u  rune
	total   uint64
	total_u rune
	perc    uint64
	// Note: Bavail is the amount of blocks that can actually be used, while
	// Bfree is the total amount of unused blocks.
}

// Copy over the provided filesystem paths and return a new routine object.
func New(paths []string, colors ...[3]string) *routine {
	var r routine

	for _, path := range paths {
		r.disks = append(r.disks, fs{path: path})
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

// For each provided filesystem, get the amounts of used and total disk space and
// convert them into a human-readable size.
func (r *routine) Update() {
	var b syscall.Statfs_t

	for i, disk := range r.disks {
		r.err = syscall.Statfs(disk.path, &b)
		if r.err != nil {
			return
		}

		total  := b.Blocks * uint64(b.Bsize)
		used   := total - (b.Bavail * uint64(b.Bsize))
		r.disks[i].perc  = (used * 100) / total

		r.disks[i].used,  r.disks[i].used_u  = shrink(used)
		r.disks[i].total, r.disks[i].total_u = shrink(total)
	}
}

// Format and print the amounts of disk space for each provided filesystem.
func (r *routine) String() string {
	var c string
	var b strings.Builder

	if r.err != nil {
		return r.colors.error + r.err.Error() + COLOR_END
	}

	for i, disk := range r.disks {
		if disk.perc > 90 {
			c = r.colors.error
		} else if disk.perc > 75 {
			c = r.colors.warning
		} else {
			c = r.colors.normal
		}

		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(c)
		fmt.Fprintf(&b, "%s: %v%c/%v%c", disk.path, disk.used, disk.used_u, disk.total, disk.total_u)
		b.WriteString(COLOR_END)
	}

	return b.String()
}

// Iteratively decrease the amount of bytes by a step of 2^10 until human-readable.
func shrink(blocks uint64) (uint64, rune) {
	var units = [...]rune{'B', 'K', 'M', 'G', 'T', 'P', 'E'}
	var i int

	for blocks > 1024 {
		blocks >>= 10
		i++
	}

	return blocks, units[i]
}
