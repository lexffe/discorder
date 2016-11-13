package discorder

import (
	"fmt"
	"github.com/jonas747/discorder/ui"
	"github.com/jonas747/discordgo"
	"log"
	"path/filepath"
	"strconv"
)

var SimpleCommands = []Command{
	&SimpleCommand{
		Name:        "commands",
		Description: "Opens up the command window with all commands available",
		Category:    []string{"Hidden"},
		RunFunc: func(app *App, args Arguments) {
			cw := NewCommandWindow(app, 5, nil, "")
			app.ViewManager.AddWindow(cw)
		},
	},
	&SimpleCommand{
		Name:        "move_cursor",
		Description: "Moves cursor in specified direction",
		Category:    []string{"hidden"},
		Args: []*ArgumentDef{
			{Name: "direction", Optional: false, Datatype: ui.DataTypeString},
			{Name: "amount", Optional: false, Datatype: ui.DataTypeInt},
			{Name: "word", Optional: true, Datatype: ui.DataTypeBool},
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
		Category:    []string{"hidden"},
		Args: []*ArgumentDef{
			{Name: "direction", Optional: false, Datatype: ui.DataTypeString},
			{Name: "amount", Optional: false, Datatype: ui.DataTypeInt},
			{Name: "words", Optional: true, Datatype: ui.DataTypeBool},
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
		Name:        "help",
		Description: "Opens up the help window",
		Category:    []string{"Windows"},
		RunFunc: func(app *App, args Arguments) {
			if app.ViewManager.CanOpenWindow() {
				hw := NewHelpWindow(app)
				app.ViewManager.AddWindow(hw)
			}
		},
	},
	&SimpleCommand{
		Name:        "scroll",
		Description: "Scrolls currently active view",
		Category:    []string{"hidden"},
		Args: []*ArgumentDef{
			{Name: "direction", Optional: false, Datatype: ui.DataTypeString},
			{Name: "amount", Optional: false, Datatype: ui.DataTypeInt},
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
		Category:    []string{"hidden"},
		RunFunc: func(app *App, args Arguments) {
			window := app.ViewManager.UIManager.CurrentWindow()
			if window == nil {
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
		Category:    []string{"hidden"},
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
		Name:        "autocomplete_selection",
		Description: "Changes the autocomplete selection",
		Category:    []string{"hidden"},
		Args: []*ArgumentDef{
			{Name: "amount", Description: "The amoount to change in", Datatype: ui.DataTypeBool},
		},
		RunFunc: func(app *App, args Arguments) {},
	},
	&SimpleCommand{
		Name:        "clear_log",
		Description: "Clear the logbuffer",
		Category:    []string{"Utils"},
		RunFunc: func(app *App, args Arguments) {
			logRoutine.Clear()
		},
	},
	&SimpleCommand{
		Name:        "reload_theme",
		Description: "Reloads the current theme",
		Category:    []string{"Utils"},
		RunFunc: func(app *App, args Arguments) {
			userTheme := app.options.CustomThemePath
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
		Name:         "theme_window",
		Description:  "Select a theme",
		Category:     []string{"Windows"},
		CustomWindow: &ThemeCommandWindow{},
	},
	&SimpleCommand{
		Name:        "delete_message",
		Description: "Deletes a message",
		Category:    []string{"Discord"},
		Args: []*ArgumentDef{
			{Name: "last_yours", Description: "If true deletes last message you sent", Datatype: ui.DataTypeBool},
			{Name: "last_any", Description: "If true deletes last message anyone sent", Datatype: ui.DataTypeBool},
			{Name: "message", Description: "Specify a message id", Datatype: ui.DataTypeString, Helper: &MessageArgumentHelper{}},
			{Name: "channel", Description: "Specify a channel id", Datatype: ui.DataTypeString, Helper: &ServerChannelArgumentHelper{Channel: true}},
		},
		ArgCombinations: [][]string{{"last_yours"}, {"last_any"}, {"message", "channel"}},
		RunFunc: func(app *App, args Arguments) {
			// We need to be logged in
			if app.session == nil {
				return
			}

			lastOwn, _ := args.Bool("last_yours")
			lastAny, _ := args.Bool("last_any")
			messageId, _ := args.String("message")
			channelId, _ := args.String("channel")

			if messageId != "" && channelId != "" {

				err := app.session.ChannelMessageDelete(channelId, messageId)
				if err != nil {
					log.Println("Failed to delete message: ", err)
				} else {
					log.Println("Deleted message ID:", messageId, "Sucessfully")
				}
				return
			}

			// Below stuff requires an active tab
			if app.ViewManager.ActiveTab == nil {
				return
			}

			tab := app.ViewManager.ActiveTab

			if (lastAny || lastOwn) && app.session.State != nil && app.session.State.User != nil {
				for _, text := range tab.MessageView.MessageTexts {
					if text.Userdata == nil {
						continue
					}

					displayMsg, ok := text.Userdata.(*DisplayMessage)
					if !ok {
						continue
					}

					if !displayMsg.IsLogMessage &&
						((lastOwn && displayMsg.DiscordMessage.Author.ID == app.session.State.User.ID) || lastAny) {

						err := app.session.ChannelMessageDelete(displayMsg.DiscordMessage.ChannelID, displayMsg.DiscordMessage.ID)
						if err != nil {
							if err != nil {
								log.Println("Failed to delete message: ", err)
							} else {
								log.Println("Deleted message ID:", messageId, "Sucessfully")
							}
						}
						return
					}
				}
				return
			}
		},
	},
	&SimpleCommand{
		Name:           "status",
		Description:    "Updates your discord status",
		CustomExecText: "Set",
		Category:       []string{"Discord"},
		Args: []*ArgumentDef{
			{Name: "game", Description: "What game you should appear playing as", Datatype: ui.DataTypeString},
			{Name: "idle", Description: "How long you've been idle in seconds", Datatype: ui.DataTypeInt},
		},
		RunFunc: func(app *App, args Arguments) {
			game, _ := args.String("game")
			idle, _ := args.Int("idle")
			app.session.UpdateStatus(idle, game)
		},
	},
	&SimpleCommand{
		Name:        "send_message",
		Description: "Sends a message, (This is a stub, not implemented :()",
		Category:    []string{"Discord"},
		RunFunc: func(app *App, args Arguments) {
		},
	},
	&SimpleCommand{
		Name:        "initiate_conversation",
		Description: "Initiate a conversation",
		Category:    []string{"Discord"},
		Args: []*ArgumentDef{
			{Name: "user", Description: "User to intiate a conversation with", Datatype: ui.DataTypeString, Helper: &UserArgumentHelper{}},
		},
		RunFunc: func(app *App, args Arguments) {
			userId, _ := args.String("user")

			if userId == "" {
				log.Println("User empty, doing nothing..")
				return
			}

			tab := app.ViewManager.ActiveTab

			// Check private channels first
			state := app.session.State
			state.RLock()
			for _, v := range state.PrivateChannels {
				if v.Recipient.ID == userId {
					tab.MessageView.AddChannel(v.ID)
					tab.SendChannel = v.ID
					state.RUnlock()
					return
				}
			}
			state.RUnlock()

			// Create one then
			channel, err := app.session.UserChannelCreate(userId)
			if err != nil {
				log.Println("Error creating userchannel", err)
				return
			}
			state.ChannelAdd(channel)

			tab.MessageView.AddChannel(channel.ID)
			tab.SendChannel = channel.ID
		},
	},
	&SimpleCommand{
		Name:           "set_nick",
		Description:    "Sets your nickname on a server (if possible)",
		CustomExecText: "Set",
		Category:       []string{"Discord"},
		Args: []*ArgumentDef{
			{Name: "name", Description: "The nickname you will set (empty to reset)", Datatype: ui.DataTypeString},
			{Name: "server", Description: "Server to set the nickname on", Datatype: ui.DataTypeString, Helper: &ServerChannelArgumentHelper{}},
			{Name: "user", Description: "Specify a user, leave empty for youself", Datatype: ui.DataTypeString, Helper: &UserArgumentHelper{}},
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
		Name:           "pin_message",
		Description:    "Pins a message, (This is a stub, not implemented :()",
		CustomExecText: "Pin!",
		Category:       []string{"Discord"},
		Args: []*ArgumentDef{
			{Name: "message", Description: "The message that will be pinned", Datatype: ui.DataTypeString, Helper: &MessageArgumentHelper{}},
			{Name: "channel", Description: "The message that will be pinned", Datatype: ui.DataTypeString, Helper: &ServerChannelArgumentHelper{Channel: true}},
		},
		RunFunc: func(app *App, args Arguments) {
		},
	},
	&SimpleCommand{
		Name:        "back",
		Description: "Closes the active window",
		Category:    []string{"hidden"},
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
				if _, ok := window.(*MessageView); ok {
					return
				}

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
		Name:        "close_windows",
		Description: "Closes all windows",
		Category:    []string{"hidden"},
		RunFunc: func(app *App, args Arguments) {
			app.ViewManager.RemoveAllWindows()
		},
	},
	&SimpleCommand{
		Name:           "discorder_settings",
		Description:    "Change settings",
		CustomExecText: "Save",
		Args: []*ArgumentDef{
			{Name: "short_guilds", Description: "Displays a mini version of guilds in message view", Datatype: ui.DataTypeBool, CurValFunc: func(app *App) string {
				return strconv.FormatBool(app.config.ShortGuilds)
			}}, {Name: "hide_nicknames", Description: "Shows usernames instead of nicknames if true", Datatype: ui.DataTypeBool, CurValFunc: func(app *App) string {
				return strconv.FormatBool(app.config.HideNicknames)
			}}, {Name: "time_format_full", Description: "Sets the full time format", Datatype: ui.DataTypeString, CurValFunc: func(app *App) string {
				return app.config.GetTimeFormatFull()
			}}, {Name: "time_format_short", Description: "Sets the short time format (for messages on the same day)", Datatype: ui.DataTypeString, CurValFunc: func(app *App) string {
				return app.config.GetTimeFormatSameDay()
			}}, {Name: "colored_guilds", Description: "Sets the guild names to a deterministic 'random' color", Datatype: ui.DataTypeBool, CurValFunc: func(app *App) string {
				return strconv.FormatBool(app.config.ColoredGuilds)
			}}, {Name: "colored_channels", Description: "Sets the channel names to a deterministic 'random' color", Datatype: ui.DataTypeBool, CurValFunc: func(app *App) string {
				return strconv.FormatBool(app.config.ColoredChannels)
			}}, {Name: "colored_users", Description: "Sets the user names to a deterministic 'random' color", Datatype: ui.DataTypeBool, CurValFunc: func(app *App) string {
				return strconv.FormatBool(app.config.ColoredUsers)
			}},
		},
		RunFunc: func(app *App, args Arguments) {
			shortGuilds, _ := args.Bool("short_guilds")
			hideNicks, _ := args.Bool("hide_nicknames")

			formatFull, _ := args.String("time_format_full")
			formatShort, _ := args.String("time_format_short")

			coloredGuilds, _ := args.Bool("colored_guilds")
			coloredChannels, _ := args.Bool("colored_channels")
			coloredUsers, _ := args.Bool("colored_users")

			displayFormatFull := formatFull

			if formatFull == DefaultTimeFormatFull || formatFull == "" {
				formatFull = ""
				displayFormatFull = "Default"
			}

			displayFormatShort := formatShort

			if formatShort == DefaultTimeFormatSameDay || formatShort == "" {
				formatShort = ""
				displayFormatShort = "Default"
			}

			app.config.ShortGuilds = shortGuilds
			app.config.HideNicknames = hideNicks
			app.config.TimeFormatFull = formatFull
			app.config.TimeFormatSameDay = formatShort

			app.config.ColoredGuilds = coloredGuilds
			app.config.ColoredChannels = coloredChannels
			app.config.ColoredUsers = coloredUsers

			log.Printf("Set short_guilds: %v; hide_nicknames: %v, time_format_same_day: %v, time_format_full: %v", shortGuilds, hideNicks, displayFormatShort, displayFormatFull)
		},
	},
	&SimpleCommand{
		Name:        "change_tab",
		Description: "Change tab",
		Category:    []string{"hidden"},
		Args: []*ArgumentDef{
			{Name: "tab", Datatype: ui.DataTypeInt},
			{Name: "change", Datatype: ui.DataTypeInt},
		},
		RunFunc: func(app *App, args Arguments) {
			index, ok := args.Int("tab")
			if ok {
				for _, tab := range app.ViewManager.Tabs {
					if tab.Index == index {
						app.ViewManager.SetActiveTab(tab)
						return
					}
				}
				app.ViewManager.CreateTab(index)
				return
			}

			change, _ := args.Int("change")
			if change == 0 || app.ViewManager.ActiveTab == nil {
				return
			}

			curIndex := -1
			for k, tab := range app.ViewManager.Tabs {
				if tab == app.ViewManager.ActiveTab {
					curIndex = k
					break
				}
			}

			if curIndex == -1 {
				return
			}

			curIndex += change
			if curIndex < 0 {
				curIndex = 0
			} else if curIndex >= len(app.ViewManager.Tabs) {
				curIndex = len(app.ViewManager.Tabs) - 1
			}
			tab := app.ViewManager.Tabs[curIndex]
			app.ViewManager.SetActiveTab(tab)
		},
	},
	&SimpleCommand{
		Name:        "remove_tab",
		Description: "Removes the active tab",
		Category:    []string{"Utils"},
		RunFunc: func(app *App, args Arguments) {
			app.ViewManager.RemoveTab(app.ViewManager.ActiveTab)
		},
	},
	&SimpleCommand{
		Name:        "rename_tab",
		Description: "Renames the currently selected tab",
		Category:    []string{"Utils"},
		Args: []*ArgumentDef{
			{Name: "name", Description: "The name you want to give", Datatype: ui.DataTypeString},
		},
		RunFunc: func(app *App, args Arguments) {
			tab := app.ViewManager.ActiveTab
			if tab == nil {
				return
			}
			name, _ := args.String("name")
			tab.SetName(name)
		},
	},
	&SimpleCommand{
		Name:        "open_last_link",
		Description: "Opens the last link",
		Category:    []string{"Utils"},
		RunFunc: func(app *App, args Arguments) {
			tab := app.ViewManager.ActiveTab
			if tab == nil {
				return
			}

			// Below stuff requires an active tab
			if app.session.State != nil && app.session.State.User != nil {
				for _, text := range tab.MessageView.MessageTexts {
					if text.Userdata == nil {
						continue
					}

					displayMsg, ok := text.Userdata.(*DisplayMessage)
					if !ok {
						continue
					}

					if !displayMsg.IsLogMessage {
						matches := linkRegex.FindAllString(displayMsg.DiscordMessage.Content, -1)
						if len(matches) > 0 {
							go app.OpenLink(matches[0])
							return
						}

						if len(displayMsg.DiscordMessage.Attachments) > 0 {
							go app.OpenLink(displayMsg.DiscordMessage.Attachments[0].URL)
							return
						}
					}
				}
				log.Println("No recent links found :'(")
			}
		},
	},
	&SimpleCommand{
		Name:        "gen_command_table",
		Description: "Generates command docs",
		Category:    []string{"Utils"},
		RunFunc: func(app *App, args Arguments) {

			out := "| Name | Description | Category | Args |"
			out += "\n| --- | --------- | ------- | ---- |"

			for _, cmd := range app.Commands {

				category := ""
				categories := cmd.GetCategory()
				for k, v := range categories {
					category += v
					if k < len(categories)-1 {
						category += "/"
					}
				}

				argumentDefinitions := cmd.GetArgs(nil)
				argString := "<ul>"
				for _, v := range argumentDefinitions {
					argString += "<li>" + v.String() + "</li>"
				}
				argString += "</ul>"

				out += fmt.Sprintf("\n| %s | %s | %s | %s | ", cmd.GetName(), cmd.GetDescription(app), category, argString)
			}

			log.Println("Command table:\n" + out)
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

func (app *App) AddCommands() {
	app.Commands = []Command{
		&ServerNotificationSettingsCommand{app: app},
	}

	app.Commands = append(app.Commands, SimpleCommands...)
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

func (app *App) GetCommandByName(name string) Command {
	for _, cmd := range app.Commands {
		if cmd.GetName() == name {
			return cmd
		}
	}
	return nil
}
