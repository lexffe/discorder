package main

import (
	"flag"
	"fmt"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	VERSION_MAJOR = 0
	VERSION_MINOR = 3
	VERSION_PATCH = 0
	VERSION_NOTE  = "Git-silly"
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
OUTER:
	for _, typing := range t.typing {
		for _, filterItem := range filter {
			if typing.t.ChannelID == filterItem {
				out = append(out, typing.t)
				continue OUTER
			}
		}
	}
	t.Unlock()
	return out
}
