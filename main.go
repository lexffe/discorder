package main

import (
	"flag"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

const (
	VERSION_MAJOR = 0
	VERSION_MINOR = 1
	VERSION_PATCH = 1
	VERSION_NOTE  = "Alpha"
)

var (
	VERSION = fmt.Sprintf("%d.%d.%d-%s", VERSION_MAJOR, VERSION_MINOR, VERSION_PATCH, VERSION_NOTE)
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
