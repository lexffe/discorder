package discorder

import (
	"github.com/jonas747/discorder/common"
	"github.com/jonas747/discorder/ui"
	"github.com/jonas747/termbox-go"
)

var HelpContent = []string{
	"Keyboard shortcuts:",
	"Look in your discorder.json file in either ~/.config/discorder or %%appdata%/discorder if you're on windows",
	"\n\nIf all private channels are selected, it will go into a All-Private\nmode meaning new messages from people you haven't talked to before\nwill also show up",
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

	window := ui.NewWindow(app.ViewManager.UIManager)
	window.Title = "Help"
	window.Footer = "Hmmm - Mr Smilery"
	window.Transform.AnchorMax = common.NewVector2F(1, 1)
	window.Layer = 10

	text := ui.NewText()
	text.Transform.AnchorMin = common.NewVector2I(0, 0)
	text.Transform.AnchorMax = common.NewVector2I(1, 1)
	app.ApplyThemeToText(text, "text_window_normal")
	window.Transform.AddChildren(text)
	text.Layer = 10

	for k, v := range HelpContent {
		if k != 0 {
			text.Text += "\n"
		}
		text.Text += v
	}

	hw.Transform.AddChildren(window)
	hw.Window = window

	hw.Transform.AnchorMax = common.NewVector2I(1, 1)
	hw.Transform.Right = 2
	hw.Transform.Left = 1

	app.ViewManager.UIManager.AddWindow(hw)
	return hw
}

func (s *HelpWindow) Enter() {}
func (s *HelpWindow) Destroy() {
	s.App.ViewManager.UIManager.RemoveWindow(s)
	s.DestroyChildren()
}

func (s *HelpWindow) HandleInput(event termbox.Event) {
	// if event.Type == termbox.EventKey {
	// 	switch event.Key {
	// 	case termbox.KeyBackspace, termbox.KeyBackspace2:
	// 		s.App.entityContainer.RemoveChild(s, true)
	// 	}
	// }
}
