# Configuring keybinds

If you wanna add/change/remove a keybind you do that in your keybinds-user.json file in either ~/.config/discorder for unix systems or %APPDATA%/discorder windows

There should also be a file called keybinds-default.json, changing this will have no effect as it's hardcoded into discorder, it's purpose is to show the defaults

For a full list of commands see commands, special keys are listed at the bottom

### Change a Keybind

If you want to change say the open server keybind from ctrl-s to ctrl-w you would put this into keybinds-user.json:

```json
[
    {"key": "CtrlS", "command": "nop"},
    {"key": "CtrlW", "command": "open_servers"}
]
```

First we unbind the CtrlS key to "nop" (No operation), since user binds overrides default binds, then we set CtrlW to open_servers

Here is a little more advanced example:

```json
{"key": "CtrlC-d", "command": "delete_message", "args":{"which": "last"}}
```

Here we have a key sequence, doing CtrlC and then d will delete our last message