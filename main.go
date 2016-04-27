package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/jonas747/discordgo"
)

const (
	VERSION_MAJOR = 0
	VERSION_MINOR = 3
	VERSION_PATCH = 0
	VERSION_NOTE  = "Git-lemony"
)

var (
	VERSION = fmt.Sprintf("%d.%d.%d-%s", VERSION_MAJOR, VERSION_MINOR, VERSION_PATCH, VERSION_NOTE)
)

var (
	channels    map[string]*discordgo.Channel
	application *App
	config      *Config

	configPath       = flag.String("config", "discorder.json", "Path to the config file")
	flagLogPath      = flag.String("log", "discorder.log", "Path to output logs")
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

	//Below used when panics thats not recovered from occurs and it smesses up the terminal :'(
	logFile, _ := os.OpenFile("hmpf", os.O_WRONLY|os.O_CREATE|os.O_SYNC, 0755)
	syscall.Dup2(int(logFile.Fd()), 1)
	syscall.Dup2(int(logFile.Fd()), 2)
	go RunPProf()

	application = NewApp(config, *flagLogPath)
	application.Run()
}

type TypingWrapper struct {
	t    *discordgo.TypingStart
	last time.Time
}

type TypingManager struct {
	sync.Mutex
	in     chan *discordgo.TypingStart
	typing []*TypingWrapper
}

func (t *TypingManager) Run() {
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ticker.C:
			t.Lock()
			newTyping := make([]*TypingWrapper, 0)
			for _, v := range t.typing {
				if time.Since(v.last) < 5*time.Second {
					newTyping = append(newTyping, v)
				}
			}
			t.typing = newTyping
			t.Unlock()
		case typingEvt := <-t.in:
			t.Lock()
			found := false
			for _, v := range t.typing {
				if v.t.ChannelID == typingEvt.ChannelID && v.t.UserID == typingEvt.UserID {
					v.last = time.Now()
					found = true
					break
				}
			}
			if !found {
				t.typing = append(t.typing, &TypingWrapper{t: typingEvt, last: time.Now()})
			}
			t.Unlock()
		}
	}
}

func (t *TypingManager) GetTyping(filter []string) []*discordgo.TypingStart {
	out := make([]*discordgo.TypingStart, 0)
	t.Lock()
	if len(filter) == 0 {
		out = make([]*discordgo.TypingStart, len(t.typing))
		for k, typing := range t.typing {
			out[k] = typing.t
		}
	} else {
	OUTER:
		for _, typing := range t.typing {
			for _, filterItem := range filter {
				if typing.t.ChannelID == filterItem {
					out = append(out, typing.t)
					continue OUTER
				}
			}
		}
	}
	t.Unlock()
	return out
}

func RunPProf() {
	log.Println(http.ListenAndServe("localhost:6060", nil))
}
