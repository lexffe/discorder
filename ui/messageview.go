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

	MessageTexts  []*Text
	CurChatScroll int
}

type DisplayMessage struct {
	discordMessage *discordgo.Message
	logMessage     *common.LogMessage
	isLogMessage   bool
	timestamp      time.Time
}

func NewMessageView(state *discordgo.State) *MessageView {
	mv := &MessageView{
		BaseEntity:   &BaseEntity{},
		Transform:    &Transform{},
		DiscordState: state,
	}
	return mv
}

func (mv *MessageView) BuildTexts() {
	// sizex, sizey := termbox.Size()
	mv.ClearChildren()
	mv.MessageTexts = make([]*Text, 0)

	rect := mv.Transform.GetRect()

	y := int(rect.W)
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

		text := NewUIText()
		text.Transform.Size = common.NewVector2F(rect.W, 0)

		if item.isLogMessage {
			//cells = GenCellSlice("Log: "+item.logMessage.content, map[int]AttribPoint{0: AttribPoint{termbox.ColorYellow, termbox.ColorDefault}})
			text.Text = "Log: " + item.logMessage.Content
			text.Attribs = map[int]AttribPair{0: AttribPair{termbox.ColorYellow, termbox.ColorDefault}}
		} else {
			msg := item.discordMessage
			if msg == nil {
				continue
			}
			author := "Unknown?"
			if msg.Author != nil {
				author = msg.Author.Username
			}
			ts := item.timestamp.Local().Format(time.Stamp) + " "
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
		//SetCells(cells, padding, y, sizex-1-padding*2, 0)
	}
}

// A target for optimisation when i get that far
// Also a target for cleaning up
// Builds a list of messages to display from all of the channels were listening to, pm's and the log
func (mv *MessageView) BuildDisplayMessages(size int) {
	// Ackquire the state, or create one if null (incase were starting)
	var state *discordgo.State
	if app.session != nil && app.session.State != nil {
		state = app.session.State
	} else {
		state = discordgo.NewState()
	}
	state.RLock()
	defer state.RUnlock()

	messages := make([]*DisplayMessage, size)

	// Holds the start indexes in the newest message search
	listeningIndexes := make([]int, len(app.listeningChannels))
	pmIndexes := make([]int, len(state.PrivateChannels))
	// Init the slices with silly vals
	for i := 0; i < len(app.listeningChannels); i++ {
		listeningIndexes[i] = -10
	}
	for i := 0; i < len(state.PrivateChannels); i++ {
		pmIndexes[i] = -10
	}
	nextLogIndex := len(app.logBuffer) - 1

	// Get a sorted list
	var lastMessage *DisplayMessage
	var beforeTime time.Time
	for i := 0; i < size; i++ {
		// Get newest message after "lastMessage", set it to curNewestMessage if its newer than that

		var newestListening *DisplayMessage
		newestListeningIndex := 0    // confusing, but the index of the indexes slice
		nextListeningStartIndex := 0 // And the actual next start index to use

		// Check the channels were listening on
		for k, listeningChannelId := range app.listeningChannels {
			// Avoid deadlock since guildchannel also calls rlock, whch will block if there was a new message in the meantime causing lock to be called
			// before that
			state.RUnlock()
			channel, err := state.GuildChannel(app.selectedServerId, listeningChannelId)
			state.RLock()
			if err != nil {
				continue
			}

			newest, nextIndex := GetNewestMessageBefore(channel.Messages, beforeTime, listeningIndexes[k])

			if newest != nil && (newestListening == nil || !newest.timestamp.Before(newestListening.timestamp)) {
				newestListening = newest
				newestListeningIndex = k
				nextListeningStartIndex = nextIndex
			}
		}

		var newestPm *DisplayMessage
		newestPmIndex := 0    // confusing, but the index of the indexes slice
		nextPmStartIndex := 0 // And the actual next start index to use

		// Check for newest pm's
		for k, privateChannel := range state.PrivateChannels {

			newest, nextIndex := GetNewestMessageBefore(privateChannel.Messages, beforeTime, pmIndexes[k])

			if newest != nil && (newestPm == nil || !newest.timestamp.Before(newestPm.timestamp)) {
				newestPm = newest
				newestPmIndex = k
				nextPmStartIndex = nextIndex
			}
		}

		newNextLogIndex := 0
		var newestLog *DisplayMessage

		// Check the logerino
		for j := nextLogIndex; j >= 0; j-- {
			msg := app.logBuffer[j]
			if !msg.timestamp.After(beforeTime) || beforeTime.IsZero() {
				if newestLog == nil || !msg.timestamp.Before(newestLog.timestamp) {
					newestLog = &DisplayMessage{
						logMessage:   msg,
						timestamp:    msg.timestamp,
						isLogMessage: true,
					}
					newNextLogIndex = j - 1
				}
				break // Newest message after last since ordered
			}
		}

		if newestListening != nil &&
			(newestPm == nil || !newestListening.timestamp.Before(newestPm.timestamp)) &&
			(newestLog == nil || !newestListening.timestamp.Before(newestLog.timestamp)) {
			messages[i] = newestListening
			listeningIndexes[newestListeningIndex] = nextListeningStartIndex

			lastMessage = newestListening
			beforeTime = lastMessage.timestamp
		} else if newestPm != nil &&
			(newestListening == nil || !newestPm.timestamp.Before(newestListening.timestamp)) &&
			(newestLog == nil || !newestPm.timestamp.Before(newestLog.timestamp)) {

			messages[i] = newestPm
			pmIndexes[newestPmIndex] = nextPmStartIndex

			lastMessage = newestPm
			beforeTime = lastMessage.timestamp
		} else if newestLog != nil {
			messages[i] = newestLog
			nextLogIndex = newNextLogIndex

			lastMessage = newestLog
			beforeTime = lastMessage.timestamp
		} else {
			break // No new shit!
		}
	}
	app.displayMessages = messages
}
func GetNewestMessageBefore(msgs []*discordgo.Message, before time.Time, startIndex int) (*DisplayMessage, int) {
	if startIndex == -10 {
		startIndex = len(msgs) - 1
	}

	for j := startIndex; j >= 0; j-- {
		msg := msgs[j]
		parsedTimestamp, _ := time.Parse(DiscordTimeFormat, msg.Timestamp)
		if !parsedTimestamp.After(before) || before.IsZero() { // Reason for !after is so that we still show all the messages with same timestamps
			curNewestMessage := &DisplayMessage{
				discordMessage: msg,
				timestamp:      parsedTimestamp,
			}
			return curNewestMessage, j - 1
		}
	}
	return nil, 0
}
