package discorder

import (
	"fmt"
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

	menuWindow := ui.NewMenuWindow(layer, app.ViewManager.UIManager)

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

	ssw.GenMenu()
	//height := float32(menuWindow.OptionsHeight() + 5)

	app.ViewManager.UIManager.AddWindow(ssw)

	return ssw
}

func (ssw *ServerSelectWindow) GenMenu() {
	state := ssw.App.session.State
	state.RLock()
	defer state.RUnlock()

	if len(state.Guilds) < 1 {
		log.Println("No guilds, probably starting up still...")
		return
	}

	// Generate guild options
	rootOptions := make([]*ui.MenuItem, len(state.Guilds)+1)
	for k, guild := range state.Guilds {
		guildOption := &ui.MenuItem{
			Name:     guild.Name,
			IsDir:    true,
			UserData: guild,
			Info:     fmt.Sprintf("Members: %d\nID:%s", len(guild.Members), guild.ID),
			Children: make([]*ui.MenuItem, len(guild.Channels)),
		}

		// Generate chanel options
		for i, channel := range guild.Channels {
			marked := false
			for _, listening := range ssw.messageView.Channels {
				if listening == channel.ID {
					marked = true
					break
				}
			}
			channelOption := &ui.MenuItem{
				Name:     "#" + channel.Name,
				UserData: channel,
				Info:     fmt.Sprintf("Topic %s", channel.Topic),
				Marked:   marked,
			}
			guildOption.Children[i] = channelOption
			if marked {
				guildOption.Marked = true
			}
		}
		rootOptions[k+1] = guildOption
	}
	rootOptions[0] = &ui.MenuItem{
		Name:        "Direct Messages",
		Highlighted: true,
		IsDir:       true,
		Children:    make([]*ui.MenuItem, len(state.PrivateChannels)),
	}

	for i, channel := range state.PrivateChannels {
		marked := false
		if ssw.messageView.ShowAllPrivate {
			marked = true
		} else {
			for _, listening := range ssw.messageView.Channels {
				if listening == channel.ID {
					marked = true
					break
				}
			}
		}

		channelOption := &ui.MenuItem{
			Name:     GetChannelNameOrRecipient(channel),
			UserData: channel,
			Info:     fmt.Sprintf("Topic %s", channel.Topic),
			Marked:   marked,
		}
		if marked {
			rootOptions[0].Marked = true
		}
		rootOptions[0].Children[i] = channelOption
	}

	ssw.menuWindow.SetOptions(rootOptions)
}

func (ssw *ServerSelectWindow) HandleInput(event termbox.Event) {
	if event.Type == termbox.EventKey {
		switch event.Key {
		case termbox.KeyEnter:
			// The below does not strictly belong here does it?
			selected := ssw.menuWindow.GetHighlighted()

			userdata, ok := selected.UserData.(*discordgo.Guild)

			var window *ChannelSelectWindow
			if ok {
				window = NewChannelSelectWindow(ssw.App, ssw.messageView, userdata.ID)
			} else {
				window = NewChannelSelectWindow(ssw.App, ssw.messageView, "")
			}

			ssw.App.ViewManager.Transform.RemoveChild(ssw, true)
			ssw.App.ViewManager.Transform.AddChildren(window)
		case termbox.KeySpace:
			// The below does not strictly belong here does it?
			selected := ssw.menuWindow.GetHighlighted()
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

func (ssw *ServerSelectWindow) Destroy() {
	ssw.App.ViewManager.UIManager.RemoveWindow(ssw)
	ssw.DestroyChildren()
}

func (ssw *ServerSelectWindow) Back() {
	if len(ssw.menuWindow.CurDir) < 1 {
		ssw.Transform.Parent.RemoveChild(ssw, true)
	}
}
