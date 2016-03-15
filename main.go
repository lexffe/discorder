// This file provides a basic "quick start" example of using the Discordgo
// package to connect to Discord using the low level API functions.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
)

var (
	channels map[string]*discordgo.Channel
	app      *App
)

func main() {

	var err error

	// Check for Username and Password CLI arguments.
	if len(os.Args) != 3 {
		fmt.Println("You must provide username and password as arguments. See below example.")
		fmt.Println(os.Args[0], " [username] [password]")
		return
	}

	app, err = Login(os.Args[1], os.Args[2])
	if err != nil {
		log.Println(err)
		return
	}
	app.Run()
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated user has access to.

// func getChannel(id string) (*discordgo.Channel, error) {
// 	if channels == nil {
// 		channels = make(map[string]*discordgo.Channel)
// 	}
// 	channel, ok := channels[id]
// 	if !ok {
// 		ch, err := session.Channel(id)
// 		if err != nil {
// 			return nil, err
// 		}
// 		channels[id] = ch
// 		channel = ch
// 		fmt.Println("Fetched new channel")
// 	}
// 	return channel, nil
// }

// func changeState(newState State) {

// }
