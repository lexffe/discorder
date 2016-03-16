package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/nsf/termbox-go"
	"log"
	"unicode/utf8"
)

type StateNormal struct {
	app *App
}

func (s *StateNormal) Start() {}

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
			state := app.session.State
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

			s.app.currentState = &StateListSelection{
				app:      s.app,
				Header:   "Select a server",
				Options:  options,
				OnSelect: serverSelected,
			}
		} else if event.Key == termbox.KeyCtrlG {
			// Select channel
			if s.app.selectedGuild == nil {
				return
			}

			options := make([]string, 0)
			for _, v := range s.app.selectedGuild.Channels {
				if v.Type == "text" {
					options = append(options, "#"+v.Name)
				}
			}

			s.app.currentState = &StateListSelection{
				app:      s.app,
				Header:   "Select a Channel",
				Options:  options,
				OnSelect: channelSelected,
			}
		} else if event.Key == termbox.KeyCtrlP {
			// Select private message channel

			state := app.session.State
			state.RLock()
			defer state.RUnlock()

			options := make([]string, len(state.PrivateChannels))
			for k, v := range state.PrivateChannels {
				options[k] = v.Recipient.Username
			}

			s.app.currentState = &StateListSelection{
				app:      s.app,
				Header:   "Select a Conversation",
				Options:  options,
				OnSelect: channelPrivateSelected,
			}
		} else if event.Key == termbox.KeyCtrlR {
			// quick respond or return
		} else if event.Key == termbox.KeyCtrlH {
			// help
			s.app.currentState = &StateHelp{app}
		} else {
			// Otherwise delegate it to the text input handler
			s.app.HandleTextInput(event)
		}
	}
}
func (s *StateNormal) RefreshDisplay() {}

func serverSelected(app *App, index int, name string) {

	state := app.session.State
	state.RLock()
	defer state.RUnlock()

	guild := state.Guilds[index]
	if guild.Name != name {
		log.Println("Name mismatch, guild list changed")
		app.currentState = &StateNormal{app}
		return
	}

	app.selectedGuild = guild
	app.selectedServerId = guild.ID

	app.currentState = &StateNormal{app}
}

func channelSelected(app *App, index int, name string) {
	state := app.session.State
	state.RLock()
	defer state.RUnlock()

	realList := make([]*discordgo.Channel, 0)
	for _, v := range app.selectedGuild.Channels {
		if v.Type == "text" {
			realList = append(realList, v)
		}
	}

	if index < len(realList) && index >= 0 {
		channel := realList[index]
		if "#"+channel.Name != name {
			log.Println("Name mismatch, channel list changed ", channel.Name, "!=", name)
			app.currentState = &StateNormal{app}
			return
		}

		app.selectedChannelId = channel.ID
		app.selectedChannel = channel
	}

	app.currentState = &StateNormal{app}
}

func channelPrivateSelected(app *App, index int, name string) {
	state := app.session.State
	state.RLock()
	defer state.RUnlock()

	if index < len(state.PrivateChannels) && index >= 0 {
		channel := state.PrivateChannels[index]
		if channel.Recipient.Username != name {
			log.Println("Name mismatch, user list changed ", channel.Name, "!=", name)
			app.currentState = &StateNormal{app}
			return
		}

		app.selectedChannelId = channel.ID
		app.selectedChannel = channel
	}

	app.currentState = &StateNormal{app}
}

type StateListSelection struct {
	app          *App
	Options      []string
	Header       string
	OnSelect     func(app *App, selectedIndex int, selectedStr string)
	curSelection int
}

func (s *StateListSelection) Start() {}

func (s *StateListSelection) HandleInput(event termbox.Event) {
	if event.Type == termbox.EventKey {
		if event.Key == termbox.KeyArrowUp {
			s.curSelection--
			if s.curSelection < 0 {
				s.curSelection = 0
			}
		} else if event.Key == termbox.KeyArrowDown {
			s.curSelection++
			if s.curSelection >= len(s.Options) {
				s.curSelection = len(s.Options) - 1
			}
		} else if event.Key == termbox.KeyEnter {
			if s.OnSelect != nil {
				s.OnSelect(s.app, s.curSelection, s.Options[s.curSelection])
			}
		} else if event.Key == termbox.KeyBackspace || event.Key == termbox.KeyBackspace2 {
			s.app.currentState = &StateNormal{s.app}
		}
	}
}

func (s *StateListSelection) RefreshDisplay() {
	if s.Header == "" {
		s.Header = "Select an item"
	}
	app.CreateListWindow(s.Header, s.Options, s.curSelection)
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

func (s *StateLogin) Start() {
	if s.app.config.Email != "" && s.app.config.Password != "" {
		s.currentlyLoggingIn = true

		err := s.app.Login(s.app.config.Email, s.app.config.Password)
		if err != nil {
			log.Println("Error logging in :", err)
		} else {
			s.app.currentState = &StateNormal{app}
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
					s.app.currentState = &StateNormal{s.app}
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

func (s *StateHelp) Start() {}
func (s *StateHelp) RefreshDisplay() {
	sizeX, sizeY := termbox.Size()

	wWidth := 70
	wHeight := 20

	startX := sizeX/2 - wWidth/2
	startY := sizeY/2 - wHeight/2
	CreateWindow("Help", startX, startY, wWidth, wHeight, termbox.ColorBlack)
	SimpleSetText(startX+1, startY+1, wWidth-2, "Keyboard shortcuts:", termbox.ColorDefault, termbox.ColorDefault)
	SimpleSetText(startX+1, startY+2, wWidth-2, "Ctrl-H: Help", termbox.ColorDefault, termbox.ColorDefault)
	SimpleSetText(startX+1, startY+3, wWidth-2, "Ctrl-S: Select server", termbox.ColorDefault, termbox.ColorDefault)
	SimpleSetText(startX+1, startY+4, wWidth-2, "Ctrl-G: Select channel", termbox.ColorDefault, termbox.ColorDefault)
	SimpleSetText(startX+1, startY+5, wWidth-2, "Ctrl-P: Select private conversation", termbox.ColorDefault, termbox.ColorDefault)
	SimpleSetText(startX+1, startY+6, wWidth-2, "Escape: Quit", termbox.ColorDefault, termbox.ColorDefault)
	SimpleSetText(startX+1, startY+7, wWidth-2, "Backspace: Close current wnidow", termbox.ColorDefault, termbox.ColorDefault)
	SimpleSetText(startX+1, startY+10, wWidth-2, "You are using Discorder version "+VERSION, termbox.ColorDefault, termbox.ColorDefault)
	SimpleSetText(startX+1, startY+11, wWidth-2, "This is still in very early development, please report any bugs you find here", termbox.ColorDefault, termbox.ColorDefault)
	SimpleSetText(startX+1, startY+13, wWidth-2, "https://github.com/jonas747/discorder", termbox.ColorDefault, termbox.ColorDefault)
}

func (s *StateHelp) HandleInput(event termbox.Event) {
	if event.Type == termbox.EventKey {
		switch event.Key {
		case termbox.KeyBackspace, termbox.KeyBackspace2:
			s.app.currentState = &StateNormal{s.app}
		}
	}
}
