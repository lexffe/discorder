package discorder

import (
	"github.com/jonas747/discorder/ui"
)

func OpenLoginWindow(app *App) *CommandWindow {
	cmdWindow := NewCommandWindow(app, 5, nil, "Because of discord api restrictions i have disabled the normal username/password login\nIf Retrieving the token from the official client fails you can input token manually\nI'm sorry for the inconvinience")
	cmdWindow.commands = LoginCommands
	cmdWindow.categories = []*CommandCategory{}

	return cmdWindow
}

var LoginCommands = []Command{
	&SimpleCommand{
		Name:           "Token login",
		Description:    "Login using a token that was peviously aquired from somewhere else (like the official client)",
		CustomExecText: "Login",
		Args: []*ArgumentDef{
			&ArgumentDef{Name: "token", Description: "The token", Datatype: ui.DataTypeBool},
		},
		RunFunc: func(app *App, args Arguments) {
		},
	},
	&SimpleCommand{
		Name:           "Retrieve token from the official client",
		Description:    "Tries to locate the location of the official discord client and take token from the currently logged in user there",
		CustomExecText: "Login",
		Args:           []*ArgumentDef{},
		RunFunc: func(app *App, args Arguments) {
		},
	},
}
