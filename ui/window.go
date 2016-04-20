package ui

import (
	"github.com/nsf/termbox-go"
	"unicode/utf8"
)

type Window struct {
	*BaseEntity

	Title  string
	Footer string

	Layer int

	Transform          *Transform
	BorderBG, BorderFG termbox.Attribute
	FillBG             termbox.Attribute
}

func NewWindow() *Window {
	w := &Window{
		BaseEntity: &BaseEntity{},
		Transform:  &Transform{},
		BorderBG:   termbox.ColorBlack | termbox.AttrBold,
		BorderFG:   termbox.ColorWhite,
		FillBG:     termbox.ColorBlack,
	}
	w.Self = w // See BaseEntity struct for why
	return w
}

func (w *Window) GetDrawLayer() int {
	return w.Layer
}

func (w *Window) Draw() {
	headerLen := utf8.RuneCountInString(w.Title)
	runeSlice := []rune(w.Title)

	rect := w.Transform.GetRect()

	headerStartPos := int((rect.W / 2) - (float32(headerLen) / 2))

	footerLen := utf8.RuneCountInString(w.Footer)
	footerSlice := []rune(w.Footer)
	footerStartPos := int((rect.W / 2) - (float32(footerLen) / 2))

	_, tSizeY := termbox.Size()

	for curX := -1; curX <= int(rect.W)+1; curX++ {
		for curY := -1; curY <= int(rect.H)+1; curY++ {
			realX := curX + int(rect.X)
			realY := curY + int(rect.Y)

			char := ' '

			atTop := curY == -1 || realY == 0
			atBottom := curY == int(rect.H)+1 || realY == tSizeY-1

			var fg, bg termbox.Attribute
			atBorder := false
			if curX >= headerStartPos && curX < headerStartPos+headerLen && atTop {
				char = runeSlice[curX-headerStartPos]
			} else if curX >= footerStartPos && curX < footerStartPos+footerLen && atBottom {
				char = footerSlice[curX-footerStartPos]
			} else if curX == -1 || curX == int(rect.W)+1 {
				char = '|'
				atBorder = true
			} else if atTop || atBottom {
				char = '-'
			}

			if atBorder || atTop || atBottom {
				fg = w.BorderFG
				bg = w.BorderBG
			} else {
				bg = w.FillBG
			}

			termbox.SetCell(realX, realY, char, fg, bg)
		}
	}
}

func (w *Window) Destroy() { w.DestroyChildren() }
func (w *Window) Init()    {}
