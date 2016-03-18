package main

import (
	"flag"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

const (
	VERSION = "Alpha 0.1.0-RC1"
)

var (
	channels    map[string]*discordgo.Channel
	application *App
	config      *Config

	configPath  = "discorder.json"
	flagLogPath = flag.String("log", "discorder.log", "Path to output logs")
)

func main() {
	flag.Parse()

	// if len(os.Args) >= 2 {
	// 	configPath = os.Args[1]
	// }

	c, err := LoadConfig(configPath)
	if err != nil {
		c = &Config{}
		fmt.Println("Failed to open config, creating new one")
		c.Save(configPath)
	}

	config = c

	application = NewApp(config, *flagLogPath)
	application.Run()
}
