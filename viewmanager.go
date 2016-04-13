package main

import (
	"github.com/jonas747/discorder/common"
	"github.com/jonas747/discorder/ui"
	"unicode/utf8"
)

type ViewManager struct {
	*ui.BaseEntity
	App *App
	mv  *ui.MessageView
}

func NewViewManager(app *App) *ViewManager {
	return &ViewManager{
		BaseEntity: &ui.BaseEntity{},
		App:        app,
	}
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
	v.App.AddEntity(login)
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
	input.Active = true
	v.AddChild(input)
}

func (v *ViewManager) Destroy() { v.DestroyChildren() }

func (v *ViewManager) Draw() {
	if v.mv != nil {
		v.mv.Logs = v.App.logBuffer
	}
}

func (v *ViewManager) GetDrawLayer() int {
	return 0
}
