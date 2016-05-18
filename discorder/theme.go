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

	WindowBorder ui.AttribPair `json:"window_border"`
	WindowFill   ui.AttribPair `json:"window_fill"`

	ListElementNormal   ui.AttribPair `json:"list_element_normal"`
	ListElementMarked   ui.AttribPair `json:"list_element_marked"`
	ListElementSelected ui.AttribPair `json:"list_element_selected"`
	ListElementBoth     ui.AttribPair `json:"list_element_selected_marked"`

	TitleBar         ui.AttribPair `json:"title_bar"`
	NotificationsBar ui.AttribPair `json:"notifications_bar"`
	TypingBar        ui.AttribPair `json:"typing_bar"`

	TextOther ui.AttribPair `json:"text_other"`

	InputChat  ui.AttribPair `json:"input_chat"`
	SendPrompt ui.AttribPair `json:"send_prompt"`

	InputOther ui.AttribPair `json:"input_other"`

	MessageTimestamp     ui.AttribPair `json:"message_timestamp"`
	MessageServerChannel ui.AttribPair `json:"message_server_channel"`
	MessageDirect        ui.AttribPair `json:"message_direct_channel"`
	MessageAuthor        ui.AttribPair `json:"message_author"`
	MessageContent       ui.AttribPair `json:"message_content"`
	MessageLog           ui.AttribPair `json:"message_log"`
}

// TODO...
func (t *Theme) ApplyList()   {}
func (t *Theme) ApplyWindow() {}
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

	WindowBorder: ui.AttribPair{termbox.ColorWhite, termbox.ColorBlack},
	WindowFill:   ui.AttribPair{0, termbox.ColorBlack},

	ListElementNormal:   ui.AttribPair{0, 0},
	ListElementMarked:   ui.AttribPair{0, termbox.ColorYellow},
	ListElementSelected: ui.AttribPair{0, termbox.ColorCyan},
	ListElementBoth:     ui.AttribPair{0, termbox.ColorBlue},

	SendPrompt: ui.AttribPair{termbox.ColorYellow | termbox.AttrBold, 0},

	MessageTimestamp:     ui.AttribPair{termbox.ColorBlue, 0},
	MessageServerChannel: ui.AttribPair{termbox.ColorGreen, 0},
	MessageDirect:        ui.AttribPair{termbox.ColorMagenta, 0},
	MessageAuthor:        ui.AttribPair{termbox.ColorCyan | termbox.AttrBold, 0},
}
