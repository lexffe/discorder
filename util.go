package main

import (
	"github.com/jonas747/discorder/common"
	"github.com/jonas747/discordgo"
	"log"
	"time"
)

// For logs, should probably move this somewhere else though
func (app *App) Write(p []byte) (n int, err error) {
	cop := string(p)
	app.HandleLogMessage(cop)
	return len(p), nil
}

// If ran from another goroutine, lock should be true
// gonna turn this into its own history manager later with a queue and everything
func (app *App) GetHistory(channelId string, limit int, beforeId, afterId string, lock bool) {
	defer func() {
		if lock {
			app.Lock()
		}
		delete(app.fetchingHistory, channelId)
		app.ViewManager.mv.DisplayMessagesDirty = true
		if lock {
			app.Unlock()
		}
		log.Println("History processing complete")
	}()

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

	if len(channel.Messages) < 1 && len(resp) > 0 {
		for i := len(resp) - 1; i >= 0; i-- {
			channel.Messages = append(channel.Messages, resp[i])
		}
		return
	} else if len(resp) < 1 {
		if lock {
			state.Unlock() // Also need to unlock this to avoid deadlocks
			app.Lock()
		}
		//log.Println("Set first message")
		app.firstMessages[channelId] = beforeId
		if lock {
			app.Unlock()
			state.Lock() // And lock it again
		}

		return
	}

	if len(resp) < limit {
		// Looks like we've hit the first message of the channel
		if lock {
			state.Unlock() // Also need to unlock this to avoid deadlocks
			app.Lock()
		}
		newFirst := resp[len(resp)-1].ID
		app.firstMessages[channelId] = newFirst

		if lock {
			app.Unlock()
			state.Lock() // And lock it again
		}
	}

	newMessages := make([]*discordgo.Message, 0)
	nextNewMessageIndex := len(resp) - 1
	nextOldMessageIndex := 0

	for {
		newOut := false // new (response) is oob
		oldOut := false // old (current channel history) is oob
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
	if len(resp) > 0 {
		app.ackRoutine.In <- resp[0]
	}
}

func (app *App) GetNotificationSettingsForChannel(channelId string) *ChannelNotificationSettings {
	channel, err := app.session.Channel(channelId)
	if err != nil {
		log.Println("Error getting channel from state", err)
		return nil
	}

	if channel.IsPrivate {
		return &ChannelNotificationSettings{Notifications: ChannelNotificationsAll}
	}

	for _, gs := range app.guildSettings {
		if gs.GuildID == channel.GuildID {

			cn := &ChannelNotificationSettings{
				Notifications:    gs.MessageNotifications,
				Muted:            gs.Muted,
				SurpressEveryone: gs.SupressEveryone,
			}
			if gs.Muted {
				return cn
			}
			for _, override := range gs.ChannelOverrides {
				if override.ChannelID == channel.ID {
					cn.Notifications = override.MessageNotifications
					cn.Muted = override.Muted
					break
				}
			}
			return cn
		}
	}

	// Use default guild settings
	guild, err := app.session.Guild(channel.GuildID)
	if err != nil {
		log.Println("Error getting guild from state", err)
		return nil
	}
	return &ChannelNotificationSettings{
		Notifications: guild.DefaultMessageNotifications,
	}
}

// Compare readstate's last_message to channel's last_message and if theres new show so
// Also number of mentions
// Take notifications settings into mind also
func (app *App) GetStartNotifications() {
	// readStates := app.session.State.ReadState
	// for _, state := range readStates {
	// }
}

const (
	ChannelNotificationsAll      = 0
	ChannelNotificationsMentions = 1
	ChannelNotificationsNothing  = 2
)

type ChannelNotificationSettings struct {
	Notifications    int // 0 all, 1 mentions, 2 nothing
	Muted            bool
	SurpressEveryone bool
}

func GetChannelNameOrRecipient(channel *discordgo.Channel) string {
	if channel.IsPrivate {
		if channel.Recipient != nil {
			return channel.Recipient.Username
		} else {
			return "Recipient is nil!?"
		}
	}
	return channel.Name
}

func GetMessageAuthor(msg *discordgo.Message) string {
	if msg.Author != nil {
		return msg.Author.Username
	}
	return "Unknwon?"
}

func (app *App) IsFirstChannelMessage(channelId, msgId string) bool {
	first, ok := app.firstMessages[channelId]
	if !ok {
		return false
	}

	if first == msgId {
		return true
	}
	return false
}
