package ui

import (
	"github.com/jonas747/discorder/common"
	"github.com/nsf/termbox-go"
	"log"
)

type MenuItem struct {
	Str         string
	Marked      bool
	Highlighted bool
	Info        string
	UserData    interface{}
}

type MenuWindow struct {
	*BaseEntity
	Window      *Window
	LowerWindow *Window

	MainContainer  *AutoLayoutContainer
	TopContainer   *Container
	LowerContainer *Container

	InfoText    *Text
	SearchInput *TextInput

	// Style
	StyleNormal         AttribPair
	StyleMarked         AttribPair
	StyleSelected       AttribPair
	StyleMarkedSelected AttribPair

	Options []*MenuItem

	Highlighted int

	texts []*Text
	Dirty bool

	Layer int
}

func NewMenuWindow(layer int, manager *Manager) *MenuWindow {
	mw := &MenuWindow{
		BaseEntity:     &BaseEntity{},
		Window:         NewWindow(manager),
		LowerWindow:    NewWindow(nil),
		MainContainer:  NewAutoLayoutContainer(),
		TopContainer:   NewContainer(),
		LowerContainer: NewContainer(),
		InfoText:       NewText(),
		Layer:          layer,
		Dirty:          true,
	}

	mw.Window.Transform.AnchorMax = common.NewVector2F(1, 1)
	mw.Window.Layer = mw.Layer
	mw.Transform.AddChildren(mw.Window)

	mw.MainContainer.ForceExpandWidth = true
	mw.TopContainer.Dynamic = true
	mw.TopContainer.AllowZeroSize = true

	mw.Window.Transform.AddChildren(mw.MainContainer)
	mw.MainContainer.Transform.AnchorMax = common.NewVector2F(1, 1)

	mw.LowerWindow.Transform.AddChildren(mw.InfoText)
	mw.LowerWindow.Transform.AnchorMax = common.NewVector2I(1, 1)

	mw.LowerWindow.Layer = layer

	mw.MainContainer.Transform.AddChildren(mw.TopContainer)

	mw.MainContainer.Transform.AddChildren(mw.LowerContainer)

	mw.LowerContainer.Transform.AddChildren(mw.LowerWindow)
	mw.LowerContainer.Transform.AnchorMax = common.NewVector2I(1, 1)
	mw.LowerContainer.ProxySize = mw.InfoText
	mw.LowerContainer.AllowZeroSize = false

	mw.InfoText.Text = "THIS IS INFOOO TEXT"
	mw.InfoText.Transform.AnchorMax = common.NewVector2I(1, 1)
	mw.InfoText.Layer = layer

	manager.AddWindow(mw)
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

func (lw *MenuWindow) SetHighlighted(Highlighted int) {
	if len(lw.Options) < 1 {
		return
	}
	// Remove previous selection
	if lw.Highlighted < len(lw.Options) && lw.Highlighted >= 0 {
		lw.Options[lw.Highlighted].Highlighted = false
	}

	Highlighted = lw.CheckBounds(Highlighted)
	lw.Options[Highlighted].Highlighted = true
	lw.Highlighted = Highlighted

	lw.Dirty = true
}

func (lw *MenuWindow) GetHighlighted() *MenuItem {
	index := lw.CheckBounds(lw.Highlighted)
	return lw.Options[index]
}

func (lw *MenuWindow) SetOptionsString(options []string) {
	lw.Options = make([]*MenuItem, len(options))
	for k, v := range options {
		lw.Options[k] = &MenuItem{
			Str:         v,
			Marked:      false,
			Highlighted: false,
		}
		if k == lw.Highlighted {
			lw.Options[k].Highlighted = true
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
	rect := lw.TopContainer.Transform.GetRect()
	_, termSizeY := termbox.Size()

	y := 0
	if requiredHeight > termSizeY || requiredHeight > int(rect.H) {
		// If window is taller then scroll
		heightPerOption := float64(requiredHeight) / float64(len(lw.Options))
		y = int(heightPerOption*(float64(len(lw.Options)-(lw.Highlighted)))) - int(rect.H*2)
		log.Println(y, heightPerOption)
	}

	for k, option := range lw.Options {
		t := NewText()
		t.Text = option.Str
		t.Transform.Position.Y = float32(y)
		t.Transform.AnchorMax.X = 1
		t.Layer = lw.Layer
		y += t.HeightRequired()

		if y >= termSizeY || y >= int(rect.H) || y <= 0 {
			// Ignore if hidden/should be hidden
			continue
		}

		lw.texts[k] = t
		lw.TopContainer.Transform.AddChildren(t)

		switch {
		case option.Highlighted && option.Marked:
			t.Style = lw.StyleMarkedSelected
		case option.Highlighted:
			t.Style = lw.StyleSelected
		case option.Marked:
			t.Style = lw.StyleMarked
		default:
			t.Style = lw.StyleNormal
		}
	}
}

func (lw *MenuWindow) HandleInput(event termbox.Event) {
	if event.Type == termbox.EventResize {
		lw.Dirty = true
	}
}

func (mw *MenuWindow) OnLayoutChanged() {
	mw.Rebuild()
}

func (lw *MenuWindow) Destroy() { lw.DestroyChildren() }
func (lw *MenuWindow) Update() {
	if lw.Dirty {
		lw.Rebuild()
		lw.InfoText.Text = lw.GetHighlighted().Info
	}
	//	lw.Dirty = false
}

func (lw *MenuWindow) Scroll(dir Direction, amount int) {
	switch dir {
	case DirUp:
		lw.SetHighlighted(lw.Highlighted - amount)
	case DirDown:
		lw.SetHighlighted(lw.Highlighted + amount)
	case DirEnd:
		lw.SetHighlighted(len(lw.Options) - 1)
	case DirStart:
		lw.SetHighlighted(0)
	}
}
