package main

import (
	"container/list"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/nsf/termbox-go"
	"log"
	"os"
	"runtime/debug"
	"time"
	"unicode/utf8"
)

const HistorySize = 1000

type App struct {
	running        bool
	stopping       bool
	stopChan       chan chan error
	msgRecvChan    chan *discordgo.Message
	session        *discordgo.Session
	inputEventChan chan termbox.Event
	logChan        chan string

	stopPollEvents chan chan bool

	selectedServerId  string
	selectedChannelId string
	selectedGuild     *discordgo.Guild
	selectedChannel   *discordgo.Channel

	history           *list.List
	listeningChannels []string

	currentState State

	config *Config

	currentTextBuffer     string
	currentCursorLocation int
}

func NewApp(config *Config) *App {
	a := &App{
		history: list.New(),
		config:  config,
	}
	return a
}

func (app *App) Login(user, password string) error {
	var session *discordgo.Session
	var err error
	if app.session != nil {
		session = app.session
		err = session.Login(user, password)
	} else {
		session, err = discordgo.New(user, password)
	}

	app.session = session

	if err != nil {
		return err
	}

	session.StateEnabled = true
	session.AddHandler(app.messageCreate)
	err = session.Open()
	return err
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

	defer func() {
		if r := recover(); r != nil {
			termbox.Close()
			fmt.Println("Panic!: ", r, string(debug.Stack()))
			os.Exit(1)
		}
	}()

	app.msgRecvChan = make(chan *discordgo.Message)
	app.stopChan = make(chan chan error)
	app.inputEventChan = make(chan termbox.Event)
	app.stopPollEvents = make(chan chan bool)
	app.logChan = make(chan string)

	log.SetOutput(app)
	log.Println("Started!")

	err := termbox.Init()
	if err != nil {
		panic(err)
	}

	// Start polling events
	go app.PollEvents()

	if app.config.LastChannel != "" {
		app.selectedChannelId = app.config.LastChannel
	}

	if app.config.LastServer != "" {
		app.selectedServerId = app.config.LastServer
	}

	app.SetState(&StateLogin{app: app})

	ticker := time.NewTicker(1000 * time.Millisecond)
	for {
		select {
		case errChan := <-app.stopChan:
			app.running = false

			app.config.LastServer = app.selectedServerId
			app.config.LastChannel = app.selectedChannelId
			app.config.Save(configPath)
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
			if app.session != nil {
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
						var err2 error
						app.selectedChannel, err2 = app.session.State.PrivateChannel(app.selectedChannelId)
						if err2 != nil {
							log.Println("App.Run: ", err, err2)
						}
					}
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

func (app *App) SetState(state State) {
	oldState := app.currentState
	if oldState != nil {
		oldState.Exit()
	}

	app.currentState = state
	state.Enter()
}
