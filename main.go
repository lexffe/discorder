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
	flagDGoDebugLvl  = flag.Int("dgo", 0, "discordgo logging level (0 - Errors, 1 - Warnings, 2 - Info, 3 - Debug")
	flagConfigPath   = flag.String("config", "discorder.json", "Path to the config file")
	flagLogPath      = flag.String("log", "discorder.log", "Path to output logs, only used with debug enabled")
	flagDebugEnabled = flag.Bool("debug", false, "Set to enable debuging mode")
)

func main() {
	flag.Parse()

	logPath := ""
	if *flagDebugEnabled {
		logPath = *flagLogPath
	}
	discorder.InitLogging(logPath)

	if *flagDebugEnabled {
		// Below used when panics thats not recovered from occurs and it smesses up the terminal :'(
		// logFile, _ := os.OpenFile("discorder_stdout_stderr.log", os.O_WRONLY|os.O_CREATE|os.O_SYNC, 0755)
		// syscall.Dup2(int(logFile.Fd()), 1)
		// syscall.Dup2(int(logFile.Fd()), 2)
		go RunPProf()
	}

	app, err := discorder.NewApp(*flagConfigPath, *flagDebugEnabled, *flagDGoDebugLvl)
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
