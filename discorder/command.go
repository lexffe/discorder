package discorder

import (
	"encoding/json"
	"errors"
	"github.com/nsf/termbox-go"
)

type ArgumentDataType int

const (
	ArgumentDataTypeInt ArgumentDataType = iota
	ArgumentDataTypeFloat
	ArgumentDataTypeString
	ArgumentDataTypeBool
)

type Command struct {
	Name        string
	Description string
	Args        []*ArgumentDef
	Category    string
	Run         func(app *App, args []*Argument)
}

type ArgumentDef struct {
	Name     string
	Optional bool
	Datatype ArgumentDataType
}

type Argument struct {
	Name string      `json:"name"`
	Val  interface{} `json:"val"`
}

func (a *Argument) Int() (int, bool) {
	fVal, ok := a.Val.(float64)
	if !ok {
		return 0, false
	}

	return int(fVal), true
}

type KeyBind struct {
	Command string      `json:"command"`
	Args    []*Argument `json:"args"`
	Key     KeyBindKey  `json:"key"`
	Alt     bool        `json:"alt"`
}

type KeyBindKey struct {
	StringKey string
	TermKey   termbox.Key
}

func (k *KeyBindKey) UnmarshalJSON(data []byte) error {
	err := json.Unmarshal(data, &k.StringKey)
	if err != nil {
		return err
	}

	key, ok := Keys[k.StringKey]
	if !ok {
		return errors.New("Key not found: " + k.StringKey)
	}

	k.TermKey = key
	return nil
}

var Keys = map[string]termbox.Key{
	"F1":         termbox.KeyF1,
	"F2":         termbox.KeyF2,
	"F3":         termbox.KeyF3,
	"F4":         termbox.KeyF4,
	"F5":         termbox.KeyF5,
	"F6":         termbox.KeyF6,
	"F7":         termbox.KeyF7,
	"F8":         termbox.KeyF8,
	"F9":         termbox.KeyF9,
	"F10":        termbox.KeyF10,
	"F11":        termbox.KeyF11,
	"F12":        termbox.KeyF12,
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
	"Backspace":      termbox.KeyBackspace,
	"CtrlH":          termbox.KeyCtrlH,
	"Tab":            termbox.KeyTab,
	"CtrlI":          termbox.KeyCtrlI,
	"CtrlJ":          termbox.KeyCtrlJ,
	"CtrlK":          termbox.KeyCtrlK,
	"CtrlL":          termbox.KeyCtrlL,
	"Enter":          termbox.KeyEnter,
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
	"Esc":            termbox.KeyEsc,
	"CtrlLsqBracket": termbox.KeyCtrlLsqBracket,
	"CtrlBackslash":  termbox.KeyCtrlBackslash,
	"CtrlRsqBracket": termbox.KeyCtrlRsqBracket,
	"CtrlSlash":      termbox.KeyCtrlSlash,
	"CtrlUnderscore": termbox.KeyCtrlUnderscore,
	"Space":          termbox.KeySpace,
	"Backspace2":     termbox.KeyBackspace2,
	"Ctrl2":          termbox.KeyCtrl2,
	"Ctrl3":          termbox.KeyCtrl3,
	"Ctrl4":          termbox.KeyCtrl4,
	"Ctrl5":          termbox.KeyCtrl5,
	"Ctrl6":          termbox.KeyCtrl6,
	"Ctrl7":          termbox.KeyCtrl7,
	"Ctrl8":          termbox.KeyCtrl8,
}
