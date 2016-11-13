# Discorder

### I'm not actively developing this anymore, so don't expect any new features anytime soon

![Ayy](https://dl.dropboxusercontent.com/u/17487167/screenshots/1476387883.png)

Discorder is an interactive command line discord client.

Discord server: https://discord.gg/0vYlUK2XBKlmxGrX

Join for updates, to suggest stuff, complain about stuff, or just talk about stuff

##Installing/Running

#### Latest Alpha

Try the latest alpha version here: https://github.com/jonas747/discorder/releases

####compiling from repo

You need go and git installed, and your gopath set up

run this command: `go get -u github.com/jonas747/discorder/cmd/discorder`

(-u: to force an update if you have a older version)

After that there should be a built executable in your $GOPATH/bin folder

##Features

Note: Discorder still lacks a lot of features, so i wouldn't consider using it as a full replacement yet

 - Light on resource usage
     + This depends on how many tabs you have open and how many channels you're listening in on offcourse
     + Maybe not so much in this early development stage where everything is still getting set up and organised, but will be a focus later on
 - Customizable
     + Discorder is very customizable, you can change the looks of it with your own themes and change the keybinds (See doc for more info)

Feature list:

- [x] Sending receiving messages and dm's
- [x] Multiple channels in one view
- [x] Initiating new dm conversations
- [x] Delete messages
- [ ] Edit messages
- [x] Custom keybinds
- [x] Custom themes
- [x] Notifications
- [x] Change server notifications settings
- [x] Mention auto completion
- [x] Tabs
- [x] Typing events
- [x] History
- [x] Persistent state, tabs will be saved to config when exiting
- [x] Nicknames with optional hiding of nicknames
- [ ] Change user settings
- [ ] Discord status (idle status, game playing), you can set it using the command currently but not view it
- [ ] Member list  
- [ ] Message pinning
- [ ] Server management
- [ ] Voice
- [ ] Friends and other relationship stuff (block etc)
- [ ] Invite (you can open them but not create)

## Usage

Run the executable and follow the instructions on screen

Keybinds: See [doc/keybinds.md](https://github.com/jonas747/discorder/blob/master/doc/keybinds.md) for keybind configuration and [doc/defaults.md](https://github.com/jonas747/discorder/blob/master/doc/defaults.md) for defaults

Quick start:

1. log in using token or username/pw
2. ctrl-s to open server/channel list
3. mark servers/channels for listening with ctrl-space
4. set as sending channel with enter
5. close out of windows wih alt-backspace
6. f1-12 for tabs
7. ctrl-x to open command menu
8. change discorder settings such as randomized colors, short guild names, hide nicknames etc in discorder_setttings which you can find in the command menu

## Dependencies

Discorder depends on termbox, discordgo, and go-runewidth at compile time

#### Optional dependencies

 - xdg-open: Used for opening links
 - notify-send: Used for Linux notifications
 - terminal-emulator: Used for macOS notifications
