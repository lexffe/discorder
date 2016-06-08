package discorder

import (
	"encoding/json"
	"github.com/jonas747/discorder/ui"
	"github.com/nsf/termbox-go"
	"strings"
)

type SimpleCommand struct {
	Name        string
	Description string
	Args        []*ArgumentDef
	Category    []string
	RunFunc     func(app *App, args Arguments)
	StatusFunc  func(app *App) string
}

func (s *SimpleCommand) GetName() string {
	return s.Name
}

func (s *SimpleCommand) GetDescription(app *App) string {
	desc := s.Description
	if s.StatusFunc != nil {
		desc += "\n" + s.StatusFunc(app)
	}
	return desc
}

func (s *SimpleCommand) GetArgs() []*ArgumentDef {
	return s.Args
}

func (s *SimpleCommand) GetCategory() []string {
	return s.Category
}

func (s *SimpleCommand) Run(app *App, args Arguments) {
	if s.RunFunc != nil {
		s.RunFunc(app, args)
	}
}

type Command interface {
	GetName() string
	GetDescription(app *App) string
	GetArgs() []*ArgumentDef
	GetCategory() []string
	Run(app *App, args Arguments)
}

func (app *App) GenMenuItemFromCommand(cmd Command) *ui.MenuItem {
	cmdItem := &ui.MenuItem{
		Name:     cmd.GetName(),
		Info:     cmd.GetDescription(app),
		UserData: cmd,
	}
	return cmdItem
}

type ArgumentDef struct {
	Name        string
	Description string
	Optional    bool
	Datatype    ui.DataType
}

type Arguments map[string]interface{}

func (a Arguments) Get(key string) (val interface{}, ok bool) {
	val, ok = map[string]interface{}(a)[key]
	return
}

func (a Arguments) Int(key string) (val int, ok bool) {
	fVal, ok := a.Float(key)
	if !ok {
		return 0, false
	}

	val = int(fVal)
	ok = true
	return
}

func (a Arguments) Float(key string) (val float64, ok bool) {
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
	Command string         `json:"command"`
	Args    Arguments      `json:"args"`
	KeyComb KeyCombination `json:"key"`
	Alt     bool           `json:"alt"`
}

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

type KeyCombination struct {
	Keys []*KeybindKey
	raw  string
}

// Alt+CtrlX-A
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
