package discorder

import (
	"github.com/jonas747/discordgo"
	"log"
	"strconv"
	"strings"
)

func (app *App) GetNotificationSettingsForChannel(channelId string) *ChannelNotificationSettings {
	channel, err := app.session.State.Channel(channelId)
	if err != nil {
		log.Println("Error getting channel from state", err)
		return &ChannelNotificationSettings{}
	}

	if channel.IsPrivate {
		return &ChannelNotificationSettings{Notifications: MessageNotificationsAll}
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
		return &ChannelNotificationSettings{}
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
	MessageNotificationsAll      = 0
	MessageNotificationsMentions = 1
	MessageNotificationsNothing  = 2
	MessageNotificationsServer   = 3
)

func StringNotificationsSettings(notifications int) string {
	switch notifications {
	case MessageNotificationsAll:
		return "all"
	case MessageNotificationsMentions:
		return "mentions"
	case MessageNotificationsNothing:
		return "nothing"
	case MessageNotificationsServer:
		return "server"

	}
	return "??? (discorder outdated, please bug me about this, " + strconv.FormatInt(int64(notifications), 10) + ")"
}

func MessageNotificationsFromString(str string) int {
	switch strings.ToLower(str) {
	case "all":
		return MessageNotificationsAll
	case "mentions", "mention":
		return MessageNotificationsMentions
	case "none", "nothing":
		return MessageNotificationsNothing
	case "server", "default":
		return MessageNotificationsServer
	}
	log.Println("Encountered unknown message notification string", str, "Defaulting to mentions")
	return MessageNotificationsMentions
}

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
