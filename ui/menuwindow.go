package ui

import (
	"github.com/jonas747/discorder/common"
	"github.com/nsf/termbox-go"
	"log"
)

type MenuItem struct {
	Str      string
	Marked   bool
	Selected bool
	Info     string
	UserData interface{}
}

type MenuWindow struct {
	*BaseEntity
	Window *Window

	MainContainer *AutoLayoutContainer
	TopContainer  *Container

	InfoText *Text

	// Style
	StyleNormal         AttribPair
	StyleMarked         AttribPair
	StyleSelected       AttribPair
	StyleMarkedSelected AttribPair

	Options []*MenuItem

	Selected int

	texts []*Text
	Dirty bool
}

func NewMenuWindow() *MenuWindow {
	mw := &MenuWindow{
		BaseEntity:    &BaseEntity{},
		MainContainer: NewAutoLayoutContainer(),
		TopContainer:  NewContainer(),
		InfoText:      NewText(),
	}

	window := NewWindow()
	window.Transform.AnchorMax = common.NewVector2F(1, 1)

	mw.Transform.AddChildren(window)
	mw.Window = window

	mw.MainContainer.ForceExpandWidth = true
	mw.TopContainer.Dynamic = true
	mw.TopContainer.AllowZeroSize = true

	window.Transform.AddChildren(mw.MainContainer)
	mw.MainContainer.Transform.AnchorMax = common.NewVector2F(1, 1)

	mw.MainContainer.Transform.AddChildren(mw.TopContainer, mw.InfoText)
	mw.InfoText.Text = "THIS IS INFOOO TEXT"
	return mw
}

// Makes sure index is within len(options)
func (lw *MenuWindow) CheckBounds(index int) int {
	if index < 0 {
		return 0
	}
	if index >= len(lw.Options) {
		return len(lw.Options) - 1
	}
	return index
}

func (lw *MenuWindow) GetIndex(item *MenuItem) int {
	for k, v := range lw.Options {
		if item == v {
			return k
		}
	}

	return -1
}

func (lw *MenuWindow) RemoveMarked(index int) {
	index = lw.CheckBounds(index)
	lw.Options[index].Marked = false

	lw.Dirty = true
}

func (lw *MenuWindow) AddMarked(index int) {
	index = lw.CheckBounds(index)
	lw.Options[index].Marked = true

	lw.Dirty = true
}

func (lw *MenuWindow) SetSelected(selected int) {
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

func (lw *MenuWindow) GetSelected() *MenuItem {
	index := lw.CheckBounds(lw.Selected)
	return lw.Options[index]
}

func (lw *MenuWindow) SetOptionsString(options []string) {
	lw.Options = make([]*MenuItem, len(options))
	for k, v := range options {
		lw.Options[k] = &MenuItem{
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

func (lw *MenuWindow) SetOptions(options []*MenuItem) {
	lw.Options = options
	lw.Dirty = true
}

func (lw *MenuWindow) OptionsHeight() int {
	h := 0
	rect := lw.Transform.GetRect()
	for _, v := range lw.Options {
		h += HeightRequired(v.Str, int(rect.W))
	}
	return h
}

func (lw *MenuWindow) Rebuild() {
	//lw.ClearChildren()
	//lw.AddChild(lw.Window)
	lw.TopContainer.Transform.ClearChildren(true)

	lw.texts = make([]*Text, len(lw.Options))

	requiredHeight := lw.OptionsHeight()
	rect := lw.Transform.GetRect()
	_, termSizeY := termbox.Size()

	y := 0
	if requiredHeight > termSizeY || requiredHeight > int(rect.H) {
		// If window is taller then scroll
		y = int(float64(requiredHeight)*(float64(len(lw.Options)-lw.Selected)/float64(len(lw.Options)))) - int(rect.H/2)
	}

	for k, option := range lw.Options {
		t := NewText()
		t.Text = option.Str
		t.Transform.Position.Y = float32(y)
		t.Transform.AnchorMax.X = 1

		y += t.HeightRequired()
		log.Println(y)
		lw.texts[k] = t
		lw.TopContainer.Transform.AddChildren(t)

		switch {
		case option.Selected && option.Marked:
			t.FG = lw.StyleMarkedSelected.FG
			t.BG = lw.StyleMarkedSelected.BG
		case option.Selected:
			t.FG = lw.StyleSelected.FG
			t.BG = lw.StyleSelected.BG
		case option.Marked:
			t.FG = lw.StyleMarked.FG
			t.BG = lw.StyleMarked.BG
		default:
			t.FG = lw.StyleNormal.FG
			t.BG = lw.StyleNormal.BG
		}
	}
}

func (lw *MenuWindow) HandleInput(event termbox.Event) {
	if event.Type == termbox.EventResize {
		lw.Dirty = true
	}
}

func (lw *MenuWindow) Destroy() { lw.DestroyChildren() }
func (lw *MenuWindow) Update() {
	if lw.Dirty {
		lw.Rebuild()
		lw.InfoText.Text = lw.GetSelected().Info
		lw.Dirty = false
	}
}

func (lw *MenuWindow) Scroll(dir Direction, amount int) {
	switch dir {
	case DirUp:
		lw.SetSelected(lw.Selected - amount)
	case DirDown:
		lw.SetSelected(lw.Selected + amount)
	case DirEnd:
		lw.SetSelected(len(lw.Options) - 1)
	case DirStart:
		lw.SetSelected(0)
	}
}
