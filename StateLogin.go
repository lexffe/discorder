package main

import (
	"github.com/nsf/termbox-go"
	"log"
	"unicode/utf8"
)

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
	DrawWindow("Login", ":)", startX, startY, 50, 10, termbox.ColorBlack)

	if s.currentlyLoggingIn {
		SimpleSetText(startX+1, startY+1, 48, "Logging in", termbox.ColorDefault, termbox.ColorBlack)
		SimpleSetText(startX+1, startY+2, 48, "Using email: "+s.app.config.Email, termbox.ColorDefault, termbox.ColorBlack)
	} else {
		SimpleSetText(startX+1, startY+1, 48, "Enter email", termbox.ColorDefault, termbox.ColorBlack)
		if s.curInputState == InputStateEmail {
			DrawPrompt("", startX+1, startY+4, 150, s.app.currentCursorLocation, s.app.currentTextBuffer, termbox.ColorYellow, termbox.ColorDefault)
		} else {
			SimpleSetText(startX+1, startY+1, 48, s.app.config.Email, termbox.ColorDefault, termbox.ColorBlack)
			SimpleSetText(startX+1, startY+2, 48, "Enter Password", termbox.ColorDefault, termbox.ColorBlack)
			str := ""
			for i := 0; i < utf8.RuneCountInString(s.app.currentTextBuffer); i++ {
				str += "*"
			}
			DrawPrompt("", startX+1, startY+4, 150, s.app.currentCursorLocation, str, termbox.ColorYellow, termbox.ColorDefault)
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
