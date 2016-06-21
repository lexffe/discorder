# Configuring keybinds

If you want to add/change/remove a keybind you can do that in your keybinds-user.json file in either ~/.config/discorder for unix systems or %APPDATA%/discorder windows

There should also be a file called keybinds-default.json, changing this will have no effect as it's hardcoded into discorder, it's purpose is to show the defaults as a reference

For a full list of commands see commands.md, special keys are listed at the bottom

NOTES: Some keys will trigger eachother, check out the bottom for a list of keys with the same keycodes, Some keys will also flat out not work such as the esc key (Reason is because discorder runs using termbox alt mode and not esc mode as input)

### Change a Keybind

If you want to change say the open server keybind from ctrl-s to ctrl-w you would put this into keybinds-user.json:

```json
[
    {"key": "CtrlS", "command": "nop"},
    {"key": "CtrlW", "command": "servers"}
]
```

First we unbind the CtrlS key to "nop" (No operation), since user binds overrides default binds, then we set CtrlW to servers

### Sequences and arguments

```json
{"key": "CtrlC-d", "command": "delete_message", "args":{"which": "last"}}
```

Here we have a key sequence, doing CtrlC and then d will delete our last message

### Open the execute window

If you set `open_exec_window` To true the execute window will open for that command

## Special keys:

 - F1
 - F2
 - F3
 - F4
 - F5
 - F6
 - F7
 - F8
 - F9
 - F10
 - F11
 - F12
 - Insert
 - Delete
 - Home
 - End
 - Pgup
 - Pgdn
 - ArrowUp
 - ArrowDown
 - ArrowLeft
 - ArrowRight
 - MouseLeft
 - MouseMiddle
 - MouseRight
 - MouseRelease
 - MouseWheelUp
 - MouseWheelDown
 - CtrlTilde
 - CtrlSpace
 - CtrlA
 - CtrlB
 - CtrlC
 - CtrlD
 - CtrlE
 - CtrlF
 - CtrlG
 - Backspace
 - CtrlH
 - Tab
 - CtrlI
 - CtrlJ
 - CtrlK
 - CtrlL
 - Enter
 - CtrlM
 - CtrlN
 - CtrlO
 - CtrlP
 - CtrlQ
 - CtrlR
 - CtrlS
 - CtrlT
 - CtrlU
 - CtrlV
 - CtrlW
 - CtrlX
 - CtrlY
 - CtrlZ
 - Esc
 - CtrlLsqBracket
 - CtrlBackslash
 - CtrlRsqBracket
 - CtrlSlash
 - CtrlUnderscore
 - Space
 - Backspace2
 - Ctrl2
 - Ctrl3
 - Ctrl4
 - Ctrl5
 - Ctrl6
 - Ctrl7
 - Ctrl8

## Keys with same keycodes

Keys with same keycode will trigger eachothers commands

 - CtrlTilde       = 0x00
 - Ctrl2           = 0x00
 - CtrlSpace       = 0x00

 - Backspace       = 0x08
 - CtrlH           = 0x08

 - Tab             = 0x09
 - CtrlI           = 0x09

 - Enter           = 0x0D
 - CtrlM           = 0x0D

 - Esc             = 0x1B
 - CtrlLsqBracket  = 0x1B
 - Ctrl3           = 0x1B

 - Ctrl4           = 0x1C
 - CtrlBackslash   = 0x1C

 - Ctrl5           = 0x1D
 - CtrlRsqBracket  = 0x1D

 - Ctrl7           = 0x1F
 - CtrlSlash       = 0x1F
 - CtrlUnderscore  = 0x1F

 - Backspace2      = 0x7F
 - Ctrl8           = 0x7F
