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

	// session := discordgo.Session{
	// 	ShouldReconnectOnError: true,
	// 	StateEnabled:           true,
	// 	State:                  discordgo.NewState(),
	// }

	app := &App{
		session: session,
		history: list.New(),
	}

	session.AddHandler(app.messageCreate)

	// err := session.Login(user, password)

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

	ticker := time.NewTicker(100 * time.Millisecond)

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
		if app.selectedServerId != "" {
			_, err := app.session.State.GuildChannel(app.selectedServerId, chId)
			if err != nil {
				// We ignore t since its not on our server
			} else {
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
	} else if event.Type == termbox.EventResize {
		app.RefreshDisplay()
	}

	app.currentState.HandleInput(event)
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
		} else if event.Key == termbox.KeyEnter {
			// send
			_, err := s.app.session.ChannelMessageSend(s.app.selectedChannelId, s.app.currentSendBuffer)
			if err != nil {
				log.Println("Error sending: ", err)
			}
			s.app.currentSendBuffer = ""
			s.app.currentCursorLocation = 0
		} else if event.Key == termbox.KeyArrowLeft {
			s.app.currentCursorLocation--
			if s.app.currentCursorLocation < 0 {
				s.app.currentCursorLocation = 0
			}
		} else if event.Key == termbox.KeyArrowRight {
			s.app.currentCursorLocation++
			bufLen := utf8.RuneCountInString(s.app.currentSendBuffer)
			if s.app.currentCursorLocation >= bufLen {
				s.app.currentCursorLocation = bufLen - 1
			}
		} else if event.Key == termbox.KeyBackspace || event.Key == termbox.KeyBackspace2 {
			bufLen := utf8.RuneCountInString(s.app.currentSendBuffer)
			if bufLen == 0 {
				return
			}
			if s.app.currentCursorLocation == bufLen-1 {
				_, size := utf8.DecodeLastRuneInString(s.app.currentSendBuffer)
				s.app.currentCursorLocation--
				s.app.currentSendBuffer = s.app.currentSendBuffer[:len(s.app.currentSendBuffer)-size]
			} else if s.app.currentCursorLocation == 1 {
				_, size := utf8.DecodeRuneInString(s.app.currentSendBuffer)
				s.app.currentCursorLocation--
				s.app.currentSendBuffer = s.app.currentSendBuffer[size:]
			} else if s.app.currentCursorLocation == 0 {
				return
			} else {
				runeSlice := []rune(s.app.currentSendBuffer)
				newSlice := append(runeSlice[:s.app.currentCursorLocation], runeSlice[s.app.currentCursorLocation+1:]...)
				s.app.currentSendBuffer = string(newSlice)
				s.app.currentCursorLocation--
			}
		} else if event.Ch != 0 || event.Key == termbox.KeySpace {
			char := event.Ch
			if event.Key == termbox.KeySpace {
				char = ' '
			}

			bufLen := utf8.RuneCountInString(s.app.currentSendBuffer)
			if s.app.currentCursorLocation == bufLen-1 {
				s.app.currentSendBuffer += string(char)
				s.app.currentCursorLocation++
			} else if s.app.currentCursorLocation == 0 {
				s.app.currentSendBuffer = string(char) + s.app.currentSendBuffer
				//s.app.currentCursorLocation++
			} else {
				bufSlice := []rune(s.app.currentSendBuffer)
				bufCopy := ""

				for i := 0; i < len(bufSlice); i++ {
					bufCopy += string(bufSlice[i])
					if i == s.app.currentCursorLocation {
						bufCopy += string(char)
					}
				}
				s.app.currentSendBuffer = bufCopy
				s.app.currentCursorLocation++
				// before := bufSlice[:s.app.currentCursorLocation]
				// after := bufSlice[s.app.currentCursorLocation:]
				// before = append(before[len(before)-1:], char)
				// before = append(before, after...)
				// str := string(before)
				// s.app.currentSendBuffer = str
				// s.app.currentCursorLocation++
			}

		}
		s.app.RefreshDisplay()
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

			realList := make([]discordgo.Channel, 0)
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
