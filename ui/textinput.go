package ui

import (
	"github.com/jonas747/discorder/common"
	"github.com/nsf/termbox-go"
	"unicode/utf8"
)

type TextInput struct {
	*BaseEntity
	Transform *Transform
	Text      *Text

	Layer          int
	TextBuffer     string
	CursorLocation int
	Active         bool
	MaskInput      bool // Replecas everything with "*"
}

func NewTextInput() *TextInput {
	t := NewText()

	input := &TextInput{
		BaseEntity: &BaseEntity{},
		Transform:  &Transform{},
	}

	t.Transform.Parent = input.Transform
	t.Transform.AnchorMax = common.NewVector2I(1, 1)
	input.Text = t
	input.AddChild(t)
	return input
}

func (ti *TextInput) HandleInput(event termbox.Event) {
	if event.Type != termbox.EventKey || !ti.Active {
		return
	}

	switch event.Key {
	case termbox.KeyArrowLeft: // Move cursor left
		ti.CursorLocation--
		if ti.CursorLocation < 0 {
			ti.CursorLocation = 0
		}
	case termbox.KeyArrowRight: // Move cusror right
		ti.CursorLocation++
		bufLen := utf8.RuneCountInString(ti.TextBuffer)
		if ti.CursorLocation > bufLen {
			ti.CursorLocation = bufLen
		}
	case termbox.KeyBackspace, termbox.KeyBackspace2:
		bufLen := utf8.RuneCountInString(ti.TextBuffer)
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
	case termbox.KeyDelete:
		bufLen := utf8.RuneCountInString(ti.TextBuffer)
		if ti.CursorLocation >= bufLen {
			break
		}
		runeSlice := []rune(ti.TextBuffer)
		runeSlice = append(runeSlice[:ti.CursorLocation], runeSlice[ti.CursorLocation+1:]...)
		ti.TextBuffer = string(runeSlice)
	default:
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
}

func (ti *TextInput) Draw() {
	if ti.Active {
		rect := ti.Transform.GetRect()
		termbox.SetCursor(ti.CursorLocation+int(rect.X), int(rect.Y))
	}
}

func (ti *TextInput) GetDrawLayer() int {
	return ti.Layer
}

func (ti *TextInput) Destroy() { ti.DestroyChildren() }
