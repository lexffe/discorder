#Discorder

An interactive command line discord client (Would not reccomend for use atleast before 0.1, 0.2 will be a lot more usable with history and other stuff)

##Installing

#### Latest Alpha

Try the latest alpha versio here: https://github.com/jonas747/discorder/releases/tag/v0.1.1

####compiling from repo

You need go and git installed

run this command: `go get -u github.com/jonas747/discorder`

(-u: to force an update if you have a older version)

After that there should be a built executable in your $GOPATH/bin folder

##Features

Yup, very much in development.

 - Should be light on resource usage
     + Maybe not so much in this early development stage where everything is still getting set up
 - Very early in development, a lot of missing features right now
 - Voice support will be added laaaaaater
 - 

Current alpha release has the following implemented:

 - Sending/receiving messages
     + You also received the changes when they get edited and removed
 - Sending/receiving direct messages EXCEPT for initiating new conversations
 - State will be saved when you leave and restored when you open again
 - Listen in on multiple channels at once!
 - History is fetched of each channel you're listening on

Latest git version has these things:
 
 - @mention autocompletion
 - Typing status, both receiving and sending
 
Next on the list to be worked on is history, and the channel listening system mentioned above

## Usage

Run the executable and follow the instructions on screen

Controlls:

 - ctrl-o: Opens help 
 - ctrl-s: Select server
     + space marks a channel for listening
     + enter selects the channel for typing
 - ctrl-g: select channel 
 - ctrl-p: select private channel (direct messages)
 - ctrl-j: Queries the log for the current channel (for debugging, or when you think it missed a new message)
 - ctrl-l: Clears the log, will later be changed to toggle hiding the log, and you can view the log in a seperate window, but thats for later...!
 - ctrl-q: Quit
 - backspace: closes the active window if any

##Screenshots

![Logging in](https://dl.dropboxusercontent.com/u/17487167/screenshots/2016-03-16T01%3A00%3A23%2B01%3A00.png)

![Channels list](https://dl.dropboxusercontent.com/u/17487167/screenshots/2016-03-16T03%3A57%3A45%2B01%3A00.png)

![Direct messages](https://dl.dropboxusercontent.com/u/17487167/screenshots/2016-03-18T04%3A15%3A40%2B01%3A00.png)


