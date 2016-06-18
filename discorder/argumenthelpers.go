package discorder

import (
	"encoding/json"
	"fmt"
	"github.com/jonas747/discorder/common"
	"github.com/jonas747/discorder/ui"
	"github.com/jonas747/discordgo"
	"io/ioutil"
	"log"
	"path/filepath"
)

type ArgumentCallback func(result string)

type ArgumentHelper interface {
	Run(app *App, uiLayer int, callback ArgumentCallback)
	GetName() string // Returns the name of the helper, to be used with preprovided info
}

type CustomCommandWindow interface {
	Run(app *App, uiLayer int)
}

type ServerChannelArgumentHelper struct {
	Channel bool // Select channel and not server

	app   *App
	layer int
	cb    ArgumentCallback
}

func (s *ServerChannelArgumentHelper) GetName() string {
	if s.Channel {
		return "channel"
	}
	return "server"
}

func (s *ServerChannelArgumentHelper) Run(app *App, uiLayer int, callback ArgumentCallback) {
	s.app = app
	s.layer = uiLayer
	s.cb = callback

	ssw := NewSelectServerWindow(s.app, nil, s.layer+2)
	s.app.ViewManager.AddWindow(ssw)
	if s.Channel {
		ssw.Mode = ServerSelectModeChannelOnly
	} else {
		ssw.Mode = ServerSelectModeServerOnly
	}

	ssw.OnSelect = func(sel interface{}) {
		id := ""
		if s.Channel {
			cast, ok := sel.(*discordgo.Channel)
			if !ok {
				return
			}
			id = cast.ID
		} else {
			cast, ok := sel.(*discordgo.Guild)
			if !ok {
				return
			}
			id = cast.ID
		}

		s.cb(id)
		s.app.ViewManager.RemoveWindow(ssw)
	}
}

type MessageArgumentHelper struct{}

func (m *MessageArgumentHelper) Run(app *App, layer int, cb ArgumentCallback) {}

func (m *MessageArgumentHelper) GetName() string {
	return "message"
}

type UserArgumentHelper struct{}

func (m *UserArgumentHelper) Run(app *App, layer int, cb ArgumentCallback) {}

func (m *UserArgumentHelper) GetName() string {
	return "user"
}

type ThemeCommandWindow struct {
}

func (t *ThemeCommandWindow) Run(app *App, layer int) {
	themeNames, err := app.GetAvailableThemes()
	if err != nil {
		log.Println("Error reading theme dir", filepath.Join(app.configDir, "themes"), err)
		return
	}

	menuWindow := ui.NewMenuWindow(layer, app.ViewManager.UIManager, true)

	menuWindow.Transform.AnchorMax = common.NewVector2F(1, 1)
	menuWindow.Transform.Left = 1
	menuWindow.Transform.Right = 2
	app.ApplyThemeToMenu(menuWindow)

	menuWindow.Window.Title = "Select a theme"

	var items []*ui.MenuItem
	for _, themeName := range themeNames {
		item := &ui.MenuItem{
			Name: themeName,
			Info: "Parsing...",
		}
		items = append(items, item)
		go t.ParseThemeAndApply(app, filepath.Join(app.configDir, "themes", themeName), item)
	}

	menuWindow.SetOptions(items)
	app.ViewManager.AddWindow(menuWindow)

	menuWindow.OnSelect = func(item *ui.MenuItem) {
		if item == nil {
			return
		}
		if item.UserData == nil {
			log.Println("This theme hasn't been parsed yet, check log for errors")
			return
		}
		theme, ok := item.UserData.(*Theme)
		if !ok {
			log.Println("Failed acessing theme?")
			return
		}

		app.SetUserTheme(theme)
		app.ViewManager.RemoveWindow(menuWindow)
	}
}

func (t *ThemeCommandWindow) ParseThemeAndApply(app *App, path string, item *ui.MenuItem) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println("Failed reading theme file", path, err)
		return
	}

	var theme *Theme
	err = json.Unmarshal(file, &theme)
	if err != nil {
		log.Println("Failed decoding theme json", path, err)
		return
	}

	app.Lock()
	item.Info = fmt.Sprintf("%s By %s\n%s\nColor Mode: %d", theme.Name, theme.Author, theme.Comment, theme.ColorMode)
	item.UserData = theme
	app.Unlock()
}
