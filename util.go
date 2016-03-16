package main

import (
	"github.com/nsf/termbox-go"
)

type ListSelection struct {
	app          *App
	Options      []string
	Header       string
	curSelection int
}

func (s *ListSelection) HandleInput(event termbox.Event) {
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
		} else if event.Key == termbox.KeyBackspace || event.Key == termbox.KeyBackspace2 {
			s.app.currentState = &StateNormal{s.app}
		}
	}
}

func (s *ListSelection) RefreshDisplay() {
	if s.Header == "" {
		s.Header = "Select an item"
	}
	s.app.CreateListWindow(s.Header, s.Options, s.curSelection)
}

func (s *ListSelection) GetCurrentSelection() string {
	return s.Options[s.curSelection]
}
