package ui

import (
	"github.com/jonas747/discorder/common"
	"github.com/jonas747/termbox-go"
	"strconv"
	"unicode/utf8"
)

type DataType int

const (
	DataTypeString DataType = iota
	DataTypePassword
	DataTypeInt
	DataTypeFloat
	DataTypeBool
)

type TextInput struct {
	*BaseEntity
	Text *Text

	TextBuffer          string
	CursorLocation      int
	Active              bool
	MaskInput           bool // Replecas everything with "*"
	HideCursorWhenEmpty bool
	Layer               int

	DataType  DataType
	MinHeight int

	Manager *Manager
}

func NewTextInput(manager *Manager, layer int) *TextInput {
	input := &TextInput{
		BaseEntity: &BaseEntity{},
		Text:       NewText(),
		Manager:    manager,
		Layer:      layer,
	}

	input.Transform.AddChildren(input.Text)
	input.Text.Transform.AnchorMax = common.NewVector2I(1, 1)
	input.Text.Layer = layer

	manager.AddInput(input, false)

	return input
}

func (ti *TextInput) HandleInput(event termbox.Event) {
	if event.Type != termbox.EventKey || !ti.Active {
		return
	}

	char := event.Ch
	if event.Key == termbox.KeySpace {
		char = ' '
	} else if event.Key == termbox.Key(0) && event.Mod == termbox.ModAlt && char == 0 {
		char = '@' // Just temporary workaround for non american keyboards on windows
		// So they're atleast able to log in
	}
	if char == 0 {
		return
	}

	switch ti.DataType {
	case DataTypeInt:
		_, err := strconv.Atoi(string(char))
		if err != nil {
			return
		}
	case DataTypeBool:
		if char == 't' || char == 'T' {
			ti.TextBuffer = "true"
		} else {
			ti.TextBuffer = "false"
		}
		return
	case DataTypeFloat:
	}

	bufLen := utf8.RuneCountInString(ti.TextBuffer)
	if ti.CursorLocation == bufLen {
		ti.TextBuffer += string(char)
		ti.CursorLocation++
	} else if ti.CursorLocation == 0 {
		ti.TextBuffer = string(char) + ti.TextBuffer
		ti.CursorLocation++
	} else {
		bufSlice := []rune(ti.TextBuffer)
		bufCopy := ""

		for i := 0; i < len(bufSlice); i++ {
			if i == ti.CursorLocation {
				bufCopy += string(char)
			}
			bufCopy += string(bufSlice[i])
		}
		ti.TextBuffer = bufCopy
		ti.CursorLocation++
	}

}

type Direction int

const (
	DirLeft Direction = iota
	DirRight
	DirUp
	DirDown
	DirEnd
	DirStart
)

func (ti *TextInput) MoveCursor(dir Direction, amount int, byWords bool) {
	switch dir {
	case DirLeft:
		ti.CursorLocation -= amount
		if ti.CursorLocation < 0 {
			ti.CursorLocation = 0
		}
	case DirRight:
		ti.CursorLocation += amount
		bufLen := utf8.RuneCountInString(ti.TextBuffer)
		if ti.CursorLocation > bufLen {
			ti.CursorLocation = bufLen
		}
	case DirUp:
	case DirDown:
	case DirEnd:
		ti.CursorLocation = utf8.RuneCountInString(ti.TextBuffer)
	case DirStart:
		ti.CursorLocation = 0
	}
}

func (ti *TextInput) Erase(dir Direction, amount int, byWords bool) {
	bufLen := utf8.RuneCountInString(ti.TextBuffer)

	if byWords {

		switch dir {
		case DirLeft:
			spaces := make([]int, 1)

			i := 0
			for _, v := range ti.TextBuffer {
				if v == ' ' || v == '_' || v == '-' {
					spaces = append(spaces, i)
				}

				if i >= ti.CursorLocation {
					break
				}
				i++
			}

			if amount >= len(spaces) {
				amount = ti.CursorLocation
			} else {
				amount = ti.CursorLocation - spaces[len(spaces)-amount]
			}

		}

	}

	switch dir {
	case DirLeft:
		if bufLen == 0 || ti.CursorLocation == 0 {
			return
		}

		runeSlice := []rune(ti.TextBuffer)
		runeSlice = append(runeSlice[:ti.CursorLocation-amount], runeSlice[ti.CursorLocation:]...)
		ti.TextBuffer = string(runeSlice)
		ti.CursorLocation -= amount
		if ti.CursorLocation < 0 {
			ti.CursorLocation = 0
		}

	case DirRight:
		if ti.CursorLocation >= bufLen {
			break
		}
		runeSlice := []rune(ti.TextBuffer)
		runeSlice = append(runeSlice[:ti.CursorLocation], runeSlice[ti.CursorLocation+amount:]...)
		ti.TextBuffer = string(runeSlice)
	case DirEnd:
	case DirStart:
	}
}

func (ti *TextInput) Update() {
	if ti.MaskInput || ti.DataType == DataTypePassword {
		ti.Text.Text = ""
		for range ti.TextBuffer {
			ti.Text.Text += "*"
		}
	} else {
		ti.Text.Text = ti.TextBuffer
	}

}

func (ti *TextInput) Draw() {
	if !ti.Active {
		return
	}
	if ti.TextBuffer != "" || (ti.TextBuffer == "" && !ti.HideCursorWhenEmpty) {
		rect := ti.Transform.GetRect()
		if rect.W < 1 {
			return
		}
		x := (ti.CursorLocation % int(rect.W)) + int(rect.X)
		y := int(rect.Y) + (ti.CursorLocation / int(rect.W))
		SafeSetCursor(x, y)
	} else if ti.HideCursorWhenEmpty {
		termbox.HideCursor()
	}
}

func (ti *TextInput) GetDrawLayer() int {
	return ti.Layer
}

func (ti *TextInput) Destroy() {
	ti.Manager.RemoveInput(ti, true)
	ti.DestroyChildren()
}

func (ti *TextInput) SetActive(active bool) {
	if active {
		ti.Manager.SetActiveInput(ti)
	} else {
		ti.Active = false
	}
}

// Implement LayoutElement
func (ti *TextInput) GetRequiredSize() common.Vector2F {
	rect := ti.Transform.GetRect()
	height := ti.Text.HeightRequired()
	if height < ti.MinHeight {
		height = ti.MinHeight
	}
	return common.NewVector2F(rect.W, float32(height))
}

func (ti *TextInput) IsLayoutDynamic() bool {
	return false
}

// Crashes on windows otherwise if going out of bounds
func SafeSetCursor(x, y int) {
	sizeX, sizeY := termbox.Size()
	if x > +sizeX {
		x = sizeX - 1
	}
	if y >= sizeY {
		y = sizeY - 1
	}
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}

	termbox.SetCursor(x, y)
}
