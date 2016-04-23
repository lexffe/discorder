package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/discorder/common"
	"log"
	"time"
)

// For logs, should probably move this somewhere else though
func (app *App) Write(p []byte) (n int, err error) {
	cop := string(p)
	app.HandleLogMessage(cop)
	return len(p), nil
}

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

func (app *App) TypingStart(s *discordgo.Session, t *discordgo.TypingStart) {
	app.typingManager.in <- t
}
