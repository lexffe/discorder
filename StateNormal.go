package main

// import (
// 	"github.com/bwmarrin/discordgo"
// 	"github.com/nsf/termbox-go"
// 	"log"
// 	"strings"
// 	"time"
// 	"unicode/utf8"
// )

// type StateNormal struct {
// 	app                     *App
// 	lastTypingSent          time.Time
// 	mentionSelect           int
// 	isAutocompletingMention bool
// 	mentionMatches          []*discordgo.Member
// }

// func (s *StateNormal) Enter() {}
// func (s *StateNormal) Exit()  {}

// func (s *StateNormal) HandleInput(event termbox.Event) {
// 	if event.Type == termbox.EventKey {
// 		switch event.Key {
// 		case termbox.KeyEnter:
// 			// send
// 			if s.isAutocompletingMention {
// 				if s.PerformAutocompleteMention() {
// 					s.isAutocompletingMention = false
// 					break
// 				}
// 			}

// 			cp := s.app.currentTextBuffer
// 			s.app.currentTextBuffer = ""
// 			s.app.currentCursorLocation = 0
// 			s.app.RefreshDisplay()
// 			_, err := s.app.session.ChannelMessageSend(s.app.selectedChannelId, cp)
// 			if err != nil {
// 				log.Println("Error sending: ", err)
// 			}
// 		case termbox.KeyCtrlS:
// 			// Select server
// 			if len(s.app.session.State.Guilds) < 0 {
// 				log.Println("No guilds, Most likely starting up still...")
// 				return
// 			}
// 			s.app.SetState(&StateSelectServer{app: s.app})
// 		case termbox.KeyCtrlG:
// 			// Select channel
// 			if s.app.selectedGuild == nil {
// 				log.Println("No valid server selected")
// 				return
// 			}
// 			s.app.SetState(&StateSelectChannel{app: s.app})
// 		case termbox.KeyCtrlP:
// 			// Select private message channel
// 			s.app.SetState(&StateSelectPrivateChannel{app: s.app})
// 		case termbox.KeyCtrlR:
// 		// quick respond or return
// 		case termbox.KeyCtrlO:
// 			// help
// 			s.app.SetState(&StateHelp{s.app})
// 		case termbox.KeyCtrlJ:
// 			go s.app.GetHistory(s.app.selectedChannelId, 10, "", "")
// 		case termbox.KeyCtrlL:
// 			s.app.logBuffer = make([]*LogMessage, 0)
// 		case termbox.KeyArrowUp:
// 			s.app.curChatScroll++
// 		case termbox.KeyArrowDown:
// 			s.app.curChatScroll--
// 			if s.app.curChatScroll < 0 {
// 				s.app.curChatScroll = 0
// 			}
// 		case termbox.KeyTab:
// 			if s.isAutocompletingMention {
// 				s.mentionSelect++
// 				if s.mentionSelect > len(s.mentionMatches)-1 {
// 					s.mentionSelect = 0
// 				}
// 			}
// 		default:
// 			// Otherwise delegate it to the text input handler
// 			s.app.HandleTextInput(event)

// 			// Autocomplete stuff
// 			split := strings.Split(s.app.currentTextBuffer, " ")
// 			currentIndex := s.FindMatchSubIndex(split)
// 			//log.Println(currentIndex, split[currentIndex])
// 			if len(split[currentIndex]) > 0 && split[currentIndex][0] == '@' {
// 				if !s.isAutocompletingMention {
// 					s.isAutocompletingMention = true
// 					s.mentionMatches = make([]*discordgo.Member, 0)
// 					s.mentionSelect = 0
// 				}

// 				s.FindMatchingMentions(currentIndex)
// 			} else {
// 				s.isAutocompletingMention = false
// 			}
// 		}
// 	}
// }
// func (s *StateNormal) RefreshDisplay() {
// 	// Typing status update
// 	if s.app.currentTextBuffer != "" && time.Since(s.lastTypingSent) > time.Second*2 {
// 		go s.app.session.ChannelTyping(s.app.selectedChannelId)
// 		s.lastTypingSent = time.Now()
// 	}

// 	// Send prompt
// 	preStr := "Send To " + s.app.selectedChannelId + ":"
// 	if s.app.selectedChannel != nil {
// 		preStr = "Send to #" + getChannelName(s.app.selectedChannel) + ":"
// 	}
// 	sizeX, sizeY := termbox.Size()
// 	DrawPrompt(preStr, 0, sizeY-1, sizeX, s.app.currentCursorLocation, s.app.currentTextBuffer, termbox.ColorDefault, termbox.ColorDefault)

// 	// @ mention autocompletion
// 	if s.isAutocompletingMention && s.app.selectedGuild != nil {
// 		s.DrawAutocomplete()
// 	}
// }

// // Dirty and messy implementation of mention autocompleting, but hey it works! somewhat..
// func (s *StateNormal) DrawAutocomplete() {
// 	selectedStart := 0
// 	selectedEnd := 0
// 	foundSelected := false
// 	fullString := ""
// 	for k, v := range s.mentionMatches {
// 		if !foundSelected {
// 			if k == s.mentionSelect {
// 				foundSelected = true
// 				selectedEnd = selectedStart + utf8.RuneCountInString(v.User.Username) + 1
// 			} else {
// 				selectedStart += utf8.RuneCountInString(v.User.Username) + 1
// 			}
// 		}
// 		fullString += v.User.Username + " "
// 	}
// 	attribs := map[int]AttribPoint{
// 		0: AttribPoint{termbox.ColorGreen, termbox.ColorDefault},
// 	}
// 	if selectedEnd != 0 && selectedEnd != selectedStart {
// 		attribs[selectedStart] = AttribPoint{termbox.ColorGreen, termbox.ColorYellow}
// 		attribs[selectedEnd] = AttribPoint{termbox.ColorGreen, termbox.ColorDefault}
// 	}
// 	cellSlice := GenCellSlice(fullString, attribs)
// 	sizeX, sizeY := termbox.Size()
// 	SetCells(cellSlice, 0, sizeY-2, sizeX, 1)
// }

// func (s *StateNormal) FindMatchingMentions(subIndex int) {
// 	split := strings.Split(s.app.currentTextBuffer, " ")

// 	if len(split[subIndex]) < 2 {
// 		s.mentionMatches = []*discordgo.Member{}
// 		return
// 	}
// 	strToSearchFor := split[subIndex][1:]

// 	matches := make([]*discordgo.Member, 0)
// 	for _, member := range s.app.selectedGuild.Members {
// 		if strings.Contains(strings.ToLower(member.User.Username), strings.ToLower(strToSearchFor)) {
// 			matches = append(matches, member)
// 		}
// 	}
// 	s.mentionMatches = matches
// }

// func (s *StateNormal) PerformAutocompleteMention() bool {
// 	if s.mentionSelect > len(s.mentionMatches)-1 {
// 		return false
// 	}

// 	split := strings.Split(s.app.currentTextBuffer, " ")
// 	currentIndex := s.FindMatchSubIndex(split)

// 	selected := s.mentionMatches[s.mentionSelect]

// 	//curBuffer := s.app.currentTextBuffer

// 	split[currentIndex] = "<@" + selected.User.ID + ">"
// 	s.app.currentCursorLocation = len(split) - 1
// 	str := ""
// 	for k, v := range split {
// 		if k <= currentIndex {
// 			s.app.currentCursorLocation += utf8.RuneCountInString(v)
// 		}
// 		str += v + " "
// 	}
// 	str = str[:len(str)-1]

// 	s.app.currentTextBuffer = str
// 	return true
// }

// func (s *StateNormal) FindMatchSubIndex(split []string) int {
// 	i := 0
// 	currentIndex := 0
// 	for k, v := range split {
// 		i += utf8.RuneCountInString(v)
// 		if i >= s.app.currentCursorLocation-k {
// 			currentIndex = k
// 			break
// 		}
// 	}
// 	return currentIndex
// }
