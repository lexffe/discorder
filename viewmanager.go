package main

import (
	"github.com/jonas747/discorder/common"
	"github.com/jonas747/discorder/ui"
	"github.com/nsf/termbox-go"
	"log"
	"unicode/utf8"
)

type ViewManager struct {
	*ui.BaseEntity
	App                 *App
	mv                  *ui.MessageView
	selectedMessageView *ui.MessageView
	activeWindow        ui.Entity
	input               *ui.TextInput
}

func NewViewManager(app *App) *ViewManager {
	mv := &ViewManager{
		BaseEntity: &ui.BaseEntity{},
		App:        app,
	}
	return mv
}

func (v *ViewManager) OnInit() {
	// Add the header
	header := ui.NewText()
	header.Text = "Discorder v" + VERSION + "(´ ▽ ` )ﾉ"
	hw := utf8.RuneCountInString(header.Text)
	header.Transform.Size = common.NewVector2I(hw, 0)
	header.Transform.AnchorMin = common.NewVector2F(0.5, 0)
	header.Transform.AnchorMax = common.NewVector2F(0.5, 0)
	header.Transform.Position.X = float32(-(hw / 2))
	v.AddChild(header)

	// Launch the login
	login := NewLoginWindow(v.App)
	v.App.AddChild(login)
	login.CheckAutoLogin()
}

func (v *ViewManager) OnReady() {
	// go into the main view

	mv := ui.NewMessageView(v.App.session.State)
	mv.Transform.AnchorMax = common.NewVector2I(1, 1)
	mv.Transform.Bottom = 2
	mv.Transform.Top = 1
	mv.ShowPrivate = true
	mv.Logs = v.App.logBuffer
	v.AddChild(mv)
	v.mv = mv

	input := ui.NewTextInput()
	input.Transform.AnchorMin = common.NewVector2F(0, 1)
	input.Transform.AnchorMax = common.NewVector2F(1, 1)
	input.Transform.Position.Y = -1
	input.Active = true
	v.AddChild(input)
	v.input = input
}

func (v *ViewManager) Destroy() { v.DestroyChildren() }

func (v *ViewManager) PreDraw() {
	if v.mv != nil {
		v.mv.Logs = v.App.logBuffer
	}
}

func (v *ViewManager) GetDrawLayer() int {
	return 0
}

func (v *ViewManager) HandleInput(event termbox.Event) {
	if event.Type == termbox.EventKey {
		switch event.Key {
		case termbox.KeyCtrlG: // Select channel
			if v.activeWindow != nil {
				break
			}
		case termbox.KeyCtrlO: // Options
			if v.activeWindow != nil {
				break
			}
			hw := NewHelpWindow(v.App)
			v.AddChild(hw)
			v.activeWindow = hw
			log.Println("Opening help")
			v.input.Active = false
		case termbox.KeyCtrlS: // Select server
			if v.activeWindow != nil {
				break
			}
			ssw := NewSelectServerWindow(v.App, v.mv)
			v.AddChild(ssw)
			v.activeWindow = ssw
			v.input.Active = false
			log.Println("Opening server select window")
		case termbox.KeyBackspace, termbox.KeyBackspace2:
			if v.activeWindow != nil {
				v.RemoveChild(v.activeWindow, true)
				v.activeWindow = nil
				v.input.Active = true
			}
		}
	}
}
