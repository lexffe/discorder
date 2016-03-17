package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/nsf/termbox-go"
	"log"
	"unicode/utf8"
)

type State interface {
	Enter()
	Exit()
	HandleInput(event termbox.Event)
	RefreshDisplay()
}

type StateNormal struct {
	app *App
}

func (s *StateNormal) Enter() {}
func (s *StateNormal) Exit()  {}

func (s *StateNormal) HandleInput(event termbox.Event) {
	if event.Type == termbox.EventKey {
		if event.Key == termbox.KeyEnter {
			// send
			cp := s.app.currentTextBuffer
			s.app.currentTextBuffer = ""
			s.app.currentCursorLocation = 0
			s.app.RefreshDisplay()
			_, err := s.app.session.ChannelMessageSend(s.app.selectedChannelId, cp)
			if err != nil {
				log.Println("Error sending: ", err)
			}
		} else if event.Key == termbox.KeyCtrlS {
			// Select server
			if len(s.app.session.State.Guilds) < 0 {
				log.Println("No guilds, Most likely starting up still...")
				return
			}
			s.app.SetState(&StateSelectServer{app: s.app})
		} else if event.Key == termbox.KeyCtrlG {
			// Select channel
			if s.app.selectedGuild == nil {
				log.Println("No valid server selected")
				return
			}
			s.app.SetState(&StateSelectChannel{app: s.app})
		} else if event.Key == termbox.KeyCtrlP {
			// Select private message channel
			s.app.SetState(&StateSelectPrivateChannel{app: s.app})
		} else if event.Key == termbox.KeyCtrlR {
			// quick respond or return
		} else if event.Key == termbox.KeyCtrlH {
			// help
			s.app.SetState(&StateHelp{s.app})
		} else {
			// Otherwise delegate it to the text input handler
			s.app.HandleTextInput(event)
		}
	}
}
func (s *StateNormal) RefreshDisplay() {}

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
			s.app.SetState(&StateNormal{s.app})
		} else {
			s.listSelection.HandleInput(event)
		}
	}
}
func (s *StateSelectServer) RefreshDisplay() {
	s.listSelection.RefreshDisplay()
}

type StateSelectChannel struct {
	app           *App
	listSelection *ListSelection
}

func (s *StateSelectChannel) Enter() {
	options := make([]string, 0)
	for _, v := range s.app.selectedGuild.Channels {
		if v.Type == "text" {
			options = append(options, "#"+v.Name)
		}
	}

	s.listSelection = &ListSelection{
		app:     s.app,
		Header:  "Select a Channel",
		Options: options,
	}

	s.SetMarked()
}
func (s *StateSelectChannel) Exit() {}
func (s *StateSelectChannel) HandleInput(event termbox.Event) {
	if event.Type == termbox.EventKey {
		if event.Key == termbox.KeyEnter {
			state := s.app.session.State
			state.RLock()
			defer state.RUnlock()

			realList := make([]*discordgo.Channel, 0)
			for _, v := range s.app.selectedGuild.Channels {
				if v.Type == "text" {
					realList = append(realList, v)
				}
			}

			index := s.listSelection.curSelection
			name := s.listSelection.GetCurrentSelection()

			if index < len(realList) && index >= 0 {
				channel := realList[index]
				if "#"+channel.Name != name {
					log.Println("Name mismatch, channel list changed ", channel.Name, "!=", name)
					s.app.SetState(&StateNormal{s.app})
					return
				}

				s.app.selectedChannelId = channel.ID
				s.app.selectedChannel = channel
			}

			s.app.SetState(&StateNormal{s.app})
		} else if event.Key == termbox.KeySpace {
			state := s.app.session.State
			state.RLock()
			defer state.RUnlock()

			realList := make([]*discordgo.Channel, 0)
			for _, v := range s.app.selectedGuild.Channels {
				if v.Type == "text" {
					realList = append(realList, v)
				}
			}

			index := s.listSelection.curSelection
			name := s.listSelection.GetCurrentSelection()

			if index < len(realList) && index >= 0 {
				channel := realList[index]
				if "#"+channel.Name != name {
					log.Println("Name mismatch, channel list changed ", channel.Name, "!=", name)
					s.app.SetState(&StateNormal{s.app})
					return
				}

				index := -1
				for i, listening := range s.app.listeningChannels {
					if listening == channel.ID {
						index = i
						break
					}
				}

				if index != -1 {
					if index == 0 {
						s.app.listeningChannels = s.app.listeningChannels[1:]
					} else if index == len(s.app.listeningChannels)-1 {
						s.app.listeningChannels = s.app.listeningChannels[:len(s.app.listeningChannels)-1]
					} else {
						s.app.listeningChannels = append(s.app.listeningChannels[:index], s.app.listeningChannels[index+1:]...)
					}
				} else {
					s.app.listeningChannels = append(s.app.listeningChannels, channel.ID)
				}
				s.SetMarked()
			}
		} else {
			s.listSelection.HandleInput(event)
		}
	}
}

func (s *StateSelectChannel) SetMarked() {
	newMarked := make([]int, 0)
	realIndex := 0
	for _, channel := range s.app.selectedGuild.Channels {
		if channel.Type == "text" {
			for _, listening := range s.app.listeningChannels {
				if listening == channel.ID {
					newMarked = append(newMarked, realIndex)
				}
			}
			realIndex++
		}
	}
	s.listSelection.marked = newMarked
}

func (s *StateSelectChannel) RefreshDisplay() {
	s.listSelection.RefreshDisplay()
}

type StateSelectPrivateChannel struct {
	app           *App
	listSelection *ListSelection
}

func (s *StateSelectPrivateChannel) Enter() {
	state := s.app.session.State
	state.RLock()
	defer state.RUnlock()

	options := make([]string, len(state.PrivateChannels))
	for k, v := range state.PrivateChannels {
		options[k] = v.Recipient.Username
	}

	s.listSelection = &ListSelection{
		app:     s.app,
		Header:  "Select a User",
		Options: options,
	}
}
func (s *StateSelectPrivateChannel) Exit() {}
func (s *StateSelectPrivateChannel) HandleInput(event termbox.Event) {
	if event.Type == termbox.EventKey {
		if event.Key == termbox.KeyEnter {
			state := s.app.session.State
			state.RLock()
			defer state.RUnlock()

			index := s.listSelection.curSelection
			name := s.listSelection.GetCurrentSelection()

			if index < len(state.PrivateChannels) && index >= 0 {
				channel := state.PrivateChannels[index]
				if channel.Recipient.Username != name {
					log.Println("Name mismatch, user list changed ", channel.Name, "!=", name)
					s.app.SetState(&StateNormal{s.app})
					return
				}

				s.app.selectedChannelId = channel.ID
				s.app.selectedChannel = channel
			}

			s.app.SetState(&StateNormal{s.app})
		} else {
			s.listSelection.HandleInput(event)
		}
	}
}
func (s *StateSelectPrivateChannel) RefreshDisplay() {
	s.listSelection.RefreshDisplay()
}

var (
	InputStateEmail    = 0
	InputStatePassword = 1
)

type StateLogin struct {
	app                *App
	currentlyLoggingIn bool

	pwBuffer      string
	curInputState int
	savePassword  bool
	err           error
}

func (s *StateLogin) Exit() {}

func (s *StateLogin) Enter() {
	if s.app.config.Email != "" && s.app.config.Password != "" {
		s.currentlyLoggingIn = true

		err := s.app.Login(s.app.config.Email, s.app.config.Password)
		if err != nil {
			log.Println("Error logging in :", err)
		} else {
			s.app.SetState(&StateNormal{s.app})
		}
		s.currentlyLoggingIn = false
	} else {
		if s.app.config.Email != "" {
			s.curInputState = InputStatePassword
		} else {
			s.curInputState = InputStateEmail
		}
	}
}

func (s *StateLogin) HandleInput(event termbox.Event) {
	if event.Type == termbox.EventKey {
		switch event.Key {
		case termbox.KeyEnter:
			if s.curInputState == InputStateEmail {
				s.app.config.Email = s.app.currentTextBuffer
				s.app.currentTextBuffer = s.pwBuffer
				s.curInputState = InputStatePassword
				s.app.currentCursorLocation = 0
			} else {
				pw := s.app.currentTextBuffer
				if s.savePassword {
					s.app.config.Password = pw
				}
				err := s.app.Login(s.app.config.Email, pw)
				if err != nil {
					s.err = err
				} else {
					s.app.config.Save(configPath)
					s.app.currentTextBuffer = ""
					s.app.currentCursorLocation = 0
					s.app.SetState(&StateNormal{s.app})
				}
			}
		case termbox.KeyCtrlS:
			if s.curInputState == InputStateEmail {
				s.app.config.Email = s.app.currentTextBuffer
				s.app.currentTextBuffer = s.pwBuffer
				s.curInputState = InputStatePassword
				s.app.currentCursorLocation = 0
			} else {
				s.pwBuffer = s.app.currentTextBuffer
				s.app.currentTextBuffer = s.app.config.Email
				s.curInputState = InputStateEmail
				s.app.currentCursorLocation = 0
			}
		case termbox.KeyCtrlT:
			s.savePassword = !s.savePassword
		default:
			s.app.HandleTextInput(event)
		}
	}
}

func (s *StateLogin) RefreshDisplay() {
	sizeX, sizeY := termbox.Size()

	startX := sizeX/2 - 25
	startY := sizeY/2 - 5
	CreateWindow("Login", startX, startY, 50, 10, termbox.ColorBlack)

	if s.currentlyLoggingIn {
		SimpleSetText(startX+1, startY+1, 48, "Logging in", termbox.ColorDefault, termbox.ColorBlack)
		SimpleSetText(startX+1, startY+2, 48, "Using email: "+s.app.config.Email, termbox.ColorDefault, termbox.ColorBlack)
	} else {
		SimpleSetText(startX+1, startY+1, 48, "Enter email", termbox.ColorDefault, termbox.ColorBlack)
		if s.curInputState == InputStateEmail {
			s.app.Prompt(startX+1, startY+4, 150, s.app.currentCursorLocation, s.app.currentTextBuffer)
		} else {
			SimpleSetText(startX+1, startY+1, 48, s.app.config.Email, termbox.ColorDefault, termbox.ColorBlack)
			SimpleSetText(startX+1, startY+2, 48, "Enter Password", termbox.ColorDefault, termbox.ColorBlack)
			str := ""
			for i := 0; i < utf8.RuneCountInString(s.app.currentTextBuffer); i++ {
				str += "*"
			}
			s.app.Prompt(startX+1, startY+4, 150, s.app.currentCursorLocation, str)
		}

		SimpleSetText(startX+1, startY+5, 48, "Ctrl-S to switch input", termbox.ColorGreen, termbox.ColorBlack)
		passSaveStr := "on"
		if !s.savePassword {
			passSaveStr = "off"
		}
		SimpleSetText(startX+1, startY+6, 48, "Ctrl-t toggle password saving("+passSaveStr+")", termbox.ColorGreen, termbox.ColorBlack)
	}

	if s.err != nil {
		SimpleSetText(startX+1, startY+7, 48, s.err.Error(), termbox.ColorRed, termbox.ColorBlack)
	}
}

type StateHelp struct {
	app *App
}

func (s *StateHelp) Enter() {}
func (s *StateHelp) Exit()  {}
func (s *StateHelp) RefreshDisplay() {
	sizeX, sizeY := termbox.Size()

	wWidth := 70
	wHeight := 20

	startX := sizeX/2 - wWidth/2
	startY := sizeY/2 - wHeight/2

	curY := startY + 1
	CreateWindow("Help", startX, startY, wWidth, wHeight, termbox.ColorBlack)
	curY += SimpleSetText(startX+1, curY, wWidth-2, "Keyboard shortcuts:", termbox.ColorDefault, termbox.ColorDefault)
	curY += SimpleSetText(startX+1, curY, wWidth-2, "Ctrl-H: Help", termbox.ColorDefault, termbox.ColorDefault)
	curY += SimpleSetText(startX+1, curY, wWidth-2, "Ctrl-S: Select server", termbox.ColorDefault, termbox.ColorDefault)
	curY += SimpleSetText(startX+1, curY, wWidth-2, "Ctrl-G: Select channels", termbox.ColorDefault, termbox.ColorDefault)
	curY += SimpleSetText(startX+1, curY, wWidth-2, "	Space: mark channel", termbox.ColorDefault, termbox.ColorDefault)
	curY += SimpleSetText(startX+1, curY, wWidth-2, "	Enter: Select as send channelf", termbox.ColorDefault, termbox.ColorDefault)
	curY += SimpleSetText(startX+1, curY, wWidth-2, "Ctrl-P: Select private conversation", termbox.ColorDefault, termbox.ColorDefault)
	curY += SimpleSetText(startX+1, curY, wWidth-2, "Ctrl-Q: Quit", termbox.ColorDefault, termbox.ColorDefault)
	curY += SimpleSetText(startX+1, curY, wWidth-2, "Backspace: Close current wnidow", termbox.ColorDefault, termbox.ColorDefault)
	curY++
	curY += SimpleSetText(startX+1, curY, wWidth-2, "You are using Discorder version "+VERSION, termbox.ColorDefault, termbox.ColorDefault)
	curY += SimpleSetText(startX+1, curY, wWidth-2, "This is still in very early development, please report any bugs you find here", termbox.ColorDefault, termbox.ColorDefault)
	curY += SimpleSetText(startX+1, curY, wWidth-2, "https://github.com/jonas747/discorder", termbox.ColorDefault, termbox.ColorDefault)
}

func (s *StateHelp) HandleInput(event termbox.Event) {
	if event.Type == termbox.EventKey {
		switch event.Key {
		case termbox.KeyBackspace, termbox.KeyBackspace2:
			s.app.SetState(&StateNormal{s.app})
		}
	}
}
