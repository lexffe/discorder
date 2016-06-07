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

	menuWindow := ui.NewMenuWindow(layer, app.ViewManager.UIManager)

	menuWindow.Transform.AnchorMax = common.NewVector2F(1, 1)
	menuWindow.Transform.Top = 1
	menuWindow.Transform.Bottom = 2

	menuWindow.Window.Title = "Execute command"
	menuWindow.Window.Footer = ":)"

	app.ApplyThemeToMenu(menuWindow)

	cw.menuWindow = menuWindow
	cw.Transform.AddChildren(menuWindow)

	cw.Transform.AnchorMin = common.NewVector2F(0.1, 0)
	cw.Transform.AnchorMax = common.NewVector2F(0.9, 1)

	app.ViewManager.UIManager.AddWindow(cw)

	return cw
}
func (cw *CommandWindow) GenMenu() {
	options := make([]*ui.MenuItem, 0)

	for _, category := range CommandCategories {
		// Category
		options = append(options, category.GenMenu(Commands, CommandCategories))
	}

	// Add the top level commands
	for _, cmd := range Commands {
		if len(cmd.Category) < 1 {
			options = append(options, cmd.GenMenuItem())
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
