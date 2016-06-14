package discorder

import (
	"github.com/jonas747/discorder/ui"
	"github.com/jonas747/discordgo"
	"log"
	"strconv"
)

type ServerNotificationSettingsCommand struct {
	app *App
}

func (s *ServerNotificationSettingsCommand) GetName() string {
	return "server_notifications_settings"
}

func (s *ServerNotificationSettingsCommand) GetDescription(app *App) string {
	return "Change notifications settings for a server"
}

func (s *ServerNotificationSettingsCommand) GetExecText() string {
	return "Save"
}

func (s *ServerNotificationSettingsCommand) GetCategory() []string {
	return []string{"Discord"}
}

func (s *ServerNotificationSettingsCommand) GetPreRunHelper() string {
	return "server"
}

func (s *ServerNotificationSettingsCommand) GetArgs(curArgs Arguments) []*ArgumentDef {
	args := []*ArgumentDef{
		&ArgumentDef{
			Name:                   "server",
			Description:            "Server to change settings on",
			Datatype:               ui.DataTypeString,
			Helper:                 &ServerChannelArgumentHelper{},
			RebuildOnChanged:       true,
			PreserveValueOnRebuild: true,
		},
	}
	if curArgs == nil {
		log.Println(args)
		return args
	}

	serverId, _ := curArgs.String("server")

	var settings *discordgo.UserGuildSettings
	for _, v := range s.app.guildSettings {
		if v.GuildID == serverId {
			settings = v
		}
	}

	if settings == nil {
		log.Println("Failed to find guild :(")
		return args
	}

	guild, err := s.app.session.State.Guild(serverId)
	if err != nil {
		log.Println("Failed to find guild in state", err)
		return args
	}

	args[0].CurVal = serverId

	args = append(args, &ArgumentDef{
		Name:        "mute_server",
		Description: "Mute the server, Muting a server prevents unread indicators and notifications from appearing unless you are mentioned.",
		Datatype:    ui.DataTypeBool,
		CurVal:      strconv.FormatBool(settings.Muted),
	}, &ArgumentDef{
		Name:        "surpress_everyone",
		Description: "Surpress @everyone and @here",
		Datatype:    ui.DataTypeBool,
		CurVal:      strconv.FormatBool(settings.SupressEveryone),
	}, &ArgumentDef{
		Name:        "mobile_push",
		Description: "Send mobile push notifications",
		Datatype:    ui.DataTypeBool,
		CurVal:      strconv.FormatBool(settings.MobilePush),
	}, &ArgumentDef{
		Name:        "server_notifications",
		Description: "Server notifications level (one of 'all', 'mentions', 'nothing')",
		Datatype:    ui.DataTypeString,
		CurVal:      StringNotificationsSettings(settings.MessageNotifications),
	})

	// Channel overrides
	for _, channel := range guild.Channels {
		if channel.Type != "text" {
			continue
		}
		var channelOverride *discordgo.UserGuildSettingsChannelOverride
		for _, cs := range settings.ChannelOverrides {
			if cs.ChannelID == channel.ID {
				channelOverride = cs
				break
			}
		}

		channelNotify := settings.MessageNotifications
		channelMuted := settings.Muted

		if channelOverride != nil {
			channelNotify = channelOverride.MessageNotifications
			channelMuted = channelOverride.Muted
		}

		args = append(args, &ArgumentDef{
			Name:        channel.ID + "_muted",
			DisplayName: "Override: " + channel.Name + " muted",
			Description: "Mute this channel?",
			Datatype:    ui.DataTypeBool,
			CurVal:      strconv.FormatBool(channelMuted),
		}, &ArgumentDef{
			Name:        channel.ID + "_notifications",
			DisplayName: "Override: " + channel.Name + " notifications",
			Description: "One of 'all', 'mentions', 'nothing', 'server', if server, then server setting is used",
			Datatype:    ui.DataTypeString,
			CurVal:      StringNotificationsSettings(channelNotify),
		})
	}
	return args
}

func (s *ServerNotificationSettingsCommand) Run(app *App, args Arguments) {
}
