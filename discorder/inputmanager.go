package discorder

import (
	"encoding/json"
	"github.com/jonas747/discorder/ui"
	"github.com/jonas747/termbox-go"
	"io/ioutil"
	"log"
	"path/filepath"
	"sync"
)

type InputManager struct {
	app  *App
	stop chan *sync.WaitGroup

	defaultKeybinds []*KeyBind
	userKeybinds    []*KeyBind
	eventBuffer     []termbox.Event
}

func NewInputManager(app *App) *InputManager {
	return &InputManager{
		app:  app,
		stop: make(chan *sync.WaitGroup),
	}
}

func (im *InputManager) Run() {
	im.InitializeKeybinds()
	im.PollEvents()
}

func (im *InputManager) HandleInputEvent(event termbox.Event) {
	commandProcessed := false

	im.eventBuffer = append(im.eventBuffer, event)
	// Check both built in and user defined keybinds, with userdefined ones as priority
	if event.Type == termbox.EventKey {
		partial, full := im.CheckBinds(im.userKeybinds)
		if partial || full {
			commandProcessed = true
		} else {
			partial, full = im.CheckBinds(im.defaultKeybinds)
			if partial || full {
				commandProcessed = true
			} else {
				im.eventBuffer = []termbox.Event{}
			}
		}
	}
	im.app.Lock()
	defer im.app.Unlock()
	if !commandProcessed {
		ui.RunFunc(im.app, func(e ui.Entity) {
			inputHandler, ok := e.(ui.InputHandler)
			if ok {
				inputHandler.HandleInput(event)
			}
		})
	}
	im.app.Draw()
}

func (im *InputManager) CheckBinds(binds []*KeyBind) (partialMatch, fullMatch bool) {
	for _, v := range binds {
		partialMatch, fullMatch = v.Check(im.eventBuffer)
		if fullMatch {
			if v.Command != "nop" {
				im.app.Lock()
				v.Run(im.app)
				im.app.Unlock()
			}
			im.eventBuffer = []termbox.Event{}
			return
		}

		if partialMatch {
			return
		}
	}
	return
}

func (im *InputManager) PollEvents() {
	for {
		evt := termbox.PollEvent()

		select {
		case wg := <-im.stop:
			wg.Done()
			log.Println("Event polling stopped")
			return
		default:
			break
		}
		im.HandleInputEvent(evt)
	}
}

func (im *InputManager) InitializeKeybinds() {
	var defaultBinds []*KeyBind
	err := json.Unmarshal(DefaultKeybinds, &defaultBinds)
	if err != nil {
		panic(err)
	}

	im.defaultKeybinds = defaultBinds

	// Create default binds file
	err = ioutil.WriteFile(filepath.Join(im.app.configDir, "keybinds-default.json"), DefaultKeybinds, 0755)
	if err != nil {
		log.Println("Error writing default keybinds", err)
	}

	// Load user binds
	file, err := ioutil.ReadFile(filepath.Join(im.app.configDir, "keybinds-user.json"))
	if err != nil {
		log.Println("Error reading user keybinds", err)
		return
	}

	var userBinds []*KeyBind
	err = json.Unmarshal(file, &userBinds)
	if err != nil {
		log.Println("Error decoding user binds:", err)
	}
	im.userKeybinds = userBinds
}
