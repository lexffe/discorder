package ui

import (
	"github.com/nsf/termbox-go"
	"math"
	"unicode/utf8"
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

func NewText() *Text {
	return &Text{
		BaseEntity: &BaseEntity{},
		Transform:  &Transform{},
	}
}

// Helper functions
// func SimpleText(pos common.Vector2F, size common.Vector2F, text string, fg, bg termbox.Attribute, layer int) *Text {
// 	t := NewUIText()
// 	t.Transform.Position = pos
// 	t.Transform.Size = size
// 	t.Text = text
// 	t.fg = fg
// 	t.BG = bg
// 	t.Layer = layer
// 	return t
// }

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

func (t *Text) HeightRequired() int {
	rect := t.Transform.GetRect()
	num := utf8.RuneCountInString(t.Text)
	return int(math.Ceil(float64(num) / float64(rect.W)))
}

func (t *Text) Destroy() {}
