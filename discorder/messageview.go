package discorder

import (
	"fmt"
	"github.com/jonas747/discorder/common"
	"github.com/jonas747/discorder/ui"
	"github.com/jonas747/discordgo"
	"github.com/jonas747/termbox-go"
	"log"
	"strings"
	"time"
	"unicode/utf8"
)

type MessageView struct {
	*ui.BaseEntity
	App             *App
	DisplayMessages []*DisplayMessage

	Channels       []string
	ShowAllPrivate bool
	Logs           []*LogMessage // Maybe move this?

	Layer int

	MessageContainer *ui.SimpleEntity
	MessageTexts     []*ui.Text
	ScrollText       *ui.Text

	ScrollAmount int

	DisplayMessagesDirty bool // Rebuilds displaymessages on next draw if set
	TextsDirty           bool // Rebuilds texts on next draw if set
	lastLogs             time.Time
	lastRect             common.Rect
}

type DisplayMessage struct {
	DiscordMessage *discordgo.Message
	LogMessage     *LogMessage
	IsLogMessage   bool
	Timestamp      time.Time
}

func NewMessageView(app *App) *MessageView {
	mv := &MessageView{
		BaseEntity:       &ui.BaseEntity{},
		App:              app,
		MessageContainer: ui.NewSimpleEntity(),
	}

	t := ui.NewText()
	t.Transform.AnchorMin.Y = 1
	t.Transform.AnchorMax = common.NewVector2I(1, 1)
	t.Transform.Position.Y = -1
	t.Layer = 6

	app.ApplyThemeToText(t, "scroll_text")

	mv.Transform.AddChildren(t)
	mv.ScrollText = t

	mv.Transform.AddChildren(mv.MessageContainer)

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

	// Grab some history if needed
	if mv.App.session == nil || mv.App.session.State == nil {
		mv.App.requestRoutine.AddRequest(NewHistoryRequest(mv.App, channel, 20, "", ""))
	} else {

		discordChannel, err := mv.App.session.State.Channel(channel)
		if err != nil {
			return
		}
		if len(discordChannel.Messages) < 10 {
			firstId := ""
			if len(discordChannel.Messages) > 0 {
				firstId = discordChannel.Messages[0].ID
			}

			if !mv.App.IsFirstChannelMessage(discordChannel.ID, firstId) {
				mv.App.requestRoutine.AddRequest(NewHistoryRequest(mv.App, channel, 10, "", ""))
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

func (mv *MessageView) Select() {
	if mv.ScrollAmount <= 0 {
		mv.App.ViewManager.SendFromTextBuffer()
		return
	}

	if len(mv.MessageTexts) < 0 {
		return
	}

	text := mv.MessageTexts[0]
	selectedDisplayMsg, ok := text.Userdata.(*DisplayMessage)
	if !ok || selectedDisplayMsg.IsLogMessage {
		return
	}

	msg := selectedDisplayMsg.DiscordMessage
	presetArgs := map[string]interface{}{"channel": msg.ChannelID, "message": msg.ID, "user": msg.Author.ID}

	info := fmt.Sprintf("%s\nMessage ID: %s\nAuthor: %s (ID: %s)", msg.ContentWithMentionsReplaced(), msg.ID, msg.Author.Username, msg.Author.ID)

	channel, err := mv.App.session.State.Channel(msg.ChannelID)
	if err != nil {
		log.Printf("Failed getting channel from state", err)
		return
	}

	if !channel.IsPrivate {
		presetArgs["server"] = channel.GuildID
	}

	info += fmt.Sprintf("\nChannel: #%s (ID: %s)", channel.Name, channel.ID)

	guild, _ := mv.App.session.State.Guild(channel.GuildID)

	if guild != nil {
		info += fmt.Sprintf("\nServer: %s (ID: %s)", guild.Name, guild.ID)
	}

	cw := NewCommandWindow(mv.App, 5, presetArgs, info)
	extraCommands := mv.App.GenMessageCommands(msg)
	commands := append(mv.App.Commands, extraCommands...)
	cw.commands = commands

	mv.App.ViewManager.AddWindow(cw)
}

func (mv *MessageView) Scroll(dir ui.Direction, amount int) {
	switch dir {
	case ui.DirUp:
		mv.ScrollAmount += amount
		mv.DisplayMessagesDirty = true
	case ui.DirDown:
		mv.ScrollAmount -= amount
		if mv.ScrollAmount < 0 {
			mv.ScrollAmount = 0
		}
		mv.DisplayMessagesDirty = true
	case ui.DirEnd, ui.DirStart:
		mv.ScrollAmount = 0
		mv.DisplayMessagesDirty = true
	}

	if mv.ScrollAmount > 0 {
		mv.ScrollText.Text = fmt.Sprintf("Scroll: %d", mv.ScrollAmount)
		mv.ScrollText.Disabled = false
	} else {
		mv.ScrollText.Disabled = true
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
	mv.MessageContainer.Transform.ClearChildren(true)
	mv.MessageTexts = make([]*ui.Text, 0)

	rect := mv.Transform.GetRect()

	realScroll := mv.ScrollAmount - 1
	if realScroll == -1 {
		realScroll = 0
	}
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

		attribs := make(map[int]ui.AttribPair)

		if item.IsLogMessage {
			text.Text = "Log: " + item.LogMessage.Content
			attribs[0] = mv.App.GetThemeAttribPair("message_log").AttribPair()
		} else {
			msg := item.DiscordMessage
			if msg == nil {
				continue
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

			isPrivate := false
			channelName := "???"
			channel, err := mv.App.session.State.Channel(msg.ChannelID)
			var guild *discordgo.Guild
			if err == nil {
				channelName = channel.Name
				isPrivate = channel.IsPrivate
				if !isPrivate {
					guild, err = mv.App.session.State.Guild(channel.GuildID)
					if err == nil {
						guildName := guild.Name
						if mv.App.config.ShortGuilds {
							guildName = ShortName(guildName)
						}
						channelName = guildName + "/#" + channelName
					}
				}
			}

			author := "Unknown?"
			if msg.Author != nil {
				if mv.App.config.HideNicknames || guild == nil {
					author = msg.Author.Username
				} else {
					member, err := mv.App.session.State.Member(guild.ID, msg.Author.ID)
					if (err == nil && member.Nick == "") || err != nil {
						author = msg.Author.Username // Fallback
					} else {
						author = member.Nick
					}
				}
			}

			authorLen := utf8.RuneCountInString(author)

			if isPrivate {
				channelName = "Direct Message"
			}

			body := msg.ContentWithMentionsReplaced()
			for _, v := range msg.Attachments {
				body += " Attachment: " + v.ProxyURL + " (original: " + v.URL + ") "
			}

			fullMsg := ts + "[" + channelName + "]" + author + ": " + body
			channelLen := utf8.RuneCountInString(channelName) + 2
			attribs = map[int]ui.AttribPair{
				0:                              mv.App.GetThemeAttribPair("message_timestamp").AttribPair(),
				tsLen:                          mv.App.GetThemeAttribPair("message_server_channel").AttribPair(),
				channelLen + tsLen:             mv.App.GetThemeAttribPair("message_author").AttribPair(),
				channelLen + authorLen + tsLen: mv.App.GetThemeAttribPair("message_content").AttribPair(),
			}
			if isPrivate {
				attribs[tsLen] = mv.App.GetThemeAttribPair("message_direct_channel").AttribPair()
			}
			text.Text = fullMsg
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
		} else if y >= int(rect.H) {
			continue
		}

		// Send ack
		if !item.IsLogMessage && !mv.App.stopping {
			msgCopy := item.DiscordMessage
			go func() {
				mv.App.ackRoutine.In <- msgCopy
			}()
		}

		if mv.ScrollAmount != 0 && isFirst {
			if item.IsLogMessage {
				mv.App.ApplyThemeToText(text, "element_selected")
			} else {
				for k, v := range attribs {
					attribs[k] = ui.AttribPair{v.FG, termbox.ColorBlue | termbox.AttrBold}
				}
			}
			isFirst = false
		}

		text.SetAttribs(attribs)
		text.Transform.Position = common.NewVector2I(padding, y)
		text.Layer = mv.Layer
		text.Userdata = item
		mv.MessageTexts = append(mv.MessageTexts, text)
		mv.MessageContainer.Transform.AddChildren(text)
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
	var state *discordgo.State
	if mv.App.ViewManager.readyReceived {
		state = mv.App.session.State
	}
	if state == nil {
		state = discordgo.NewState()
	}
	state.RLock()
	defer state.RUnlock()

	messages := make([]*DisplayMessage, size)

	// Holds the start indexes in the newest message search
	listeningIndexes := make([]int, len(mv.Channels))
	var pmIndexes []int
	if mv.App.ViewManager.readyReceived {
		pmIndexes = make([]int, len(state.PrivateChannels))
	}

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
		if mv.App.ViewManager.readyReceived {
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

func (mv *MessageView) Update() {
	curRect := mv.Transform.GetRect()
	if !curRect.Equals(mv.lastRect) {
		mv.lastRect = curRect
		mv.DisplayMessagesDirty = true
	}

	if logRoutine.HasChangedSince(mv.lastLogs) {
		mv.Logs = logRoutine.GetCopy()
		mv.DisplayMessagesDirty = true
		mv.lastLogs = time.Now()
	}

	if mv.ScrollAmount == 0 {
		mv.ScrollText.Disabled = true
	} else {
		mv.ScrollText.Disabled = false
		mv.ScrollText.Text = fmt.Sprintf("Scroll: %d", mv.ScrollAmount)
	}
	if mv.DisplayMessagesDirty {
		h := int(curRect.H)
		if h < 0 {
			h = 0
		}
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
	return mv.Layer
}

func (mv *MessageView) GetNewestMessageBefore(channel *discordgo.Channel, before time.Time, startIndex int) (*DisplayMessage, int) {
	msgs := channel.Messages
	if startIndex == -10 {
		startIndex = len(msgs) - 1
	}
	for j := startIndex; j >= 0; j-- {
		msg := msgs[j]
		parsedTimestamp, _ := time.Parse(DiscordTimeFormat, msg.Timestamp)
		if !parsedTimestamp.After(before) || before.IsZero() { // Reason for !after is so that we still show all the messages with same timestamps
			curNewestMessage := &DisplayMessage{
				DiscordMessage: msg,
				Timestamp:      parsedTimestamp,
			}
			return curNewestMessage, j - 1
		}
	}

	if len(msgs) > 0 && !mv.App.stopping {
		oldest := msgs[0]
		if !mv.App.IsFirstChannelMessage(channel.ID, oldest.ID) {
			// Grab history
			mv.App.requestRoutine.AddRequest(NewHistoryRequest(mv.App, channel.ID, 25, oldest.ID, ""))
		}
	}

	return nil, 0
}

// Implement LayoutElement
func (mv *MessageView) GetRequiredSize() common.Vector2F {
	//log.Println("Called getreuidesize")
	return common.Vector2F{}
}

func (mv *MessageView) IsLayoutDynamic() bool {
	return true
}

func ShortName(in string) string {
	fields := strings.Fields(in)
	out := ""
	for _, field := range fields {
		firstRune := []rune(field)[0]
		out += string(firstRune)
	}
	return out
}
