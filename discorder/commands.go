package discorder

import (
	"github.com/jonas747/discorder/ui"
	"log"
	"path/filepath"
)

var Commands = []*Command{
	&Command{
		Name:        "commands",
		Description: "Opens up the command window with all commands available",
		Category:    "Hidden",
		Run:         func(app *App, args Arguments) {},
	},
	&Command{
		Name:        "settings",
		Description: "Opens up the help window",
		Category:    "Main",
		Run: func(app *App, args Arguments) {
			if app.ViewManager.CanOpenWindow() {
				// hw := NewHelpWindow(app)
				// app.ViewManager.SetActiveWindow(hw)
			}
		},
	},
	&Command{
		Name:        "move_cursor",
		Description: "Moves cursor in specified direction",
		Category:    "Misc",
		Args: []*ArgumentDef{
			&ArgumentDef{Name: "direction", Optional: false, Datatype: ArgumentDataTypeString},
			&ArgumentDef{Name: "amount", Optional: false, Datatype: ArgumentDataTypeInt},
			&ArgumentDef{Name: "word", Optional: true, Datatype: ArgumentDataTypeBool},
		},
		Run: func(app *App, args Arguments) {
			amount, _ := args.Int("amount")
			words, _ := args.Bool("words")
			dir, _ := args.String("direction")
			moveDir := StringToDir(dir)

			if app.ViewManager.ActiveInput != nil && app.ViewManager.ActiveInput.Active {
				app.ViewManager.ActiveInput.MoveCursor(moveDir, amount, words)
			}
		},
	},
	&Command{
		Name:        "erase",
		Description: "Erase text",
		Category:    "Misc",
		Args: []*ArgumentDef{
			&ArgumentDef{Name: "direction", Optional: false, Datatype: ArgumentDataTypeString},
			&ArgumentDef{Name: "amount", Optional: false, Datatype: ArgumentDataTypeInt},
			&ArgumentDef{Name: "words", Optional: true, Datatype: ArgumentDataTypeBool},
		},
		Run: func(app *App, args Arguments) {
			amount, _ := args.Int("amount")
			words, _ := args.Bool("words")
			dir, _ := args.String("direction")

			moveDir := StringToDir(dir)

			if app.ViewManager.ActiveInput != nil && app.ViewManager.ActiveInput.Active {
				app.ViewManager.ActiveInput.Erase(moveDir, amount, words)
			}
		},
	},
	&Command{
		Name:        "servers",
		Description: "Opens up the server window",
		Category:    "Main",
		Run: func(app *App, args Arguments) {
			if app.ViewManager.CanOpenWindow() {
				ssw := NewSelectServerWindow(app, app.ViewManager.SelectedMessageView)
				app.ViewManager.SetActiveWindow(ssw)
			}
		},
	},
	&Command{
		Name:        "channels",
		Description: "Opens up the channel window",
		Category:    "Main",
		Run: func(app *App, args Arguments) {
			if app.ViewManager.CanOpenWindow() {
				// ssw := NewChannelSelectWindow(app, app.ViewManager.SelectedMessageView, guild)
				// app.ViewManager.SetActiveWindow(ssw)
			}
		},
	},
	&Command{
		Name:        "help",
		Description: "Opens up the help window",
		Category:    "Main",
		Run: func(app *App, args Arguments) {
			if app.ViewManager.CanOpenWindow() {
				hw := NewHelpWindow(app)
				app.ViewManager.SetActiveWindow(hw)
			}
		},
	},
	&Command{
		Name:        "message_window",
		Description: "Opens message window",
		Category:    "Misc",
		Args: []*ArgumentDef{
			&ArgumentDef{Name: "message", Optional: true, Datatype: ArgumentDataTypeString},
		},
		Run: func(app *App, args Arguments) {
			app.ViewManager.SelectedMessageView.OpenMessageSelectWindow("")
		},
	},
	&Command{
		Name:        "scroll",
		Description: "Scrolls currently active view",
		Category:    "Misc",
		Args: []*ArgumentDef{
			&ArgumentDef{Name: "direction", Optional: false, Datatype: ArgumentDataTypeString},
			&ArgumentDef{Name: "amount", Optional: false, Datatype: ArgumentDataTypeInt},
		},
		Run: func(app *App, args Arguments) {
			amount, _ := args.Int("amount")
			dir, _ := args.String("direction")
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
		Name:        "select",
		Description: "Select the currently highlighted element",
		Category:    "Misc",
	},
	&Command{
		Name:        "mark",
		Description: "Toggles the currently highlited element",
		Category:    "Misc",
	},
	&Command{
		Name:        "clear_log",
		Description: "Clear the logbuffer",
		Category:    "Main",
		Run: func(app *App, args Arguments) {
			logRoutine.Clear()
		},
	},
	&Command{
		Name:        "reload_theme",
		Description: "Reloads the current theme",
		Category:    "Main",
		Run: func(app *App, args Arguments) {
			userTheme := app.themePath
			if userTheme == "" {
				userTheme = filepath.Join(app.configDir, "themes", app.config.Theme)
			}
			if userTheme == "" {
				log.Println("No theme selected")
				return
			}

			app.userTheme = LoadTheme(userTheme)
			app.ViewManager.ApplyTheme()
		},
	},
	&Command{
		Name:        "delete",
		Description: "Deletes a message",
		Category:    "Util",
	},
	&Command{
		Name:        "game",
		Description: "Sets the game you're playing",
		Category:    "Util",
	},
	&Command{
		Name:        "send_message",
		Description: "Sends a message",
		Category:    "Util",
		Run: func(app *App, args Arguments) {
		},
	},
	&Command{
		Name:        "set_nick",
		Description: "Sets your nickname on a server (if possible)",
		Category:    "Main",
		Run: func(app *App, args Arguments) {
		},
	},
	&Command{
		Name:        "close_window",
		Description: "Closes the active window",
		Category:    "Main",
		Run: func(app *App, args Arguments) {
			if app.ViewManager.activeWindow != nil {
				app.ViewManager.CloseActiveWindow()
			}
		},
	},
	&Command{
		Name:        "quit",
		Description: "Quit discorder",
		Category:    "Main",
		Run: func(app *App, args Arguments) {
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
