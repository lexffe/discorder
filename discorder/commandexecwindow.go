package discorder

import (
	"github.com/jonas747/discorder/common"
	"github.com/jonas747/discorder/ui"
)

type CommandExecWindow struct {
	*ui.BaseEntity
	app        *App
	menuWindow *ui.MenuWindow
	command    Command
}

func NewCommandExecWindow(layer int, app *App, command Command) *CommandExecWindow {
	execWindow := &CommandExecWindow{
		BaseEntity: &ui.BaseEntity{},
		app:        app,
		menuWindow: ui.NewMenuWindow(layer, app.ViewManager.UIManager, false),
		command:    command,
	}

	execWindow.menuWindow.Transform.AnchorMax = common.NewVector2F(1, 1)
	execWindow.menuWindow.Transform.Top = 1
	execWindow.menuWindow.Transform.Bottom = 2

	execWindow.menuWindow.Window.Title = "Execute command"
	execWindow.menuWindow.Window.Footer = ":)"

	app.ApplyThemeToMenu(execWindow.menuWindow)

	execWindow.Transform.AddChildren(execWindow.menuWindow)

	execWindow.Transform.AnchorMin = common.NewVector2F(0.1, 0)
	execWindow.Transform.AnchorMax = common.NewVector2F(0.9, 1)

	app.ViewManager.UIManager.AddWindow(execWindow)

	execWindow.GenMenu()

	return execWindow
}

func (cew *CommandExecWindow) Destroy() {
	cew.app.ViewManager.UIManager.RemoveWindow(cew)
	cew.DestroyChildren()
}

func (cew *CommandExecWindow) GenMenu() {
	items := make([]*ui.MenuItem, 0)
	for _, arg := range cew.command.GetArgs() {
		helper := &ui.MenuItem{
			Name: arg.Name,
			Info: arg.Description,
		}
		input := &ui.MenuItem{
			Name:      arg.Name,
			Info:      arg.Description,
			IsInput:   true,
			InputType: arg.Datatype,
			UserData:  arg,
		}
		items = append(items, helper, input)
	}

	exec := &ui.MenuItem{
		Name: "Execute",
		Info: "Execute the commadn with specified args",
	}
	items = append(items, exec)
	cew.menuWindow.SetOptions(items)
}
