package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/nsf/termbox-go"
	"log"
	"os"
	"runtime/debug"
	"strings"
	"sync"
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
	logFile         *os.File
	logFileLock     sync.Mutex

	currentState State

	config *Config

	currentTextBuffer     string
	currentCursorLocation int

	curChatScroll int
}

func NewApp(config *Config, logPath string) *App {
	logFile, err := os.Create(logPath)

	a := &App{
		//history: list.New(),
		config: config,
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
	termbox.SetInputMode(termbox.InputAlt)

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
			app.BuildDisplayMessages(size + app.curChatScroll - 2)
			app.RefreshDisplay()
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
	for _, splitStr := range split {
		if splitStr == "" {
			continue
		}
		obj := &LogMessage{
			timestamp: now,
			content:   splitStr,
		}
		app.logBuffer = append(app.logBuffer, obj)
	}

	//app.logFile.Write([]byte(msg))
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
// Also a target for cleaning up
// Builds a list of messages to display from all of the channels were listening to, pm's and the log
func (app *App) BuildDisplayMessages(size int) {
	// Ackquire the state, or create one if null (incase were starting)
	var state *discordgo.State
	if app.session != nil && app.session.State != nil {
		state = app.session.State
	} else {
		state = discordgo.NewState()
	}
	state.RLock()
	defer state.RUnlock()

	messages := make([]*DisplayMessage, size)

	// Holds the start indexes in the newest message search
	listeningIndexes := make([]int, len(app.listeningChannels))
	pmIndexes := make([]int, len(state.PrivateChannels))
	// Init the slices with silly vals
	for i := 0; i < len(app.listeningChannels); i++ {
		listeningIndexes[i] = -10
	}
	for i := 0; i < len(state.PrivateChannels); i++ {
		pmIndexes[i] = -10
	}
	nextLogIndex := len(app.logBuffer) - 1

	// Get a sorted list
	var lastMessage *DisplayMessage
	var beforeTime time.Time
	for i := 0; i < size; i++ {
		// Get newest message after "lastMessage", set it to curNewestMessage if its newer than that

		var newestListening *DisplayMessage
		newestListeningIndex := 0    // confusing, but the index of the indexes slice
		nextListeningStartIndex := 0 // And the actual next start index to use

		// Check the channels were listening on
		for k, listeningChannelId := range app.listeningChannels {
			// Avoid deadlock since guildchannel also calls rlock, whch will block if there was a new message in the meantime causing lock to be called
			// before that
			state.RUnlock()
			channel, err := state.GuildChannel(app.selectedServerId, listeningChannelId)
			state.RLock()
			if err != nil {
				continue
			}

			newest, nextIndex := GetNewestMessageBefore(channel.Messages, beforeTime, listeningIndexes[k])

			if newest != nil && (newestListening == nil || !newest.timestamp.Before(newestListening.timestamp)) {
				newestListening = newest
				newestListeningIndex = k
				nextListeningStartIndex = nextIndex
			}
		}

		var newestPm *DisplayMessage
		newestPmIndex := 0    // confusing, but the index of the indexes slice
		nextPmStartIndex := 0 // And the actual next start index to use

		// Check for newest pm's
		for k, privateChannel := range state.PrivateChannels {

			newest, nextIndex := GetNewestMessageBefore(privateChannel.Messages, beforeTime, pmIndexes[k])

			if newest != nil && (newestPm == nil || !newest.timestamp.Before(newestPm.timestamp)) {
				newestPm = newest
				newestPmIndex = k
				nextPmStartIndex = nextIndex
			}
		}

		newNextLogIndex := 0
		var newestLog *DisplayMessage

		// Check the logerino
		for j := nextLogIndex; j >= 0; j-- {
			msg := app.logBuffer[j]
			if !msg.timestamp.After(beforeTime) || beforeTime.IsZero() {
				if newestLog == nil || !msg.timestamp.Before(newestLog.timestamp) {
					newestLog = &DisplayMessage{
						logMessage:   msg,
						timestamp:    msg.timestamp,
						isLogMessage: true,
					}
					newNextLogIndex = j - 1
				}
				break // Newest message after last since ordered
			}
		}

		if newestListening != nil &&
			(newestPm == nil || !newestListening.timestamp.Before(newestPm.timestamp)) &&
			(newestLog == nil || !newestListening.timestamp.Before(newestLog.timestamp)) {
			messages[i] = newestListening
			listeningIndexes[newestListeningIndex] = nextListeningStartIndex

			lastMessage = newestListening
			beforeTime = lastMessage.timestamp
		} else if newestPm != nil &&
			(newestListening == nil || !newestPm.timestamp.Before(newestListening.timestamp)) &&
			(newestLog == nil || !newestPm.timestamp.Before(newestLog.timestamp)) {

			messages[i] = newestPm
			pmIndexes[newestPmIndex] = nextPmStartIndex

			lastMessage = newestPm
			beforeTime = lastMessage.timestamp
		} else if newestLog != nil {
			messages[i] = newestLog
			nextLogIndex = newNextLogIndex

			lastMessage = newestLog
			beforeTime = lastMessage.timestamp
		} else {
			break // No new shit!
		}
	}
	app.displayMessages = messages
}

func (app *App) Ready(s *discordgo.Session, r *discordgo.Ready) {
	log.Println("Received ready!")

	app.session.State.Lock()
	app.session.State.Ready = *r
	app.session.State.Unlock()

	for _, g := range r.Guilds {
		for _, ch := range g.Channels {
			for _, ls := range app.listeningChannels {
				if ch.ID == ls {
					go app.GetHistory(ls, 25, "", "")
					break
				}
			}
		}
	}
	app.PrintWelcome()
}

func (app *App) PrintWelcome() {
	log.Println("You are using Discorder V" + VERSION + "! If you stumble upon any issues or bugs then please let me know!\n(Press ctrl-o For help)")
}

func GetNewestMessageBefore(msgs []*discordgo.Message, before time.Time, startIndex int) (*DisplayMessage, int) {
	if startIndex == -10 {
		startIndex = len(msgs) - 1
	}

	for j := startIndex; j >= 0; j-- {
		msg := msgs[j]
		parsedTimestamp, _ := time.Parse(DiscordTimeFormat, msg.Timestamp)
		if !parsedTimestamp.After(before) || before.IsZero() { // Reason for !after is so that we still show all the messages with same timestamps
			curNewestMessage := &DisplayMessage{
				discordMessage: msg,
				timestamp:      parsedTimestamp,
			}
			return curNewestMessage, j - 1
		}
	}
	return nil, 0
}
