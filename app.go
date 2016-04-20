package main

import (
	"fmt"
	"github.com/0xAX/notificator"
	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/discorder/common"
	"github.com/jonas747/discorder/ui"
	"github.com/nsf/termbox-go"
	"log"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

type App struct {
	running        bool
	stopping       bool
	stopChan       chan chan error
	msgRecvChan    chan *discordgo.Message // Unused atm... remove?
	session        *discordgo.Session
	inputEventChan chan termbox.Event

	typingManager TypingManager

	stopPollEvents chan chan bool

	logBuffer []*common.LogMessage
	logFile   *os.File
	logLock   sync.Mutex

	*ui.BaseEntity

	ViewManager *ViewManager

	notifications *notificator.Notificator
	config        *Config
}

func NewApp(config *Config, logPath string) *App {
	logFile, err := os.Create(logPath)

	notify := notificator.New(notificator.Options{
		AppName: "Discorder",
	})

	a := &App{
		config:        config,
		notifications: notify,
		BaseEntity:    &ui.BaseEntity{},
	}
	if err == nil {
		a.logFile = logFile
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
	session.State.MaxMessageCount = 100
	app.session.AddHandler(app.Ready)
	app.session.AddHandler(app.TypingStart)
	app.session.AddHandler(app.messageCreate)
	err = session.Open()
	return err
}

func (app *App) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Mentions != nil {
		for _, v := range m.Mentions {
			if v.ID == s.State.User.ID {
				if app.notifications != nil {
					author := "Unknown?"
					if m.Author != nil {
						author = m.Author.Username
					}
					app.notifications.Push(author, m.ContentWithMentionsReplaced(), "", notificator.UR_NORMAL)
				}
			}
		}
	}
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

func (app *App) Init() {
	// Initialize the channels
	app.msgRecvChan = make(chan *discordgo.Message)
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

	app.typingManager = TypingManager{
		in: make(chan *discordgo.TypingStart),
	}
	go app.typingManager.Run()

	// Start polling events
	go app.PollEvents()

	app.Init()
	log.Println("Initialized!")
	app.ViewManager.OnInit()

	ticker := time.NewTicker(1000 * time.Millisecond)
	for {
		select {
		case errChan := <-app.stopChan:
			app.running = false
			// app.config.LastServer = app.selectedServerId
			// app.config.LastChannel = app.selectedChannelId
			// app.config.ListeningChannels = app.listeningChannels
			app.config.Save(configPath)
			pollStopped := make(chan bool)
			// Stop the event polling
			go delayedInterrupt(1 * time.Millisecond) // might not work 100% all cases? should probably replace
			app.stopPollEvents <- pollStopped
			<-pollStopped
			errChan <- nil
			return
		case _ = <-app.msgRecvChan:
			//app.HandleMessage(*msg)
		case evt := <-app.inputEventChan:
			app.HandleInputEvent(evt)
		case <-ticker.C:
			app.Draw()
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
		obj := &common.LogMessage{
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

func (app *App) Ready(s *discordgo.Session, r *discordgo.Ready) {
	log.Println("Received ready!")

	// app.session.State.Lock()
	// app.session.State.Ready = *r
	// app.session.State.Unlock()

	// for _, g := range r.Guilds {
	// 	for _, ch := range g.Channels {
	// 		for _, ls := range app.listeningChannels {
	// 			if ch.ID == ls {
	// 				go app.GetHistory(ls, 25, "", "")
	// 				break
	// 			}
	// 		}
	// 	}
	// }
	app.ViewManager.OnReady()
	app.PrintWelcome()
}

func (app *App) PrintWelcome() {
	log.Println("You are using Discorder V" + VERSION + "! If you stumble upon any issues or bugs then please let me know!\n(Press ctrl-o For help)")
}

func (app *App) Destroy() { app.DestroyChildren() }
