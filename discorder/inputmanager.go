package discorder

import (
	"github.com/jonas747/discorder/ui"
	"github.com/nsf/termbox-go"
	"log"
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
	im.PollEvents()
}

func (im *InputManager) HandleInputEvent(event termbox.Event) {
	im.eventBuffer = append(im.eventBuffer, event)
	// Check both built in and user defined keybinds, with userdefined ones as priority
	if event.Type == termbox.EventKey {
		partial, full := im.CheckBinds(im.userKeybinds)
		if partial || full {
			return
		}
		partial, full = im.CheckBinds(im.defaultKeybinds)
		if partial || full {
			return
		}
	}

	im.app.Lock()
	defer im.app.Unlock()
	ui.RunFunc(im.app, func(e ui.Entity) {
		inputHandler, ok := e.(ui.InputHandler)
		if ok {
			inputHandler.HandleInput(event)
		}
	})
	im.app.Draw()
}

func (im *InputManager) CheckBinds(binds []*KeyBind) (partialMatch, fullMatch bool) {
	for _, v := range binds {
		partialMatch, fullMatch = v.Check(im.eventBuffer)
		if fullMatch {
			if v.Command != "nop" {
				im.app.Lock()
				im.app.RunCommand(GetCommandByName(v.Command), v.Args)
				im.app.Unlock()
			}
			im.eventBuffer = []termbox.Event{}
			return
		}

		if partialMatch {
			return
		}
		im.eventBuffer = []termbox.Event{}
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
