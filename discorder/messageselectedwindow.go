package discorder

import (
	"github.com/jonas747/discorder/common"
	"github.com/jonas747/discorder/ui"
	"github.com/jonas747/discordgo"
	"github.com/nsf/termbox-go"
	"log"
)

type MessageSelectedWindow struct {
	*ui.SimpleEntity
	App        *App
	msg        *discordgo.Message
	listWindow *ui.ListWindow
}

func NewMessageSelectedWindow(app *App, msg *discordgo.Message) *MessageSelectedWindow {
	lw := ui.NewListWindow()

	msw := &MessageSelectedWindow{
		SimpleEntity: ui.NewSimpleEntity(),
		App:          app,
		msg:          msg,
		listWindow:   lw,
	}
	msw.AddChild(lw)

	options := []string{
		"Message " + GetMessageAuthor(msg),
		"Edit - Not added yet :(",
		"Remove - Not added yet :(",
		"Copy link[x] - Not added yet :(",
	}

	lw.SetOptionsString(options)

	lw.Transform.AnchorMax = common.NewVector2F(0.5, 0.5)
	lw.Transform.AnchorMin = common.NewVector2F(0.5, 0.5)

	width := 70
	height := 20

	lw.Transform.Size = common.NewVector2I(70, 20)
	lw.Transform.Position = common.NewVector2I(-width/2, -height/2)
	lw.Window.Title = "ID: " + msg.ID

	return msw
}

func (mw *MessageSelectedWindow) HandleInput(event termbox.Event) {
	if event.Type == termbox.EventKey {
		if event.Key == termbox.KeyEnter {
			option := mw.listWindow.Selected
			switch option {
			case 0:
				if mw.msg.Author == nil {
					log.Println("Sorry it appears that theres no author for this message? e.e")
				} else {
					log.Println("Should message", GetMessageAuthor(mw.msg))
					mw.InitiateConvo(mw.msg.Author.ID)
				}
			default:
				log.Println("This hasn't been implemented yet :(")
			}
			mw.App.ViewManager.CloseActiveWindow()
		}
	}
}

func (mw *MessageSelectedWindow) InitiateConvo(userId string) {
	// Check private channels first
	state := mw.App.session.State
	state.RLock()
	for _, v := range state.PrivateChannels {
		if v.Recipient.ID == userId {
			mw.App.ViewManager.mv.AddChannel(v.ID)
			mw.App.ViewManager.talkingChannel = v.ID
			state.RUnlock()
			return
		}
	}
	state.RUnlock()

	// Create one then
	channel, err := mw.App.session.UserChannelCreate(userId)
	if err != nil {
		log.Println("Error creating userchannel", err)
		return
	}
	state.ChannelAdd(channel)

	mw.App.ViewManager.mv.AddChannel(channel.ID)
	mw.App.ViewManager.talkingChannel = channel.ID
}
