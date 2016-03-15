package main

import (
	"container/list"
	"github.com/bwmarrin/discordgo"
	"github.com/nsf/termbox-go"
	"log"
	"time"
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
	session        discordgo.Session
	inputEventChan chan termbox.Event
	logChan        chan string

	curChannel discordgo.Channel
	history    *list.List

	currentState  State
	selectedIndex int

	stopPollEvents chan chan bool
}

func Login(user, password string) (*App, error) {
	session := discordgo.Session{
		ShouldReconnectOnError: true,
		StateEnabled:           true,
		State:                  discordgo.NewState(),
	}

	app := &App{
		session: session,
		history: list.New(),
	}

	session.AddHandler(app.messageCreate)

	err := session.Login(user, password)
	if err != nil {
		return nil, err
	}

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
	app.history.PushFront(msg)

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
		}
	}
}
func (s *StateNormal) RefreshDisplay() {}

type StateServerSelection struct {
	app         *App
	curSelecton int
}

func (s *StateServerSelection) HandleInput(event termbox.Event) {
	if event.Type == termbox.EventKey {
		if event.Key == termbox.KeyArrowUp {
			s.curSelecton--
		} else if event.Key == termbox.KeyArrowDown {
			s.curSelecton++
		} else if event.Key == termbox.KeyEnter {
			s.app.currentState = &StateNormal{s.app}
		}
	}
}
func (s *StateServerSelection) RefreshDisplay() {
	app.CreateServerWindow(s.curSelecton)
}
