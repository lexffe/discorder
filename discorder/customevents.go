package discorder

import (
	"github.com/jonas747/discordgo"
)

type MessageCreateHandler interface {
	HandleMessageCreate(m *discordgo.Message)
}

type MessageUpdateHandler interface {
	HandleMessageUpdate(m *discordgo.Message)
}

type MessageRemoveHandler interface {
	HandleMessageRemove(m *discordgo.Message)
}
