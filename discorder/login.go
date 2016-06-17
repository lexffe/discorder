package discorder

import (
	"github.com/jonas747/discorder/ui"
	"log"
)

func OpenLoginWindow(app *App) *CommandWindow {
	cmdWindow := NewCommandWindow(app, 5, nil, "Because of discord api restrictions i can not add 2 factor authentication, they made it very clear that only official clients can use that api endpoint so go nag them and not me about it\n\nITo grab a token from the official client, open the developer tools with 'ctrl-i' and enter this into the console\n'localStorage.getItem(\"token\")'\nThen copy the token without the quotes\n")
	cmdWindow.menuWindow.SearchInput.SetActive(false)
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
			&ArgumentDef{Name: "token", Description: "The token", Datatype: ui.DataTypeString},
		},
		RunFunc: func(app *App, args Arguments) {
			token, _ := args.String("token")
			if token == "" {
				log.Println("No token provided")
				return
			} else {
				log.Println("Trying to log in using a token, If it gets stuck here either the api is having problems or the token is invalid")
				err := app.Login("", "", token)
				if err != nil {
					log.Println("Failed logging in with token:", err)
				}
			}
		},
	},
	&SimpleCommand{
		Name:           "Email and password login (no 2fa)",
		Description:    "Login using email and password",
		CustomExecText: "Login",
		Args: []*ArgumentDef{
			&ArgumentDef{Name: "Email", Description: "Your email", Datatype: ui.DataTypeString},
			&ArgumentDef{Name: "Password", Description: "Your email", Datatype: ui.DataTypePassword},
		},
		RunFunc: func(app *App, args Arguments) {
			email, _ := args.String("Email")
			pw, _ := args.String("Password")
			if email == "" {
				log.Println("Email empty")
				return
			}

			if pw == "" {
				log.Println("Password empty")
				return
			}
			err := app.Login(email, pw, "")
			log.Println("Trying to log in using email and password")
			if err != nil {
				log.Println("Error logging in", err)
			} else {
				log.Println("Sucessfully logged in!")
			}
		},
	},
}
