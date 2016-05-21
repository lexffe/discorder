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
	im.app.Lock()
	defer im.app.Unlock()

	if event.Type == termbox.EventKey {
		//im.app.CheckCommand(event)
	}

	ui.RunFunc(im.app, func(e ui.Entity) {
		inputHandler, ok := e.(ui.InputHandler)
		if ok {
			inputHandler.HandleInput(event)
		}
	})
	im.app.Draw()
}

func (im *InputManager) CheckBinds(seq []termbox.Event, binds []*KeyBind) (partialMatch, fullMatch bool) {
	for _, v := range binds {
		partialMatch, fullMatch = v.Check(seq)
		if fullMatch {
			im.app.RunCommand(GetCommandByName(v.Command), v.Args)
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
