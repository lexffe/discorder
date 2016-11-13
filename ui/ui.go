package ui

import (
	"github.com/jonas747/termbox-go"
	"unicode/utf8"
)

// Simple bg fg attribute pair
type AttribPair struct {
	FG termbox.Attribute `json:"fg"`
	BG termbox.Attribute `json:"bg"`
}

// Helper functions

// A Helper for drawing simple text, returns number of lines
func SimpleSetText(startX, startY, width int, str string, fg, bg termbox.Attribute) int {
	cells := GenCellSlice(str, map[int]AttribPair{0: {fg, bg}})
	return SetCells(cells, startX, startY, width, 0)
}

// Generates a cellslice with attributes
func GenCellSlice(str string, points map[int]AttribPair) []termbox.Cell {
	index := 0
	curAttribs := AttribPair{}
	cells := make([]termbox.Cell, utf8.RuneCountInString(str))
	for _, ch := range str {
		newAttribs, ok := points[index]
		if ok {
			curAttribs = newAttribs
		}
		cell := termbox.Cell{
			Ch: ch,
			Fg: curAttribs.FG,
			Bg: curAttribs.BG,
		}
		cells[index] = cell
		index++
	}
	return cells
}

// Sets the cells and returns number of lines
func SetCells(cells []termbox.Cell, startX, startY, width, height int) int {
	x := 0
	y := 0

	for _, cell := range cells {
		if cell.Ch != '\n' {
			termbox.SetCell(x+startX, y+startY, cell.Ch, cell.Fg, cell.Bg)
			x++
		} else {
			x = width
		}

		if x >= width {
			y++
			x = 0
			if height != 0 && y >= height {
				return y
			}
		}
	}
	return y + 1
}

// Returns the number of lines required
func HeightRequired(str string, width int) int {
	if str == "" {
		return 0
	}

	lines := BuildTextLines(str, width)
	return (len(lines))
}
