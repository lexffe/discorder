package main

import (
	"github.com/jonas747/discordgo"
	"log"
	"sync"
	"time"
)

type TypingWrapper struct {
	t    *discordgo.TypingStart
	last time.Time
}

type TypingRoutine struct {
	sync.Mutex
	app          *App
	typingEvtIn  chan *discordgo.TypingStart
	selfTypingIn chan string
	typing       []*TypingWrapper
	stop         chan bool
}

func NewTypingRoutine(app *App) *TypingRoutine {
	return &TypingRoutine{
		app:          app,
		typingEvtIn:  make(chan *discordgo.TypingStart),
		selfTypingIn: make(chan string),
		typing:       make([]*TypingWrapper, 0),
		stop:         make(chan bool),
	}
}

func (t *TypingRoutine) Run() {
	ticker := time.NewTicker(5 * time.Second)
	selfTyping := ""
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

			if selfTyping != "" {
				err := t.app.session.ChannelTyping(selfTyping)
				if err != nil {
					log.Println("Error sending typing: ", err)
				}
				selfTyping = ""
			}

			t.Unlock()
		case selfTyping = <-t.selfTypingIn:
			// ...
		case typingEvt := <-t.typingEvtIn:
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
		case <-t.stop:
			ticker.Stop()
			return
		}
	}
}

func (t *TypingRoutine) GetTyping(filter []string) []*discordgo.TypingStart {
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
