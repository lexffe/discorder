package main

import (
	"github.com/jonas747/discorder/common"
	"github.com/jonas747/discorder/ui"
	"github.com/jonas747/discordgo"
	"github.com/nsf/termbox-go"
	"time"
	"unicode/utf8"
)

type MessageView struct {
	*ui.BaseEntity
	Transform       *ui.Transform
	App             *App
	DisplayMessages []*DisplayMessage

	Channels       []string
	ShowAllPrivate bool
	Logs           []*common.LogMessage // Maybe move this?

	Layer int

	MessageTexts  []*ui.Text
	CurChatScroll int

	DisplayMessagesDirty bool // Rebuilds displaymessages on next draw if set
	TextsDirty           bool // Rebuilds texts on next draw if set
}

type DisplayMessage struct {
	DiscordMessage *discordgo.Message
	LogMessage     *common.LogMessage
	IsLogMessage   bool
	Timestamp      time.Time
}

func NewMessageView(app *App) *MessageView {
	mv := &MessageView{
		BaseEntity: &ui.BaseEntity{},
		Transform:  &ui.Transform{},
		App:        app,
	}
	return mv
}

func (mv *MessageView) AddChannel(channel string) {
	if mv.Channels == nil {
		mv.Channels = []string{channel}
	} else {
		for _, v := range mv.Channels {
			if v == channel {
				return
			}
		}
		mv.Channels = append(mv.Channels, channel)
	}

	mv.DisplayMessagesDirty = true

	discordChannel, err := mv.App.session.State.Channel(channel)
	if err != nil {
		return
	}
	// Grab some history
	if len(discordChannel.Messages) < 10 {
		mv.App.GetHistory(channel, 10, "", "")
	}
}

func (mv *MessageView) RemoveChannel(channel string) {
	index := -1
	for k, v := range mv.Channels {
		if channel == v {
			index = k
			break
		}
	}

	if index == -1 {
		return // no channel by that name
	}

	if index == len(mv.Channels)-1 {
		mv.Channels = mv.Channels[:index]
	} else {
		mv.Channels = append(mv.Channels[:index], mv.Channels[index+1:]...)
	}
}

func (mv *MessageView) HandleInput(event termbox.Event) {
	if event.Type == termbox.EventResize || event.Type == termbox.EventKey {
		mv.TextsDirty = true //  ;)
	}
}

func (mv *MessageView) HandleMessageCreate(session *discordgo.Session, msg *discordgo.Message) {
	// Check if its private and if this messagegview shows private messages
	pChannel, err := mv.App.session.State.PrivateChannel(msg.ChannelID)
	if pChannel != nil && err != nil {
		mv.DisplayMessagesDirty = true
		return
	}

	// Check if its a message were listening to
	for _, v := range mv.Channels {
		if v == msg.ChannelID {
			mv.DisplayMessagesDirty = true
			break
		}
	}
}

func (mv *MessageView) HandleMessageEdit(session *discordgo.Session, msg *discordgo.Message) {
	mv.HandleMessageCreate(session, msg)
}

func (mv *MessageView) HandleMessageRemove(session *discordgo.Session, msg *discordgo.Message) {
	mv.HandleMessageCreate(session, msg)
}

func (mv *MessageView) BuildTexts() {
	// sizex, sizey := termbox.Size()
	mv.ClearChildren()
	mv.MessageTexts = make([]*ui.Text, 0)

	rect := mv.Transform.GetRect()

	y := int(rect.H)
	padding := 0

	// Build it!!
	for k, item := range mv.DisplayMessages {
		//var cells []termbox.Cell

		if item == nil {
			continue
		}

		if k < mv.CurChatScroll {
			continue
		}

		text := ui.NewText()
		text.Transform.Size = common.NewVector2F(rect.W, 0)

		if item.IsLogMessage {
			//cells = GenCellSlice("Log: "+item.logMessage.content, map[int]AttribPoint{0: AttribPoint{termbox.ColorYellow, termbox.ColorDefault}})
			text.Text = "Log: " + item.LogMessage.Content
			//text.Attribs = map[int]AttribPair{0: AttribPair{termbox.ColorYellow, termbox.ColorDefault}}
		} else {
			msg := item.DiscordMessage
			if msg == nil {
				continue
			}
			author := "Unknown?"
			if msg.Author != nil {
				author = msg.Author.Username
			}
			ts := item.Timestamp.Local().Format(time.Stamp) + " "
			tsLen := utf8.RuneCountInString(ts)

			authorLen := utf8.RuneCountInString(author)
			channel, err := mv.App.session.State.Channel(msg.ChannelID)
			if err != nil {
				errStr := "(error getting channel" + err.Error() + ") "
				fullMsg := ts + errStr + author + ": " + msg.ContentWithMentionsReplaced()
				errLen := utf8.RuneCountInString(errStr)
				points := map[int]ui.AttribPair{
					0:                          ui.AttribPair{termbox.ColorBlue, termbox.ColorRed},
					tsLen:                      ui.AttribPair{termbox.ColorWhite, termbox.ColorRed},
					errLen + tsLen:             ui.AttribPair{termbox.ColorCyan | termbox.AttrBold, termbox.ColorDefault},
					errLen + authorLen + tsLen: ui.AttribPair{},
				}
				text.Text = fullMsg
				text.Attribs = points
			} else {
				name := channel.Name
				dm := false
				if name == "" {
					name = "Direct Message"
					dm = true
				}

				fullMsg := ts + "[" + name + "]" + author + ": " + msg.ContentWithMentionsReplaced()
				channelLen := utf8.RuneCountInString(name) + 2
				points := map[int]ui.AttribPair{
					0:                              ui.AttribPair{termbox.ColorBlue, termbox.ColorDefault},
					tsLen:                          ui.AttribPair{termbox.ColorGreen, termbox.ColorDefault},
					channelLen + tsLen:             ui.AttribPair{termbox.ColorCyan | termbox.AttrBold, termbox.ColorDefault},
					channelLen + authorLen + tsLen: ui.AttribPair{},
				}
				if dm {
					points[tsLen] = ui.AttribPair{termbox.ColorMagenta, termbox.ColorDefault}
				}
				text.Text = fullMsg
				text.Attribs = points
			}
		}

		lines := ui.HeightRequired(utf8.RuneCountInString(text.Text), int(rect.W)-padding*2)
		y -= lines
		if y < 0 {
			break
		}
		text.Transform.Position = common.NewVector2I(int(rect.X)+padding, int(rect.Y)+y)
		text.Layer = mv.Layer
		mv.AddChild(text)
	}
}

// TODO: Merge private and normal channels to make this a little big ligther
// A target for optimisation when i get that far
// Also a target for cleaning up
// Builds a list of messages to display from all of the channels were listening to, pm's and the log
func (mv *MessageView) BuildDisplayMessages(size int) {
	// Ackquire the state, or create one if null (incase were starting)
	state := mv.App.session.State
	if state == nil {
		state = discordgo.NewState()
	}
	state.RLock()
	defer state.RUnlock()

	messages := make([]*DisplayMessage, size)

	// Holds the start indexes in the newest message search
	listeningIndexes := make([]int, len(mv.Channels))
	pmIndexes := make([]int, len(state.PrivateChannels))
	// Init the slices with silly vals
	for i := 0; i < len(mv.Channels); i++ {
		listeningIndexes[i] = -10
	}
	if mv.ShowAllPrivate {
		for i := 0; i < len(state.PrivateChannels); i++ {
			pmIndexes[i] = -10
		}
	}
	nextLogIndex := len(mv.Logs) - 1

	// Get a sorted list
	var lastMessage *DisplayMessage
	var beforeTime time.Time
	for i := 0; i < size; i++ {
		// Get newest message after "lastMessage", set it to curNewestMessage if its newer than that

		var newestListening *DisplayMessage
		newestListeningIndex := 0    // confusing, but the index of the indexes slice
		nextListeningStartIndex := 0 // And the actual next start index to use

		// Check the channels were listening on
		for k, listeningChannelId := range mv.Channels {
			// Avoid deadlock since guildchannel also calls rlock, whch will block if there was a new message in the meantime causing lock to be called
			// before that
			state.RUnlock()
			channel, err := state.Channel(listeningChannelId)
			state.RLock()
			if err != nil || (channel.IsPrivate && mv.ShowAllPrivate) {
				continue
			}

			newest, nextIndex := GetNewestMessageBefore(channel.Messages, beforeTime, listeningIndexes[k])

			if newest != nil && (newestListening == nil || !newest.Timestamp.Before(newestListening.Timestamp)) {
				newestListening = newest
				newestListeningIndex = k
				nextListeningStartIndex = nextIndex
			}
		}

		var newestPm *DisplayMessage
		newestPmIndex := 0    // confusing, but the index of the indexes slice
		nextPmStartIndex := 0 // And the actual next start index to use

		// Check for newest pm's
		if mv.ShowAllPrivate {
			for k, privateChannel := range state.PrivateChannels {

				newest, nextIndex := GetNewestMessageBefore(privateChannel.Messages, beforeTime, pmIndexes[k])

				if newest != nil && (newestPm == nil || !newest.Timestamp.Before(newestPm.Timestamp)) {
					newestPm = newest
					newestPmIndex = k
					nextPmStartIndex = nextIndex
				}
			}
		}

		newNextLogIndex := 0
		var newestLog *DisplayMessage

		// Check the logerino
		for j := nextLogIndex; j >= 0; j-- {
			msg := mv.Logs[j]
			if !msg.Timestamp.After(beforeTime) || beforeTime.IsZero() {
				if newestLog == nil || !msg.Timestamp.Before(newestLog.Timestamp) {
					newestLog = &DisplayMessage{
						LogMessage:   msg,
						Timestamp:    msg.Timestamp,
						IsLogMessage: true,
					}
					newNextLogIndex = j - 1
				}
				break // Newest message after last since ordered
			}
		}

		if newestListening != nil &&
			(newestPm == nil || !newestListening.Timestamp.Before(newestPm.Timestamp)) &&
			(newestLog == nil || !newestListening.Timestamp.Before(newestLog.Timestamp)) {
			messages[i] = newestListening
			listeningIndexes[newestListeningIndex] = nextListeningStartIndex

			lastMessage = newestListening
			beforeTime = lastMessage.Timestamp
		} else if newestPm != nil &&
			(newestListening == nil || !newestPm.Timestamp.Before(newestListening.Timestamp)) &&
			(newestLog == nil || !newestPm.Timestamp.Before(newestLog.Timestamp)) {

			messages[i] = newestPm
			pmIndexes[newestPmIndex] = nextPmStartIndex

			lastMessage = newestPm
			beforeTime = lastMessage.Timestamp
		} else if newestLog != nil {
			messages[i] = newestLog
			nextLogIndex = newNextLogIndex

			lastMessage = newestLog
			beforeTime = lastMessage.Timestamp
		} else {
			break // No new shit!
		}
	}
	mv.DisplayMessages = messages
}

func (mv *MessageView) Destroy() { mv.DestroyChildren() }

func (mv *MessageView) PreDraw() {
	h := int(mv.Transform.GetRect().H)
	if h < 0 {
		h = 0
	}
	mv.BuildDisplayMessages(h)
	mv.BuildTexts()
}

func (mv *MessageView) GetDrawLayer() int {
	return 9
}

func GetNewestMessageBefore(msgs []*discordgo.Message, before time.Time, startIndex int) (*DisplayMessage, int) {
	if startIndex == -10 {
		startIndex = len(msgs) - 1
	}

	for j := startIndex; j >= 0; j-- {
		msg := msgs[j]
		parsedTimestamp, _ := time.Parse(common.DiscordTimeFormat, msg.Timestamp)
		if !parsedTimestamp.After(before) || before.IsZero() { // Reason for !after is so that we still show all the messages with same timestamps
			curNewestMessage := &DisplayMessage{
				DiscordMessage: msg,
				Timestamp:      parsedTimestamp,
			}
			return curNewestMessage, j - 1
		}
	}
	return nil, 0
}
