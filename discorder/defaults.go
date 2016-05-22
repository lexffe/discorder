package discorder

var DefaultKeybinds = []byte(`[
	{"key": "CtrlQ", "command": "quit"},
	{"key": "CtrlS", "command": "servers"},
	{"key": "CtrlO", "command": "settings"},
	{"key": "CtrlX", "command": "commands"},
	{"key": "CtrlL", "command": "clear_log"},
	{"key": "CtrlT", "command": "reload_theme"},

	{"key": "ArrowLeft", "command": "move_cursor", "args": {"amount": 1, "direction": "left"}},
	{"key": "ArrowRight", "command": "move_cursor", "args": {"amount": 1, "direction": "right"}},
	{"key": "ArrowUp", "command": "scroll", "args": {"amount": 1, "direction": "up"}},
	{"key": "ArrowDown", "command": "scroll", "args": {"amount": 1, "direction": "down"}},
	
	{"key": "Enter", "command": "select"},
	{"key": "CtrlSpace", "command": "mark"},
	{"key": "Backspace", "command": "erase", "args": {"amount": 1, "direction": "left"}},
	{"key": "Backspace2", "command": "erase", "args": {"amount": 1, "direction": "left"}},
	{"key": "Delete", "command": "erase", "args": {"amount": 1, "direction": "right"}},

	{"key": "Alt+Backspace", "command": "close_window"},
	{"key": "Alt+Backspace2", "command": "close_window"}

]`)

var DefaultTheme = []byte(`{
    "name": "Default theme",
    "author": "jonas747",
    "comment": "The default discorder theme",
    "color_mode": 0,
    "theme":{
		"window_border": {
			"fg": {"color": "white", "bold": true},
			"bg": {"color": "black"}
		},
		"window_fill": {
			"bg": {"color": "black"}
		},
		"list_element_normal": {
			"bg": {"color": "black"}
		},
		"element_marked": {
			"bg": {"color": "yellow"}
		},
		"element_selected": {
			"bg": {"color": "cyan"}
		},
		"element_selected_marked": {
			"bg": {"color": "blue"}
		},
		"message_timestamp": {
			"fg": {"color": "blue"}
		},
		"message_server_channel": {
			"fg": {"color": "green"}
		},
		"message_direct_channel": {
			"fg": {"color": "magenta", "bold": true}
		},
		"message_author": {
			"fg": {"color": "cyan", "bold": true}
		},
		"message_content": {},
		"message_log": {},
		"title_bar": {
			"fg": {"color": "green", "bold": true, "underline": true}
		},
		"notifications_bar": {
			"bg": {"color": "blue"}
		},
		"typing_bar": {
			"fg": {"color": "cyan"}
		},
		"text_other": {},
		"input_chat": {},
		"send_prompt": {
			"fg": {"color": "green", "bold": true}
		},
		"input_other": {}
	}
}`)
