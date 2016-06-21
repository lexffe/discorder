# Discorder

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

Note: Discorder still lacks a lot of features, so i wouldn't consider using it as a full replacement just yet

 - Light on resource usage
     + This depends on how many tabs you have open and how many channels you're listening in on offcourse
     + Maybe not so much in this early development stage where everything is still getting set up and organised, but will be a focus later on
 - Customizeble
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

Keybinds:

After you run it once, a keybind file will be generated in the config dir either at ~/.config/discorder for unix or %appdata%/discorder for windows

Look there for keybinds

## Dependencies

Discorder depends on termbox, discordgo, and go-runewidth at compile time

#### Optional dependencies

 - xdg-open: Used for opening links
 - notify-send: Used for notifications

##Screenshots

![Typing status](https://dl.dropboxusercontent.com/u/17487167/screenshots/2016-04-07T16%3A18%3A02%2B02%3A00.png)

![Mention auto complete](https://dl.dropboxusercontent.com/u/17487167/screenshots/2016-04-07T16%3A19%3A10%2B02%3A00.png)


![Logging in](https://dl.dropboxusercontent.com/u/17487167/screenshots/2016-03-16T01%3A00%3A23%2B01%3A00.png)

![Channels list](https://dl.dropboxusercontent.com/u/17487167/screenshots/2016-03-16T03%3A57%3A45%2B01%3A00.png)

![Direct messages](https://dl.dropboxusercontent.com/u/17487167/screenshots/2016-03-18T04%3A15%3A40%2B01%3A00.png)


