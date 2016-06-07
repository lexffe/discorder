package discorder

import (
	"github.com/jonas747/discorder/ui"
)

type CommandExecWindow struct {
	*ui.BaseEntity
}

func NewCommandExecWindow(app *App, command *Command) *CommandExecWindow {
	window := &CommandExecWindow{}
	return window
}

func (cew *CommandExecWindow) Destroy() {
	cew.DestroyChildren()
}
