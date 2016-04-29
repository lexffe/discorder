package main

import (
	"fmt"
	"github.com/jonas747/discorder/common"
	"github.com/jonas747/discorder/ui"
	"github.com/jonas747/discordgo"
	"github.com/nsf/termbox-go"
	"log"
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

	MessageContainer *ui.SimpleEntity
	MessageTexts     []*ui.Text
	ScrollText       *ui.Text

	ScrollAmount int

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
		BaseEntity:       &ui.BaseEntity{},
		Transform:        &ui.Transform{},
		App:              app,
		MessageContainer: ui.NewSimpleEntity(),
	}

	t := ui.NewText()
	t.Transform.Parent = mv.Transform
	t.Transform.AnchorMin.Y = 1
	t.Transform.AnchorMax = common.NewVector2I(1, 1)
	t.Layer = 9
	t.BG = termbox.ColorYellow

	mv.AddChild(t)
	mv.ScrollText = t

	mv.AddChild(mv.MessageContainer)

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
		firstId := ""
		if len(discordChannel.Messages) > 0 {
			firstId = discordChannel.Messages[0].ID
		}

		if !mv.App.IsFirstChannelMessage(discordChannel.ID, firstId) {
			mv.App.fetchingHistory[channel] = true
			go mv.App.GetHistory(channel, 10, "", "", true)
			if *flagDebugEnabled {
				log.Println("Fetching :" + GetChannelNameOrRecipient(discordChannel))
			}
		}

	}
}

func (mv *MessageView) RemoveChannel(channel string) {
	for k, v := range mv.Channels {
		if channel == v {
			mv.Channels = append(mv.Channels[:k], mv.Channels[k+1:]...)
			break
		}
	}
}

func (mv *MessageView) HandleInput(event termbox.Event) {
	if event.Type == termbox.EventResize || event.Type == termbox.EventKey {
		mv.TextsDirty = true //  ;)

		if mv.App.ViewManager.activeWindow != nil {
			return
		}

		switch event.Key {
		case termbox.KeyArrowUp:
			mv.ScrollAmount += 1
			mv.DisplayMessagesDirty = true
		case termbox.KeyArrowDown:
			mv.ScrollAmount -= 1
			if mv.ScrollAmount < 0 {
				mv.ScrollAmount = 0
			}
			mv.DisplayMessagesDirty = true
		case termbox.KeyPgup:
			mv.ScrollAmount += 10
			mv.DisplayMessagesDirty = true
		case termbox.KeyPgdn:
			mv.ScrollAmount -= 10
			if mv.ScrollAmount < 0 {
				mv.ScrollAmount = 0
			}
			mv.DisplayMessagesDirty = true
		case termbox.KeyHome, termbox.KeyEnd:
			mv.ScrollAmount = 0
			mv.DisplayMessagesDirty = true
		case termbox.KeyEnter:
			if mv.ScrollAmount < 1 || len(mv.MessageTexts) < 1 {
				return
			}

			text := mv.MessageTexts[0]
			selectedDisplayMsg, ok := text.Userdata.(*DisplayMessage)
			if !ok || selectedDisplayMsg.IsLogMessage {
				return
			}
			msw := NewMessageSelectedWindow(mv.App, selectedDisplayMsg.DiscordMessage)
			mv.App.ViewManager.AddChild(msw)
			mv.App.ViewManager.activeWindow = msw
		}
		if mv.ScrollAmount != 0 {
			mv.App.ViewManager.input.Active = false
		} else {
			mv.App.ViewManager.input.Active = true
		}
	}
}

func (mv *MessageView) HandleMessageCreate(msg *discordgo.Message) {
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

func (mv *MessageView) HandleMessageEdit(msg *discordgo.Message) {
	mv.HandleMessageCreate(msg)
}

func (mv *MessageView) HandleMessageRemove(msg *discordgo.Message) {
	mv.HandleMessageCreate(msg)
}

func (mv *MessageView) BuildTexts() {
	// sizex, sizey := termbox.Size()
	mv.MessageContainer.ClearChildren()
	mv.MessageTexts = make([]*ui.Text, 0)

	rect := mv.Transform.GetRect()

	realScroll := mv.ScrollAmount
	y := int(rect.H) + realScroll
	padding := 0

	now := time.Now()
	thisYear, thisMonth, thisDay := now.Date()

	isFirst := true

	// Build it!!
	for _, item := range mv.DisplayMessages {
		if item == nil {
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

			ts := ""
			thenYear, thenMonth, thenDay := item.Timestamp.Date()
			if thisYear == thenYear && thisMonth == thenMonth && thisDay == thenDay {
				ts = item.Timestamp.Local().Format("15:04:05")
			} else {
				ts = item.Timestamp.Local().Format(time.Stamp)
			}
			ts += " "
			tsLen := utf8.RuneCountInString(ts)

			authorLen := utf8.RuneCountInString(author)
			channel, err := mv.App.session.State.Channel(msg.ChannelID)
			channelName := "???"
			isPrivate := false
			if err != nil {
				log.Println("Error getting channel", err)
			} else {
				channelName = channel.Name
				isPrivate = channel.IsPrivate
				if !isPrivate {
					guild, err := mv.App.session.State.Guild(channel.GuildID)
					if err == nil {
						channelName = guild.Name + "/#" + channelName
					}
				}
			}

			if isPrivate {
				channelName = "Direct Message"
			}

			fullMsg := ts + "[" + channelName + "]" + author + ": " + msg.ContentWithMentionsReplaced()
			channelLen := utf8.RuneCountInString(channelName) + 2
			points := map[int]ui.AttribPair{
				0:                              ui.AttribPair{termbox.ColorBlue, termbox.ColorDefault},
				tsLen:                          ui.AttribPair{termbox.ColorGreen, termbox.ColorDefault},
				channelLen + tsLen:             ui.AttribPair{termbox.ColorCyan | termbox.AttrBold, termbox.ColorDefault},
				channelLen + authorLen + tsLen: ui.AttribPair{},
			}
			if isPrivate {
				points[tsLen] = ui.AttribPair{termbox.ColorMagenta, termbox.ColorDefault}
			}
			text.Text = fullMsg
			text.Attribs = points
		}

		lines := text.HeightRequired()
		//lines := ui.HeightRequired(utf8.RuneCountInString(text.Text), int(rect.W)-padding*2)
		y -= lines
		if y < 0 {
			if y+lines > 0 {
				toSkip := -y
				text.SkipLines = toSkip
			} else {
				break
			}
		} else if y > int(rect.H) {
			continue
		}

		// Send ack
		if !item.IsLogMessage {
			mv.App.ackRoutine.In <- item.DiscordMessage
		}

		if mv.ScrollAmount != 0 && isFirst {
			if item.IsLogMessage {
				text.BG = termbox.ColorBlue | termbox.AttrBold
			} else {
				for k, v := range text.Attribs {
					text.Attribs[k] = ui.AttribPair{v.Fg, termbox.ColorBlue | termbox.AttrBold}
				}
			}
			isFirst = false
		}

		text.Transform.Position = common.NewVector2I(int(rect.X)+padding, int(rect.Y)+y)
		text.Layer = mv.Layer
		text.Userdata = item
		mv.MessageTexts = append(mv.MessageTexts, text)
		mv.MessageContainer.AddChild(text)
		if y < 0 {
			break
		}
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

			newest, nextIndex := mv.GetNewestMessageBefore(channel, beforeTime, listeningIndexes[k])

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

				newest, nextIndex := mv.GetNewestMessageBefore(privateChannel, beforeTime, pmIndexes[k])

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

	if mv.ScrollAmount == 0 {
		mv.ScrollText.Disabled = true
	} else {
		mv.ScrollText.Disabled = false
		mv.ScrollText.Text = fmt.Sprintf("Scroll: %d", mv.ScrollAmount)
	}
	if mv.DisplayMessagesDirty {
		mv.BuildDisplayMessages(h + mv.ScrollAmount)
		mv.BuildTexts()
		mv.DisplayMessagesDirty = false
		mv.TextsDirty = false
	} else if mv.TextsDirty {
		mv.BuildTexts()
		mv.TextsDirty = false
	}
}

func (mv *MessageView) GetDrawLayer() int {
	return 9
}

func (mv *MessageView) GetNewestMessageBefore(channel *discordgo.Channel, before time.Time, startIndex int) (*DisplayMessage, int) {
	msgs := channel.Messages
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

	if len(msgs) > 0 {
		name := GetChannelNameOrRecipient(channel)
		oldest := msgs[0]
		if !mv.App.IsFirstChannelMessage(channel.ID, oldest.ID) {
			// Grab history
			current, _ := mv.App.fetchingHistory[channel.ID]
			if !current {
				if *flagDebugEnabled {
					log.Println("Should grab history for ", name)
				}
				mv.App.fetchingHistory[channel.ID] = true
				go mv.App.GetHistory(channel.ID, 10, oldest.ID, "", true)
			}
		}
	}

	return nil, 0
}
