package discorder

import (
	"github.com/jonas747/discorder/ui"
	"github.com/jonas747/discordgo"
	"log"
	"path/filepath"
	"strconv"
)

var Commands = []Command{
	&SimpleCommand{
		Name:        "commands",
		Description: "Opens up the command window with all commands available",
		Category:    []string{"Hidden"},
		RunFunc: func(app *App, args Arguments) {
			cw := NewCommandWindow(app, 5)
			app.ViewManager.AddWindow(cw)
		},
	},
	&SimpleCommand{
		Name:        "move_cursor",
		Description: "Moves cursor in specified direction",
		Category:    []string{"Misc"},
		Args: []*ArgumentDef{
			&ArgumentDef{Name: "direction", Optional: false, Datatype: ui.DataTypeString},
			&ArgumentDef{Name: "amount", Optional: false, Datatype: ui.DataTypeInt},
			&ArgumentDef{Name: "word", Optional: true, Datatype: ui.DataTypeBool},
		},
		RunFunc: func(app *App, args Arguments) {
			amount, _ := args.Int("amount")
			words, _ := args.Bool("words")
			dir, _ := args.String("direction")
			moveDir := StringToDir(dir)
			input := app.ViewManager.UIManager.ActiveInput
			if input != nil && input.Active {
				input.MoveCursor(moveDir, amount, words)
			}
		},
	},
	&SimpleCommand{
		Name:        "erase",
		Description: "Erase text",
		Category:    []string{"Misc"},
		Args: []*ArgumentDef{
			&ArgumentDef{Name: "direction", Optional: false, Datatype: ui.DataTypeString},
			&ArgumentDef{Name: "amount", Optional: false, Datatype: ui.DataTypeInt},
			&ArgumentDef{Name: "words", Optional: true, Datatype: ui.DataTypeBool},
		},
		RunFunc: func(app *App, args Arguments) {
			amount, _ := args.Int("amount")
			words, _ := args.Bool("words")
			dir, _ := args.String("direction")

			moveDir := StringToDir(dir)
			input := app.ViewManager.UIManager.ActiveInput
			if input != nil && input.Active {
				input.Erase(moveDir, amount, words)
			}
		},
	},
	&SimpleCommand{
		Name:        "servers",
		Description: "Opens up the server window",
		Category:    []string{"Windows"},
		RunFunc: func(app *App, args Arguments) {
			if app.ViewManager.CanOpenWindow() {
				ssw := NewSelectServerWindow(app, app.ViewManager.ActiveTab.MessageView, 6)
				app.ViewManager.AddWindow(ssw)
				ssw.OnSelect = func(element interface{}) {
					cast, ok := element.(*discordgo.Channel)
					if !ok {
						return
					}

					log.Println("Selected ", GetChannelNameOrRecipient(cast))
					app.ViewManager.ActiveTab.SendChannel = cast.ID
				}
			}
		},
	},
	&SimpleCommand{
		Name:        "channels",
		Description: "Opens up the channel window",
		Category:    []string{"Windows"},
		RunFunc: func(app *App, args Arguments) {
			if app.ViewManager.CanOpenWindow() {
				// ssw := NewChannelSelectWindow(app, app.ViewManager.SelectedMessageView, guild)
				// app.ViewManager.SetActiveWindow(ssw)
			}
		},
	},
	&SimpleCommand{
		Name:        "help",
		Description: "Opens up the help window",
		RunFunc: func(app *App, args Arguments) {
			if app.ViewManager.CanOpenWindow() {
				hw := NewHelpWindow(app)
				app.ViewManager.AddWindow(hw)
			}
		},
	},
	&SimpleCommand{
		Name:        "message_window",
		Description: "Opens message window",
		Category:    []string{"Misc"},
		Args: []*ArgumentDef{
			&ArgumentDef{Name: "message", Optional: true, Datatype: ui.DataTypeString},
		},
		RunFunc: func(app *App, args Arguments) {
			//app.ViewManager.SelectedMessageView.OpenMessageSelectWindow("")
		},
	},
	&SimpleCommand{
		Name:        "scroll",
		Description: "Scrolls currently active view",
		Category:    []string{"Misc"},
		Args: []*ArgumentDef{
			&ArgumentDef{Name: "direction", Optional: false, Datatype: ui.DataTypeString},
			&ArgumentDef{Name: "amount", Optional: false, Datatype: ui.DataTypeInt},
		},
		RunFunc: func(app *App, args Arguments) {
			amount, _ := args.Int("amount")
			dir, _ := args.String("direction")
			moveDir := StringToDir(dir)
			window := app.ViewManager.UIManager.CurrentWindow()
			if window != nil {
				ui.RunFuncCondTraverse(window, func(e ui.Entity) bool {
					scrollable, ok := e.(ui.Scrollable)
					if ok {
						scrollable.Scroll(moveDir, amount)
						return false
					}
					return true
				})
			} else if app.ViewManager.ActiveTab != nil {
				app.ViewManager.ActiveTab.MessageView.Scroll(moveDir, amount)
			}
		},
	},
	&SimpleCommand{
		Name:        "select",
		Description: "Select the currently highlighted element",
		Category:    []string{"Misc"},
		RunFunc: func(app *App, args Arguments) {
			window := app.ViewManager.UIManager.CurrentWindow()
			if window == nil {
				app.ViewManager.SendFromTextBuffer()
				return
			}

			ui.RunFuncCond(window, func(e ui.Entity) bool {
				cast, ok := e.(ui.SelectAble)
				if ok {
					cast.Select()
					return false
				}

				return true
			})
		},
	},
	&SimpleCommand{
		Name:        "toggle",
		Description: "Toggles the currently highlited element",
		Category:    []string{"Misc"},
		RunFunc: func(app *App, args Arguments) {
			window := app.ViewManager.UIManager.CurrentWindow()
			if window == nil {
				return
			}

			ui.RunFuncCond(window, func(e ui.Entity) bool {
				cast, ok := e.(ui.ToggleAble)
				if ok {
					cast.Toggle()
					return false
				}

				return true
			})
		},
	},
	&SimpleCommand{
		Name:        "clear_log",
		Description: "Clear the logbuffer",
		RunFunc: func(app *App, args Arguments) {
			logRoutine.Clear()
		},
	},
	&SimpleCommand{
		Name:        "reload_theme",
		Description: "Reloads the current theme",
		RunFunc: func(app *App, args Arguments) {
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
	&SimpleCommand{
		Name:        "delete",
		Description: "Deletes a message",
		Category:    []string{"Discord"},
	},
	&SimpleCommand{
		Name:        "status",
		Description: "Updates your discord status",
		Category:    []string{"Discord"},
		Args: []*ArgumentDef{
			&ArgumentDef{Name: "game", Description: "What game you should appear playing as", Datatype: ui.DataTypeString},
			&ArgumentDef{Name: "idle", Description: "How long you've been idle in seconds", Datatype: ui.DataTypeInt},
		},
		RunFunc: func(app *App, args Arguments) {
			game, _ := args.String("game")
			idle, _ := args.Int("idle")
			app.session.UpdateStatus(idle, game)
		},
	},
	&SimpleCommand{
		Name:        "send_message",
		Description: "Sends a message",
		Category:    []string{"Discord"},
		RunFunc: func(app *App, args Arguments) {
		},
	},
	&SimpleCommand{
		Name:        "set_nick",
		Description: "Sets your nickname on a server (if possible)",
		Category:    []string{"Discord"},
		Args: []*ArgumentDef{
			&ArgumentDef{Name: "name", Description: "The nickname you will set (empty to reset)", Datatype: ui.DataTypeString},
			&ArgumentDef{Name: "server", Description: "Server to set the nickname on", Datatype: ui.DataTypeString, HelperDataType: HelperDataTypeServer},
			&ArgumentDef{Name: "user", Description: "Specify a user, leave empty for youself", Datatype: ui.DataTypeString},
		},
		RunFunc: func(app *App, args Arguments) {
			serverId, _ := args.String("server")
			name, _ := args.String("name")
			user, _ := args.String("user")

			userId := "@me/nick"
			if user != "" {
				userId = user
			}

			err := app.session.GuildMemberNickname(serverId, userId, name)
			if err != nil {
				log.Println("Error setting nickname", err)
			}
		},
	},
	&SimpleCommand{
		Name:        "back",
		Description: "Closes the active window",
		Category:    []string{"Misc"},
		RunFunc: func(app *App, args Arguments) {
			window := app.ViewManager.UIManager.CurrentWindow()
			if window == nil {
				return
			}

			handled := false
			ui.RunFuncCond(window, func(e ui.Entity) bool {
				cast, ok := e.(ui.BackHandler)
				if ok {
					handled = cast.Back()
					if handled {
						return false
					}
				}

				return true
			})

			if !handled { // Do the default action
				parent := window.GetTransform().Parent
				if parent == app.ViewManager.menuContainer.GetTransform() {
					app.ViewManager.RemoveWindow(window)
				} else {
					parent.RemoveChild(window, true)
				}
			}
		},
	},
	&SimpleCommand{
		Name:        "short_guilds",
		Description: "Displays a mini version of guilds in message view",
		Args: []*ArgumentDef{
			&ArgumentDef{Name: "on", Datatype: ui.DataTypeBool, CurVal: func(app *App) string {
				return strconv.FormatBool(app.config.ShortGuilds)
			}},
		},
		RunFunc: func(app *App, args Arguments) {
			shortguilds, _ := args.Bool("on")
			app.config.ShortGuilds = shortguilds
			log.Println("Set short_guilds to", shortguilds)
		},
	},
	&SimpleCommand{
		Name:        "change_tab",
		Description: "Displays a mini version of guilds in message view",
		Args: []*ArgumentDef{
			&ArgumentDef{Name: "tab", Datatype: ui.DataTypeInt},
		},
		RunFunc: func(app *App, args Arguments) {
			index, _ := args.Int("tab")
			for _, tab := range app.ViewManager.Tabs {
				if tab.Index == index {
					app.ViewManager.SetActiveTab(tab)
					return
				}
			}
			app.ViewManager.CreateTab(index)
		},
	},
	&SimpleCommand{
		Name:        "quit",
		Description: "Quit discorder",
		RunFunc: func(app *App, args Arguments) {
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

func GetCommandByName(name string) Command {
	for _, cmd := range Commands {
		if cmd.GetName() == name {
			return cmd
		}
	}
	return nil
}
