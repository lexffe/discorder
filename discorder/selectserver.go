package discorder

import (
	"github.com/jonas747/discorder/common"
	//"github.com/jonas747/discorder/common"
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
	menuWindow  *ui.MenuWindow
	messageView *MessageView
	viewManager *ViewManager
	Layer       int
}

func NewSelectServerWindow(app *App, messageView *MessageView, layer int) *ServerSelectWindow {
	ssw := &ServerSelectWindow{
		BaseEntity:  &ui.BaseEntity{},
		App:         app,
		messageView: messageView,
		Layer:       layer,
	}

	state := app.session.State
	state.RLock()
	defer state.RUnlock()

	if len(state.Guilds) < 1 {
		log.Println("No guilds, probably starting up still...")
		return nil
	}

	options := make([]*ui.MenuItem, len(state.Guilds)+1)
	for k, v := range state.Guilds {
		options[k+1] = &ui.MenuItem{
			Str:      v.Name,
			Info:     v.Name + "\n" + v.ID,
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
	options[0] = &ui.MenuItem{
		Str:      "Direct Messages",
		Selected: true,
	}

	menuWindow := ui.NewMenuWindow(layer)
	menuWindow.SetOptions(options)

	menuWindow.Transform.AnchorMax = common.NewVector2F(1, 1)
	menuWindow.Transform.Top = 1
	menuWindow.Transform.Bottom = 2

	menuWindow.Window.Footer = ServerSelectFooter
	menuWindow.Window.Title = ServerSelectTitle

	app.ApplyThemeToMenu(menuWindow)

	ssw.menuWindow = menuWindow
	ssw.Transform.AddChildren(menuWindow)

	ssw.Transform.AnchorMin = common.NewVector2F(0.1, 0)
	ssw.Transform.AnchorMax = common.NewVector2F(0.9, 1)

	//height := float32(menuWindow.OptionsHeight() + 5)

	return ssw
}

func (ssw *ServerSelectWindow) HandleInput(event termbox.Event) {
	if event.Type == termbox.EventKey {
		switch event.Key {
		case termbox.KeyEnter:
			// The below does not strictly belong here does it?
			selected := ssw.menuWindow.GetSelected()

			userdata, ok := selected.UserData.(*discordgo.Guild)

			var window *ChannelSelectWindow
			if ok {
				window = NewChannelSelectWindow(ssw.App, ssw.messageView, userdata.ID)
			} else {
				window = NewChannelSelectWindow(ssw.App, ssw.messageView, "")
			}

			ssw.App.ViewManager.Transform.RemoveChild(ssw, true)
			ssw.App.ViewManager.SetActiveWindow(window)
		case termbox.KeySpace:
			// The below does not strictly belong here does it?
			selected := ssw.menuWindow.GetSelected()
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
				if v.Type != "text" && !v.IsPrivate {
					continue
				}

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
