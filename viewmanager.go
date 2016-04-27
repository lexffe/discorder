package main

import (
	"fmt"
	"github.com/jonas747/discorder/common"
	"github.com/jonas747/discorder/ui"
	"github.com/nsf/termbox-go"
	"log"
	"unicode/utf8"
)

type ViewManager struct {
	*ui.BaseEntity
	App                  *App
	mv                   *MessageView
	SelectedMessageView  *MessageView
	activeWindow         ui.Entity
	inputHelper          *ui.Text
	input                *ui.TextInput
	debugText            *ui.Text
	readyReceived        bool
	talkingChannel       string
	mentionAutocompleter *MentionAutoCompletion
	notificationsManager *NotificationsManager
}

func NewViewManager(app *App) *ViewManager {
	mv := &ViewManager{
		BaseEntity: &ui.BaseEntity{},
		App:        app,
	}
	return mv
}

func (v *ViewManager) OnInit() {
	// Add the header
	header := ui.NewText()
	header.Text = "Discorder v" + VERSION + "(´ ▽ ` )ﾉ"
	hw := utf8.RuneCountInString(header.Text)
	header.Transform.Size = common.NewVector2I(hw, 0)
	header.Transform.AnchorMin = common.NewVector2F(0.5, 0)
	header.Transform.AnchorMax = common.NewVector2F(0.5, 0)
	header.Transform.Position.X = float32(-(hw / 2))
	v.AddChild(header)

	if *flagDebugEnabled {
		debugBar := ui.NewText()
		debugBar.Text = "debug"
		debugBar.Transform.AnchorMin = common.NewVector2F(0, 0)
		debugBar.Transform.AnchorMax = common.NewVector2F(1, 0)
		debugBar.Transform.Position.Y = 2
		debugBar.Layer = 9
		v.AddChild(debugBar)
		v.debugText = debugBar
	}

	// Launch the login
	login := NewLoginWindow(v.App)
	v.App.AddChild(login)
	login.CheckAutoLogin()
}

func (v *ViewManager) OnReady() {
	// go into the main view
	v.readyReceived = true

	mv := NewMessageView(v.App)
	mv.Transform.AnchorMax = common.NewVector2I(1, 1)
	mv.Transform.Bottom = 3
	mv.Transform.Top = 2
	mv.ShowAllPrivate = true
	mv.Logs = v.App.logBuffer
	v.AddChild(mv)
	v.mv = mv
	v.SelectedMessageView = mv
	if *flagDebugEnabled {
		mv.Transform.Top = 3
	}

	input := ui.NewTextInput()
	input.Transform.AnchorMin = common.NewVector2F(0, 1)
	input.Transform.AnchorMax = common.NewVector2F(1, 1)
	input.Transform.Position.Y = -1
	input.Active = true
	v.AddChild(input)
	v.input = input

	inputHelper := ui.NewText()
	inputHelper.Transform.AnchorMax = common.NewVector2I(0, 1)
	inputHelper.Transform.AnchorMin = common.NewVector2I(0, 1)
	inputHelper.FG = termbox.ColorYellow | termbox.AttrBold
	inputHelper.Text = "Select a channel to send to"
	length := utf8.RuneCountInString(inputHelper.Text)
	inputHelper.Transform.Size.X = float32(length)
	inputHelper.Transform.Position.Y = -1
	v.inputHelper = inputHelper
	v.AddChild(inputHelper)

	v.input.Transform.Left = length + 1

	v.mentionAutocompleter = NewMentionAutoCompletion(v.App, input)
	v.mentionAutocompleter.Transform.AnchorMin.Y = 1
	v.mentionAutocompleter.Transform.AnchorMax = common.NewVector2I(1, 1)
	v.mentionAutocompleter.Transform.Position.Y = -2
	v.AddChild(v.mentionAutocompleter)

	typingDisplay := NewTypingDisplay(v.App)
	typingDisplay.Transform.AnchorMin.Y = 1
	typingDisplay.Transform.AnchorMax = common.NewVector2I(1, 1)
	typingDisplay.Transform.Position.Y = -2
	v.AddChild(typingDisplay)

	v.notificationsManager = NewNotificationsManager(v.App)
	v.notificationsManager.Transform.AnchorMax.X = 1
	v.notificationsManager.Transform.Position.Y = 1
	v.AddChild(v.notificationsManager)

	v.ApplyConfig()

}

func (v *ViewManager) ApplyConfig() {
	for _, channel := range v.App.config.ListeningChannels {
		v.SelectedMessageView.AddChannel(channel)
	}
	v.talkingChannel = v.App.config.LastChannel
}

func (v *ViewManager) Destroy() { v.DestroyChildren() }

func (v *ViewManager) PreDraw() {
	if v.mv != nil {
		v.mv.Logs = v.App.logBuffer
	}

	// Update the prompt
	if v.talkingChannel != "" {
		preStr := "Send to"

		channel, err := v.App.session.State.Channel(v.talkingChannel)
		name := v.talkingChannel

		if channel != nil && err == nil {
			name = channel.Name
			if !channel.IsPrivate {
				guild, err := v.App.session.State.Guild(channel.GuildID)
				if err == nil {
					preStr += " " + guild.Name + "/"
				} else {
					preStr += " " + channel.GuildID + "/"
				}
			}
		}

		v.inputHelper.Text = preStr + "#" + name
		length := utf8.RuneCountInString(v.inputHelper.Text)
		v.inputHelper.Transform.Size.X = float32(length)
		v.input.Transform.Left = length + 1
	}

	if *flagDebugEnabled {
		children := v.App.Children(true)
		v.debugText.Text = fmt.Sprintf("Number of entities %d", len(children))
	}
}

func (v *ViewManager) HandleInput(event termbox.Event) {
	if !v.readyReceived {
		return
	}

	if event.Type == termbox.EventKey {
		switch event.Key {
		case termbox.KeyCtrlG: // Select channel
			if v.activeWindow != nil {
				break
			}
		case termbox.KeyCtrlO: // Options
			if v.activeWindow != nil {
				break
			}
			hw := NewHelpWindow(v.App)
			v.AddChild(hw)
			v.activeWindow = hw
			log.Println("Opening help")
			v.input.Active = false
		case termbox.KeyCtrlS: // Select server
			if v.activeWindow != nil {
				break
			}
			ssw := NewSelectServerWindow(v.App, v.mv)
			v.AddChild(ssw)
			v.activeWindow = ssw
			v.input.Active = false
			log.Println("Opening server select window")
		case termbox.KeyBackspace, termbox.KeyBackspace2:
			if v.activeWindow != nil {
				v.RemoveChild(v.activeWindow, true)
				v.activeWindow = nil
				v.input.Active = true
			}
		case termbox.KeyCtrlL:
			v.App.logBuffer = []*common.LogMessage{}
		case termbox.KeyEnter:
			if v.talkingChannel == "" {
				log.Println("you're trying to send a message to nobody buddy D:")
				break
			}

			if v.mentionAutocompleter.isAutocompletingMention {
				if v.mentionAutocompleter.PerformAutocompleteMention() {
					v.mentionAutocompleter.isAutocompletingMention = false
				}
			} else {
				toSend := v.input.TextBuffer
				v.input.TextBuffer = ""
				v.input.CursorLocation = 0
				v.App.session.ChannelMessageSend(v.talkingChannel, toSend)
			}

		}
	}
}
