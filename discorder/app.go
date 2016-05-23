package discorder

import (
	"encoding/json"
	"fmt"
	"github.com/0xAX/notificator"
	"github.com/jonas747/discorder/common"
	"github.com/jonas747/discorder/ui"
	"github.com/jonas747/discordgo"
	"github.com/nsf/termbox-go"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"sync"
	"time"
)

const (
	WindowTextBG = termbox.ColorBlack
)

type App struct {
	sync.RWMutex
	*ui.BaseEntity

	running  bool
	stopping bool             // true if in the process of stopping
	stopChan chan interface{} // Sending on this channel will instantly stop (not gracefull)

	session        *discordgo.Session
	typingRoutine  *TypingRoutine
	ackRoutine     *AckRoutine
	requestRoutine *RequestRoutine

	ViewManager  *ViewManager
	InputManager *InputManager

	config       *Config
	defaultTheme *Theme
	userTheme    *Theme

	notifications *notificator.Notificator
	settings      *discordgo.Settings
	guildSettings []*discordgo.UserGuildSettings
	firstMessages map[string]string

	configDir string

	configPath  string
	themePath   string
	debug       bool
	dGoDebugLvl int
}

func NewApp(configPath, themePath string, debug bool, dgoDebug int) (*App, error) {
	notify := notificator.New(notificator.Options{
		AppName: "Discorder",
	})

	configDir, err := GetCreateConfigDir()
	if err != nil {
		log.Println("Failed getting proper config dirs, falling back to current directory")
		configDir = ""
	}

	app := &App{
		BaseEntity:    &ui.BaseEntity{},
		notifications: notify,
		firstMessages: make(map[string]string),
		debug:         debug,
		dGoDebugLvl:   dgoDebug,
		configPath:    configPath,
		themePath:     themePath,
		configDir:     configDir,
	}
	app.Transform.AnchorMax = common.NewVector2I(1, 1)

	err = app.InitializeConfigFiles()
	return app, err
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

	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	termbox.SetInputMode(termbox.InputAlt)

	app.ViewManager = NewViewManager(app)
	app.Transform.AddChildren(app.ViewManager)

	app.typingRoutine = NewTypingRoutine(app)
	go app.typingRoutine.Run()

	app.ackRoutine = NewAckRoutine(app)
	go app.ackRoutine.Run()

	app.requestRoutine = NewRequestRoutine()
	go app.requestRoutine.Run()

	app.InputManager = NewInputManager(app)
	go app.InputManager.Run()

	// out, err := DefaultTheme.Read()
	// if err == nil {
	// 	log.Println(string(out))
	// }
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

	// Initialize and run
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

// Todo remove 10 layer lazy limit... Maps?
func (app *App) Draw() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	// Run Update
	ui.RunFunc(app, func(e ui.Entity) {
		updater, ok := e.(ui.UpdateHandler)
		if ok {
			updater.Update()
		}
	})

	// Run Update
	ui.RunFunc(app, func(e ui.Entity) {
		updater, ok := e.(ui.LateUpdateHandler)
		if ok {
			updater.LateUpdate()
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

func (app *App) RunCommand(command *Command, args Arguments) {
	if command == nil {
		log.Println("Tried to run a nonexstant command")
		return
	}
	if command.Run != nil {
		command.Run(app, args)
	}

	ui.RunFunc(app, func(e ui.Entity) {
		cmdHandler, ok := e.(CommandHandler)
		if ok {
			cmdHandler.OnCommand(command, args)
		}
	})
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

	savePath := app.configPath
	if app.configPath == "" {
		savePath = filepath.Join(app.configDir, "discorder.json")
	}

	app.config.Save(savePath)

	if app.session != nil {
		err := app.session.Close()
		if err != nil {
			log.Println("Error closing:", err)
		}
	}

	app.Unlock()

	var wg sync.WaitGroup
	wg.Add(4)

	// Stop the event polling
	go delayedInterrupt(10 * time.Millisecond)

	app.InputManager.stop <- &wg
	app.ackRoutine.stop <- &wg
	app.typingRoutine.stop <- &wg
	app.requestRoutine.stop <- &wg

	wg.Wait()
	app.stopChan <- true
}

func (app *App) Destroy() { app.DestroyChildren() }

func (app *App) InitializeConfigFiles() error {
	// Load general config
	configPath := ""
	if app.configPath != "" {
		configPath = app.configPath
	} else {
		configPath = filepath.Join(app.configDir, "discorder.json")
	}

	config, err := LoadOrCreateConfig(configPath)
	if err != nil {
		return err
	}
	app.config = config

	// Load default theme
	var defaultTheme Theme
	err = json.Unmarshal(DefaultTheme, &defaultTheme)
	if err != nil {
		panic(err) // Panic cuase were in serious trouble then
	}
	app.defaultTheme = &defaultTheme

	// Check if theme dir exists
	skipCreateFile := false
	themesDir := filepath.Join(app.configDir, "themes")
	_, err = os.Stat(themesDir)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(themesDir, 0755)
			if err != nil {
				log.Println("Failed creating themes dir", err)
				skipCreateFile = true
			}
		} else {
			skipCreateFile = true
			log.Println("Failed checking if themes dir exist", err)
		}
	}

	// Write default theme file
	if !skipCreateFile {
		outFilePath := filepath.Join(themesDir, "default.json")
		err = ioutil.WriteFile(outFilePath, DefaultTheme, 0755)
		if err != nil {
			log.Println("Error creating default theme file", err)
		}
	}

	// Load user theme
	if app.themePath != "" {
		app.userTheme = LoadTheme(app.themePath)
	} else if app.config.Theme != "" {
		app.userTheme = LoadTheme(filepath.Join(themesDir, app.config.Theme))
	}
	return nil
}
