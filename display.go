package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/nsf/termbox-go"
	"math"
	"time"
	"unicode/utf8"
)

func (app *App) RefreshDisplay() {
	// Main display
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	// Always display messages incase of log messages and whatnot
	app.DisplayMessages()

	headerStr := " Discorder (v" + VERSION + ") (╯°□°） ╯ ︵ ┻━┻"
	if app.selectedGuild != nil {
		headerStr += " Server: " + app.selectedGuild.Name
		if app.selectedChannel != nil {
			chName := getChannelName(app.selectedChannel)
			headerStr += ", Active Channel: " + chName
		} else {
			headerStr += ", Ctrl+h to select a channel"
		}
	} else {
		headerStr += " Ctrl+s to select a server "
	}
	headerStr += " "
	CreateHeader(headerStr)
	if app.session != nil && app.session.Token != "" {
		app.CreateFooter()
	}

	app.currentState.RefreshDisplay() // state specific stuff

	termbox.Flush()
}

func (app *App) DisplayMessages() {
	sizex, sizey := termbox.Size()

	y := sizey - 2
	padding := 2

	// Iterate through list and print its contents.
	for k, item := range app.displayMessages {
		var cells []termbox.Cell
		if item == nil {
			break
		}
		if k < app.curChatScroll {
			continue
		}
		if item.isLogMessage {
			cells = GenCellSlice("Log: "+item.logMessage.content, map[int]AttribPoint{0: AttribPoint{termbox.ColorYellow, termbox.ColorDefault}})
		} else {
			msg := item.discordMessage
			if msg == nil {
				continue
			}
			author := "Unknown?"
			if msg.Author != nil {
				author = msg.Author.Username
			}
			ts := item.timestamp.Local().Format(time.Stamp) + " "
			tsLen := utf8.RuneCountInString(ts)

			authorLen := utf8.RuneCountInString(author)
			channel, err := app.session.State.Channel(msg.ChannelID)
			if err != nil {
				errStr := "(error getting channel" + err.Error() + ") "
				fullMsg := ts + errStr + author + ": " + msg.ContentWithMentionsReplaced()
				errLen := utf8.RuneCountInString(errStr)
				points := map[int]AttribPoint{
					0:                          AttribPoint{termbox.ColorBlue, termbox.ColorRed},
					tsLen:                      AttribPoint{termbox.ColorWhite, termbox.ColorRed},
					errLen + tsLen:             AttribPoint{termbox.ColorCyan | termbox.AttrBold, termbox.ColorDefault},
					errLen + authorLen + tsLen: ResetAttribPoint,
				}
				cells = GenCellSlice(fullMsg, points)
			} else {
				name := channel.Name
				dm := false
				if name == "" {
					name = "Direct Message"
					dm = true
				}

				fullMsg := ts + "[" + name + "]" + author + ": " + msg.ContentWithMentionsReplaced()
				channelLen := utf8.RuneCountInString(name) + 2
				points := map[int]AttribPoint{
					0:                              AttribPoint{termbox.ColorBlue, termbox.ColorDefault},
					tsLen:                          AttribPoint{termbox.ColorGreen, termbox.ColorDefault},
					channelLen + tsLen:             AttribPoint{termbox.ColorCyan | termbox.AttrBold, termbox.ColorDefault},
					channelLen + authorLen + tsLen: ResetAttribPoint,
				}
				if dm {
					points[1] = AttribPoint{termbox.ColorMagenta, termbox.ColorDefault}
				}
				cells = GenCellSlice(fullMsg, points)
			}
		}

		lines := HeightRequired(len(cells), sizex-padding*2)
		y -= lines
		SetCells(cells, padding, y, sizex-padding*2, 0)
	}
}

type AttribPoint struct {
	fg termbox.Attribute
	bg termbox.Attribute
}

var ResetAttribPoint = AttribPoint{termbox.ColorDefault, termbox.ColorDefault}

func SimpleSetText(startX, startY, width int, str string, fg, bg termbox.Attribute) int {
	cells := GenCellSlice(str, map[int]AttribPoint{0: AttribPoint{fg, bg}})
	return SetCells(cells, startX, startY, width, 0)
}

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
func SetCells(cells []termbox.Cell, startX, startY, width, height int) int {
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
	sizeX, sizeY := termbox.Size()

	if app.curChatScroll > 0 {
		SimpleSetText(0, sizeY-2, sizeX, fmt.Sprintf("Scroll: %d", app.curChatScroll), termbox.ColorDefault, termbox.ColorYellow)
	}

	preStr := "Send To " + app.selectedChannelId + ":"
	if app.selectedChannel != nil {
		preStr = "Send to #" + getChannelName(app.selectedChannel) + ":"
	}
	preStrLen := utf8.RuneCountInString(preStr)

	body := app.currentTextBuffer + " "
	//bodyLen := utf8.RuneCountInString(body)

	pointTyped := AttribPoint{termbox.ColorDefault, termbox.ColorDefault}

	cells := GenCellSlice(preStr+body, map[int]AttribPoint{
		0:         AttribPoint{termbox.AttrBold | termbox.ColorYellow, termbox.ColorDefault},
		preStrLen: pointTyped,
	})

	SetCells(cells, 0, sizeY-1, sizeX, 1)
	termbox.SetCursor(preStrLen+app.currentCursorLocation, sizeY-1)
}

func (app *App) Prompt(x, y, width int, cursor int, buffer string) {
	///body := app.currentSendBuffer + " "
	//bodyLen := utf8.RuneCountInString(body)

	cells := GenCellSlice(buffer, map[int]AttribPoint{
		0: AttribPoint{termbox.AttrBold | termbox.ColorYellow, termbox.ColorDefault},
	})

	//sizeX, sizeY := termbox.Size()
	SetCells(cells, x, y, width, 1)
	termbox.SetCursor(x+cursor, y)
}

func CreateWindow(title string, startX, startY, width, height int, color termbox.Attribute) {
	headerLen := utf8.RuneCountInString(title)
	runeSlice := []rune(title)
	headerStartPos := (width / 2) - (headerLen / 2)

	for curX := 0; curX <= width; curX++ {
		for curY := 0; curY <= height; curY++ {
			realX := curX + startX
			realY := curY + startY

			char := ' '
			if curX >= headerStartPos && curX < headerStartPos+headerLen && curY == 0 {
				char = runeSlice[curX-headerStartPos]
			} else if curX == 0 || curX == width {
				char = '|'
			} else if curY == 0 || curY == height {
				char = '-'
			}
			termbox.SetCell(realX, realY, char, termbox.ColorDefault, color)
		}
	}
}

func (app *App) CreateListWindow(title string, list []string, cursor int, selections []int) {
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

	if height > sizeY {
		// If window is taller then scroll
		added := int(float64(height)*(float64(len(list)-cursor)/float64(len(list)))) - height/2
		y += added
	}

	for k, v := range list {
		bg := termbox.ColorBlack
		for _, selected := range selections {
			if selected == k {
				bg = termbox.ColorYellow
				break
			}
		}
		if k == cursor {
			if bg == termbox.ColorYellow {
				// Both cursor and selected! What a world we live in!
				bg = termbox.ColorBlue
			} else {
				bg = termbox.ColorCyan
			}
		}
		cells := GenCellSlice(v, map[int]AttribPoint{0: AttribPoint{termbox.ColorDefault, bg}})
		mod := SetCells(cells, startX+1, y, windowWidth, 0)
		y += mod
	}
}

func getChannelName(channel *discordgo.Channel) string {
	if channel.IsPrivate {
		return channel.Recipient.Username
	}
	return channel.Name
}
