package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/discorder/common"
	"github.com/jonas747/discorder/ui"
	"github.com/nsf/termbox-go"
	"log"
)

type ServerSelectWindow struct {
	*ui.BaseEntity
	App         *App
	listWindow  *ui.ListWindow
	messageView *MessageView
	viewManager *ViewManager
}

func NewSelectServerWindow(app *App, messageView *MessageView) *ServerSelectWindow {
	ssw := &ServerSelectWindow{
		BaseEntity:  &ui.BaseEntity{},
		App:         app,
		messageView: messageView,
	}

	state := app.session.State
	state.RLock()
	defer state.RUnlock()

	if len(state.Guilds) < 1 {
		log.Println("No guilds, probably starting up still...")
		return nil
	}

	options := make([]*ui.ListItem, len(state.Guilds)+1)
	for k, v := range state.Guilds {
		options[k+1] = &ui.ListItem{
			Str:      v.Name,
			UserData: v,
		}
	}
	options[0] = &ui.ListItem{
		Str: "Direct Messages",
	}

	listWindow := ui.NewListWindow()
	listWindow.Transform.AnchorMin = common.NewVector2F(0.1, 0.5)
	listWindow.Transform.AnchorMax = common.NewVector2F(0.9, 0.5)
	listWindow.Transform.Size.Y = float32(len(options))
	listWindow.Transform.Top = 1
	listWindow.Transform.Bottom = 1
	listWindow.SetOptions(options)
	ssw.listWindow = listWindow
	ssw.AddChild(listWindow)
	return ssw
}

func (ssw *ServerSelectWindow) HandleInput(event termbox.Event) {
	if event.Type == termbox.EventKey {
		switch event.Key {
		case termbox.KeyEnter:
			// The below does not strictly belong here does it?
			selected := ssw.listWindow.GetSelected()

			userdata, ok := selected.UserData.(*discordgo.Guild)

			var window *ChannelSelectWindow
			if ok {
				window = NewChannelSelectWindow(ssw.App, ssw.messageView, userdata.ID)
			} else {
				window = NewChannelSelectWindow(ssw.App, ssw.messageView, "")
			}

			ssw.App.ViewManager.RemoveChild(ssw, true)
			ssw.App.ViewManager.AddChild(window)
			ssw.App.ViewManager.activeWindow = window
		}
	}
}

func (ssw *ServerSelectWindow) Destroy() { ssw.DestroyChildren() }
