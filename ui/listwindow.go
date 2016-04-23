package ui

import (
	"github.com/jonas747/discorder/common"
	"github.com/nsf/termbox-go"
)

// Set some default styles
const (
	DefaultListNormalBG = termbox.ColorDefault
	DefaultListNormalFG = termbox.ColorDefault

	DefaultListMarkedFG = termbox.ColorDefault
	DefaultListMarkedBG = termbox.ColorYellow

	DefaultListSelectedFG = termbox.ColorDefault
	DefaultListSelectedBG = termbox.ColorCyan
)

type ListItem struct {
	Str      string
	Marked   bool
	Selected bool
	UserData interface{}
}

type ListWindow struct {
	*BaseEntity
	Transform *Transform
	Window    *Window

	// Style
	NormalBG, NormalFG     termbox.Attribute
	MarkedBG, MarkedFG     termbox.Attribute
	SelectedBG, SelectedFG termbox.Attribute

	Options []*ListItem

	Selected int

	texts []*Text
	Dirty bool
}

func NewListWindow() *ListWindow {
	lw := &ListWindow{
		BaseEntity: &BaseEntity{},
		Transform:  &Transform{},
		// Defaults
		NormalFG:   DefaultListNormalFG,
		NormalBG:   DefaultListNormalBG,
		MarkedFG:   DefaultListMarkedFG,
		MarkedBG:   DefaultListMarkedBG,
		SelectedFG: DefaultListSelectedFG,
		SelectedBG: DefaultListSelectedBG,
	}

	window := NewWindow()
	window.Transform.Parent = lw.Transform
	window.Transform.AnchorMax = common.NewVector2F(1, 1)
	lw.AddChild(window)

	lw.Window = window

	return lw
}

// Makes sure index is within len(options)
func (lw *ListWindow) CheckBounds(index int) int {
	if index < 0 {
		return 0
	}
	if index >= len(lw.Options) {
		return len(lw.Options) - 1
	}
	return index
}

func (lw *ListWindow) GetIndex(item *ListItem) int {
	for k, v := range lw.Options {
		if item == v {
			return k
		}
	}

	return -1
}

func (lw *ListWindow) RemoveMarked(index int) {
	index = lw.CheckBounds(index)
	lw.Options[index].Marked = false

	lw.Dirty = true
}

func (lw *ListWindow) AddMarked(index int) {
	index = lw.CheckBounds(index)
	lw.Options[index].Marked = true

	lw.Dirty = true
}

func (lw *ListWindow) SetSelected(selected int) {
	if len(lw.Options) < 1 {
		return
	}
	// Remove previous selection
	if lw.Selected < len(lw.Options) && lw.Selected >= 0 {
		lw.Options[lw.Selected].Selected = false
	}

	selected = lw.CheckBounds(selected)
	lw.Options[selected].Selected = true
	lw.Selected = selected

	lw.Dirty = true
}

func (lw *ListWindow) GetSelected() *ListItem {
	index := lw.CheckBounds(lw.Selected)
	return lw.Options[index]
}

func (lw *ListWindow) SetOptionsString(options []string) {
	lw.Options = make([]*ListItem, len(options))
	for k, v := range options {
		lw.Options[k] = &ListItem{
			Str:      v,
			Marked:   false,
			Selected: false,
		}
		if k == lw.Selected {
			lw.Options[k].Selected = true
		}
	}
	lw.Dirty = true
}

func (lw *ListWindow) SetOptions(options []*ListItem) {
	lw.Options = options
	lw.Dirty = true
}

func (lw *ListWindow) Rebuild() {
	lw.ClearChildren()
	lw.AddChild(lw.Window)

	lw.texts = make([]*Text, len(lw.Options))

	y := 0
	for k, option := range lw.Options {
		t := NewText()
		t.Text = option.Str
		t.Transform.Position.Y = float32(y)
		t.Transform.AnchorMax.X = 1

		y += t.HeightRequired()
		lw.texts[k] = t
		lw.Window.AddChild(t)
		t.Transform.Parent = lw.Window.Transform

		switch {
		case option.Selected:
			t.FG = lw.SelectedFG
			t.BG = lw.SelectedBG
		case option.Marked:
			t.FG = lw.MarkedFG
			t.BG = lw.MarkedBG
		default:
			t.FG = lw.NormalFG
			t.BG = lw.NormalBG
		}
	}
}

func (lw *ListWindow) HandleInput(event termbox.Event) {
	if event.Type == termbox.EventKey {

		switch event.Key {
		case termbox.KeyArrowUp:
			lw.SetSelected(lw.Selected - 1)
		case termbox.KeyArrowDown:
			lw.SetSelected(lw.Selected + 1)
		}
	}
}

func (lw *ListWindow) Destroy() { lw.DestroyChildren() }
func (lw *ListWindow) PreDraw() {
	if lw.Dirty {
		lw.Rebuild()
		lw.Dirty = false
	}
}
