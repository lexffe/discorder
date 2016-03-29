package main

import (
	"github.com/nsf/termbox-go"
	"log"
	"time"
)

type StateNormal struct {
	app            *App
	lastTypingSent time.Time
}

func (s *StateNormal) Enter() {}
func (s *StateNormal) Exit()  {}

func (s *StateNormal) HandleInput(event termbox.Event) {
	if event.Type == termbox.EventKey {
		switch event.Key {
		case termbox.KeyEnter:
			// send
			cp := s.app.currentTextBuffer
			s.app.currentTextBuffer = ""
			s.app.currentCursorLocation = 0
			s.app.RefreshDisplay()
			_, err := s.app.session.ChannelMessageSend(s.app.selectedChannelId, cp)
			if err != nil {
				log.Println("Error sending: ", err)
			}
		case termbox.KeyCtrlS:
			// Select server
			if len(s.app.session.State.Guilds) < 0 {
				log.Println("No guilds, Most likely starting up still...")
				return
			}
			s.app.SetState(&StateSelectServer{app: s.app})
		case termbox.KeyCtrlG:
			// Select channel
			if s.app.selectedGuild == nil {
				log.Println("No valid server selected")
				return
			}
			s.app.SetState(&StateSelectChannel{app: s.app})
		case termbox.KeyCtrlP:
			// Select private message channel
			s.app.SetState(&StateSelectPrivateChannel{app: s.app})
		case termbox.KeyCtrlR:
		// quick respond or return
		case termbox.KeyCtrlO:
			// help
			s.app.SetState(&StateHelp{s.app})
		case termbox.KeyCtrlJ:
			go s.app.GetHistory(s.app.selectedChannelId, 10, "", "")
		case termbox.KeyCtrlL:
			s.app.logBuffer = make([]*LogMessage, 0)
		case termbox.KeyArrowUp:
			s.app.curChatScroll++
		case termbox.KeyArrowDown:
			s.app.curChatScroll--
			if s.app.curChatScroll < 0 {
				s.app.curChatScroll = 0
			}
		default:
			// Otherwise delegate it to the text input handler
			s.app.HandleTextInput(event)
		}
	}
}
func (s *StateNormal) RefreshDisplay() {
	if s.app.currentTextBuffer != "" && time.Since(s.lastTypingSent) > time.Second*2 {
		go s.app.session.ChannelTyping(s.app.selectedChannelId)
		s.lastTypingSent = time.Now()
	}

	preStr := "Send To " + s.app.selectedChannelId + ":"
	if s.app.selectedChannel != nil {
		preStr = "Send to #" + getChannelName(s.app.selectedChannel) + ":"
	}
	sizeX, sizeY := termbox.Size()
	DrawPrompt(preStr, 0, sizeY-1, sizeX, s.app.currentCursorLocation, s.app.currentTextBuffer, termbox.ColorDefault, termbox.ColorDefault)
}
