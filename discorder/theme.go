package discorder

import (
	"encoding/json"
	"github.com/jonas747/discorder/ui"
	"github.com/nsf/termbox-go"
)

type Theme struct {
	Name    string `json:"name"`
	Author  string `json:"author"`
	Comment string `json:"comment"`

	ColorMode termbox.OutputMode `json:"color_mode"` // see termbox.OutputMode for info

	WindowBorder ThemeAttribPair `json:"window_border"`
	WindowFill   ThemeAttribPair `json:"window_fill"`

	ListElementNormal   ThemeAttribPair `json:"list_element_normal"`
	ListElementMarked   ThemeAttribPair `json:"list_element_marked"`
	ListElementSelected ThemeAttribPair `json:"list_element_selected"`
	ListElementBoth     ThemeAttribPair `json:"list_element_selected_marked"`

	TitleBar         ThemeAttribPair `json:"title_bar"`
	NotificationsBar ThemeAttribPair `json:"notifications_bar"`
	TypingBar        ThemeAttribPair `json:"typing_bar"`

	TextOther ThemeAttribPair `json:"text_other"`

	InputChat  ThemeAttribPair `json:"input_chat"`
	SendPrompt ThemeAttribPair `json:"send_prompt"`

	InputOther ThemeAttribPair `json:"input_other"`

	MessageTimestamp     ThemeAttribPair `json:"message_timestamp"`
	MessageServerChannel ThemeAttribPair `json:"message_server_channel"`
	MessageDirect        ThemeAttribPair `json:"message_direct_channel"`
	MessageAuthor        ThemeAttribPair `json:"message_author"`
	MessageContent       ThemeAttribPair `json:"message_content"`
	MessageLog           ThemeAttribPair `json:"message_log"`
}

func (t *Theme) ApplyList(list *ui.ListWindow) {
	list.NormalFG = t.ListElementNormal.FG.Attribute()
	list.NormalBG = t.ListElementNormal.BG.Attribute()
	list.MarkedFG = t.ListElementMarked.FG.Attribute()
	list.MarkedBG = t.ListElementMarked.BG.Attribute()
	list.SelectedFG = t.ListElementSelected.FG.Attribute()
	list.SelectedBG = t.ListElementSelected.BG.Attribute()
	list.MarkedSelectedFG = t.ListElementBoth.FG.Attribute()
	list.MarkedSelectedBG = t.ListElementBoth.BG.Attribute()

	t.ApplyWindow(list.Window)
}

func (t *Theme) ApplyWindow(window *ui.Window) {
	window.BorderFG = t.WindowBorder.FG.Attribute()
	window.BorderBG = t.WindowBorder.BG.Attribute()
	window.FillBG = t.WindowFill.BG.Attribute()
}

func ApplyThemeText(text *ui.Text, pair ThemeAttribPair) {
	text.BG = pair.BG.Attribute()
	text.FG = pair.FG.Attribute()
}

func (t *Theme) Read() ([]byte, error) {
	out, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}
	return out, nil
}

var DefaultTheme = &Theme{
	Name:    "Default Theme",
	Author:  "jonas747",
	Comment: "The default theme for discorder",

	WindowBorder: ThemeAttribPair{ThemeAttribFromTermbox(termbox.ColorWhite), ThemeAttribFromTermbox(termbox.ColorBlack)},
	WindowFill:   ThemeAttribPair{ThemeAttribute{}, ThemeAttribFromTermbox(termbox.ColorBlack)},

	ListElementNormal:   ThemeAttribPair{ThemeAttribute{}, ThemeAttribute{}},
	ListElementMarked:   ThemeAttribPair{ThemeAttribute{}, ThemeAttribFromTermbox(termbox.ColorYellow)},
	ListElementSelected: ThemeAttribPair{ThemeAttribute{}, ThemeAttribFromTermbox(termbox.ColorCyan)},
	ListElementBoth:     ThemeAttribPair{ThemeAttribute{}, ThemeAttribFromTermbox(termbox.ColorBlue)},

	SendPrompt: ThemeAttribPair{ThemeAttribFromTermbox(termbox.ColorYellow | termbox.AttrBold), ThemeAttribute{}},

	MessageTimestamp:     ThemeAttribPair{ThemeAttribFromTermbox(termbox.ColorBlue), ThemeAttribute{}},
	MessageServerChannel: ThemeAttribPair{ThemeAttribFromTermbox(termbox.ColorGreen), ThemeAttribute{}},
	MessageDirect:        ThemeAttribPair{ThemeAttribFromTermbox(termbox.ColorMagenta | termbox.AttrBold), ThemeAttribute{}},
	MessageAuthor:        ThemeAttribPair{ThemeAttribFromTermbox(termbox.ColorCyan | termbox.AttrBold), ThemeAttribute{}},

	TypingBar:        ThemeAttribPair{ThemeAttribFromTermbox(termbox.ColorCyan), ThemeAttribute{}},
	NotificationsBar: ThemeAttribPair{ThemeAttribute{}, ThemeAttribFromTermbox(termbox.ColorYellow)},
}

type ThemeAttribute struct {
	Color     uint8 `json:"color"`
	Bold      bool  `json:"bold"`
	Underline bool  `json:"underline"`
	Reverse   bool  `json:"reverse"`
}

func ThemeAttribFromTermbox(attr termbox.Attribute) ThemeAttribute {
	return ThemeAttribute{
		Color:     uint8(attr & 0xff),
		Bold:      attr&termbox.AttrBold != 0,
		Underline: attr&termbox.AttrUnderline != 0,
		Reverse:   attr&termbox.AttrReverse != 0,
	}
}

func (t *ThemeAttribute) Attribute() termbox.Attribute {
	attr := termbox.Attribute(t.Color)
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

func (t *ThemeAttribPair) AttribPair() ui.AttribPair {
	return ui.AttribPair{
		FG: t.FG.Attribute(),
		BG: t.BG.Attribute(),
	}
}
