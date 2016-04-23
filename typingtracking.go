package main

import (
	"github.com/jonas747/discorder/common"
	"github.com/jonas747/discorder/ui"
	"github.com/nsf/termbox-go"
)

// Shows whos typing
type TypingDisplay struct {
	*ui.BaseEntity
	Transform *ui.Transform

	App *App

	text *ui.Text
}

func NewTypingDisplay(app *App) *TypingDisplay {
	td := &TypingDisplay{
		BaseEntity: &ui.BaseEntity{},
		Transform:  &ui.Transform{},
		App:        app,
	}

	t := ui.NewText()

	t.Transform.Parent = td.Transform
	t.Transform.AnchorMax = common.NewVector2I(1, 1)
	t.FG = termbox.ColorCyan
	td.AddChild(t)
	td.text = t
	return td
}

func (t *TypingDisplay) PreDraw() {
	typing := t.App.typingManager.GetTyping([]string{})

	if len(typing) > 0 {

		typingStr := "Typing: "

		for _, v := range typing {
			channel, err := t.App.session.State.Channel(v.ChannelID)
			if err != nil {
				continue
			}

			member, err := t.App.session.State.Member(channel.GuildID, v.UserID)
			if err != nil {
				continue
			}
			typingStr += channel.Name + ":" + member.User.Username + ", "
		}
		// Remove trailing ","
		typingStr = typingStr[:len(typingStr)-1]
		t.text.Text = typingStr
	} else {
		t.text.Text = "No one is typing :'("
	}
}

func (t *TypingDisplay) Destroy() { t.DestroyChildren() }
