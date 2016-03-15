package main

import (
	"container/list"
	"github.com/bwmarrin/discordgo"
	"github.com/nsf/termbox-go"
	"log"
	"time"
	"unicode/utf8"
)

const HistorySize = 1000

type State interface {
	HandleInput(event termbox.Event)
	RefreshDisplay()
}

type App struct {
	running        bool
	stopping       bool
	stopChan       chan chan error
	msgRecvChan    chan *discordgo.Message
	session        *discordgo.Session
	inputEventChan chan termbox.Event
	logChan        chan string

	selectedServerId  string
	selectedChannelId string

	selectedGuild   *discordgo.Guild
	selectedChannel *discordgo.Channel

	history *list.List

	currentState State

	stopPollEvents chan chan bool

	currentSendBuffer     string
	currentCursorLocation int
}

func Login(user, password string) (*App, error) {
	session, err := discordgo.New(user, password)
	if err != nil {
		return nil, err
	}

	session.StateEnabled = true

	app := &App{
		session: session,
		history: list.New(),
	}

	session.AddHandler(app.messageCreate)

	err = session.Open()
	return app, err
}

func (app *App) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	app.msgRecvChan <- m.Message
}

func (app *App) Stop() error {
	if app.stopping {
		return nil
	}
	app.stopping = true
	errChan := make(chan error)
	if app.running {
		app.stopChan <- errChan
		return <-errChan
	}
	return nil
}

// Lsiten on the channels for incoming messages
func (app *App) Run() {
	// Some initialization
	if app.running {
		log.Println("Tried to run app while already running")
		return
	}
	app.running = true

	app.msgRecvChan = make(chan *discordgo.Message)
	app.stopChan = make(chan chan error)
	app.inputEventChan = make(chan termbox.Event)
	app.stopPollEvents = make(chan chan bool)
	app.logChan = make(chan string)
	app.currentState = &StateNormal{app}
	log.SetOutput(app)
	log.Println("Starting...")
	// Start polling events
	go app.PollEvents()

	ticker := time.NewTicker(1000 * time.Millisecond)

	for {
		select {
		case errChan := <-app.stopChan:
			app.running = false
			pollStopped := make(chan bool)

			// Stop the event polling
			go delayedInterrupt(1 * time.Millisecond) // might not work 100% all cases? should probably replace
			app.stopPollEvents <- pollStopped
			<-pollStopped
			errChan <- nil
			return
		case msg := <-app.msgRecvChan:
			app.HandleMessage(*msg)
		case evt := <-app.inputEventChan:
			app.HandleInputEvent(evt)
		case msg := <-app.logChan:
			app.HandleMessage(msg)
		case <-ticker.C:
			var err error
			if app.selectedServerId != "" {
				app.selectedGuild, err = app.session.State.Guild(app.selectedServerId)
				if err != nil {
					log.Println("App.Run: ", err)
				}
			}
			if app.selectedChannelId != "" && app.selectedGuild != nil {
				app.selectedChannel, err = app.session.State.GuildChannel(app.selectedServerId, app.selectedChannelId)
				if err != nil {
					log.Println("App.Run: ", err)
				}
			}

			app.RefreshDisplay()
		}
	}
}

func delayedInterrupt(d time.Duration) {
	time.Sleep(d)
	termbox.Interrupt()
}

func (app *App) HandleMessage(msg interface{}) {
	cast, ok := msg.(discordgo.Message)
	if ok {
		chId := cast.ChannelID
		state := app.session.State
		// Check if its a private channel
		_, err := state.PrivateChannel(cast.ChannelID)
		if err == nil {
			app.history.PushFront(msg)
		}

		if app.selectedServerId != "" {
			_, err := app.session.State.GuildChannel(app.selectedServerId, chId)
			if err == nil {
				app.history.PushFront(msg)
			}
		}
	} else {
		// Let log messages through
		app.history.PushFront(msg)
	}

	for app.history.Len() > HistorySize {
		app.history.Remove(app.history.Back())
	}

	//app.RefreshDisplay()
}

func (app *App) HandleInputEvent(event termbox.Event) {
	if event.Type == termbox.EventKey {
		if event.Key == termbox.KeyEsc {
			log.Println("Stopping...")
			go app.Stop()
		}
	}

	app.currentState.HandleInput(event)
	app.RefreshDisplay()
}

func (app *App) PollEvents() {
	if !termbox.IsInit {
		err := termbox.Init()
		if err != nil {
			panic(err)
		}
	}

	for {
		evt := termbox.PollEvent()

		select {
		case retChan := <-app.stopPollEvents:
			if termbox.IsInit {
				termbox.Close()
			}
			retChan <- true
			return
		default:
			break
		}

		app.inputEventChan <- evt
	}
}

// For logs
func (app *App) Write(p []byte) (n int, err error) {
	// since we might log from the same goroutine deadlocks may occour, should probably do a queue system or something instead...
	go func() {
		app.logChan <- string(p)
	}()

	return len(p), nil
}

func (app *App) HandleTextInput(event termbox.Event) {
	if event.Type == termbox.EventKey {
		if event.Key == termbox.KeyEnter {
			// send
			cp := app.currentSendBuffer
			app.currentSendBuffer = ""
			app.currentCursorLocation = 0
			app.RefreshDisplay()
			_, err := app.session.ChannelMessageSend(app.selectedChannelId, cp)
			if err != nil {
				log.Println("Error sending: ", err)
			}
		} else if event.Key == termbox.KeyArrowLeft {
			app.currentCursorLocation--
			if app.currentCursorLocation < 0 {
				app.currentCursorLocation = 0
			}
		} else if event.Key == termbox.KeyArrowRight {
			app.currentCursorLocation++
			bufLen := utf8.RuneCountInString(app.currentSendBuffer)
			if app.currentCursorLocation > bufLen {
				app.currentCursorLocation = bufLen
			}
		} else if event.Key == termbox.KeyBackspace || event.Key == termbox.KeyBackspace2 {
			bufLen := utf8.RuneCountInString(app.currentSendBuffer)
			if bufLen == 0 {
				return
			}
			if app.currentCursorLocation == bufLen {
				_, size := utf8.DecodeLastRuneInString(app.currentSendBuffer)
				app.currentCursorLocation--
				app.currentSendBuffer = app.currentSendBuffer[:len(app.currentSendBuffer)-size]
			} else if app.currentCursorLocation == 1 {
				_, size := utf8.DecodeRuneInString(app.currentSendBuffer)
				app.currentCursorLocation--
				app.currentSendBuffer = app.currentSendBuffer[size:]
			} else if app.currentCursorLocation == 0 {
				return
			} else {
				runeSlice := []rune(app.currentSendBuffer)
				newSlice := append(runeSlice[:app.currentCursorLocation], runeSlice[app.currentCursorLocation+1:]...)
				app.currentSendBuffer = string(newSlice)
				app.currentCursorLocation--
			}
		} else if event.Ch != 0 || event.Key == termbox.KeySpace {
			char := event.Ch
			if event.Key == termbox.KeySpace {
				char = ' '
			}

			bufLen := utf8.RuneCountInString(app.currentSendBuffer)
			if app.currentCursorLocation == bufLen {
				app.currentSendBuffer += string(char)
				app.currentCursorLocation++
			} else if app.currentCursorLocation == 0 {
				app.currentSendBuffer = string(char) + app.currentSendBuffer
				app.currentCursorLocation++
			} else {
				bufSlice := []rune(app.currentSendBuffer)
				bufCopy := ""

				for i := 0; i < len(bufSlice); i++ {
					if i == app.currentCursorLocation {
						bufCopy += string(char)
					}
					bufCopy += string(bufSlice[i])
				}
				app.currentSendBuffer = bufCopy
				app.currentCursorLocation++
			}
		}
	}
}

type StateNormal struct {
	app *App
}

func (s *StateNormal) HandleInput(event termbox.Event) {
	if event.Type == termbox.EventKey {
		if event.Key == termbox.KeyCtrlS {
			s.app.currentState = &StateServerSelection{
				app: s.app,
			}
		} else if event.Key == termbox.KeyCtrlH {
			s.app.currentState = &StateChannelSelection{
				app: s.app,
			}
		} else {
			s.app.HandleTextInput(event)
		}
	}
}
func (s *StateNormal) RefreshDisplay() {}

type StateServerSelection struct {
	app          *App
	curSelection int
}

func (s *StateServerSelection) HandleInput(event termbox.Event) {
	if event.Type == termbox.EventKey {
		if event.Key == termbox.KeyArrowUp {
			s.curSelection--
		} else if event.Key == termbox.KeyArrowDown {
			s.curSelection++
		} else if event.Key == termbox.KeyEnter {

			state := s.app.session.State
			state.RLock()
			defer state.RUnlock()

			if s.curSelection < len(state.Guilds) || s.curSelection >= 0 {
				s.app.selectedServerId = state.Guilds[s.curSelection].ID
			}

			s.app.currentState = &StateNormal{s.app}
		}
	}
}
func (s *StateServerSelection) RefreshDisplay() {
	app.CreateServerWindow(s.curSelection)
}

type StateChannelSelection struct {
	app          *App
	curSelection int
}

func (s *StateChannelSelection) HandleInput(event termbox.Event) {
	if event.Type == termbox.EventKey {
		if event.Key == termbox.KeyArrowUp {
			s.curSelection--
		} else if event.Key == termbox.KeyArrowDown {
			s.curSelection++
		} else if event.Key == termbox.KeyEnter {
			state := s.app.session.State
			state.RLock()
			defer state.RUnlock()

			curGuild, err := state.Guild(s.app.selectedServerId)
			if err != nil {
				log.Println("Error getting guild: ", err.Error())
				return
			}

			realList := make([]*discordgo.Channel, 0)
			for _, v := range curGuild.Channels {
				if v.Type == "text" {
					realList = append(realList, v)
				}
			}

			if s.curSelection < len(realList) || s.curSelection >= 0 {
				s.app.selectedChannelId = realList[s.curSelection].ID
			}

			s.app.currentState = &StateNormal{s.app}
		}
	}
}
func (s *StateChannelSelection) RefreshDisplay() {
	app.CreateChannelWindow(s.curSelection)
}
