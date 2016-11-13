package discorder

import (
	"encoding/json"
	"github.com/jonas747/discorder/ui"
	"github.com/jonas747/termbox-go"
	"log"
	"strings"
)

// Simple command struct that implements the Command interface, should cover most cases
type SimpleCommand struct {
	Name            string
	Description     string
	CustomExecText  string
	Args            []*ArgumentDef
	ArgCombinations [][]string
	Category        []string
	PreRunHelper    string
	RunFunc         func(app *App, args Arguments)
	StatusFunc      func(app *App) string
	CustomWindow    CustomCommandWindow
	IgnoreFilter    bool
}

func (s *SimpleCommand) GetName() string {
	return s.Name
}

func (s *SimpleCommand) GetDescription(app *App) string {
	desc := s.Description
	if s.StatusFunc != nil {
		desc += "\n" + s.StatusFunc(app)
	}

	if len(s.Args) > 0 {
		for _, v := range s.Args {
			curVal := v.GetCurrentVal(app)
			if curVal != "" {
				desc += "\nCurrent " + v.Name + ":" + curVal
			}
		}
	}

	return desc
}

func (s *SimpleCommand) GetArgs(curArgs Arguments) []*ArgumentDef {
	return s.Args
}

func (s *SimpleCommand) GetCategory() []string {
	return s.Category
}

func (s *SimpleCommand) GetExecText() string {
	return s.CustomExecText
}

func (s *SimpleCommand) Run(app *App, args Arguments) {
	if s.RunFunc != nil {
		s.RunFunc(app, args)
	}
}

func (s *SimpleCommand) GetPreRunHelper() string {
	return s.PreRunHelper
}

func (s *SimpleCommand) GetArgCombinations() [][]string {
	return s.ArgCombinations
}

func (s *SimpleCommand) GetCustomWindow() CustomCommandWindow {
	return s.CustomWindow
}

func (s *SimpleCommand) GetIgnoreFilter() bool {
	return s.IgnoreFilter
}

type Command interface {
	GetName() string                          // Name of the command
	GetDescription(app *App) string           // Decsription
	GetArgs(curArgs Arguments) []*ArgumentDef // Argumend definitions
	GetArgCombinations() [][]string           // Returns possible argument combinations
	GetCustomWindow() CustomCommandWindow     // An optional custom command window
	GetPreRunHelper() string                  // Helper to be ran before main exec window
	GetCategory() []string                    // Category
	GetExecText() string                      // Custom exec button text
	GetIgnoreFilter() bool
	Run(app *App, args Arguments) // Called when the command should be run
}

func (app *App) GenMenuItemFromCommand(cmd Command) *ui.MenuItem {
	cmdItem := &ui.MenuItem{
		Name:     cmd.GetName(),
		Info:     cmd.GetDescription(app),
		UserData: cmd,
	}
	return cmdItem
}

// Argument definition
type ArgumentDef struct {
	Name                   string                // Unique Name for this argument
	DisplayName            string                // Display name that will be shown, if empty name will be shown in exec window
	Description            string                // Simple description
	Optional               bool                  // Wether optional or not- currently unused
	Datatype               ui.DataType           // Datatype for this arg
	Helper                 ArgumentHelper        // Helper to help pick a value from a predefined list
	CurVal                 string                // The current/default value
	CurValFunc             func(app *App) string // Same as above but a function that is ran instead
	RebuildOnChanged       bool                  // Rebuilds the command exec menu if changed
	PreserveValueOnRebuild bool                  // Wether to preserve the value after a rebuild occured
}

func (a *ArgumentDef) GetName() string {
	if a.DisplayName != "" {
		return a.DisplayName
	}
	return a.Name
}

func (a *ArgumentDef) GetCurrentVal(app *App) string {
	if a.CurValFunc != nil {
		return a.CurValFunc(app)
	}
	return a.CurVal
}

func (a *ArgumentDef) String() string {
	out := a.GetName()

	switch a.Datatype {
	case ui.DataTypeString:
		out += ":string"
	case ui.DataTypePassword:
		out += ":password"
	case ui.DataTypeFloat:
		out += ":float"
	case ui.DataTypeInt:
		out += ":int"
	case ui.DataTypeBool:
		out += ":boolean"
	}
	if a.Description != "" {
		out += " - " + a.Description
	}
	return out
}

type Arguments map[string]interface{}

func (a Arguments) Get(key string) (val interface{}, ok bool) {
	val, ok = map[string]interface{}(a)[key]
	return
}

func (a Arguments) Int(key string) (val int, ok bool) {
	var i64 int64
	i64, ok = a.Int64(key)
	val = int(i64)
	return
}

func (a Arguments) Int64(key string) (val int64, ok bool) {
	raw, ok := a.Get(key)
	if !ok {
		return
	}

	switch v := raw.(type) {
	case float64:
		var temp float64
		temp, ok = a.Float64(key)
		val = int64(temp)
	case int64:
		ok = true
		val = v
	}

	return
}

func (a Arguments) Float64(key string) (val float64, ok bool) {
	raw, ok := a.Get(key)
	if !ok {
		return
	}
	val, ok = raw.(float64)
	return
}

func (a Arguments) String(key string) (val string, ok bool) {
	raw, ok := a.Get(key)
	if !ok {
		return
	}
	val, ok = raw.(string)
	return
}

func (a Arguments) Bool(key string) (val bool, ok bool) {
	raw, ok := a.Get(key)
	if !ok {
		return
	}
	val, ok = raw.(bool)
	return
}

type KeyBind struct {
	Command        string         `json:"command"`
	Args           Arguments      `json:"args"`
	KeyComb        KeyCombination `json:"key"`              // Key combination with custom json unmarshaling
	Helpers        []string       `json:"helpers"`          // Runs the helpers for these arguements before running the command
	OpenExecWindow bool           `json:"open_exec_window"` // Instead of running the command opens the command exec window
}

// Fullmatch is a full match, all combinations matched
// Partial means that atleast one of the combinations were matched (in order)
func (k KeyBind) Check(seq []termbox.Event) (partialMatch, fullMatch bool) {
	if len(seq) > len(k.KeyComb.Keys) {
		return
	}

	for i, event := range seq {
		keybindKey := k.KeyComb.Keys[i]
		if (event.Mod&termbox.ModAlt != 0 && !keybindKey.Alt) || (event.Mod&termbox.ModAlt == 0 && keybindKey.Alt) {
			return
		}
		if keybindKey.Char != "" {
			if string(event.Ch) != keybindKey.Char {
				return
			}
		} else {
			if event.Key != keybindKey.Special || event.Ch != 0 {
				return
			}
		}
	}
	if len(seq) < len(k.KeyComb.Keys) {
		partialMatch = true
	} else {
		fullMatch = true
	}
	return
}

func (k KeyBind) Run(app *App) {
	cmd := app.GetCommandByName(k.Command)
	if cmd == nil {
		log.Println("Unknown command", k.Command)
		return
	}

	if k.OpenExecWindow {
		cew := NewCommandExecWindow(6, app, cmd, nil)
		app.ViewManager.AddWindow(cew)
	} else {
		if len(k.Helpers) < 1 {
			app.RunCommand(cmd, k.Args)
		} else {
			k.RunHelper(app, 0, cmd)
		}
	}
}

func (k KeyBind) RunHelper(app *App, index int, cmd Command) {
	args := cmd.GetArgs(nil)
	var arg *ArgumentDef
	for _, v := range args {
		if v.Name == k.Helpers[index] {
			arg = v
			break
		}
	}

	if arg == nil {
		log.Println("Could not run helper for", k.Helpers[index], "; Argument not found")
		return
	}

	if arg.Helper == nil {
		log.Println("Could not run helper for", k.Helpers[index], "; No helper for argument")
		return
	}

	arg.Helper.Run(app, 6, func(result string) {
		k.Args[arg.Name] = ParseArgumentString(result, arg.Datatype)
		if len(k.Helpers) > index+1 {
			k.RunHelper(app, index+1, cmd) // Run the next helper
		} else {
			app.RunCommand(cmd, k.Args)
		}
	})
}

type KeyCombination struct {
	Keys []*KeybindKey
	raw  string
}

// Parses key combinations and sequences
// Keys can be any normal alphanumric key and some special keys found below
// + is additive, work for shift and alt
// - seperates combinations, Eg for the below
// Alt+CtrlX-A
// You have to press alta and ctrl x at the same time followed by A after
func (k *KeyCombination) UnmarshalJSON(data []byte) error {
	raw := ""
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return err
	}
	k.raw = raw
	k.Keys = make([]*KeybindKey, 0)

	seqSplit := strings.Split(raw, "-")
	for _, sequence := range seqSplit {
		modSplit := strings.Split(sequence, "+")
		key := &KeybindKey{}
		for _, mod := range modSplit {
			if mod == "Alt" || mod == "alt" {
				key.Alt = true
				continue
			}
			special, ok := SpecialKeys[mod]
			if ok {
				key.Special = special
				continue
			}
			key.Char = mod
		}
		k.Keys = append(k.Keys, key)
	}
	return nil
}

type KeybindKey struct {
	Alt     bool
	Special termbox.Key
	Char    string
}

// Generic command handler
type CommandHandler interface {
	OnCommand(cmd Command, args Arguments)
}

var SpecialKeys = map[string]termbox.Key{
	"F1":  termbox.KeyF1,
	"F2":  termbox.KeyF2,
	"F3":  termbox.KeyF3,
	"F4":  termbox.KeyF4,
	"F5":  termbox.KeyF5,
	"F6":  termbox.KeyF6,
	"F7":  termbox.KeyF7,
	"F8":  termbox.KeyF8,
	"F9":  termbox.KeyF9,
	"F10": termbox.KeyF10,
	"F11": termbox.KeyF11,
	"F12": termbox.KeyF12,

	"Tab":        termbox.KeyTab,
	"Esc":        termbox.KeyEsc,
	"Enter":      termbox.KeyEnter,
	"Space":      termbox.KeySpace,
	"Backspace":  termbox.KeyBackspace,
	"Backspace2": termbox.KeyBackspace2,
	"Insert":     termbox.KeyInsert,
	"Delete":     termbox.KeyDelete,
	"Home":       termbox.KeyHome,
	"End":        termbox.KeyEnd,
	"Pgup":       termbox.KeyPgup,
	"Pgdn":       termbox.KeyPgdn,
	"ArrowUp":    termbox.KeyArrowUp,
	"ArrowDown":  termbox.KeyArrowDown,
	"ArrowLeft":  termbox.KeyArrowLeft,
	"ArrowRight": termbox.KeyArrowRight,

	"MouseLeft":      termbox.MouseLeft,
	"MouseMiddle":    termbox.MouseMiddle,
	"MouseRight":     termbox.MouseRight,
	"MouseRelease":   termbox.MouseRelease,
	"MouseWheelUp":   termbox.MouseWheelUp,
	"MouseWheelDown": termbox.MouseWheelDown,

	"CtrlTilde":      termbox.KeyCtrlTilde,
	"CtrlSpace":      termbox.KeyCtrlSpace,
	"CtrlA":          termbox.KeyCtrlA,
	"CtrlB":          termbox.KeyCtrlB,
	"CtrlC":          termbox.KeyCtrlC,
	"CtrlD":          termbox.KeyCtrlD,
	"CtrlE":          termbox.KeyCtrlE,
	"CtrlF":          termbox.KeyCtrlF,
	"CtrlG":          termbox.KeyCtrlG,
	"CtrlH":          termbox.KeyCtrlH,
	"CtrlI":          termbox.KeyCtrlI,
	"CtrlJ":          termbox.KeyCtrlJ,
	"CtrlK":          termbox.KeyCtrlK,
	"CtrlL":          termbox.KeyCtrlL,
	"CtrlM":          termbox.KeyCtrlM,
	"CtrlN":          termbox.KeyCtrlN,
	"CtrlO":          termbox.KeyCtrlO,
	"CtrlP":          termbox.KeyCtrlP,
	"CtrlQ":          termbox.KeyCtrlQ,
	"CtrlR":          termbox.KeyCtrlR,
	"CtrlS":          termbox.KeyCtrlS,
	"CtrlT":          termbox.KeyCtrlT,
	"CtrlU":          termbox.KeyCtrlU,
	"CtrlV":          termbox.KeyCtrlV,
	"CtrlW":          termbox.KeyCtrlW,
	"CtrlX":          termbox.KeyCtrlX,
	"CtrlY":          termbox.KeyCtrlY,
	"CtrlZ":          termbox.KeyCtrlZ,
	"CtrlLsqBracket": termbox.KeyCtrlLsqBracket,
	"CtrlBackslash":  termbox.KeyCtrlBackslash,
	"CtrlRsqBracket": termbox.KeyCtrlRsqBracket,
	"CtrlSlash":      termbox.KeyCtrlSlash,
	"CtrlUnderscore": termbox.KeyCtrlUnderscore,

	"Ctrl2": termbox.KeyCtrl2,
	"Ctrl3": termbox.KeyCtrl3,
	"Ctrl4": termbox.KeyCtrl4,
	"Ctrl5": termbox.KeyCtrl5,
	"Ctrl6": termbox.KeyCtrl6,
	"Ctrl7": termbox.KeyCtrl7,
	"Ctrl8": termbox.KeyCtrl8,
}
