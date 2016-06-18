package ui

import (
	"github.com/jonas747/discorder/common"
	"github.com/jonas747/go-runewidth"
	"github.com/jonas747/termbox-go"
)

const (
	TextModeOverflow = iota
	TextModeHide
	TextModeWrap
)

type Text struct {
	*BaseEntity
	Disabled bool // won't draw then

	Text string

	SkipLines int

	Mode int

	// If attribs is empty, uses style instead
	attribs []*AttribPair
	Style   AttribPair

	Layer    int
	Userdata interface{}
}

func NewText() *Text {
	t := &Text{
		BaseEntity: &BaseEntity{},
	}
	return t
}

func (t *Text) GetDrawLayer() int {
	return t.Layer
}

func (t *Text) SetAttribs(attribs map[int]AttribPair) {
	highest := 0
	for key, _ := range attribs {
		if key > highest {
			highest = key
		}
	}

	t.attribs = make([]*AttribPair, highest+1)
	for key, pair := range attribs {
		c := pair
		t.attribs[key] = &c
	}
}

func (t *Text) Draw() {
	if t.Disabled {
		return
	}

	rect := t.Transform.GetRect()

	var attribs []*AttribPair
	if t.attribs != nil && len(t.attribs) > 0 {
		attribs = t.attribs
	} else {
		attribs = []*AttribPair{&t.Style}
	}

	// The actual drawing happens here
	x := 0
	y := 0
	i := 0
	skip := t.SkipLines
	height := int(rect.H)
	width := int(rect.W)
	var curAttribs AttribPair

	for _, char := range t.Text {
		if i < len(attribs) {
			newAttribs := attribs[i]
			if newAttribs != nil {
				curAttribs = *newAttribs
			}
		}
		charWidth := runewidth.RuneWidth(char)
		if charWidth == 0 {
			continue
		}
		if char != '\n' {
			if skip <= 0 {
				termbox.SetCell(x+int(rect.X), y+int(rect.Y), char, curAttribs.FG, curAttribs.BG)
			}
			x += charWidth
		} else {
			x = width
		}

		if x >= width {
			skip--
			y++
			x = 0
			if height != 0 && y >= height {
				return
			}
		}
		i++
	}

	// cellSlice := GenCellSlice(t.Text, attribs)
	// SetCells(cellSlice, int(rect.X), int(rect.Y), int(rect.W), int(rect.H))
}

func (t *Text) HeightRequired() int {
	if t.Disabled {
		return 0
	}

	rect := t.Transform.GetRect()
	return HeightRequired(t.Text, int(rect.W))
}

// Implement LayoutElement
func (t *Text) GetRequiredSize() common.Vector2F {
	//rect := t.Transform.GetRect()
	return common.NewVector2I(runewidth.StringWidth(t.Text), t.HeightRequired())
}

func (t *Text) IsLayoutDynamic() bool {
	return false
}

func (t *Text) Destroy() { t.DestroyChildren() }
