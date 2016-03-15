package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/nsf/termbox-go"
	"log"
)

type StateNormal struct {
	app *App
}

func (s *StateNormal) HandleInput(event termbox.Event) {
	if event.Type == termbox.EventKey {
		if event.Key == termbox.KeyCtrlS {
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
		} else if event.Key == termbox.KeyCtrlH {
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
		} else {
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

type StateListSelection struct {
	app          *App
	Options      []string
	Header       string
	OnSelect     func(app *App, selectedIndex int, selectedStr string)
	curSelection int
}

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
