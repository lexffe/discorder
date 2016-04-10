package ui

import (
	"github.com/nsf/termbox-go"
)

const (
	TextModeOverflow = iota
	TextModeHide
	TextModeWrap
)

type Text struct {
	*BaseEntity

	Transform *Transform
	Text      string

	Mode int

	Attribs map[int]AttribPair
	BG, Fg  termbox.Attribute // Not used if Attribs is specified

	Layer int
}

func NewUIText() *Text {
	return &Text{
		BaseEntity: &BaseEntity{},
		Transform:  &Transform{},
	}
}

func (t *Text) GetDrawLayer() int {
	return t.Layer
}

func (t *Text) Draw() {
	rect := t.Transform.GetRect()

	var attribs map[int]AttribPair
	if t.Attribs != nil && len(t.Attribs) > 0 {
		attribs = t.Attribs
	} else {
		attribs = map[int]AttribPair{
			0: AttribPair{t.Fg, t.BG},
		}
	}

	cellSlice := GenCellSlice(t.Text, attribs)
	SetCells(cellSlice, int(rect.X), int(rect.Y), int(rect.W), int(rect.H))
}

func (t *Text) Destroy() {}
