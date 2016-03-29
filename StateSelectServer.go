package main

import (
	"github.com/nsf/termbox-go"
	"log"
)

type StateSelectServer struct {
	app           *App
	listSelection *ListSelection
}

func (s *StateSelectServer) Enter() {
	state := s.app.session.State
	state.RLock()
	defer state.RUnlock()

	if len(state.Guilds) < 1 {
		log.Println("No guilds, probably starting up still...")
		return
	}

	options := make([]string, len(state.Guilds))
	for k, v := range state.Guilds {
		options[k] = v.Name
	}

	s.listSelection = &ListSelection{
		app:     s.app,
		Header:  "Select a Server",
		Options: options,
	}
}
func (s *StateSelectServer) Exit() {}
func (s *StateSelectServer) HandleInput(event termbox.Event) {
	if event.Type == termbox.EventKey {
		if event.Key == termbox.KeyEnter {
			state := s.app.session.State
			state.RLock()
			defer state.RUnlock()

			if s.listSelection.curSelection >= len(state.Guilds) {
				log.Println("Guild list changed while selecting.. aborting")
				s.app.SetState(&StateNormal{s.app})
				return
			}

			guild := state.Guilds[s.listSelection.curSelection]
			if guild.Name != s.listSelection.GetCurrentSelection() {
				log.Println("Name mismatch, guild list changed")
				s.app.SetState(&StateNormal{s.app})
				return
			}

			s.app.selectedGuild = guild
			s.app.selectedServerId = guild.ID
			s.app.listeningChannels = make([]string, 0)
			s.app.SetState(&StateSelectChannel{app: s.app})
		} else {
			s.listSelection.HandleInput(event)
		}
	}
}
func (s *StateSelectServer) RefreshDisplay() {
	if s.listSelection == nil {
		return
	}
	s.listSelection.RefreshDisplay()
}
