package main

import (
	"github.com/jonas747/discordgo"
)

type SubbedMessageCreateHandler interface {
	HandleSubbedMessageCreate(m *discordgo.Message)
}

type SubbedMessageUpdateHandler interface {
	HandleSubbedMessageUpdate(m *discordgo.Message)
}

type SubbedMessageRemoveHandler interface {
	HandleSubbedMessageRemove(m *discordgo.Message)
}
