package discorder

var DefaultKeybinds = []byte(`[
	{"key": "CtrlQ", "command": "quit"},
	{"key": "CtrlS", "command": "servers"},
	{"key": "CtrlO", "command": "settings"},
	{"key": "CtrlX", "command": "commands"},
	{"key": "CtrlL", "command": "clear_log"},
	{"key": "CtrlT", "command": "reload_theme"},
	{"key": "CtrlN", "command": "set_nick", "open_exec_window": true},

	{"key": "ArrowLeft", "command": "move_cursor", "args": {"amount": 1, "direction": "left"}},
	{"key": "ArrowRight", "command": "move_cursor", "args": {"amount": 1, "direction": "right"}},
	{"key": "ArrowUp", "command": "scroll", "args": {"amount": 1, "direction": "up"}},
	{"key": "ArrowDown", "command": "scroll", "args": {"amount": 1, "direction": "down"}},
	{"key": "Pgup", "command": "scroll", "args": {"amount": 10, "direction": "up"}},
	{"key": "Pgdn", "command": "scroll", "args": {"amount": 10, "direction": "down"}},
	
	{"key": "Enter", "command": "select"},
	{"key": "CtrlSpace", "command": "toggle"},
	{"key": "Tab", "command": "switch", "args": {"amount": 1}},
	{"key": "Tab", "command": "switch", "args": {"amount": -1}},
	
	{"key": "Backspace", "command": "erase", "args": {"amount": 1, "direction": "left"}},
	{"key": "Backspace2", "command": "erase", "args": {"amount": 1, "direction": "left"}},
	{"key": "Delete", "command": "erase", "args": {"amount": 1, "direction": "right"}},
		
	{"key": "Alt+Backspace", "command": "back"},
	{"key": "Alt+Backspace2", "command": "back"},
	
	{"key": "F1", "command": "change_tab", "args": {"tab": 1}},
	{"key": "F2", "command": "change_tab", "args": {"tab": 2}},
	{"key": "F3", "command": "change_tab", "args": {"tab": 3}},
	{"key": "F4", "command": "change_tab", "args": {"tab": 4}},
	{"key": "F5", "command": "change_tab", "args": {"tab": 5}},
	{"key": "F6", "command": "change_tab", "args": {"tab": 6}},
	{"key": "F7", "command": "change_tab", "args": {"tab": 7}},
	{"key": "F8", "command": "change_tab", "args": {"tab": 8}},
	{"key": "F9", "command": "change_tab", "args": {"tab": 9}},
	{"key": "F10", "command": "change_tab", "args": {"tab": 10}},
	{"key": "F11", "command": "change_tab", "args": {"tab": 11}},
	{"key": "F12", "command": "change_tab", "args": {"tab": 12}}
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
		"window_fill": { "bg": {"color": "black"} },
		
		"element_normal": { "bg": {"color": "black"} },
		"element_input_normal": { "fg": {"color": "yellow"}, "bg": {"color": "black"} },
		"element_marked": { "bg": {"color": "yellow"} },
		"element_selected": { "bg": {"color": "cyan"} },
		"element_selected_marked": { "bg": {"color": "blue"} },
		
		"message_timestamp": { "fg": {"color": "blue"} },
		"message_server_channel": { "fg": {"color": "green"} },
		"message_direct_channel": { "fg": {"color": "magenta", "bold": true} },
		"message_author": { "fg": {"color": "cyan", "bold": true} },
		"message_content": {},
		"message_log": {"fg": {"color":"yellow"}},
		
		"title_bar": { "fg": {"color": "green", "bold": true, "underline": true} },
		"notifications_bar": { "bg": {"color": "blue"}	},
		"typing_bar": { "fg": {"color": "cyan"}	},
		"text_other": {},
		"text_special": { "fg": {"color": "cyan"} },
		"input_chat": {},
		"input_other": {},
		"send_prompt": { "fg": {"color": "green", "bold": true}	},
		"search": { "fg": {"color": "yellow", "bold": true}	},
		
		"tab_normal": {},
		"tab_selected": { "bg": {"color": "blue"}},
		"tab_activity": { "bg": {"color": "yellow"}	},
		"tab_mention": { "bg": {"color": "cyan"} }
	}
}`)
