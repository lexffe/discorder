package main

import (
	"github.com/nsf/termbox-go"
)

type StateHelp struct {
	app *App
}

func (s *StateHelp) Enter() {}
func (s *StateHelp) Exit()  {}
func (s *StateHelp) RefreshDisplay() {
	sizeX, sizeY := termbox.Size()

	wWidth := 70
	wHeight := 20

	startX := sizeX/2 - wWidth/2
	startY := sizeY/2 - wHeight/2

	curY := startY + 1
	DrawWindow("Help", "Hmmm - Mr Smilery", startX, startY, wWidth, wHeight, termbox.ColorBlack)
	curY += SimpleSetText(startX+1, curY, wWidth-2, "Keyboard shortcuts:", termbox.ColorDefault, termbox.ColorDefault)
	curY += SimpleSetText(startX+1, curY, wWidth-2, "Ctrl-O: Help", termbox.ColorDefault, termbox.ColorDefault)
	curY += SimpleSetText(startX+1, curY, wWidth-2, "Ctrl-S: Select server", termbox.ColorDefault, termbox.ColorDefault)
	curY += SimpleSetText(startX+1, curY, wWidth-2, "Ctrl-G: Select channels", termbox.ColorDefault, termbox.ColorDefault)
	curY += SimpleSetText(startX+1, curY, wWidth-2, "    Space: mark channel", termbox.ColorDefault, termbox.ColorDefault)
	curY += SimpleSetText(startX+1, curY, wWidth-2, "    Enter: Select as send channel (Also Mark)", termbox.ColorDefault, termbox.ColorDefault)
	curY += SimpleSetText(startX+1, curY, wWidth-2, "Ctrl-P: Select private conversation", termbox.ColorDefault, termbox.ColorDefault)
	curY += SimpleSetText(startX+1, curY, wWidth-2, "Ctrl-Q: Quit", termbox.ColorDefault, termbox.ColorDefault)
	curY += SimpleSetText(startX+1, curY, wWidth-2, "Ctrl-J: Queries the history of the current channel (For debugging)", termbox.ColorDefault, termbox.ColorDefault)
	curY += SimpleSetText(startX+1, curY, wWidth-2, "Ctrl-L: Clear log messages", termbox.ColorDefault, termbox.ColorDefault)
	curY += SimpleSetText(startX+1, curY, wWidth-2, "Up: Scroll up", termbox.ColorDefault, termbox.ColorDefault)
	curY += SimpleSetText(startX+1, curY, wWidth-2, "Down: Scroll down", termbox.ColorDefault, termbox.ColorDefault)
	curY += SimpleSetText(startX+1, curY, wWidth-2, "Backspace: Close current wnidow", termbox.ColorDefault, termbox.ColorDefault)
	curY++
	curY += SimpleSetText(startX+1, curY, wWidth-2, "You are using Discorder version "+VERSION, termbox.ColorDefault, termbox.ColorDefault)
	curY += SimpleSetText(startX+1, curY, wWidth-2, "This is still in very early development, please report any bugs you find here", termbox.ColorDefault, termbox.ColorDefault)
	curY += SimpleSetText(startX+1, curY, wWidth-2, "https://github.com/jonas747/discorder", termbox.ColorDefault, termbox.ColorDefault)
}

func (s *StateHelp) HandleInput(event termbox.Event) {
	if event.Type == termbox.EventKey {
		switch event.Key {
		case termbox.KeyBackspace, termbox.KeyBackspace2:
			s.app.SetState(&StateNormal{s.app})
		}
	}
}
