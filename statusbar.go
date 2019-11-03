package statusbar

import (
	"fmt"
)

type Routine interface {
	Update() error
	String() string
	Sleep()
}

type Bar []Routine
