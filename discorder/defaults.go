package discorder

// The reason i have this in json encoded constants instead of struct literals
// Is to have the formatting when written to disk be neat

var DefaultKeybinds = []byte(`[
	{"key": "CtrlQ", "command": "quit"},
	{"key": "CtrlS", "command": "servers"},
	{"key": "CtrlO", "command": "help"},
	{"key": "CtrlX", "command": "commands"},
	{"key": "CtrlL", "command": "clear_log"},
	{"key": "CtrlT", "command": "reload_theme"},
	{"key": "CtrlN", "command": "set_nick", "open_exec_window": true},
	{"key": "CtrlZ", "command": "open_last_link"},
	{"key": "CtrlC-Backspace", "command": "close_windows"},
	{"key": "CtrlC-Backspace2", "command": "close_windows"},

	{"key": "ArrowLeft", "command": "move_cursor", "args": {"amount": 1, "direction": "left"}},
	{"key": "ArrowRight", "command": "move_cursor", "args": {"amount": 1, "direction": "right"}},
	{"key": "ArrowUp", "command": "scroll", "args": {"amount": 1, "direction": "up"}},
	{"key": "ArrowDown", "command": "scroll", "args": {"amount": 1, "direction": "down"}},
	{"key": "Pgup", "command": "scroll", "args": {"amount": 10, "direction": "up"}},
	{"key": "Pgdn", "command": "scroll", "args": {"amount": 10, "direction": "down"}},
	
	{"key": "Enter", "command": "select"},
	{"key": "CtrlSpace", "command": "toggle"},
	{"key": "Tab", "command": "autocomplete_selection", "args": {"amount": 1}},
	
	{"key": "Backspace", "command": "erase", "args": {"amount": 1, "direction": "left"}},
	{"key": "Backspace2", "command": "erase", "args": {"amount": 1, "direction": "left"}},
	{"key": "CtrlW", "command": "erase", "args": {"amount": 1, "direction": "left", "words": true}},
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
	{"key": "F12", "command": "change_tab", "args": {"tab": 12}},
	

	{"key": "CtrlC-ArrowLeft", "command": "change_tab", "args": {"change": -1}},
	{"key": "CtrlC-ArrowRight", "command": "change_tab", "args": {"change": 1}}
]`)

var DefaultTheme = []byte(`{
    "name": "Default theme",
    "author": "jonas747",
    "comment": "The default discorder theme",
    "color_mode": 1,
    "discrim_table":[
    	{"fg": {"color":"red", "bold": true} },
    	{"fg": {"color":"green", "bold": true} },
    	{"fg": {"color":"green"} },
    	{"fg": {"color":"yellow", "bold": true} },
    	{"fg": {"color":"yellow"} },
    	{"fg": {"color":"blue", "bold": true} },
    	{"fg": {"color":"blue"} },
    	{"fg": {"color":"magenta", "bold": true} },
    	{"fg": {"color":"magenta"} },
    	{"fg": {"color":"cyan", "bold": true} },
    	{"fg": {"color":"cyan"} }
    ],
    "theme":{
		"window_border": {
			"fg": {"color": "white", "bold": true},
			"bg": {"color": "black"}
		},
		"window_fill": { "bg": {"color": "black"} },
		
		"menu_info_border": {"bg": {"color": "black"}},
		"menu_info_fill": {"bg": {"color": "black"}},
		"menu_info_text": {"bg": {"color": "black"}},
		"menu_search": {"fg": {"color": "cyan"}, "bg": {"color": "black"}},
		"element_normal": { "bg": {"color": "black"} },
		"element_input_normal": { "fg": {"color": "yellow"}, "bg": {"color": "black"} },
		"element_marked": { "bg": {"color": "yellow"} },
		"element_selected": { "bg": {"color": "cyan"} },
		"element_selected_marked": { "bg": {"color": "blue"} },

		"message_timestamp": { "fg": {"color": "blue"} },
		"message_server": { "fg": {"color": "green"} },
		"message_server_channel": { "fg": {"color": "green"} },
		"message_direct_channel": { "fg": {"color": "magenta", "bold": true} },
		"message_author": { "fg": {"color": "cyan", "bold": true} },
		"message_content": {},
		"message_log": {"fg": {"color":"yellow"}},
		
		"scroll_text": {"fg": {"color": "cyan"}},

		"title_bar": { "fg": {"color": "green", "bold": true, "underline": true} },
		"notifications_bar": { "bg": {"color": "blue"}	},
		"typing_bar": { "fg": {"color": "cyan"}	},
		
		"text_window_normal": { "bg": {"color": "black"} },
		"text_window_special": { "fg": {"color": "cyan"}, "bg": {"color": "black"} },
		
		
		"input_chat": {},
		"send_prompt": { "fg": {"color": "green", "bold": true}	},
		"search": { "fg": {"color": "yellow", "bold": true}	},
		
		"tab_normal": {},
		"tab_selected": { "bg": {"color": "blue"}},
		"tab_activity": { "bg": {"color": "yellow"}	},
		"tab_mention": { "bg": {"color": "cyan"} }
	}
}`)
