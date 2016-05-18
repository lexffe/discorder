package discorder

import (
	"github.com/jonas747/discorder/common"
	"github.com/jonas747/discorder/ui"
	"github.com/jonas747/discordgo"
	"github.com/nsf/termbox-go"
	"log"
)

const (
	ChannelSelectFooter = "(a)all (space)toggle selected (enter)set sending"
	ChannelSelectTitle  = "Select Channels"
)

type ChannelSelectWindow struct {
	*ui.BaseEntity
	Guild       string
	App         *App
	listWindow  *ui.ListWindow
	messageView *MessageView
	private     bool
}

// If guild is empty, selects private channel
func NewChannelSelectWindow(app *App, mv *MessageView, guild string) *ChannelSelectWindow {
	csw := &ChannelSelectWindow{
		BaseEntity:  &ui.BaseEntity{},
		App:         app,
		Guild:       guild,
		messageView: mv,
	}
	if guild == "" {
		csw.private = true
	}

	state := app.session.State
	state.RLock()
	defer state.RUnlock()

	var channels []*discordgo.Channel
	if !csw.private {
		g, err := state.Guild(guild)
		if err != nil {
			log.Println("Error getting guild, maybe we haven't recevied ready yet?")
			return nil
		}
		channels = g.Channels
	} else {
		channels = state.PrivateChannels
	}

	options := make([]*ui.ListItem, 0)

	for k, v := range channels {
		if v.Type != "text" && !v.IsPrivate {
			continue
		}

		name := v.Name
		if v.IsPrivate {
			name = v.Recipient.Username
		}
		item := &ui.ListItem{
			Str:      name,
			UserData: v,
		}
		if k == 0 {
			item.Selected = true
		}

		// Check if we are listening to it
		for _, listeningChannel := range mv.Channels {
			if listeningChannel == v.ID {
				item.Marked = true
			}
		}

		options = append(options, item)
	}

	listWindow := ui.NewListWindow()
	listWindow.Transform.AnchorMin = common.NewVector2F(0.1, 0.5)
	listWindow.Transform.AnchorMax = common.NewVector2F(0.9, 0.5)
	listWindow.Transform.Size.Y = float32(len(options))

	listWindow.Transform.Position.Y = -float32(len(options)) / 2
	listWindow.Window.Title = ChannelSelectTitle
	listWindow.Window.Footer = ChannelSelectFooter

	app.theme.ApplyList(listWindow)

	listWindow.SetOptions(options)
	csw.listWindow = listWindow
	csw.AddChild(listWindow)
	csw.CheckAll()
	return csw
}

func (csw *ChannelSelectWindow) HandleInput(event termbox.Event) {
	if event.Type == termbox.EventKey {
		switch event.Key {
		case termbox.KeyEnter:
			selected := csw.listWindow.GetSelected()
			userdata, ok := selected.UserData.(*discordgo.Channel)
			if ok {
				log.Println("Selected ", GetChannelNameOrRecipient(userdata))
				csw.App.ViewManager.talkingChannel = userdata.ID
			}
		case termbox.KeySpace:
			selected := csw.listWindow.GetSelected()
			// Toggle
			csw.TogggleMarked(selected)
		default:
			switch event.Ch {
			case 'a', 'A':
				to := false
				for k, v := range csw.listWindow.Options {
					if k == 0 {
						to = v.Marked
					}
					if to == v.Marked {
						csw.TogggleMarked(v)
					}
				}
			}
		}
		csw.CheckAll()
	}
}

func (csw *ChannelSelectWindow) CheckAll() {
	if csw.private {
		all := true
		for _, v := range csw.listWindow.Options {
			if !v.Marked {
				all = false
				break
			}
		}

		if all {
			csw.App.ViewManager.mv.ShowAllPrivate = true
			csw.listWindow.Window.Title = ChannelSelectTitle + " ALLPRIVATEMODE: ON"
		} else {
			csw.App.ViewManager.mv.ShowAllPrivate = false
			csw.listWindow.Window.Title = ChannelSelectTitle + " ALLPRIVATEMODE: OFF"
		}
	}
}

func (csw *ChannelSelectWindow) TogggleMarked(item *ui.ListItem) {
	item.Marked = !item.Marked
	csw.listWindow.Dirty = true

	channel, ok := item.UserData.(*discordgo.Channel)
	if !ok {
		return
	}

	if item.Marked {
		csw.messageView.AddChannel(channel.ID)
	} else {
		csw.messageView.RemoveChannel(channel.ID)
	}
	// Reflect changes to messageview
}

func (csw *ChannelSelectWindow) Destroy() { csw.DestroyChildren() }
