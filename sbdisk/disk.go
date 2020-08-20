// Package sbdisk displays disk resources for each filesystem provided.
package sbdisk

import (
	"errors"
	"fmt"
	"strings"
	"syscall"
)

var colorEnd = "^d^"

// Routine is the main object for this package.
type Routine struct {
	// Error encountered along the way, if any.
	err error

	// Slice of provided filesystems to stat.
	disks []fs

	// Trio of user-provided colors for displaying various states.
	colors struct {
		normal  string
		warning string
		error   string
	}
}

// fs holds information about a single filesystem.
type fs struct {
	// Given path that will be used to stat the partition.
	path string

	// Used bytes for this filesystem.
	used uint64

	// Unit for the used bytes.
	usedUnit rune

	// Total bytes for this filesystem.
	total uint64

	// Unit for the total bytes.
	totalUnit rune

	// Percentage of total disk space used.
	// Note: Bavail is the amount of blocks that can actually be used, while Bfree is the total amount of unused blocks.
	perc uint64
}

// New copies over the provided filesystem paths and makes a new routine object. colors is an optional triplet of hex
// color codes for colorizing the output based on these rules:
// 1. Normal color, disk is less than 75% full.
// 2. Warning color, disk is between 75% and 90% full.
// 3. Error color, disk is over 90% full.
func New(paths []string, colors ...[3]string) *Routine {
	var r Routine

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
		r.colors.normal = "^c" + colors[0][0] + "^"
		r.colors.warning = "^c" + colors[0][1] + "^"
		r.colors.error = "^c" + colors[0][2] + "^"
	} else {
		// If a color array wasn't passed in, then we don't want to print this.
		colorEnd = ""
	}

	return &r
}

// Update gets the amount of used and total disk space and converts them into a human-readable size for each provided
// filesystem.
func (r *Routine) Update() {
	var b syscall.Statfs_t

	for i, disk := range r.disks {
		r.err = syscall.Statfs(disk.path, &b)
		if r.err != nil {
			return
		}

		total := b.Blocks * uint64(b.Bsize)
		used := total - (b.Bavail * uint64(b.Bsize))
		r.disks[i].perc = (used * 100) / total

		r.disks[i].used, r.disks[i].usedUnit = shrink(used)
		r.disks[i].total, r.disks[i].totalUnit = shrink(total)
	}
}

// String formats and prints the amounts of disk space for each provided filesystem.
func (r *Routine) String() string {
	var c string
	var b strings.Builder

	if r.err != nil {
		return r.colors.error + r.err.Error() + colorEnd
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
		fmt.Fprintf(&b, "%s: %v%c/%v%c", disk.path, disk.used, disk.usedUnit, disk.total, disk.totalUnit)
		b.WriteString(colorEnd)
	}

	return b.String()
}

// Shrink iteratively decreases the amount of bytes by a step of 2^10 until human-readable.
func shrink(blocks uint64) (uint64, rune) {
	var units = [...]rune{'B', 'K', 'M', 'G', 'T', 'P', 'E'}
	var i int

	for blocks > 1024 {
		blocks >>= 10
		i++
	}

	return blocks, units[i]
}
