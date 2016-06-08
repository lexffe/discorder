package discorder

import (
	"github.com/jonas747/discorder/common"
	"github.com/jonas747/discorder/ui"
)

type CommandWindow struct {
	*ui.BaseEntity
	app        *App
	menuWindow *ui.MenuWindow
	generated  bool
}

func NewCommandWindow(app *App, layer int) *CommandWindow {
	cw := &CommandWindow{
		BaseEntity: &ui.BaseEntity{},
		app:        app,
	}

	menuWindow := ui.NewMenuWindow(layer, app.ViewManager.UIManager, true)

	menuWindow.Transform.AnchorMax = common.NewVector2F(1, 1)
	menuWindow.Transform.Top = 1
	menuWindow.Transform.Bottom = 2

	menuWindow.Window.Title = "Execute command"
	menuWindow.Window.Footer = ":)"

	app.ApplyThemeToMenu(menuWindow)

	cw.menuWindow = menuWindow
	cw.Transform.AddChildren(menuWindow)

	cw.Transform.AnchorMin = common.NewVector2F(0, 0)
	cw.Transform.AnchorMax = common.NewVector2F(1, 1)
	cw.Transform.Right = 2
	cw.Transform.Left = 1

	app.ViewManager.UIManager.AddWindow(cw)

	return cw
}
func (cw *CommandWindow) GenMenu() {
	options := make([]*ui.MenuItem, 0)

	for _, category := range CommandCategories {
		// Category
		options = append(options, category.GenMenu(cw.app, Commands, CommandCategories))
	}

	// Add the top level commands
	for _, cmd := range Commands {
		if len(cmd.GetCategory()) < 1 {
			options = append(options, cw.app.GenMenuItemFromCommand(cmd))
		}
	}

	cw.menuWindow.SetOptions(options)
}

func (cw *CommandWindow) Destroy() {
	cw.app.ViewManager.UIManager.RemoveWindow(cw)
	cw.DestroyChildren()
}

func (cw *CommandWindow) Update() {
	if !cw.generated {
		cw.GenMenu()
		cw.generated = true
	}
}

func (cw *CommandWindow) Select() {
	element := cw.menuWindow.GetHighlighted()
	if element == nil {
		return
	}

	if element.IsCategory {
		cw.menuWindow.Select()
		return
	}

	if element.UserData == nil {
		return // We can't deal
	}

	cmd, ok := element.UserData.(Command)
	if !ok {
		return // We must be doing something very wrong somewhere
	}

	execWindow := NewCommandExecWindow(7, cw.app, cmd)
	cw.app.ViewManager.Transform.AddChildren(execWindow)

	//cw.Transform.Parent.RemoveChild(cw, true)
}
