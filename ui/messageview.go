package ui

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/discorder/common"
	"github.com/nsf/termbox-go"
	"time"
	"unicode/utf8"
)

type MessageView struct {
	*BaseEntity
	Transform       *Transform
	DiscordState    *discordgo.State
	DisplayMessages []*DisplayMessage

	Guild       string
	Channels    []string
	ShowPrivate bool
	Logs        []*common.LogMessage // Maybe move this?

	Layer int

	MessageTexts  []*Text
	CurChatScroll int
}

type DisplayMessage struct {
	DiscordMessage *discordgo.Message
	LogMessage     *common.LogMessage
	IsLogMessage   bool
	Timestamp      time.Time
}

func NewMessageView(state *discordgo.State) *MessageView {
	mv := &MessageView{
		BaseEntity:   &BaseEntity{},
		Transform:    &Transform{},
		DiscordState: state,
	}
	return mv
}

func (mv *MessageView) HandleInput(event termbox.Event) {
	if event.Type == termbox.EventResize || event.Type == termbox.EventKey {
		mv.Update()
	}
}

func (mv *MessageView) HandleMessageCreate(session *discordgo.Session, msg *discordgo.Message) {
	// Check if its private and if this messagegview shows private messages
	pChannel, err := mv.DiscordState.PrivateChannel(msg.ChannelID)
	if pChannel != nil && err != nil {
		if mv.ShowPrivate {
			mv.Update()
		} else {
			return
		}
	}

	// Check if its a message were listening to
	for _, v := range mv.Channels {
		if v == msg.ChannelID {
			mv.Update()
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

func (mv *MessageView) Update() {
	mv.BuildTexts()
	h := mv.Transform.GetRect().H
	mv.BuildDisplayMessages(int(h))
}

func (mv *MessageView) BuildTexts() {
	// sizex, sizey := termbox.Size()
	mv.ClearChildren()
	mv.MessageTexts = make([]*Text, 0)

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

		text := NewText()
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
			channel, err := mv.DiscordState.Channel(msg.ChannelID)
			if err != nil {
				errStr := "(error getting channel" + err.Error() + ") "
				fullMsg := ts + errStr + author + ": " + msg.ContentWithMentionsReplaced()
				errLen := utf8.RuneCountInString(errStr)
				points := map[int]AttribPair{
					0:                          AttribPair{termbox.ColorBlue, termbox.ColorRed},
					tsLen:                      AttribPair{termbox.ColorWhite, termbox.ColorRed},
					errLen + tsLen:             AttribPair{termbox.ColorCyan | termbox.AttrBold, termbox.ColorDefault},
					errLen + authorLen + tsLen: AttribPair{},
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
				points := map[int]AttribPair{
					0:                              AttribPair{termbox.ColorBlue, termbox.ColorDefault},
					tsLen:                          AttribPair{termbox.ColorGreen, termbox.ColorDefault},
					channelLen + tsLen:             AttribPair{termbox.ColorCyan | termbox.AttrBold, termbox.ColorDefault},
					channelLen + authorLen + tsLen: AttribPair{},
				}
				if dm {
					points[tsLen] = AttribPair{termbox.ColorMagenta, termbox.ColorDefault}
				}
				text.Text = fullMsg
				text.Attribs = points
			}
		}

		lines := HeightRequired(utf8.RuneCountInString(text.Text), int(rect.W)-padding*2)
		y -= lines
		text.Transform.Position = common.NewVector2I(int(rect.X)+padding, int(rect.Y)+y)
		text.Layer = mv.Layer
		mv.AddChild(text)
	}
}

// A target for optimisation when i get that far
// Also a target for cleaning up
// Builds a list of messages to display from all of the channels were listening to, pm's and the log
func (mv *MessageView) BuildDisplayMessages(size int) {
	// Ackquire the state, or create one if null (incase were starting)
	state := mv.DiscordState
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
	if mv.ShowPrivate {
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
			channel, err := state.GuildChannel(mv.Guild, listeningChannelId)
			state.RLock()
			if err != nil {
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
		if mv.ShowPrivate {
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
	mv.BuildDisplayMessages(int(mv.Transform.GetRect().H))
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
