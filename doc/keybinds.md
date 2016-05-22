# Configuring keybinds

If you wanna add/change/remove a keybind you do that in your keybinds-user.json file in either ~/.config/discorder for unix systems or %APPDATA%/discorder windows

There should also be a file called keybinds-default.json, changing this will have no effect as it's hardcoded into discorder, it's purpose is to show the defaults

For a full list of commands see commands, special keys are listed at the bottom

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