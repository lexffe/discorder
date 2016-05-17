package discorder

import (
	"fmt"
	"github.com/0xAX/notificator"
	"github.com/jonas747/discorder/ui"
	"github.com/jonas747/discordgo"
	"github.com/nsf/termbox-go"
	"log"
	"os"
	"runtime/debug"
	"sync"
	"time"
)

const (
	WindowTextBG = termbox.ColorBlack
)

type App struct {
	sync.RWMutex
	running        bool
	stopping       bool
	stopChan       chan interface{}
	session        *discordgo.Session
	inputEventChan chan termbox.Event

	typingRoutine  *TypingRoutine
	ackRoutine     *AckRoutine
	requestRoutine *RequestRoutine

	stopPollEvents chan *sync.WaitGroup

	*ui.BaseEntity

	ViewManager *ViewManager

	notifications *notificator.Notificator
	config        *Config
	settings      *discordgo.Settings
	guildSettings []*discordgo.UserGuildSettings
	firstMessages map[string]string

	configPath  string
	debug       bool
	dGoDebugLvl int
}

func NewApp(configPath string, debug bool, dgoDebug int) (*App, error) {
	notify := notificator.New(notificator.Options{
		AppName: "Discorder",
	})

	config, err := LoadOrCreateConfig(configPath)
	if err != nil {
		return nil, err
	}

	app := &App{
		config:        config,
		notifications: notify,
		BaseEntity:    &ui.BaseEntity{},
		firstMessages: make(map[string]string),
		configPath:    configPath,
		debug:         debug,
		dGoDebugLvl:   dgoDebug,
	}
	return app, nil
}

func (app *App) Login(user, password, token string) error {
	var session *discordgo.Session
	var err error
	if app.session != nil {
		session = app.session
		err = session.Login(user, password)
	} else {
		if token != "" {
			session, err = discordgo.New(token)
		} else {
			session, err = discordgo.New(user, password)
		}
	}

	session.LogLevel = app.dGoDebugLvl
	app.session = session

	if err != nil {
		return err
	}

	session.StateEnabled = true
	session.State.MaxMessageCount = 100

	app.session.AddHandler(app.Ready)

	app.session.AddHandler(app.typingStart)
	app.session.AddHandler(app.messageCreate)
	app.session.AddHandler(app.messageDelete)
	app.session.AddHandler(app.messageUpdate)
	app.session.AddHandler(app.messageAck)
	app.session.AddHandler(app.guildSettingsUpdated)
	app.session.AddHandler(app.userSettingsUpdated)
	app.session.AddHandler(app.guildCreated)

	err = session.Open()
	return err
}

func (app *App) Stop() {
	app.Lock()

	if app.stopping {
		app.Unlock()
		return
	}
	app.stopping = true
	if app.running {
		app.Unlock()
		app.shutdown()
		return
	}

	app.Unlock()
	return
}

func (app *App) init() {
	// Initialize the channels
	app.stopChan = make(chan interface{})
	app.inputEventChan = make(chan termbox.Event)
	app.stopPollEvents = make(chan *sync.WaitGroup)

	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	termbox.SetInputMode(termbox.InputAlt)

	app.ViewManager = NewViewManager(app)
	app.AddChild(app.ViewManager)

	app.typingRoutine = NewTypingRoutine(app)
	go app.typingRoutine.Run()

	app.ackRoutine = NewAckRoutine(app)
	go app.ackRoutine.Run()

	app.requestRoutine = NewRequestRoutine()
	go app.requestRoutine.Run()
}

// Lsiten on the channels for incoming messages
func (app *App) Run() {
	// Some initialization
	app.Lock()
	if app.running {
		log.Println("Tried to run app while already running")
		app.Unlock()
		return
	}
	app.running = true

	defer func() {
		if r := recover(); r != nil {
			if termbox.IsInit {
				termbox.Close()
			}
			fmt.Println("Panic!: ", r, string(debug.Stack()))
			os.Exit(1)
		}
	}()

	// Start polling events
	go app.PollEvents()

	app.init()
	log.Println("Initialized!")
	app.ViewManager.OnInit()
	app.Unlock()

	ticker := time.NewTicker(1000 * time.Millisecond)
	for {
		select {
		case _ = <-app.stopChan:
			if termbox.IsInit {
				termbox.Close()
			}
			return
		case evt := <-app.inputEventChan:
			app.Lock()
			app.HandleInputEvent(evt)
			app.Unlock()
		case <-ticker.C:
			app.Lock()
			app.Draw()
			app.Unlock()
		}
	}
}

func delayedInterrupt(d time.Duration) {
	time.Sleep(d)
	termbox.Interrupt()
}

func (app *App) HandleInputEvent(event termbox.Event) {
	if event.Type == termbox.EventKey {
		if event.Key == termbox.KeyCtrlQ {
			log.Println("Stopping...")
			go app.Stop()
		}
	}

	entities := app.Children(true)
	for _, entity := range entities {
		inputHandler, ok := entity.(ui.InputHandler)
		if ok {
			inputHandler.HandleInput(event)
		}
	}
	app.Draw()
}

// Todo remove 10 layer lazy limit... Maps?
func (app *App) Draw() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	// Run predraw
	ui.RunFunc(app, func(e ui.Entity) {
		updater, ok := e.(ui.PreDrawHandler)
		if ok {
			updater.PreDraw()
		}
	})

	// Build the layers
	layers := make([][]ui.DrawHandler, 10)

	entities := app.Children(true)
	for _, entity := range entities {
		drawable, ok := entity.(ui.DrawHandler)
		if ok {
			layer := drawable.GetDrawLayer()
			layers[layer] = append(layers[layer], drawable)
		}
	}

	for i := 0; i < 10; i++ {
		for _, drawable := range layers[i] {
			drawable.Draw()
		}
	}
	termbox.Flush()
}

func (app *App) PollEvents() {
	for {
		evt := termbox.PollEvent()

		select {
		case wg := <-app.stopPollEvents:
			wg.Done()
			log.Println("Event polling stopped")
			return
		default:
			break
		}
		app.inputEventChan <- evt
	}
}

func (app *App) PrintWelcome() {
	log.Println("You are using Discorder V" + VERSION + "! If you stumble upon any issues or bugs then please let me know!\n(Press ctrl-o For help)")
}

func (app *App) shutdown() {
	app.Lock()
	app.running = false
	// app.config.LastServer = app.selectedServerId
	// app.config.LastChannel = app.selectedChannelId
	if app.ViewManager != nil && app.ViewManager.SelectedMessageView != nil {
		app.config.ListeningChannels = app.ViewManager.SelectedMessageView.Channels
		app.config.LastChannel = app.ViewManager.talkingChannel
		app.config.AllPrivateMode = app.ViewManager.SelectedMessageView.ShowAllPrivate
	}

	if app.session != nil {
		app.config.AuthToken = app.session.Token
	}

	app.config.Save(app.configPath)

	err := app.session.Close()
	if err != nil {
		log.Println("Error closing:", err)
	}

	app.Unlock()

	var wg sync.WaitGroup
	wg.Add(4)

	// Stop the event polling
	go delayedInterrupt(10 * time.Millisecond)

	app.stopPollEvents <- &wg
	app.ackRoutine.stop <- &wg
	app.typingRoutine.stop <- &wg
	app.requestRoutine.stop <- &wg

	wg.Wait()
	app.stopChan <- true
}

func (app *App) Destroy() { app.DestroyChildren() }
