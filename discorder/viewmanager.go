package discorder

import (
	"fmt"
	"github.com/jonas747/discorder/common"
	"github.com/jonas747/discorder/ui"
	"log"
	"time"
	"unicode/utf8"
)

type ViewManager struct {
	*ui.BaseEntity
	App *App

	rootContainer *ui.AutoLayoutContainer

	middleContainer       *ui.Container
	middleLayoutContainer *ui.AutoLayoutContainer
	menuContainer         *ui.Container

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
	rootContainer := ui.NewAutoLayoutContainer()
	rootContainer.Transform.AnchorMax = common.NewVector2F(1, 1)
	rootContainer.LayoutType = ui.LayoutTypeVertical
	rootContainer.ForceExpandWidth = true

	v.Transform.AddChildren(rootContainer)
	v.rootContainer = rootContainer

	// Add the header
	header := ui.NewText()
	header.Text = "Discorder v" + VERSION + "(´ ▽ ` )ﾉ"
	header.Transform.AnchorMin = common.NewVector2F(0.5, 0)
	header.Transform.AnchorMax = common.NewVector2F(0.5, 0)
	header.Transform.Position.X = float32(-utf8.RuneCountInString(header.Text) / 2)

	rootContainer.Transform.AddChildren(header)
	v.header = header

	if v.App.debug {
		debugBar := ui.NewText()
		debugBar.Text = "debug"
		debugBar.Layer = 9

		rootContainer.Transform.AddChildren(debugBar)
		v.debugText = debugBar
	}

	v.notificationsManager = NewNotificationsManager(v.App)
	v.rootContainer.Transform.AddChildren(v.notificationsManager)

	v.middleLayoutContainer = ui.NewAutoLayoutContainer()
	v.middleLayoutContainer.Transform.AnchorMax = common.NewVector2F(1, 1)
	v.middleLayoutContainer.LayoutType = ui.LayoutTypeHorizontal
	v.middleLayoutContainer.ForceExpandHeight = true

	v.rootContainer.Transform.AddChildren(v.middleLayoutContainer)

	// Menu container
	v.menuContainer = ui.NewContainer()
	v.menuContainer.AllowZeroSize = true
	v.menuContainer.Dynamic = false
	v.middleLayoutContainer.Transform.AddChildren(v.menuContainer)

	// Initialize all the ui entities
	mv := NewMessageView(v.App)
	v.middleLayoutContainer.Transform.AddChildren(mv)
	v.mv = mv
	v.SelectedMessageView = mv

	// Launch the login
	login := NewLoginWindow(v.App)
	v.AddWindow(login)
	login.CheckAutoLogin()
}

func (v *ViewManager) OnReady() {
	// go into the main view
	v.readyReceived = true

	// Typing display
	typingDisplay := NewTypingDisplay(v.App)
	typingDisplay.text.Layer = 9
	v.rootContainer.Transform.AddChildren(typingDisplay)
	v.typingDisplay = typingDisplay

	// Footer
	footerContainer := ui.NewContainer()
	footerContainer.AllowZeroSize = false
	v.rootContainer.Transform.AddChildren(footerContainer)

	// Main input
	MainInput := ui.NewTextInput(v.UIManager, 5)
	MainInput.Transform.AnchorMax = common.NewVector2F(1, 1)
	MainInput.SetActive(true)

	footerContainer.Transform.AddChildren(MainInput)
	v.MainInput = MainInput
	footerContainer.ProxySize = MainInput

	// Prompt
	inputHelper := ui.NewText()
	inputHelper.Transform.AnchorMax = common.NewVector2I(1, 1)
	inputHelper.Layer = 5
	v.inputHelper = inputHelper
	footerContainer.Transform.AddChildren(inputHelper)

	inputHelper.Text = "Select a channel to send to"
	length := utf8.RuneCountInString(inputHelper.Text)
	v.MainInput.Transform.Left = length + 1

	// Mention autocompleter
	v.mentionAutocompleter = NewMentionAutoCompletion(v.App, MainInput)
	v.rootContainer.Transform.AddChildren(v.mentionAutocompleter)

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

func (v *ViewManager) SendFromTextBuffer() {
	if v.talkingChannel == "" {
		log.Println("you're trying to send a message to nobody buddy D:")
		return
	}

	if v.MainInput.TextBuffer == "" {
		return // Nothing to see here...
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

func (v *ViewManager) CanOpenWindow() bool {
	return v.readyReceived
}

func (v *ViewManager) ApplyTheme() {
	v.App.ApplyThemeToText(v.inputHelper, "send_prompt")
	v.App.ApplyThemeToText(v.MainInput.Text, "input_chat")
	v.App.ApplyThemeToText(v.header, "title_bar")
	v.App.ApplyThemeToText(v.typingDisplay.text, "typing_bar")
	v.App.ApplyThemeToText(v.notificationsManager.text, "notifications_bar")

	ui.RunFuncCondTraverse(v, func(e ui.Entity) bool {
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

func (v *ViewManager) AddWindow(e ui.Entity) {
	v.menuContainer.Transform.AddChildren(e)
	v.menuContainer.Dynamic = true
}

func (v *ViewManager) RemoveWindow(e ui.Entity) {
	v.menuContainer.Transform.RemoveChild(e, true)

	if len(v.menuContainer.Children(false)) > 0 {
		v.menuContainer.Dynamic = true
	} else {
		v.menuContainer.Dynamic = false
		v.menuContainer.Transform.Size = common.NewVector2I(0, 0)
	}

}
