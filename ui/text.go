package ui

import (
	"github.com/jonas747/discorder/common"
	"github.com/nsf/termbox-go"
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

	Attribs map[int]AttribPair
	BG, FG  termbox.Attribute // Not used if Attribs is specified

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

func (t *Text) Draw() {
	if t.Disabled {
		return
	}

	rect := t.Transform.GetRect()

	var attribs map[int]AttribPair
	if t.Attribs != nil && len(t.Attribs) > 0 {
		attribs = t.Attribs
	} else {
		attribs = map[int]AttribPair{
			0: AttribPair{t.FG, t.BG},
		}
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
		newAttribs, ok := attribs[i]
		if ok {
			curAttribs = newAttribs
		}

		if char != '\n' {
			if skip <= 0 {
				termbox.SetCell(x+int(rect.X), y+int(rect.Y), char, curAttribs.FG, curAttribs.BG)
			}
			x++
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
	rect := t.Transform.GetRect()
	return common.NewVector2F(rect.W, float32(t.HeightRequired()))
}

func (t *Text) IsLayoutDynamic() bool {
	return false
}

func (t *Text) Destroy() { t.DestroyChildren() }
