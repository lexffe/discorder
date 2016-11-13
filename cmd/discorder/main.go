package main

import (
	"flag"
	"fmt"
	"github.com/jonas747/discorder/discorder"
	"log"
	"net/http"
	_ "net/http/pprof"
)

var (
	flagConfigPath   = flag.String("c", "", "Custom path to the config file")
	flagThemePath    = flag.String("t", "", "Path to a theme file, as opposed to the one in the config file")
	flagLogPath      = flag.String("l", "discorder.log", "Path to output logs, only used with debug enabled")
	flagClearToken   = flag.Bool("r", false, "Set to reset token")
	flagDGoDebugLvl  = flag.Int("g", 0, "discordgo logging level (0 - Errors, 1 - Warnings, 2 - Info, 3 - Debug")
	flagDebugEnabled = flag.Bool("d", false, "Set to enable debuging mode")
)

func main() {
	flag.Parse()

	logPath := ""
	if *flagDebugEnabled {
		logPath = *flagLogPath
	}
	discorder.InitLogging(logPath)

	if *flagDebugEnabled {
		go RunPProf()
	}

	path, err := discorder.GetCreateConfigDir()
	if err != nil {
		panic(err)
	}
	log.Println("Config path is", path)

	options := &discorder.AppOptions{
		CustomConfigPath:    *flagConfigPath,
		CustomThemePath:     *flagThemePath,
		DebugEnabled:        *flagDebugEnabled,
		DiscordgoDebugLevel: *flagDGoDebugLvl,
		ClearToken:          *flagClearToken,
	}

	app, err := discorder.NewApp(options)
	if err != nil {
		log.Println("Error setting up discorder :(", err)
		return
	}
	app.Run()

	discorder.StopLogger()

	fmt.Println("bye :'(.....")
}

func RunPProf() {
	log.Println(http.ListenAndServe("localhost:6060", nil))
}
