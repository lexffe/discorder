package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/nsf/termbox-go"
	"math"
	"unicode/utf8"
)

func (app *App) RefreshDisplay() {
	if !termbox.IsInit {
		err := termbox.Init()
		if err != nil {
			panic(err)
		}
	}
	// Main display
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	sizex, sizey := termbox.Size()

	y := sizey - 2
	padding := 2

	// Iterate through list and print its contents.
	for e := app.history.Front(); e != nil; e = e.Next() {
		var cells []termbox.Cell
		switch msg := e.Value.(type) {
		case discordgo.Message:

			authorLen := utf8.RuneCountInString(msg.Author.Username)

			channel, err := app.session.State.Channel(msg.ChannelID)
			if err != nil {
				errStr := "(error getting channel" + err.Error() + ") "
				fullMsg := errStr + msg.Author.Username + ": " + msg.ContentWithMentionsReplaced()
				errLen := utf8.RuneCountInString(errStr)
				points := map[int]AttribPoint{
					0:                  AttribPoint{termbox.ColorWhite, termbox.ColorRed},
					errLen:             AttribPoint{termbox.ColorCyan | termbox.AttrBold, termbox.ColorDefault},
					errLen + authorLen: ResetAttribPoint,
				}
				cells = GenCellSlice(fullMsg, points)
			} else {
				fullMsg := "[" + channel.Name + "]" + msg.Author.Username + ": " + msg.ContentWithMentionsReplaced()
				channelLen := utf8.RuneCountInString(channel.Name) + 2
				points := map[int]AttribPoint{
					0:                      AttribPoint{termbox.ColorGreen, termbox.ColorDefault},
					channelLen:             AttribPoint{termbox.ColorCyan | termbox.AttrBold, termbox.ColorDefault},
					channelLen + authorLen: ResetAttribPoint,
				}
				cells = GenCellSlice(fullMsg, points)
			}

		case string:
			cells = GenCellSlice("Log: "+msg, map[int]AttribPoint{0: AttribPoint{termbox.ColorYellow, termbox.ColorDefault}})
		}

		lines := HeightRequired(len(cells), sizex-padding*2)
		y -= lines
		app.SetCells(cells, padding, y, sizex-padding*2, 0)
	}

	CreateHeader(" Discorder DEV (╯°□°）╯︵ ┻━┻ ")
	//CreateWindow("A window!", sizex/2-30, sizey/2-10, 60, 20, termbox.ColorBlack)
	//app.CreateServerWindow(0)

	app.currentState.RefreshDisplay()

	termbox.Flush()
}

type AttribPoint struct {
	fg termbox.Attribute
	bg termbox.Attribute
}

var ResetAttribPoint = AttribPoint{termbox.ColorDefault, termbox.ColorDefault}

func GenCellSlice(str string, points map[int]AttribPoint) []termbox.Cell {
	index := 0
	curAttribs := ResetAttribPoint
	cells := make([]termbox.Cell, utf8.RuneCountInString(str))
	for _, ch := range str {
		newAttribs, ok := points[index]
		if ok {
			curAttribs = newAttribs
		}
		cell := termbox.Cell{
			Ch: ch,
			Fg: curAttribs.fg,
			Bg: curAttribs.bg,
		}
		cells[index] = cell
		index++
	}
	return cells
}

// Returns number of lines
func (app *App) SetCells(cells []termbox.Cell, startX, startY, width, height int) int {
	x := 0
	y := 0

	for _, cell := range cells {
		termbox.SetCell(x+startX, y+startY, cell.Ch, cell.Fg, cell.Bg)

		x++
		if x > width {
			y++
			x = 0
			if height != 0 && y >= height {
				return y
			}
		}
	}
	return y + 1
}

func HeightRequired(num, width int) int {
	return int(math.Ceil(float64(num) / float64(width)))
}

func CreateHeader(header string) {
	headerLen := utf8.RuneCountInString(header)
	runeSlice := []rune(header)
	sizeX, _ := termbox.Size()
	headerStartPos := (sizeX / 2) - (headerLen / 2)
	for i := 0; i < sizeX; i++ {
		if i >= headerStartPos && i < headerStartPos+headerLen {
			termbox.SetCell(i, 0, runeSlice[i-headerStartPos], termbox.ColorDefault, termbox.ColorDefault)
		} else {
			termbox.SetCell(i, 0, '=', termbox.ColorDefault, termbox.ColorDefault)
		}
	}
}

func CreateWindow(title string, startX, startY, width, height int, color termbox.Attribute) {
	headerLen := utf8.RuneCountInString(title)
	runeSlice := []rune(title)
	headerStartPos := (width / 2) - (headerLen / 2)
	for curX := 0; curX < width; curX++ {
		for curY := 0; curY < height; curY++ {
			realX := curX + startX
			realY := curY + startY

			char := ' '
			if curX >= headerStartPos && curX < headerStartPos+headerLen && curY == 0 {
				char = runeSlice[curX-headerStartPos]
			} else if curX == 0 || curX == width-1 {
				char = '|'
			} else if curY == 0 || curY == height-1 {
				char = '-'
			}
			termbox.SetCell(realX, realY, char, termbox.ColorDefault, color)
		}
	}

}

// Need a scrollbar
func (app *App) CreateServerWindow(selected int) {
	state := app.session.State
	state.RLock()
	defer state.RUnlock()

	sizeX, sizeY := termbox.Size()

	windowWidth := int(float64(sizeX) / 1.5)

	serverStrings := make([]string, len(state.Guilds))
	height := 2
	for k, v := range state.Guilds {
		serverStrings[k] = v.Name
		height += HeightRequired(utf8.RuneCountInString(v.Name), windowWidth)
	}

	CreateWindow("Servers", (sizeX/2)-(windowWidth/2), sizeY/2-height/2, windowWidth, height, termbox.ColorBlack)
	startY := sizeY/2 - height/2
	y := startY + 1
	x := sizeX/2 - windowWidth/2
	for k, v := range serverStrings {
		bg := termbox.ColorBlack
		if k == selected {
			bg = termbox.ColorYellow
		}
		cells := GenCellSlice(v, map[int]AttribPoint{0: AttribPoint{termbox.ColorDefault, bg}})
		mod := app.SetCells(cells, x+1, y, windowWidth, 0)
		y += mod
	}
}
