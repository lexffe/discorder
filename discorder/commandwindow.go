package discorder

import (
	"github.com/jonas747/discorder/common"
	"github.com/jonas747/discorder/ui"
)

type CommandWindow struct {
	*ui.BaseEntity
	app          *App
	menuWindow   *ui.MenuWindow
	generated    bool
	providedArgs map[string]interface{}
	header       string

	commands   []Command
	categories []*CommandCategory
}

func NewCommandWindow(app *App, layer int, providedArgs map[string]interface{}, header string) *CommandWindow {
	cw := &CommandWindow{
		BaseEntity: &ui.BaseEntity{},
		app:        app,
	}

	menuWindow := ui.NewMenuWindow(layer, app.ViewManager.UIManager, true)

	menuWindow.Transform.AnchorMax = common.NewVector2F(1, 1)
	menuWindow.Transform.Top = 1
	menuWindow.Transform.Bottom = 2

	menuWindow.Window.Title = "Commands"
	menuWindow.Window.Footer = ":)"

	app.ApplyThemeToMenu(menuWindow)

	cw.menuWindow = menuWindow
	cw.Transform.AddChildren(menuWindow)

	cw.Transform.AnchorMin = common.NewVector2F(0, 0)
	cw.Transform.AnchorMax = common.NewVector2F(1, 1)
	cw.Transform.Right = 2
	cw.Transform.Left = 1

	cw.providedArgs = providedArgs
	cw.header = header

	app.ViewManager.UIManager.AddWindow(cw)

	return cw
}

func (cw *CommandWindow) GenMenu() {
	options := make([]*ui.MenuItem, 0)

	if cw.header != "" {
		options = append(options, &ui.MenuItem{
			Name:       cw.header,
			Info:       "What do you call an ak47 that fires blanks? jk47",
			Decorative: true,
		})
	}
	if cw.commands == nil {
		cw.commands = cw.app.Commands
	}

	if cw.categories == nil {
		cw.categories = CommandCategories
	}

	commands := cw.commands

	// Filter out only the commands related to the provided args
	if len(cw.providedArgs) > 0 {

		filtered := make([]Command, 0)
		for _, v := range commands {
			matches := cw.GetArgMatchesForCommand(v)
			if len(matches) > 0 || v.GetIgnoreFilter() {
				filtered = append(filtered, v)
			}
		}
		commands = filtered
	}

	for _, category := range cw.categories {
		// Category
		options = append(options, category.GenMenu(cw.app, commands, cw.categories))
	}

	// Add the top level commands
	for _, cmd := range commands {
		if len(cmd.GetCategory()) < 1 {
			options = append(options, cw.app.GenMenuItemFromCommand(cmd))
		}
	}

	cw.menuWindow.SetOptions(options)
}

func (cw *CommandWindow) GetArgMatchesForCommand(cmd Command) []string {
	matches := []string{}
	args := cmd.GetArgs(nil)
	for _, arg := range args {
		if arg.Helper == nil {
			continue
		}

		name := arg.Helper.GetName()
		for key := range cw.providedArgs {
			if name == key {
				matches = append(matches, key)
				break
			}
		}
	}
	return matches
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

	customWindow := cmd.GetCustomWindow()
	if customWindow != nil {
		customWindow.Run(cw.app, 7)
		return
	}
	var presetArgs Arguments
	if len(cw.providedArgs) > 0 {
		presetArgs = make(map[string]interface{})

		matches := cw.GetArgMatchesForCommand(cmd)
		cmdArgs := cmd.GetArgs(nil)

		if len(matches) > 0 {
		OUTER:
			for _, match := range matches {
				for _, arg := range cmdArgs {
					if arg.Helper == nil {
						continue
					}

					if match == arg.Helper.GetName() {
						presetArgs[arg.Name] = cw.providedArgs[match]
						continue OUTER
					}
				}
			}
		}
	}

	execWindow := NewCommandExecWindow(7, cw.app, cmd, presetArgs)
	if execWindow != nil {
		cw.app.ViewManager.AddWindow(execWindow)
	}
	cw.app.ViewManager.RemoveWindow(cw)
	//cw.Transform.Parent.RemoveChild(cw, true)
}
