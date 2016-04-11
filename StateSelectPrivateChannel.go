package main

// import (
// 	"github.com/nsf/termbox-go"
// 	"log"
// )

// type StateSelectPrivateChannel struct {
// 	app           *App
// 	listSelection *ListSelection
// }

// func (s *StateSelectPrivateChannel) Enter() {
// 	state := s.app.session.State
// 	state.RLock()
// 	defer state.RUnlock()

// 	options := make([]string, len(state.PrivateChannels))
// 	for k, v := range state.PrivateChannels {
// 		options[k] = v.Recipient.Username
// 	}

// 	s.listSelection = &ListSelection{
// 		app:     s.app,
// 		Header:  "Select a User",
// 		Options: options,
// 	}
// }
// func (s *StateSelectPrivateChannel) Exit() {}
// func (s *StateSelectPrivateChannel) HandleInput(event termbox.Event) {
// 	if event.Type == termbox.EventKey {
// 		if event.Key == termbox.KeyEnter {
// 			state := s.app.session.State
// 			state.RLock()
// 			defer state.RUnlock()

// 			index := s.listSelection.curSelection
// 			name := s.listSelection.GetCurrentSelection()

// 			if index < len(state.PrivateChannels) && index >= 0 {
// 				channel := state.PrivateChannels[index]
// 				if channel.Recipient.Username != name {
// 					log.Println("Name mismatch, user list changed ", channel.Name, "!=", name)
// 					s.app.SetState(&StateNormal{app: s.app})
// 					return
// 				}

// 				s.app.selectedChannelId = channel.ID
// 				s.app.selectedChannel = channel
// 			}

// 			s.app.SetState(&StateNormal{app: s.app})
// 		} else {
// 			s.listSelection.HandleInput(event)
// 		}
// 	}
// }
// func (s *StateSelectPrivateChannel) RefreshDisplay() {
// 	s.listSelection.RefreshDisplay()
// }
