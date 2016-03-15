package main

import (
	"fmt"
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
)

const (
	VERSION = "0.0.1 DEV"
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
