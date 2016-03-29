package main

import (
	"github.com/nsf/termbox-go"
)

type State interface {
	Enter()
	Exit()
	HandleInput(event termbox.Event)
	RefreshDisplay()
}
