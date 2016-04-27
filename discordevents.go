package main

import (
	"github.com/0xAX/notificator"
	"github.com/jonas747/discordgo"
	"log"
)

func (app *App) Ready(s *discordgo.Session, r *discordgo.Ready) {
	app.Lock()
	defer app.Unlock()
	log.Println("Received ready from discord!")

	app.settings = r.Settings
	app.ViewManager.OnReady()
	app.PrintWelcome()
}

func (app *App) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// channel, err := app.session.State.Channel(m.ChannelID)
	// if err != nil {
	// 	log.Println("Error retrieving channel from state", err)
	// 	return
	// }
	app.Lock()
	defer app.Unlock()
	settings := app.GetNotificationSettingsForChannel(m.ChannelID)

	author := "Unknown?"
	if m.Author != nil { // Yes this does happen...
		author = m.Author.Username
	}

	shouldNotify := false

	if !settings.Muted && settings.Notifications == ChannelNotificationsAll {
		shouldNotify = true
	} else if !settings.Muted && settings.Notifications == ChannelNotificationsMentions {
		for _, v := range m.Mentions {
			if v.ID == s.State.User.ID {
				shouldNotify = true
				break
			}
		}
	} else if !settings.SurpressEveryone && m.MentionEveryone {
		shouldNotify = true
	}

	if shouldNotify {
		if app.notifications != nil {
			app.notifications.Push(author, m.ContentWithMentionsReplaced(), "", notificator.UR_NORMAL)
		}
		app.ViewManager.notificationsManager.AddMention(m.Message)
	}

	// Update last message
	channel, err := app.session.State.Channel(m.ChannelID)
	if err != nil {
		log.Println("Error getting channel", err)
	} else {
		channel.LastMessageID = m.ID
	}

	// Check if we should ack the message, moved this to messageview for now
	// msgVisible := false

	// channel, err := app.session.State.Channel(m.ChannelID)
	// if err != nil {
	// 	log.Println("Error getting channel", err)
	// }
	// if channel != nil && channel.IsPrivate && app.ViewManager.mv.ShowAllPrivate {
	// 	msgVisible = true
	// } else {
	// 	for _, c := range app.ViewManager.mv.Channels {
	// 		if c == m.ChannelID {
	// 			msgVisible = true
	// 			break
	// 		}
	// 	}
	// }

	// if msgVisible {
	// 	if m.Author != nil {
	// 		if m.Author.ID != app.session.State.User.ID {
	// 			app.ackRoutine.In <- m.Message
	// 		}
	// 	}
	// }

}

func (app *App) messageUpdate(s *discordgo.Session, m *discordgo.MessageUpdate) {

}

func (app *App) messageDelete(s *discordgo.Session, m *discordgo.MessageDelete) {

}

func (app *App) messageAck(s *discordgo.Session, a *discordgo.MessageAck) {
	log.Println("Received ack!")
	app.ViewManager.notificationsManager.HandleAck(a)
}

func (app *App) guildSettingsUpdated(s *discordgo.Session, a *discordgo.UserGuildSettingsUpdate) {

}

func (app *App) userSettingsUpdated(s *discordgo.Session, u *discordgo.UserSettingsUpdate) {
	// for k, v := range map[string]interface{}(*u) {
	// 	switch k {
	// 	case "theme":
	// 		str, _ := v.(string)
	// 		app.settings.Theme = str
	// 	case "friend_source_flags":
	// 		flags, ok := v.(map[string]interface{})
	// 		log.Println(ok, flags)
	// 	case "restricted_guilds":
	// 		slice, ok := v.([]interface{})
	// 		log.Println(ok, slice)
	// 	}
	// }
}

func (app *App) typingStart(s *discordgo.Session, t *discordgo.TypingStart) {
	app.typingManager.in <- t
}

func (app *App) guildCreated(s *discordgo.Session, g *discordgo.GuildCreate) {
	log.Println("Guild created!", g.Guild)
}