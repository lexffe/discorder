package discorder

import (
	"github.com/jonas747/discordgo"
	"log"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

var linkRegex = regexp.MustCompile(`(https?|ftp):\/\/[-a-zA-Z0-9@:%._\+~#=]{2,256}\.[a-z]{2,6}\b([-a-zA-Z0-9@:%_\+.~#?&\/\/=]*)`)

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

func (app *App) GenMessageCommands(msg *discordgo.Message) []Command {
	content := msg.ContentWithMentionsReplaced()

	matches := linkRegex.FindAllString(content, -1)
	if len(matches) < 1 && len(msg.Attachments) < 1 {
		return nil
	}

	out := make([]Command, 0)
	for _, v := range matches {
		linkCopy := v
		out = append(out, &SimpleCommand{
			Name:         "Open " + v,
			Description:  "Open the link using xdg-open or join the server",
			IgnoreFilter: true,
			RunFunc: func(app *App, args Arguments) {
				go app.OpenLink(linkCopy)
			},
		})
	}

	for _, v := range msg.Attachments {
		linkCopy := v.URL
		out = append(out, &SimpleCommand{
			Name:         "Open " + linkCopy,
			Description:  "Open the link using xdg-open or join the server",
			IgnoreFilter: true,
			RunFunc: func(app *App, args Arguments) {
				go app.OpenLink(linkCopy)
			},
		})
	}

	return out
}

func (app *App) OpenLink(link string) {

	// Checks if it is in macOS
	// In Windows, start [URL] might work, though it is not implemented here.

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", link)
		break
	case "linux":
	default:
		cmd = exec.Command("xdg-open", link)
		break
	}

	err := cmd.Run()
	if err != nil {
		log.Println("Failed to run opening link with xdg-open / open", err)
	}
}
