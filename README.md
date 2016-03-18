#Discorder

An interactive command line discord client (Would not reccomend for use atleast before 0.1, 0.2 will be a lot more usable with history and other stuff)

Yup, very much in development.

 - Should be light on resource usage
     + Maybe not so much in this early development stage where everything is still getting set up
 - Just started so not much added, and what is is somewhat buggy, see tasks.TODO for details
 - Voice support might be added laaaaaater

What works:
 - Sending/receiving messages
     + You also received the changes when they get edited and removed
 - Sending/receiving direct messages EXCEPT for initiating new conversations
 - State will be saved when you leave and restored when you open again

Next on the list to be worked on is history, and the channel listening system mentioned above

Controlls:

 - ctrl-h: Opens help 
 - ctrl-s: Select server
     + space marks a channel for listening
     + enter selects the channel for typing
 - ctrl-g: select channel 
 - ctrl-p: select private channel (direct messages)
 - ctrl-j: Queries the log for the current channel (for debugging, or when you think it missed a new message)
 - ctrl-l: Clears the log, will later be changed to toggle hiding the log, and you can view the log in a seperate window, but thats for later...!
 - ctrl-q: Quit
 - backspace: closes the active window if any

Heres an image of its current state

![Logging in](https://dl.dropboxusercontent.com/u/17487167/screenshots/2016-03-16T01%3A00%3A23%2B01%3A00.png)

![Channels list](https://dl.dropboxusercontent.com/u/17487167/screenshots/2016-03-16T03%3A57%3A45%2B01%3A00.png)

![Direct messages](https://dl.dropboxusercontent.com/u/17487167/screenshots/2016-03-18T04%3A15%3A40%2B01%3A00.png)


