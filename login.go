package main

import (
	"github.com/jonas747/discorder/common"
	"github.com/jonas747/discorder/ui"
	"github.com/nsf/termbox-go"
	"log"
)

var (
	InputStateEmail    = 0
	InputStatePassword = 1
)

type LoginWindow struct {
	*ui.BaseEntity
	App *App

	CurInputState int
	PWInput       *ui.TextInput
	EmailInput    *ui.TextInput

	loggingIn bool

	Footer *ui.Text
	Helper *ui.Text

	SavePassword       bool
	currentlyLoggingIn bool
	// pwBuffer      string
	// savePassword  bool
	// err           error
}

func NewLoginWindow(app *App) *LoginWindow {
	window := ui.NewWindow()
	window.Transform.AnchorMax = common.NewVector2F(0.5, 0.5)
	window.Transform.AnchorMin = common.NewVector2F(0.5, 0.5)
	window.Transform.Size = common.NewVector2I(50, 10)
	window.Transform.Position = common.NewVector2I(-25, -5)
	window.Title = "Login"

	helper := ui.NewText()
	helper.Text = "Enter Email"
	helper.Transform.Parent = window.Transform
	helper.Transform.Position = common.NewVector2I(1, 2)
	helper.Transform.Size = common.NewVector2I(45, 1)
	helper.Layer = 5
	window.AddChild(helper)

	mailInput := ui.NewTextInput()
	mailInput.Transform.Parent = window.Transform
	mailInput.Transform.Position = common.NewVector2I(1, 3)
	mailInput.Transform.Size = common.NewVector2I(45, 0)
	mailInput.Active = true
	mailInput.Layer = 5
	window.AddChild(mailInput)

	pwInput := ui.NewTextInput()
	pwInput.Transform.Parent = window.Transform
	pwInput.Transform.Position = common.NewVector2I(1, 5)
	pwInput.Transform.Size = common.NewVector2I(45, 0)
	pwInput.MaskInput = true
	pwInput.Layer = 5
	window.AddChild(pwInput)

	footer1 := ui.NewText()
	footer1.Text = "Ctrl-t toggle password saving(currently off)"
	footer1.Transform.Parent = window.Transform
	footer1.Transform.Position = common.NewVector2I(1, 7)
	footer1.Transform.Size = common.NewVector2I(45, 1)
	footer1.Layer = 5
	window.AddChild(footer1)

	footer2 := ui.NewText()
	footer2.Text = "Ctrl-s switch between email and password"
	footer2.Transform.Parent = window.Transform
	footer2.Transform.Position = common.NewVector2I(1, 8)
	footer2.Transform.Size = common.NewVector2I(45, 1)
	footer2.Layer = 5
	window.AddChild(footer2)

	lw := &LoginWindow{
		BaseEntity: &ui.BaseEntity{},
		Footer:     footer1,
		PWInput:    pwInput,
		EmailInput: mailInput,
		App:        app,
		Helper:     helper,
	}
	lw.AddChild(window)
	return lw
}

func (lw *LoginWindow) CheckAutoLogin() {
	if lw.App.config.Email != "" && lw.App.config.Password != "" {
		lw.Trylogin(lw.App.config.Email, lw.App.config.Password)
	}
}

func (lw *LoginWindow) Destroy() { lw.DestroyChildren() }

func (lw *LoginWindow) HandleInput(event termbox.Event) {
	if event.Type == termbox.EventKey {
		switch event.Key {
		case termbox.KeyEnter:
			if lw.CurInputState == InputStateEmail {
				lw.App.config.Email = lw.EmailInput.TextBuffer
				lw.CurInputState = InputStatePassword
				lw.EmailInput.Active = false
				lw.PWInput.Active = true
			} else {
				pw := lw.PWInput.TextBuffer
				if lw.SavePassword {
					lw.App.config.Password = pw
				}
				lw.Trylogin(lw.App.config.Email, pw)
				// err := lw.App.Login(lw.App.config.Email, pw)
				// if err != nil {
				// 	log.Println("Error logging in: ", err)
				// } else {
				// 	lw.App.config.Save(configPath)
				// 	// TODO leave window cause we logged in and whatnot
				// }
			}
		case termbox.KeyCtrlS:
			if lw.CurInputState == InputStateEmail {
				lw.EmailInput.Active = false
				lw.PWInput.Active = true
				lw.CurInputState = InputStatePassword
			} else {
				lw.EmailInput.Active = true
				lw.PWInput.Active = false
				lw.CurInputState = InputStateEmail
			}
		case termbox.KeyCtrlT:
			lw.SavePassword = !lw.SavePassword

			statusStr := "on"
			if !lw.SavePassword {
				statusStr = "off"
			}

			lw.Footer.Text = "Ctrl-t toggle password saving(currently " + statusStr + ")"
		}
	}
	log.Println("Handled input in loginwindow")
}

func (lw *LoginWindow) Trylogin(user, pw string) {
	log.Println("Attempting login...")

	lw.loggingIn = true
	lw.App.Draw()

	err := lw.App.Login(lw.App.config.Email, pw)
	if err != nil {
		log.Println("Error logging in: ", err)
	} else {
		lw.App.config.Save(*configPath)
		lw.App.RemoveChild(lw, true)
	}
	lw.loggingIn = false
}

func (lw *LoginWindow) PreDraw() {
	if lw.Helper == nil {
		return
	}
	if lw.CurInputState == InputStateEmail {
		lw.Helper.Text = "Enter Email"
	} else {
		lw.Helper.Text = "Enter Password"
	}

	if lw.loggingIn {
		lw.Helper.Text = "Logging in..."
	}
}
