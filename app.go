package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/nsf/termbox-go"
	"log"
	"os"
	"runtime/debug"
	"time"
)

const (
	DiscordTimeFormat = "2006-01-02T15:04:05-07:00"
)

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

	//history           *list.List
	listeningChannels []string

	displayMessages []*DisplayMessage
	logBuffer       []*LogMessage

	currentState State

	config *Config

	currentTextBuffer     string
	currentCursorLocation int
}

func NewApp(config *Config) *App {
	a := &App{
		//history: list.New(),
		config: config,
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
	if app.config.ListeningChannels != nil && len(app.config.ListeningChannels) > 0 {
		app.listeningChannels = app.config.ListeningChannels
	}

	app.SetState(&StateLogin{app: app})

	ticker := time.NewTicker(1000 * time.Millisecond)
	for {
		select {
		case errChan := <-app.stopChan:
			app.running = false
			app.config.LastServer = app.selectedServerId
			app.config.LastChannel = app.selectedChannelId
			app.config.ListeningChannels = app.listeningChannels
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
		case msg := <-app.logChan:
			app.HandleLogMessage(msg)
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
			_, size := termbox.Size()
			app.BuildDisplayMessages(size - 2)
			app.RefreshDisplay()
		}
	}
}

func delayedInterrupt(d time.Duration) {
	time.Sleep(d)
	termbox.Interrupt()
}

func (app *App) HandleLogMessage(msg string) {
	now := time.Now()
	obj := &LogMessage{
		timestamp: now,
		content:   msg,
	}
	app.logBuffer = append(app.logBuffer, obj)
}

func (app *App) HandleInputEvent(event termbox.Event) {
	if event.Type == termbox.EventKey {
		if event.Key == termbox.KeyCtrlQ {
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

type DisplayMessage struct {
	discordMessage *discordgo.Message
	logMessage     *LogMessage
	isLogMessage   bool
	timestamp      time.Time
}

type LogMessage struct {
	timestamp time.Time
	content   string
}

// A target for optimisation when i get that far
// Builds a list of messages to display from all of the channels were listening to, pm's and the log
func (app *App) BuildDisplayMessages(size int) {
	state := app.session.State
	state.RLock()
	defer state.RUnlock()

	messages := make([]*DisplayMessage, size)

	// Get a sorted list
	var lastMessage *DisplayMessage
	for i := 0; i < size; i++ {
		// Get newest message after "lastMessage", set it to curNewestMessage if its newer than that

		var curNewestMessage *DisplayMessage

		// Check the channels were listening on
		for _, listeningChannelId := range app.listeningChannels {
			// Avoid deadlock since guildchannel also calls rlock, whch will block if there was a new message in the meantime causing lock to be called
			// before that
			state.RUnlock()
			channel, err := state.GuildChannel(app.selectedServerId, listeningChannelId)
			state.RLock()
			if err != nil {
				continue
			}
			for j := len(channel.Messages) - 1; j >= 0; j-- {
				msg := channel.Messages[j]
				parsedTimestamp, _ := time.Parse(DiscordTimeFormat, msg.Timestamp)
				if lastMessage == nil || parsedTimestamp.Before(lastMessage.timestamp) {
					if curNewestMessage == nil || parsedTimestamp.After(curNewestMessage.timestamp) {
						curNewestMessage = &DisplayMessage{
							discordMessage: msg,
							timestamp:      parsedTimestamp,
						}
					}
					break // Newest message after last since ordered
				}
			}

		}
		// Check for newest pm's
		for _, privateChannel := range state.PrivateChannels {
			for j := len(privateChannel.Messages) - 1; j >= 0; j-- {
				msg := privateChannel.Messages[j]
				parsedTimestamp, _ := time.Parse(DiscordTimeFormat, msg.Timestamp)
				if lastMessage == nil || parsedTimestamp.Before(lastMessage.timestamp) {
					if curNewestMessage == nil || parsedTimestamp.After(curNewestMessage.timestamp) {
						curNewestMessage = &DisplayMessage{
							discordMessage: msg,
							timestamp:      parsedTimestamp,
						}
					}
					break // Newest message after last since ordered
				}
			}
		}

		// Check the logerino
		for j := len(app.logBuffer) - 1; j >= 0; j-- {
			msg := app.logBuffer[j]
			if lastMessage == nil || msg.timestamp.Before(lastMessage.timestamp) {
				if curNewestMessage == nil || msg.timestamp.After(curNewestMessage.timestamp) {
					curNewestMessage = &DisplayMessage{
						logMessage:   msg,
						timestamp:    msg.timestamp,
						isLogMessage: true,
					}
				}
				break // Newest message after last since ordered
			}
		}
		if curNewestMessage == nil {
			// Looks like we ran out of messages to display! D:
			break
		}
		messages[i] = curNewestMessage
		lastMessage = curNewestMessage
	}
	app.displayMessages = messages
}
