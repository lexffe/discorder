package discorder

import (
	"fmt"
	"github.com/jonas747/discorder/common"
	"github.com/jonas747/discorder/ui"
	"github.com/nsf/termbox-go"
	"log"
	"time"
	"unicode/utf8"
)

type ViewManager struct {
	*ui.BaseEntity
	App *App

	mainContainer *ui.AutoLayoutContainer

	mv                  *MessageView // Will be changed when multiple message views
	SelectedMessageView *MessageView

	UIManager *ui.Manager

	inputHelper *ui.Text
	MainInput   *ui.TextInput
	debugText   *ui.Text
	header      *ui.Text

	mentionAutocompleter *MentionAutoCompletion
	notificationsManager *NotificationsManager
	typingDisplay        *TypingDisplay

	readyReceived  bool
	talkingChannel string
	lastLog        time.Time
}

func NewViewManager(app *App) *ViewManager {
	vm := &ViewManager{
		BaseEntity: &ui.BaseEntity{},
		App:        app,
		UIManager:  ui.NewManager(),
	}
	vm.Transform.AnchorMax = common.NewVector2I(1, 1)
	return vm
}

func (v *ViewManager) OnInit() {
	mainContainer := ui.NewAutoLayoutContainer()
	mainContainer.Transform.AnchorMax = common.NewVector2F(1, 1)
	mainContainer.LayoutType = ui.LayoutTypeVertical
	mainContainer.ForceExpandWidth = true

	v.Transform.AddChildren(mainContainer)
	v.mainContainer = mainContainer

	// Add the header
	header := ui.NewText()
	header.Text = "Discorder v" + VERSION + "(´ ▽ ` )ﾉ"
	header.Transform.AnchorMin = common.NewVector2F(0.5, 0)
	header.Transform.AnchorMax = common.NewVector2F(0.5, 0)
	header.Transform.Position.X = float32(-utf8.RuneCountInString(header.Text))

	mainContainer.Transform.AddChildren(header)
	v.header = header

	if v.App.debug {
		debugBar := ui.NewText()
		debugBar.Text = "debug"
		debugBar.Layer = 9

		mainContainer.Transform.AddChildren(debugBar)
		v.debugText = debugBar
	}

	// Launch the login
	login := NewLoginWindow(v.App)

	v.App.Transform.AddChildren(login)
	login.CheckAutoLogin()
}

func (v *ViewManager) OnReady() {
	// go into the main view
	v.readyReceived = true

	v.notificationsManager = NewNotificationsManager(v.App)
	v.mainContainer.Transform.AddChildren(v.notificationsManager)

	// Initialize all the ui entities
	mv := NewMessageView(v.App)
	v.mainContainer.Transform.AddChildren(mv)
	v.mv = mv
	v.SelectedMessageView = mv

	typingDisplay := NewTypingDisplay(v.App)
	typingDisplay.text.Layer = 9
	v.mainContainer.Transform.AddChildren(typingDisplay)
	v.typingDisplay = typingDisplay

	footerContainer := ui.NewContainer()
	footerContainer.AllowZeroSize = false
	v.mainContainer.Transform.AddChildren(footerContainer)

	MainInput := ui.NewTextInput(v.UIManager)
	MainInput.Transform.AnchorMax = common.NewVector2F(1, 1)
	MainInput.Layer = 5
	MainInput.SetActive(true)

	footerContainer.Transform.AddChildren(MainInput)
	v.MainInput = MainInput
	footerContainer.ProxySize = MainInput

	inputHelper := ui.NewText()
	inputHelper.Transform.AnchorMax = common.NewVector2I(1, 1)
	inputHelper.Layer = 5
	v.inputHelper = inputHelper
	footerContainer.Transform.AddChildren(inputHelper)

	inputHelper.Text = "Select a channel to send to"
	length := utf8.RuneCountInString(inputHelper.Text)
	v.MainInput.Transform.Left = length + 1

	v.mentionAutocompleter = NewMentionAutoCompletion(v.App, MainInput)
	v.mainContainer.Transform.AddChildren(v.mentionAutocompleter)

	v.ApplyConfig()
	v.ApplyTheme()
}

func (v *ViewManager) ApplyConfig() {
	for _, channel := range v.App.config.ListeningChannels {
		v.SelectedMessageView.AddChannel(channel)
	}
	v.talkingChannel = v.App.config.LastChannel
	v.SelectedMessageView.ShowAllPrivate = v.App.config.AllPrivateMode
}

func (v *ViewManager) Destroy() { v.DestroyChildren() }

func (v *ViewManager) Update() {
	if v.mv != nil {
		if logRoutine.HasChangedSince(v.lastLog) {
			v.mv.Logs = logRoutine.GetCopy()
			v.mv.DisplayMessagesDirty = true
		}
	}

	// Update the prompt
	if v.talkingChannel != "" {
		preStr := "Send to "

		channel, err := v.App.session.State.Channel(v.talkingChannel)
		name := v.talkingChannel

		if channel != nil && err == nil {
			name = GetChannelNameOrRecipient(channel)
			if !channel.IsPrivate {
				guild, err := v.App.session.State.Guild(channel.GuildID)
				if err == nil {
					preStr += guild.Name + "/"
				} else {
					preStr += channel.GuildID + "/"
				}
			}
		}

		v.inputHelper.Text = preStr + "#" + name + ":"
		length := utf8.RuneCountInString(v.inputHelper.Text)
		v.inputHelper.Transform.Size.X = float32(length)
		v.MainInput.Transform.Left = length
	}

	if v.App.debug {
		children := v.App.Children(true)
		v.debugText.Text = fmt.Sprintf("Number of entities %d, Req queue length: %d", len(children), v.App.requestRoutine.GetQueueLenth())
	}

	if v.MainInput != nil && v.MainInput.TextBuffer != "" {
		v.App.typingRoutine.selfTypingIn <- v.talkingChannel
	}
}

func (v *ViewManager) HandleInput(event termbox.Event) {
	if !v.readyReceived {
		return
	}

	if event.Type == termbox.EventKey {
		switch event.Key {
		// case termbox.KeyCtrlG: // Select channel
		// 	if v.activeWindow != nil {
		// 		break
		// 	}
		// case termbox.KeyCtrlO: // Options
		// 	if v.activeWindow != nil {
		// 		break
		// 	}
		// 	hw := NewHelpWindow(v.App)
		// 	v.AddChild(hw)
		// 	v.activeWindow = hw
		// 	v.MainInput.Active = false
		// case termbox.KeyCtrlS: // Select server
		// 	if v.activeWindow != nil {
		// 		break
		// 	}
		// 	ssw := NewSelectServerWindow(v.App, v.mv)
		// 	v.SetActiveWindow(ssw)
		case termbox.KeyEnter:
			if v.talkingChannel == "" {
				log.Println("you're trying to send a message to nobody buddy D:")
				break
			}

			if v.MainInput.TextBuffer == "" {
				break // Nothing to see here...
			}

			if v.mentionAutocompleter.isAutocompletingMention {
				if v.mentionAutocompleter.PerformAutocompleteMention() {
					v.mentionAutocompleter.isAutocompletingMention = false
				}
			} else {
				toSend := v.MainInput.TextBuffer
				v.MainInput.TextBuffer = ""
				v.MainInput.CursorLocation = 0
				go func() {
					_, err := v.App.session.ChannelMessageSend(v.talkingChannel, toSend)
					if err != nil {
						log.Println("Error sending message: ", err)
					}
				}()
			}
		}
	}
}

func (v *ViewManager) CanOpenWindow() bool {
	return v.readyReceived
}

func (v *ViewManager) ApplyTheme() {
	v.App.ApplyThemeToText(v.inputHelper, "send_prompt")
	v.App.ApplyThemeToText(v.MainInput.Text, "input_chat")
	v.App.ApplyThemeToText(v.header, "title_bar")
	v.App.ApplyThemeToText(v.typingDisplay.text, "typing_bar")
	v.App.ApplyThemeToText(v.notificationsManager.text, "notifications_bar")

	ui.RunFuncConditional(v, func(e ui.Entity) bool {
		menu, ok := e.(*ui.MenuWindow)
		if ok {
			v.App.ApplyThemeToMenu(menu)
			return false
		}

		window, ok := e.(*ui.Window)
		if ok {
			v.App.ApplyThemeToWindow(window)
			return false
		}
		return true
	})
}
