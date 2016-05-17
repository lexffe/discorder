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
	"strings"
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
	stopChan       chan chan error
	session        *discordgo.Session
	inputEventChan chan termbox.Event

	typingRoutine  *TypingRoutine
	ackRoutine     *AckRoutine
	requestRoutine *RequestRoutine

	stopPollEvents chan chan bool

	logBuffer []*LogMessage
	logFile   *os.File
	logLock   sync.Mutex

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

func NewApp(configPath, logPath string, debug bool, dgoDebug int) (*App, error) {
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

	if debug {
		logFile, err := os.Create(logPath)
		if err == nil {
			app.logFile = logFile
		}
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

func (app *App) Stop() error {
	app.Lock()

	if app.stopping {
		app.Unlock()
		return nil
	}
	app.stopping = true
	errChan := make(chan error)
	if app.running {
		app.Unlock()
		app.stopChan <- errChan
		return <-errChan
	}

	app.Unlock()
	return nil
}

func (app *App) init() {
	// Initialize the channels
	app.stopChan = make(chan chan error)
	app.inputEventChan = make(chan termbox.Event)
	app.stopPollEvents = make(chan chan bool)
	log.SetOutput(app)

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
			termbox.Close()
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
		case errChan := <-app.stopChan:
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
			pollStopped := make(chan bool)

			// Stop the event polling
			go delayedInterrupt(100 * time.Millisecond) // might not work 100% all cases? should probably replace

			app.Unlock()
			app.stopPollEvents <- pollStopped
			app.ackRoutine.Stop <- true
			<-pollStopped
			errChan <- nil
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

func (app *App) HandleLogMessage(msg string) {
	split := strings.Split(msg, "\n")
	now := time.Now()

	app.logLock.Lock()
	for _, splitStr := range split {
		if splitStr == "" {
			continue
		}
		obj := &LogMessage{
			Timestamp: now,
			Content:   splitStr,
		}
		app.logBuffer = append(app.logBuffer, obj)
	}
	if app.logFile != nil {
		app.logFile.Write([]byte(msg)) // TODO: Move this somewhere else, to its own goroutine to avoid slowdowns
	}
	app.logLock.Unlock()
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

func (app *App) PrintWelcome() {
	log.Println("You are using Discorder V" + VERSION + "! If you stumble upon any issues or bugs then please let me know!\n(Press ctrl-o For help)")
}

func (app *App) Destroy() { app.DestroyChildren() }
