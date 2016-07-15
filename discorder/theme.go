package discorder

import (
	"encoding/json"
	"github.com/jonas747/discorder/ui"
	"github.com/jonas747/termbox-go"
	"io/ioutil"
	"log"
	"path/filepath"
	"strconv"
)

type Theme struct {
	Name         string                     `json:"name"`
	Author       string                     `json:"author"`
	Comment      string                     `json:"comment"`
	ColorMode    termbox.OutputMode         `json:"color_mode"`    // see termbox.OutputMode for info
	DiscrimTable []ThemeAttribPair          `json:"discrim_table"` // Used for different color for users and channels & servers
	Theme        map[string]ThemeAttribPair `json:"theme"`
}

func (t *Theme) GetAttribute(key string, fg bool) (attrib termbox.Attribute, ok bool) {
	var pair ThemeAttribPair
	pair, ok = t.Theme[key]
	if !ok {
		return
	}

	if fg {
		attrib = pair.FG.Attribute()
	} else {
		attrib = pair.BG.Attribute()
	}
	return
}

func (app *App) GetThemeAttribPair(key string) ThemeAttribPair {
	if app.userTheme != nil {
		pair, ok := app.userTheme.Theme[key]
		if ok {
			return pair
		}
	}

	pair, _ := app.defaultTheme.Theme[key]
	return pair
}

func (app *App) GetThemeAttribute(key string, fg bool) termbox.Attribute {
	if app.userTheme != nil {
		userAttrib, ok := app.userTheme.GetAttribute(key, fg)
		if ok {
			return userAttrib
		}
	}
	defaultAttrib, _ := app.defaultTheme.GetAttribute(key, fg)
	return defaultAttrib
}

func (app *App) GetThemeDiscrim(id string) ThemeAttribPair {
	parsed, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		log.Println("Failed parsign id", err)
	}

	discrims := app.defaultTheme.DiscrimTable
	if app.userTheme != nil && len(app.userTheme.DiscrimTable) > 0 {
		discrims = app.userTheme.DiscrimTable
	}

	return discrims[uint64(parsed)%uint64(len(discrims))]
}

func (app *App) ApplyThemeToMenu(menu *ui.MenuWindow) {
	menu.StyleNormal = app.GetThemeAttribPair("element_normal").AttribPair()
	menu.StyleMarked = app.GetThemeAttribPair("element_marked").AttribPair()
	menu.StyleSelected = app.GetThemeAttribPair("element_selected").AttribPair()
	menu.StyleMarkedSelected = app.GetThemeAttribPair("element_selected_marked").AttribPair()
	menu.StyleInputNormal = app.GetThemeAttribPair("element_input_normal").AttribPair()

	app.ApplyThemeToText(menu.SearchInput.Text, "menu_search")
	app.ApplyThemeToText(menu.InfoText, "menu_info_text")

	menu.LowerWindow.Border = app.GetThemeAttribPair("menu_info_border").AttribPair()
	menu.LowerWindow.FillBG = app.GetThemeAttribute("menu_info_fill", false)

	app.ApplyThemeToWindow(menu.Window)
}

func (app *App) ApplyThemeToWindow(window *ui.Window) {
	window.Border = app.GetThemeAttribPair("window_border").AttribPair()
	window.FillBG = app.GetThemeAttribute("window_fill", false)
}

func (app *App) ApplyThemeToText(text *ui.Text, key string) {
	pair := app.GetThemeAttribPair(key)
	text.Style = pair.AttribPair()
}

func (t *Theme) Read() ([]byte, error) {
	out, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type ThemeAttribute struct {
	Color     *Color `json:"color"`
	Bold      bool   `json:"bold"`
	Underline bool   `json:"underline"`
	Reverse   bool   `json:"reverse"`
}

func (t *ThemeAttribute) GetColor() uint8 {
	if t.Color == nil {
		return 0
	}
	return uint8(*t.Color)
}

func (t *ThemeAttribute) Attribute() termbox.Attribute {

	attr := termbox.Attribute(t.GetColor())
	if t.Bold {
		attr |= termbox.AttrBold
	}
	if t.Underline {
		attr |= termbox.AttrUnderline
	}
	if t.Reverse {
		attr |= termbox.AttrReverse
	}
	return attr
}

type ThemeAttribPair struct {
	FG ThemeAttribute `json:"fg"`
	BG ThemeAttribute `json:"bg"`
}

func (t ThemeAttribPair) AttribPair() ui.AttribPair {
	return ui.AttribPair{
		FG: t.FG.Attribute(),
		BG: t.BG.Attribute(),
	}
}

type Color uint8

func (c *Color) UnmarshalJSON(data []byte) error {
	var raw interface{}
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return err
	}

	switch t := raw.(type) {
	case string:
		col, ok := Colors[t]
		if !ok {
			log.Println("Color not found", t)
		}
		*c = Color(col)
	case float64:
		intColor := uint8(t)
		*c = Color(intColor)
	}
	return nil
}

var Colors = map[string]uint8{
	"default": 0,
	"black":   1,
	"red":     2,
	"green":   3,
	"yellow":  4,
	"blue":    5,
	"magenta": 6,
	"cyan":    7,
	"white":   8,
}

func (app *App) GetAvailableThemes() ([]string, error) {
	files, err := ioutil.ReadDir(filepath.Join(app.configDir, "themes"))
	if err != nil {
		return nil, err
	}

	out := make([]string, 0)
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		out = append(out, file.Name())
	}
	return out, nil
}

func (app *App) SetUserTheme(theme *Theme) {
	app.userTheme = theme
	termbox.SetOutputMode(termbox.OutputMode(theme.ColorMode))
	log.Println("Set theme to ", theme.Name)

}
