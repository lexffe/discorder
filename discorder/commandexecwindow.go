package discorder

import (
	"fmt"
	"github.com/jonas747/discorder/common"
	"github.com/jonas747/discorder/ui"
	"log"
	"strconv"
	"strings"
)

type CommandExecWindow struct {
	*ui.BaseEntity
	app        *App
	layer      int
	menuWindow *ui.MenuWindow
	command    Command

	curArgs Arguments
}

type CustomMenuType int

const (
	CustomMenuExecute CustomMenuType = iota
)

func NewCommandExecWindow(layer int, app *App, command Command, presetArgs Arguments) *CommandExecWindow {
	execWindow := &CommandExecWindow{
		BaseEntity: &ui.BaseEntity{},
		app:        app,
		menuWindow: ui.NewMenuWindow(layer, app.ViewManager.UIManager, false),
		command:    command,
		layer:      layer,
	}

	execWindow.menuWindow.Transform.AnchorMax = common.NewVector2F(1, 1)
	execWindow.menuWindow.Transform.Top = 1
	execWindow.menuWindow.Transform.Bottom = 2

	execWindow.menuWindow.Window.Title = "Execute command"
	execWindow.menuWindow.Window.Footer = ":)"

	app.ApplyThemeToMenu(execWindow.menuWindow)

	execWindow.Transform.AddChildren(execWindow.menuWindow)

	execWindow.Transform.AnchorMax = common.NewVector2F(1, 1)

	execWindow.Transform.Right = 2
	execWindow.Transform.Left = 1
	execWindow.curArgs = presetArgs

	app.ViewManager.UIManager.AddWindow(execWindow)
	preRunHelper := command.GetPreRunHelper()
	if preRunHelper != "" {
		execWindow.RunPreHelper(preRunHelper)
	} else {
		if !execWindow.CheckAutoExec() {
			execWindow.GenMenu()
		} else {
			return nil
		}
	}

	return execWindow
}

func (cew *CommandExecWindow) RunPreHelper(helperArg string) {
	var arg *ArgumentDef

	args := cew.command.GetArgs(nil)
	for _, v := range args {
		if v.Name == helperArg {
			arg = v
			break
		}
	}

	if arg == nil {
		log.Println("could not find arg", helperArg)
		return
	} else if arg.Helper == nil {
		log.Println("Argument has no helper", helperArg)
		return
	}

	arg.Helper.Run(cew.app, cew.layer+2, func(result string) {
		if cew.curArgs == nil {
			cew.curArgs = make(map[string]interface{})
		}
		cew.curArgs[helperArg] = ParseArgumentString(result, arg.Datatype)
		if !cew.CheckAutoExec() {
			cew.GenMenu()
		}
	})
}

func (cew *CommandExecWindow) Destroy() {
	cew.app.ViewManager.UIManager.RemoveWindow(cew)
	cew.DestroyChildren()
}

func (cew *CommandExecWindow) GenMenu() {
	items := make([]*ui.MenuItem, 0)
	for _, arg := range cew.command.GetArgs(cew.curArgs) {
		helper := &ui.MenuItem{
			Name:       arg.GetName(),
			Info:       arg.Description,
			Decorative: true,
		}
		input := &ui.MenuItem{
			Name:      arg.Name,
			Info:      arg.Description,
			IsInput:   true,
			InputType: arg.Datatype,
			UserData:  arg,
		}
		// Set the preset arg if any
		set := false
		for key, presetArg := range cew.curArgs {
			if key == arg.Name {
				input.InputDefaultText = fmt.Sprintf("%v", presetArg)
				set = true
				break
			}
		}
		// Set to default value
		if !set {
			curVal := arg.GetCurrentVal(cew.app)
			if curVal != "" {
				input.InputDefaultText = curVal
			}
		}

		items = append(items, helper, input)
	}
	execText := cew.command.GetExecText()
	if execText == "" {
		execText = "Run"
	}

	exec := &ui.MenuItem{
		Name:     execText,
		Info:     "Execute the command with specified args",
		UserData: CustomMenuExecute,
	}
	items = append(items, exec)
	cew.menuWindow.SetOptions(items)
}

func (cew *CommandExecWindow) Select() {
	element := cew.menuWindow.GetHighlighted()
	if element == nil {
		return
	}

	if element.IsCategory {
		cew.menuWindow.Select()
		return
	}

	if element.UserData == nil {
		return
	}

	switch t := element.UserData.(type) {
	case CustomMenuType:
		switch t {
		case CustomMenuExecute:
			cew.Execute(true)
		}
	// Run a argument helper if any
	case *ArgumentDef:
		if t.Helper != nil {
			t.Helper.Run(cew.app, cew.layer+2, func(result string) {
				if element.Input != nil {
					element.Input.TextBuffer = result
					element.Input.CursorLocation = 0
					if t.RebuildOnChanged {
						cew.Rebuild()
					}
				}
			})
		}
	}
}

func (cew *CommandExecWindow) Rebuild() {
	cew.curArgs = cew.ParseArgs()
	log.Println(cew.curArgs)
	cew.GenMenu()
}

func (cew *CommandExecWindow) Execute(parseArgs bool) {
	args := cew.curArgs
	if parseArgs {
		args = cew.ParseArgs()
	}
	cew.app.RunCommand(cew.command, Arguments(args))
	parent := cew.Transform.Parent
	if parent != nil {
		if parent == cew.app.ViewManager.menuContainer.GetTransform() {
			cew.app.ViewManager.RemoveWindow(cew)
		} else {
			cew.Transform.Parent.RemoveChild(cew, true)
		}
	} else {
		// Manually destroy
		cew.Destroy()
	}
}

func (cew *CommandExecWindow) ParseArgs() Arguments {
	args := make(map[string]interface{})
	for _, item := range cew.menuWindow.Options {
		if !item.IsInput {
			continue
		}
		buf := item.Input.TextBuffer
		if buf == "" {
			continue
		}
		args[item.Name] = ParseArgumentString(buf, item.InputType)
	}
	return args
}

func (cew *CommandExecWindow) CheckAutoExec() bool {
	args := cew.curArgs
	argDefs := cew.command.GetArgs(cew.curArgs)

	if len(args) < 1 && len(argDefs) < 1 {
		cew.Execute(false)
		return true
	} else if len(args) < 1 {
		return false
	}

	combinations := cew.command.GetArgCombinations()

	if len(combinations) > 0 {
	OUTER:
		for _, combination := range combinations {
			for _, step := range combination {
				found := false
				for key := range args {
					if key == step {
						found = true
						break
					}
				}
				if !found {
					continue OUTER
				}
			}
			// all steps of the combination matched
			cew.Execute(false)
			return true
		}
	} else if len(args) == len(argDefs) {
		cew.Execute(false)
		return true
	}

	return false
}

func ParseArgumentString(arg string, dataType ui.DataType) interface{} {
	switch dataType {
	case ui.DataTypeBool:
		lowerBuf := strings.ToLower(arg)
		b, _ := strconv.ParseBool(lowerBuf)
		return b
	case ui.DataTypeInt:
		i, _ := strconv.ParseInt(arg, 10, 64)
		return i
	case ui.DataTypeFloat:
		f, _ := strconv.ParseFloat(arg, 64)
		return f
	}

	return arg
}
