package discorder

import (
	"github.com/jonas747/discorder/ui"
)

var Commands = []*Command{
	&Command{
		Name:        "OpenCommands",
		Description: "Opens up the command window with all commands available",
		Category:    "Hidden",
		Run:         func(app *App, args []*Argument) {},
	},
	&Command{
		Name:        "MoveCursor",
		Description: "Moves cursor in specified direction",
		Category:    "Misc",
		Args: []*ArgumentDef{
			&ArgumentDef{Name: "direction", Optional: false, Datatype: ArgumentDataTypeString},
			&ArgumentDef{Name: "amount", Optional: false, Datatype: ArgumentDataTypeInt},
			&ArgumentDef{Name: "word", Optional: true, Datatype: ArgumentDataTypeBool},
		},
		Run: func(app *App, args []*Argument) {
			amount := 1
			word := false
			dir := ""

			for _, v := range args {
				switch v.Name {
				case "direction":
					dir, _ = v.Val.(string)
				case "amount":
					amount, _ = v.Int()
				case "word":
					word, _ = v.Val.(bool)
				}
			}

			moveDir := StringToDir(dir)

			if app.ViewManager.ActiveInput != nil && app.ViewManager.ActiveInput.Active {
				app.ViewManager.ActiveInput.MoveCursor(moveDir, amount, word)
			}
		},
	},
	&Command{
		Name:        "Erase",
		Description: "Erase text",
		Category:    "Misc",
		Args: []*ArgumentDef{
			&ArgumentDef{Name: "direction", Optional: false, Datatype: ArgumentDataTypeString},
			&ArgumentDef{Name: "amount", Optional: false, Datatype: ArgumentDataTypeInt},
			&ArgumentDef{Name: "word", Optional: true, Datatype: ArgumentDataTypeBool},
		},
		Run: func(app *App, args []*Argument) {
			amount := 1
			word := false
			dir := ""

			for _, v := range args {
				switch v.Name {
				case "direction":
					dir, _ = v.Val.(string)
				case "amount":
					amount, _ = v.Int()
				case "word":
					word, _ = v.Val.(bool)
				}
			}

			moveDir := StringToDir(dir)

			if app.ViewManager.ActiveInput != nil && app.ViewManager.ActiveInput.Active {
				app.ViewManager.ActiveInput.Erase(moveDir, amount, word)
			}
		},
	},
	&Command{
		Name:        "OpenServers",
		Description: "Opens up the server window",
		Category:    "Main",
		Run: func(app *App, args []*Argument) {
			if app.ViewManager.CanOpenWindow() {
				ssw := NewSelectServerWindow(app, app.ViewManager.SelectedMessageView)
				app.ViewManager.SetActiveWindow(ssw)
			}
		},
	},
	&Command{
		Name:        "OpenHelp",
		Description: "Opens up the help window",
		Category:    "Main",
		Run: func(app *App, args []*Argument) {
			if app.ViewManager.CanOpenWindow() {
				hw := NewHelpWindow(app)
				app.ViewManager.SetActiveWindow(hw)
			}
		},
	},
	&Command{
		Name:        "OpenMessage",
		Description: "Opens message window",
		Category:    "Misc",
		Args: []*ArgumentDef{
			&ArgumentDef{Name: "message", Optional: true, Datatype: ArgumentDataTypeString},
		},
		Run: func(app *App, args []*Argument) {
			app.ViewManager.SelectedMessageView.OpenMessageSelectWindow("")
		},
	},
	&Command{
		Name:        "Scroll",
		Description: "Scrolls currently active view",
		Category:    "Misc",
		Args: []*ArgumentDef{
			&ArgumentDef{Name: "direction", Optional: false, Datatype: ArgumentDataTypeString},
			&ArgumentDef{Name: "amount", Optional: false, Datatype: ArgumentDataTypeInt},
			&ArgumentDef{Name: "word", Optional: true, Datatype: ArgumentDataTypeBool},
		},
		Run: func(app *App, args []*Argument) {
			amount := 1
			dir := ""

			for _, v := range args {
				switch v.Name {
				case "Direction":
					dir, _ = v.Val.(string)
				case "Amount":
					amount, _ = v.Int()
				}
			}

			moveDir := StringToDir(dir)

			if app.ViewManager.activeWindow != nil {
				scrollable, ok := app.ViewManager.activeWindow.(ui.Scrollable)
				if ok {
					scrollable.Scroll(moveDir, amount)
				}
			} else if app.ViewManager.SelectedMessageView != nil {
				app.ViewManager.SelectedMessageView.Scroll(moveDir, amount)
			}
		},
	},
	&Command{
		Name:        "Select",
		Description: "Select the currently highlighted element",
		Category:    "Misc",
	},
	&Command{
		Name:        "Mark",
		Description: "Toggles the currently highlited element",
		Category:    "Misc",
	},
	&Command{
		Name:        "ClearLog",
		Description: "Clear the logbuffer",
		Category:    "Main",
	},
	&Command{
		Name:        "Quit",
		Description: "Quit discorder",
		Category:    "Main",
		Run: func(app *App, args []*Argument) {
			go app.Stop()
		},
	},
}

func StringToDir(dir string) ui.Direction {
	switch dir {
	case "left":
		return ui.DirLeft
	case "right":
		return ui.DirRight
	case "up":
		return ui.DirUp
	case "down":
		return ui.DirDown
	case "end":
		return ui.DirEnd
	case "start":
		return ui.DirStart
	}
	return ui.DirLeft
}

func GetCommandByName(name string) *Command {
	for _, cmd := range Commands {
		if cmd.Name == name {
			return cmd
		}
	}
	return nil
}
