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

func New() Bar {
	var b Bar
	return b
}
