package main

import (
	"github.com/nsf/termbox-go"
)

type ListSelection struct {
	app          *App
	Options      []string
	Header       string
	curSelection int
	marked       []int
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
	if s.marked == nil {
		s.marked = []int{}
	}
	s.app.CreateListWindow(s.Header, s.Options, s.curSelection, s.marked)
}

func (s *ListSelection) GetCurrentSelection() string {
	return s.Options[s.curSelection]
}

// For logs
func (app *App) Write(p []byte) (n int, err error) {
	// since we might log from the same goroutine deadlocks may occour, should probably do a queue system or something instead...
	go func() {
		app.logChan <- string(p)
	}()

	return len(p), nil
}

func (app *App) HandleTextInput(event termbox.Event) {
	if event.Type == termbox.EventKey {
		if event.Key == termbox.KeyArrowLeft {
			app.currentCursorLocation--
			if app.currentCursorLocation < 0 {
				app.currentCursorLocation = 0
			}
		} else if event.Key == termbox.KeyArrowRight {
			app.currentCursorLocation++
			bufLen := utf8.RuneCountInString(app.currentTextBuffer)
			if app.currentCursorLocation > bufLen {
				app.currentCursorLocation = bufLen
			}
		} else if event.Key == termbox.KeyBackspace || event.Key == termbox.KeyBackspace2 {
			bufLen := utf8.RuneCountInString(app.currentTextBuffer)
			if bufLen == 0 {
				return
			}
			if app.currentCursorLocation == bufLen {
				_, size := utf8.DecodeLastRuneInString(app.currentTextBuffer)
				app.currentCursorLocation--
				app.currentTextBuffer = app.currentTextBuffer[:len(app.currentTextBuffer)-size]
			} else if app.currentCursorLocation == 1 {
				_, size := utf8.DecodeRuneInString(app.currentTextBuffer)
				app.currentCursorLocation--
				app.currentTextBuffer = app.currentTextBuffer[size:]
			} else if app.currentCursorLocation == 0 {
				return
			} else {
				runeSlice := []rune(app.currentTextBuffer)
				newSlice := append(runeSlice[:app.currentCursorLocation], runeSlice[app.currentCursorLocation+1:]...)
				app.currentTextBuffer = string(newSlice)
				app.currentCursorLocation--
			}
		} else if event.Ch != 0 || event.Key == termbox.KeySpace {
			char := event.Ch
			if event.Key == termbox.KeySpace {
				char = ' '
			}

			bufLen := utf8.RuneCountInString(app.currentTextBuffer)
			if app.currentCursorLocation == bufLen {
				app.currentTextBuffer += string(char)
				app.currentCursorLocation++
			} else if app.currentCursorLocation == 0 {
				app.currentTextBuffer = string(char) + app.currentTextBuffer
				app.currentCursorLocation++
			} else {
				bufSlice := []rune(app.currentTextBuffer)
				bufCopy := ""

				for i := 0; i < len(bufSlice); i++ {
					if i == app.currentCursorLocation {
						bufCopy += string(char)
					}
					bufCopy += string(bufSlice[i])
				}
				app.currentTextBuffer = bufCopy
				app.currentCursorLocation++
			}
		}
	}
}
