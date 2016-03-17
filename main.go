package main

import (
	"fmt"
	"os"

	"github.com/bwmarrin/discordgo"
)

const (
	VERSION = "0.0.3 DEV"
)

var (
	channels    map[string]*discordgo.Channel
	application *App
	config      *Config

	configPath = "config.json"
)

func main() {

	// Check for Username and Password CLI arguments.
	// if len(os.Args) != 3 {
	// 	fmt.Println("You must provide username and password as arguments. See below example.")
	// 	fmt.Println(os.Args[0], " [username] [password]")
	// 	return
	// }

	if len(os.Args) >= 2 {
		configPath = os.Args[1]
	}

	c, err := LoadConfig(configPath)
	if err != nil {
		c = &Config{}
		fmt.Println("Failed to open config, creating new one")
		c.Save(configPath)
	}

	config = c

	application = NewApp(config)
	application.Run()
}
