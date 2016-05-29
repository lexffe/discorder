package discorder

import (
	"github.com/jonas747/discorder/common"
	"github.com/jonas747/discorder/ui"
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

	Helper *ui.Text

	SavePassword       bool
	currentlyLoggingIn bool
}

func NewLoginWindow(app *App) *LoginWindow {
	window := ui.NewWindow(app.ViewManager.UIManager)
	window.Transform.AnchorMax = common.NewVector2F(1, 1)
	window.Title = "Login"

	helper := ui.NewText()
	helper.Text = "Enter Email"
	helper.Transform.Position = common.NewVector2I(1, 2)
	helper.Transform.Size = common.NewVector2I(45, 1)
	helper.Layer = 5
	window.Transform.AddChildren(helper)

	mailInput := ui.NewTextInput(app.ViewManager.UIManager)
	mailInput.Transform.Position = common.NewVector2I(1, 3)
	mailInput.Transform.Size = common.NewVector2I(45, 0)
	mailInput.Active = true
	mailInput.Layer = 5
	mailInput.TextBuffer = app.config.Email

	window.Transform.AddChildren(mailInput)

	pwInput := ui.NewTextInput(app.ViewManager.UIManager)
	pwInput.Transform.Position = common.NewVector2I(1, 5)
	pwInput.Transform.Size = common.NewVector2I(45, 0)
	pwInput.Active = false
	pwInput.MaskInput = true
	pwInput.Layer = 5
	window.Transform.AddChildren(pwInput)

	app.ViewManager.UIManager.SetActiveInput(mailInput)

	footer2 := ui.NewText()
	footer2.Text = "Ctrl-s switch between email and password"
	footer2.Transform.Position = common.NewVector2I(1, 8)
	footer2.Transform.Size = common.NewVector2I(45, 1)
	footer2.Layer = 5
	window.Transform.AddChildren(footer2)

	lw := &LoginWindow{
		BaseEntity: &ui.BaseEntity{},
		PWInput:    pwInput,
		EmailInput: mailInput,
		App:        app,
		Helper:     helper,
	}
	lw.Transform.AddChildren(window)
	lw.Transform.AnchorMin = common.NewVector2F(0.5, 0.5)
	lw.Transform.AnchorMax = common.NewVector2F(0.5, 0.5)
	lw.Transform.Size = common.NewVector2I(50, 10)
	lw.Transform.Position = common.NewVector2I(-25, -5)
	return lw
}

func (lw *LoginWindow) CheckAutoLogin() {
	if lw.App.config.AuthToken != "" {
		lw.Trylogin("", "", lw.App.config.AuthToken)
	}
}

func (lw *LoginWindow) Destroy() { lw.DestroyChildren() }

func (lw *LoginWindow) Trylogin(email, pw, token string) {
	log.Println("Attempting login...")

	lw.loggingIn = true
	lw.App.Draw()

	err := lw.App.Login(email, pw, token)
	if err != nil {
		log.Println("Error logging in: ", err)
	} else {
		log.Println("Logged in!")
		lw.App.config.Save(lw.App.configPath)
		lw.App.Transform.RemoveChild(lw, true)
	}
	lw.loggingIn = false
}

func (lw *LoginWindow) Update() {
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
func (lw *LoginWindow) OnCommand(cmd *Command, args Arguments) {
	if cmd.Name == "select" {
		if lw.CurInputState == InputStateEmail {
			lw.App.config.Email = lw.EmailInput.TextBuffer
			lw.CurInputState = InputStatePassword
			lw.App.ViewManager.UIManager.SetActiveInput(lw.PWInput)
		} else {
			pw := lw.PWInput.TextBuffer
			lw.Trylogin(lw.App.config.Email, pw, "")
		}
	} else if cmd.Name == "scroll" {
		if lw.CurInputState == InputStateEmail {
			lw.App.ViewManager.UIManager.SetActiveInput(lw.PWInput)
			lw.CurInputState = InputStatePassword
		} else {
			lw.App.ViewManager.UIManager.SetActiveInput(lw.EmailInput)
			lw.CurInputState = InputStateEmail
		}
	}
}
