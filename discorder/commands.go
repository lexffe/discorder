package discorder

var (
	CommandOpenCommands = &Command{
		Name:        "OpenCommands",
		Description: "Opens up the command window with all commands available",
		Category:    "Hidden",
		Run:         func(app *App, args []*Argument) {},
	}

	CommandOpenServers = &Command{
		Name:        "OpenServers",
		Description: "Opens up the command window with all commands available",
		Category:    "Hidden",
		Run: func(app *App, args []*Argument) {
			if app.ViewManager.CanOpenWindow() {
				ssw := NewSelectServerWindow(app, app.ViewManager.SelectedMessageView)
				app.ViewManager.SetActiveWindow(ssw)
			}
		},
	}
)
