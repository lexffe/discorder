package main

import (
	"github.com/jonas747/discordgo"
)

type VisibleMessageCreateHandler interface {
	HandleVisibleMessageCreate(m *discordgo.Message)
}

type VisibleMessageUpdateHandler interface {
	HandleVisibleMessageUpdate(m *discordgo.Message)
}

type VisibleMessageRemoveHandler interface {
	HandleVisibleMessageRemove(m *discordgo.Message)
}
