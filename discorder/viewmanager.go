package discorder

import (
	"fmt"
	"github.com/jonas747/discorder/common"
	"github.com/jonas747/discorder/ui"
	"github.com/jonas747/discordgo"
	"github.com/jonas747/go-runewidth"
	"log"
	"sort"
)

type ViewManager struct {
	*ui.BaseEntity
	App *App

	rootContainer *ui.AutoLayoutContainer

	middleContainer       *ui.Container
	middleLayoutContainer *ui.AutoLayoutContainer
	menuContainer         *ui.Container

	Tabs      TabSlice
	ActiveTab *Tab

	UIManager *ui.Manager

	inputHelper *ui.Text
	MainInput   *ui.TextInput
	debugText   *ui.Text
	header      *ui.Text

	mentionAutocompleter *MentionAutoCompletion
	notificationsManager *NotificationsManager
	typingDisplay        *TypingDisplay

	tabContainer *ui.AutoLayoutContainer

	readyReceived bool
}

func NewViewManager(app *App) *ViewManager {
	vm := &ViewManager{
		BaseEntity: &ui.BaseEntity{},
		App:        app,
		UIManager:  ui.NewManager(),
		Tabs:       make([]*Tab, 0),
	}
	vm.Transform.AnchorMax = common.NewVector2I(1, 1)
	return vm
}

func (v *ViewManager) OnInit() {
	rootContainer := ui.NewAutoLayoutContainer()
	rootContainer.Transform.AnchorMax = common.NewVector2F(1, 1)
	rootContainer.LayoutType = ui.LayoutTypeVertical
	rootContainer.ForceExpandWidth = true
	rootContainer.LayoutDynamic = true

	v.Transform.AddChildren(rootContainer)
	v.rootContainer = rootContainer

	// Add the header
	header := ui.NewText()
	header.Text = "Discorder v" + VERSION + "(´ ▽ ` )ﾉ"
	header.Transform.AnchorMin = common.NewVector2F(0.5, 0)
	header.Transform.AnchorMax = common.NewVector2F(0.5, 0)
	header.Transform.Position.X = float32(-runewidth.StringWidth(header.Text) / 2)

	rootContainer.Transform.AddChildren(header)
	v.header = header

	if v.App.options.DebugEnabled {
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
	v.middleLayoutContainer.LayoutDynamic = true

	v.rootContainer.Transform.AddChildren(v.middleLayoutContainer)

	// Menu container
	v.menuContainer = ui.NewContainer()
	v.menuContainer.AllowZeroSize = true
	v.menuContainer.Dynamic = false
	v.middleLayoutContainer.Transform.AddChildren(v.menuContainer)

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
	length := runewidth.StringWidth(inputHelper.Text)
	v.MainInput.Transform.Left = length + 1

	// Mention autocompleter
	v.mentionAutocompleter = NewMentionAutoCompletion(v.App, MainInput)
	v.rootContainer.Transform.AddChildren(v.mentionAutocompleter)

	// Tab container
	v.tabContainer = ui.NewAutoLayoutContainer()
	v.tabContainer.Transform.Size.Y = 1
	v.tabContainer.LayoutType = ui.LayoutTypeHorizontal
	v.tabContainer.ForceExpandHeight = true
	v.rootContainer.Transform.AddChildren(v.tabContainer)

	// Initialize tabs
	v.InitializeTabs()
	//v.UpdateTabIndicators()
	v.ApplyTheme()

	// Launch the login
	login := OpenLoginWindow(v.App)
	v.AddWindow(login)
	v.App.CheckAutoLogin()
}

func (v *ViewManager) OnReady() {

	// go into the main view
	if v.readyReceived {
		return // Only run once, not on reconnects
	}
	v.readyReceived = true
}

func (v *ViewManager) Destroy() { v.DestroyChildren() }

func (v *ViewManager) Update() {
	// Update the prompt
	talkingChannel := ""
	if v.ActiveTab != nil {
		talkingChannel = v.ActiveTab.SendChannel
	}
	if v.readyReceived {
		preStr := "Send to "

		channel, err := v.App.session.State.Channel(talkingChannel)
		name := talkingChannel

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
		} else {
			name = "Select a channel! >:O"
		}

		v.inputHelper.Text = preStr + "#" + name + ":"
		length := runewidth.StringWidth(v.inputHelper.Text)
		v.inputHelper.Transform.Size.X = float32(length)
		v.MainInput.Transform.Left = length
	}

	if v.App.options.DebugEnabled {
		children := v.App.Children(true)
		v.debugText.Text = fmt.Sprintf("Number of entities %d, Req queue length: %d", len(children), v.App.requestRoutine.GetQueueLenth())
	}

	if v.MainInput != nil && v.MainInput.TextBuffer != "" {
		v.App.typingRoutine.selfTypingIn <- talkingChannel
	}
}

func (v *ViewManager) SendFromTextBuffer() {
	if v.ActiveTab == nil || v.ActiveTab.SendChannel == "" {
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
			_, err := v.App.session.ChannelMessageSend(v.ActiveTab.SendChannel, toSend)
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

func (v *ViewManager) RemoveAllWindows() {
	windows := v.UIManager.Windows
	for _, window := range windows {
		_, ok := window.(*MessageView)
		if ok {
			continue
		}

		v.RemoveWindow(window)
	}
}

func (v *ViewManager) InitializeTabs() {
	tabConfig := v.App.config.Tabs
	if tabConfig == nil || len(tabConfig) < 1 {
		v.CreateTab(1)
		return
	}

	for _, t := range v.App.config.Tabs {
		v.CreateTab(t.Index)
		for _, c := range t.ListeningChannels {
			v.ActiveTab.MessageView.AddChannel(c)
		}
		v.ActiveTab.SendChannel = t.SendChannel
		v.ActiveTab.MessageView.ShowAllPrivate = t.AllPrivateMode
		v.ActiveTab.SetName(t.Name)
	}
}

func (v *ViewManager) CreateTab(index int) {
	for _, t := range v.Tabs {
		if t.Index == index {
			log.Println("Trying to create an existing tab")
			return
		}
	}
	tab := NewTab(v.App, index)

	v.Tabs = append(v.Tabs, tab)
	v.SetActiveTab(tab)
	v.UpdateTabIndicators()
}

func (v *ViewManager) RemoveTab(t *Tab) {
	for k, ct := range v.Tabs {
		if ct == t {
			ct.Destroy()
			v.Tabs = append(v.Tabs[:k], v.Tabs[k+1:]...)
			v.UpdateTabIndicators()
			break
		}
	}
}

func (v *ViewManager) SetActiveTab(t *Tab) {
	if t == v.ActiveTab {
		return
	}

	if v.ActiveTab != nil {
		v.ActiveTab.Transform.Parent.RemoveChild(v.ActiveTab, false)
		v.ActiveTab.SetActive(false)
		v.UIManager.RemoveWindow(v.ActiveTab.MessageView)
		if len(v.ActiveTab.MessageView.Channels) < 1 && !v.ActiveTab.MessageView.ShowAllPrivate {
			// Remove it
			v.RemoveTab(v.ActiveTab)
		}
	}

	v.middleLayoutContainer.Transform.AddChildren(t)
	v.UIManager.AddWindowFront(t.MessageView)
	t.SetActive(true)
	v.ActiveTab = t
}

func (v *ViewManager) UpdateTabIndicators() {
	v.tabContainer.Transform.ClearChildren(false)
	if len(v.Tabs) > 1 {
		sort.Sort(v.Tabs)
		for _, tab := range v.Tabs {
			v.tabContainer.Transform.AddChildren(tab.Indicator)
		}
		v.tabContainer.Transform.Size.Y = 1
	} else {
		v.tabContainer.Transform.Size.Y = 0
	}
}
func (v *ViewManager) HandleMessageCreate(m *discordgo.Message) {
	mentioned := false
	isPrivate := false
	channel, err := v.App.session.State.Channel(m.ChannelID)
	if err == nil {
		if channel.IsPrivate {
			mentioned = true
			isPrivate = true
		}
	}
	if !mentioned {
		for _, mention := range m.Mentions {
			if mention.ID == v.App.session.State.User.ID {
				mentioned = true
				break
			}
		}
	}

	for _, tab := range v.Tabs {
		if tab.MessageView == nil || tab.IndicatorMarked {
			continue
		}
		if isPrivate && tab.MessageView.ShowAllPrivate {
			v.App.ApplyThemeToText(tab.Indicator, "tab_mention")
			tab.IndicatorMarked = true
			continue
		}

		for _, c := range tab.MessageView.Channels {
			if c == m.ChannelID {
				if !tab.Active {
					if mentioned {
						v.App.ApplyThemeToText(tab.Indicator, "tab_mention")
						tab.IndicatorMarked = true
					} else {
						v.App.ApplyThemeToText(tab.Indicator, "tab_activity")
					}
				}
				break
			}
		}
	}
}
