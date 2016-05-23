package ui

import (
	"github.com/jonas747/discorder/common"
	"github.com/nsf/termbox-go"
	"unicode/utf8"
)

type TextInput struct {
	*BaseEntity
	Text *Text

	Layer          int
	TextBuffer     string
	CursorLocation int
	Active         bool
	MaskInput      bool // Replecas everything with "*"
}

func NewTextInput() *TextInput {
	input := &TextInput{
		BaseEntity: &BaseEntity{},
		Text:       NewText(),
	}

	input.Transform.AddChildren(input.Text)
	input.Text.Transform.AnchorMax = common.NewVector2I(1, 1)

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

	switch dir {
	case DirLeft:
		if bufLen == 0 {
			return
		}
		if ti.CursorLocation == bufLen {
			_, size := utf8.DecodeLastRuneInString(ti.TextBuffer)
			ti.CursorLocation--
			ti.TextBuffer = ti.TextBuffer[:len(ti.TextBuffer)-size]
		} else if ti.CursorLocation == 1 {
			_, size := utf8.DecodeRuneInString(ti.TextBuffer)
			ti.CursorLocation--
			ti.TextBuffer = ti.TextBuffer[size:]
		} else if ti.CursorLocation == 0 {
			return
		} else {
			runeSlice := []rune(ti.TextBuffer)
			newSlice := append(runeSlice[:ti.CursorLocation-1], runeSlice[ti.CursorLocation:]...)
			ti.TextBuffer = string(newSlice)
			ti.CursorLocation--
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
	if ti.MaskInput {
		ti.Text.Text = ""
		for _, _ = range ti.TextBuffer {
			ti.Text.Text += "*"
		}
	} else {
		ti.Text.Text = ti.TextBuffer
	}
	ti.Text.Layer = ti.Layer
}

func (ti *TextInput) Draw() {
	if ti.Active {
		rect := ti.Transform.GetRect()
		if rect.W < 1 {
			return
		}
		x := (ti.CursorLocation % int(rect.W)) + int(rect.X)
		y := int(rect.Y) + (ti.CursorLocation / int(rect.W))
		SafeSetCursor(x, y)
	}
}

func (ti *TextInput) GetDrawLayer() int {
	return ti.Layer
}

func (ti *TextInput) Destroy() { ti.DestroyChildren() }

// Implement LayoutElement
func (ti *TextInput) GetRequiredSize() common.Vector2F {
	rect := ti.Transform.GetRect()
	return common.NewVector2F(rect.W, float32(ti.Text.HeightRequired()))
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
