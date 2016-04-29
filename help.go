package main

import (
	"github.com/jonas747/discorder/common"
	"github.com/jonas747/discorder/ui"
	"github.com/nsf/termbox-go"
)

var HelpContent = []string{
	"Keyboard shortcuts:",
	"Ctrl-O: Help",
	"Ctrl-S: Select server/channel",
	"    Space: toggle",
	"    Enter: Select as send channel",
	//"Ctrl-G: Select channels (removed, use ctrl-s)",
	//"Ctrl-P: Select private conversation (removed, use ctrl-s)",
	"Ctrl-Q: Quit",
	"Ctrl-L: Clear log messages",
	"Up: Scroll up",
	"Down: Scroll down",
	"Backspace: Close current wnidow",
	"\nIf all private channels are selected, it will go into a AllPrivate\nmode meaning new messages from people you haven't talked to before\nwill also show up",
	"--------------",
	"You are using Discorder version " + VERSION,
	"This is still in very early development, please report any bugs you find here",
	"https://github.com/jonas747/discorder",
}

type HelpWindow struct {
	*ui.BaseEntity
	App    *App
	Window *ui.Window
}

func NewHelpWindow(app *App) *HelpWindow {
	hw := &HelpWindow{
		BaseEntity: &ui.BaseEntity{},
		App:        app,
	}

	wWidth := 70
	wHeight := 21

	curY := 1

	window := ui.NewWindow()
	window.Title = "Help"
	window.Footer = "Hmmm - Mr Smilery"
	window.Transform.AnchorMax = common.NewVector2F(0.5, 0.5)
	window.Transform.AnchorMin = common.NewVector2F(0.5, 0.5)
	window.Transform.Position = common.NewVector2I(-(wWidth / 2), -(wHeight / 2))
	window.Transform.Size = common.NewVector2I(wWidth, wHeight)

	for _, v := range HelpContent {
		text := ui.NewText()
		text.Transform.AnchorMin = common.NewVector2I(0, 0)
		text.Transform.AnchorMax = common.NewVector2I(1, 0)
		text.Transform.Position = common.NewVector2I(0, curY)
		text.Transform.Parent = window.Transform
		text.BG = WindowTextBG
		text.Text = v
		curY += text.HeightRequired()
		window.AddChild(text)
	}
	hw.AddChild(window)
	hw.Window = window
	return hw
}

func (s *HelpWindow) Enter()   {}
func (s *HelpWindow) Destroy() { s.DestroyChildren() }

func (s *HelpWindow) HandleInput(event termbox.Event) {
	// if event.Type == termbox.EventKey {
	// 	switch event.Key {
	// 	case termbox.KeyBackspace, termbox.KeyBackspace2:
	// 		s.App.entityContainer.RemoveChild(s, true)
	// 	}
	// }
}
