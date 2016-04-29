package main

import (
	"github.com/jonas747/discorder/common"
	"github.com/jonas747/discorder/ui"
	"github.com/jonas747/discordgo"
	"github.com/nsf/termbox-go"
	"log"
)

const (
	ServerSelectTitle  = "Select a server"
	ServerSelectFooter = "(Space) Toggle whole server, (enter) select"
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

		// Check if were listening to one of the channel on this servers
	OUTER:
		for _, channel := range v.Channels {
			for _, listening := range messageView.Channels {
				if listening == channel.ID {
					options[k+1].Marked = true
					break OUTER // We only need 1 so no need to continue
				}
			}
		}
	}
	options[0] = &ui.ListItem{
		Str:      "Direct Messages",
		Selected: true,
	}

	listWindow := ui.NewListWindow()
	listWindow.Transform.AnchorMin = common.NewVector2F(0.1, 0.5)
	listWindow.Transform.AnchorMax = common.NewVector2F(0.9, 0.5)
	listWindow.Transform.Size.Y = float32(len(options))
	listWindow.Window.Footer = ServerSelectFooter
	listWindow.Window.Title = ServerSelectTitle
	listWindow.Transform.Position.X = -float32(len(options)) / 2

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
		case termbox.KeySpace:
			// The below does not strictly belong here does it?
			selected := ssw.listWindow.GetSelected()
			userdata, ok := selected.UserData.(*discordgo.Guild)
			if !ok {
				break
			}
			toggleTo := true
		OUTER:
			for _, v := range userdata.Channels {
				for _, c := range ssw.messageView.Channels {
					if v.ID == c {
						toggleTo = false
						break OUTER
					}
				}
			}

			for _, v := range userdata.Channels {
				if toggleTo {
					ssw.messageView.AddChannel(v.ID)
					selected.Marked = true
				} else {
					ssw.messageView.RemoveChannel(v.ID)
					selected.Marked = false
				}
			}
		}
	}
}

func (ssw *ServerSelectWindow) Destroy() { ssw.DestroyChildren() }
