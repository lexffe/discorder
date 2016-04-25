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

	activeNotifications []*discordgo.Message
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
	if len(nm.activeNotifications) > 0 {
		str = fmt.Sprintf("%d Unread: ", len(nm.activeNotifications))
		for i, v := range nm.activeNotifications {
			channel, err := nm.App.session.State.Channel(v.ChannelID)
			if err != nil {
				log.Println("Error getting channel:", err)
				continue
			}

			guild, err := nm.App.session.State.Guild(channel.GuildID)
			if err != nil {
				log.Println("Error getting guild:", err)
				continue
			}

			author := "Unknown?"
			if v.Author != nil {
				author = v.Author.Username
			}
			str += fmt.Sprintf("%s/%s@%s", guild.Name, GetChannelNameOrRecipient(channel), author)
			if i != len(nm.activeNotifications)-1 {
				str += ", "
			}
		}
	}
	nm.text.Text = str
}

func (nm *NotificationsManager) AddMessageNotification(msg *discordgo.Message) {
	log.Println("Added a notification")
	nm.activeNotifications = append(nm.activeNotifications, msg)
}

func (nm *NotificationsManager) RemoveMessageNotification(msg *discordgo.Message) bool {
	for i, v := range nm.activeNotifications {
		if v.ID == msg.ID {
			nm.activeNotifications = append(nm.activeNotifications[:i], nm.activeNotifications[i+1:]...)
			return true
		}
	}
	return false
}
func (nm *NotificationsManager) Destroy() { nm.DestroyChildren() }
