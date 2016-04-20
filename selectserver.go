package main

import (
	"github.com/jonas747/discorder/common"
	"github.com/jonas747/discorder/ui"
	"github.com/nsf/termbox-go"
	"log"
)

type ServerSelectWindow struct {
	*ui.BaseEntity
	App         *App
	listWindow  *ui.ListWindow
	messageView *ui.MessageView
}

func NewSelectServerWindow(app *App, messageView *ui.MessageView) *ServerSelectWindow {
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

	options := make([]ui.ListItem, len(state.Guilds))
	for k, v := range state.Guilds {
		options[k] = ui.ListItem{
			Str:      v.Name,
			UserData: v,
		}
		if k == 0 {
			options[k].Selected = true
		}
	}

	listWindow := ui.NewListWindow()
	listWindow.Transform.AnchorMin = common.NewVector2F(0.1, 0)
	listWindow.Transform.AnchorMax = common.NewVector2F(0.9, 1)
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
			// state := ssw.app.session.State
			// state.RLock()
			// defer state.RUnlock()

			// if ssw.listWindow >= len(state.Guilds) {
			// 	log.Println("Guild list changed while selecting.. aborting")
			// 	s.app.SetState(&StateNormal{app: s.app})
			// 	return
			// }

			// guild := state.Guilds[s.listSelection.curSelection]
			// if guild.Name != s.listSelection.GetCurrentSelection() {
			// 	log.Println("Name mismatch, guild list changed")
			// 	s.app.SetState(&StateNormal{app: s.app})
			// 	return
			// }

			// s.app.selectedGuild = guild
			// s.app.selectedServerId = guild.ID
			// s.app.listeningChannels = make([]string, 0)
			// s.app.SetState(&StateSelectChannel{app: s.app})

		}
	}
}

func (ssw *ServerSelectWindow) Destroy() { ssw.DestroyChildren() }
