package main

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
				log.Println("Should message", GetMessageAuthor(mw.msg))
			default:
				log.Println("This hasn't been implemented yet :(")
			}
			mw.App.ViewManager.CloseActiveWindow()
		}
	}
}
