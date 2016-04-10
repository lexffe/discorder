package ui

import (
	"github.com/jonas747/discorder/common"
	"github.com/nsf/termbox-go"
)

type ListWindow struct {
	*BaseEntity
	Transform *Transform
	Window    *Window

	// Style
	NormalBG, NormalFG     termbox.Attribute
	MarkedBG, MarkedFG     termbox.Attribute
	SelectedBG, SelectedFG termbox.Attribute

	Options []string

	marked   []int
	selected int

	texts []*Text
}

func NewListWindow(options []string) *ListWindow {
	lw := &ListWindow{
		BaseEntity: &BaseEntity{},
		Transform:  &Transform{},
		Options:    options,
	}

	window := NewWindow()
	window.Transform.Parent = lw.Transform
	window.Transform.AnchorMax = common.NewVector2F(1, 1)
	lw.entities = append(lw.entities, window)

	lw.Window = window

	return lw
}

func (lw *ListWindow) SetMarked(marked []int) {
}

func (lw *ListWindow) SetSelected(selected int) {

}

func (lw *ListWindow) HandleInput(event termbox.Event) {
	if event.Type == termbox.EventKey {

		switch event.Key {
		case termbox.KeyArrowUp:
			lw.selected--
			if lw.selected < 0 {
				lw.selected = 0
			}
		case termbox.KeyArrowDown:
			lw.selected++
			if lw.selected >= len(lw.Options) {
				lw.selected = len(lw.Options) - 1
			}
		}
	}
}

func (lw *ListWindow) Destroy() { lw.DestroyChildren() }
func (lw *ListWindow) Init()    {}

// type ListSelection struct {
// 	app          *App
// 	Options      []string
// 	Header       string
// 	Footer       string
// 	curSelection int
// 	marked       []int
// }

// func (s *ListSelection) HandleInput(event termbox.Event) {
// 	if event.Type == termbox.EventKey {
// 		if event.Key == termbox.KeyArrowUp {
// 			s.curSelection--
// 			if s.curSelection < 0 {
// 				s.curSelection = 0
// 			}
// 		} else if event.Key == termbox.KeyArrowDown {
// 			s.curSelection++
// 			if s.curSelection >= len(s.Options) {
// 				s.curSelection = len(s.Options) - 1
// 			}
// 		} else if event.Key == termbox.KeyBackspace || event.Key == termbox.KeyBackspace2 {
// 			s.app.currentState = &StateNormal{app: s.app}
// 		}
// 	}
// }

// func (s *ListSelection) RefreshDisplay() {
// 	if s.Header == "" {
// 		s.Header = "Select an item"
// 	}
// 	if s.marked == nil {
// 		s.marked = []int{}
// 	}
// 	CreateListWindow(s.Header, s.Footer, s.Options, s.curSelection, s.marked)
// }

// func (s *ListSelection) GetCurrentSelection() string {
// 	return s.Options[s.curSelection]
// }
