package discorder

import (
	"github.com/jonas747/discordgo"
)

type ArgumentCallback func(result string)

type ArgumentHelper interface {
	Run(app *App, uiLayer int, callback ArgumentCallback)
	GetName() string // Returns the name of the helper, to be used with preprovided info
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
