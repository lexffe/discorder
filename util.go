package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/discorder/common"
	"log"
	"time"
)

// For logs
func (app *App) Write(p []byte) (n int, err error) {
	cop := string(p)

	// since we might log from the same goroutine deadlocks may occour, should probably do a queue system or something instead...
	go func() {
		app.logChan <- cop
	}()

	if app.logFile != nil {
		app.logFileLock.Lock()
		defer app.logFileLock.Unlock()
		app.logFile.Write(p)
	}

	return len(p), nil
}

// func (app *App) HandleTextInput(event termbox.Event) {
// 	if event.Type == termbox.EventKey {

// 		switch event.Key {

// 		case termbox.KeyArrowLeft:
// 			app.currentCursorLocation--
// 			if app.currentCursorLocation < 0 {
// 				app.currentCursorLocation = 0
// 			}
// 		case termbox.KeyArrowRight:
// 			app.currentCursorLocation++
// 			bufLen := utf8.RuneCountInString(app.currentTextBuffer)
// 			if app.currentCursorLocation > bufLen {
// 				app.currentCursorLocation = bufLen
// 			}
// 		case termbox.KeyBackspace, termbox.KeyBackspace2:
// 			bufLen := utf8.RuneCountInString(app.currentTextBuffer)
// 			if bufLen == 0 {
// 				return
// 			}
// 			if app.currentCursorLocation == bufLen {
// 				_, size := utf8.DecodeLastRuneInString(app.currentTextBuffer)
// 				app.currentCursorLocation--
// 				app.currentTextBuffer = app.currentTextBuffer[:len(app.currentTextBuffer)-size]
// 			} else if app.currentCursorLocation == 1 {
// 				_, size := utf8.DecodeRuneInString(app.currentTextBuffer)
// 				app.currentCursorLocation--
// 				app.currentTextBuffer = app.currentTextBuffer[size:]
// 			} else if app.currentCursorLocation == 0 {
// 				return
// 			} else {
// 				runeSlice := []rune(app.currentTextBuffer)
// 				newSlice := append(runeSlice[:app.currentCursorLocation-1], runeSlice[app.currentCursorLocation:]...)
// 				app.currentTextBuffer = string(newSlice)
// 				app.currentCursorLocation--
// 			}
// 		default:
// 			char := event.Ch
// 			if event.Key == termbox.KeySpace {
// 				char = ' '
// 			} else if event.Key == termbox.Key(0) && event.Mod == termbox.ModAlt && char == 0 {
// 				char = '@' // Just temporary workaround for non american keyboards on windows
// 				// So they're atleast able to log in
// 			}

// 			bufLen := utf8.RuneCountInString(app.currentTextBuffer)
// 			if app.currentCursorLocation == bufLen {
// 				app.currentTextBuffer += string(char)
// 				app.currentCursorLocation++
// 			} else if app.currentCursorLocation == 0 {
// 				app.currentTextBuffer = string(char) + app.currentTextBuffer
// 				app.currentCursorLocation++
// 			} else {
// 				bufSlice := []rune(app.currentTextBuffer)
// 				bufCopy := ""

// 				for i := 0; i < len(bufSlice); i++ {
// 					if i == app.currentCursorLocation {
// 						bufCopy += string(char)
// 					}
// 					bufCopy += string(bufSlice[i])
// 				}
// 				app.currentTextBuffer = bufCopy
// 				app.currentCursorLocation++
// 			}
// 		}

// 	}
// }

func (app *App) GetHistory(channelId string, limit int, beforeId, afterId string) {
	state := app.session.State
	channel, err := state.Channel(channelId)
	if err != nil {
		log.Println("History error: ", err)
		return
	}

	// func (s *Session) ChannelMessages(channelID string, limit int, beforeID, afterID string) (st []*Message, err error)
	resp, err := app.session.ChannelMessages(channelId, limit, beforeId, afterId)
	if err != nil {
		log.Println("History error: ", err)
		return
	}

	state.Lock()
	defer state.Unlock()

	newMessages := make([]*discordgo.Message, 0)

	if len(channel.Messages) < 1 && len(resp) > 0 {
		for i := len(resp) - 1; i >= 0; i-- {
			channel.Messages = append(channel.Messages, resp[i])
		}
		return
	} else if len(resp) < 1 {
		return
	}

	nextNewMessageIndex := len(resp) - 1
	nextOldMessageIndex := 0

	for {
		newOut := false
		oldOut := false
		var nextOldMessage *discordgo.Message
		if nextOldMessageIndex >= len(channel.Messages) {
			oldOut = true
		} else {
			nextOldMessage = channel.Messages[nextOldMessageIndex]
		}

		var nextNewMessage *discordgo.Message
		if nextNewMessageIndex < 0 {
			newOut = true
		} else {
			nextNewMessage = resp[nextNewMessageIndex]
		}

		if newOut && !oldOut {
			newMessages = append(newMessages, nextOldMessage)
			nextOldMessageIndex++
			continue
		} else if !newOut && oldOut {
			newMessages = append(newMessages, nextNewMessage)
			nextNewMessageIndex--
			continue
		} else if newOut && oldOut {
			break
		}

		if nextNewMessage.ID == nextOldMessage.ID {
			newMessages = append(newMessages, nextNewMessage)
			nextNewMessageIndex--
			nextOldMessageIndex++
			continue
		}

		parsedNew, _ := time.Parse(common.DiscordTimeFormat, nextNewMessage.Timestamp)
		parsedOld, _ := time.Parse(common.DiscordTimeFormat, nextOldMessage.Timestamp)

		if parsedNew.Before(parsedOld) {
			newMessages = append(newMessages, nextNewMessage)
			nextNewMessageIndex--
		} else {
			newMessages = append(newMessages, nextOldMessage)
			nextOldMessageIndex++
		}
	}
	channel.Messages = newMessages
	log.Println("History processing completed!")
}

// func (app *App) ToggleListeningChannel(chId string) {
// 	if !app.AddListeningChannel(chId) {
// 		app.RemoveListeningChannel(chId)
// 	}
// }

func (app *App) TypingStart(s *discordgo.Session, t *discordgo.TypingStart) {
	app.typingManager.in <- t
}
