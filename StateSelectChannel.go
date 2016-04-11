package main

// import (
// 	"github.com/bwmarrin/discordgo"
// 	"github.com/nsf/termbox-go"
// 	"log"
// )

// type StateSelectChannel struct {
// 	app           *App
// 	listSelection *ListSelection
// }

// func (s *StateSelectChannel) Enter() {
// 	options := make([]string, 0)
// 	for _, v := range s.app.selectedGuild.Channels {
// 		if v.Type == "text" {
// 			options = append(options, "#"+v.Name)
// 		}
// 	}

// 	s.listSelection = &ListSelection{
// 		app:     s.app,
// 		Header:  "Select a Channel (Space: mark, Enter: select)",
// 		Options: options,
// 		Footer:  "(Space): toggle channel, (a): Toggle all",
// 	}

// 	s.SetMarked()
// }
// func (s *StateSelectChannel) Exit() {}
// func (s *StateSelectChannel) HandleInput(event termbox.Event) {
// 	if event.Type == termbox.EventKey {
// 		switch event.Key {
// 		case termbox.KeyEnter:
// 			state := s.app.session.State
// 			state.RLock()
// 			defer state.RUnlock()

// 			realList := make([]*discordgo.Channel, 0)
// 			for _, v := range s.app.selectedGuild.Channels {
// 				if v.Type == "text" {
// 					realList = append(realList, v)
// 				}
// 			}

// 			index := s.listSelection.curSelection
// 			name := s.listSelection.GetCurrentSelection()

// 			if index < len(realList) && index >= 0 {
// 				channel := realList[index]
// 				if "#"+channel.Name != name {
// 					log.Println("Name mismatch, channel list changed ", channel.Name, "!=", name)
// 					s.app.SetState(&StateNormal{app: s.app})
// 					return
// 				}
// 				s.app.AddListeningChannel(channel.ID)
// 				s.app.selectedChannelId = channel.ID
// 				s.app.selectedChannel = channel
// 			}

// 			s.app.SetState(&StateNormal{app: s.app})
// 		case termbox.KeySpace:
// 			state := s.app.session.State
// 			state.RLock()
// 			defer state.RUnlock()

// 			realList := make([]*discordgo.Channel, 0)
// 			for _, v := range s.app.selectedGuild.Channels {
// 				if v.Type == "text" {
// 					realList = append(realList, v)
// 				}
// 			}

// 			index := s.listSelection.curSelection
// 			name := s.listSelection.GetCurrentSelection()

// 			if index < len(realList) && index >= 0 {
// 				channel := realList[index]
// 				if "#"+channel.Name != name {
// 					log.Println("Name mismatch, channel list changed ", channel.Name, "!=", name)
// 					s.app.SetState(&StateNormal{app: s.app})
// 					return
// 				}

// 				s.app.ToggleListeningChannel(channel.ID)
// 				s.SetMarked()
// 			}

// 		default:
// 			switch event.Ch {
// 			case 'a':
// 				found := false
// 				for _, v := range s.listSelection.marked {
// 					if v == s.listSelection.curSelection {
// 						found = true
// 						break
// 					}
// 				}

// 				if found {
// 					s.app.listeningChannels = make([]string, 0)
// 				} else {
// 					for _, v := range s.app.selectedGuild.Channels {
// 						s.app.AddListeningChannel(v.ID)
// 					}
// 				}
// 				s.SetMarked()
// 			default:
// 				s.listSelection.HandleInput(event)
// 			}
// 		}
// 	}
// }

// func (s *StateSelectChannel) SetMarked() {
// 	newMarked := make([]int, 0)
// 	realIndex := 0
// 	for _, channel := range s.app.selectedGuild.Channels {
// 		if channel.Type == "text" {
// 			for _, listening := range s.app.listeningChannels {
// 				if listening == channel.ID {
// 					newMarked = append(newMarked, realIndex)
// 				}
// 			}
// 			realIndex++
// 		}
// 	}
// 	s.listSelection.marked = newMarked
// }

// func (s *StateSelectChannel) RefreshDisplay() {
// 	if s.listSelection == nil {
// 		return
// 	}
// 	s.listSelection.RefreshDisplay()
// }
