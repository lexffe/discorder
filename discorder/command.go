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
)

type Command struct {
	Name        string
	Description string
	Arguments   []ArgumentDef
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
	Shift   bool        `json:"shift"`
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
		return errors.New("Key not found", k.StringKey)
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

	"CtrlTilde":      ermbox.KeyCtrlTilde,
	"CtrlSpace":      ermbox.KeyCtrlSpace,
	"CtrlA":          ermbox.KeyCtrlA,
	"CtrlB":          ermbox.KeyCtrlB,
	"CtrlC":          ermbox.KeyCtrlC,
	"CtrlD":          ermbox.KeyCtrlD,
	"CtrlE":          ermbox.KeyCtrlE,
	"CtrlF":          ermbox.KeyCtrlF,
	"CtrlG":          ermbox.KeyCtrlG,
	"Backspace":      ermbox.KeyBackspace,
	"CtrlH":          ermbox.KeyCtrlH,
	"Tab":            ermbox.KeyTab,
	"CtrlI":          ermbox.KeyCtrlI,
	"CtrlJ":          ermbox.KeyCtrlJ,
	"CtrlK":          ermbox.KeyCtrlK,
	"CtrlL":          ermbox.KeyCtrlL,
	"Enter":          ermbox.KeyEnter,
	"CtrlM":          ermbox.KeyCtrlM,
	"CtrlN":          ermbox.KeyCtrlN,
	"CtrlO":          ermbox.KeyCtrlO,
	"CtrlP":          ermbox.KeyCtrlP,
	"CtrlQ":          ermbox.KeyCtrlQ,
	"CtrlR":          ermbox.KeyCtrlR,
	"CtrlS":          ermbox.KeyCtrlS,
	"CtrlT":          ermbox.KeyCtrlT,
	"CtrlU":          ermbox.KeyCtrlU,
	"CtrlV":          ermbox.KeyCtrlV,
	"CtrlW":          ermbox.KeyCtrlW,
	"CtrlX":          ermbox.KeyCtrlX,
	"CtrlY":          ermbox.KeyCtrlY,
	"CtrlZ":          ermbox.KeyCtrlZ,
	"Esc":            ermbox.KeyEsc,
	"CtrlLsqBracket": ermbox.KeyCtrlLsqBracket,
	"CtrlBackslash":  ermbox.KeyCtrlBackslash,
	"CtrlRsqBracket": ermbox.KeyCtrlRsqBracket,
	"CtrlSlash":      ermbox.KeyCtrlSlash,
	"CtrlUnderscore": ermbox.KeyCtrlUnderscore,
	"Space":          ermbox.KeySpace,
	"Backspace2":     ermbox.KeyBackspace2,
	"Ctrl2":          ermbox.KeyCtrl2,
	"Ctrl3":          ermbox.KeyCtrl3,
	"Ctrl4":          ermbox.KeyCtrl4,
	"Ctrl5":          ermbox.KeyCtrl5,
	"Ctrl6":          ermbox.KeyCtrl6,
	"Ctrl7":          ermbox.KeyCtrl7,
	"Ctrl8":          ermbox.KeyCtrl8,
}
