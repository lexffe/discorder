package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/nsf/termbox-go"
	"log"
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

	app.DisplayMessages()

	headerStr := " Discorder (v" + VERSION + ") (╯°□°） ╯ ︵ ┻━┻"
	if app.selectedGuild != nil {
		headerStr += " Server: " + app.selectedGuild.Name
		if app.selectedChannel != nil {
			headerStr += ", Active Channel: " + app.selectedChannel.Name
		} else {
			headerStr += ", Ctrl+h to select a channel"
		}
	} else {
		headerStr += " Ctrl+s to select a server "
	}
	headerStr += " "
	CreateHeader(headerStr)

	app.CreateFooter()

	app.currentState.RefreshDisplay() // state specific stuff

	termbox.Flush()
}

func (app *App) DisplayMessages() {
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
				name := channel.Name
				dm := false
				if name == "" {
					name = "Direct Message"
					dm = true
				}

				fullMsg := "[" + channel.Name + "]" + msg.Author.Username + ": " + msg.ContentWithMentionsReplaced()
				channelLen := utf8.RuneCountInString(channel.Name) + 2
				points := map[int]AttribPoint{
					0:                      AttribPoint{termbox.ColorGreen, termbox.ColorDefault},
					channelLen:             AttribPoint{termbox.ColorCyan | termbox.AttrBold, termbox.ColorDefault},
					channelLen + authorLen: ResetAttribPoint,
				}
				if dm {
					points[0] = AttribPoint{termbox.ColorMagenta, termbox.ColorDefault}
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

func (app *App) CreateFooter() {
	preStr := "Send To " + app.selectedChannelId + ":"
	if app.selectedChannel != nil {
		preStr = "Send to #" + app.selectedChannel.Name + ":"
	}
	preStrLen := utf8.RuneCountInString(preStr)

	body := app.currentSendBuffer + " "
	//bodyLen := utf8.RuneCountInString(body)

	pointTyped := AttribPoint{termbox.ColorDefault, termbox.ColorDefault}

	cells := GenCellSlice(preStr+body, map[int]AttribPoint{
		0:         AttribPoint{termbox.AttrBold | termbox.ColorYellow, termbox.ColorDefault},
		preStrLen: pointTyped,
		//preStrLen + app.currentCursorLocation:     AttribPoint{termbox.ColorDefault, termbox.ColorYellow},
		//preStrLen + app.currentCursorLocation + 1: pointTyped,
	})

	sizeX, sizeY := termbox.Size()
	app.SetCells(cells, 0, sizeY-1, sizeX, 1)
	termbox.SetCursor(preStrLen+app.currentCursorLocation, sizeY-1)
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

func (app *App) CreateServerWindow(selected int) {
	state := app.session.State
	state.RLock()
	defer state.RUnlock()

	strList := make([]string, len(state.Guilds))
	for k, v := range state.Guilds {
		strList[k] = v.Name
	}
	app.CreateListWindow("Servers", strList, selected)
}

func (app *App) CreateChannelWindow(selected int) {
	g, err := app.session.State.Guild(app.selectedServerId)
	if err != nil {
		log.Println("Error getting guild ", err.Error())
		return
	}
	strList := make([]string, 0)
	for _, v := range g.Channels {
		if v.Type == "text" {
			strList = append(strList, "#"+v.Name)
		}
	}
	app.CreateListWindow("Channels", strList, selected)
}

// Need a scrollbar
func (app *App) CreateListWindow(title string, list []string, selected int) {
	sizeX, sizeY := termbox.Size()
	windowWidth := int(float64(sizeX) / 1.5)

	height := 2
	for _, v := range list {
		height += HeightRequired(utf8.RuneCountInString(v), windowWidth)
	}

	startX := sizeX/2 - windowWidth/2
	startY := sizeY/2 - height/2

	CreateWindow(title, startX, startY, windowWidth, height, termbox.ColorBlack)

	y := startY + 1

	for k, v := range list {
		bg := termbox.ColorBlack
		if k == selected {
			bg = termbox.ColorYellow
		}
		cells := GenCellSlice(v, map[int]AttribPoint{0: AttribPoint{termbox.ColorDefault, bg}})
		mod := app.SetCells(cells, startX+1, y, windowWidth, 0)
		y += mod
	}
}
