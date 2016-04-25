package main

import (
	"fmt"
	"github.com/jonas747/discorder/common"
	"github.com/jonas747/discorder/ui"
	"github.com/jonas747/discordgo"
	"github.com/nsf/termbox-go"
	"log"
)

type NotificationsManager struct {
	*ui.BaseEntity
	Transform *ui.Transform
	App       *App
	text      *ui.Text

	unread map[string]int
}

func NewNotificationsManager(app *App) *NotificationsManager {
	t := ui.NewText()

	nm := &NotificationsManager{
		BaseEntity: &ui.BaseEntity{},
		Transform:  &ui.Transform{},
		App:        app,
		text:       t,
	}

	t.Transform.Parent = nm.Transform
	t.Transform.AnchorMax = common.NewVector2I(1, 1)
	t.BG = termbox.ColorYellow
	nm.AddChild(t)
	return nm
}

func (nm *NotificationsManager) PreDraw() {
	str := ""
	if len(nm.unread) > 0 {
		total := 0
		for k, v := range nm.unread {
			total += v
			channel, err := nm.App.session.State.Channel(k)
			if err != nil {
				log.Println("Error getting channel:", err)
				continue
			}

			guild, err := nm.App.session.State.Guild(channel.GuildID)
			if err != nil {
				log.Println("Error getting guild:", err)
				continue
			}

			str += fmt.Sprintf("%s/%s: %d, ", guild.Name, GetChannelNameOrRecipient(channel), v)
		}
		str = str[:len(str)-1]
		str = fmt.Sprintf("Unread messages: %d (%s)", total, str)
	}
	nm.text.Text = str
}

func (nm *NotificationsManager) AddMessageNotification(msg *discordgo.Message) {
	nm.unread[msg.ChannelID]++
}

func (nm *NotificationsManager) RemoveMessageNotification(msg *discordgo.Message) {
	nm.unread[msg.ChannelID]--
}
func (nm *NotificationsManager) Destroy() { nm.DestroyChildren() }
