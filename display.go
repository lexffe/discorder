package main

// import (
// 	"fmt"
// 	"github.com/bwmarrin/discordgo"
// 	"github.com/nsf/termbox-go"
// 	"log"
// 	"math"
// 	"time"
// 	"unicode/utf8"
// )

// func (app *App) RefreshDisplay() {
// 	// Main display
// 	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

// 	// Always display messages incase of log messages and whatnot
// 	app.DisplayMessages()

// 	headerStr := " Discorder (v" + VERSION + ") (╯°□°） ╯ ︵ ┻━┻"
// 	if app.selectedGuild != nil {
// 		headerStr += " Server: " + app.selectedGuild.Name
// 		if app.selectedChannel != nil {
// 			chName := getChannelName(app.selectedChannel)
// 			headerStr += ", Active Channel: " + chName
// 		} else {
// 			headerStr += ", Ctrl+h to select a channel"
// 		}
// 	} else {
// 		headerStr += " Ctrl+s to select a server "
// 	}
// 	headerStr += " "
// 	DrawHeader(headerStr)
// 	if app.session != nil && app.session.Token != "" {
// 		app.DrawFooter()
// 	}

// 	app.currentState.RefreshDisplay() // state specific stuff

// 	termbox.Flush()
// }

// func (app *App) DisplayMessages() {
// 	sizex, sizey := termbox.Size()

// 	y := sizey - 2
// 	padding := 0

// 	// Iterate through list and print its contents.
// 	for k, item := range app.displayMessages {
// 		var cells []termbox.Cell
// 		if item == nil {
// 			break
// 		}
// 		if k < app.curChatScroll {
// 			continue
// 		}
// 		if item.isLogMessage {
// 			cells = GenCellSlice("Log: "+item.logMessage.content, map[int]AttribPoint{0: AttribPoint{termbox.ColorYellow, termbox.ColorDefault}})
// 		} else {
// 			msg := item.discordMessage
// 			if msg == nil {
// 				continue
// 			}
// 			author := "Unknown?"
// 			if msg.Author != nil {
// 				author = msg.Author.Username
// 			}
// 			ts := item.timestamp.Local().Format(time.Stamp) + " "
// 			tsLen := utf8.RuneCountInString(ts)

// 			authorLen := utf8.RuneCountInString(author)
// 			channel, err := app.session.State.Channel(msg.ChannelID)
// 			if err != nil {
// 				errStr := "(error getting channel" + err.Error() + ") "
// 				fullMsg := ts + errStr + author + ": " + msg.ContentWithMentionsReplaced()
// 				errLen := utf8.RuneCountInString(errStr)
// 				points := map[int]AttribPoint{
// 					0:                          AttribPoint{termbox.ColorBlue, termbox.ColorRed},
// 					tsLen:                      AttribPoint{termbox.ColorWhite, termbox.ColorRed},
// 					errLen + tsLen:             AttribPoint{termbox.ColorCyan | termbox.AttrBold, termbox.ColorDefault},
// 					errLen + authorLen + tsLen: ResetAttribPoint,
// 				}
// 				cells = GenCellSlice(fullMsg, points)
// 			} else {
// 				name := channel.Name
// 				dm := false
// 				if name == "" {
// 					name = "Direct Message"
// 					dm = true
// 				}

// 				fullMsg := ts + "[" + name + "]" + author + ": " + msg.ContentWithMentionsReplaced()
// 				channelLen := utf8.RuneCountInString(name) + 2
// 				points := map[int]AttribPoint{
// 					0:                              AttribPoint{termbox.ColorBlue, termbox.ColorDefault},
// 					tsLen:                          AttribPoint{termbox.ColorGreen, termbox.ColorDefault},
// 					channelLen + tsLen:             AttribPoint{termbox.ColorCyan | termbox.AttrBold, termbox.ColorDefault},
// 					channelLen + authorLen + tsLen: ResetAttribPoint,
// 				}
// 				if dm {
// 					points[tsLen] = AttribPoint{termbox.ColorMagenta, termbox.ColorDefault}
// 				}
// 				cells = GenCellSlice(fullMsg, points)
// 			}
// 		}

// 		lines := HeightRequired(len(cells), sizex-padding*2)
// 		y -= lines
// 		SetCells(cells, padding, y, sizex-1-padding*2, 0)
// 	}
// }

// type AttribPoint struct {
// 	fg termbox.Attribute
// 	bg termbox.Attribute
// }

// var ResetAttribPoint = AttribPoint{termbox.ColorDefault, termbox.ColorDefault}

// func DrawHeader(header string) {
// 	headerLen := utf8.RuneCountInString(header)
// 	runeSlice := []rune(header)
// 	sizeX, _ := termbox.Size()
// 	headerStartPos := (sizeX / 2) - (headerLen / 2)
// 	for i := 0; i < sizeX; i++ {
// 		if i >= headerStartPos && i < headerStartPos+headerLen {
// 			termbox.SetCell(i, 0, runeSlice[i-headerStartPos], termbox.ColorDefault, termbox.ColorDefault)
// 		} else {
// 			termbox.SetCell(i, 0, '=', termbox.ColorDefault, termbox.ColorDefault)
// 		}
// 	}
// }

// func (app *App) DrawFooter() {
// 	sizeX, sizeY := termbox.Size()

// 	typing := app.typingManager.GetTyping(app.listeningChannels)
// 	if len(typing) > 0 {

// 		typingStr := " is typing..."
// 		for _, v := range typing {
// 			member, err := app.session.State.Member(app.selectedServerId, v.UserID)
// 			if err != nil {
// 				log.Println("Error getting member in drawfooter", err)
// 				continue
// 			}

// 			if member.User.Username != "" {
// 				typingStr = ", " + member.User.Username + typingStr
// 			}
// 		}
// 		typingStr = typingStr[2:]
// 		SimpleSetText(0, sizeY-2, sizeX, typingStr, termbox.ColorCyan, termbox.ColorDefault)
// 	}

// 	if app.curChatScroll > 0 {
// 		SimpleSetText(0, sizeY-2, sizeX, fmt.Sprintf("Scroll: %d", app.curChatScroll), termbox.ColorDefault, termbox.ColorYellow)
// 	}
// }

// func DrawPrompt(pre string, x, y, width int, cursor int, buffer string, fg, bg termbox.Attribute) {
// 	preStrLen := utf8.RuneCountInString(pre)
// 	cells := GenCellSlice(pre+buffer, map[int]AttribPoint{
// 		0:         AttribPoint{termbox.ColorYellow | termbox.AttrBold, termbox.ColorDefault},
// 		preStrLen: AttribPoint{termbox.AttrBold | fg, bg},
// 	})

// 	SetCells(cells, x, y, width, 1)
// 	termbox.SetCursor(x+preStrLen+cursor, y)
// }

// func DrawWindow(title, footer string, startX, startY, width, height int, bg termbox.Attribute) {
// 	headerLen := utf8.RuneCountInString(title)
// 	runeSlice := []rune(title)
// 	headerStartPos := (width / 2) - (headerLen / 2)

// 	footerLen := utf8.RuneCountInString(footer)
// 	footerSlice := []rune(footer)
// 	footerStartPos := (width / 2) - (footerLen / 2)

// 	_, tSizeY := termbox.Size()

// 	for curX := 0; curX <= width; curX++ {
// 		for curY := 0; curY <= height; curY++ {
// 			realX := curX + startX
// 			realY := curY + startY

// 			char := ' '

// 			atTop := curY == 0 || realY == 0
// 			atBottom := curY == height || realY == tSizeY-1

// 			if curX >= headerStartPos && curX < headerStartPos+headerLen && atTop {
// 				char = runeSlice[curX-headerStartPos]
// 			} else if curX >= footerStartPos && curX < footerStartPos+footerLen && atBottom {
// 				char = footerSlice[curX-footerStartPos]
// 			} else if curX == 0 || curX == width {
// 				char = '|'
// 			} else if atTop || atBottom {
// 				char = '-'
// 			}
// 			termbox.SetCell(realX, realY, char, termbox.ColorMagenta, bg)
// 		}
// 	}
// }

// func CreateListWindow(title, footer string, list []string, cursor int, selections []int) {
// 	sizeX, sizeY := termbox.Size()
// 	windowWidth := int(float64(sizeX) / 1.5)

// 	height := 2
// 	for _, v := range list {
// 		height += HeightRequired(utf8.RuneCountInString(v), windowWidth)
// 	}

// 	startX := sizeX/2 - windowWidth/2
// 	startY := sizeY/2 - height/2

// 	DrawWindow(title, footer, startX, startY, windowWidth, height, termbox.ColorBlack)

// 	y := startY + 1

// 	if height > sizeY {
// 		// If window is taller then scroll
// 		added := int(float64(height)*(float64(len(list)-cursor)/float64(len(list)))) - height/2
// 		y += added
// 	}

// 	for k, v := range list {
// 		bg := termbox.ColorBlack
// 		for _, selected := range selections {
// 			if selected == k {
// 				bg = termbox.ColorYellow
// 				break
// 			}
// 		}
// 		if k == cursor {
// 			if bg == termbox.ColorYellow {
// 				// Both cursor and selected! What a world we live in!
// 				bg = termbox.ColorBlue
// 			} else {
// 				bg = termbox.ColorCyan
// 			}
// 		}
// 		cells := GenCellSlice(v, map[int]AttribPoint{0: AttribPoint{termbox.ColorDefault, bg}})
// 		mod := SetCells(cells, startX+1, y, windowWidth, 0)
// 		y += mod
// 	}
// }

// func getChannelName(channel *discordgo.Channel) string {
// 	if channel.IsPrivate {
// 		return channel.Recipient.Username
// 	}
// 	return channel.Name
// }
