package main

import (
	"flag"
	"fmt"
	"github.com/jonas747/discordgo"
	"log"
	"net/http"
	_ "net/http/pprof"
)

const (
	VERSION_MAJOR = 0
	VERSION_MINOR = 3
	VERSION_PATCH = 0
	VERSION_NOTE  = "Fruity-Alpha"
)

var (
	VERSION = fmt.Sprintf("%d.%d.%d-%s", VERSION_MAJOR, VERSION_MINOR, VERSION_PATCH, VERSION_NOTE)
)

var (
	channels    map[string]*discordgo.Channel
	application *App
	config      *Config

	configPath       = flag.String("config", "discorder.json", "Path to the config file")
	flagLogPath      = flag.String("log", "discorder.log", "Path to output logs, only used with debug enabled")
	flagDebugEnabled = flag.Bool("debug", false, "Set to enable debuging mode")
	flagDumpAPI      = flag.Bool("dumpapi", false, "Set to enable debug in discordgo")
)

func main() {
	flag.Parse()

	c, err := LoadConfig(*configPath)
	if err != nil {
		c = &Config{}
		log.Println("Failed to open config, creating new one")
		c.Save(*configPath)
	}

	config = c

	if *flagDebugEnabled {
		// Below used when panics thats not recovered from occurs and it smesses up the terminal :'(
		// logFile, _ := os.OpenFile("discorder_stdout_stderr.log", os.O_WRONLY|os.O_CREATE|os.O_SYNC, 0755)
		// syscall.Dup2(int(logFile.Fd()), 1)
		// syscall.Dup2(int(logFile.Fd()), 2)
		go RunPProf()
	}

	application = NewApp(config, *flagLogPath)
	application.Run()
}

func RunPProf() {
	log.Println(http.ListenAndServe("localhost:6060", nil))
}
